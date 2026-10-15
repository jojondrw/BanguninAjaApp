package billing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	onchainMethod = "onchain"

	onchainInvoiceTarget    = "invoice"
	onchainReceivableTarget = "receivable"

	onchainPending   = "pending"
	onchainPaid      = "paid"
	onchainExpired   = "expired"
	onchainFailed    = "failed"
	onchainCancelled = "cancelled"

	verificationPaid    = "paid"
	verificationWaiting = "waiting"
	verificationFailed  = "failed"

	onchainTokenBytes  = 24
	onchainTokenLength = 32
	paymentNoteLimit   = 200
	failureReasonLimit = 300

	messagePaid = "Pembayaran sudah diterima dan tercatat di tagihan"
)

var (
	errOnchainDisabled      = apperror.New(http.StatusServiceUnavailable, "onchain_disabled", "Pembayaran on-chain belum diaktifkan di server ini")
	errOnchainNotFound      = apperror.NotFound("onchain_request_not_found", "Permintaan bayar tidak ditemukan. Periksa lagi tautannya")
	errOnchainTargetMissing = apperror.NotFound("onchain_target_not_found", "Tagihan untuk permintaan bayar ini sudah tidak ada")
	errOnchainVendorInvoice = apperror.Unprocessable("onchain_vendor_invoice", "Pembayaran on-chain hanya untuk faktur ke pelanggan, bukan faktur vendor")
	errOnchainSettled       = apperror.Unprocessable("onchain_nothing_outstanding", "Tagihan ini sudah lunas, tidak ada sisa yang bisa dibayar")
	errOnchainActive        = apperror.Conflict("onchain_request_active", "Masih ada permintaan bayar on-chain yang aktif untuk tagihan ini. Batalkan dulu atau tunggu sampai kedaluwarsa")
	errOnchainNotPending    = apperror.Conflict("onchain_request_not_pending", "Permintaan bayar ini sudah tidak menunggu pembayaran")
	errOnchainCancelled     = apperror.Conflict("onchain_request_cancelled", "Permintaan bayar ini sudah dibatalkan. Minta tautan baru ke bagian keuangan")
	errOnchainPaidByOther   = apperror.Conflict("onchain_request_paid", "Permintaan bayar ini sudah lunas dengan transaksi lain")
	errOnchainTxUsed        = apperror.Conflict("onchain_tx_used", "Hash transaksi ini sudah dipakai untuk pembayaran lain")

	onchainReadErrors  = database.ErrorMap{NotFound: errOnchainNotFound}
	onchainWriteErrors = database.ErrorMap{NotFound: errOnchainNotFound, Duplicate: errOnchainTxUsed}

	settlementRejections = []string{"payment_exceeds_outstanding", "payment_parent_not_found", "invoice_not_found", "receivable_not_found"}
)

func onchainAmountExceeds(outstanding int64) *apperror.Error {
	return apperror.Unprocessable("onchain_amount_exceeds_outstanding", fmt.Sprintf(
		"Nominal melebihi sisa tagihan, sisanya tinggal %s", rupiah(outstanding),
	))
}

func onchainFailedEarlier(reason string) *apperror.Error {
	return apperror.Conflict("onchain_request_failed", fmt.Sprintf(
		"Permintaan bayar ini gagal diproses: %s. Hubungi bagian keuangan", reason,
	))
}

func onchainSettlementFailed(reason string) *apperror.Error {
	return apperror.Unprocessable("onchain_settlement_failed", fmt.Sprintf(
		"Transaksi sudah diterima di jaringan, tetapi tidak bisa dicatat ke tagihan: %s. Hubungi bagian keuangan", reason,
	))
}

type onchainTarget struct {
	kind string
	id   uuid.UUID
}

type OnchainService interface {
	Config() OnchainConfigResponse
	CreateInvoiceRequest(ctx context.Context, id uuid.UUID, request OnchainRequestCreate, actor *uuid.UUID) (OnchainRequestResponse, error)
	CreateReceivableRequest(ctx context.Context, id uuid.UUID, request OnchainRequestCreate, actor *uuid.UUID) (OnchainRequestResponse, error)
	InvoiceRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[OnchainRequestResponse], error)
	ReceivableRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[OnchainRequestResponse], error)
	CancelRequest(ctx context.Context, id uuid.UUID) (OnchainRequestResponse, error)
	PublicPayment(ctx context.Context, token string) (PublicPaymentResponse, error)
	ConfirmPayment(ctx context.Context, token string, request OnchainConfirmRequest) (OnchainConfirmResponse, error)
}

type onchainService struct {
	repository Repository
	chain      ChainReader
	settings   config.Onchain
	now        func() time.Time
	newToken   func() (string, error)
}

func NewOnchainService(repository Repository, chain ChainReader, settings config.Onchain) OnchainService {
	return &onchainService{repository: repository, chain: chain, settings: settings, now: time.Now, newToken: randomToken}
}

func IsPaymentToken(token string) bool {
	if len(token) != onchainTokenLength {
		return false
	}
	_, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil
}

func randomToken() (string, error) {
	raw := make([]byte, onchainTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate payment token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *onchainService) Config() OnchainConfigResponse {
	if !s.settings.Enabled() {
		return OnchainConfigResponse{ChainID: s.settings.ChainID}
	}
	return OnchainConfigResponse{
		Enabled:           true,
		ChainID:           s.settings.ChainID,
		Receiver:          s.settings.ReceiverAddress,
		IDRPerETH:         s.settings.IDRPerETH,
		ExplorerURL:       s.settings.ExplorerURL,
		MinConfirmations:  s.settings.MinConfirmations,
		RequestTTLMinutes: int(s.settings.RequestTTL / time.Minute),
	}
}

func (s *onchainService) CreateInvoiceRequest(ctx context.Context, id uuid.UUID, request OnchainRequestCreate, actor *uuid.UUID) (OnchainRequestResponse, error) {
	return s.create(ctx, onchainTarget{kind: onchainInvoiceTarget, id: id}, request.Amount, actor)
}

func (s *onchainService) CreateReceivableRequest(ctx context.Context, id uuid.UUID, request OnchainRequestCreate, actor *uuid.UUID) (OnchainRequestResponse, error) {
	return s.create(ctx, onchainTarget{kind: onchainReceivableTarget, id: id}, request.Amount, actor)
}

func (s *onchainService) create(ctx context.Context, target onchainTarget, amount int64, actor *uuid.UUID) (OnchainRequestResponse, error) {
	if !s.settings.Enabled() {
		return OnchainRequestResponse{}, errOnchainDisabled
	}
	now := s.now()
	var created OnchainRequest
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		parents, err := lockOnchainTarget(ctx, repository, target)
		if err != nil {
			return err
		}
		requested, err := requestableAmount(parents, amount)
		if err != nil {
			return err
		}
		if err := ensureNoActiveRequest(ctx, repository, parents.onchainTargets(), now); err != nil {
			return err
		}
		created, err = s.newRequest(target, requested, actor, now)
		if err != nil {
			return err
		}
		return onchainWriteErrors.Resolve(repository.CreateOnchainRequest(ctx, &created))
	})
	if err != nil {
		return OnchainRequestResponse{}, apperror.From(err)
	}
	return s.requestResponse(ctx, created.ID)
}

func lockOnchainTarget(ctx context.Context, repository Repository, target onchainTarget) (paymentParents, error) {
	if target.kind == onchainInvoiceTarget {
		return lockInvoicePayment(ctx, repository, target.id)
	}
	return lockReceivablePayment(ctx, repository, target.id)
}

func requestableAmount(parents paymentParents, requested int64) (int64, error) {
	if parents.invoice != nil && parents.invoice.PartyType != customerParty {
		return 0, errOnchainVendorInvoice
	}
	outstanding := parents.outstanding()
	switch {
	case outstanding <= 0:
		return 0, errOnchainSettled
	case requested == 0:
		return outstanding, nil
	case requested > outstanding:
		return 0, onchainAmountExceeds(outstanding)
	default:
		return requested, nil
	}
}

func (p paymentParents) outstanding() int64 {
	remaining := int64(math.MaxInt64)
	if p.invoice != nil {
		remaining = min(remaining, p.invoice.Amount-p.invoice.PaidAmount)
	}
	if p.receivable != nil {
		remaining = min(remaining, p.receivable.Amount-p.receivable.PaidAmount)
	}
	return remaining
}

func (p paymentParents) onchainTargets() OnchainTargets {
	var targets OnchainTargets
	if p.invoice != nil {
		targets.InvoiceID = p.invoice.ID
	}
	if p.receivable != nil {
		targets.ReceivableID = p.receivable.ID
	}
	return targets
}

func ensureNoActiveRequest(ctx context.Context, repository Repository, targets OnchainTargets, now time.Time) error {
	active, err := repository.ActiveOnchainRequestExists(ctx, targets, now)
	if err != nil {
		return apperror.Internal(err)
	}
	if active {
		return errOnchainActive
	}
	return nil
}

func (s *onchainService) newRequest(target onchainTarget, amount int64, actor *uuid.UUID, now time.Time) (OnchainRequest, error) {
	token, err := s.newToken()
	if err != nil {
		return OnchainRequest{}, apperror.Internal(err)
	}
	request := OnchainRequest{
		TargetKind: target.kind,
		TargetID:   target.id,
		AmountIDR:  amount,
		AmountWei:  ethereum.WeiForAmount(amount, s.settings.IDRPerETH).String(),
		IDRPerETH:  s.settings.IDRPerETH,
		Receiver:   s.settings.ReceiverAddress,
		ChainID:    s.settings.ChainID,
		Token:      token,
		Status:     onchainPending,
		CreatedBy:  actor,
		ExpiresAt:  now.Add(s.settings.RequestTTL),
	}
	request.CreatedAt = now
	request.UpdatedAt = now
	return request, nil
}

func (s *onchainService) InvoiceRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[OnchainRequestResponse], error) {
	if _, err := s.repository.FindInvoice(ctx, id); err != nil {
		return pagination.Page[OnchainRequestResponse]{}, invoiceReadErrors.Resolve(err)
	}
	targets := OnchainTargets{InvoiceID: id}
	linked, err := receivableOfInvoice(ctx, s.repository, id)
	if err != nil {
		return pagination.Page[OnchainRequestResponse]{}, err
	}
	if linked != nil {
		targets.ReceivableID = linked.ID
	}
	return s.list(ctx, targets, query)
}

func (s *onchainService) ReceivableRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[OnchainRequestResponse], error) {
	receivable, err := s.repository.FindReceivable(ctx, id)
	if err != nil {
		return pagination.Page[OnchainRequestResponse]{}, receivableReadErrors.Resolve(err)
	}
	targets := OnchainTargets{ReceivableID: id}
	if receivable.InvoiceID != nil {
		targets.InvoiceID = *receivable.InvoiceID
	}
	return s.list(ctx, targets, query)
}

func (s *onchainService) list(ctx context.Context, targets OnchainTargets, query OnchainRequestQuery) (pagination.Page[OnchainRequestResponse], error) {
	rows, total, err := s.repository.ListOnchainRequests(ctx, OnchainRequestFilter{
		Targets: targets,
		Offset:  query.Offset(),
		Limit:   query.Size(),
	})
	if err != nil {
		return pagination.Page[OnchainRequestResponse]{}, apperror.Internal(err)
	}
	now := s.now()
	responses := pagination.Map(rows, func(row OnchainRequestRow) OnchainRequestResponse {
		return newOnchainRequestResponse(row, now, s.settings.ExplorerURL)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *onchainService) CancelRequest(ctx context.Context, id uuid.UUID) (OnchainRequestResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		request, err := repository.LockOnchainRequest(ctx, id)
		if err != nil {
			return onchainReadErrors.Resolve(err)
		}
		if request.Status != onchainPending {
			return errOnchainNotPending
		}
		request.Status = onchainCancelled
		return onchainWriteErrors.Resolve(repository.SaveOnchainRequest(ctx, &request))
	})
	if err != nil {
		return OnchainRequestResponse{}, apperror.From(err)
	}
	return s.requestResponse(ctx, id)
}

func (s *onchainService) PublicPayment(ctx context.Context, token string) (PublicPaymentResponse, error) {
	request, err := s.repository.FindOnchainRequestByToken(ctx, token)
	if err != nil {
		return PublicPaymentResponse{}, onchainReadErrors.Resolve(err)
	}
	return s.publicResponse(ctx, request)
}

func (s *onchainService) ConfirmPayment(ctx context.Context, token string, body OnchainConfirmRequest) (OnchainConfirmResponse, error) {
	hash := strings.ToLower(body.TxHash)
	request, err := s.repository.FindOnchainRequestByToken(ctx, token)
	if err != nil {
		return OnchainConfirmResponse{}, onchainReadErrors.Resolve(err)
	}
	if request.Status != onchainPending {
		return s.finishedOutcome(ctx, request, hash)
	}
	if !s.settings.Enabled() {
		return OnchainConfirmResponse{}, errOnchainDisabled
	}
	if err := s.ensureHashUnused(ctx, hash, request.ID); err != nil {
		return OnchainConfirmResponse{}, err
	}
	now := s.now()
	verifier := transferVerifier{chain: s.chain, minConfirmations: s.settings.MinConfirmations, now: now}
	check, err := verifier.verify(ctx, request, hash)
	if err != nil {
		return OnchainConfirmResponse{}, s.rejected(ctx, request.ID, hash, err)
	}
	if !check.verified {
		return s.confirmResponse(ctx, request, verificationWaiting, check.message, check.confirmations)
	}
	return s.settle(ctx, request.ID, hash, check, now)
}

func (s *onchainService) finishedOutcome(ctx context.Context, request OnchainRequest, hash string) (OnchainConfirmResponse, error) {
	sameTransaction := request.TxHash != nil && *request.TxHash == hash
	switch {
	case request.Status == onchainPaid && sameTransaction:
		return s.confirmResponse(ctx, request, verificationPaid, messagePaid, 0)
	case request.Status == onchainFailed && sameTransaction:
		return s.confirmResponse(ctx, request, verificationFailed, request.FailureReason, 0)
	default:
		return OnchainConfirmResponse{}, finishedError(request)
	}
}

func finishedError(request OnchainRequest) error {
	switch request.Status {
	case onchainPaid:
		return errOnchainPaidByOther
	case onchainFailed:
		return onchainFailedEarlier(request.FailureReason)
	case onchainCancelled:
		return errOnchainCancelled
	default:
		return errOnchainExpired
	}
}

func (s *onchainService) ensureHashUnused(ctx context.Context, hash string, requestID uuid.UUID) error {
	taken, err := s.repository.OnchainTxHashTaken(ctx, hash, requestID)
	if err != nil {
		return apperror.Internal(err)
	}
	if taken {
		return errOnchainTxUsed
	}
	return nil
}

func (s *onchainService) rejected(ctx context.Context, id uuid.UUID, hash string, err error) error {
	if errors.Is(err, errTxAfterExpiry) {
		reason := fmt.Sprintf("transaksi %s masuk setelah batas waktu", hash)
		_ = s.closeRequest(ctx, id, onchainExpired, reason, nil, nil)
	}
	return err
}

func (s *onchainService) settle(ctx context.Context, id uuid.UUID, hash string, check transferCheck, now time.Time) (OnchainConfirmResponse, error) {
	var settled OnchainRequest
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		request, err := repository.LockOnchainRequest(ctx, id)
		if err != nil {
			return onchainReadErrors.Resolve(err)
		}
		settled = request
		if request.Status == onchainPaid && request.TxHash != nil && *request.TxHash == hash {
			return nil
		}
		if request.Status != onchainPending {
			return finishedError(request)
		}
		if err := s.recordPayment(ctx, repository, request, hash, check.payer, now); err != nil {
			return err
		}
		markPaid(&settled, hash, check.payer, now)
		return onchainWriteErrors.Resolve(repository.SaveOnchainRequest(ctx, &settled))
	})
	if err != nil {
		return OnchainConfirmResponse{}, s.settlementFailed(ctx, id, hash, check.payer, err)
	}
	return s.confirmResponse(ctx, settled, verificationPaid, messagePaid, check.confirmations)
}

func (s *onchainService) recordPayment(ctx context.Context, repository Repository, request OnchainRequest, hash, payer string, now time.Time) error {
	parents, err := lockOnchainTarget(ctx, repository, onchainTarget{kind: request.TargetKind, id: request.TargetID})
	if err != nil {
		return err
	}
	payment := PaymentRequest{
		Amount:    request.AmountIDR,
		Method:    onchainMethod,
		Reference: hash,
		Note:      truncateRunes(fmt.Sprintf("Dibayar on-chain dari %s, bukti %s", payer, s.txURL(hash)), paymentNoteLimit),
	}
	return settlePayment(ctx, repository, parents, payment, request.CreatedBy, now)
}

func markPaid(request *OnchainRequest, hash, payer string, now time.Time) {
	request.Status = onchainPaid
	request.TxHash = &hash
	request.PayerAddress = &payer
	request.PaidAt = &now
	request.FailureReason = ""
}

func (s *onchainService) settlementFailed(ctx context.Context, id uuid.UUID, hash, payer string, err error) error {
	appError := apperror.From(err)
	if !slices.Contains(settlementRejections, appError.Code) {
		return appError
	}
	_ = s.closeRequest(ctx, id, onchainFailed, appError.Message, &hash, &payer)
	return onchainSettlementFailed(appError.Message)
}

func (s *onchainService) closeRequest(ctx context.Context, id uuid.UUID, status, reason string, hash, payer *string) error {
	return s.repository.Transaction(ctx, func(repository Repository) error {
		request, err := repository.LockOnchainRequest(ctx, id)
		if err != nil || request.Status != onchainPending {
			return err
		}
		request.Status = status
		request.FailureReason = truncateRunes(reason, failureReasonLimit)
		request.TxHash = hash
		request.PayerAddress = payer
		return repository.SaveOnchainRequest(ctx, &request)
	})
}

func (s *onchainService) requestResponse(ctx context.Context, id uuid.UUID) (OnchainRequestResponse, error) {
	row, err := s.repository.FindOnchainRequestRow(ctx, id)
	if err != nil {
		return OnchainRequestResponse{}, onchainReadErrors.Resolve(err)
	}
	return newOnchainRequestResponse(row, s.now(), s.settings.ExplorerURL), nil
}

func (s *onchainService) publicResponse(ctx context.Context, request OnchainRequest) (PublicPaymentResponse, error) {
	label, err := onchainTargetLabel(ctx, s.repository, request)
	if err != nil {
		return PublicPaymentResponse{}, err
	}
	return newPublicPaymentResponse(request, label, s.now(), s.settings.ExplorerURL), nil
}

func (s *onchainService) confirmResponse(ctx context.Context, request OnchainRequest, verification, message string, confirmations uint64) (OnchainConfirmResponse, error) {
	public, err := s.publicResponse(ctx, request)
	if err != nil {
		return OnchainConfirmResponse{}, err
	}
	return OnchainConfirmResponse{
		PublicPaymentResponse: public,
		Verification:          verification,
		Message:               message,
		Confirmations:         confirmations,
	}, nil
}

func (s *onchainService) txURL(hash string) string {
	return s.settings.ExplorerURL + "/tx/" + hash
}

type targetLabel struct {
	number      string
	projectName *string
}

func onchainTargetLabel(ctx context.Context, repository Repository, request OnchainRequest) (targetLabel, error) {
	if request.TargetKind == onchainInvoiceTarget {
		row, err := repository.FindInvoiceRow(ctx, request.TargetID)
		if err != nil {
			return targetLabel{}, database.ErrorMap{NotFound: errOnchainTargetMissing}.Resolve(err)
		}
		return targetLabel{number: row.Number, projectName: row.ProjectName}, nil
	}
	row, err := repository.FindReceivableRow(ctx, request.TargetID)
	if err != nil {
		return targetLabel{}, database.ErrorMap{NotFound: errOnchainTargetMissing}.Resolve(err)
	}
	return targetLabel{number: row.Reference, projectName: row.ProjectName}, nil
}

func effectiveStatus(request OnchainRequest, now time.Time) string {
	if request.Status == onchainPending && now.After(request.ExpiresAt) {
		return onchainExpired
	}
	return request.Status
}

func paymentURI(request OnchainRequest) string {
	return fmt.Sprintf("ethereum:%s@%d?value=%s", request.Receiver, request.ChainID, request.AmountWei)
}

func transactionURL(explorerURL string, hash *string) *string {
	if hash == nil || explorerURL == "" {
		return nil
	}
	link := explorerURL + "/tx/" + *hash
	return &link
}

func etherAmount(wei string) string {
	value, ok := new(big.Int).SetString(wei, decimalBase)
	if !ok {
		return ""
	}
	return ethereum.FormatEther(value)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
