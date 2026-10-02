package site

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/news"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/scoring"
)

type fakeRepository struct {
	projects map[uuid.UUID]bool
	saved    *SavedRecord
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) FindBuildingProfile(_ context.Context, id uuid.UUID) (scoring.BuildingProfile, error) {
	profile := scoring.BuildingProfile{Code: "housing"}
	profile.ID = id
	return profile, nil
}

func (f *fakeRepository) DimensionIDByCode(_ context.Context) (map[string]uuid.UUID, error) {
	byCode := make(map[string]uuid.UUID, len(scoredDimensions))
	for _, code := range scoredDimensions {
		byCode[code] = uuid.New()
	}
	return byCode, nil
}

func (f *fakeRepository) ProjectExists(_ context.Context, id uuid.UUID) (bool, error) {
	return f.projects[id], nil
}

func (f *fakeRepository) CreateSavedLocation(_ context.Context, saved *SavedRecord) error {
	saved.Location.ID = uuid.New()
	f.saved = saved
	return nil
}

type fakeRegulation struct{}

func (fakeRegulation) GetByPoint(context.Context, float64, float64) (*regulation.Response, error) {
	return nil, nil
}

type fakeNews struct{}

func (fakeNews) FetchNews(context.Context, news.NewsQuery) ([]news.NewsResponse, error) {
	return nil, nil
}

func pointer[T any](value T) *T {
	return &value
}

func newTestService(repository *fakeRepository) Service {
	return NewService(repository, stubScoreClient{}, fakeRegulation{}, fakeNews{})
}

func evaluateRequest(projectID *uuid.UUID) EvaluateRequest {
	return EvaluateRequest{
		ProjectID:         projectID,
		Latitude:          pointer(-6.914744),
		Longitude:         pointer(107.609810),
		BuildingProfileID: uuid.New(),
		Name:              "Lahan Dago",
	}
}

func TestEvaluateLinksSavedLocationToProject(t *testing.T) {
	projectID := uuid.New()
	repository := &fakeRepository{projects: map[uuid.UUID]bool{projectID: true}}

	response, err := newTestService(repository).Evaluate(context.Background(), uuid.New(), evaluateRequest(&projectID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.saved == nil || repository.saved.Location.ProjectID == nil {
		t.Fatal("saved location was not linked to the project")
	}
	if *repository.saved.Location.ProjectID != projectID {
		t.Fatalf("got project %v, want %v", *repository.saved.Location.ProjectID, projectID)
	}
	if response.SavedLocationID != repository.saved.Location.ID {
		t.Fatalf("response id %v does not match saved id %v", response.SavedLocationID, repository.saved.Location.ID)
	}
}

func TestEvaluateWithoutProjectLeavesLinkEmpty(t *testing.T) {
	repository := &fakeRepository{}

	if _, err := newTestService(repository).Evaluate(context.Background(), uuid.New(), evaluateRequest(nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.saved.Location.ProjectID != nil {
		t.Fatalf("expected no project, got %v", *repository.saved.Location.ProjectID)
	}
}

func TestEvaluateRejectsUnknownProject(t *testing.T) {
	repository := &fakeRepository{}
	unknown := uuid.New()

	_, err := newTestService(repository).Evaluate(context.Background(), uuid.New(), evaluateRequest(&unknown))
	if !errors.Is(err, errProjectNotFound) {
		t.Fatalf("got %v, want errProjectNotFound", err)
	}
	if repository.saved != nil {
		t.Fatal("nothing should be saved for an unknown project")
	}
}
