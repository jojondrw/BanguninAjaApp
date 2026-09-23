package scoring

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	weighted bool
	deleted  bool
	replaced []Weight
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) FindBuildingProfile(context.Context, uuid.UUID) (BuildingProfile, error) {
	return BuildingProfile{}, nil
}

func (f *fakeRepository) ReplaceWeights(_ context.Context, _ uuid.UUID, weights []Weight) error {
	f.replaced = weights
	return nil
}

func (f *fakeRepository) DimensionWeighted(context.Context, uuid.UUID) (bool, error) {
	return f.weighted, nil
}

func (f *fakeRepository) DeleteDimension(context.Context, uuid.UUID) error {
	f.deleted = true
	return nil
}

func TestValidateWeights(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	cases := []struct {
		name  string
		items []WeightItemRequest
		want  error
	}{
		{"sums to hundred", []WeightItemRequest{{first, 60}, {second, 40}}, nil},
		{"below hundred", []WeightItemRequest{{first, 60}, {second, 30}}, errWeightTotal},
		{"duplicate dimension", []WeightItemRequest{{first, 50}, {first, 50}}, errWeightDuplicate},
	}
	for _, tc := range cases {
		if got := validateWeights(tc.items); !errors.Is(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestReplaceWeightsStoresEveryDimension(t *testing.T) {
	repository := &fakeRepository{}
	profileID := uuid.New()
	request := WeightsRequest{Weights: []WeightItemRequest{{uuid.New(), 70}, {uuid.New(), 30}}}

	weights, err := NewService(repository).ReplaceWeights(context.Background(), profileID, request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(weights) != 2 || len(repository.replaced) != 2 || repository.replaced[0].BuildingProfileID != profileID {
		t.Fatalf("got %+v stored %+v", weights, repository.replaced)
	}
}

func TestWeightedDimensionCannotBeDeleted(t *testing.T) {
	repository := &fakeRepository{weighted: true}

	err := NewService(repository).DeleteDimension(context.Background(), uuid.New())
	if !errors.Is(err, errDimensionInUse) {
		t.Fatalf("got %v, want errDimensionInUse", err)
	}
	if repository.deleted {
		t.Fatal("dimension must not be deleted")
	}
}
