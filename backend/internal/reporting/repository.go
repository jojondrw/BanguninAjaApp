package reporting

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type ReportFilter struct {
	Search    string
	Type      string
	Format    string
	ProjectID *uuid.UUID
	Offset    int
	Limit     int
}

type Repository interface {
	ListReports(ctx context.Context, filter ReportFilter) ([]Report, int64, error)
	FindReport(ctx context.Context, id uuid.UUID) (Report, error)
	CreateReport(ctx context.Context, report *Report) error
	DeleteReport(ctx context.Context, id uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListReports(ctx context.Context, filter ReportFilter) ([]Report, int64, error) {
	return database.FindPage[Report](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "created_at DESC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindReport(ctx context.Context, id uuid.UUID) (Report, error) {
	var report Report
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&report).Error
	return report, database.Translate(err)
}

func (r *gormRepository) CreateReport(ctx context.Context, report *Report) error {
	return database.Translate(r.db.WithContext(ctx).Create(report).Error)
}

func (r *gormRepository) DeleteReport(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Report{})
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (f ReportFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		db = db.Where("name ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.Type != "" {
		db = db.Where("type = ?", f.Type)
	}
	if f.Format != "" {
		db = db.Where("format = ?", f.Format)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	return db
}
