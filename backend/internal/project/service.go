package project

import (
	"context"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	statusPlanning  = "planning"
	statusOngoing   = "ongoing"
	statusOnHold    = "on_hold"
	statusCompleted = "completed"
	statusCancelled = "cancelled"

	phaseNotStarted = "not_started"
	phaseInProgress = "in_progress"
	phaseCompleted  = "completed"

	permitIssued    = "issued"
	fullProgress    = 100
	volumePrecision = 100
)

var projectTransitions = map[string][]string{
	statusPlanning: {statusOngoing, statusCancelled},
	statusOngoing:  {statusOnHold, statusCompleted, statusCancelled},
	statusOnHold:   {statusOngoing, statusCancelled},
}

var (
	errProjectNotFound          = apperror.NotFound("project_not_found", "Proyek tidak ditemukan")
	errProjectCodeUsed          = apperror.Conflict("project_code_used", "Kode proyek sudah dipakai")
	errProjectInUse             = apperror.Conflict("project_in_use", "Proyek masih dipakai oleh data lain")
	errProjectClosed            = apperror.Conflict("project_closed", "Proyek sudah selesai atau dibatalkan sehingga tidak bisa diubah")
	errProjectTransition        = apperror.Unprocessable("project_status_transition_invalid", "Perubahan status proyek tidak diizinkan")
	errProjectUnfinished        = apperror.Unprocessable("project_unfinished", "Proyek belum bisa diselesaikan karena masih ada tahap yang belum 100%")
	errRegionNotFound           = apperror.Unprocessable("region_not_found", "Wilayah tidak ditemukan")
	errInvalidDateRange         = apperror.Unprocessable("invalid_date_range", "Tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	errPhaseNotFound            = apperror.NotFound("phase_not_found", "Tahap proyek tidak ditemukan")
	errBudgetItemNotFound       = apperror.NotFound("budget_item_not_found", "Item RAB tidak ditemukan")
	errBudgetItemCodeUsed       = apperror.Conflict("budget_item_code_used", "Kode item RAB sudah dipakai di proyek ini")
	errUnitOfMeasureNotFound    = apperror.Unprocessable("unit_of_measure_not_found", "Satuan tidak ditemukan")
	errPermitNotFound           = apperror.NotFound("permit_not_found", "Izin tidak ditemukan")
	errPermitIssuedDateRequired = apperror.Unprocessable("permit_issued_date_required", "Izin yang sudah terbit wajib punya tanggal terbit")
)

var (
	projectReadErrors   = database.ErrorMap{NotFound: errProjectNotFound}
	projectWriteErrors  = database.ErrorMap{NotFound: errProjectNotFound, Duplicate: errProjectCodeUsed, Referenced: errRegionNotFound}
	projectDeleteErrors = database.ErrorMap{NotFound: errProjectNotFound, Referenced: errProjectInUse}
	phaseErrors         = database.ErrorMap{NotFound: errPhaseNotFound}
	budgetItemErrors    = database.ErrorMap{NotFound: errBudgetItemNotFound, Duplicate: errBudgetItemCodeUsed, Referenced: errUnitOfMeasureNotFound}
	permitErrors        = database.ErrorMap{NotFound: errPermitNotFound, Referenced: errProjectNotFound}
)

type Service interface {
	ListProjects(ctx context.Context, query ProjectQuery) (pagination.Page[ProjectResponse], error)
	GetProject(ctx context.Context, id uuid.UUID) (ProjectResponse, error)
	CreateProject(ctx context.Context, request ProjectRequest) (ProjectResponse, error)
	UpdateProject(ctx context.Context, id uuid.UUID, request ProjectRequest) (ProjectResponse, error)
	UpdateProjectStatus(ctx context.Context, id uuid.UUID, request ProjectStatusRequest) (ProjectResponse, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error

	ListPhases(ctx context.Context, projectID uuid.UUID) ([]PhaseResponse, error)
	CreatePhase(ctx context.Context, projectID uuid.UUID, request PhaseRequest) (PhaseResponse, error)
	UpdatePhase(ctx context.Context, projectID, phaseID uuid.UUID, request PhaseRequest) (PhaseResponse, error)
	DeletePhase(ctx context.Context, projectID, phaseID uuid.UUID) error

	ListBudgetItems(ctx context.Context, projectID uuid.UUID, query BudgetItemQuery) (BudgetItemPage, error)
	CreateBudgetItem(ctx context.Context, projectID uuid.UUID, request BudgetItemRequest) (BudgetItemResponse, error)
	UpdateBudgetItem(ctx context.Context, projectID, itemID uuid.UUID, request BudgetItemRequest) (BudgetItemResponse, error)
	DeleteBudgetItem(ctx context.Context, projectID, itemID uuid.UUID) error

	ListPermits(ctx context.Context, projectID uuid.UUID) ([]PermitResponse, error)
	CreatePermit(ctx context.Context, projectID uuid.UUID, request PermitRequest) (PermitResponse, error)
	UpdatePermit(ctx context.Context, projectID, permitID uuid.UUID, request PermitRequest) (PermitResponse, error)
	DeletePermit(ctx context.Context, projectID, permitID uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListProjects(ctx context.Context, query ProjectQuery) (pagination.Page[ProjectResponse], error) {
	projects, total, err := s.repository.ListProjects(ctx, ProjectFilter{
		Search:   query.Search,
		Status:   query.Status,
		RegionID: query.RegionID,
		Offset:   query.Offset(),
		Limit:    query.Size(),
	})
	if err != nil {
		return pagination.Page[ProjectResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(projects, newProjectResponse), query.Query, total), nil
}

func (s *service) GetProject(ctx context.Context, id uuid.UUID) (ProjectResponse, error) {
	project, err := s.repository.FindProject(ctx, id)
	if err != nil {
		return ProjectResponse{}, projectReadErrors.Resolve(err)
	}
	return newProjectResponse(project), nil
}

func (s *service) CreateProject(ctx context.Context, request ProjectRequest) (ProjectResponse, error) {
	if !dateRangeValid(request.StartDate, request.TargetEndDate) {
		return ProjectResponse{}, errInvalidDateRange
	}

	project := Project{Status: statusPlanning}
	applyProjectRequest(&project, request)
	if err := s.repository.CreateProject(ctx, &project); err != nil {
		return ProjectResponse{}, projectWriteErrors.Resolve(err)
	}
	return newProjectResponse(project), nil
}

func (s *service) UpdateProject(ctx context.Context, id uuid.UUID, request ProjectRequest) (ProjectResponse, error) {
	if !dateRangeValid(request.StartDate, request.TargetEndDate) {
		return ProjectResponse{}, errInvalidDateRange
	}

	project, err := s.repository.FindProject(ctx, id)
	if err != nil {
		return ProjectResponse{}, projectReadErrors.Resolve(err)
	}

	applyProjectRequest(&project, request)
	if err := s.repository.SaveProject(ctx, &project); err != nil {
		return ProjectResponse{}, projectWriteErrors.Resolve(err)
	}
	return newProjectResponse(project), nil
}

func (s *service) UpdateProjectStatus(ctx context.Context, id uuid.UUID, request ProjectStatusRequest) (ProjectResponse, error) {
	var project Project
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		var err error
		project, err = changeProjectStatus(ctx, repository, id, request.Status)
		return err
	})
	if err != nil {
		return ProjectResponse{}, apperror.From(err)
	}
	return newProjectResponse(project), nil
}

func (s *service) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return projectDeleteErrors.Resolve(s.repository.DeleteProject(ctx, id))
}

func (s *service) ListPhases(ctx context.Context, projectID uuid.UUID) ([]PhaseResponse, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return nil, projectReadErrors.Resolve(err)
	}

	phases, err := s.repository.ListPhases(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(phases, newPhaseResponse), nil
}

func (s *service) CreatePhase(ctx context.Context, projectID uuid.UUID, request PhaseRequest) (PhaseResponse, error) {
	if !dateRangeValid(request.StartDate, request.TargetEndDate) {
		return PhaseResponse{}, errInvalidDateRange
	}

	phase := ProjectPhase{ProjectID: projectID}
	applyPhaseRequest(&phase, request)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createPhase(ctx, repository, &phase)
	})
	if err != nil {
		return PhaseResponse{}, apperror.From(err)
	}
	return newPhaseResponse(phase), nil
}

func (s *service) UpdatePhase(ctx context.Context, projectID, phaseID uuid.UUID, request PhaseRequest) (PhaseResponse, error) {
	if !dateRangeValid(request.StartDate, request.TargetEndDate) {
		return PhaseResponse{}, errInvalidDateRange
	}

	var phase ProjectPhase
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		var err error
		phase, err = updatePhase(ctx, repository, projectID, phaseID, request)
		return err
	})
	if err != nil {
		return PhaseResponse{}, apperror.From(err)
	}
	return newPhaseResponse(phase), nil
}

func (s *service) DeletePhase(ctx context.Context, projectID, phaseID uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return deletePhase(ctx, repository, projectID, phaseID)
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) ListBudgetItems(ctx context.Context, projectID uuid.UUID, query BudgetItemQuery) (BudgetItemPage, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return BudgetItemPage{}, projectReadErrors.Resolve(err)
	}

	items, total, err := s.repository.ListBudgetItems(ctx, BudgetItemFilter{
		ProjectID: projectID,
		Search:    query.Search,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return BudgetItemPage{}, apperror.Internal(err)
	}

	grandTotal, err := s.repository.SumBudgetItems(ctx, projectID)
	if err != nil {
		return BudgetItemPage{}, apperror.Internal(err)
	}
	return BudgetItemPage{
		Page:       pagination.New(pagination.Map(items, newBudgetItemResponse), query.Query, total),
		GrandTotal: grandTotal,
	}, nil
}

func (s *service) CreateBudgetItem(ctx context.Context, projectID uuid.UUID, request BudgetItemRequest) (BudgetItemResponse, error) {
	if err := s.ensureOpenProject(ctx, projectID); err != nil {
		return BudgetItemResponse{}, err
	}

	item := BudgetItem{ProjectID: projectID}
	applyBudgetItemRequest(&item, request)
	if err := s.repository.CreateBudgetItem(ctx, &item); err != nil {
		return BudgetItemResponse{}, budgetItemErrors.Resolve(err)
	}
	return newBudgetItemResponse(item), nil
}

func (s *service) UpdateBudgetItem(ctx context.Context, projectID, itemID uuid.UUID, request BudgetItemRequest) (BudgetItemResponse, error) {
	if err := s.ensureOpenProject(ctx, projectID); err != nil {
		return BudgetItemResponse{}, err
	}

	item, err := s.repository.FindBudgetItem(ctx, projectID, itemID)
	if err != nil {
		return BudgetItemResponse{}, budgetItemErrors.Resolve(err)
	}

	applyBudgetItemRequest(&item, request)
	if err := s.repository.SaveBudgetItem(ctx, &item); err != nil {
		return BudgetItemResponse{}, budgetItemErrors.Resolve(err)
	}
	return newBudgetItemResponse(item), nil
}

func (s *service) DeleteBudgetItem(ctx context.Context, projectID, itemID uuid.UUID) error {
	if err := s.ensureOpenProject(ctx, projectID); err != nil {
		return err
	}
	return budgetItemErrors.Resolve(s.repository.DeleteBudgetItem(ctx, projectID, itemID))
}

func (s *service) ListPermits(ctx context.Context, projectID uuid.UUID) ([]PermitResponse, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return nil, projectReadErrors.Resolve(err)
	}

	permits, err := s.repository.ListPermits(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(permits, newPermitResponse), nil
}

func (s *service) CreatePermit(ctx context.Context, projectID uuid.UUID, request PermitRequest) (PermitResponse, error) {
	if err := validatePermit(request); err != nil {
		return PermitResponse{}, err
	}

	permit := Permit{ProjectID: projectID}
	applyPermitRequest(&permit, request)
	if err := s.repository.CreatePermit(ctx, &permit); err != nil {
		return PermitResponse{}, permitErrors.Resolve(err)
	}
	return newPermitResponse(permit), nil
}

func (s *service) UpdatePermit(ctx context.Context, projectID, permitID uuid.UUID, request PermitRequest) (PermitResponse, error) {
	if err := validatePermit(request); err != nil {
		return PermitResponse{}, err
	}

	permit, err := s.repository.FindPermit(ctx, projectID, permitID)
	if err != nil {
		return PermitResponse{}, permitErrors.Resolve(err)
	}

	applyPermitRequest(&permit, request)
	if err := s.repository.SavePermit(ctx, &permit); err != nil {
		return PermitResponse{}, permitErrors.Resolve(err)
	}
	return newPermitResponse(permit), nil
}

func (s *service) DeletePermit(ctx context.Context, projectID, permitID uuid.UUID) error {
	return permitErrors.Resolve(s.repository.DeletePermit(ctx, projectID, permitID))
}

func (s *service) ensureOpenProject(ctx context.Context, id uuid.UUID) error {
	project, err := s.repository.FindProject(ctx, id)
	if err != nil {
		return projectReadErrors.Resolve(err)
	}
	return ensureOpen(project)
}

func changeProjectStatus(ctx context.Context, repository Repository, id uuid.UUID, next string) (Project, error) {
	project, err := repository.LockProject(ctx, id)
	if err != nil {
		return Project{}, projectReadErrors.Resolve(err)
	}
	if project.Status == next {
		return project, nil
	}
	if !canTransitionProject(project.Status, next) {
		return Project{}, errProjectTransition
	}
	if next == statusCompleted {
		if err := completeProject(ctx, repository, &project); err != nil {
			return Project{}, err
		}
	}

	project.Status = next
	if err := repository.SaveProject(ctx, &project); err != nil {
		return Project{}, projectWriteErrors.Resolve(err)
	}
	return project, nil
}

func completeProject(ctx context.Context, repository Repository, project *Project) error {
	phases, err := repository.ListPhases(ctx, project.ID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !phasesFinished(phases) {
		return errProjectUnfinished
	}
	project.Progress = fullProgress
	return nil
}

func createPhase(ctx context.Context, repository Repository, phase *ProjectPhase) error {
	if err := lockOpenProject(ctx, repository, phase.ProjectID); err != nil {
		return err
	}
	if err := repository.CreatePhase(ctx, phase); err != nil {
		return phaseErrors.Resolve(err)
	}
	return refreshProjectProgress(ctx, repository, phase.ProjectID)
}

func updatePhase(ctx context.Context, repository Repository, projectID, phaseID uuid.UUID, request PhaseRequest) (ProjectPhase, error) {
	if err := lockOpenProject(ctx, repository, projectID); err != nil {
		return ProjectPhase{}, err
	}

	phase, err := repository.FindPhase(ctx, projectID, phaseID)
	if err != nil {
		return ProjectPhase{}, phaseErrors.Resolve(err)
	}

	applyPhaseRequest(&phase, request)
	if err := repository.SavePhase(ctx, &phase); err != nil {
		return ProjectPhase{}, phaseErrors.Resolve(err)
	}
	return phase, refreshProjectProgress(ctx, repository, projectID)
}

func deletePhase(ctx context.Context, repository Repository, projectID, phaseID uuid.UUID) error {
	if err := lockOpenProject(ctx, repository, projectID); err != nil {
		return err
	}
	if err := repository.DeletePhase(ctx, projectID, phaseID); err != nil {
		return phaseErrors.Resolve(err)
	}
	return refreshProjectProgress(ctx, repository, projectID)
}

func lockOpenProject(ctx context.Context, repository Repository, id uuid.UUID) error {
	project, err := repository.LockProject(ctx, id)
	if err != nil {
		return projectReadErrors.Resolve(err)
	}
	return ensureOpen(project)
}

func refreshProjectProgress(ctx context.Context, repository Repository, projectID uuid.UUID) error {
	phases, err := repository.ListPhases(ctx, projectID)
	if err != nil {
		return apperror.Internal(err)
	}
	if err := repository.UpdateProjectProgress(ctx, projectID, averageProgress(phases)); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func ensureOpen(project Project) error {
	if project.Status == statusCompleted || project.Status == statusCancelled {
		return errProjectClosed
	}
	return nil
}

func canTransitionProject(current, next string) bool {
	return slices.Contains(projectTransitions[current], next)
}

func phasesFinished(phases []ProjectPhase) bool {
	for _, phase := range phases {
		if phase.Progress < fullProgress {
			return false
		}
	}
	return true
}

func averageProgress(phases []ProjectPhase) int {
	if len(phases) == 0 {
		return 0
	}

	sum := 0
	for _, phase := range phases {
		sum += phase.Progress
	}
	return int(math.Round(float64(sum) / float64(len(phases))))
}

func phaseStatus(progress int) string {
	switch {
	case progress <= 0:
		return phaseNotStarted
	case progress >= fullProgress:
		return phaseCompleted
	default:
		return phaseInProgress
	}
}

func budgetTotal(volume float64, unitPrice int64) int64 {
	return int64(math.Round(volume * float64(unitPrice)))
}

func roundVolume(volume float64) float64 {
	return math.Round(volume*volumePrecision) / volumePrecision
}

func dateRangeValid(start, end *time.Time) bool {
	return start == nil || end == nil || !end.Before(*start)
}

func validatePermit(request PermitRequest) error {
	if request.Status == permitIssued && request.IssuedDate == nil {
		return errPermitIssuedDateRequired
	}
	if !dateRangeValid(request.IssuedDate, request.ValidUntil) {
		return errInvalidDateRange
	}
	return nil
}

func applyProjectRequest(project *Project, request ProjectRequest) {
	project.Code = strings.TrimSpace(request.Code)
	project.Name = strings.TrimSpace(request.Name)
	project.RegionID = request.RegionID
	project.Type = strings.TrimSpace(request.Type)
	project.StartDate = request.StartDate
	project.TargetEndDate = request.TargetEndDate
	project.ContractValue = request.ContractValue
}

func applyPhaseRequest(phase *ProjectPhase, request PhaseRequest) {
	phase.Name = strings.TrimSpace(request.Name)
	phase.SortOrder = request.SortOrder
	phase.StartDate = request.StartDate
	phase.TargetEndDate = request.TargetEndDate
	phase.Progress = request.Progress
	phase.Status = phaseStatus(request.Progress)
}

func applyBudgetItemRequest(item *BudgetItem, request BudgetItemRequest) {
	item.Code = strings.TrimSpace(request.Code)
	item.Description = strings.TrimSpace(request.Description)
	item.Volume = roundVolume(request.Volume)
	item.UnitOfMeasureID = request.UnitOfMeasureID
	item.UnitPrice = request.UnitPrice
	item.Total = budgetTotal(item.Volume, request.UnitPrice)
}

func applyPermitRequest(permit *Permit, request PermitRequest) {
	permit.Type = strings.TrimSpace(request.Type)
	permit.Number = strings.TrimSpace(request.Number)
	permit.IssuedDate = request.IssuedDate
	permit.ValidUntil = request.ValidUntil
	permit.Status = request.Status
}
