package finance

import (
	"context"
	"database/sql/driver"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var (
	cashTransactionColumns = []string{"id", "created_at", "updated_at", "date", "type", "account_id", "project_id", "amount", "note"}
	journalEntryColumns    = []string{"id", "created_at", "updated_at", "number", "date", "note", "source", "cash_transaction_id"}
	recordedAt             = time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
)

type cashLineWant struct {
	entryID  string
	debited  uuid.UUID
	credited uuid.UUID
	amount   int64
}

func accountResult(id uuid.UUID) databasetest.Result {
	return databasetest.Result{Columns: []string{"id"}, Rows: [][]driver.Value{{id.String()}}}
}

func sequenceResult(value int64) databasetest.Result {
	return databasetest.Result{Columns: []string{"nextval"}, Rows: [][]driver.Value{{value}}}
}

func cashTransactionRow(id uuid.UUID, kind string, accountID uuid.UUID, amount int64, date time.Time) []driver.Value {
	return []driver.Value{id.String(), recordedAt, recordedAt, date, kind, accountID.String(), nil, amount, "UJI- transaksi lama"}
}

func cashTransactionsResult(rows ...[]driver.Value) databasetest.Result {
	return databasetest.Result{Columns: cashTransactionColumns, Rows: rows}
}

func journalEntryResult(id, cashTransactionID uuid.UUID, number string) databasetest.Result {
	return databasetest.Result{Columns: journalEntryColumns, Rows: [][]driver.Value{{
		id.String(), recordedAt, recordedAt, number, recordedAt, "UJI- transaksi lama", cashJournalSource, cashTransactionID.String(),
	}}}
}

func statements(queries []databasetest.Query) []string {
	sql := make([]string, 0, len(queries))
	for _, query := range queries {
		sql = append(sql, query.SQL)
	}
	return sql
}

func queriesContaining(queries []databasetest.Query, fragment string) []databasetest.Query {
	var matches []databasetest.Query
	for _, query := range queries {
		if strings.Contains(query.SQL, fragment) {
			matches = append(matches, query)
		}
	}
	return matches
}

func onlyQuery(t *testing.T, queries []databasetest.Query, fragment string) databasetest.Query {
	t.Helper()
	matches := queriesContaining(queries, fragment)
	if len(matches) != 1 {
		t.Fatalf("want exactly one statement with %q, got %d:\n%s", fragment, len(matches), strings.Join(statements(queries), "\n"))
	}
	return matches[0]
}

func indexOf(queries []databasetest.Query, fragment string) int {
	return slices.IndexFunc(queries, func(query databasetest.Query) bool { return strings.Contains(query.SQL, fragment) })
}

func insertedRows(t *testing.T, query databasetest.Query) []map[string]any {
	t.Helper()
	open, end := strings.Index(query.SQL, "("), strings.Index(query.SQL, ")")
	if open < 0 || end < open {
		t.Fatalf("not an insert: %s", query.SQL)
	}
	columns := strings.Split(query.SQL[open+1:end], ",")
	if len(query.Args)%len(columns) != 0 {
		t.Fatalf("got %d args for %d columns: %s", len(query.Args), len(columns), query.SQL)
	}

	rows := make([]map[string]any, 0, len(query.Args)/len(columns))
	for start := 0; start < len(query.Args); start += len(columns) {
		row := make(map[string]any, len(columns))
		for index, column := range columns {
			row[strings.Trim(column, `" `)] = query.Args[start+index]
		}
		rows = append(rows, row)
	}
	return rows
}

func assertTransactionCommitted(t *testing.T, queries []databasetest.Query) {
	t.Helper()
	if len(queries) < 2 || queries[0].SQL != databasetest.Begin || queries[len(queries)-1].SQL != databasetest.Commit {
		t.Fatalf("every statement must run inside one committed transaction:\n%s", strings.Join(statements(queries), "\n"))
	}
	if len(queriesContaining(queries, databasetest.Begin)) != 1 {
		t.Fatalf("want a single transaction:\n%s", strings.Join(statements(queries), "\n"))
	}
}

func assertRolledBack(t *testing.T, queries []databasetest.Query) {
	t.Helper()
	if len(queries) == 0 || queries[0].SQL != databasetest.Begin || queries[len(queries)-1].SQL != databasetest.Rollback {
		t.Fatalf("want the transaction rolled back:\n%s", strings.Join(statements(queries), "\n"))
	}
}

func assertCashLines(t *testing.T, lines []map[string]any, want cashLineWant) {
	t.Helper()
	if len(lines) != 2 {
		t.Fatalf("a cash journal has exactly two lines, got %+v", lines)
	}
	debit, credit := lines[0], lines[1]
	if debit["journal_entry_id"] != want.entryID || credit["journal_entry_id"] != want.entryID {
		t.Errorf("lines must belong to journal %s, got %+v", want.entryID, lines)
	}
	if debit["account_id"] != want.debited.String() || debit["debit"] != want.amount || debit["kredit"] != int64(0) {
		t.Errorf("debit line: got %+v, want account %s debit %d", debit, want.debited, want.amount)
	}
	if credit["account_id"] != want.credited.String() || credit["kredit"] != want.amount || credit["debit"] != int64(0) {
		t.Errorf("credit line: got %+v, want account %s credit %d", credit, want.credited, want.amount)
	}
}

func TestCreateCashTransactionPostsBalancedJournal(t *testing.T) {
	cashAccountID, counterID := uuid.New(), uuid.New()
	cases := []struct {
		kind      string
		note      string
		wantNote  string
		cashDebit bool
	}{
		{kind: cashOut, note: "UJI- pekerjaan pondasi", wantNote: "UJI- pekerjaan pondasi", cashDebit: false},
		{kind: cashIn, note: "", wantNote: "Kas masuk", cashDebit: true},
	}

	for _, tc := range cases {
		db, recorder := databasetest.Open(t,
			accountResult(cashAccountID),
			databasetest.Result{},
			sequenceResult(123),
			databasetest.Result{},
			databasetest.Result{},
		)
		date := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)

		response, err := NewService(NewRepository(db)).CreateCashTransaction(context.Background(), CashTransactionRequest{
			Date: date, Type: tc.kind, AccountID: counterID, Amount: 1_250_000, Note: tc.note,
		})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.kind, err)
		}

		queries := recorder.Queries()
		assertTransactionCommitted(t, queries)
		onlyQuery(t, queries, `INSERT INTO "cash_transaction"`)
		if indexOf(queries, `INSERT INTO "cash_transaction"`) > indexOf(queries, `INSERT INTO "journal_entry"`) {
			t.Errorf("%s: the cash transaction must exist before its journal", tc.kind)
		}

		entry := insertedRows(t, onlyQuery(t, queries, `INSERT INTO "journal_entry"`))[0]
		if entry["number"] != "KAS-2026-000123" || entry["source"] != cashJournalSource || entry["note"] != tc.wantNote {
			t.Errorf("%s: got journal %+v", tc.kind, entry)
		}
		if entry["cash_transaction_id"] != response.ID.String() || entry["date"] != date {
			t.Errorf("%s: journal must point at cash transaction %s on %s, got %+v", tc.kind, response.ID, date, entry)
		}
		if response.JournalEntryID == nil || response.JournalEntryID.String() != entry["id"] {
			t.Errorf("%s: response must carry journal id %v, got %v", tc.kind, entry["id"], response.JournalEntryID)
		}

		want := cashLineWant{entryID: response.JournalEntryID.String(), debited: counterID, credited: cashAccountID, amount: 1_250_000}
		if tc.cashDebit {
			want.debited, want.credited = cashAccountID, counterID
		}
		assertCashLines(t, insertedRows(t, onlyQuery(t, queries, `INSERT INTO "journal_line"`)), want)
	}
}

func TestCreateCashTransactionResolvesCashAccountByCode(t *testing.T) {
	db, recorder := databasetest.Open(t, accountResult(uuid.New()))

	_, _ = NewService(NewRepository(db)).CreateCashTransaction(context.Background(), CashTransactionRequest{
		Date: recordedAt, Type: cashOut, AccountID: uuid.New(), Amount: 100,
	})

	lookup := onlyQuery(t, recorder.Queries(), `FROM "account"`)
	if !strings.Contains(lookup.SQL, "code = $1") || len(lookup.Args) == 0 || lookup.Args[0] != cashAccountCode {
		t.Fatalf("cash account must be found by code %s, got %s %v", cashAccountCode, lookup.SQL, lookup.Args)
	}
}

func TestCreateCashTransactionFailsWithoutCashAccount(t *testing.T) {
	db, recorder := databasetest.Open(t)

	_, err := NewService(NewRepository(db)).CreateCashTransaction(context.Background(), CashTransactionRequest{
		Date: recordedAt, Type: cashOut, AccountID: uuid.New(), Amount: 100,
	})
	if !errors.Is(err, errCashAccountMissing) {
		t.Fatalf("got %v, want errCashAccountMissing", err)
	}

	queries := recorder.Queries()
	assertRolledBack(t, queries)
	if len(queriesContaining(queries, "INSERT")) != 0 {
		t.Fatalf("nothing may be written without a cash account:\n%s", strings.Join(statements(queries), "\n"))
	}
}

func TestCreateCashTransactionRejectsCashAsCounterAccount(t *testing.T) {
	cashAccountID := uuid.New()
	db, recorder := databasetest.Open(t, accountResult(cashAccountID))

	_, err := NewService(NewRepository(db)).CreateCashTransaction(context.Background(), CashTransactionRequest{
		Date: recordedAt, Type: cashIn, AccountID: cashAccountID, Amount: 100,
	})
	if !errors.Is(err, errCashCounterIsCash) {
		t.Fatalf("got %v, want errCashCounterIsCash", err)
	}
	if len(queriesContaining(recorder.Queries(), "INSERT")) != 0 {
		t.Fatal("a rejected transaction must not be stored")
	}
}

func TestUpdateCashTransactionRewritesLinkedJournal(t *testing.T) {
	cashAccountID, oldCounterID, newCounterID := uuid.New(), uuid.New(), uuid.New()
	transactionID, entryID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t,
		cashTransactionsResult(cashTransactionRow(transactionID, cashOut, oldCounterID, 500_000, recordedAt)),
		accountResult(cashAccountID),
		databasetest.Result{Affected: 1},
		journalEntryResult(entryID, transactionID, "KAS-2026-000007"),
		databasetest.Result{Affected: 1},
		databasetest.Result{Affected: 2},
		databasetest.Result{},
	)
	newDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	response, err := NewService(NewRepository(db)).UpdateCashTransaction(context.Background(), transactionID, CashTransactionRequest{
		Date: newDate, Type: cashIn, AccountID: newCounterID, Amount: 750_000, Note: "UJI- koreksi",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.JournalEntryID == nil || *response.JournalEntryID != entryID {
		t.Fatalf("the linked journal must be kept, got %v", response.JournalEntryID)
	}

	queries := recorder.Queries()
	assertTransactionCommitted(t, queries)
	if len(queriesContaining(queries, "nextval")) != 0 || len(queriesContaining(queries, `INSERT INTO "journal_entry"`)) != 0 {
		t.Fatalf("an update must not post a second journal:\n%s", strings.Join(statements(queries), "\n"))
	}

	update := onlyQuery(t, queries, `UPDATE "journal_entry"`)
	for _, want := range []any{"KAS-2026-000007", newDate, "UJI- koreksi", entryID.String()} {
		if !slices.Contains(update.Args, want) {
			t.Errorf("journal update is missing %v, got %v", want, update.Args)
		}
	}

	onlyQuery(t, queries, `UPDATE "cash_transaction"`)
	removal := onlyQuery(t, queries, `DELETE FROM "journal_line"`)
	if !strings.Contains(removal.SQL, "journal_entry_id = $1") || removal.Args[0] != entryID.String() {
		t.Errorf("old lines must be removed by journal id, got %s %v", removal.SQL, removal.Args)
	}
	if indexOf(queries, `DELETE FROM "journal_line"`) > indexOf(queries, `INSERT INTO "journal_line"`) {
		t.Error("old lines must be removed before the new ones are written")
	}
	assertCashLines(t, insertedRows(t, onlyQuery(t, queries, `INSERT INTO "journal_line"`)), cashLineWant{
		entryID: entryID.String(), debited: cashAccountID, credited: newCounterID, amount: 750_000,
	})
}

func TestUpdateCashTransactionPostsJournalWhenNoneExists(t *testing.T) {
	cashAccountID, counterID, transactionID := uuid.New(), uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t,
		cashTransactionsResult(cashTransactionRow(transactionID, cashOut, counterID, 500_000, recordedAt)),
		accountResult(cashAccountID),
		databasetest.Result{Affected: 1},
		databasetest.Result{},
		sequenceResult(9),
		databasetest.Result{},
		databasetest.Result{},
	)

	response, err := NewService(NewRepository(db)).UpdateCashTransaction(context.Background(), transactionID, CashTransactionRequest{
		Date: recordedAt, Type: cashOut, AccountID: counterID, Amount: 500_000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queries := recorder.Queries()
	assertTransactionCommitted(t, queries)
	entry := insertedRows(t, onlyQuery(t, queries, `INSERT INTO "journal_entry"`))[0]
	if entry["number"] != "KAS-2026-000009" || entry["cash_transaction_id"] != transactionID.String() {
		t.Fatalf("got %+v", entry)
	}
	if response.JournalEntryID == nil || response.JournalEntryID.String() != entry["id"] {
		t.Fatalf("got journal id %v, want %v", response.JournalEntryID, entry["id"])
	}
}

func TestDeleteCashTransactionRemovesLinkedJournal(t *testing.T) {
	transactionID := uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{Affected: 2},
		databasetest.Result{Affected: 1},
		databasetest.Result{Affected: 1},
	)

	if err := NewService(NewRepository(db)).DeleteCashTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queries := recorder.Queries()
	assertTransactionCommitted(t, queries)
	if len(queries) != 5 {
		t.Fatalf("want lines, journal, and transaction removed:\n%s", strings.Join(statements(queries), "\n"))
	}
	wants := []string{
		`DELETE FROM "journal_line" WHERE journal_entry_id IN (SELECT "id" FROM "journal_entry" WHERE cash_transaction_id = $1)`,
		`DELETE FROM "journal_entry" WHERE cash_transaction_id = $1`,
		`DELETE FROM "cash_transaction" WHERE id = $1`,
	}
	for index, want := range wants {
		query := queries[index+1]
		if !strings.Contains(query.SQL, want) || len(query.Args) != 1 || query.Args[0] != transactionID.String() {
			t.Errorf("statement %d: got %s %v, want %s", index+1, query.SQL, query.Args, want)
		}
	}
}

func TestDeleteMissingCashTransactionRollsBack(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{}, databasetest.Result{}, databasetest.Result{})

	err := NewService(NewRepository(db)).DeleteCashTransaction(context.Background(), uuid.New())
	if !errors.Is(err, errCashTransactionNotFound) {
		t.Fatalf("got %v, want errCashTransactionNotFound", err)
	}
	assertRolledBack(t, recorder.Queries())
}

func TestPostMissingCashJournalsBackfillsOnlyUnlinkedTransactions(t *testing.T) {
	cashAccountID, expenseID, depositID := uuid.New(), uuid.New(), uuid.New()
	outID, inID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t,
		cashTransactionsResult(
			cashTransactionRow(outID, cashOut, expenseID, 1_250_000_000, time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)),
			cashTransactionRow(inID, cashIn, depositID, 300_000_000, time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)),
		),
		accountResult(cashAccountID),
		sequenceResult(1), databasetest.Result{}, databasetest.Result{},
		sequenceResult(2), databasetest.Result{}, databasetest.Result{},
	)

	posted, err := PostMissingCashJournals(context.Background(), NewRepository(db))
	if err != nil || posted != 2 {
		t.Fatalf("got %d, %v; want 2 posted", posted, err)
	}

	queries := recorder.Queries()
	assertTransactionCommitted(t, queries)
	missing := onlyQuery(t, queries, `FROM "cash_transaction"`)
	if !strings.Contains(missing.SQL, "NOT EXISTS (SELECT 1 FROM journal_entry WHERE journal_entry.cash_transaction_id = cash_transaction.id)") {
		t.Errorf("backfill must only pick transactions without a journal, got %s", missing.SQL)
	}

	entries := queriesContaining(queries, `INSERT INTO "journal_entry"`)
	lines := queriesContaining(queries, `INSERT INTO "journal_line"`)
	if len(entries) != 2 || len(lines) != 2 {
		t.Fatalf("want one journal per transaction:\n%s", strings.Join(statements(queries), "\n"))
	}
	first, second := insertedRows(t, entries[0])[0], insertedRows(t, entries[1])[0]
	if first["number"] != "KAS-2026-000001" || first["cash_transaction_id"] != outID.String() {
		t.Errorf("got first journal %+v", first)
	}
	if second["number"] != "KAS-2026-000002" || second["cash_transaction_id"] != inID.String() {
		t.Errorf("got second journal %+v", second)
	}
	assertCashLines(t, insertedRows(t, lines[0]), cashLineWant{entryID: first["id"].(string), debited: expenseID, credited: cashAccountID, amount: 1_250_000_000})
	assertCashLines(t, insertedRows(t, lines[1]), cashLineWant{entryID: second["id"].(string), debited: cashAccountID, credited: depositID, amount: 300_000_000})
}

func TestPostMissingCashJournalsTwiceWritesNothingTheSecondTime(t *testing.T) {
	db, recorder := databasetest.Open(t, cashTransactionsResult())

	posted, err := PostMissingCashJournals(context.Background(), NewRepository(db))
	if err != nil || posted != 0 {
		t.Fatalf("got %d, %v; want nothing posted", posted, err)
	}

	queries := recorder.Queries()
	assertTransactionCommitted(t, queries)
	if len(queries) != 3 {
		t.Fatalf("a rerun must only look for missing journals:\n%s", strings.Join(statements(queries), "\n"))
	}
}

func TestPostMissingCashJournalsStopsWithoutCashAccount(t *testing.T) {
	db, recorder := databasetest.Open(t, cashTransactionsResult(cashTransactionRow(uuid.New(), cashOut, uuid.New(), 100, recordedAt)))

	posted, err := PostMissingCashJournals(context.Background(), NewRepository(db))
	if !errors.Is(err, errCashAccountMissing) || posted != 0 {
		t.Fatalf("got %d, %v; want errCashAccountMissing", posted, err)
	}
	assertRolledBack(t, recorder.Queries())
}

func TestManualJournalCannotUseCashSystemSourceOrNumber(t *testing.T) {
	lines := []JournalLineRequest{{AccountID: uuid.New(), Debit: 100}, {AccountID: uuid.New(), Credit: 100}}
	cases := []struct {
		request JournalEntryRequest
		want    error
	}{
		{JournalEntryRequest{Number: "UJI-JU-001", Source: "kas", Lines: lines}, errJournalSourceReserved},
		{JournalEntryRequest{Number: "UJI-JU-002", Source: " Kas ", Lines: lines}, errJournalSourceReserved},
		{JournalEntryRequest{Number: "kas-2026-000001", Source: "Manual", Lines: lines}, errJournalNumberReserved},
	}

	for _, tc := range cases {
		repository := &fakeRepository{}
		_, err := NewService(repository).RecordJournalEntry(context.Background(), tc.request)
		if !errors.Is(err, tc.want) {
			t.Errorf("%+v: got %v, want %v", tc.request, err, tc.want)
		}
		if len(repository.entries) != 0 || len(repository.lines) != 0 {
			t.Errorf("%+v: a rejected journal must not be stored", tc.request)
		}
	}
}

func TestListCashTransactionsCarriesJournalEntryID(t *testing.T) {
	linkedID, legacyID, entryID := uuid.New(), uuid.New(), uuid.New()
	columns := []string{"id", "date", "type", "account_id", "project_id", "amount", "note", "created_at", "updated_at", "journal_entry_id"}
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(2)}}},
		databasetest.Result{Columns: columns, Rows: [][]driver.Value{
			{linkedID.String(), recordedAt, cashOut, uuid.NewString(), nil, int64(500), "UJI-", recordedAt, recordedAt, entryID.String()},
			{legacyID.String(), recordedAt, cashIn, uuid.NewString(), nil, int64(700), "UJI-", recordedAt, recordedAt, nil},
		}},
	)

	page, err := NewService(NewRepository(db)).ListCashTransactions(context.Background(), CashTransactionQuery{Type: cashOut})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 || page.Items[0].JournalEntryID == nil || *page.Items[0].JournalEntryID != entryID || page.Items[1].JournalEntryID != nil {
		t.Fatalf("got %+v", page.Items)
	}

	listing := recorder.Queries()[1].SQL
	for _, fragment := range []string{
		"LEFT JOIN journal_entry ON journal_entry.cash_transaction_id = cash_transaction.id",
		"journal_entry.id AS journal_entry_id",
		"cash_transaction.type = $1",
		"ORDER BY cash_transaction.date DESC, cash_transaction.created_at DESC",
	} {
		if !strings.Contains(listing, fragment) {
			t.Errorf("listing is missing %q:\n%s", fragment, listing)
		}
	}
}

func TestJournalEntryCarriesCashTransactionID(t *testing.T) {
	entryID, transactionID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{
			Columns: []string{"id", "number", "date", "note", "source", "cash_transaction_id", "created_at", "total"},
			Rows:    [][]driver.Value{{entryID.String(), "KAS-2026-000001", recordedAt, "UJI-", cashJournalSource, transactionID.String(), recordedAt, int64(500)}},
		},
	)

	detail, err := NewService(NewRepository(db)).GetJournalEntry(context.Background(), entryID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.CashTransactionID == nil || *detail.CashTransactionID != transactionID {
		t.Fatalf("got %v, want %s", detail.CashTransactionID, transactionID)
	}
	if !strings.Contains(recorder.Queries()[0].SQL, "journal_entry.cash_transaction_id") {
		t.Fatalf("journal query must select the cash link:\n%s", recorder.Queries()[0].SQL)
	}
	if manual := newJournalEntryResponse(JournalEntryRow{ID: uuid.New()}); manual.CashTransactionID != nil {
		t.Fatalf("a manual journal has no cash link, got %v", manual.CashTransactionID)
	}
}
