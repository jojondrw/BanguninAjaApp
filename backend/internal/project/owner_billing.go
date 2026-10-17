package project

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

var (
	errOwnerContractNotFound = apperror.NotFound("owner_contract_not_found", "Kontrak dengan pemberi kerja untuk proyek ini belum diatur")
	errOwnerCustomer         = apperror.Unprocessable("customer_not_found", "Pemberi kerja tidak ditemukan di daftar pelanggan")
	errOwnerContractLocked   = apperror.Conflict("owner_contract_locked", "Sudah ada termin yang ditagihkan, jadi pemberi kerja, nilai, retensi, dan tempo bayar tidak bisa diubah lagi")

	ownerContractErrors = database.ErrorMap{NotFound: errOwnerContractNotFound, Referenced: errOwnerCustomer}
	ownerCustomerErrors = database.ErrorMap{NotFound: errOwnerCustomer}
)

type OwnerContract struct {
	entity.Base
	ProjectID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_owner_contract_project"`
	CustomerID       uuid.UUID `gorm:"type:uuid;not null;index:idx_owner_contract_customer"`
	Number           string    `gorm:"type:varchar(60);not null"`
	Value            int64     `gorm:"not null"`
	RetentionPercent float64   `gorm:"type:numeric(4,2);not null;default:0"`
	PaymentTermDays  int       `gorm:"not null;default:30"`
}

func (OwnerContract) TableName() string {
	return "owner_contract"
}

type OwnerTerm struct {
	entity.Base
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_owner_term_project"`
	Kind             string     `gorm:"type:varchar(20);not null"`
	RequiredProgress int        `gorm:"not null;default:0"`
	Percent          float64    `gorm:"type:numeric(5,2);not null;default:0"`
	GrossAmount      int64      `gorm:"not null;default:0"`
	RetentionAmount  int64      `gorm:"not null;default:0"`
	ReceivableID     *uuid.UUID `gorm:"type:uuid;uniqueIndex:uq_owner_term_receivable"`
	ClaimedOn        *time.Time `gorm:"type:date"`
	ClaimedProgress  *int
	ClaimedBy        *uuid.UUID `gorm:"type:uuid;index:idx_owner_term_claimed_by"`
}

func (OwnerTerm) TableName() string {
	return "owner_term"
}

type OwnerContractRequest struct {
	CustomerID       uuid.UUID `json:"customerId" binding:"required"`
	Number           string    `json:"number" binding:"required,max=60"`
	Value            int64     `json:"value" binding:"required,gt=0"`
	RetentionPercent *float64  `json:"retentionPercent" binding:"required,min=0,max=10"`
	PaymentTermDays  *int      `json:"paymentTermDays" binding:"required,min=0,max=180"`
}

type OwnerContractResponse struct {
	CustomerID       uuid.UUID `json:"customerId"`
	CustomerName     string    `json:"customerName"`
	Number           string    `json:"number"`
	Value            int64     `json:"value"`
	RetentionPercent float64   `json:"retentionPercent"`
	PaymentTermDays  int       `json:"paymentTermDays"`
}

type OwnerTermResponse struct {
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
	ReceivableID     *uuid.UUID `json:"receivableId"`
	ClaimedOn        *time.Time `json:"claimedOn"`
	ClaimedProgress  *int       `json:"claimedProgress"`
}

type OwnerScheduleResponse struct {
	Contract            *OwnerContractResponse `json:"contract"`
	Progress            int                    `json:"progress"`
	Terms               []OwnerTermResponse    `json:"terms"`
	PlannedPercent      float64                `json:"plannedPercent"`
	ClaimedAmount       int64                  `json:"claimedAmount"`
	PaidAmount          int64                  `json:"paidAmount"`
	RetentionHeld       int64                  `json:"retentionHeld"`
	RetentionReleased   bool                   `json:"retentionReleased"`
	ContractLocked      bool                   `json:"contractLocked"`
	CanReleaseRetention bool                   `json:"canReleaseRetention"`
}

func (s *service) OwnerSchedule(ctx context.Context, projectID uuid.UUID) (OwnerScheduleResponse, error) {
	project, err := s.repository.FindProject(ctx, projectID)
	if err != nil {
		return OwnerScheduleResponse{}, projectReadErrors.Resolve(err)
	}
	response, err := ownerSchedule(ctx, s.repository, project)
	if err != nil {
		return OwnerScheduleResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) SaveOwnerContract(ctx context.Context, projectID uuid.UUID, request OwnerContractRequest) (OwnerScheduleResponse, error) {
	return s.changeOwnerSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		if _, err := repository.FindOwnerCustomerName(ctx, request.CustomerID); err != nil {
			return ownerCustomerErrors.Resolve(err)
		}

		contract, err := repository.FindOwnerContract(ctx, projectID)
		isNew := errors.Is(err, database.ErrNotFound)
		if err != nil && !isNew {
			return apperror.Internal(err)
		}
		terms, err := repository.ListOwnerTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}

		next := contract
		next.ProjectID = projectID
		applyOwnerContractRequest(&next, request)
		if !isNew && hasClaims(terms) && ownerTermsChanged(contract, next) {
			return errOwnerContractLocked
		}
		if isNew {
			return ownerContractErrors.Resolve(repository.CreateOwnerContract(ctx, &next))
		}
		return ownerContractErrors.Resolve(repository.SaveOwnerContract(ctx, &next))
	})
}

func (s *service) CreateOwnerTerm(ctx context.Context, projectID uuid.UUID, request PaymentTermRequest) (OwnerScheduleResponse, error) {
	return s.changeOwnerSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		if _, err := repository.FindOwnerContract(ctx, projectID); err != nil {
			return ownerContractErrors.Resolve(err)
		}
		terms, err := repository.ListOwnerTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		if err := validatePaymentTerm(terms, request); err != nil {
			return err
		}
		term := OwnerTerm{
			ProjectID:        projectID,
			Kind:             termProgress,
			RequiredProgress: *request.RequiredProgress,
			Percent:          roundPercent(request.Percent),
		}
		return termErrors.Resolve(repository.CreateOwnerTerm(ctx, &term))
	})
}

func (s *service) DeleteOwnerTerm(ctx context.Context, projectID, termID uuid.UUID) (OwnerScheduleResponse, error) {
	return s.changeOwnerSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		term, err := repository.FindOwnerTerm(ctx, projectID, termID)
		if err != nil {
			return termErrors.Resolve(err)
		}
		if term.Kind != termProgress {
			return errTermNotFound
		}
		if term.ReceivableID != nil {
			return errTermClaimed
		}
		return termErrors.Resolve(repository.DeleteOwnerTerm(ctx, projectID, termID))
	})
}

func (s *service) ClaimOwnerTerm(ctx context.Context, projectID, termID, actorID uuid.UUID) (OwnerScheduleResponse, error) {
	today := s.today()
	return s.changeOwnerSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if project.Status == statusCancelled {
			return errTermProjectCancelled
		}
		contract, err := repository.FindOwnerContract(ctx, projectID)
		if err != nil {
			return ownerContractErrors.Resolve(err)
		}
		term, err := repository.FindOwnerTerm(ctx, projectID, termID)
		if err != nil {
			return termErrors.Resolve(err)
		}
		if term.Kind != termProgress {
			return errTermNotFound
		}
		if term.ReceivableID != nil {
			return errTermClaimed
		}

		terms, err := repository.ListOwnerTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		sortTerms(terms)
		if err := claimBlocker(terms, PaymentTerm{RequiredProgress: term.RequiredProgress, Kind: term.Kind, Base: term.Base}, project.Progress); err != nil {
			return err
		}
		amount := plannedAmounts(terms, contract.Value, contract.RetentionPercent)[term.ID]
		if amount.gross-amount.retention <= 0 {
			return errTermNothingToClaim
		}
		receivableID, err := repository.Billing().RecordReceivable(ctx, ReceivableEntry{
			CustomerID: contract.CustomerID,
			ProjectID:  projectID,
			Reference:  fmt.Sprintf("Tagihan termin %d %s", termPosition(terms, term.ID), project.Code),
			DueDate:    today.AddDate(0, 0, contract.PaymentTermDays),
			Amount:     amount.gross - amount.retention,
		})
		if err != nil {
			return err
		}

		term.GrossAmount = amount.gross
		term.RetentionAmount = amount.retention
		markOwnerClaimed(&term, receivableID, project.Progress, actorID, today)
		return termErrors.Resolve(repository.SaveOwnerTerm(ctx, &term))
	})
}

func (s *service) ReleaseOwnerRetention(ctx context.Context, projectID, actorID uuid.UUID) (OwnerScheduleResponse, error) {
	today := s.today()
	return s.changeOwnerSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if project.Status != statusCompleted {
			return errRetentionNotDue
		}
		contract, err := repository.FindOwnerContract(ctx, projectID)
		if err != nil {
			return ownerContractErrors.Resolve(err)
		}
		terms, err := repository.ListOwnerTerms(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		held, previous, err := releasableRetention(terms)
		if err != nil {
			return err
		}

		receivableID, err := repository.Billing().RecordReceivable(ctx, ReceivableEntry{
			CustomerID: contract.CustomerID,
			ProjectID:  projectID,
			Reference:  "Tagihan retensi " + project.Code,
			DueDate:    today.AddDate(0, 0, contract.PaymentTermDays),
			Amount:     held,
		})
		if err != nil {
			return err
		}

		release := OwnerTerm{ProjectID: projectID, Kind: termRetention, RequiredProgress: fullProgress}
		if previous != nil {
			release, err = repository.FindOwnerTerm(ctx, projectID, *previous)
			if err != nil {
				return termErrors.Resolve(err)
			}
		}
		release.GrossAmount = held
		release.RetentionAmount = 0
		markOwnerClaimed(&release, receivableID, project.Progress, actorID, today)
		if previous != nil {
			return termErrors.Resolve(repository.SaveOwnerTerm(ctx, &release))
		}
		return termErrors.Resolve(repository.CreateOwnerTerm(ctx, &release))
	})
}

func (s *service) changeOwnerSchedule(ctx context.Context, projectID uuid.UUID, change func(Repository, Project) error) (OwnerScheduleResponse, error) {
	var response OwnerScheduleResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		project, err := repository.LockProject(ctx, projectID)
		if err != nil {
			return projectReadErrors.Resolve(err)
		}
		if err := change(repository, project); err != nil {
			return err
		}
		response, err = ownerSchedule(ctx, repository, project)
		return err
	})
	if err != nil {
		return OwnerScheduleResponse{}, apperror.From(err)
	}
	return response, nil
}

func ownerSchedule(ctx context.Context, repository Repository, project Project) (OwnerScheduleResponse, error) {
	terms, err := repository.ListOwnerTerms(ctx, project.ID)
	if err != nil {
		return OwnerScheduleResponse{}, apperror.Internal(err)
	}
	sortTerms(terms)

	contract, err := repository.FindOwnerContract(ctx, project.ID)
	if errors.Is(err, database.ErrNotFound) {
		return newOwnerScheduleResponse(project, nil, terms), nil
	}
	if err != nil {
		return OwnerScheduleResponse{}, apperror.Internal(err)
	}
	customer, err := repository.FindOwnerCustomerName(ctx, contract.CustomerID)
	if err != nil {
		return OwnerScheduleResponse{}, apperror.Internal(err)
	}
	return newOwnerScheduleResponse(project, &OwnerContractResponse{
		CustomerID:       contract.CustomerID,
		CustomerName:     customer,
		Number:           contract.Number,
		Value:            contract.Value,
		RetentionPercent: contract.RetentionPercent,
		PaymentTermDays:  contract.PaymentTermDays,
	}, terms), nil
}

func newOwnerScheduleResponse(project Project, contract *OwnerContractResponse, terms []PaymentTermRow) OwnerScheduleResponse {
	var basis *ContractResponse
	if contract != nil {
		basis = &ContractResponse{CurrentValue: contract.Value, RetentionPercent: contract.RetentionPercent}
	}
	schedule := newPaymentScheduleResponse(project, basis, terms, nil)
	response := OwnerScheduleResponse{
		Contract:            contract,
		Progress:            schedule.Progress,
		Terms:               make([]OwnerTermResponse, 0, len(schedule.Terms)),
		PlannedPercent:      schedule.PlannedPercent,
		ClaimedAmount:       schedule.ClaimedAmount,
		PaidAmount:          schedule.PaidAmount,
		RetentionHeld:       schedule.RetentionHeld,
		RetentionReleased:   schedule.RetentionReleased,
		ContractLocked:      schedule.ContractLocked,
		CanReleaseRetention: schedule.CanReleaseRetention,
	}
	for _, term := range schedule.Terms {
		response.Terms = append(response.Terms, OwnerTermResponse{
			ID:               term.ID,
			Label:            term.Label,
			Kind:             term.Kind,
			State:            term.State,
			RequiredProgress: term.RequiredProgress,
			Percent:          term.Percent,
			GrossAmount:      term.GrossAmount,
			RetentionAmount:  term.RetentionAmount,
			NetAmount:        term.NetAmount,
			PaidAmount:       term.PaidAmount,
			ReceivableID:     term.PayableID,
			ClaimedOn:        term.ClaimedOn,
			ClaimedProgress:  term.ClaimedProgress,
		})
	}
	return response
}

func markOwnerClaimed(term *OwnerTerm, receivableID uuid.UUID, progress int, actorID uuid.UUID, today time.Time) {
	term.ReceivableID = &receivableID
	term.ClaimedOn = &today
	term.ClaimedProgress = &progress
	term.ClaimedBy = &actorID
}

func ownerTermsChanged(current, next OwnerContract) bool {
	return current.CustomerID != next.CustomerID ||
		current.Value != next.Value ||
		current.PaymentTermDays != next.PaymentTermDays ||
		percentUnits(current.RetentionPercent) != percentUnits(next.RetentionPercent)
}

func applyOwnerContractRequest(contract *OwnerContract, request OwnerContractRequest) {
	contract.CustomerID = request.CustomerID
	contract.Number = strings.TrimSpace(request.Number)
	contract.Value = request.Value
	contract.RetentionPercent = roundPercent(*request.RetentionPercent)
	contract.PaymentTermDays = *request.PaymentTermDays
}

const ownerTermColumns = "owner_term.id, owner_term.kind, owner_term.required_progress, owner_term.percent, " +
	"owner_term.gross_amount, owner_term.retention_amount, owner_term.receivable_id AS payable_id, " +
	"owner_term.claimed_on, owner_term.claimed_progress, owner_term.created_at, " +
	"COALESCE(receivable.amount, 0) AS payable_amount, COALESCE(receivable.paid_amount, 0) AS payable_paid"

func (r *gormRepository) ListOwnerTerms(ctx context.Context, projectID uuid.UUID) ([]PaymentTermRow, error) {
	rows := []PaymentTermRow{}
	err := r.db.WithContext(ctx).
		Table("owner_term").
		Select(ownerTermColumns).
		Joins("LEFT JOIN receivable ON receivable.id = owner_term.receivable_id").
		Where("owner_term.project_id = ?", projectID).
		Order("owner_term.required_progress, owner_term.created_at, owner_term.id").
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) FindOwnerTerm(ctx context.Context, projectID, id uuid.UUID) (OwnerTerm, error) {
	var term OwnerTerm
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&term).Error
	return term, database.Translate(err)
}

func (r *gormRepository) CreateOwnerTerm(ctx context.Context, term *OwnerTerm) error {
	return database.Translate(r.db.WithContext(ctx).Create(term).Error)
}

func (r *gormRepository) SaveOwnerTerm(ctx context.Context, term *OwnerTerm) error {
	return database.Translate(r.db.WithContext(ctx).Save(term).Error)
}

func (r *gormRepository) DeleteOwnerTerm(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &OwnerTerm{}, "id = ? AND project_id = ?", id, projectID)
}

func (r *gormRepository) FindOwnerContract(ctx context.Context, projectID uuid.UUID) (OwnerContract, error) {
	var contract OwnerContract
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Take(&contract).Error
	return contract, database.Translate(err)
}

func (r *gormRepository) CreateOwnerContract(ctx context.Context, contract *OwnerContract) error {
	return database.Translate(r.db.WithContext(ctx).Create(contract).Error)
}

func (r *gormRepository) SaveOwnerContract(ctx context.Context, contract *OwnerContract) error {
	return database.Translate(r.db.WithContext(ctx).Save(contract).Error)
}

func (r *gormRepository) FindOwnerCustomerName(ctx context.Context, id uuid.UUID) (string, error) {
	var customer struct{ Name string }
	err := r.db.WithContext(ctx).Table("customer").Select("name").Where("id = ?", id).Take(&customer).Error
	return customer.Name, database.Translate(err)
}
