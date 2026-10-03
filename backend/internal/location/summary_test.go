package location

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func pickResult(id uuid.UUID, name string, score int64, at time.Time) databasetest.Result {
	return databasetest.Result{
		Columns: []string{"id", "name", "score", "saved_at"},
		Rows:    [][]driver.Value{{id.String(), name, score, at}},
	}
}

func TestSummarizeSavedLocationsAggregatesEveryLocationOfTheProject(t *testing.T) {
	bestID, latestID := uuid.New(), uuid.New()
	bestAt := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	latestAt := time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count", "average_score"}, Rows: [][]driver.Value{{int64(150), int64(64)}}},
		pickResult(bestID, "UJI-Cibubur", 91, bestAt),
		pickResult(latestID, "UJI-Bekasi", 58, latestAt),
	)
	userID, projectID := uuid.New(), uuid.New()

	totals, err := NewRepository(db).SummarizeSavedLocations(context.Background(), SavedLocationFilter{UserID: userID, ProjectID: &projectID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if totals.Count != 150 || totals.AverageScore != 64 {
		t.Fatalf("got %+v", totals.SavedLocationAggregate)
	}
	if totals.Best == nil || totals.Best.ID != bestID || totals.Best.Score != 91 || totals.Best.Name != "UJI-Cibubur" {
		t.Fatalf("got best %+v", totals.Best)
	}
	if totals.Latest == nil || totals.Latest.ID != latestID || !totals.Latest.SavedAt.Equal(latestAt) {
		t.Fatalf("got latest %+v", totals.Latest)
	}

	queries := recorder.Queries()
	if len(queries) != 3 {
		t.Fatalf("got %d queries, want aggregate, best and latest", len(queries))
	}
	if !strings.Contains(queries[0].SQL, "COUNT(*) AS count, COALESCE(ROUND(AVG(score)), 0)::bigint AS average_score") ||
		strings.Contains(queries[0].SQL, "LIMIT") {
		t.Errorf("aggregate must cover every location:\n%s", queries[0].SQL)
	}
	if !strings.Contains(queries[1].SQL, "ORDER BY score DESC, saved_at DESC") || !strings.Contains(queries[2].SQL, "ORDER BY saved_at DESC") {
		t.Errorf("best and latest must be ordered by score and by date:\n%s\n%s", queries[1].SQL, queries[2].SQL)
	}
	for _, query := range queries {
		if !strings.Contains(query.SQL, "user_id = $1") || !strings.Contains(query.SQL, "project_id = $2") {
			t.Errorf("summary must stay inside the user's locations of the project:\n%s", query.SQL)
		}
		if len(query.Args) < 2 || query.Args[0] != userID.String() || query.Args[1] != projectID.String() {
			t.Errorf("got args %v", query.Args)
		}
	}
}

func TestSummarizeSavedLocationsWithoutLocationsSkipsPicks(t *testing.T) {
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count", "average_score"}, Rows: [][]driver.Value{{int64(0), int64(0)}}},
	)

	totals, err := NewRepository(db).SummarizeSavedLocations(context.Background(), SavedLocationFilter{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if totals.Count != 0 || totals.Best != nil || totals.Latest != nil {
		t.Fatalf("got %+v", totals)
	}
	if len(recorder.Queries()) != 1 {
		t.Fatalf("an empty project needs only the aggregate query, got %d", len(recorder.Queries()))
	}
}

type summaryRepository struct {
	Repository
	filter SavedLocationFilter
	totals SavedLocationTotals
}

func (f *summaryRepository) SummarizeSavedLocations(_ context.Context, filter SavedLocationFilter) (SavedLocationTotals, error) {
	f.filter = filter
	return f.totals, nil
}

func TestSummarizeSavedLocationsScopesToOwnerAndMapsPicks(t *testing.T) {
	userID, projectID := uuid.New(), uuid.New()
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	repository := &summaryRepository{totals: SavedLocationTotals{
		SavedLocationAggregate: SavedLocationAggregate{Count: 2, AverageScore: 70},
		Best:                   &SavedLocationPick{ID: uuid.New(), Name: "UJI-A", Score: 80, SavedAt: at},
	}}

	summary, err := NewService(repository).SummarizeSavedLocations(context.Background(), userID, SavedLocationSummaryQuery{ProjectID: &projectID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.filter.UserID != userID || repository.filter.ProjectID == nil || *repository.filter.ProjectID != projectID {
		t.Fatalf("got filter %+v", repository.filter)
	}
	if repository.filter.Limit != 0 || repository.filter.Offset != 0 {
		t.Fatalf("summary must not page, got filter %+v", repository.filter)
	}
	if summary.Count != 2 || summary.AverageScore != 70 || summary.Best == nil || summary.Best.Name != "UJI-A" || summary.Latest != nil {
		t.Fatalf("got %+v", summary)
	}
}
