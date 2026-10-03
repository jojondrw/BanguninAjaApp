package master

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

var countResult = databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}}

func assertContains(t *testing.T, sql string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(sql, fragment) {
			t.Errorf("SQL is missing %q:\n%s", fragment, sql)
		}
	}
}

func TestListAccountsPostableKeepsOnlyLeafAccounts(t *testing.T) {
	accountID := uuid.New()
	createdAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "created_at", "updated_at", "code", "name", "type", "parent_id"},
		Rows:    [][]driver.Value{{accountID.String(), createdAt, createdAt, "5110", "UJI Beban Material", "expense", nil}},
	})

	page, err := NewService(NewRepository(db)).ListAccounts(context.Background(), AccountQuery{
		Query:    pagination.Query{PageSize: 20},
		Search:   "beban",
		Type:     "expense",
		Postable: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.TotalItems != 1 || len(page.Items) != 1 || page.Items[0].Code != "5110" {
		t.Fatalf("got %+v", page)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want count and page", len(queries))
	}
	for _, query := range queries {
		assertContains(t, query.SQL,
			"(name ILIKE $1 OR code ILIKE $2)",
			"type = $3",
			"NOT EXISTS (SELECT 1 FROM account AS child WHERE child.parent_id = account.id)",
		)
	}
	assertContains(t, queries[1].SQL, "ORDER BY code ASC")
	if queries[1].Args[0] != "%beban%" || queries[1].Args[2] != "expense" {
		t.Fatalf("got args %v", queries[1].Args)
	}
}

func TestListAccountsWithoutPostableIncludesParents(t *testing.T) {
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{})

	if _, err := NewService(NewRepository(db)).ListAccounts(context.Background(), AccountQuery{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, query := range recorder.Queries() {
		if strings.Contains(query.SQL, "NOT EXISTS") {
			t.Fatalf("parent accounts must stay listed without postable:\n%s", query.SQL)
		}
	}
}

func TestListUnitsOfMeasureSearchesCodeOrName(t *testing.T) {
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{})

	if _, err := NewService(NewRepository(db)).ListUnitsOfMeasure(context.Background(), UnitOfMeasureQuery{Search: "m3"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want count and page", len(queries))
	}
	assertContains(t, queries[1].SQL, `FROM "unit_of_measure"`, "(name ILIKE $1 OR code ILIKE $2)", "ORDER BY code ASC")
	if queries[1].Args[0] != "%m3%" {
		t.Fatalf("got args %v", queries[1].Args)
	}
}
