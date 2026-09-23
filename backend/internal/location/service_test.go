package location

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	owned  map[uuid.UUID]bool
	scores []DimensionScore
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) CreateSavedLocation(_ context.Context, location *SavedLocation) error {
	location.ID = uuid.New()
	return nil
}

func (f *fakeRepository) ReplaceDimensionScores(_ context.Context, _ uuid.UUID, scores []DimensionScore) error {
	f.scores = scores
	return nil
}

func (f *fakeRepository) FindSavedLocations(_ context.Context, _ uuid.UUID, ids []uuid.UUID) ([]SavedLocation, error) {
	var locations []SavedLocation
	for _, id := range ids {
		if f.owned[id] {
			locations = append(locations, SavedLocation{})
		}
	}
	return locations, nil
}

func pointer[T any](value T) *T {
	return &value
}

func TestBoundsOf(t *testing.T) {
	if bounds, err := boundsOf(SavedLocationQuery{}); bounds != nil || err != nil {
		t.Fatalf("no bounds must mean no filter, got %v %v", bounds, err)
	}
	partial := SavedLocationQuery{MinLongitude: pointer(106.7), MinLatitude: pointer(-6.3)}
	if _, err := boundsOf(partial); !errors.Is(err, errLocationBounds) {
		t.Fatalf("partial bounds: got %v", err)
	}
	inverted := SavedLocationQuery{MinLongitude: pointer(107.0), MinLatitude: pointer(-6.3), MaxLongitude: pointer(106.7), MaxLatitude: pointer(-6.1)}
	if _, err := boundsOf(inverted); !errors.Is(err, errLocationBounds) {
		t.Fatalf("inverted bounds: got %v", err)
	}
	valid := SavedLocationQuery{MinLongitude: pointer(106.7), MinLatitude: pointer(-6.3), MaxLongitude: pointer(106.9), MaxLatitude: pointer(-6.1)}
	if bounds, err := boundsOf(valid); err != nil || bounds.MaxLongitude != 106.9 {
		t.Fatalf("valid bounds: got %+v %v", bounds, err)
	}
}

func TestCreateSavedLocationLinksDimensionScores(t *testing.T) {
	repository := &fakeRepository{}
	request := SavedLocationRequest{
		Name:            "Cempaka Putih Timur",
		Latitude:        pointer(-6.17),
		Longitude:       pointer(106.87),
		Score:           83,
		FloodIndex:      0.12345,
		DimensionScores: []DimensionScoreRequest{{DimensionID: uuid.New(), Value: 74}},
	}

	detail, err := NewService(repository).CreateSavedLocation(context.Background(), uuid.New(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.scores) != 1 || repository.scores[0].SavedLocationID != detail.ID {
		t.Fatalf("scores not linked to location: %+v", repository.scores)
	}
	if detail.FloodIndex != 0.123 || detail.Latitude != -6.17 {
		t.Fatalf("got flood %v latitude %v", detail.FloodIndex, detail.Latitude)
	}
}

func TestDuplicateDimensionScoreIsRejected(t *testing.T) {
	dimension := uuid.New()
	err := validateDimensionScores([]DimensionScoreRequest{{DimensionID: dimension}, {DimensionID: dimension}})
	if !errors.Is(err, errDimensionScoreDuplicate) {
		t.Fatalf("got %v, want errDimensionScoreDuplicate", err)
	}
}

func TestComparisonRejectsForeignLocation(t *testing.T) {
	mine, foreign := uuid.New(), uuid.New()
	repository := &fakeRepository{owned: map[uuid.UUID]bool{mine: true}}

	_, err := NewService(repository).CreateComparison(context.Background(), uuid.New(), ComparisonRequest{
		Name:             "Jakarta Utara",
		SavedLocationIDs: []uuid.UUID{mine, foreign},
	})
	if !errors.Is(err, errComparisonLocationUnknown) {
		t.Fatalf("got %v, want errComparisonLocationUnknown", err)
	}
}
