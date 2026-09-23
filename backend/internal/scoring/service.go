package scoring

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const fullWeight = 100

var (
	errDimensionNotFound = apperror.NotFound("dimension_not_found", "Dimensi penilaian tidak ditemukan")
	errDimensionCodeUsed = apperror.Conflict("dimension_code_used", "Kode dimensi sudah dipakai")
	errDimensionInUse    = apperror.Conflict("dimension_in_use", "Dimensi masih dipakai oleh bobot atau skor lokasi")

	errProfileNotFound = apperror.NotFound("building_profile_not_found", "Profil bangunan tidak ditemukan")
	errProfileCodeUsed = apperror.Conflict("building_profile_code_used", "Kode profil bangunan sudah dipakai")

	errWeightTotal          = apperror.Unprocessable("weight_total_invalid", "Total bobot satu profil harus tepat 100%")
	errWeightDuplicate      = apperror.Unprocessable("weight_dimension_duplicate", "Satu dimensi hanya boleh muncul sekali dalam bobot")
	errWeightDimensionUnset = apperror.Unprocessable("dimension_not_found", "Dimensi penilaian tidak ditemukan")
)

var (
	dimensionReadErrors   = database.ErrorMap{NotFound: errDimensionNotFound}
	dimensionWriteErrors  = database.ErrorMap{NotFound: errDimensionNotFound, Duplicate: errDimensionCodeUsed}
	dimensionDeleteErrors = database.ErrorMap{NotFound: errDimensionNotFound, Referenced: errDimensionInUse}

	profileReadErrors  = database.ErrorMap{NotFound: errProfileNotFound}
	profileWriteErrors = database.ErrorMap{NotFound: errProfileNotFound, Duplicate: errProfileCodeUsed}

	weightWriteErrors = database.ErrorMap{Duplicate: errWeightDuplicate, Referenced: errWeightDimensionUnset}
)

type Service interface {
	ListDimensions(ctx context.Context) ([]DimensionResponse, error)
	GetDimension(ctx context.Context, id uuid.UUID) (DimensionResponse, error)
	CreateDimension(ctx context.Context, request DimensionRequest) (DimensionResponse, error)
	UpdateDimension(ctx context.Context, id uuid.UUID, request DimensionRequest) (DimensionResponse, error)
	DeleteDimension(ctx context.Context, id uuid.UUID) error

	ListBuildingProfiles(ctx context.Context) ([]BuildingProfileResponse, error)
	GetBuildingProfile(ctx context.Context, id uuid.UUID) (BuildingProfileDetailResponse, error)
	CreateBuildingProfile(ctx context.Context, request BuildingProfileRequest) (BuildingProfileResponse, error)
	UpdateBuildingProfile(ctx context.Context, id uuid.UUID, request BuildingProfileRequest) (BuildingProfileResponse, error)
	DeleteBuildingProfile(ctx context.Context, id uuid.UUID) error

	ListWeights(ctx context.Context, profileID uuid.UUID) ([]WeightResponse, error)
	ReplaceWeights(ctx context.Context, profileID uuid.UUID, request WeightsRequest) ([]WeightResponse, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListDimensions(ctx context.Context) ([]DimensionResponse, error) {
	dimensions, err := s.repository.ListDimensions(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(dimensions, newDimensionResponse), nil
}

func (s *service) GetDimension(ctx context.Context, id uuid.UUID) (DimensionResponse, error) {
	dimension, err := s.repository.FindDimension(ctx, id)
	if err != nil {
		return DimensionResponse{}, dimensionReadErrors.Resolve(err)
	}
	return newDimensionResponse(dimension), nil
}

func (s *service) CreateDimension(ctx context.Context, request DimensionRequest) (DimensionResponse, error) {
	var dimension Dimension
	applyDimensionRequest(&dimension, request)
	if err := s.repository.CreateDimension(ctx, &dimension); err != nil {
		return DimensionResponse{}, dimensionWriteErrors.Resolve(err)
	}
	return newDimensionResponse(dimension), nil
}

func (s *service) UpdateDimension(ctx context.Context, id uuid.UUID, request DimensionRequest) (DimensionResponse, error) {
	dimension, err := s.repository.FindDimension(ctx, id)
	if err != nil {
		return DimensionResponse{}, dimensionReadErrors.Resolve(err)
	}

	applyDimensionRequest(&dimension, request)
	if err := s.repository.SaveDimension(ctx, &dimension); err != nil {
		return DimensionResponse{}, dimensionWriteErrors.Resolve(err)
	}
	return newDimensionResponse(dimension), nil
}

func (s *service) DeleteDimension(ctx context.Context, id uuid.UUID) error {
	weighted, err := s.repository.DimensionWeighted(ctx, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if weighted {
		return errDimensionInUse
	}
	return dimensionDeleteErrors.Resolve(s.repository.DeleteDimension(ctx, id))
}

func (s *service) ListBuildingProfiles(ctx context.Context) ([]BuildingProfileResponse, error) {
	profiles, err := s.repository.ListBuildingProfiles(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(profiles, newBuildingProfileResponse), nil
}

func (s *service) GetBuildingProfile(ctx context.Context, id uuid.UUID) (BuildingProfileDetailResponse, error) {
	profile, err := s.repository.FindBuildingProfile(ctx, id)
	if err != nil {
		return BuildingProfileDetailResponse{}, profileReadErrors.Resolve(err)
	}

	weights, err := s.repository.ListWeights(ctx, id)
	if err != nil {
		return BuildingProfileDetailResponse{}, apperror.Internal(err)
	}
	return BuildingProfileDetailResponse{
		BuildingProfileResponse: newBuildingProfileResponse(profile),
		Weights:                 pagination.Map(weights, newWeightResponse),
	}, nil
}

func (s *service) CreateBuildingProfile(ctx context.Context, request BuildingProfileRequest) (BuildingProfileResponse, error) {
	var profile BuildingProfile
	applyBuildingProfileRequest(&profile, request)
	if err := s.repository.CreateBuildingProfile(ctx, &profile); err != nil {
		return BuildingProfileResponse{}, profileWriteErrors.Resolve(err)
	}
	return newBuildingProfileResponse(profile), nil
}

func (s *service) UpdateBuildingProfile(ctx context.Context, id uuid.UUID, request BuildingProfileRequest) (BuildingProfileResponse, error) {
	profile, err := s.repository.FindBuildingProfile(ctx, id)
	if err != nil {
		return BuildingProfileResponse{}, profileReadErrors.Resolve(err)
	}

	applyBuildingProfileRequest(&profile, request)
	if err := s.repository.SaveBuildingProfile(ctx, &profile); err != nil {
		return BuildingProfileResponse{}, profileWriteErrors.Resolve(err)
	}
	return newBuildingProfileResponse(profile), nil
}

func (s *service) DeleteBuildingProfile(ctx context.Context, id uuid.UUID) error {
	return profileReadErrors.Resolve(s.repository.DeleteBuildingProfile(ctx, id))
}

func (s *service) ListWeights(ctx context.Context, profileID uuid.UUID) ([]WeightResponse, error) {
	if _, err := s.repository.FindBuildingProfile(ctx, profileID); err != nil {
		return nil, profileReadErrors.Resolve(err)
	}

	weights, err := s.repository.ListWeights(ctx, profileID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(weights, newWeightResponse), nil
}

func (s *service) ReplaceWeights(ctx context.Context, profileID uuid.UUID, request WeightsRequest) ([]WeightResponse, error) {
	if err := validateWeights(request.Weights); err != nil {
		return nil, err
	}

	weights := newWeights(profileID, request.Weights)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return replaceWeights(ctx, repository, profileID, weights)
	})
	if err != nil {
		return nil, apperror.From(err)
	}
	return pagination.Map(weights, newWeightResponse), nil
}

func replaceWeights(ctx context.Context, repository Repository, profileID uuid.UUID, weights []Weight) error {
	if _, err := repository.FindBuildingProfile(ctx, profileID); err != nil {
		return profileReadErrors.Resolve(err)
	}
	return weightWriteErrors.Resolve(repository.ReplaceWeights(ctx, profileID, weights))
}

func validateWeights(items []WeightItemRequest) error {
	seen := make(map[uuid.UUID]bool, len(items))
	total := 0
	for _, item := range items {
		if seen[item.DimensionID] {
			return errWeightDuplicate
		}
		seen[item.DimensionID] = true
		total += item.Percent
	}
	if total != fullWeight {
		return errWeightTotal
	}
	return nil
}

func newWeights(profileID uuid.UUID, items []WeightItemRequest) []Weight {
	weights := make([]Weight, 0, len(items))
	for _, item := range items {
		weights = append(weights, Weight{
			BuildingProfileID: profileID,
			DimensionID:       item.DimensionID,
			Percent:           item.Percent,
		})
	}
	return weights
}

func applyDimensionRequest(dimension *Dimension, request DimensionRequest) {
	dimension.Code = strings.TrimSpace(request.Code)
	dimension.Name = strings.TrimSpace(request.Name)
	dimension.Description = strings.TrimSpace(request.Description)
	dimension.SortOrder = request.SortOrder
}

func applyBuildingProfileRequest(profile *BuildingProfile, request BuildingProfileRequest) {
	profile.Code = strings.TrimSpace(request.Code)
	profile.Name = strings.TrimSpace(request.Name)
}
