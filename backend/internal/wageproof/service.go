package wageproof

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	KindWageRoot = "wage_root"
	tokenBytes   = 24
)

var (
	periodPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

	errPeriod        = apperror.BadRequest("wage_period", "Periode harus berformat YYYY-MM")
	errNoRegion      = apperror.Unprocessable("wage_no_minimum", "Upah minimum untuk wilayah dan tahun proyek ini belum tersedia")
	errNoPayroll     = apperror.Unprocessable("wage_no_payroll", "Belum ada penggajian yang dibayar untuk proyek dan periode ini")
	errTooMany       = apperror.Unprocessable("wage_too_many", fmt.Sprintf("Prototipe ini menangani paling banyak %d pekerja per bukti", MaxWorkers))
	errProofNotFound = apperror.NotFound("wage_proof_not_found", "Bukti upah tidak ditemukan. Periksa lagi tautannya")
	errReceiptAbsent = apperror.NotFound("wage_receipt_not_found", "Kuitansi tidak ditemukan. Periksa lagi tautannya")
)

type WageSource interface {
	PaidWages(ctx context.Context, projectID uuid.UUID, period string) ([]WageRowInput, error)
	Place(ctx context.Context, projectID uuid.UUID) (Place, error)
}

type WageRowInput struct {
	EmployeeID uuid.UUID
	Name       string
	Gross      int64
	Days       int
}

type Place struct {
	Code       string
	Name       string
	RegionCode string
	RegionName string
}

type AnchorStatus struct {
	Status     string     `json:"status"`
	AnchoredAt *time.Time `json:"anchoredAt"`
	TxHash     string     `json:"txHash"`
	TxURL      string     `json:"txUrl"`
}

type Anchors interface {
	Status(ctx context.Context, projectID uuid.UUID, kind string, subjectID uuid.UUID) (AnchorStatus, error)
}

type Service struct {
	repository Repository
	wages      WageSource
	prover     *Prover
	anchors    Anchors
}

func NewService(repository Repository, wages WageSource, anchors Anchors) *Service {
	service := &Service{repository: repository, wages: wages, anchors: anchors}
	service.prover = NewProver(
		func() (Keys, bool, error) { return repository.LoadKeys(context.Background()) },
		func(keys Keys) error { return repository.StoreKeys(context.Background(), keys) },
	)
	return service
}

type BelowMinimum struct {
	Name     string `json:"name"`
	Pay      int64  `json:"pay"`
	Days     int    `json:"days"`
	Required int64  `json:"required"`
}

type CreateRequest struct {
	Period string `json:"period" binding:"required"`
}

type CreateResult struct {
	Proof *ProofResponse `json:"proof"`
	Below []BelowMinimum `json:"below"`
}

type ReceiptLink struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type ProofResponse struct {
	ID          uuid.UUID     `json:"id"`
	Period      string        `json:"period"`
	RegionName  string        `json:"regionName"`
	MinimumWage int64         `json:"minimumWage"`
	WageSource  string        `json:"wageSource"`
	WorkerCount int           `json:"workerCount"`
	Root        string        `json:"root"`
	Token       string        `json:"token"`
	Receipts    []ReceiptLink `json:"receipts"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type PublicProof struct {
	ProjectCode     string       `json:"projectCode"`
	ProjectName     string       `json:"projectName"`
	Period          string       `json:"period"`
	RegionName      string       `json:"regionName"`
	MinimumWage     int64        `json:"minimumWage"`
	DailyMinimum    int64        `json:"dailyMinimum"`
	WageSource      string       `json:"wageSource"`
	WorkerCount     int          `json:"workerCount"`
	Root            string       `json:"root"`
	Verified        bool         `json:"verified"`
	Anchor          AnchorStatus `json:"anchor"`
	ProofHex        string       `json:"proofHex"`
	VerifyingKeyHex string       `json:"verifyingKeyHex"`
}

type PublicReceipt struct {
	Name        string       `json:"name"`
	ProjectCode string       `json:"projectCode"`
	ProjectName string       `json:"projectName"`
	Period      string       `json:"period"`
	Pay         int64        `json:"pay"`
	Days        int          `json:"days"`
	LeafIndex   int          `json:"leafIndex"`
	Salt        string       `json:"salt"`
	Path        []string     `json:"path"`
	Root        string       `json:"root"`
	Included    bool         `json:"included"`
	ProofToken  string       `json:"proofToken"`
	Anchor      AnchorStatus `json:"anchor"`
}

func newToken() (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *Service) Create(ctx context.Context, projectID, actorID uuid.UUID, period string) (CreateResult, error) {
	if !periodPattern.MatchString(period) {
		return CreateResult{}, errPeriod
	}
	place, err := s.wages.Place(ctx, projectID)
	if err != nil {
		return CreateResult{}, apperror.From(err)
	}
	year, _ := strconv.Atoi(period[:4])
	minimum, ok := MinimumWageFor(place.RegionCode, year)
	if !ok {
		return CreateResult{}, errNoRegion
	}
	rows, err := s.wages.PaidWages(ctx, projectID, period)
	if err != nil {
		return CreateResult{}, apperror.Internal(err)
	}
	if len(rows) == 0 {
		return CreateResult{}, errNoPayroll
	}
	if len(rows) > MaxWorkers {
		return CreateResult{}, errTooMany
	}

	leaves := make([]Leaf, len(rows))
	below := []BelowMinimum{}
	for index, row := range rows {
		salt, err := NewSalt()
		if err != nil {
			return CreateResult{}, apperror.Internal(err)
		}
		leaves[index] = Leaf{Worker: WorkerKey(row.EmployeeID), Pay: row.Gross, Days: row.Days, Salt: salt}
		if !Meets(leaves[index], minimum.Amount) {
			required := (minimum.Amount*int64(row.Days) + workDaysBase - 1) / workDaysBase
			below = append(below, BelowMinimum{Name: row.Name, Pay: row.Gross, Days: row.Days, Required: required})
		}
	}
	if len(below) > 0 {
		return CreateResult{Below: below}, nil
	}

	root, encoded, err := s.prover.Prove(leaves, minimum.Amount)
	if err != nil {
		return CreateResult{}, apperror.Internal(err)
	}
	token, err := newToken()
	if err != nil {
		return CreateResult{}, apperror.Internal(err)
	}
	proof := Proof{
		ProjectID: projectID, Period: period, RegionCode: place.RegionCode, RegionName: place.RegionName,
		MinimumWage: minimum.Amount, WageSource: minimum.Source, WorkerCount: len(rows), Root: Hex(root),
		Proof: encoded, Token: token, CreatedBy: &actorID,
	}
	receipts := make([]Receipt, len(rows))
	for index, row := range rows {
		receiptToken, err := newToken()
		if err != nil {
			return CreateResult{}, apperror.Internal(err)
		}
		receipts[index] = Receipt{EmployeeID: row.EmployeeID, Name: row.Name, LeafIndex: index, Pay: row.Gross,
			Days: row.Days, Salt: Hex(leaves[index].Salt), Token: receiptToken}
	}
	if err := s.repository.ReplaceProof(ctx, &proof, receipts); err != nil {
		return CreateResult{}, apperror.Internal(err)
	}
	response := proofResponse(proof, receipts)
	return CreateResult{Proof: &response, Below: below}, nil
}

func proofResponse(proof Proof, receipts []Receipt) ProofResponse {
	links := []ReceiptLink{}
	for _, receipt := range receipts {
		if receipt.ProofID == proof.ID {
			links = append(links, ReceiptLink{Name: receipt.Name, Token: receipt.Token})
		}
	}
	return ProofResponse{
		ID: proof.ID, Period: proof.Period, RegionName: proof.RegionName, MinimumWage: proof.MinimumWage,
		WageSource: proof.WageSource, WorkerCount: proof.WorkerCount, Root: proof.Root, Token: proof.Token,
		Receipts: links, CreatedAt: proof.CreatedAt,
	}
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID) ([]ProofResponse, error) {
	proofs, err := s.repository.ListProofs(ctx, &projectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	ids := make([]uuid.UUID, len(proofs))
	for index, proof := range proofs {
		ids[index] = proof.ID
	}
	receipts, err := s.repository.ReceiptsFor(ctx, ids)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	responses := make([]ProofResponse, len(proofs))
	for index, proof := range proofs {
		responses[index] = proofResponse(proof, receipts)
	}
	return responses, nil
}

func (s *Service) anchor(ctx context.Context, proof Proof) AnchorStatus {
	if s.anchors == nil {
		return AnchorStatus{Status: "pending"}
	}
	status, err := s.anchors.Status(ctx, proof.ProjectID, KindWageRoot, proof.ID)
	if err != nil {
		return AnchorStatus{Status: "pending"}
	}
	return status
}

func (s *Service) PublicProof(ctx context.Context, token string) (PublicProof, error) {
	proof, err := s.repository.FindProofByToken(ctx, token)
	if err != nil {
		return PublicProof{}, database.ErrorMap{NotFound: errProofNotFound}.Resolve(err)
	}
	place, err := s.wages.Place(ctx, proof.ProjectID)
	if err != nil {
		return PublicProof{}, apperror.From(err)
	}
	root, ok := ParseHex(proof.Root)
	verified := false
	if ok {
		verified, err = s.prover.Verify(proof.Proof, root, proof.MinimumWage, proof.WorkerCount)
		if err != nil {
			return PublicProof{}, apperror.Internal(err)
		}
	}
	verifyingKey, err := s.prover.VerifyingKey()
	if err != nil {
		return PublicProof{}, apperror.Internal(err)
	}
	return PublicProof{
		ProjectCode: place.Code, ProjectName: place.Name, Period: proof.Period, RegionName: proof.RegionName,
		MinimumWage: proof.MinimumWage, DailyMinimum: (proof.MinimumWage + workDaysBase - 1) / workDaysBase,
		WageSource: proof.WageSource, WorkerCount: proof.WorkerCount, Root: proof.Root, Verified: verified,
		Anchor: s.anchor(ctx, proof), ProofHex: fmt.Sprintf("0x%x", proof.Proof), VerifyingKeyHex: fmt.Sprintf("0x%x", verifyingKey),
	}, nil
}

func (s *Service) PublicReceipt(ctx context.Context, token string) (PublicReceipt, error) {
	receipt, err := s.repository.FindReceiptByToken(ctx, token)
	if err != nil {
		return PublicReceipt{}, database.ErrorMap{NotFound: errReceiptAbsent}.Resolve(err)
	}
	proof, err := s.repository.FindProof(ctx, receipt.ProofID)
	if err != nil {
		return PublicReceipt{}, errReceiptAbsent
	}
	place, err := s.wages.Place(ctx, proof.ProjectID)
	if err != nil {
		return PublicReceipt{}, apperror.From(err)
	}
	receipts, err := s.repository.ReceiptsFor(ctx, []uuid.UUID{proof.ID})
	if err != nil {
		return PublicReceipt{}, apperror.Internal(err)
	}
	leaves := make([]Leaf, len(receipts))
	for index, item := range receipts {
		salt, _ := ParseHex(item.Salt)
		leaves[index] = Leaf{Worker: WorkerKey(item.EmployeeID), Pay: item.Pay, Days: item.Days, Salt: salt}
	}
	tree, err := BuildTree(leaves)
	if err != nil {
		return PublicReceipt{}, apperror.Internal(err)
	}
	path := tree.Path(receipt.LeafIndex)
	hexPath := make([]string, len(path))
	for index, node := range path {
		hexPath[index] = Hex(node)
	}
	root, _ := ParseHex(proof.Root)
	salt, _ := ParseHex(receipt.Salt)
	own := Leaf{Worker: WorkerKey(receipt.EmployeeID), Pay: receipt.Pay, Days: receipt.Days, Salt: salt}
	return PublicReceipt{
		Name: receipt.Name, ProjectCode: place.Code, ProjectName: place.Name, Period: proof.Period, Pay: receipt.Pay,
		Days: receipt.Days, LeafIndex: receipt.LeafIndex, Salt: receipt.Salt, Path: hexPath, Root: proof.Root,
		Included: root != nil && VerifyPath(own, receipt.LeafIndex, path, root), ProofToken: proof.Token,
		Anchor: s.anchor(ctx, proof),
	}, nil
}

type RootFeed struct {
	repository Repository
}

func NewRootFeed(repository Repository) RootFeed {
	return RootFeed{repository: repository}
}

type RootFact struct {
	ProofID   uuid.UUID
	ProjectID uuid.UUID
	Date      time.Time
	Label     string
	Fields    map[string]any
}

func (f RootFeed) Roots(ctx context.Context, projectID *uuid.UUID) ([]RootFact, error) {
	proofs, err := f.repository.ListProofs(ctx, projectID)
	if err != nil {
		return nil, err
	}
	facts := make([]RootFact, len(proofs))
	for index, proof := range proofs {
		facts[index] = RootFact{
			ProofID: proof.ID, ProjectID: proof.ProjectID, Date: proof.CreatedAt,
			Label: fmt.Sprintf("Akar penggajian %s, %d pekerja", proof.Period, proof.WorkerCount),
			Fields: map[string]any{
				"period": proof.Period, "root": proof.Root, "minimumWage": proof.MinimumWage, "workerCount": proof.WorkerCount,
			},
		}
	}
	return facts, nil
}
