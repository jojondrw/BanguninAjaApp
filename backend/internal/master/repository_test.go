package master

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const foreignKeyViolation = "23503"

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
	parentID := uuid.New()
	createdAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	postable := true
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "code", "name", "type", "parent_id", "created_at", "updated_at", "parent_code", "parent_name", "postable"},
		Rows: [][]driver.Value{{
			accountID.String(), "5110", "UJI Beban Material", "expense", parentID.String(), createdAt, createdAt,
			"5100", "Beban Konstruksi", true,
		}},
	})

	page, err := NewService(NewRepository(db)).ListAccounts(context.Background(), AccountQuery{
		Query:    pagination.Query{PageSize: 20},
		Search:   "beban",
		Type:     "expense",
		Postable: &postable,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.TotalItems != 1 || len(page.Items) != 1 || page.Items[0].Code != "5110" {
		t.Fatalf("got %+v", page)
	}
	item := page.Items[0]
	if item.ParentCode == nil || *item.ParentCode != "5100" || item.ParentName == nil || *item.ParentName != "Beban Konstruksi" || !item.Postable {
		t.Fatalf("parent and postable must come from the row, got %+v", item)
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
	assertContains(t, queries[1].SQL,
		"AS parent_code",
		"AS parent_name",
		"NOT EXISTS (SELECT 1 FROM account AS child WHERE child.parent_id = account.id) AS postable",
		"ORDER BY account.code ASC",
	)
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
		if strings.Contains(query.SQL, "WHERE") && !strings.Contains(query.SQL, "AS postable") {
			t.Fatalf("parent accounts must stay listed without postable:\n%s", query.SQL)
		}
	}
	if count := recorder.Queries()[0].SQL; strings.Contains(count, "EXISTS") {
		t.Fatalf("count must not filter on children without postable:\n%s", count)
	}
}

func TestListAccountsNotPostableKeepsOnlyParentAccounts(t *testing.T) {
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{})
	postable := false

	if _, err := NewService(NewRepository(db)).ListAccounts(context.Background(), AccountQuery{Postable: &postable}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	queries := recorder.Queries()
	assertContains(t, queries[0].SQL, "WHERE EXISTS (SELECT 1 FROM account AS child WHERE child.parent_id = account.id)")
	if strings.Contains(queries[0].SQL, "NOT EXISTS") {
		t.Fatalf("postable=false must keep parent accounts only:\n%s", queries[0].SQL)
	}
}

func TestListRegionsCarriesParentName(t *testing.T) {
	regionID := uuid.New()
	parentID := uuid.New()
	createdAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "code", "name", "type", "parent_id", "created_at", "updated_at", "parent_code", "parent_name"},
		Rows:    [][]driver.Value{{regionID.String(), "3201", "Bogor", "regency", parentID.String(), createdAt, createdAt, "32", "Jawa Barat"}},
	})

	page, err := NewService(NewRepository(db)).ListRegions(context.Background(), RegionQuery{Type: "regency", ParentID: &parentID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ParentName == nil || *page.Items[0].ParentName != "Jawa Barat" {
		t.Fatalf("got %+v", page.Items)
	}

	queries := recorder.Queries()
	assertContains(t, queries[1].SQL,
		"(SELECT parent.name FROM region AS parent WHERE parent.id = region.parent_id) AS parent_name",
		"type = $1",
		"parent_id = $2",
		"ORDER BY region.name ASC, region.id ASC",
	)
}

func TestCreateAccountAnswersWithParentAndPostable(t *testing.T) {
	accountID := uuid.New()
	createdAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, databasetest.Result{Affected: 1}, databasetest.Result{
		Columns: []string{"id", "code", "name", "type", "parent_id", "created_at", "updated_at", "parent_code", "parent_name", "postable"},
		Rows:    [][]driver.Value{{accountID.String(), "UJI-1", "UJI Akun", "asset", nil, createdAt, createdAt, nil, nil, true}},
	})

	account, err := NewService(NewRepository(db)).CreateAccount(context.Background(), AccountRequest{Code: " UJI-1 ", Name: "UJI Akun", Type: "asset"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account.Code != "UJI-1" || account.ParentName != nil || !account.Postable {
		t.Fatalf("got %+v", account)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want insert and read back", len(queries))
	}
	assertContains(t, queries[0].SQL, `INSERT INTO "account"`)
	assertContains(t, queries[1].SQL, "AS postable", "WHERE account.id = $1")
	if !slices.Contains(queries[0].Args, any("UJI-1")) {
		t.Fatalf("code must be trimmed before insert, got args %v", queries[0].Args)
	}
}

func TestDeleteAccountNamesWhereItIsStillUsed(t *testing.T) {
	accountID := uuid.New()
	createdAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t,
		databasetest.Result{
			Columns: []string{"id", "created_at", "updated_at", "code", "name", "type", "parent_id"},
			Rows:    [][]driver.Value{{accountID.String(), createdAt, createdAt, "5100", "Beban Konstruksi", "expense", nil}},
		},
		databasetest.Result{Err: &pgconn.PgError{Code: foreignKeyViolation}},
		existsResult(true),
		existsResult(false),
		existsResult(true),
	)

	err := NewService(NewRepository(db)).DeleteAccount(context.Background(), accountID)

	assertRefusal(t, err, http.StatusConflict, "account_in_use", "Akun masih dipakai di akun turunan dan jurnal, jadi tidak bisa dihapus")
	queries := recorder.Queries()
	if len(queries) != 5 {
		t.Fatalf("got %d queries, want find, delete, and three usage checks", len(queries))
	}
	assertContains(t, queries[1].SQL, `DELETE FROM "account"`)
	assertContains(t, queries[2].SQL, "FROM account WHERE parent_id = $1")
	assertContains(t, queries[3].SQL, "FROM cash_transaction WHERE account_id = $1")
	assertContains(t, queries[4].SQL, "FROM journal_line WHERE account_id = $1")
	if queries[4].Args[0] != accountID.String() && queries[4].Args[0] != accountID {
		t.Fatalf("usage check must look for the deleted account, got %v", queries[4].Args)
	}
}

func TestDeleteRegionNamesWhereItIsStillUsed(t *testing.T) {
	db, _ := databasetest.Open(t,
		databasetest.Result{Err: &pgconn.PgError{Code: foreignKeyViolation}},
		existsResult(false),
		existsResult(true),
		existsResult(false),
	)

	err := NewService(NewRepository(db)).DeleteRegion(context.Background(), uuid.New())

	assertRefusal(t, err, http.StatusConflict, "region_in_use", "Wilayah masih dipakai di proyek, jadi tidak bisa dihapus")
}

func TestDeleteUnitOfMeasureNamesEveryUsage(t *testing.T) {
	db, _ := databasetest.Open(t,
		databasetest.Result{Err: &pgconn.PgError{Code: foreignKeyViolation}},
		existsResult(true),
		existsResult(true),
		existsResult(false),
		existsResult(true),
	)

	err := NewService(NewRepository(db)).DeleteUnitOfMeasure(context.Background(), uuid.New())

	assertRefusal(t, err, http.StatusConflict, "unit_of_measure_in_use",
		"Satuan masih dipakai di material, RAB proyek, dan pesanan pembelian, jadi tidak bisa dihapus")
}

func TestDeleteUnitOfMeasureFallsBackWhenUsageIsUnknown(t *testing.T) {
	db, _ := databasetest.Open(t,
		databasetest.Result{Err: &pgconn.PgError{Code: foreignKeyViolation}},
		existsResult(false),
		existsResult(false),
		existsResult(false),
		existsResult(false),
	)

	err := NewService(NewRepository(db)).DeleteUnitOfMeasure(context.Background(), uuid.New())

	assertRefusal(t, err, http.StatusConflict, "unit_of_measure_in_use", errUnitOfMeasureInUse.Message)
}

func TestDeleteUnusedUnitOfMeasureSkipsUsageChecks(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{Affected: 1})

	if err := NewService(NewRepository(db)).DeleteUnitOfMeasure(context.Background(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if queries := recorder.Queries(); len(queries) != 1 {
		t.Fatalf("got %d queries, want only the delete", len(queries))
	}
}

func existsResult(used bool) databasetest.Result {
	return databasetest.Result{Columns: []string{"exists"}, Rows: [][]driver.Value{{used}}}
}

func assertRefusal(t *testing.T, err error, status int, code, message string) {
	t.Helper()
	var refusal *apperror.Error
	if !errors.As(err, &refusal) {
		t.Fatalf("got %v, want an app error", err)
	}
	if refusal.Status != status || refusal.Code != code || refusal.Message != message {
		t.Fatalf("got %d %s %q, want %d %s %q", refusal.Status, refusal.Code, refusal.Message, status, code, message)
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
