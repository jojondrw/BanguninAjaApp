package reporting

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

var (
	errReportNotFound   = apperror.NotFound("report_not_found", "Laporan tidak ditemukan")
	errProjectNotFound  = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")
	errInvalidDateRange = apperror.Unprocessable("invalid_date_range", "Tanggal selesai tidak boleh lebih awal dari tanggal mulai")
)

var (
	reportReadErrors  = database.ErrorMap{NotFound: errReportNotFound}
	reportWriteErrors = database.ErrorMap{Referenced: errProjectNotFound}
)

type Service interface {
	ListReports(ctx context.Context, query ReportQuery) (pagination.Page[ReportResponse], error)
	GetReport(ctx context.Context, id uuid.UUID) (ReportResponse, error)
	CreateReport(ctx context.Context, userID uuid.UUID, request ReportRequest) (ReportResponse, error)
	DeleteReport(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListReports(ctx context.Context, query ReportQuery) (pagination.Page[ReportResponse], error) {
	reports, total, err := s.repository.ListReports(ctx, ReportFilter{
		Search:    query.Search,
		Type:      strings.TrimSpace(query.Type),
		Format:    query.Format,
		ProjectID: query.ProjectID,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[ReportResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(reports, newReportResponse), query.Query, total), nil
}

func (s *service) GetReport(ctx context.Context, id uuid.UUID) (ReportResponse, error) {
	report, err := s.repository.FindReport(ctx, id)
	if err != nil {
		return ReportResponse{}, reportReadErrors.Resolve(err)
	}
	return newReportResponse(report), nil
}

func (s *service) CreateReport(ctx context.Context, userID uuid.UUID, request ReportRequest) (ReportResponse, error) {
	if !dateRangeValid(request.PeriodStart, request.PeriodEnd) {
		return ReportResponse{}, errInvalidDateRange
	}

	report := newReport(userID, request)
	if err := s.repository.CreateReport(ctx, &report); err != nil {
		return ReportResponse{}, reportWriteErrors.Resolve(err)
	}
	return newReportResponse(report), nil
}

func (s *service) DeleteReport(ctx context.Context, id uuid.UUID) error {
	return reportReadErrors.Resolve(s.repository.DeleteReport(ctx, id))
}

func dateRangeValid(start, end *time.Time) bool {
	return start == nil || end == nil || !end.Before(*start)
}

func newReport(userID uuid.UUID, request ReportRequest) Report {
	return Report{
		Name:        strings.TrimSpace(request.Name),
		Type:        strings.TrimSpace(request.Type),
		Scope:       strings.TrimSpace(request.Scope),
		ProjectID:   request.ProjectID,
		PeriodStart: request.PeriodStart,
		PeriodEnd:   request.PeriodEnd,
		Format:      request.Format,
		CreatedBy:   userID,
	}
}
