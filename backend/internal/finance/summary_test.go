package finance

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var budgetTotalsResult = databasetest.Result{
	Columns: []string{"budget", "realized"},
	Rows:    [][]driver.Value{{"15000000000", "6200000000"}},
}

func TestSummarizeBudgetsAggregatesInDatabase(t *testing.T) {
	db, recorder := databasetest.Open(t, budgetTotalsResult)
	projectID := uuid.New()

	summary, err := NewBudgetSummaryService(NewBudgetSummaryRepository(db)).
		SummarizeBudgets(context.Background(), BudgetSummaryQuery{ProjectID: &projectID, Year: 2026})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary != (BudgetSummaryResponse{Budget: 15000000000, Realized: 6200000000}) {
		t.Fatalf("got %+v", summary)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want a single aggregate query", len(queries))
	}
	query := queries[0]
	for _, fragment := range []string{
		"COALESCE(SUM(budget.value), 0) AS budget",
		"COALESCE(SUM(spent.amount), 0) AS realized",
		"LEFT JOIN LATERAL (SELECT SUM(cash_transaction.amount) AS amount FROM cash_transaction",
		"cash_transaction.project_id = budget.project_id AND cash_transaction.type = $1",
		"make_date((budget.year + 1)::int, 1, 1)) AS spent ON TRUE",
		"budget.project_id = $2",
		"budget.year = $3",
	} {
		if !strings.Contains(query.SQL, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query.SQL)
		}
	}
	if strings.Contains(query.SQL, "LIMIT") || strings.Contains(query.SQL, "GROUP BY") {
		t.Errorf("summary must fold every budget into one row:\n%s", query.SQL)
	}
	if len(query.Args) != 3 || query.Args[0] != cashOut || query.Args[2] != int64(2026) {
		t.Errorf("got args %v", query.Args)
	}
}

func TestSummarizeBudgetsWithoutFilterCoversAllBudgets(t *testing.T) {
	db, recorder := databasetest.Open(t, budgetTotalsResult)

	if _, err := NewBudgetSummaryRepository(db).SummarizeBudgets(context.Background(), BudgetFilter{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := recorder.Queries()[0]
	if strings.Contains(query.SQL, "budget.project_id =") || strings.Contains(query.SQL, "budget.year =") || len(query.Args) != 1 {
		t.Fatalf("unfiltered summary must not narrow rows, got %s %v", query.SQL, query.Args)
	}
}
