package project

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func TestSummarizeProjectsAggregatesInDatabase(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{
		Columns: []string{"count", "planning", "ongoing", "on_hold", "completed", "cancelled", "contract_value", "average_progress"},
		Rows:    [][]driver.Value{{int64(140), int64(30), int64(70), int64(10), int64(25), int64(5), "912000000000", "47"}},
	})

	totals, err := NewRepository(db).SummarizeProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := ProjectTotals{Count: 140, Planning: 30, Ongoing: 70, OnHold: 10, Completed: 25, Cancelled: 5, ContractValue: 912000000000, AverageProgress: 47}
	if totals != want {
		t.Fatalf("got %+v, want %+v", totals, want)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want a single aggregate query", len(queries))
	}
	query := queries[0]
	for _, fragment := range []string{
		"COUNT(*) AS count",
		"COUNT(*) FILTER (WHERE status = $3) AS on_hold",
		"COALESCE(SUM(contract_value), 0) AS contract_value",
		"COALESCE(ROUND(AVG(progress)), 0) AS average_progress",
		`FROM "project"`,
	} {
		if !strings.Contains(query.SQL, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query.SQL)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(query.SQL), `FROM "project"`) {
		t.Errorf("summary must cover every project:\n%s", query.SQL)
	}
	wantArgs := []any{statusPlanning, statusOngoing, statusOnHold, statusCompleted, statusCancelled}
	if len(query.Args) != len(wantArgs) {
		t.Fatalf("got args %v, want %v", query.Args, wantArgs)
	}
	for index, arg := range wantArgs {
		if query.Args[index] != arg {
			t.Errorf("arg %d: got %v, want %v", index, query.Args[index], arg)
		}
	}
}

func TestProjectSummaryResponseKeysStatusCounts(t *testing.T) {
	response := newProjectSummaryResponse(ProjectTotals{Count: 3, Planning: 1, OnHold: 2, ContractValue: 500, AverageProgress: 33})

	if response.Count != 3 || response.ContractValue != 500 || response.AverageProgress != 33 {
		t.Fatalf("got %+v", response)
	}
	if response.ByStatus != (ProjectStatusCounts{Planning: 1, OnHold: 2}) {
		t.Fatalf("got by status %+v", response.ByStatus)
	}
}
