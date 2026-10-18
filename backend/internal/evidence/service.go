package evidence

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const (
	KindPayment = "payment"

	StatusAnchored = "anchored"
	StatusPending  = "pending"
	StatusChanged  = "changed"
	StatusDeleted  = "deleted"

	anchorConfirmed = "anchored"
	anchorWaiting   = "waiting"

	claimTokenBytes  = 24
	successfulStatus = 1
)

var (
	errNothingToAnchor = apperror.Unprocessable("evidence_nothing_new", "Semua bukti sudah tercatat. Belum ada bukti baru atau yang berubah")
	errBatchNotFound   = apperror.NotFound("evidence_batch_not_found", "Kumpulan bukti tidak ditemukan")
	errBatchAnchored   = apperror.Conflict("evidence_batch_anchored", "Kumpulan bukti ini sudah tercatat di blockchain")
	errBadTxHash       = apperror.BadRequest("evidence_tx_hash", "Hash transaksi harus 0x diikuti 64 karakter heksadesimal")
	errTxWrongData     = apperror.Unprocessable("evidence_tx_wrong_data", "Data transaksi ini bukan sidik jari kumpulan bukti yang diminta")
	errTxWrongChain    = apperror.Unprocessable("evidence_tx_wrong_chain", "Transaksi ini dari jaringan lain")
	errTxReverted      = apperror.Unprocessable("evidence_tx_reverted", "Transaksi ini gagal di jaringan (reverted)")
	errChainDown       = apperror.New(http.StatusBadGateway, "chain_unavailable", "Node jaringan blockchain tidak bisa dihubungi. Coba lagi sebentar")
	errApproverAddress = apperror.Unprocessable("evidence_approver_address", "Alamat dompet harus 0x diikuti 40 karakter heksadesimal, salin langsung dari dompet")
	errClaimNotFound   = apperror.NotFound("evidence_claim_not_found", "Tautan verifikasi tidak ditemukan. Periksa lagi tautannya")
	errNoApprover      = apperror.Unprocessable("evidence_no_approver", "Pengembang belum mendaftarkan alamat dompet pemberi kerja untuk proyek ini")
	errClaimApproved   = apperror.Conflict("evidence_claim_approved", "Tagihan termin ini sudah disetujui")
	errBadSignature    = apperror.Unprocessable("evidence_bad_signature", "Tanda tangan tidak valid")
	errWrongSigner     = apperror.Unprocessable("evidence_wrong_signer", "Tanda tangan bukan dari dompet pemberi kerja yang terdaftar")
)

type Fact struct {
	Kind      string
	SubjectID uuid.UUID
	ProjectID *uuid.UUID
	Date      time.Time
	Label     string
	Fields    map[string]any
}

type Claim struct {
	TermID          uuid.UUID
	ProjectID       uuid.UUID
	ProjectCode     string
	ProjectName     string
	ContractNumber  string
	OwnerName       string
	Label           string
	Progress        int
	GrossAmount     int64
	RetentionAmount int64
	NetAmount       int64
	ClaimedOn       time.Time
}

type Sources interface {
	ProjectFacts(ctx context.Context, projectID *uuid.UUID) ([]Fact, error)
	PaymentFacts(ctx context.Context, projectID *uuid.UUID) ([]Fact, error)
	OwnerClaim(ctx context.Context, projectID, termID uuid.UUID) (Claim, error)
}

type ChainReader interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionByHash(ctx context.Context, hash string) (ethereum.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, hash string) (ethereum.Receipt, bool, error)
	BlockTime(ctx context.Context, number uint64) (time.Time, error)
}

type Service struct {
	repository Repository
	sources    Sources
	chain      ChainReader
	config     config.Onchain
	now        func() time.Time
}

func NewService(repository Repository, sources Sources, chain ChainReader, onchain config.Onchain) *Service {
	return &Service{repository: repository, sources: sources, chain: chain, config: onchain, now: time.Now}
}

type Item struct {
	Kind       string     `json:"kind"`
	SubjectID  uuid.UUID  `json:"subjectId"`
	Label      string     `json:"label"`
	Date       *time.Time `json:"date"`
	Status     string     `json:"status"`
	AnchoredAt *time.Time `json:"anchoredAt"`
	TxHash     string     `json:"txHash"`
	TxURL      string     `json:"txUrl"`
}

type Summary struct {
	Anchored int `json:"anchored"`
	Pending  int `json:"pending"`
	Changed  int `json:"changed"`
	Deleted  int `json:"deleted"`
}

type ClaimStatus struct {
	TermID     uuid.UUID  `json:"termId"`
	Token      string     `json:"token"`
	Signer     string     `json:"signer"`
	ApprovedAt *time.Time `json:"approvedAt"`
	StillSame  bool       `json:"stillSame"`
}

type ProjectEvidence struct {
	Items           []Item        `json:"items"`
	Summary         Summary       `json:"summary"`
	EvidenceRoot    string        `json:"evidenceRoot"`
	ApproverAddress string        `json:"approverAddress"`
	Claims          []ClaimStatus `json:"claims"`
	ChainID         int64         `json:"chainId"`
}

type BatchResponse struct {
	ID          uuid.UUID  `json:"id"`
	Root        string     `json:"root"`
	Calldata    string     `json:"calldata"`
	RecordCount int        `json:"recordCount"`
	ChainID     int64      `json:"chainId"`
	TxHash      string     `json:"txHash"`
	TxURL       string     `json:"txUrl"`
	AnchoredAt  *time.Time `json:"anchoredAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type AnchorResult struct {
	Status        string        `json:"status"`
	Message       string        `json:"message"`
	Confirmations uint64        `json:"confirmations"`
	Batch         BatchResponse `json:"batch"`
}

type AnchorRequest struct {
	TxHash string `json:"txHash" binding:"required"`
}

type ApproverRequest struct {
	Address string `json:"address" binding:"required"`
}

type ApproveRequest struct {
	Signature string `json:"signature" binding:"required"`
}

type PublicClaim struct {
	ProjectCode     string      `json:"projectCode"`
	ProjectName     string      `json:"projectName"`
	ContractNumber  string      `json:"contractNumber"`
	OwnerName       string      `json:"ownerName"`
	Label           string      `json:"label"`
	Progress        int         `json:"progress"`
	GrossAmount     int64       `json:"grossAmount"`
	RetentionAmount int64       `json:"retentionAmount"`
	NetAmount       int64       `json:"netAmount"`
	ClaimedOn       time.Time   `json:"claimedOn"`
	Items           []Item      `json:"items"`
	Summary         Summary     `json:"summary"`
	TypedData       TypedData   `json:"typedData"`
	ApproverAddress string      `json:"approverAddress"`
	Approval        ClaimStatus `json:"approval"`
	ExplorerURL     string      `json:"explorerUrl"`
}

func (s *Service) facts(ctx context.Context, projectID *uuid.UUID) ([]Fact, error) {
	projectFacts, err := s.sources.ProjectFacts(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	payments, err := s.sources.PaymentFacts(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return append(projectFacts, payments...), nil
}

func (s *Service) txURL(hash string) string {
	if hash == "" {
		return ""
	}
	return strings.TrimRight(s.config.ExplorerURL, "/") + "/tx/" + hash
}

func (s *Service) evaluate(ctx context.Context, projectID *uuid.UUID) ([]Item, Summary, string, error) {
	facts, err := s.facts(ctx, projectID)
	if err != nil {
		return nil, Summary{}, "", err
	}
	records, err := s.repository.RecordsFor(ctx, projectID)
	if err != nil {
		return nil, Summary{}, "", apperror.Internal(err)
	}
	batchIDs := []uuid.UUID{}
	for _, record := range records {
		if record.BatchID != nil {
			batchIDs = append(batchIDs, *record.BatchID)
		}
	}
	anchored, err := s.repository.AnchoredBatches(ctx, batchIDs)
	if err != nil {
		return nil, Summary{}, "", apperror.Internal(err)
	}

	type proof struct {
		digest string
		batch  Batch
	}
	latest := map[string]proof{}
	for _, record := range records {
		if record.BatchID == nil {
			continue
		}
		batch, ok := anchored[*record.BatchID]
		if !ok {
			continue
		}
		key := record.Kind + ":" + record.SubjectID.String()
		if current, seen := latest[key]; !seen || batch.AnchoredAt.After(*current.batch.AnchoredAt) {
			latest[key] = proof{digest: record.Digest, batch: batch}
		}
	}

	items := make([]Item, 0, len(facts))
	summary := Summary{}
	leaves := make([]leaf, 0, len(facts))
	present := map[string]bool{}
	for _, fact := range facts {
		digest := Digest(fact.Kind, fact.Fields)
		key := fact.Kind + ":" + fact.SubjectID.String()
		present[key] = true
		leaves = append(leaves, leaf{kind: fact.Kind, subject: fact.SubjectID.String(), digest: digest})
		date := fact.Date
		item := Item{Kind: fact.Kind, SubjectID: fact.SubjectID, Label: fact.Label, Date: &date, Status: StatusPending}
		if found, ok := latest[key]; ok {
			item.AnchoredAt = found.batch.AnchoredAt
			item.TxHash = *found.batch.TxHash
			item.TxURL = s.txURL(item.TxHash)
			item.Status = StatusChanged
			if found.digest == digest {
				item.Status = StatusAnchored
			}
		}
		items = append(items, item)
		count(&summary, item.Status)
	}
	for key, found := range latest {
		if present[key] {
			continue
		}
		kind, subject, _ := strings.Cut(key, ":")
		subjectID, _ := uuid.Parse(subject)
		item := Item{Kind: kind, SubjectID: subjectID, Status: StatusDeleted, AnchoredAt: found.batch.AnchoredAt,
			TxHash: *found.batch.TxHash, TxURL: s.txURL(*found.batch.TxHash)}
		items = append(items, item)
		count(&summary, item.Status)
	}
	sort.SliceStable(items, func(a, b int) bool {
		left, right := items[a].Date, items[b].Date
		if left == nil || right == nil {
			return left != nil
		}
		return left.After(*right)
	})
	return items, summary, MerkleRoot(leaves), nil
}

func count(summary *Summary, status string) {
	switch status {
	case StatusAnchored:
		summary.Anchored++
	case StatusChanged:
		summary.Changed++
	case StatusDeleted:
		summary.Deleted++
	default:
		summary.Pending++
	}
}

func (s *Service) ProjectEvidence(ctx context.Context, projectID uuid.UUID) (ProjectEvidence, error) {
	items, summary, root, err := s.evaluate(ctx, &projectID)
	if err != nil {
		return ProjectEvidence{}, err
	}
	approver, _, err := s.repository.FindApprover(ctx, projectID)
	if err != nil {
		return ProjectEvidence{}, apperror.Internal(err)
	}
	claims, err := s.repository.ListClaims(ctx, projectID)
	if err != nil {
		return ProjectEvidence{}, apperror.Internal(err)
	}
	statuses := make([]ClaimStatus, len(claims))
	for index, claim := range claims {
		statuses[index] = claimStatus(claim, root)
	}
	return ProjectEvidence{
		Items: items, Summary: summary, EvidenceRoot: root, ApproverAddress: approver.Address,
		Claims: statuses, ChainID: s.config.ChainID,
	}, nil
}

func claimStatus(claim ClaimLink, currentRoot string) ClaimStatus {
	return ClaimStatus{
		TermID: claim.TermID, Token: claim.Token, Signer: claim.Signer, ApprovedAt: claim.ApprovedAt,
		StillSame: claim.ApprovedAt != nil && claim.EvidenceRoot == currentRoot,
	}
}

func (s *Service) CreateBatch(ctx context.Context, actorID uuid.UUID) (BatchResponse, error) {
	facts, err := s.facts(ctx, nil)
	if err != nil {
		return BatchResponse{}, err
	}
	var created Batch
	err = s.repository.Transaction(ctx, func(repository Repository) error {
		if pending, found, err := repository.PendingBatch(ctx); err != nil {
			return apperror.Internal(err)
		} else if found {
			if err := repository.ReleaseBatch(ctx, pending.ID); err != nil {
				return apperror.Internal(err)
			}
		}
		for _, fact := range facts {
			digest := Digest(fact.Kind, fact.Fields)
			exists, err := repository.RecordExists(ctx, fact.Kind, fact.SubjectID, digest)
			if err != nil {
				return apperror.Internal(err)
			}
			if exists {
				continue
			}
			record := Record{Kind: fact.Kind, SubjectID: fact.SubjectID, Digest: digest, ProjectID: fact.ProjectID}
			if err := repository.CreateRecord(ctx, &record); err != nil {
				return apperror.Internal(err)
			}
		}
		records, err := repository.UnbatchedRecords(ctx)
		if err != nil {
			return apperror.Internal(err)
		}
		if len(records) == 0 {
			return errNothingToAnchor
		}
		leaves := make([]leaf, len(records))
		for index, record := range records {
			leaves[index] = leaf{kind: record.Kind, subject: record.SubjectID.String(), digest: record.Digest}
		}
		created = Batch{Root: MerkleRoot(leaves), RecordCount: len(records), ChainID: s.config.ChainID, CreatedBy: &actorID}
		return repository.CreateBatch(ctx, &created, records)
	})
	if err != nil {
		return BatchResponse{}, apperror.From(err)
	}
	return s.batchResponse(created), nil
}

func (s *Service) batchResponse(batch Batch) BatchResponse {
	response := BatchResponse{
		ID: batch.ID, Root: batch.Root, Calldata: Calldata(batch.Root), RecordCount: batch.RecordCount,
		ChainID: batch.ChainID, AnchoredAt: batch.AnchoredAt, CreatedAt: batch.CreatedAt,
	}
	if batch.TxHash != nil {
		response.TxHash = *batch.TxHash
		response.TxURL = s.txURL(*batch.TxHash)
	}
	return response
}

func (s *Service) ListBatches(ctx context.Context) ([]BatchResponse, error) {
	batches, err := s.repository.ListBatches(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	responses := make([]BatchResponse, len(batches))
	for index, batch := range batches {
		responses[index] = s.batchResponse(batch)
	}
	return responses, nil
}

func (s *Service) Anchor(ctx context.Context, batchID uuid.UUID, txHash string) (AnchorResult, error) {
	hash := strings.ToLower(strings.TrimSpace(txHash))
	if !ethereum.IsTransactionHash(hash) {
		return AnchorResult{}, errBadTxHash
	}
	batch, err := s.repository.FindBatch(ctx, batchID)
	if err != nil {
		return AnchorResult{}, database.ErrorMap{NotFound: errBatchNotFound}.Resolve(err)
	}
	if batch.TxHash != nil {
		if *batch.TxHash == hash {
			return AnchorResult{Status: anchorConfirmed, Message: "Bukti sudah tercatat di blockchain", Batch: s.batchResponse(batch)}, nil
		}
		return AnchorResult{}, errBatchAnchored
	}

	transaction, found, err := s.chain.TransactionByHash(ctx, hash)
	if err != nil {
		return AnchorResult{}, errChainDown.WithCause(err)
	}
	if !found {
		return s.waiting(batch, "Transaksi belum terlihat di jaringan. Kalau baru saja dikirim, tunggu beberapa detik", 0), nil
	}
	if transaction.ChainID != nil && transaction.ChainID.Cmp(big.NewInt(batch.ChainID)) != 0 {
		return AnchorResult{}, errTxWrongChain
	}
	if transaction.Input != Calldata(batch.Root) {
		return AnchorResult{}, errTxWrongData
	}
	receipt, found, err := s.chain.TransactionReceipt(ctx, hash)
	if err != nil {
		return AnchorResult{}, errChainDown.WithCause(err)
	}
	if !found {
		return s.waiting(batch, "Transaksi sudah terkirim dan sedang menunggu masuk blok", 0), nil
	}
	if receipt.Status != successfulStatus {
		return AnchorResult{}, errTxReverted
	}
	head, err := s.chain.BlockNumber(ctx)
	if err != nil {
		return AnchorResult{}, errChainDown.WithCause(err)
	}
	confirmations := uint64(0)
	if head >= receipt.BlockNumber {
		confirmations = head - receipt.BlockNumber + 1
	}
	if confirmations < s.config.MinConfirmations {
		return s.waiting(batch, fmt.Sprintf("Transaksi sudah masuk blok, menunggu konfirmasi (%d dari %d)", confirmations, s.config.MinConfirmations), confirmations), nil
	}
	blockTime, err := s.chain.BlockTime(ctx, receipt.BlockNumber)
	if err != nil {
		return AnchorResult{}, errChainDown.WithCause(err)
	}

	block := int64(receipt.BlockNumber)
	batch.TxHash = &hash
	batch.AnchoredFrom = transaction.From
	batch.BlockNumber = &block
	batch.AnchoredAt = &blockTime
	if err := s.repository.SaveBatch(ctx, &batch); err != nil {
		return AnchorResult{}, apperror.From(database.ErrorMap{Duplicate: errBatchAnchored}.Resolve(err))
	}
	return AnchorResult{Status: anchorConfirmed, Message: "Bukti tercatat di blockchain", Confirmations: confirmations, Batch: s.batchResponse(batch)}, nil
}

func (s *Service) waiting(batch Batch, message string, confirmations uint64) AnchorResult {
	return AnchorResult{Status: anchorWaiting, Message: message, Confirmations: confirmations, Batch: s.batchResponse(batch)}
}

func (s *Service) SetApprover(ctx context.Context, projectID uuid.UUID, address string) (string, error) {
	checksummed, err := ethereum.ChecksumAddress(address)
	if err != nil {
		return "", errApproverAddress
	}
	approver, _, err := s.repository.FindApprover(ctx, projectID)
	if err != nil {
		return "", apperror.Internal(err)
	}
	approver.ProjectID = projectID
	approver.Address = checksummed
	if err := s.repository.SaveApprover(ctx, &approver); err != nil {
		return "", apperror.From(database.ErrorMap{Referenced: errClaimNotFound}.Resolve(err))
	}
	return checksummed, nil
}

func (s *Service) CreateClaimLink(ctx context.Context, projectID, termID uuid.UUID) (ClaimStatus, error) {
	if _, err := s.sources.OwnerClaim(ctx, projectID, termID); err != nil {
		return ClaimStatus{}, apperror.From(err)
	}
	claim, found, err := s.repository.FindClaimByTerm(ctx, termID)
	if err != nil {
		return ClaimStatus{}, apperror.Internal(err)
	}
	if !found {
		token, err := newToken()
		if err != nil {
			return ClaimStatus{}, apperror.Internal(err)
		}
		claim = ClaimLink{ProjectID: projectID, TermID: termID, Token: token}
		if err := s.repository.SaveClaim(ctx, &claim); err != nil {
			return ClaimStatus{}, apperror.Internal(err)
		}
	}
	return claimStatus(claim, claim.EvidenceRoot), nil
}

func newToken() (string, error) {
	raw := make([]byte, claimTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *Service) loadClaim(ctx context.Context, token string) (ClaimLink, Claim, error) {
	link, err := s.repository.FindClaimByToken(ctx, token)
	if err != nil {
		return ClaimLink{}, Claim{}, database.ErrorMap{NotFound: errClaimNotFound}.Resolve(err)
	}
	claim, err := s.sources.OwnerClaim(ctx, link.ProjectID, link.TermID)
	if err != nil {
		return ClaimLink{}, Claim{}, errClaimNotFound
	}
	return link, claim, nil
}

func approvalMessage(claim Claim, root string) ApprovalMessage {
	return ApprovalMessage{
		Project: claim.ProjectCode + " " + claim.ProjectName, Contract: claim.ContractNumber, Term: claim.Label,
		Progress: strconv.Itoa(claim.Progress), Amount: strconv.FormatInt(claim.NetAmount, 10),
		ClaimedOn: claim.ClaimedOn.Format(time.DateOnly), EvidenceRoot: root,
	}
}

func (s *Service) PublicClaim(ctx context.Context, token string) (PublicClaim, error) {
	link, claim, err := s.loadClaim(ctx, token)
	if err != nil {
		return PublicClaim{}, err
	}
	items, summary, root, err := s.evaluate(ctx, &link.ProjectID)
	if err != nil {
		return PublicClaim{}, err
	}
	approver, _, err := s.repository.FindApprover(ctx, link.ProjectID)
	if err != nil {
		return PublicClaim{}, apperror.Internal(err)
	}
	return PublicClaim{
		ProjectCode: claim.ProjectCode, ProjectName: claim.ProjectName, ContractNumber: claim.ContractNumber,
		OwnerName: claim.OwnerName, Label: claim.Label, Progress: claim.Progress, GrossAmount: claim.GrossAmount,
		RetentionAmount: claim.RetentionAmount, NetAmount: claim.NetAmount, ClaimedOn: claim.ClaimedOn,
		Items: items, Summary: summary, TypedData: NewTypedData(s.config.ChainID, approvalMessage(claim, root)),
		ApproverAddress: approver.Address, Approval: claimStatus(link, root), ExplorerURL: s.config.ExplorerURL,
	}, nil
}

func (s *Service) Approve(ctx context.Context, token, signature string) (ClaimStatus, error) {
	link, claim, err := s.loadClaim(ctx, token)
	if err != nil {
		return ClaimStatus{}, err
	}
	if link.ApprovedAt != nil {
		return ClaimStatus{}, errClaimApproved
	}
	approver, found, err := s.repository.FindApprover(ctx, link.ProjectID)
	if err != nil {
		return ClaimStatus{}, apperror.Internal(err)
	}
	if !found {
		return ClaimStatus{}, errNoApprover
	}
	_, _, root, err := s.evaluate(ctx, &link.ProjectID)
	if err != nil {
		return ClaimStatus{}, err
	}
	signer, err := ethereum.RecoverAddress(ApprovalDigest(s.config.ChainID, approvalMessage(claim, root)), signature)
	if err != nil {
		return ClaimStatus{}, errBadSignature
	}
	if !ethereum.SameAddress(signer, approver.Address) {
		return ClaimStatus{}, errWrongSigner
	}
	approvedAt := s.now()
	link.Signature = signature
	link.Signer = signer
	link.EvidenceRoot = root
	link.ApprovedAt = &approvedAt
	if err := s.repository.SaveClaim(ctx, &link); err != nil {
		return ClaimStatus{}, apperror.Internal(err)
	}
	return claimStatus(link, root), nil
}

func (s *Service) SubjectStatus(ctx context.Context, projectID uuid.UUID, kind string, subjectID uuid.UUID) (Item, error) {
	items, _, _, err := s.evaluate(ctx, &projectID)
	if err != nil {
		return Item{}, err
	}
	for _, item := range items {
		if item.Kind == kind && item.SubjectID == subjectID {
			return item, nil
		}
	}
	return Item{Kind: kind, SubjectID: subjectID, Status: StatusPending}, nil
}
