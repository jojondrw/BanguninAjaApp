package location

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting/export"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const pdfContentType = "application/pdf"

type ReportFile struct {
	Name        string
	ContentType string
	Content     []byte
}

type RegulationLookup interface {
	GetByPoint(ctx context.Context, latitude, longitude float64) (*regulation.Response, error)
}

type ReportService interface {
	SiteReport(ctx context.Context, userID, id uuid.UUID) (ReportFile, error)
	ComparisonReport(ctx context.Context, userID, id uuid.UUID) (ReportFile, error)
}

type reportService struct {
	repository  Repository
	regulations RegulationLookup
	now         func() time.Time
}

func NewReportService(repository Repository, regulations RegulationLookup) ReportService {
	return &reportService{repository: repository, regulations: regulations, now: time.Now}
}

func (s *reportService) SiteReport(ctx context.Context, userID, id uuid.UUID) (ReportFile, error) {
	location, err := s.repository.FindSavedLocation(ctx, userID, id)
	if err != nil {
		return ReportFile{}, locationReadErrors.Resolve(err)
	}
	report, err := s.collectSite(ctx, location)
	if err != nil {
		return ReportFile{}, apperror.Internal(err)
	}
	return renderReport(newSiteDocument(report), siteFileName(location.Name, report.GeneratedAt))
}

func (s *reportService) ComparisonReport(ctx context.Context, userID, id uuid.UUID) (ReportFile, error) {
	comparison, err := s.repository.FindComparison(ctx, userID, id)
	if err != nil {
		return ReportFile{}, comparisonReadErrors.Resolve(err)
	}
	report, err := s.collectComparison(ctx, comparison)
	if err != nil {
		return ReportFile{}, apperror.Internal(err)
	}
	return renderReport(newComparisonDocument(report), comparisonFileName(comparison.Name, report.GeneratedAt))
}

func (s *reportService) collectSite(ctx context.Context, location SavedLocation) (siteReport, error) {
	report := siteReport{Location: location, GeneratedAt: s.generatedAt()}
	scores, err := s.repository.ListDimensionScores(ctx, []uuid.UUID{location.ID})
	if err != nil {
		return siteReport{}, fmt.Errorf("list dimension scores: %w", err)
	}
	dimensions, err := s.repository.ListDimensions(ctx)
	if err != nil {
		return siteReport{}, fmt.Errorf("list dimensions: %w", err)
	}
	report.Scores, report.Dimensions = scores, dimensions
	if err := s.describeProfile(ctx, &report); err != nil {
		return siteReport{}, err
	}
	if err := s.describePlace(ctx, &report); err != nil {
		return siteReport{}, err
	}
	report.Regulation = s.regulationAt(ctx, location.Point)
	return report, nil
}

func (s *reportService) describeProfile(ctx context.Context, report *siteReport) error {
	profileID := report.Location.BuildingProfileID
	if profileID == nil {
		return nil
	}
	name, err := optional(s.repository.FindBuildingProfileName(ctx, *profileID))
	if err != nil {
		return fmt.Errorf("find building profile: %w", err)
	}
	weights, err := s.repository.ListProfileWeights(ctx, *profileID)
	if err != nil {
		return fmt.Errorf("list profile weights: %w", err)
	}
	report.ProfileName, report.Weights = name, weightsByDimension(weights)
	return nil
}

func (s *reportService) describePlace(ctx context.Context, report *siteReport) error {
	if regionID := report.Location.RegionID; regionID != nil {
		name, err := optional(s.repository.FindRegionName(ctx, *regionID))
		if err != nil {
			return fmt.Errorf("find region: %w", err)
		}
		report.RegionName = name
	}
	if projectID := report.Location.ProjectID; projectID != nil {
		project, err := optional(s.repository.FindProjectLabel(ctx, *projectID))
		if err != nil {
			return fmt.Errorf("find project: %w", err)
		}
		report.Project = presentProject(project)
	}
	return nil
}

func (s *reportService) regulationAt(ctx context.Context, point geo.Point) *regulation.Response {
	rule, err := s.regulations.GetByPoint(ctx, point.Lat, point.Lon)
	if err != nil {
		return nil
	}
	return rule
}

func (s *reportService) collectComparison(ctx context.Context, comparison Comparison) (comparisonReport, error) {
	items, err := s.repository.ListComparisonItems(ctx, comparison.ID)
	if err != nil {
		return comparisonReport{}, fmt.Errorf("list comparison items: %w", err)
	}
	report := comparisonReport{Comparison: comparison, GeneratedAt: s.generatedAt()}
	if err := s.collectComparedLocations(ctx, &report, items); err != nil {
		return comparisonReport{}, err
	}
	if report.Dimensions, err = s.repository.ListDimensions(ctx); err != nil {
		return comparisonReport{}, fmt.Errorf("list dimensions: %w", err)
	}
	return report, nil
}

func (s *reportService) collectComparedLocations(ctx context.Context, report *comparisonReport, items []ComparisonItem) error {
	ids := pagination.Map(items, func(item ComparisonItem) uuid.UUID { return item.SavedLocationID })
	if len(ids) == 0 {
		return nil
	}
	locations, err := s.repository.FindSavedLocations(ctx, report.Comparison.UserID, ids)
	if err != nil {
		return fmt.Errorf("find compared locations: %w", err)
	}
	scores, err := s.repository.ListDimensionScores(ctx, ids)
	if err != nil {
		return fmt.Errorf("list compared scores: %w", err)
	}
	report.Locations, report.Scores = orderLocations(items, locations), scores
	return nil
}

func (s *reportService) generatedAt() time.Time {
	return s.now().In(reportZone)
}

func renderReport(document export.Document, name string) (ReportFile, error) {
	var content bytes.Buffer
	if err := export.WritePDF(&content, document); err != nil {
		return ReportFile{}, apperror.Internal(err)
	}
	return ReportFile{Name: name, ContentType: pdfContentType, Content: content.Bytes()}, nil
}

func optional[T any](value T, err error) (T, error) {
	if errors.Is(err, database.ErrNotFound) {
		var zero T
		return zero, nil
	}
	return value, err
}

func presentProject(project ProjectLabel) *ProjectLabel {
	if project.Code == "" && project.Name == "" {
		return nil
	}
	return &project
}
