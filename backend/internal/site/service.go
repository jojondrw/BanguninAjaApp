package site

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
)

// Error envelope per the T1 contract §2.4.
var (
	errProfileNotFound = apperror.NotFound("building_profile_not_found", "Profil bangunan tidak ditemukan")
	errProjectNotFound = apperror.NotFound("project_not_found", "Proyek tidak ditemukan atau bukan milik Anda")
	errScoringUnavail  = apperror.New(502, "scoring_unavailable", "Layanan penilaian sedang tidak tersedia, coba lagi nanti")
)

var profileReadErrors = database.ErrorMap{NotFound: errProfileNotFound}

type Service interface {
	Evaluate(ctx context.Context, userID uuid.UUID, request EvaluateRequest) (EvaluateResponse, error)
}

type service struct {
	repository Repository
	scores     ScoreClient
}

func NewService(repository Repository, scores ScoreClient) Service {
	return &service{repository: repository, scores: scores}
}

func (s *service) Evaluate(ctx context.Context, userID uuid.UUID, request EvaluateRequest) (EvaluateResponse, error) {
	// 1. Resolve building_profile_id (UUID) -> building_profile_code (string) for /score (§3.3).
	profile, err := s.repository.FindBuildingProfile(ctx, request.BuildingProfileID)
	if err != nil {
		return EvaluateResponse{}, profileReadErrors.Resolve(err)
	}

	// 2. Option (b) for GAP #1: validate project_id if present, but do not persist it (§3.2).
	if err := s.ensureProject(ctx, request.ProjectID); err != nil {
		return EvaluateResponse{}, err
	}

	// 3. Call the internal scoring service for the predictive score (§1).
	result, err := s.scores.Score(ctx, ScoreInput{
		Latitude:            *request.Latitude,
		Longitude:           *request.Longitude,
		BuildingProfileCode: profile.Code,
	})
	if err != nil {
		// Unknown profile means the DB and scoring service disagree on codes — a
		// server-side config problem, surfaced as 502 not 400 (§2.4 mapping note).
		if errors.Is(err, errUnknownProfile) {
			return EvaluateResponse{}, errScoringUnavail
		}
		return EvaluateResponse{}, errScoringUnavail
	}

	// 4. Persist the SavedLocation + per-dimension breakdown (§3.4).
	saved, err := s.persist(ctx, userID, request, profile.ID, result)
	if err != nil {
		return EvaluateResponse{}, err
	}

	// 5. Assemble the public response. Per contract §2.3 (Option A), descriptive is
	// always null until Ian's T18 lands; the shape is reserved but not populated here.
	return EvaluateResponse{
		SavedLocationID: saved.Location.ID,
		Predictive: Predictive{
			OverallScore:    result.OverallScore,
			DimensionScores: result.DimensionScores,
			RiskFlags:       nonNilFlags(result.RiskFlags),
		},
		Descriptive: nil,
	}, nil
}

func (s *service) ensureProject(ctx context.Context, projectID *uuid.UUID) error {
	if projectID == nil {
		return nil
	}
	exists, err := s.repository.ProjectExists(ctx, *projectID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !exists {
		return errProjectNotFound
	}
	return nil
}

func (s *service) persist(ctx context.Context, userID uuid.UUID, request EvaluateRequest, profileID uuid.UUID, result ScoreResult) (SavedRecord, error) {
	dimensionIDs, err := s.repository.DimensionIDByCode(ctx)
	if err != nil {
		return SavedRecord{}, apperror.Internal(err)
	}

	scores := make([]location.DimensionScore, 0, len(result.DimensionScores))
	for _, dimension := range result.DimensionScores {
		id, ok := dimensionIDs[dimension.DimensionCode]
		if !ok {
			// A returned code with no seeded dimension row is a server-side config
			// mismatch — do not silently drop it (§3.4).
			return SavedRecord{}, errScoringUnavail
		}
		scores = append(scores, location.DimensionScore{DimensionID: id, Value: dimension.Value})
	}

	saved := SavedRecord{
		Location: location.SavedLocation{
			UserID:            userID,
			Name:              strings.TrimSpace(request.Name),
			BuildingProfileID: &profileID,
			Point:             geo.Point{Lon: *request.Longitude, Lat: *request.Latitude},
			Score:             result.OverallScore,
		},
		Scores: scores,
	}

	err = s.repository.Transaction(ctx, func(repository Repository) error {
		return repository.CreateSavedLocation(ctx, &saved)
	})
	if err != nil {
		return SavedRecord{}, apperror.From(err)
	}
	return saved, nil
}

func nonNilFlags(flags []RiskFlag) []RiskFlag {
	if flags == nil {
		return []RiskFlag{}
	}
	return flags
}
