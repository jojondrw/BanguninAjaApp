package reporting

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

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

type ReportResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Scope       string     `json:"scope"`
	ProjectID   *uuid.UUID `json:"projectId"`
	PeriodStart *time.Time `json:"periodStart"`
	PeriodEnd   *time.Time `json:"periodEnd"`
	Format      string     `json:"format"`
	CreatedBy   uuid.UUID  `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func newReportResponse(report Report) ReportResponse {
	return ReportResponse{
		ID:          report.ID,
		Name:        report.Name,
		Type:        report.Type,
		Scope:       report.Scope,
		ProjectID:   report.ProjectID,
		PeriodStart: report.PeriodStart,
		PeriodEnd:   report.PeriodEnd,
		Format:      report.Format,
		CreatedBy:   report.CreatedBy,
		CreatedAt:   report.CreatedAt,
		UpdatedAt:   report.UpdatedAt,
	}
}
