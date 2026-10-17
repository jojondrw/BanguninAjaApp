package project

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

const (
	termProgress  = "progress"
	termRetention = "retention"

	termPlanned = "planned"
	termReady   = "ready"
	termClaimed = "claimed"
	termPaid    = "paid"

	percentUnit  = 100
	wholePercent = 100 * percentUnit
)

var (
	errContractNotFound       = apperror.NotFound("construction_contract_not_found", "Kontrak kontraktor untuk proyek ini belum diatur")
	errContractVendor         = apperror.Unprocessable("vendor_not_found", "Kontraktor tidak ditemukan di daftar vendor")
	errContractVendorInactive = apperror.Unprocessable("vendor_inactive", "Kontraktor ini sudah tidak aktif di daftar vendor")
	errContractLocked         = apperror.Conflict("construction_contract_locked", "Sudah ada termin yang diajukan, jadi kontraktor, nilai, dan retensi tidak bisa diubah lagi")
	errTermNotFound           = apperror.NotFound("payment_term_not_found", "Termin tidak ditemukan")
	errTermPercentExceeded    = apperror.Unprocessable("payment_term_percent_exceeded", "Jumlah persen semua termin tidak boleh lebih dari 100% nilai kontrak")
	errTermBeforeClaimed      = apperror.Unprocessable("payment_term_before_claimed", "Syarat progres tidak boleh lebih rendah dari termin yang sudah diajukan")
	errTermClaimed            = apperror.Conflict("payment_term_claimed", "Termin ini sudah diajukan sehingga tidak bisa diubah atau dihapus")
	errTermOutOfOrder         = apperror.Unprocessable("payment_term_out_of_order", "Ajukan dulu termin sebelumnya")
	errTermProjectCancelled   = apperror.Conflict("payment_term_project_cancelled", "Proyek sudah dibatalkan sehingga termin tidak bisa diajukan")
	errTermNothingToClaim     = apperror.Unprocessable("payment_term_nothing_to_claim", "Nilai termin ini sudah habis tertagih oleh termin sebelumnya")
	errRetentionNotDue        = apperror.Unprocessable("retention_not_due", "Retensi baru bisa dilepas setelah proyek selesai dan diserahterimakan")
	errRetentionTermsOpen     = apperror.Unprocessable("retention_terms_open", "Masih ada termin progres yang belum diajukan")
	errRetentionEmpty         = apperror.Unprocessable("retention_empty", "Belum ada retensi yang ditahan")
	errRetentionReleased      = apperror.Conflict("retention_released", "Retensi proyek ini sudah dilepas")

	contractErrors = database.ErrorMap{NotFound: errContractNotFound, Referenced: errContractVendor}
	vendorErrors   = database.ErrorMap{NotFound: errContractVendor}
	termErrors     = database.ErrorMap{NotFound: errTermNotFound}
)

type ConstructionContract struct {
	entity.Base
	ProjectID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_construction_contract_project"`
	VendorID         uuid.UUID `gorm:"type:uuid;not null;index:idx_construction_contract_vendor"`
	Number           string    `gorm:"type:varchar(60);not null"`
	Value            int64     `gorm:"not null"`
	RetentionPercent float64   `gorm:"type:numeric(4,2);not null;default:0"`
}

func (ConstructionContract) TableName() string {
	return "construction_contract"
}

type PaymentTerm struct {
	entity.Base
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_payment_term_project"`
	Kind             string     `gorm:"type:varchar(20);not null"`
	RequiredProgress int        `gorm:"not null;default:0"`
	Percent          float64    `gorm:"type:numeric(5,2);not null;default:0"`
	GrossAmount      int64      `gorm:"not null;default:0"`
	RetentionAmount  int64      `gorm:"not null;default:0"`
	PayableID        *uuid.UUID `gorm:"type:uuid;uniqueIndex:uq_payment_term_payable"`
	ClaimedOn        *time.Time `gorm:"type:date"`
	ClaimedProgress  *int
	ClaimedBy        *uuid.UUID `gorm:"type:uuid;index:idx_payment_term_claimed_by"`
}

func (PaymentTerm) TableName() string {
	return "payment_term"
}

type PayableEntry struct {
	VendorID  uuid.UUID
	ProjectID uuid.UUID
	Reference string
	DueDate   time.Time
	Amount    int64
}

type ReceivableEntry struct {
	CustomerID uuid.UUID
	ProjectID  uuid.UUID
	Reference  string
	DueDate    time.Time
	Amount     int64
}

type BillingLedger interface {
	RecordPayable(ctx context.Context, entry PayableEntry) (uuid.UUID, error)
	RecordReceivable(ctx context.Context, entry ReceivableEntry) (uuid.UUID, error)
}

type ContractVendor struct {
	Name            string
	PaymentTermDays int
	Active          bool
}

type PaymentTermRow struct {
	ID               uuid.UUID
	Kind             string
	RequiredProgress int
	Percent          float64
	GrossAmount      int64
	RetentionAmount  int64
	PayableID        *uuid.UUID
	PayableAmount    int64
	PayablePaid      int64
	ClaimedOn        *time.Time
	ClaimedProgress  *int
	CreatedAt        time.Time
}

type ContractRequest struct {
	VendorID         uuid.UUID `json:"vendorId" binding:"required"`
	Number           string    `json:"number" binding:"required,max=60"`
	Value            int64     `json:"value" binding:"required,gt=0"`
	RetentionPercent *float64  `json:"retentionPercent" binding:"required,min=0,max=10"`
}

type PaymentTermRequest struct {
	RequiredProgress *int    `json:"requiredProgress" binding:"required,min=0,max=100"`
	Percent          float64 `json:"percent" binding:"required,gt=0,max=100"`
}

type ContractResponse struct {
	VendorID         uuid.UUID `json:"vendorId"`
	VendorName       string    `json:"vendorName"`
	PaymentTermDays  int       `json:"paymentTermDays"`
	Number           string    `json:"number"`
	Value            int64     `json:"value"`
	CurrentValue     int64     `json:"currentValue"`
	ApprovedChange   int64     `json:"approvedChange"`
	ExtensionDays    int       `json:"extensionDays"`
	RetentionPercent float64   `json:"retentionPercent"`
}

type termAmount struct {
	gross     int64
	retention int64
}

type PaymentTermResponse struct {
	ID               uuid.UUID  `json:"id"`
	Label            string     `json:"label"`
	Kind             string     `json:"kind"`
	State            string     `json:"state"`
	RequiredProgress int        `json:"requiredProgress"`
	Percent          float64    `json:"percent"`
	GrossAmount      int64      `json:"grossAmount"`
	RetentionAmount  int64      `json:"retentionAmount"`
	NetAmount        int64      `json:"netAmount"`
	PaidAmount       int64      `json:"paidAmount"`
	PayableID        *uuid.UUID `json:"payableId"`
	ClaimedOn        *time.Time `json:"claimedOn"`
	ClaimedProgress  *int       `json:"claimedProgress"`
}

type PaymentScheduleResponse struct {
	Contract            *ContractResponse        `json:"contract"`
	Progress            int                      `json:"progress"`
	Terms               []PaymentTermResponse    `json:"terms"`
	Changes             []ContractChangeResponse `json:"changes"`
	PlannedPercent      float64                  `json:"plannedPercent"`
	ClaimedAmount       int64                    `json:"claimedAmount"`
	PaidAmount          int64                    `json:"paidAmount"`
	RetentionHeld       int64                    `json:"retentionHeld"`
	RetentionReleased   bool                     `json:"retentionReleased"`
	ContractLocked      bool                     `json:"contractLocked"`
	CanReleaseRetention bool                     `json:"canReleaseRetention"`
}

func (s *service) PaymentSchedule(ctx context.Context, projectID uuid.UUID) (PaymentScheduleResponse, error) {
	project, err := s.repository.FindProject(ctx, projectID)
	if err != nil {
		return PaymentScheduleResponse{}, projectReadErrors.Resolve(err)
	}
	response, err := paymentSchedule(ctx, s.repository, project)
	if err != nil {
		return PaymentScheduleResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) SaveContract(ctx context.Context, projectID uuid.UUID, request ContractRequest) (PaymentScheduleResponse, error) {
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		vendor, err := repository.FindContractVendor(ctx, request.VendorID)
		if err != nil {
			return vendorErrors.Resolve(err)
		}

		contract, err := repository.FindContract(ctx, projectID)
		isNew := errors.Is(err, database.ErrNotFound)
		if err != nil && !isNew {
			return apperror.Internal(err)
		}
		if isNew || contract.VendorID != request.VendorID {
			if !vendor.Active {
				return errContractVendorInactive
			}
		}

		terms, err := repository.ListPaymentTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		next := contract
		next.ProjectID = projectID
		applyContractRequest(&next, request)
		if !isNew && hasClaims(terms) && contractTermsChanged(contract, next) {
			return errContractLocked
		}

		if isNew {
			return contractErrors.Resolve(repository.CreateContract(ctx, &next))
		}
		return contractErrors.Resolve(repository.SaveContract(ctx, &next))
	})
}

func (s *service) CreatePaymentTerm(ctx context.Context, projectID uuid.UUID, request PaymentTermRequest) (PaymentScheduleResponse, error) {
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		if _, err := repository.FindContract(ctx, projectID); err != nil {
			return contractErrors.Resolve(err)
		}
		terms, err := repository.ListPaymentTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		if err := validatePaymentTerm(terms, request); err != nil {
			return err
		}

		term := PaymentTerm{
			ProjectID:        projectID,
			Kind:             termProgress,
			RequiredProgress: *request.RequiredProgress,
			Percent:          roundPercent(request.Percent),
		}
		return termErrors.Resolve(repository.CreatePaymentTerm(ctx, &term))
	})
}

func (s *service) DeletePaymentTerm(ctx context.Context, projectID, termID uuid.UUID) (PaymentScheduleResponse, error) {
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		term, err := repository.FindPaymentTerm(ctx, projectID, termID)
		if err != nil {
			return termErrors.Resolve(err)
		}
		if term.Kind != termProgress {
			return errTermNotFound
		}
		if term.PayableID != nil {
			return errTermClaimed
		}
		return termErrors.Resolve(repository.DeletePaymentTerm(ctx, projectID, termID))
	})
}

func (s *service) ClaimPaymentTerm(ctx context.Context, projectID, termID, actorID uuid.UUID) (PaymentScheduleResponse, error) {
	today := s.today()
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if project.Status == statusCancelled {
			return errTermProjectCancelled
		}
		contract, err := repository.FindContract(ctx, projectID)
		if err != nil {
			return contractErrors.Resolve(err)
		}
		term, err := repository.FindPaymentTerm(ctx, projectID, termID)
		if err != nil {
			return termErrors.Resolve(err)
		}
		if term.Kind != termProgress {
			return errTermNotFound
		}
		if term.PayableID != nil {
			return errTermClaimed
		}

		terms, err := repository.ListPaymentTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		sortTerms(terms)
		if err := claimBlocker(terms, term, project.Progress); err != nil {
			return err
		}
		changes, err := repository.ListContractChanges(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}

		current := contract.Value + sumApproved(changes).value
		amount := plannedAmounts(terms, current, contract.RetentionPercent)[term.ID]
		if amount.gross-amount.retention <= 0 {
			return errTermNothingToClaim
		}
		payableID, err := recordPayable(ctx, repository, contract, PayableEntry{
			ProjectID: projectID,
			Reference: fmt.Sprintf("Termin %d %s", termPosition(terms, term.ID), project.Code),
			Amount:    amount.gross - amount.retention,
		}, today)
		if err != nil {
			return err
		}

		term.GrossAmount = amount.gross
		term.RetentionAmount = amount.retention
		markClaimed(&term, payableID, project.Progress, actorID, today)
		return termErrors.Resolve(repository.SavePaymentTerm(ctx, &term))
	})
}

func (s *service) ReleaseRetention(ctx context.Context, projectID, actorID uuid.UUID) (PaymentScheduleResponse, error) {
	today := s.today()
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if project.Status != statusCompleted {
			return errRetentionNotDue
		}
		contract, err := repository.FindContract(ctx, projectID)
		if err != nil {
			return contractErrors.Resolve(err)
		}
		terms, err := repository.ListPaymentTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		held, previous, err := releasableRetention(terms)
		if err != nil {
			return err
		}

		payableID, err := recordPayable(ctx, repository, contract, PayableEntry{
			ProjectID: projectID,
			Reference: "Retensi " + project.Code,
			Amount:    held,
		}, today)
		if err != nil {
			return err
		}

		release := PaymentTerm{ProjectID: projectID, Kind: termRetention, RequiredProgress: fullProgress}
		if previous != nil {
			release, err = repository.FindPaymentTerm(ctx, projectID, *previous)
			if err != nil {
				return termErrors.Resolve(err)
			}
		}
		release.GrossAmount = held
		release.RetentionAmount = 0
		markClaimed(&release, payableID, project.Progress, actorID, today)
		if previous != nil {
			return termErrors.Resolve(repository.SavePaymentTerm(ctx, &release))
		}
		return termErrors.Resolve(repository.CreatePaymentTerm(ctx, &release))
	})
}

func (s *service) changeSchedule(ctx context.Context, projectID uuid.UUID, change func(Repository, Project) error) (PaymentScheduleResponse, error) {
	var response PaymentScheduleResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		project, err := repository.LockProject(ctx, projectID)
		if err != nil {
			return projectReadErrors.Resolve(err)
		}
		if err := change(repository, project); err != nil {
			return err
		}
		response, err = paymentSchedule(ctx, repository, project)
		return err
	})
	if err != nil {
		return PaymentScheduleResponse{}, apperror.From(err)
	}
	return response, nil
}

func paymentSchedule(ctx context.Context, repository Repository, project Project) (PaymentScheduleResponse, error) {
	terms, err := repository.ListPaymentTerms(ctx, project.ID)
	if err != nil {
		return PaymentScheduleResponse{}, apperror.Internal(err)
	}
	sortTerms(terms)
	changes, err := repository.ListContractChanges(ctx, project.ID)
	if err != nil {
		return PaymentScheduleResponse{}, apperror.Internal(err)
	}

	contract, err := repository.FindContract(ctx, project.ID)
	if errors.Is(err, database.ErrNotFound) {
		return newPaymentScheduleResponse(project, nil, terms, changes), nil
	}
	if err != nil {
		return PaymentScheduleResponse{}, apperror.Internal(err)
	}
	vendor, err := repository.FindContractVendor(ctx, contract.VendorID)
	if err != nil {
		return PaymentScheduleResponse{}, apperror.Internal(err)
	}
	approved := sumApproved(changes)
	return newPaymentScheduleResponse(project, &ContractResponse{
		VendorID:         contract.VendorID,
		VendorName:       vendor.Name,
		PaymentTermDays:  vendor.PaymentTermDays,
		Number:           contract.Number,
		Value:            contract.Value,
		CurrentValue:     contract.Value + approved.value,
		ApprovedChange:   approved.value,
		ExtensionDays:    approved.days,
		RetentionPercent: contract.RetentionPercent,
	}, terms, changes), nil
}

func recordPayable(ctx context.Context, repository Repository, contract ConstructionContract, entry PayableEntry, today time.Time) (uuid.UUID, error) {
	vendor, err := repository.FindContractVendor(ctx, contract.VendorID)
	if err != nil {
		return uuid.Nil, vendorErrors.Resolve(err)
	}
	entry.VendorID = contract.VendorID
	entry.DueDate = today.AddDate(0, 0, vendor.PaymentTermDays)
	return repository.Billing().RecordPayable(ctx, entry)
}

func markClaimed(term *PaymentTerm, payableID uuid.UUID, progress int, actorID uuid.UUID, today time.Time) {
	term.PayableID = &payableID
	term.ClaimedOn = &today
	term.ClaimedProgress = &progress
	term.ClaimedBy = &actorID
}

func newPaymentScheduleResponse(project Project, contract *ContractResponse, terms []PaymentTermRow, changes []ContractChangeRow) PaymentScheduleResponse {
	response := PaymentScheduleResponse{
		Contract: contract,
		Progress: project.Progress,
		Terms:    []PaymentTermResponse{},
		Changes:  newContractChangeResponses(changes),
	}
	amounts := map[uuid.UUID]termAmount{}
	if contract != nil {
		amounts = plannedAmounts(terms, contract.CurrentValue, contract.RetentionPercent)
	}
	var plannedUnits int64
	openTerms := false
	position := 0
	for _, term := range terms {
		if term.Kind == termRetention {
			if term.PayableID == nil {
				continue
			}
			response.RetentionReleased = true
			response.PaidAmount += term.PayablePaid
			response.Terms = append(response.Terms, retentionTermResponse(term))
			continue
		}

		position++
		plannedUnits += percentUnits(term.Percent)
		item := progressTermResponse(term, position, amounts[term.ID])
		if term.PayableID == nil {
			if !openTerms && contract != nil && project.Status != statusCancelled && project.Progress >= term.RequiredProgress {
				item.State = termReady
			}
			openTerms = true
		} else {
			response.ContractLocked = true
			response.ClaimedAmount += term.GrossAmount
			response.RetentionHeld += term.RetentionAmount
			response.PaidAmount += term.PayablePaid
		}
		response.Terms = append(response.Terms, item)
	}

	response.PlannedPercent = float64(plannedUnits) / percentUnit
	response.CanReleaseRetention = contract != nil && project.Status == statusCompleted &&
		!openTerms && response.RetentionHeld > 0 && !response.RetentionReleased
	return response
}

func progressTermResponse(term PaymentTermRow, position int, planned termAmount) PaymentTermResponse {
	item := PaymentTermResponse{
		ID:               term.ID,
		Label:            fmt.Sprintf("Termin %d", position),
		Kind:             term.Kind,
		State:            termPlanned,
		RequiredProgress: term.RequiredProgress,
		Percent:          term.Percent,
		GrossAmount:      term.GrossAmount,
		RetentionAmount:  term.RetentionAmount,
		PaidAmount:       term.PayablePaid,
		PayableID:        term.PayableID,
		ClaimedOn:        term.ClaimedOn,
		ClaimedProgress:  term.ClaimedProgress,
	}
	if term.PayableID == nil {
		item.PaidAmount = 0
		item.ClaimedOn = nil
		item.ClaimedProgress = nil
		item.GrossAmount, item.RetentionAmount = planned.gross, planned.retention
	} else {
		item.State = payableState(term)
	}
	item.NetAmount = item.GrossAmount - item.RetentionAmount
	return item
}

func retentionTermResponse(term PaymentTermRow) PaymentTermResponse {
	return PaymentTermResponse{
		ID:               term.ID,
		Label:            "Pelepasan retensi",
		Kind:             term.Kind,
		State:            payableState(term),
		RequiredProgress: term.RequiredProgress,
		GrossAmount:      term.GrossAmount,
		NetAmount:        term.GrossAmount,
		PaidAmount:       term.PayablePaid,
		PayableID:        term.PayableID,
		ClaimedOn:        term.ClaimedOn,
		ClaimedProgress:  term.ClaimedProgress,
	}
}

func payableState(term PaymentTermRow) string {
	if term.PayableAmount > 0 && term.PayablePaid >= term.PayableAmount {
		return termPaid
	}
	return termClaimed
}

func validatePaymentTerm(terms []PaymentTermRow, request PaymentTermRequest) error {
	var planned int64
	for _, term := range terms {
		if term.Kind != termProgress {
			continue
		}
		planned += percentUnits(term.Percent)
		if term.PayableID != nil && term.RequiredProgress > *request.RequiredProgress {
			return errTermBeforeClaimed
		}
	}
	if planned+percentUnits(request.Percent) > wholePercent {
		return errTermPercentExceeded
	}
	return nil
}

func claimBlocker(sorted []PaymentTermRow, target PaymentTerm, progress int) error {
	if progress < target.RequiredProgress {
		return apperror.Unprocessable("payment_term_progress_not_reached",
			fmt.Sprintf("Progres berbobot proyek baru %d%%, termin ini mensyaratkan %d%%", progress, target.RequiredProgress))
	}
	for _, term := range sorted {
		if term.ID == target.ID {
			return nil
		}
		if term.Kind == termProgress && term.PayableID == nil {
			return errTermOutOfOrder
		}
	}
	return nil
}

func releasableRetention(terms []PaymentTermRow) (int64, *uuid.UUID, error) {
	var held int64
	var previous *uuid.UUID
	for _, term := range terms {
		switch {
		case term.Kind == termRetention && term.PayableID != nil:
			return 0, nil, errRetentionReleased
		case term.Kind == termRetention:
			id := term.ID
			previous = &id
		case term.PayableID == nil:
			return 0, nil, errRetentionTermsOpen
		default:
			held += term.RetentionAmount
		}
	}
	if held == 0 {
		return 0, nil, errRetentionEmpty
	}
	return held, previous, nil
}

func plannedAmounts(sorted []PaymentTermRow, contractValue int64, retentionPercent float64) map[uuid.UUID]termAmount {
	amounts := map[uuid.UUID]termAmount{}
	var cumulativeUnits, billed int64
	for _, term := range sorted {
		if term.Kind != termProgress {
			continue
		}
		cumulativeUnits += percentUnits(term.Percent)
		if term.PayableID != nil {
			billed += term.GrossAmount
			continue
		}
		gross := max(0, shareOf(contractValue, cumulativeUnits)-billed)
		billed += gross
		amounts[term.ID] = termAmount{gross: gross, retention: shareOf(gross, percentUnits(retentionPercent))}
	}
	return amounts
}

func shareOf(amount, units int64) int64 {
	return (amount*units + wholePercent/2) / wholePercent
}

func termPosition(sorted []PaymentTermRow, id uuid.UUID) int {
	position := 0
	for _, term := range sorted {
		if term.Kind != termProgress {
			continue
		}
		position++
		if term.ID == id {
			return position
		}
	}
	return position
}

func sortTerms(terms []PaymentTermRow) {
	slices.SortStableFunc(terms, func(a, b PaymentTermRow) int {
		return cmp.Or(
			cmp.Compare(kindOrder(a.Kind), kindOrder(b.Kind)),
			cmp.Compare(a.RequiredProgress, b.RequiredProgress),
			a.CreatedAt.Compare(b.CreatedAt),
			strings.Compare(a.ID.String(), b.ID.String()),
		)
	})
}

func kindOrder(kind string) int {
	if kind == termRetention {
		return 1
	}
	return 0
}

func hasClaims(terms []PaymentTermRow) bool {
	return slices.ContainsFunc(terms, func(term PaymentTermRow) bool { return term.PayableID != nil })
}

func contractTermsChanged(current, next ConstructionContract) bool {
	return current.VendorID != next.VendorID ||
		current.Value != next.Value ||
		percentUnits(current.RetentionPercent) != percentUnits(next.RetentionPercent)
}

func applyContractRequest(contract *ConstructionContract, request ContractRequest) {
	contract.VendorID = request.VendorID
	contract.Number = strings.TrimSpace(request.Number)
	contract.Value = request.Value
	contract.RetentionPercent = roundPercent(*request.RetentionPercent)
}

func percentUnits(percent float64) int64 {
	return int64(math.Round(percent * percentUnit))
}

func roundPercent(percent float64) float64 {
	return float64(percentUnits(percent)) / percentUnit
}

const paymentTermColumns = "payment_term.id, payment_term.kind, payment_term.required_progress, payment_term.percent, " +
	"payment_term.gross_amount, payment_term.retention_amount, payment_term.payable_id, " +
	"payment_term.claimed_on, payment_term.claimed_progress, payment_term.created_at, " +
	"COALESCE(payable.amount, 0) AS payable_amount, COALESCE(payable.paid_amount, 0) AS payable_paid"

func (r *gormRepository) ListPaymentTerms(ctx context.Context, projectID uuid.UUID) ([]PaymentTermRow, error) {
	rows := []PaymentTermRow{}
	err := r.db.WithContext(ctx).
		Table("payment_term").
		Select(paymentTermColumns).
		Joins("LEFT JOIN payable ON payable.id = payment_term.payable_id").
		Where("payment_term.project_id = ?", projectID).
		Order("payment_term.required_progress, payment_term.created_at, payment_term.id").
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) FindPaymentTerm(ctx context.Context, projectID, id uuid.UUID) (PaymentTerm, error) {
	var term PaymentTerm
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&term).Error
	return term, database.Translate(err)
}

func (r *gormRepository) CreatePaymentTerm(ctx context.Context, term *PaymentTerm) error {
	return database.Translate(r.db.WithContext(ctx).Create(term).Error)
}

func (r *gormRepository) SavePaymentTerm(ctx context.Context, term *PaymentTerm) error {
	return database.Translate(r.db.WithContext(ctx).Save(term).Error)
}

func (r *gormRepository) DeletePaymentTerm(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &PaymentTerm{}, "id = ? AND project_id = ?", id, projectID)
}

func (r *gormRepository) FindContract(ctx context.Context, projectID uuid.UUID) (ConstructionContract, error) {
	var contract ConstructionContract
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Take(&contract).Error
	return contract, database.Translate(err)
}

func (r *gormRepository) CreateContract(ctx context.Context, contract *ConstructionContract) error {
	return database.Translate(r.db.WithContext(ctx).Create(contract).Error)
}

func (r *gormRepository) SaveContract(ctx context.Context, contract *ConstructionContract) error {
	return database.Translate(r.db.WithContext(ctx).Save(contract).Error)
}

func (r *gormRepository) FindContractVendor(ctx context.Context, id uuid.UUID) (ContractVendor, error) {
	var vendor ContractVendor
	err := r.db.WithContext(ctx).
		Table("vendor").
		Select("name, payment_term_days, active").
		Where("id = ?", id).
		Take(&vendor).Error
	return vendor, database.Translate(err)
}

func (r *gormRepository) Billing() BillingLedger {
	return r.billing(r.db)
}

type BillingLedgerFactory func(db *gorm.DB) BillingLedger
