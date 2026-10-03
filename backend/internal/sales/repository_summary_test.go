package sales

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var contractTotalsResult = databasetest.Result{
	Columns: []string{"count", "draft", "active", "paid", "cancelled", "total_value"},
	Rows:    [][]driver.Value{{int64(130), int64(10), int64(90), int64(25), int64(5), "98500000000"}},
}

func TestSummarizeContractsAggregatesInDatabase(t *testing.T) {
	db, recorder := databasetest.Open(t, contractTotalsResult)
	projectID := uuid.New()

	totals, err := NewRepository(db).SummarizeContracts(context.Background(), &projectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := ContractTotals{Count: 130, Draft: 10, Active: 90, Paid: 25, Cancelled: 5, TotalValue: 98500000000}
	if totals != want {
		t.Fatalf("got %+v, want %+v", totals, want)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want a single aggregate query", len(queries))
	}
	query := queries[0]
	for _, fragment := range []string{
		"COUNT(*) FILTER (WHERE contract.status = $2) AS active",
		"COALESCE(SUM(contract.value) FILTER (WHERE contract.status <> $5), 0) AS total_value",
		"JOIN unit ON unit.id = contract.unit_id",
		"unit.project_id = $6",
	} {
		if !strings.Contains(query.SQL, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query.SQL)
		}
	}
	if strings.Contains(query.SQL, "LIMIT") {
		t.Errorf("summary must cover every contract, got a limited query:\n%s", query.SQL)
	}
	wantArgs := []any{contractDraft, contractActive, contractPaid, contractCancelled, contractCancelled}
	if len(query.Args) != len(wantArgs)+1 {
		t.Fatalf("got args %v", query.Args)
	}
	for index, arg := range wantArgs {
		if query.Args[index] != arg {
			t.Errorf("arg %d: got %v, want %v", index, query.Args[index], arg)
		}
	}
}

func TestSummarizeContractsWithoutProjectSkipsUnitJoin(t *testing.T) {
	db, recorder := databasetest.Open(t, contractTotalsResult)

	if _, err := NewRepository(db).SummarizeContracts(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := recorder.Queries()[0]
	if strings.Contains(query.SQL, "JOIN") || strings.Contains(query.SQL, "project_id") || len(query.Args) != 5 {
		t.Fatalf("unfiltered summary must not narrow rows, got %s %v", query.SQL, query.Args)
	}
}

func TestContractSummaryResponseKeepsStatusCounts(t *testing.T) {
	response := newContractSummaryResponse(ContractTotals{Count: 4, Draft: 1, Active: 2, Cancelled: 1, TotalValue: 750})

	if response.Count != 4 || response.TotalValue != 750 {
		t.Fatalf("got %+v", response)
	}
	if response.ByStatus != (ContractStatusCounts{Draft: 1, Active: 2, Cancelled: 1}) {
		t.Fatalf("got by status %+v", response.ByStatus)
	}
}
