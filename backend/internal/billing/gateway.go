package billing

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	qrisMethod  = "qris"
	qrisChannel = "QRIS"

	QRISMaxAmount = 10_000_000

	gatewayUnknownReason = "gerbang pembayaran tidak lagi mengenal permintaan ini"
)

var (
	errGatewayDisabled      = apperror.New(http.StatusServiceUnavailable, "gateway_disabled", "Pembayaran QRIS belum diaktifkan di server ini")
	errGatewayUnavailable   = apperror.New(http.StatusBadGateway, "gateway_unavailable", "Gerbang pembayaran tidak bisa dihubungi. Coba lagi sebentar lagi")
	errGatewayNotFound      = apperror.NotFound("gateway_request_not_found", "Permintaan bayar QRIS tidak ditemukan. Periksa lagi tautannya")
	errGatewayVendorInvoice = apperror.Unprocessable("gateway_vendor_invoice", "Pembayaran QRIS hanya untuk faktur ke pelanggan, bukan faktur vendor")
	errGatewayActive        = apperror.Conflict("gateway_request_active", "Masih ada QRIS yang aktif untuk tagihan ini. Batalkan dulu atau tunggu sampai kedaluwarsa")
	errGatewayNotPending    = apperror.Conflict("gateway_request_not_pending", "Permintaan bayar ini sudah tidak menunggu pembayaran")
	errGatewaySimulationOff = apperror.Forbidden("gateway_simulation_disabled", "Simulasi pembayaran hanya tersedia di mode uji")
	errGatewayCallbackToken = apperror.Unauthorized("gateway_callback_rejected", "Token callback tidak cocok")

	gatewayReadErrors  = database.ErrorMap{NotFound: errGatewayNotFound}
	gatewayWriteErrors = database.ErrorMap{NotFound: errGatewayNotFound, Duplicate: errGatewayActive}
)

func gatewayAmountTooLarge(amount int64) *apperror.Error {
	return apperror.Unprocessable("gateway_amount_over_limit", fmt.Sprintf(
		"QRIS dibatasi %s per transaksi, sedangkan nominalnya %s. Tagih sebagian dulu", rupiah(QRISMaxAmount), rupiah(amount),
	))
}

func gatewaySettlementFailed(reason string) *apperror.Error {
	return apperror.Unprocessable("gateway_settlement_failed", fmt.Sprintf(
		"Pembayaran sudah diterima gerbang pembayaran, tetapi tidak bisa dicatat ke tagihan: %s. Hubungi bagian keuangan", reason,
	))
}

type GatewayService interface {
	Config() GatewayConfigResponse
	CreateInvoiceRequest(ctx context.Context, id uuid.UUID, request GatewayRequestCreate, actor *uuid.UUID) (GatewayRequestResponse, error)
	CreateReceivableRequest(ctx context.Context, id uuid.UUID, request GatewayRequestCreate, actor *uuid.UUID) (GatewayRequestResponse, error)
	InvoiceRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[GatewayRequestResponse], error)
	ReceivableRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[GatewayRequestResponse], error)
	CancelRequest(ctx context.Context, id uuid.UUID) (GatewayRequestResponse, error)
	PublicPayment(ctx context.Context, token string) (GatewayPublicResponse, error)
	RefreshPayment(ctx context.Context, token string) (GatewayPublicResponse, error)
	SimulatePayment(ctx context.Context, token string) (GatewayPublicResponse, error)
	HandleWebhook(ctx context.Context, callbackToken string, body GatewayWebhook) error
}

type gatewayService struct {
	repository Repository
	gateway    PaymentGateway
	settings   config.Gateway
	now        func() time.Time
	newToken   func() (string, error)
}

func NewGatewayService(repository Repository, gateway PaymentGateway, settings config.Gateway) GatewayService {
	return &gatewayService{repository: repository, gateway: gateway, settings: settings, now: time.Now, newToken: randomToken}
}

func (s *gatewayService) enabled() bool {
	return s.gateway != nil
}

func (s *gatewayService) Config() GatewayConfigResponse {
	if !s.enabled() {
		return GatewayConfigResponse{Channel: qrisChannel, MaxAmount: QRISMaxAmount}
	}
	return GatewayConfigResponse{
		Enabled:           true,
		Provider:          s.gateway.Provider(),
		TestMode:          s.settings.TestMode(),
		Channel:           qrisChannel,
		MaxAmount:         QRISMaxAmount,
		RequestTTLMinutes: int(s.settings.RequestTTL / time.Minute),
	}
}

func (s *gatewayService) CreateInvoiceRequest(ctx context.Context, id uuid.UUID, request GatewayRequestCreate, actor *uuid.UUID) (GatewayRequestResponse, error) {
	return s.create(ctx, onchainTarget{kind: onchainInvoiceTarget, id: id}, request.Amount, actor)
}

func (s *gatewayService) CreateReceivableRequest(ctx context.Context, id uuid.UUID, request GatewayRequestCreate, actor *uuid.UUID) (GatewayRequestResponse, error) {
	return s.create(ctx, onchainTarget{kind: onchainReceivableTarget, id: id}, request.Amount, actor)
}

func (s *gatewayService) create(ctx context.Context, target onchainTarget, amount int64, actor *uuid.UUID) (GatewayRequestResponse, error) {
	if !s.enabled() {
		return GatewayRequestResponse{}, errGatewayDisabled
	}
	now := s.now()
	var created GatewayRequest
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		parents, err := lockOnchainTarget(ctx, repository, target)
		if err != nil {
			return err
		}
		requested, err := gatewayAmount(parents, amount)
		if err != nil {
			return err
		}
		if err := s.ensureNoActiveRequest(ctx, repository, parents.onchainTargets(), now); err != nil {
			return err
		}
		created, err = s.charge(ctx, target, requested, actor, now)
		if err != nil {
			return err
		}
		return gatewayWriteErrors.Resolve(repository.CreateGatewayRequest(ctx, &created))
	})
	if err != nil {
		return GatewayRequestResponse{}, apperror.From(err)
	}
	return s.requestResponse(ctx, created.ID)
}

func gatewayAmount(parents paymentParents, requested int64) (int64, error) {
	if parents.invoice != nil && parents.invoice.PartyType != customerParty {
		return 0, errGatewayVendorInvoice
	}
	amount, err := requestableAmount(parents, requested)
	if err != nil {
		return 0, err
	}
	if amount > QRISMaxAmount {
		return 0, gatewayAmountTooLarge(amount)
	}
	return amount, nil
}

func (s *gatewayService) ensureNoActiveRequest(ctx context.Context, repository Repository, targets OnchainTargets, now time.Time) error {
	active, err := repository.ActiveGatewayRequestExists(ctx, targets, now)
	if err != nil {
		return apperror.Internal(err)
	}
	if active {
		return errGatewayActive
	}
	return nil
}

func (s *gatewayService) charge(ctx context.Context, target onchainTarget, amount int64, actor *uuid.UUID, now time.Time) (GatewayRequest, error) {
	token, err := s.newToken()
	if err != nil {
		return GatewayRequest{}, apperror.Internal(err)
	}
	id := uuid.New()
	expiresAt := now.Add(s.settings.RequestTTL)
	qr, err := s.gateway.CreateQRIS(ctx, GatewayCharge{
		ReferenceID: id.String(),
		Amount:      amount,
		ExpiresAt:   expiresAt,
		Description: fmt.Sprintf("BanguninAja %s %s", target.kind, id.String()[:8]),
	})
	if err != nil {
		return GatewayRequest{}, errGatewayUnavailable.WithCause(err)
	}
	request := GatewayRequest{
		TargetKind: target.kind,
		TargetID:   target.id,
		AmountIDR:  amount,
		Provider:   s.gateway.Provider(),
		Channel:    qrisChannel,
		ProviderID: qr.ProviderID,
		QRString:   qr.QRString,
		Token:      token,
		Status:     onchainPending,
		CreatedBy:  actor,
		ExpiresAt:  expiresAt,
	}
	request.ID = id
	request.CreatedAt = now
	request.UpdatedAt = now
	return request, nil
}

func (s *gatewayService) InvoiceRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[GatewayRequestResponse], error) {
	if _, err := s.repository.FindInvoice(ctx, id); err != nil {
		return pagination.Page[GatewayRequestResponse]{}, invoiceReadErrors.Resolve(err)
	}
	targets := OnchainTargets{InvoiceID: id}
	linked, err := receivableOfInvoice(ctx, s.repository, id)
	if err != nil {
		return pagination.Page[GatewayRequestResponse]{}, err
	}
	if linked != nil {
		targets.ReceivableID = linked.ID
	}
	return s.list(ctx, targets, query)
}

func (s *gatewayService) ReceivableRequests(ctx context.Context, id uuid.UUID, query OnchainRequestQuery) (pagination.Page[GatewayRequestResponse], error) {
	receivable, err := s.repository.FindReceivable(ctx, id)
	if err != nil {
		return pagination.Page[GatewayRequestResponse]{}, receivableReadErrors.Resolve(err)
	}
	targets := OnchainTargets{ReceivableID: id}
	if receivable.InvoiceID != nil {
		targets.InvoiceID = *receivable.InvoiceID
	}
	return s.list(ctx, targets, query)
}

func (s *gatewayService) list(ctx context.Context, targets OnchainTargets, query OnchainRequestQuery) (pagination.Page[GatewayRequestResponse], error) {
	rows, total, err := s.repository.ListGatewayRequests(ctx, OnchainRequestFilter{Targets: targets, Offset: query.Offset(), Limit: query.Size()})
	if err != nil {
		return pagination.Page[GatewayRequestResponse]{}, apperror.Internal(err)
	}
	now := s.now()
	responses := pagination.Map(rows, func(row GatewayRequestRow) GatewayRequestResponse {
		return newGatewayRequestResponse(row, now)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *gatewayService) CancelRequest(ctx context.Context, id uuid.UUID) (GatewayRequestResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		request, err := repository.LockGatewayRequest(ctx, id)
		if err != nil {
			return gatewayReadErrors.Resolve(err)
		}
		if request.Status != onchainPending {
			return errGatewayNotPending
		}
		if s.enabled() {
			if err := s.gateway.Cancel(ctx, request.ProviderID); err != nil && !errors.Is(err, errGatewayUnknownRequest) {
				return errGatewayUnavailable.WithCause(err)
			}
		}
		request.Status = onchainCancelled
		return gatewayWriteErrors.Resolve(repository.SaveGatewayRequest(ctx, &request))
	})
	if err != nil {
		return GatewayRequestResponse{}, apperror.From(err)
	}
	return s.requestResponse(ctx, id)
}

func (s *gatewayService) PublicPayment(ctx context.Context, token string) (GatewayPublicResponse, error) {
	request, err := s.repository.FindGatewayRequestByToken(ctx, token)
	if err != nil {
		return GatewayPublicResponse{}, gatewayReadErrors.Resolve(err)
	}
	return s.publicResponse(ctx, request)
}

func (s *gatewayService) RefreshPayment(ctx context.Context, token string) (GatewayPublicResponse, error) {
	request, err := s.repository.FindGatewayRequestByToken(ctx, token)
	if err != nil {
		return GatewayPublicResponse{}, gatewayReadErrors.Resolve(err)
	}
	if request.Status == onchainPending && s.enabled() {
		if err := s.sync(ctx, request.ID); err != nil {
			return GatewayPublicResponse{}, err
		}
	}
	return s.PublicPayment(ctx, token)
}

func (s *gatewayService) SimulatePayment(ctx context.Context, token string) (GatewayPublicResponse, error) {
	if !s.enabled() {
		return GatewayPublicResponse{}, errGatewayDisabled
	}
	if !s.settings.TestMode() {
		return GatewayPublicResponse{}, errGatewaySimulationOff
	}
	request, err := s.repository.FindGatewayRequestByToken(ctx, token)
	if err != nil {
		return GatewayPublicResponse{}, gatewayReadErrors.Resolve(err)
	}
	if gatewayEffectiveStatus(request, s.now()) != onchainPending {
		return GatewayPublicResponse{}, errGatewayNotPending
	}
	if err := s.gateway.Simulate(ctx, request.ProviderID, request.AmountIDR); err != nil {
		return GatewayPublicResponse{}, errGatewayUnavailable.WithCause(err)
	}
	return s.RefreshPayment(ctx, token)
}

func (s *gatewayService) HandleWebhook(ctx context.Context, callbackToken string, body GatewayWebhook) error {
	if !s.enabled() {
		return errGatewayDisabled
	}
	expected := s.settings.XenditCallbackToken
	if expected == "" || subtle.ConstantTimeCompare([]byte(callbackToken), []byte(expected)) != 1 {
		return errGatewayCallbackToken
	}
	providerID := body.Data.PaymentRequestID
	if providerID == "" {
		return nil
	}
	request, err := s.repository.FindGatewayRequestByProviderID(ctx, providerID)
	if errors.Is(err, database.ErrNotFound) {
		return nil
	}
	if err != nil {
		return apperror.Internal(err)
	}
	return s.sync(ctx, request.ID)
}

func (s *gatewayService) sync(ctx context.Context, id uuid.UUID) error {
	request, err := s.repository.FindGatewayRequestRow(ctx, id)
	if err != nil {
		return gatewayReadErrors.Resolve(err)
	}
	if request.Status != onchainPending {
		return nil
	}
	status, err := s.gateway.Status(ctx, request.ProviderID)
	switch {
	case errors.Is(err, errGatewayUnknownRequest):
		return s.closeRequest(ctx, id, onchainFailed, gatewayUnknownReason)
	case err != nil:
		return errGatewayUnavailable.WithCause(err)
	}
	switch status.State {
	case gatewaySucceeded:
		return s.settle(ctx, id, status)
	case gatewayExpired:
		return s.closeRequest(ctx, id, onchainExpired, "")
	case gatewayCancelled:
		return s.closeRequest(ctx, id, onchainCancelled, "")
	case gatewayFailed:
		return s.closeRequest(ctx, id, onchainFailed, "pembayaran ditolak oleh bank atau dompet digital")
	default:
		return nil
	}
}

func (s *gatewayService) settle(ctx context.Context, id uuid.UUID, status GatewayStatus) error {
	now := s.now()
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		request, err := repository.LockGatewayRequest(ctx, id)
		if err != nil {
			return gatewayReadErrors.Resolve(err)
		}
		if request.Status != onchainPending {
			return nil
		}
		if status.Amount != request.AmountIDR {
			return apperror.Unprocessable("gateway_amount_mismatch", fmt.Sprintf(
				"nominal di gerbang pembayaran %s tidak sama dengan permintaan %s", rupiah(status.Amount), rupiah(request.AmountIDR),
			))
		}
		parents, err := lockOnchainTarget(ctx, repository, onchainTarget{kind: request.TargetKind, id: request.TargetID})
		if err != nil {
			return err
		}
		payment := PaymentRequest{
			Amount:    request.AmountIDR,
			Method:    qrisMethod,
			Reference: request.ProviderID,
			Note:      truncateRunes(fmt.Sprintf("Dibayar lewat QRIS (%s), dikonfirmasi otomatis oleh gerbang pembayaran", providerName(request.Provider)), paymentNoteLimit),
		}
		if err := settlePayment(ctx, repository, parents, payment, request.CreatedBy, now); err != nil {
			return err
		}
		request.Status = onchainPaid
		request.PaidAt = &now
		request.FailureReason = ""
		return gatewayWriteErrors.Resolve(repository.SaveGatewayRequest(ctx, &request))
	})
	if err == nil {
		return nil
	}
	appError := apperror.From(err)
	if appError.Code != "gateway_amount_mismatch" && !slices.Contains(settlementRejections, appError.Code) {
		return appError
	}
	_ = s.closeRequest(ctx, id, onchainFailed, appError.Message)
	return gatewaySettlementFailed(appError.Message)
}

func (s *gatewayService) closeRequest(ctx context.Context, id uuid.UUID, status, reason string) error {
	return s.repository.Transaction(ctx, func(repository Repository) error {
		request, err := repository.LockGatewayRequest(ctx, id)
		if err != nil || request.Status != onchainPending {
			return err
		}
		request.Status = status
		request.FailureReason = truncateRunes(reason, failureReasonLimit)
		return repository.SaveGatewayRequest(ctx, &request)
	})
}

func (s *gatewayService) requestResponse(ctx context.Context, id uuid.UUID) (GatewayRequestResponse, error) {
	row, err := s.repository.FindGatewayRequestRow(ctx, id)
	if err != nil {
		return GatewayRequestResponse{}, gatewayReadErrors.Resolve(err)
	}
	return newGatewayRequestResponse(row, s.now()), nil
}

func (s *gatewayService) publicResponse(ctx context.Context, request GatewayRequest) (GatewayPublicResponse, error) {
	label, err := onchainTargetLabel(ctx, s.repository, OnchainRequest{TargetKind: request.TargetKind, TargetID: request.TargetID})
	if err != nil {
		return GatewayPublicResponse{}, err
	}
	return newGatewayPublicResponse(request, label, s.now(), s.enabled() && s.settings.TestMode()), nil
}

func providerName(provider string) string {
	switch provider {
	case config.GatewayXendit:
		return "Xendit"
	case config.GatewayFake:
		return "simulator lokal"
	default:
		return provider
	}
}

func gatewayEffectiveStatus(request GatewayRequest, now time.Time) string {
	if request.Status == onchainPending && now.After(request.ExpiresAt) {
		return onchainExpired
	}
	return request.Status
}
