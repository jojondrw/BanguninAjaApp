package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	FactDailyReport = "daily_report"
	FactSitePhoto   = "site_photo"
	FactProgress    = "progress"
	FactOwnerClaim  = "owner_claim"
)

type EvidenceFact struct {
	Kind      string
	SubjectID uuid.UUID
	ProjectID uuid.UUID
	Date      time.Time
	Label     string
	Fields    map[string]any
}

type EvidenceFeed struct {
	db      *gorm.DB
	service Service
}

func NewEvidenceFeed(db *gorm.DB) EvidenceFeed {
	return EvidenceFeed{db: db, service: NewService(NewRepository(db, nil))}
}

func (f EvidenceFeed) Facts(ctx context.Context, projectID *uuid.UUID) ([]EvidenceFact, error) {
	readers := []func(context.Context, *uuid.UUID) ([]EvidenceFact, error){
		f.dailyReports, f.sitePhotos, f.progress, f.ownerClaims,
	}
	var facts []EvidenceFact
	for _, read := range readers {
		found, err := read(ctx, projectID)
		if err != nil {
			return nil, err
		}
		facts = append(facts, found...)
	}
	return facts, nil
}

func (f EvidenceFeed) scoped(ctx context.Context, table string, projectID *uuid.UUID) *gorm.DB {
	query := f.db.WithContext(ctx).Table(table)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	return query
}

func (f EvidenceFeed) dailyReports(ctx context.Context, projectID *uuid.UUID) ([]EvidenceFact, error) {
	var rows []DailyReport
	if err := f.scoped(ctx, "daily_report", projectID).Order("report_date").Find(&rows).Error; err != nil {
		return nil, database.Translate(err)
	}
	facts := make([]EvidenceFact, len(rows))
	for index, row := range rows {
		facts[index] = EvidenceFact{
			Kind: FactDailyReport, SubjectID: row.ID, ProjectID: row.ProjectID, Date: row.ReportDate,
			Label: row.WorkDone,
			Fields: map[string]any{
				"reportDate": row.ReportDate.Format(time.DateOnly), "weatherMorning": row.WeatherMorning,
				"weatherAfternoon": row.WeatherAfternoon, "workers": row.Workers, "equipment": row.Equipment,
				"workDone": row.WorkDone, "issues": row.Issues,
			},
		}
	}
	return facts, nil
}

func (f EvidenceFeed) sitePhotos(ctx context.Context, projectID *uuid.UUID) ([]EvidenceFact, error) {
	var rows []struct {
		ID            uuid.UUID
		ProjectID     uuid.UUID
		TakenOn       time.Time
		Caption       string
		ContentType   string
		ContentSHA256 string
	}
	err := f.scoped(ctx, "site_photo", projectID).
		Select("id, project_id, taken_on, caption, content_type, encode(sha256(content), 'hex') AS content_sha256").
		Order("taken_on").Scan(&rows).Error
	if err != nil {
		return nil, database.Translate(err)
	}
	facts := make([]EvidenceFact, len(rows))
	for index, row := range rows {
		facts[index] = EvidenceFact{
			Kind: FactSitePhoto, SubjectID: row.ID, ProjectID: row.ProjectID, Date: row.TakenOn, Label: row.Caption,
			Fields: map[string]any{
				"takenOn": row.TakenOn.Format(time.DateOnly), "caption": row.Caption,
				"contentType": row.ContentType, "contentSha256": row.ContentSHA256,
			},
		}
	}
	return facts, nil
}

func (f EvidenceFeed) progress(ctx context.Context, projectID *uuid.UUID) ([]EvidenceFact, error) {
	var rows []ProgressSnapshot
	if err := f.scoped(ctx, "project_progress", projectID).Order("recorded_on").Find(&rows).Error; err != nil {
		return nil, database.Translate(err)
	}
	facts := make([]EvidenceFact, len(rows))
	for index, row := range rows {
		facts[index] = EvidenceFact{
			Kind: FactProgress, SubjectID: row.ID, ProjectID: row.ProjectID, Date: row.RecordedOn,
			Fields: map[string]any{"recordedOn": row.RecordedOn.Format(time.DateOnly), "progress": row.Progress},
		}
	}
	return facts, nil
}

func (f EvidenceFeed) ownerClaims(ctx context.Context, projectID *uuid.UUID) ([]EvidenceFact, error) {
	var rows []OwnerTerm
	if err := f.scoped(ctx, "owner_term", projectID).Where("receivable_id IS NOT NULL").Order("claimed_on").Find(&rows).Error; err != nil {
		return nil, database.Translate(err)
	}
	facts := make([]EvidenceFact, 0, len(rows))
	for _, row := range rows {
		fields := map[string]any{
			"kind": row.Kind, "requiredProgress": row.RequiredProgress, "percent": row.Percent,
			"grossAmount": row.GrossAmount, "retentionAmount": row.RetentionAmount,
		}
		date := row.CreatedAt
		if row.ClaimedOn != nil {
			date = *row.ClaimedOn
			fields["claimedOn"] = row.ClaimedOn.Format(time.DateOnly)
		}
		if row.ClaimedProgress != nil {
			fields["claimedProgress"] = *row.ClaimedProgress
		}
		facts = append(facts, EvidenceFact{Kind: FactOwnerClaim, SubjectID: row.ID, ProjectID: row.ProjectID, Date: date, Fields: fields})
	}
	return facts, nil
}

type OwnerClaim struct {
	TermID          uuid.UUID
	ProjectID       uuid.UUID
	ProjectCode     string
	ProjectName     string
	ContractNumber  string
	OwnerName       string
	Label           string
	Progress        int
	GrossAmount     int64
	RetentionAmount int64
	NetAmount       int64
	ClaimedOn       time.Time
}

func (f EvidenceFeed) OwnerClaim(ctx context.Context, projectID, termID uuid.UUID) (OwnerClaim, error) {
	project, err := f.service.GetProject(ctx, projectID)
	if err != nil {
		return OwnerClaim{}, err
	}
	schedule, err := f.service.OwnerSchedule(ctx, projectID)
	if err != nil {
		return OwnerClaim{}, err
	}
	for _, term := range schedule.Terms {
		if term.ID != termID || term.ClaimedOn == nil || schedule.Contract == nil {
			continue
		}
		claim := OwnerClaim{
			TermID: term.ID, ProjectID: projectID, ProjectCode: project.Code, ProjectName: project.Name,
			ContractNumber: schedule.Contract.Number, OwnerName: schedule.Contract.CustomerName, Label: term.Label,
			GrossAmount: term.GrossAmount, RetentionAmount: term.RetentionAmount, NetAmount: term.NetAmount,
			ClaimedOn: *term.ClaimedOn,
		}
		if term.ClaimedProgress != nil {
			claim.Progress = *term.ClaimedProgress
		}
		return claim, nil
	}
	return OwnerClaim{}, errTermNotFound
}
