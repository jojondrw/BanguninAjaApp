package reporting

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting/export"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const previewRowLimit = 200

type ReportRequest struct {
	Name        string     `json:"name" binding:"required,max=160"`
	Type        string     `json:"type" binding:"required,max=40"`
	Scope       string     `json:"scope" binding:"max=120"`
	ProjectID   *uuid.UUID `json:"projectId"`
	PeriodStart *time.Time `json:"periodStart"`
	PeriodEnd   *time.Time `json:"periodEnd"`
	Format      string     `json:"format" binding:"required,oneof=pdf xlsx csv"`
}

type ReportQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=160"`
	Type      string     `form:"type" binding:"omitempty,max=40"`
	Format    string     `form:"format" binding:"omitempty,oneof=pdf xlsx csv"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
}

type PreviewQuery struct {
	Type        string     `form:"type" binding:"required,max=40"`
	ProjectID   *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	PeriodStart *time.Time `form:"periodStart" time_format:"2006-01-02"`
	PeriodEnd   *time.Time `form:"periodEnd" time_format:"2006-01-02"`
}

type ReportResponse struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Scope         string     `json:"scope"`
	ProjectID     *uuid.UUID `json:"projectId"`
	ProjectCode   string     `json:"projectCode"`
	ProjectName   string     `json:"projectName"`
	PeriodStart   *time.Time `json:"periodStart"`
	PeriodEnd     *time.Time `json:"periodEnd"`
	Format        string     `json:"format"`
	CreatedBy     uuid.UUID  `json:"createdBy"`
	CreatedByName string     `json:"createdByName"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type PreviewSection struct {
	Title     string          `json:"title"`
	Columns   []export.Column `json:"columns"`
	Rows      [][]any         `json:"rows"`
	Total     []any           `json:"total,omitempty"`
	TotalRows int             `json:"totalRows"`
}

type PreviewResponse struct {
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	Filters     string           `json:"filters"`
	PeriodStart *time.Time       `json:"periodStart"`
	PeriodEnd   *time.Time       `json:"periodEnd"`
	GeneratedAt time.Time        `json:"generatedAt"`
	Sections    []PreviewSection `json:"sections"`
}

func newReportResponse(row ReportRow) ReportResponse {
	report := row.Report
	return ReportResponse{
		ID:            report.ID,
		Name:          report.Name,
		Type:          report.Type,
		Scope:         report.Scope,
		ProjectID:     report.ProjectID,
		ProjectCode:   row.ProjectCode,
		ProjectName:   row.ProjectName,
		PeriodStart:   report.PeriodStart,
		PeriodEnd:     report.PeriodEnd,
		Format:        report.Format,
		CreatedBy:     report.CreatedBy,
		CreatedByName: row.CreatedByName,
		CreatedAt:     report.CreatedAt,
		UpdatedAt:     report.UpdatedAt,
	}
}

func newPreviewResponse(spec reportSpec, document export.Document) PreviewResponse {
	sections := make([]PreviewSection, 0, len(document.Sections))
	for _, section := range document.Sections {
		rows := section.Rows
		if len(rows) > previewRowLimit {
			rows = rows[:previewRowLimit]
		}
		if rows == nil {
			rows = [][]any{}
		}
		sections = append(sections, PreviewSection{
			Title: section.Title, Columns: section.Columns, Rows: rows, Total: section.Total, TotalRows: len(section.Rows),
		})
	}
	return PreviewResponse{
		Type:        spec.kind.key,
		Title:       document.Title,
		Filters:     document.Filters,
		PeriodStart: spec.periodStart,
		PeriodEnd:   spec.periodEnd,
		GeneratedAt: document.GeneratedAt,
		Sections:    sections,
	}
}
