package reporting

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting/export"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	formatPDF  = "pdf"
	formatXLSX = "xlsx"
	formatCSV  = "csv"
)

var (
	errReportNotFound          = apperror.NotFound("report_not_found", "Laporan tidak ditemukan")
	errProjectNotFound         = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")
	errInvalidDateRange        = apperror.Unprocessable("invalid_date_range", "Tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	errReportTypeInvalid       = apperror.Unprocessable("report_type_invalid", "Jenis laporan tidak dikenal")
	errReportPeriodUnsupported = apperror.Unprocessable("report_period_unsupported", "Jenis laporan ini tidak memakai periode")
	errReportFormatInvalid     = apperror.Unprocessable("report_format_invalid", "Format berkas laporan tidak dikenal")
)

var (
	reportReadErrors  = database.ErrorMap{NotFound: errReportNotFound}
	reportWriteErrors = database.ErrorMap{Referenced: errProjectNotFound}
	projectReadErrors = database.ErrorMap{NotFound: errProjectNotFound}
)

type fileFormat struct {
	contentType string
	write       func(io.Writer, export.Document) error
}

var fileFormats = map[string]fileFormat{
	formatPDF:  {contentType: "application/pdf", write: export.WritePDF},
	formatXLSX: {contentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", write: export.WriteXLSX},
	formatCSV:  {contentType: "text/csv; charset=utf-8", write: export.WriteCSV},
}

type ReportFile struct {
	Name        string
	ContentType string
	Content     []byte
}

type Service interface {
	ListReports(ctx context.Context, role string, query ReportQuery) (pagination.Page[ReportResponse], error)
	GetReport(ctx context.Context, role string, id uuid.UUID) (ReportResponse, error)
	CreateReport(ctx context.Context, userID uuid.UUID, role string, request ReportRequest) (ReportResponse, error)
	DeleteReport(ctx context.Context, role string, id uuid.UUID) error
	PreviewReport(ctx context.Context, role string, query PreviewQuery) (PreviewResponse, error)
	DownloadReport(ctx context.Context, role string, id uuid.UUID) (ReportFile, error)
}

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) ListReports(ctx context.Context, role string, query ReportQuery) (pagination.Page[ReportResponse], error) {
	types := readableKinds(role)
	if len(types) == 0 {
		return pagination.New([]ReportResponse{}, query.Query, 0), nil
	}
	reports, total, err := s.repository.ListReports(ctx, ReportFilter{
		Search:    query.Search,
		Type:      strings.TrimSpace(query.Type),
		Types:     types,
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

func (s *service) GetReport(ctx context.Context, role string, id uuid.UUID) (ReportResponse, error) {
	report, err := s.readableReport(ctx, role, id)
	if err != nil {
		return ReportResponse{}, err
	}
	return newReportResponse(report), nil
}

func (s *service) CreateReport(ctx context.Context, userID uuid.UUID, role string, request ReportRequest) (ReportResponse, error) {
	spec, err := s.readableSpec(role, request.Type, request.ProjectID, request.PeriodStart, request.PeriodEnd)
	if err != nil {
		return ReportResponse{}, err
	}

	report := newReport(userID, request, spec)
	if err := s.repository.CreateReport(ctx, &report); err != nil {
		return ReportResponse{}, reportWriteErrors.Resolve(err)
	}
	return newReportResponse(ReportRow{Report: report}), nil
}

func (s *service) DeleteReport(ctx context.Context, role string, id uuid.UUID) error {
	if _, err := s.readableReport(ctx, role, id); err != nil {
		return err
	}
	return reportReadErrors.Resolve(s.repository.DeleteReport(ctx, id))
}

func (s *service) PreviewReport(ctx context.Context, role string, query PreviewQuery) (PreviewResponse, error) {
	spec, err := s.readableSpec(role, query.Type, query.ProjectID, query.PeriodStart, query.PeriodEnd)
	if err != nil {
		return PreviewResponse{}, err
	}
	document, _, err := s.buildDocument(ctx, spec)
	if err != nil {
		return PreviewResponse{}, err
	}
	return newPreviewResponse(spec, document), nil
}

func (s *service) DownloadReport(ctx context.Context, role string, id uuid.UUID) (ReportFile, error) {
	report, err := s.readableReport(ctx, role, id)
	if err != nil {
		return ReportFile{}, err
	}
	format, ok := fileFormats[report.Format]
	if !ok {
		return ReportFile{}, errReportFormatInvalid
	}
	spec, err := newReportSpec(report.Type, report.ProjectID, report.PeriodStart, report.PeriodEnd, s.today())
	if err != nil {
		return ReportFile{}, err
	}
	document, project, err := s.buildDocument(ctx, spec)
	if err != nil {
		return ReportFile{}, err
	}
	var content bytes.Buffer
	if err := format.write(&content, document); err != nil {
		return ReportFile{}, apperror.Internal(err)
	}
	return ReportFile{
		Name:        spec.fileName(project.Code, report.Format, s.today()),
		ContentType: format.contentType,
		Content:     content.Bytes(),
	}, nil
}

func (s *service) readableSpec(role, typeName string, projectID *uuid.UUID, start, end *time.Time) (reportSpec, error) {
	if kind, known := resolveKind(typeName); known {
		if err := authorizeType(role, kind.key); err != nil {
			return reportSpec{}, err
		}
	}
	return newReportSpec(typeName, projectID, start, end, s.today())
}

func (s *service) readableReport(ctx context.Context, role string, id uuid.UUID) (ReportRow, error) {
	report, err := s.repository.FindReport(ctx, id)
	if err != nil {
		return ReportRow{}, reportReadErrors.Resolve(err)
	}
	if err := authorizeType(role, report.Type); err != nil {
		return ReportRow{}, err
	}
	return report, nil
}

func (s *service) buildDocument(ctx context.Context, spec reportSpec) (export.Document, ProjectLabel, error) {
	project, err := s.projectLabel(ctx, spec.projectID)
	if err != nil {
		return export.Document{}, ProjectLabel{}, err
	}
	today := s.today()
	filter := DatasetFilter{ProjectID: spec.projectID, From: spec.periodStart, To: spec.periodEnd}
	sections, err := sectionBuilders[spec.kind.key](ctx, s.repository, filter, today)
	if err != nil {
		return export.Document{}, ProjectLabel{}, apperror.Internal(err)
	}
	return export.Document{
		Title:       spec.title(),
		Filters:     filtersLine(spec, project, today),
		GeneratedAt: s.now(),
		Sections:    sections,
	}, project, nil
}

func (s *service) projectLabel(ctx context.Context, projectID *uuid.UUID) (ProjectLabel, error) {
	if projectID == nil {
		return ProjectLabel{}, nil
	}
	project, err := s.repository.FindProjectLabel(ctx, *projectID)
	if err != nil {
		return ProjectLabel{}, projectReadErrors.Resolve(err)
	}
	return project, nil
}

func (s *service) today() time.Time {
	return duedate.Today(s.now())
}

func filtersLine(spec reportSpec, project ProjectLabel, today time.Time) string {
	parts := []string{"Semua proyek"}
	if project.Code != "" {
		parts[0] = "Proyek " + project.Code + " " + project.Name
	}
	if period := spec.periodText(); period != "" {
		parts = append(parts, period)
	}
	if spec.kind.key == typeAging || spec.kind.key == typeStock {
		parts = append(parts, "Posisi per "+today.Format(compactDate))
	}
	return strings.Join(parts, filterSeparator)
}

func dateRangeValid(start, end *time.Time) bool {
	return start == nil || end == nil || !end.Before(*start)
}

func newReport(userID uuid.UUID, request ReportRequest, spec reportSpec) Report {
	return Report{
		Name:        strings.TrimSpace(request.Name),
		Type:        spec.kind.key,
		Scope:       strings.TrimSpace(request.Scope),
		ProjectID:   request.ProjectID,
		PeriodStart: spec.periodStart,
		PeriodEnd:   spec.periodEnd,
		Format:      request.Format,
		CreatedBy:   userID,
	}
}
