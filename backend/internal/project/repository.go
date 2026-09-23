package project

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type ProjectFilter struct {
	Search   string
	Status   string
	RegionID *uuid.UUID
	Offset   int
	Limit    int
}

type BudgetItemFilter struct {
	ProjectID uuid.UUID
	Search    string
	Offset    int
	Limit     int
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, int64, error)
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
	FindBudgetItem(ctx context.Context, projectID, id uuid.UUID) (BudgetItem, error)
	CreateBudgetItem(ctx context.Context, item *BudgetItem) error
	SaveBudgetItem(ctx context.Context, item *BudgetItem) error
	DeleteBudgetItem(ctx context.Context, projectID, id uuid.UUID) error

	ListPermits(ctx context.Context, projectID uuid.UUID) ([]Permit, error)
	FindPermit(ctx context.Context, projectID, id uuid.UUID) (Permit, error)
	CreatePermit(ctx context.Context, permit *Permit) error
	SavePermit(ctx context.Context, permit *Permit) error
	DeletePermit(ctx context.Context, projectID, id uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Transaction(ctx context.Context, work func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(&gormRepository{db: tx})
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
