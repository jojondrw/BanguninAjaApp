package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
)

type progressRecorder interface {
	progressDays(ctx context.Context, projectID uuid.UUID) (int64, error)
	recordProgress(ctx context.Context, projectID uuid.UUID, on time.Time, progress int) error
}

type gormProgressRecorder struct {
	db *gorm.DB
}

func (r gormProgressRecorder) progressDays(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&project.ProgressSnapshot{}).Where("project_id = ?", projectID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count progress history: %w", err)
	}
	return count, nil
}

func (r gormProgressRecorder) recordProgress(ctx context.Context, projectID uuid.UUID, on time.Time, progress int) error {
	snapshot := project.ProgressSnapshot{ProjectID: projectID, RecordedOn: on, Progress: progress}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&snapshot).Error
	if err != nil {
		return fmt.Errorf("record progress history: %w", err)
	}
	return nil
}

type historyPoint struct {
	day      time.Time
	progress int
}

func (s *seeder) seedProgressHistory() error {
	for _, seed := range demoProjects() {
		id, known := s.ids["project:"+seed.code]
		if !known {
			continue
		}
		recorded, err := s.history.progressDays(s.ctx, id)
		if err != nil {
			return err
		}
		if recorded > 1 {
			s.tally.skip("project_progress")
			continue
		}
		curve, err := s.services.projects.ProgressCurve(s.ctx, id)
		if err != nil {
			return fmt.Errorf("progress curve %s: %w", seed.code, err)
		}
		for _, point := range demoHistory(curve, s.today) {
			if err := s.history.recordProgress(s.ctx, id, point.day, point.progress); err != nil {
				return err
			}
			s.tally.add("project_progress")
		}
	}
	return nil
}

func demoHistory(curve project.ProgressCurveResponse, today time.Time) []historyPoint {
	if curve.Today.Planned == nil || *curve.Today.Planned <= 0 || curve.Today.Actual <= 0 {
		return nil
	}
	pace := float64(curve.Today.Actual) / *curve.Today.Planned

	history := []historyPoint{}
	reached := 0
	for _, point := range curve.Points {
		day, err := time.Parse(time.DateOnly, point.Date)
		if err != nil || !day.Before(today) {
			continue
		}
		progress := int(math.Round(point.Planned * pace))
		progress = min(max(progress, reached), curve.Today.Actual)
		reached = progress
		history = append(history, historyPoint{day: day, progress: progress})
	}
	return history
}
