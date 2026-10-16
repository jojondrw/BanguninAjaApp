package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const projectTotalsColumns = "COUNT(*) AS count, " +
	"COUNT(*) FILTER (WHERE status = ?) AS planning, " +
	"COUNT(*) FILTER (WHERE status = ?) AS ongoing, " +
	"COUNT(*) FILTER (WHERE status = ?) AS on_hold, " +
	"COUNT(*) FILTER (WHERE status = ?) AS completed, " +
	"COUNT(*) FILTER (WHERE status = ?) AS cancelled, " +
	"COALESCE(SUM(contract_value), 0) AS contract_value, " +
	"COALESCE(ROUND(AVG(progress)), 0) AS average_progress"

const memberColumns = "project_member.id, project_member.project_id, project_member.user_id, " +
	"users.name, users.email, users.active, project_member.position, " +
	"role.code AS role_code, role.name AS role_name, project_member.created_at, project_member.updated_at"

type ProjectFilter struct {
	Search   string
	Status   string
	RegionID *uuid.UUID
	MemberID *uuid.UUID
	Offset   int
	Limit    int
}

type ProjectTotals struct {
	Count           int64
	Planning        int64
	Ongoing         int64
	OnHold          int64
	Completed       int64
	Cancelled       int64
	ContractValue   int64
	AverageProgress int64
}

type BudgetItemFilter struct {
	ProjectID uuid.UUID
	Search    string
	Offset    int
	Limit     int
}

type MemberRow struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Name      string
	Email     string
	Active    bool
	Position  *string
	RoleCode  string
	RoleName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemberUser struct {
	ID     uuid.UUID
	Active bool
	RoleID *uuid.UUID
}

type MemberCandidate struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type MemberCandidateFilter struct {
	ProjectID uuid.UUID
	Search    string
	Offset    int
	Limit     int
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, int64, error)
	SummarizeProjects(ctx context.Context) (ProjectTotals, error)
	FindProject(ctx context.Context, id uuid.UUID) (Project, error)
	LockProject(ctx context.Context, id uuid.UUID) (Project, error)
	CreateProject(ctx context.Context, project *Project) error
	SaveProject(ctx context.Context, project *Project) error
	UpdateProjectProgress(ctx context.Context, id uuid.UUID, progress int) error
	DeleteProject(ctx context.Context, id uuid.UUID) error

	ListPhases(ctx context.Context, projectID uuid.UUID) ([]ProjectPhase, error)
	FindPhase(ctx context.Context, projectID, id uuid.UUID) (ProjectPhase, error)
	CreatePhase(ctx context.Context, phase *ProjectPhase) error
	SavePhase(ctx context.Context, phase *ProjectPhase) error
	DeletePhase(ctx context.Context, projectID, id uuid.UUID) error

	ListBudgetItems(ctx context.Context, filter BudgetItemFilter) ([]BudgetItem, int64, error)
	SumBudgetItems(ctx context.Context, projectID uuid.UUID) (int64, error)
	SumBudgetByPhase(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]int64, error)
	RecordProgress(ctx context.Context, snapshot ProgressSnapshot) error
	ListProjectIDs(ctx context.Context) ([]uuid.UUID, error)
	ListProgressSnapshots(ctx context.Context, projectID uuid.UUID) ([]ProgressSnapshot, error)
	ListDailyReports(ctx context.Context, filter DailyReportFilter) ([]DailyReport, int64, error)
	CountRecentDailyReports(ctx context.Context, projectID uuid.UUID, since time.Time) (int64, int64, error)
	FindDailyReport(ctx context.Context, projectID, id uuid.UUID) (DailyReport, error)
	CreateDailyReport(ctx context.Context, report *DailyReport) error
	SaveDailyReport(ctx context.Context, report *DailyReport) error
	DeleteDailyReport(ctx context.Context, projectID, id uuid.UUID) error
	ListPaymentTerms(ctx context.Context, projectID uuid.UUID) ([]PaymentTermRow, error)
	FindPaymentTerm(ctx context.Context, projectID, id uuid.UUID) (PaymentTerm, error)
	CreatePaymentTerm(ctx context.Context, term *PaymentTerm) error
	SavePaymentTerm(ctx context.Context, term *PaymentTerm) error
	DeletePaymentTerm(ctx context.Context, projectID, id uuid.UUID) error
	FindContract(ctx context.Context, projectID uuid.UUID) (ConstructionContract, error)
	CreateContract(ctx context.Context, contract *ConstructionContract) error
	SaveContract(ctx context.Context, contract *ConstructionContract) error
	FindContractVendor(ctx context.Context, id uuid.UUID) (ContractVendor, error)
	Payables() PayableLedger
	ListContractChanges(ctx context.Context, projectID uuid.UUID) ([]ContractChangeRow, error)
	FindContractChange(ctx context.Context, projectID, id uuid.UUID) (ContractChange, error)
	CreateContractChange(ctx context.Context, change *ContractChange) error
	SaveContractChange(ctx context.Context, change *ContractChange) error
	DeleteContractChange(ctx context.Context, projectID, id uuid.UUID) error
	FindBudgetItem(ctx context.Context, projectID, id uuid.UUID) (BudgetItem, error)
	CreateBudgetItem(ctx context.Context, item *BudgetItem) error
	SaveBudgetItem(ctx context.Context, item *BudgetItem) error
	DeleteBudgetItem(ctx context.Context, projectID, id uuid.UUID) error

	ListPermits(ctx context.Context, projectID uuid.UUID) ([]Permit, error)
	FindPermit(ctx context.Context, projectID, id uuid.UUID) (Permit, error)
	CreatePermit(ctx context.Context, permit *Permit) error
	SavePermit(ctx context.Context, permit *Permit) error
	DeletePermit(ctx context.Context, projectID, id uuid.UUID) error

	ListMembers(ctx context.Context, projectID uuid.UUID) ([]MemberRow, error)
	FindMemberRow(ctx context.Context, projectID, id uuid.UUID) (MemberRow, error)
	FindMember(ctx context.Context, projectID, id uuid.UUID) (ProjectMember, error)
	CreateMember(ctx context.Context, member *ProjectMember) error
	UpdateMember(ctx context.Context, member *ProjectMember) error
	DeleteMember(ctx context.Context, projectID, id uuid.UUID) error
	FindMemberUser(ctx context.Context, userID uuid.UUID) (MemberUser, error)
	FindRoleID(ctx context.Context, code string) (uuid.UUID, error)
	ListMemberCandidates(ctx context.Context, filter MemberCandidateFilter) ([]MemberCandidate, int64, error)
}

type gormRepository struct {
	db       *gorm.DB
	payables PayableLedgerFactory
}

func NewRepository(db *gorm.DB, payables PayableLedgerFactory) Repository {
	return &gormRepository{db: db, payables: payables}
}

func (r *gormRepository) Transaction(ctx context.Context, work func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(&gormRepository{db: tx, payables: r.payables})
	})
}

func (r *gormRepository) ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, int64, error) {
	return database.FindPage[Project](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "created_at DESC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) SummarizeProjects(ctx context.Context) (ProjectTotals, error) {
	var totals ProjectTotals
	err := r.db.WithContext(ctx).
		Model(&Project{}).
		Select(projectTotalsColumns, statusPlanning, statusOngoing, statusOnHold, statusCompleted, statusCancelled).
		Scan(&totals).Error
	return totals, database.Translate(err)
}

func (r *gormRepository) FindProject(ctx context.Context, id uuid.UUID) (Project, error) {
	var project Project
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&project).Error
	return project, database.Translate(err)
}

func (r *gormRepository) LockProject(ctx context.Context, id uuid.UUID) (Project, error) {
	var project Project
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("id = ?", id).
		Take(&project).Error
	return project, database.Translate(err)
}

func (r *gormRepository) CreateProject(ctx context.Context, project *Project) error {
	return database.Translate(r.db.WithContext(ctx).Create(project).Error)
}

func (r *gormRepository) SaveProject(ctx context.Context, project *Project) error {
	return database.Translate(r.db.WithContext(ctx).Save(project).Error)
}

func (r *gormRepository) UpdateProjectProgress(ctx context.Context, id uuid.UUID, progress int) error {
	err := r.db.WithContext(ctx).Model(&Project{}).Where("id = ?", id).Update("progress", progress).Error
	return database.Translate(err)
}

func (r *gormRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Project{}, "id = ?", id)
}

func (r *gormRepository) ListPhases(ctx context.Context, projectID uuid.UUID) ([]ProjectPhase, error) {
	var phases []ProjectPhase
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("sort_order ASC, created_at ASC").
		Find(&phases).Error
	return phases, database.Translate(err)
}

func (r *gormRepository) FindPhase(ctx context.Context, projectID, id uuid.UUID) (ProjectPhase, error) {
	var phase ProjectPhase
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&phase).Error
	return phase, database.Translate(err)
}

func (r *gormRepository) CreatePhase(ctx context.Context, phase *ProjectPhase) error {
	return database.Translate(r.db.WithContext(ctx).Create(phase).Error)
}

func (r *gormRepository) SavePhase(ctx context.Context, phase *ProjectPhase) error {
	return database.Translate(r.db.WithContext(ctx).Save(phase).Error)
}

func (r *gormRepository) DeletePhase(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &ProjectPhase{}, "id = ? AND project_id = ?", id, projectID)
}

func (r *gormRepository) ListBudgetItems(ctx context.Context, filter BudgetItemFilter) ([]BudgetItem, int64, error) {
	return database.FindPage[BudgetItem](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) SumBudgetItems(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&BudgetItem{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(SUM(total), 0)").
		Scan(&total).Error
	return total, database.Translate(err)
}

func (r *gormRepository) SumBudgetByPhase(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]int64, error) {
	var rows []struct {
		PhaseID uuid.UUID
		Total   int64
	}
	err := r.db.WithContext(ctx).
		Model(&BudgetItem{}).
		Select("phase_id, COALESCE(SUM(total), 0) AS total").
		Where("project_id = ? AND phase_id IS NOT NULL", projectID).
		Group("phase_id").
		Scan(&rows).Error
	if err != nil {
		return nil, database.Translate(err)
	}
	totals := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		totals[row.PhaseID] = row.Total
	}
	return totals, nil
}

func (r *gormRepository) RecordProgress(ctx context.Context, snapshot ProgressSnapshot) error {
	return database.Translate(r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "recorded_on"}},
		DoUpdates: clause.Assignments(map[string]any{"progress": snapshot.Progress, "updated_at": time.Now()}),
	}).Create(&snapshot).Error)
}

func (r *gormRepository) ListProjectIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&Project{}).Order("code").Pluck("id", &ids).Error
	return ids, database.Translate(err)
}

func (r *gormRepository) ListProgressSnapshots(ctx context.Context, projectID uuid.UUID) ([]ProgressSnapshot, error) {
	var snapshots []ProgressSnapshot
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("recorded_on").
		Find(&snapshots).Error
	return snapshots, database.Translate(err)
}

func (r *gormRepository) FindBudgetItem(ctx context.Context, projectID, id uuid.UUID) (BudgetItem, error) {
	var item BudgetItem
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&item).Error
	return item, database.Translate(err)
}

func (r *gormRepository) CreateBudgetItem(ctx context.Context, item *BudgetItem) error {
	return database.Translate(r.db.WithContext(ctx).Create(item).Error)
}

func (r *gormRepository) SaveBudgetItem(ctx context.Context, item *BudgetItem) error {
	return database.Translate(r.db.WithContext(ctx).Save(item).Error)
}

func (r *gormRepository) DeleteBudgetItem(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &BudgetItem{}, "id = ? AND project_id = ?", id, projectID)
}

func (r *gormRepository) ListPermits(ctx context.Context, projectID uuid.UUID) ([]Permit, error) {
	var permits []Permit
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("valid_until ASC NULLS LAST, created_at ASC").
		Find(&permits).Error
	return permits, database.Translate(err)
}

func (r *gormRepository) FindPermit(ctx context.Context, projectID, id uuid.UUID) (Permit, error) {
	var permit Permit
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&permit).Error
	return permit, database.Translate(err)
}

func (r *gormRepository) CreatePermit(ctx context.Context, permit *Permit) error {
	return database.Translate(r.db.WithContext(ctx).Create(permit).Error)
}

func (r *gormRepository) SavePermit(ctx context.Context, permit *Permit) error {
	return database.Translate(r.db.WithContext(ctx).Save(permit).Error)
}

func (r *gormRepository) DeletePermit(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Permit{}, "id = ? AND project_id = ?", id, projectID)
}

func (r *gormRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]MemberRow, error) {
	rows := []MemberRow{}
	err := r.memberRows(ctx).
		Where("project_member.project_id = ?", projectID).
		Order("users.name ASC, project_member.id ASC").
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) FindMemberRow(ctx context.Context, projectID, id uuid.UUID) (MemberRow, error) {
	var rows []MemberRow
	err := r.memberRows(ctx).
		Where("project_member.id = ? AND project_member.project_id = ?", id, projectID).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return MemberRow{}, database.Translate(err)
	}
	if len(rows) == 0 {
		return MemberRow{}, database.ErrNotFound
	}
	return rows[0], nil
}

func (r *gormRepository) FindMember(ctx context.Context, projectID, id uuid.UUID) (ProjectMember, error) {
	var member ProjectMember
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&member).Error
	return member, database.Translate(err)
}

func (r *gormRepository) CreateMember(ctx context.Context, member *ProjectMember) error {
	return database.Translate(r.db.WithContext(ctx).Create(member).Error)
}

func (r *gormRepository) UpdateMember(ctx context.Context, member *ProjectMember) error {
	updatedAt := time.Now()
	result := r.db.WithContext(ctx).
		Model(&ProjectMember{}).
		Where("id = ? AND project_id = ?", member.ID, member.ProjectID).
		Updates(map[string]any{"role_id": member.RoleID, "position": member.Position, "updated_at": updatedAt})
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	member.UpdatedAt = updatedAt
	return nil
}

func (r *gormRepository) DeleteMember(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &ProjectMember{}, "id = ? AND project_id = ?", id, projectID)
}

func (r *gormRepository) FindMemberUser(ctx context.Context, userID uuid.UUID) (MemberUser, error) {
	var users []MemberUser
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.id, users.active, users.role_id").
		Where("users.id = ?", userID).
		Limit(1).
		Scan(&users).Error
	if err != nil {
		return MemberUser{}, database.Translate(err)
	}
	if len(users) == 0 {
		return MemberUser{}, database.ErrNotFound
	}
	return users[0], nil
}

func (r *gormRepository) FindRoleID(ctx context.Context, code string) (uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("role").Where("role.code = ?", code).Limit(1).Pluck("role.id", &ids).Error
	if err != nil {
		return uuid.Nil, database.Translate(err)
	}
	if len(ids) == 0 {
		return uuid.Nil, database.ErrNotFound
	}
	return ids[0], nil
}

func (r *gormRepository) ListMemberCandidates(ctx context.Context, filter MemberCandidateFilter) ([]MemberCandidate, int64, error) {
	var total int64
	if err := r.memberCandidates(ctx, filter).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	candidates := make([]MemberCandidate, 0, filter.Limit)
	err := r.memberCandidates(ctx, filter).
		Select("users.id, users.name, users.email").
		Order("users.name ASC, users.id ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&candidates).Error
	return candidates, total, database.Translate(err)
}

func (r *gormRepository) memberRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("project_member").
		Select(memberColumns).
		Joins("JOIN users ON users.id = project_member.user_id").
		Joins("JOIN role ON role.id = project_member.role_id")
}

func (r *gormRepository) memberCandidates(ctx context.Context, filter MemberCandidateFilter) *gorm.DB {
	return r.db.WithContext(ctx).Table("users").Scopes(filter.apply)
}

func (r *gormRepository) deleteWhere(ctx context.Context, model any, condition string, args ...any) error {
	result := r.db.WithContext(ctx).Where(condition, args...).Delete(model)
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (f ProjectFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if f.RegionID != nil {
		db = db.Where("region_id = ?", *f.RegionID)
	}
	if f.MemberID != nil {
		db = db.Where("id IN (SELECT project_id FROM project_member WHERE user_id = ?)", *f.MemberID)
	}
	return db
}

func (f BudgetItemFilter) apply(db *gorm.DB) *gorm.DB {
	db = db.Where("project_id = ?", f.ProjectID)
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(description ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	return db
}

func (f MemberCandidateFilter) apply(db *gorm.DB) *gorm.DB {
	db = db.Where("users.active = ?", true).
		Where("NOT EXISTS (SELECT 1 FROM project_member WHERE project_member.user_id = users.id AND project_member.project_id = ?)", f.ProjectID)
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(users.name ILIKE ? OR users.email ILIKE ?)", pattern, pattern)
	}
	return db
}
