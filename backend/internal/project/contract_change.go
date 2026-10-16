package project

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

const (
	changeProposed = "proposed"
	changeApproved = "approved"
	changeRejected = "rejected"
)

var (
	errChangeNotFound         = apperror.NotFound("contract_change_not_found", "Adendum tidak ditemukan")
	errChangeEmpty            = apperror.Unprocessable("contract_change_empty", "Adendum harus mengubah nilai kontrak atau waktu pelaksanaan")
	errChangeDecided          = apperror.Conflict("contract_change_decided", "Adendum ini sudah diputuskan sehingga tidak bisa diubah atau dihapus")
	errChangeApprover         = apperror.Forbidden("contract_change_approver", "Hanya Admin yang bisa menyetujui atau menolak adendum")
	errChangeNonPositive      = apperror.Unprocessable("contract_change_non_positive", "Nilai kontrak setelah adendum harus lebih dari nol")
	errChangeBelowClaimed     = apperror.Unprocessable("contract_change_below_claimed", "Nilai kontrak setelah adendum tidak boleh di bawah nilai termin yang sudah diajukan")
	errChangeProjectCancelled = apperror.Conflict("contract_change_project_cancelled", "Proyek sudah dibatalkan sehingga adendum tidak bisa diputuskan")

	changeErrors = database.ErrorMap{NotFound: errChangeNotFound}
)

type ContractChange struct {
	entity.Base
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_contract_change_project"`
	Sequence    int        `gorm:"not null"`
	Description string     `gorm:"type:varchar(300);not null"`
	ValueChange int64      `gorm:"not null;default:0"`
	DaysChange  int        `gorm:"not null;default:0"`
	Status      string     `gorm:"type:varchar(20);not null"`
	ProposedBy  *uuid.UUID `gorm:"type:uuid;index:idx_contract_change_proposed_by"`
	DecidedBy   *uuid.UUID `gorm:"type:uuid;index:idx_contract_change_decided_by"`
	DecidedOn   *time.Time `gorm:"type:date"`
}

func (ContractChange) TableName() string {
	return "contract_change"
}

type ContractChangeRow struct {
	ID           uuid.UUID
	Sequence     int
	Description  string
	ValueChange  int64
	DaysChange   int
	Status       string
	ProposerName *string
	DeciderName  *string
	DecidedOn    *time.Time
	CreatedAt    time.Time
}

type ContractChangeRequest struct {
	Description string `json:"description" binding:"required,max=300"`
	ValueChange int64  `json:"valueChange"`
	DaysChange  int    `json:"daysChange" binding:"min=0,max=3650"`
}

type ContractChangeDecision struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

type ContractChangeResponse struct {
	ID           uuid.UUID  `json:"id"`
	Label        string     `json:"label"`
	Description  string     `json:"description"`
	ValueChange  int64      `json:"valueChange"`
	DaysChange   int        `json:"daysChange"`
	Status       string     `json:"status"`
	ProposerName *string    `json:"proposerName"`
	DeciderName  *string    `json:"deciderName"`
	DecidedOn    *time.Time `json:"decidedOn"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type approvedChanges struct {
	value int64
	days  int
}

func (s *service) ProposeContractChange(ctx context.Context, projectID, actorID uuid.UUID, request ContractChangeRequest) (PaymentScheduleResponse, error) {
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		if request.ValueChange == 0 && request.DaysChange == 0 {
			return errChangeEmpty
		}
		contract, err := repository.FindContract(ctx, projectID)
		if err != nil {
			return contractErrors.Resolve(err)
		}
		changes, err := repository.ListContractChanges(ctx, projectID)
		if err != nil {
			return apperror.Internal(err)
		}
		if contract.Value+sumApproved(changes).value+request.ValueChange <= 0 {
			return errChangeNonPositive
		}

		change := ContractChange{
			ProjectID:   projectID,
			Sequence:    nextChangeSequence(changes),
			Description: strings.TrimSpace(request.Description),
			ValueChange: request.ValueChange,
			DaysChange:  request.DaysChange,
			Status:      changeProposed,
			ProposedBy:  &actorID,
		}
		return changeErrors.Resolve(repository.CreateContractChange(ctx, &change))
	})
}

func (s *service) DecideContractChange(ctx context.Context, projectID, changeID, actorID uuid.UUID, role string, decision ContractChangeDecision) (PaymentScheduleResponse, error) {
	today := s.today()
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if role != access.RoleAdmin {
			return errChangeApprover
		}
		if project.Status == statusCancelled {
			return errChangeProjectCancelled
		}
		change, err := repository.FindContractChange(ctx, projectID, changeID)
		if err != nil {
			return changeErrors.Resolve(err)
		}
		if change.Status != changeProposed {
			return errChangeDecided
		}
		if decision.Status == changeApproved {
			if err := ensureApprovable(ctx, repository, projectID, change); err != nil {
				return err
			}
		}

		change.Status = decision.Status
		change.DecidedBy = &actorID
		change.DecidedOn = &today
		return changeErrors.Resolve(repository.SaveContractChange(ctx, &change))
	})
}

func (s *service) DeleteContractChange(ctx context.Context, projectID, changeID uuid.UUID) (PaymentScheduleResponse, error) {
	return s.changeSchedule(ctx, projectID, func(repository Repository, project Project) error {
		if err := ensureOpen(project); err != nil {
			return err
		}
		change, err := repository.FindContractChange(ctx, projectID, changeID)
		if err != nil {
			return changeErrors.Resolve(err)
		}
		if change.Status != changeProposed {
			return errChangeDecided
		}
		return changeErrors.Resolve(repository.DeleteContractChange(ctx, projectID, changeID))
	})
}

func ensureApprovable(ctx context.Context, repository Repository, projectID uuid.UUID, change ContractChange) error {
	contract, err := repository.FindContract(ctx, projectID)
	if err != nil {
		return contractErrors.Resolve(err)
	}
	changes, err := repository.ListContractChanges(ctx, projectID)
	if err != nil {
		return apperror.Internal(err)
	}
	terms, err := repository.ListPaymentTerms(ctx, projectID)
	if err != nil {
		return apperror.Internal(err)
	}

	next := contract.Value + sumApproved(changes).value + change.ValueChange
	if next <= 0 {
		return errChangeNonPositive
	}
	if next < claimedGross(terms) {
		return errChangeBelowClaimed
	}
	return nil
}

func sumApproved(changes []ContractChangeRow) approvedChanges {
	var total approvedChanges
	for _, change := range changes {
		if change.Status == changeApproved {
			total.value += change.ValueChange
			total.days += change.DaysChange
		}
	}
	return total
}

func nextChangeSequence(changes []ContractChangeRow) int {
	highest := 0
	for _, change := range changes {
		highest = max(highest, change.Sequence)
	}
	return highest + 1
}

func claimedGross(terms []PaymentTermRow) int64 {
	var total int64
	for _, term := range terms {
		if term.Kind == termProgress && term.PayableID != nil {
			total += term.GrossAmount
		}
	}
	return total
}

func newContractChangeResponses(changes []ContractChangeRow) []ContractChangeResponse {
	sorted := slices.Clone(changes)
	slices.SortFunc(sorted, func(a, b ContractChangeRow) int { return a.Sequence - b.Sequence })
	responses := make([]ContractChangeResponse, 0, len(sorted))
	for _, change := range sorted {
		responses = append(responses, ContractChangeResponse{
			ID:           change.ID,
			Label:        fmt.Sprintf("Adendum %d", change.Sequence),
			Description:  change.Description,
			ValueChange:  change.ValueChange,
			DaysChange:   change.DaysChange,
			Status:       change.Status,
			ProposerName: change.ProposerName,
			DeciderName:  change.DeciderName,
			DecidedOn:    change.DecidedOn,
			CreatedAt:    change.CreatedAt,
		})
	}
	return responses
}

const contractChangeColumns = "contract_change.id, contract_change.sequence, contract_change.description, " +
	"contract_change.value_change, contract_change.days_change, contract_change.status, " +
	"proposer.name AS proposer_name, decider.name AS decider_name, " +
	"contract_change.decided_on, contract_change.created_at"

func (r *gormRepository) ListContractChanges(ctx context.Context, projectID uuid.UUID) ([]ContractChangeRow, error) {
	rows := []ContractChangeRow{}
	err := r.db.WithContext(ctx).
		Table("contract_change").
		Select(contractChangeColumns).
		Joins("LEFT JOIN users AS proposer ON proposer.id = contract_change.proposed_by").
		Joins("LEFT JOIN users AS decider ON decider.id = contract_change.decided_by").
		Where("contract_change.project_id = ?", projectID).
		Order("contract_change.sequence").
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) FindContractChange(ctx context.Context, projectID, id uuid.UUID) (ContractChange, error) {
	var change ContractChange
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&change).Error
	return change, database.Translate(err)
}

func (r *gormRepository) CreateContractChange(ctx context.Context, change *ContractChange) error {
	return database.Translate(r.db.WithContext(ctx).Create(change).Error)
}

func (r *gormRepository) SaveContractChange(ctx context.Context, change *ContractChange) error {
	return database.Translate(r.db.WithContext(ctx).Save(change).Error)
}

func (r *gormRepository) DeleteContractChange(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &ContractChange{}, "id = ? AND project_id = ?", id, projectID)
}
