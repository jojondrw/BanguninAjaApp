package location

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	areaPrecision  = 100
	indexPrecision = 1000
	boundsParts    = 4
)

var (
	errLocationNotFound        = apperror.NotFound("saved_location_not_found", "Lokasi tersimpan tidak ditemukan")
	errLocationReference       = apperror.Unprocessable("location_reference_not_found", "Wilayah, profil bangunan, atau dimensi tidak ditemukan")
	errLocationBounds          = apperror.Unprocessable("location_bounds_invalid", "Batas area peta harus lengkap dan nilai minimum harus lebih kecil dari maksimum")
	errDimensionScoreDuplicate = apperror.Unprocessable("dimension_score_duplicate", "Satu dimensi hanya boleh dinilai sekali per lokasi")

	errComparisonNotFound          = apperror.NotFound("comparison_not_found", "Perbandingan tidak ditemukan")
	errComparisonLocationDuplicate = apperror.Unprocessable("comparison_location_duplicate", "Lokasi yang sama tidak boleh dibandingkan dua kali")
	errComparisonLocationUnknown   = apperror.Unprocessable("comparison_location_not_found", "Ada lokasi yang tidak ditemukan di daftar lokasi tersimpan Anda")
)

var (
	locationReadErrors  = database.ErrorMap{NotFound: errLocationNotFound}
	locationWriteErrors = database.ErrorMap{NotFound: errLocationNotFound, Duplicate: errDimensionScoreDuplicate, Referenced: errLocationReference}

	comparisonReadErrors  = database.ErrorMap{NotFound: errComparisonNotFound}
	comparisonWriteErrors = database.ErrorMap{NotFound: errComparisonNotFound, Duplicate: errComparisonLocationDuplicate, Referenced: errComparisonLocationUnknown}
)

type Service interface {
	ListSavedLocations(ctx context.Context, userID uuid.UUID, query SavedLocationQuery) (pagination.Page[SavedLocationResponse], error)
	GetSavedLocation(ctx context.Context, userID, id uuid.UUID) (SavedLocationDetailResponse, error)
	CreateSavedLocation(ctx context.Context, userID uuid.UUID, request SavedLocationRequest) (SavedLocationDetailResponse, error)
	UpdateSavedLocation(ctx context.Context, userID, id uuid.UUID, request SavedLocationRequest) (SavedLocationDetailResponse, error)
	DeleteSavedLocation(ctx context.Context, userID, id uuid.UUID) error

	ListComparisons(ctx context.Context, userID uuid.UUID, query pagination.Query) (pagination.Page[ComparisonResponse], error)
	GetComparison(ctx context.Context, userID, id uuid.UUID) (ComparisonDetailResponse, error)
	CreateComparison(ctx context.Context, userID uuid.UUID, request ComparisonRequest) (ComparisonDetailResponse, error)
	UpdateComparison(ctx context.Context, userID, id uuid.UUID, request ComparisonRequest) (ComparisonDetailResponse, error)
	DeleteComparison(ctx context.Context, userID, id uuid.UUID) error
}

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) ListSavedLocations(ctx context.Context, userID uuid.UUID, query SavedLocationQuery) (pagination.Page[SavedLocationResponse], error) {
	bounds, err := boundsOf(query)
	if err != nil {
		return pagination.Page[SavedLocationResponse]{}, err
	}

	locations, total, err := s.repository.ListSavedLocations(ctx, SavedLocationFilter{
		UserID:            userID,
		Search:            query.Search,
		RegionID:          query.RegionID,
		BuildingProfileID: query.BuildingProfileID,
		MinScore:          query.MinScore,
		Bounds:            bounds,
		Offset:            query.Offset(),
		Limit:             query.Size(),
	})
	if err != nil {
		return pagination.Page[SavedLocationResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(locations, newSavedLocationResponse), query.Query, total), nil
}

func (s *service) GetSavedLocation(ctx context.Context, userID, id uuid.UUID) (SavedLocationDetailResponse, error) {
	location, err := s.repository.FindSavedLocation(ctx, userID, id)
	if err != nil {
		return SavedLocationDetailResponse{}, locationReadErrors.Resolve(err)
	}

	scores, err := s.repository.ListDimensionScores(ctx, []uuid.UUID{id})
	if err != nil {
		return SavedLocationDetailResponse{}, apperror.Internal(err)
	}
	return newSavedLocationDetail(location, scores), nil
}

func (s *service) CreateSavedLocation(ctx context.Context, userID uuid.UUID, request SavedLocationRequest) (SavedLocationDetailResponse, error) {
	if err := validateDimensionScores(request.DimensionScores); err != nil {
		return SavedLocationDetailResponse{}, err
	}

	location := SavedLocation{UserID: userID, SavedAt: s.now()}
	applySavedLocationRequest(&location, request)
	scores := newDimensionScores(request.DimensionScores)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createLocation(ctx, repository, &location, scores)
	})
	if err != nil {
		return SavedLocationDetailResponse{}, apperror.From(err)
	}
	return newSavedLocationDetail(location, scores), nil
}

func (s *service) UpdateSavedLocation(ctx context.Context, userID, id uuid.UUID, request SavedLocationRequest) (SavedLocationDetailResponse, error) {
	if err := validateDimensionScores(request.DimensionScores); err != nil {
		return SavedLocationDetailResponse{}, err
	}

	var location SavedLocation
	scores := newDimensionScores(request.DimensionScores)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		var err error
		location, err = s.updateLocation(ctx, repository, userID, id, request, scores)
		return err
	})
	if err != nil {
		return SavedLocationDetailResponse{}, apperror.From(err)
	}
	return newSavedLocationDetail(location, scores), nil
}

func (s *service) DeleteSavedLocation(ctx context.Context, userID, id uuid.UUID) error {
	return locationReadErrors.Resolve(s.repository.DeleteSavedLocation(ctx, userID, id))
}

func (s *service) ListComparisons(ctx context.Context, userID uuid.UUID, query pagination.Query) (pagination.Page[ComparisonResponse], error) {
	rows, total, err := s.repository.ListComparisons(ctx, userID, query.Offset(), query.Size())
	if err != nil {
		return pagination.Page[ComparisonResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newComparisonResponse), query, total), nil
}

func (s *service) GetComparison(ctx context.Context, userID, id uuid.UUID) (ComparisonDetailResponse, error) {
	comparison, err := s.repository.FindComparison(ctx, userID, id)
	if err != nil {
		return ComparisonDetailResponse{}, comparisonReadErrors.Resolve(err)
	}

	items, err := s.repository.ListComparisonItems(ctx, id)
	if err != nil {
		return ComparisonDetailResponse{}, apperror.Internal(err)
	}
	return s.describeComparison(ctx, comparison, items)
}

func (s *service) CreateComparison(ctx context.Context, userID uuid.UUID, request ComparisonRequest) (ComparisonDetailResponse, error) {
	if err := validateComparisonLocations(request.SavedLocationIDs); err != nil {
		return ComparisonDetailResponse{}, err
	}

	comparison := Comparison{UserID: userID, Name: strings.TrimSpace(request.Name)}
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createComparison(ctx, repository, &comparison, request.SavedLocationIDs)
	})
	if err != nil {
		return ComparisonDetailResponse{}, apperror.From(err)
	}
	return s.GetComparison(ctx, userID, comparison.ID)
}

func (s *service) UpdateComparison(ctx context.Context, userID, id uuid.UUID, request ComparisonRequest) (ComparisonDetailResponse, error) {
	if err := validateComparisonLocations(request.SavedLocationIDs); err != nil {
		return ComparisonDetailResponse{}, err
	}

	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return updateComparison(ctx, repository, userID, id, request)
	})
	if err != nil {
		return ComparisonDetailResponse{}, apperror.From(err)
	}
	return s.GetComparison(ctx, userID, id)
}

func (s *service) DeleteComparison(ctx context.Context, userID, id uuid.UUID) error {
	return comparisonReadErrors.Resolve(s.repository.DeleteComparison(ctx, userID, id))
}

func (s *service) updateLocation(ctx context.Context, repository Repository, userID, id uuid.UUID, request SavedLocationRequest, scores []DimensionScore) (SavedLocation, error) {
	location, err := s.resaveLocation(ctx, repository, userID, id, request)
	if err != nil {
		return SavedLocation{}, err
	}
	return location, storeDimensionScores(ctx, repository, id, scores)
}

func (s *service) resaveLocation(ctx context.Context, repository Repository, userID, id uuid.UUID, request SavedLocationRequest) (SavedLocation, error) {
	location, err := repository.FindSavedLocation(ctx, userID, id)
	if err != nil {
		return SavedLocation{}, locationReadErrors.Resolve(err)
	}

	previousScore := location.Score
	applySavedLocationRequest(&location, request)
	if location.Score != previousScore {
		location.SavedAt = s.now()
	}
	if err := repository.SaveSavedLocation(ctx, &location); err != nil {
		return SavedLocation{}, locationWriteErrors.Resolve(err)
	}
	return location, nil
}

func (s *service) describeComparison(ctx context.Context, comparison Comparison, items []ComparisonItem) (ComparisonDetailResponse, error) {
	ids := pagination.Map(items, func(item ComparisonItem) uuid.UUID { return item.SavedLocationID })
	locations, err := s.repository.FindSavedLocations(ctx, comparison.UserID, ids)
	if err != nil {
		return ComparisonDetailResponse{}, apperror.Internal(err)
	}

	scores, err := s.repository.ListDimensionScores(ctx, ids)
	if err != nil {
		return ComparisonDetailResponse{}, apperror.Internal(err)
	}
	return ComparisonDetailResponse{
		ID:        comparison.ID,
		Name:      comparison.Name,
		Items:     newComparisonItems(items, locations, scores),
		CreatedAt: comparison.CreatedAt,
		UpdatedAt: comparison.UpdatedAt,
	}, nil
}

func createLocation(ctx context.Context, repository Repository, location *SavedLocation, scores []DimensionScore) error {
	if err := repository.CreateSavedLocation(ctx, location); err != nil {
		return locationWriteErrors.Resolve(err)
	}
	return storeDimensionScores(ctx, repository, location.ID, scores)
}

func createComparison(ctx context.Context, repository Repository, comparison *Comparison, locationIDs []uuid.UUID) error {
	if err := ensureOwnedLocations(ctx, repository, comparison.UserID, locationIDs); err != nil {
		return err
	}
	if err := repository.CreateComparison(ctx, comparison); err != nil {
		return comparisonWriteErrors.Resolve(err)
	}
	return storeComparisonItems(ctx, repository, comparison.ID, locationIDs)
}

func updateComparison(ctx context.Context, repository Repository, userID, id uuid.UUID, request ComparisonRequest) error {
	if err := renameComparison(ctx, repository, userID, id, request.Name); err != nil {
		return err
	}
	if err := ensureOwnedLocations(ctx, repository, userID, request.SavedLocationIDs); err != nil {
		return err
	}
	return storeComparisonItems(ctx, repository, id, request.SavedLocationIDs)
}

func renameComparison(ctx context.Context, repository Repository, userID, id uuid.UUID, name string) error {
	comparison, err := repository.FindComparison(ctx, userID, id)
	if err != nil {
		return comparisonReadErrors.Resolve(err)
	}
	comparison.Name = strings.TrimSpace(name)
	return comparisonWriteErrors.Resolve(repository.SaveComparison(ctx, &comparison))
}

func ensureOwnedLocations(ctx context.Context, repository Repository, userID uuid.UUID, ids []uuid.UUID) error {
	locations, err := repository.FindSavedLocations(ctx, userID, ids)
	if err != nil {
		return apperror.Internal(err)
	}
	if len(locations) != len(ids) {
		return errComparisonLocationUnknown
	}
	return nil
}

func storeDimensionScores(ctx context.Context, repository Repository, locationID uuid.UUID, scores []DimensionScore) error {
	for index := range scores {
		scores[index].SavedLocationID = locationID
	}
	return locationWriteErrors.Resolve(repository.ReplaceDimensionScores(ctx, locationID, scores))
}

func storeComparisonItems(ctx context.Context, repository Repository, comparisonID uuid.UUID, locationIDs []uuid.UUID) error {
	items := make([]ComparisonItem, 0, len(locationIDs))
	for index, locationID := range locationIDs {
		items = append(items, ComparisonItem{ComparisonID: comparisonID, SavedLocationID: locationID, SortOrder: index + 1})
	}
	return comparisonWriteErrors.Resolve(repository.ReplaceComparisonItems(ctx, comparisonID, items))
}

func boundsOf(query SavedLocationQuery) (*Bounds, error) {
	parts := []*float64{query.MinLongitude, query.MinLatitude, query.MaxLongitude, query.MaxLatitude}
	provided := 0
	for _, part := range parts {
		if part != nil {
			provided++
		}
	}
	if provided == 0 {
		return nil, nil
	}
	if provided != boundsParts {
		return nil, errLocationBounds
	}

	bounds := Bounds{MinLongitude: *query.MinLongitude, MinLatitude: *query.MinLatitude, MaxLongitude: *query.MaxLongitude, MaxLatitude: *query.MaxLatitude}
	if bounds.MinLongitude >= bounds.MaxLongitude || bounds.MinLatitude >= bounds.MaxLatitude {
		return nil, errLocationBounds
	}
	return &bounds, nil
}

func validateDimensionScores(items []DimensionScoreRequest) error {
	seen := make(map[uuid.UUID]bool, len(items))
	for _, item := range items {
		if seen[item.DimensionID] {
			return errDimensionScoreDuplicate
		}
		seen[item.DimensionID] = true
	}
	return nil
}

func validateComparisonLocations(ids []uuid.UUID) error {
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return errComparisonLocationDuplicate
		}
		seen[id] = true
	}
	return nil
}

func newDimensionScores(items []DimensionScoreRequest) []DimensionScore {
	scores := make([]DimensionScore, 0, len(items))
	for _, item := range items {
		scores = append(scores, DimensionScore{DimensionID: item.DimensionID, Value: item.Value})
	}
	return scores
}

func newSavedLocationDetail(location SavedLocation, scores []DimensionScore) SavedLocationDetailResponse {
	return SavedLocationDetailResponse{
		SavedLocationResponse: newSavedLocationResponse(location),
		DimensionScores:       pagination.Map(scores, newDimensionScoreResponse),
	}
}

func newComparisonItems(items []ComparisonItem, locations []SavedLocation, scores []DimensionScore) []ComparisonItemResponse {
	byID := make(map[uuid.UUID]SavedLocation, len(locations))
	for _, location := range locations {
		byID[location.ID] = location
	}
	grouped := make(map[uuid.UUID][]DimensionScoreResponse, len(locations))
	for _, score := range scores {
		grouped[score.SavedLocationID] = append(grouped[score.SavedLocationID], newDimensionScoreResponse(score))
	}

	responses := make([]ComparisonItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, newComparisonItemResponse(item, byID[item.SavedLocationID], grouped[item.SavedLocationID]))
	}
	return responses
}

func roundTo(value float64, precision float64) float64 {
	return math.Round(value*precision) / precision
}

func applySavedLocationRequest(location *SavedLocation, request SavedLocationRequest) {
	location.Name = strings.TrimSpace(request.Name)
	location.RegionID = request.RegionID
	location.BuildingProfileID = request.BuildingProfileID
	location.Point = geo.Point{Lon: *request.Longitude, Lat: *request.Latitude}
	location.AreaSqm = roundTo(request.AreaSqm, areaPrecision)
	location.LandPricePerSqm = request.LandPricePerSqm
	location.Score = request.Score
	location.FloodIndex = roundTo(request.FloodIndex, indexPrecision)
	location.EarthquakeIndex = roundTo(request.EarthquakeIndex, indexPrecision)
	location.Note = strings.TrimSpace(request.Note)
}
