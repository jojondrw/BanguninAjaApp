package reporting

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const reportColumns = "report.id, report.name, report.type, report.scope, report.project_id, report.period_start, report.period_end, " +
	"report.format, report.created_by, report.created_at, report.updated_at, " +
	"(SELECT project.code FROM project WHERE project.id = report.project_id) AS project_code, " +
	"(SELECT project.name FROM project WHERE project.id = report.project_id) AS project_name, " +
	"(SELECT users.name FROM users WHERE users.id = report.created_by) AS created_by_name"

type ReportFilter struct {
	Search    string
	Type      string
	Types     []string
	Format    string
	ProjectID *uuid.UUID
	Offset    int
	Limit     int
}

type ReportRow struct {
	Report
	ProjectCode   string
	ProjectName   string
	CreatedByName string
}

type Repository interface {
	ListReports(ctx context.Context, filter ReportFilter) ([]ReportRow, int64, error)
	FindReport(ctx context.Context, id uuid.UUID) (ReportRow, error)
	CreateReport(ctx context.Context, report *Report) error
	DeleteReport(ctx context.Context, id uuid.UUID) error

	FindProjectLabel(ctx context.Context, id uuid.UUID) (ProjectLabel, error)
	SumCashByMonth(ctx context.Context, filter DatasetFilter) ([]CashMonthRow, error)
	CashBalanceBefore(ctx context.Context, projectID *uuid.UUID, before time.Time) (int64, error)
	ListCashLines(ctx context.Context, filter DatasetFilter) ([]CashLineRow, error)
	ListBudgetRealization(ctx context.Context, filter DatasetFilter) ([]BudgetRow, error)
	ListProjectProgress(ctx context.Context, filter DatasetFilter) ([]ProjectRow, error)
	ListOpenReceivables(ctx context.Context, filter DatasetFilter) ([]OpenItemRow, error)
	ListOpenPayables(ctx context.Context, filter DatasetFilter) ([]OpenItemRow, error)
	ListStockLevels(ctx context.Context, filter DatasetFilter) ([]StockRow, error)
	ListLowStock(ctx context.Context, filter DatasetFilter) ([]LowStockRow, error)
	ListPayrolls(ctx context.Context, filter DatasetFilter) ([]PayrollRow, error)
	ListContracts(ctx context.Context, filter DatasetFilter, today time.Time) ([]ContractRow, error)
	ListInstallments(ctx context.Context, filter DatasetFilter) ([]InstallmentRow, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListReports(ctx context.Context, filter ReportFilter) ([]ReportRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Report{}).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}
	rows := make([]ReportRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).Table("report").
		Select(reportColumns).
		Scopes(filter.apply).
		Order("report.created_at DESC, report.id ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindReport(ctx context.Context, id uuid.UUID) (ReportRow, error) {
	var row ReportRow
	result := r.db.WithContext(ctx).Table("report").Select(reportColumns).Where("report.id = ?", id).Limit(1).Scan(&row)
	if result.Error != nil {
		return ReportRow{}, database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return ReportRow{}, database.ErrNotFound
	}
	return row, nil
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
		db = db.Where("report.name ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.Type != "" {
		db = db.Where("report.type = ?", f.Type)
	}
	if f.Format != "" {
		db = db.Where("report.format = ?", f.Format)
	}
	if f.ProjectID != nil {
		db = db.Where("report.project_id = ?", *f.ProjectID)
	}
	return db.Where("report.type IN ?", f.Types)
}
