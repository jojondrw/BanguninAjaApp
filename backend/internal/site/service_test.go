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

type fakeRegulation struct {
	response *regulation.Response
}

func (f fakeRegulation) GetByPoint(context.Context, float64, float64) (*regulation.Response, error) {
	return f.response, nil
}

type fakeNews struct {
	queries []news.NewsQuery
}

func (f *fakeNews) FetchNews(_ context.Context, query news.NewsQuery) ([]news.NewsResponse, error) {
	f.queries = append(f.queries, query)
	return []news.NewsResponse{{Title: "Berita " + query.Region + query.District}}, nil
}

type regionScoreClient struct {
	region string
}

func (c regionScoreClient) Score(ctx context.Context, input ScoreInput) (ScoreResult, error) {
	result, err := stubScoreClient{}.Score(ctx, input)
	result.Region = c.region
	return result, err
}

func pointer[T any](value T) *T {
	return &value
}

func newTestService(repository *fakeRepository) Service {
	return NewService(repository, stubScoreClient{}, fakeRegulation{}, &fakeNews{})
}

func evaluateWithNews(t *testing.T, scores ScoreClient, regulationData *regulation.Response) (EvaluateResponse, *fakeNews) {
	t.Helper()

	newsService := &fakeNews{}
	evaluator := NewService(&fakeRepository{}, scores, fakeRegulation{response: regulationData}, newsService)

	response, err := evaluator.Evaluate(context.Background(), uuid.New(), evaluateRequest(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return response, newsService
}

func simulatedRegulation() *regulation.Response {
	return &regulation.Response{ZoneName: "Kawasan Budidaya", IsSimulated: true}
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

func TestEvaluateQueriesNewsByRegulationDistrict(t *testing.T) {
	regulationData := &regulation.Response{ZoneName: "Perumahan", District: " Coblong "}

	response, newsService := evaluateWithNews(t, regionScoreClient{region: "Kota Bandung"}, regulationData)

	want := news.NewsQuery{District: "Coblong"}
	if len(newsService.queries) != 1 || newsService.queries[0] != want {
		t.Fatalf("got news queries %+v, want [%+v]", newsService.queries, want)
	}
	if len(response.Descriptive.News) != 1 {
		t.Fatalf("got %d news items, want 1", len(response.Descriptive.News))
	}
}

func TestEvaluateFallsBackToScoringRegionForNews(t *testing.T) {
	response, newsService := evaluateWithNews(t, regionScoreClient{region: "Kota Bandung"}, simulatedRegulation())

	want := news.NewsQuery{Region: "Kota Bandung"}
	if len(newsService.queries) != 1 || newsService.queries[0] != want {
		t.Fatalf("got news queries %+v, want [%+v]", newsService.queries, want)
	}
	if len(response.Descriptive.News) != 1 {
		t.Fatalf("got %d news items, want 1", len(response.Descriptive.News))
	}
	if response.Descriptive.Regulasi == nil || !response.Descriptive.Regulasi.IsSimulated {
		t.Fatal("simulated regulation should still be returned")
	}
}

func TestEvaluateFallsBackToScoringRegionWhenRegulationMissing(t *testing.T) {
	_, newsService := evaluateWithNews(t, regionScoreClient{region: "Kota Bandung"}, nil)

	want := news.NewsQuery{Region: "Kota Bandung"}
	if len(newsService.queries) != 1 || newsService.queries[0] != want {
		t.Fatalf("got news queries %+v, want [%+v]", newsService.queries, want)
	}
}

func TestEvaluateSkipsNewsWithoutDistrictOrRegion(t *testing.T) {
	response, newsService := evaluateWithNews(t, stubScoreClient{}, simulatedRegulation())

	if len(newsService.queries) != 0 {
		t.Fatalf("expected no news call, got %+v", newsService.queries)
	}
	if response.Descriptive.News == nil || len(response.Descriptive.News) != 0 {
		t.Fatalf("expected an empty non-nil news list, got %+v", response.Descriptive.News)
	}
}
