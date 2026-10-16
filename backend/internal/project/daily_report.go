package project

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const recentWindowDays = 30

var rainyWeather = []string{"rain", "heavy_rain"}

var (
	errDailyReportNotFound = apperror.NotFound("daily_report_not_found", "Laporan harian tidak ditemukan")
	errDailyReportDateUsed = apperror.Conflict("daily_report_date_used", "Laporan untuk tanggal ini sudah ada. Ubah laporan tersebut")
	errDailyReportFuture   = apperror.Unprocessable("daily_report_in_future", "Laporan harian tidak bisa dibuat untuk tanggal yang belum terjadi")

	dailyReportErrors = database.ErrorMap{NotFound: errDailyReportNotFound, Duplicate: errDailyReportDateUsed, Referenced: errProjectNotFound}
)

type DailyReport struct {
	entity.Base
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_daily_report_project"`
	ReportDate       time.Time  `gorm:"type:date;not null"`
	WeatherMorning   string     `gorm:"type:varchar(20);not null"`
	WeatherAfternoon string     `gorm:"type:varchar(20);not null"`
	Workers          int        `gorm:"not null;default:0"`
	Equipment        string     `gorm:"type:varchar(300);not null;default:''"`
	WorkDone         string     `gorm:"type:text;not null"`
	Issues           string     `gorm:"type:text;not null;default:''"`
	CreatedBy        *uuid.UUID `gorm:"type:uuid;index:idx_daily_report_created_by"`
}

func (DailyReport) TableName() string {
	return "daily_report"
}

type DailyReportRequest struct {
	ReportDate       *time.Time `json:"reportDate" binding:"required"`
	WeatherMorning   string     `json:"weatherMorning" binding:"required,oneof=sunny cloudy drizzle rain heavy_rain"`
	WeatherAfternoon string     `json:"weatherAfternoon" binding:"required,oneof=sunny cloudy drizzle rain heavy_rain"`
	Workers          *int       `json:"workers" binding:"required,min=0,max=10000"`
	Equipment        string     `json:"equipment" binding:"max=300"`
	WorkDone         string     `json:"workDone" binding:"required,max=2000"`
	Issues           string     `json:"issues" binding:"max=2000"`
}

type DailyReportQuery struct {
	pagination.Query
}

type DailyReportResponse struct {
	ID               uuid.UUID  `json:"id"`
	ProjectID        uuid.UUID  `json:"projectId"`
	ReportDate       time.Time  `json:"reportDate"`
	WeatherMorning   string     `json:"weatherMorning"`
	WeatherAfternoon string     `json:"weatherAfternoon"`
	Workers          int        `json:"workers"`
	Equipment        string     `json:"equipment"`
	WorkDone         string     `json:"workDone"`
	Issues           string     `json:"issues"`
	CreatedBy        *uuid.UUID `json:"createdBy"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type DailyReportPage struct {
	pagination.Page[DailyReportResponse]
	ReportsLast30  int64 `json:"reportsLast30"`
	RainDaysLast30 int64 `json:"rainDaysLast30"`
}

type DailyReportFilter struct {
	ProjectID uuid.UUID
	Offset    int
	Limit     int
}

func newDailyReportResponse(report DailyReport) DailyReportResponse {
	return DailyReportResponse{
		ID:               report.ID,
		ProjectID:        report.ProjectID,
		ReportDate:       report.ReportDate,
		WeatherMorning:   report.WeatherMorning,
		WeatherAfternoon: report.WeatherAfternoon,
		Workers:          report.Workers,
		Equipment:        report.Equipment,
		WorkDone:         report.WorkDone,
		Issues:           report.Issues,
		CreatedBy:        report.CreatedBy,
		CreatedAt:        report.CreatedAt,
		UpdatedAt:        report.UpdatedAt,
	}
}

func (s *service) ListDailyReports(ctx context.Context, projectID uuid.UUID, query DailyReportQuery) (DailyReportPage, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return DailyReportPage{}, projectReadErrors.Resolve(err)
	}

	reports, total, err := s.repository.ListDailyReports(ctx, DailyReportFilter{
		ProjectID: projectID,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return DailyReportPage{}, apperror.Internal(err)
	}

	since := s.today().AddDate(0, 0, -(recentWindowDays - 1))
	recent, rainy, err := s.repository.CountRecentDailyReports(ctx, projectID, since)
	if err != nil {
		return DailyReportPage{}, apperror.Internal(err)
	}

	return DailyReportPage{
		Page:           pagination.New(pagination.Map(reports, newDailyReportResponse), query.Query, total),
		ReportsLast30:  recent,
		RainDaysLast30: rainy,
	}, nil
}

func (s *service) CreateDailyReport(ctx context.Context, projectID, actorID uuid.UUID, request DailyReportRequest) (DailyReportResponse, error) {
	if err := s.validateDailyReport(request); err != nil {
		return DailyReportResponse{}, err
	}
	if err := s.ensureOpenProject(ctx, projectID); err != nil {
		return DailyReportResponse{}, err
	}

	report := DailyReport{ProjectID: projectID, CreatedBy: &actorID}
	applyDailyReportRequest(&report, request)
	if err := s.repository.CreateDailyReport(ctx, &report); err != nil {
		return DailyReportResponse{}, dailyReportErrors.Resolve(err)
	}
	return newDailyReportResponse(report), nil
}

func (s *service) UpdateDailyReport(ctx context.Context, projectID, reportID uuid.UUID, request DailyReportRequest) (DailyReportResponse, error) {
	if err := s.validateDailyReport(request); err != nil {
		return DailyReportResponse{}, err
	}
	if err := s.ensureOpenProject(ctx, projectID); err != nil {
		return DailyReportResponse{}, err
	}

	report, err := s.repository.FindDailyReport(ctx, projectID, reportID)
	if err != nil {
		return DailyReportResponse{}, dailyReportErrors.Resolve(err)
	}
	applyDailyReportRequest(&report, request)
	if err := s.repository.SaveDailyReport(ctx, &report); err != nil {
		return DailyReportResponse{}, dailyReportErrors.Resolve(err)
	}
	return newDailyReportResponse(report), nil
}

func (s *service) DeleteDailyReport(ctx context.Context, projectID, reportID uuid.UUID) error {
	if err := s.ensureOpenProject(ctx, projectID); err != nil {
		return err
	}
	return dailyReportErrors.Resolve(s.repository.DeleteDailyReport(ctx, projectID, reportID))
}

func (s *service) validateDailyReport(request DailyReportRequest) error {
	if dateOnly(*request.ReportDate).After(s.today()) {
		return errDailyReportFuture
	}
	return nil
}

func applyDailyReportRequest(report *DailyReport, request DailyReportRequest) {
	report.ReportDate = dateOnly(*request.ReportDate)
	report.WeatherMorning = request.WeatherMorning
	report.WeatherAfternoon = request.WeatherAfternoon
	report.Workers = *request.Workers
	report.Equipment = strings.TrimSpace(request.Equipment)
	report.WorkDone = strings.TrimSpace(request.WorkDone)
	report.Issues = strings.TrimSpace(request.Issues)
}

func (r *gormRepository) ListDailyReports(ctx context.Context, filter DailyReportFilter) ([]DailyReport, int64, error) {
	return database.FindPage[DailyReport](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "report_date DESC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (f DailyReportFilter) apply(db *gorm.DB) *gorm.DB {
	return db.Where("project_id = ?", f.ProjectID)
}

func (r *gormRepository) CountRecentDailyReports(ctx context.Context, projectID uuid.UUID, since time.Time) (int64, int64, error) {
	var counts struct {
		Reports  int64
		RainDays int64
	}
	err := r.db.WithContext(ctx).
		Model(&DailyReport{}).
		Select("COUNT(*) AS reports, "+
			"COUNT(*) FILTER (WHERE weather_morning IN ? OR weather_afternoon IN ?) AS rain_days", rainyWeather, rainyWeather).
		Where("project_id = ? AND report_date >= ?", projectID, since).
		Scan(&counts).Error
	return counts.Reports, counts.RainDays, database.Translate(err)
}

func (r *gormRepository) FindDailyReport(ctx context.Context, projectID, id uuid.UUID) (DailyReport, error) {
	var report DailyReport
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&report).Error
	return report, database.Translate(err)
}

func (r *gormRepository) CreateDailyReport(ctx context.Context, report *DailyReport) error {
	return database.Translate(r.db.WithContext(ctx).Create(report).Error)
}

func (r *gormRepository) SaveDailyReport(ctx context.Context, report *DailyReport) error {
	return database.Translate(r.db.WithContext(ctx).Save(report).Error)
}

func (r *gormRepository) DeleteDailyReport(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &DailyReport{}, "id = ? AND project_id = ?", id, projectID)
}
