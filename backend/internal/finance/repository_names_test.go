package finance

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var (
	countOne = databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}}

	cashNameColumns = []string{
		"id", "date", "type", "account_id", "project_id", "amount", "note", "created_at", "updated_at",
		"journal_entry_id", "account_code", "account_name", "project_name",
	}
	cashNameSelects = []string{
		"(SELECT account.code FROM account WHERE account.id = cash_transaction.account_id) AS account_code",
		"(SELECT account.name FROM account WHERE account.id = cash_transaction.account_id) AS account_name",
		"(SELECT project.name FROM project WHERE project.id = cash_transaction.project_id) AS project_name",
	}
)

func assertSQLContains(t *testing.T, sql string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(sql, fragment) {
			t.Errorf("SQL is missing %q:\n%s", fragment, sql)
		}
	}
}

func TestListCashTransactionsReturnsAccountAndProjectNames(t *testing.T) {
	projectID, accountID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, countOne, databasetest.Result{Columns: cashNameColumns, Rows: [][]driver.Value{{
		uuid.NewString(), recordedAt, cashOut, accountID.String(), projectID.String(), int64(500), "UJI-", recordedAt, recordedAt,
		uuid.NewString(), "5110", "UJI Beban Material", "UJI Proyek Cibubur",
	}}})

	page, err := NewService(NewRepository(db)).ListCashTransactions(context.Background(), CashTransactionQuery{ProjectID: &projectID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("got %d items", len(page.Items))
	}
	item := page.Items[0]
	if item.AccountCode != "5110" || item.AccountName != "UJI Beban Material" || item.ProjectName != "UJI Proyek Cibubur" {
		t.Fatalf("got %+v", item)
	}

	queries := recorder.Queries()
	assertSQLContains(t, queries[1].SQL, append(cashNameSelects, "cash_transaction.project_id = $1")...)
	if strings.Contains(queries[0].SQL, "account_name") {
		t.Errorf("the count query must not select names:\n%s", queries[0].SQL)
	}
}

func TestGetCashTransactionReturnsNamesAndHeadOfficeHasNoProjectName(t *testing.T) {
	id := uuid.New()
	db, recorder := databasetest.Open(t, databasetest.Result{Columns: cashNameColumns, Rows: [][]driver.Value{{
		id.String(), recordedAt, cashIn, uuid.NewString(), nil, int64(900), "UJI-", recordedAt, recordedAt,
		nil, "4110", "UJI Pendapatan Penjualan", nil,
	}}})

	response, err := NewService(NewRepository(db)).GetCashTransaction(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.AccountCode != "4110" || response.AccountName != "UJI Pendapatan Penjualan" || response.ProjectName != "" || response.ProjectID != nil {
		t.Fatalf("got %+v", response)
	}
	assertSQLContains(t, recorder.Queries()[0].SQL, append(cashNameSelects, "cash_transaction.id = $1")...)
}

func TestListBudgetsReturnsProjectName(t *testing.T) {
	projectID := uuid.New()
	db, recorder := databasetest.Open(t, countOne, databasetest.Result{
		Columns: []string{"id", "project_id", "year", "value", "note", "created_at", "updated_at", "project_name", "realized"},
		Rows: [][]driver.Value{{
			uuid.NewString(), projectID.String(), int64(2026), int64(1000), "UJI-", recordedAt, recordedAt, "UJI Proyek Cibubur", int64(250),
		}},
	})

	page, err := NewService(NewRepository(db)).ListBudgets(context.Background(), BudgetQuery{Year: 2026})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ProjectName != "UJI Proyek Cibubur" || page.Items[0].Remaining != 750 {
		t.Fatalf("got %+v", page.Items)
	}
	assertSQLContains(t, recorder.Queries()[1].SQL,
		"(SELECT project.name FROM project WHERE project.id = budget.project_id) AS project_name",
		"budget.year = $2",
	)
}

func TestJournalEntryLinesReturnAccountCodeAndName(t *testing.T) {
	entryID, debitID, creditID := uuid.New(), uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{
			Columns: []string{"id", "number", "date", "note", "source", "cash_transaction_id", "created_at", "total"},
			Rows:    [][]driver.Value{{entryID.String(), "UJI-JU-001", recordedAt, "UJI-", "Manual", nil, recordedAt, int64(500)}},
		},
		databasetest.Result{
			Columns: []string{"id", "account_id", "debit", "credit", "account_code", "account_name"},
			Rows: [][]driver.Value{
				{uuid.NewString(), debitID.String(), int64(500), int64(0), "5110", "UJI Beban Material"},
				{uuid.NewString(), creditID.String(), int64(0), int64(500), "1110", "UJI Kas"},
			},
		},
	)

	detail, err := NewService(NewRepository(db)).GetJournalEntry(context.Background(), entryID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detail.Lines) != 2 {
		t.Fatalf("got %d lines", len(detail.Lines))
	}
	debit, credit := detail.Lines[0], detail.Lines[1]
	if debit.AccountID != debitID || debit.AccountCode != "5110" || debit.AccountName != "UJI Beban Material" || debit.Debit != 500 {
		t.Fatalf("got debit line %+v", debit)
	}
	if credit.AccountCode != "1110" || credit.AccountName != "UJI Kas" || credit.Credit != 500 || credit.Debit != 0 {
		t.Fatalf("got credit line %+v", credit)
	}
	assertSQLContains(t, recorder.Queries()[1].SQL,
		"JOIN account ON account.id = journal_line.account_id",
		"journal_line.kredit AS credit",
		"account.code AS account_code, account.name AS account_name",
		"journal_line.journal_entry_id = $1",
		"ORDER BY journal_line.created_at ASC, journal_line.id ASC",
	)
}
