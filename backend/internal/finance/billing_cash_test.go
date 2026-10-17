package finance

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func TestBillingCashPostsPaymentsAgainstReceivableOrPayableAccount(t *testing.T) {
	cashAccountID, counterID, projectID := uuid.New(), uuid.New(), uuid.New()
	cases := []struct {
		name      string
		incoming  bool
		wantCode  string
		wantType  string
		cashDebit bool
	}{
		{name: "faktur dibayar", incoming: true, wantCode: receivableAccountCode, wantType: cashIn, cashDebit: true},
		{name: "hutang dibayar", incoming: false, wantCode: payableAccountCode, wantType: cashOut, cashDebit: false},
	}

	for _, tc := range cases {
		db, recorder := databasetest.Open(t,
			accountResult(counterID),
			accountResult(cashAccountID),
			databasetest.Result{},
			sequenceResult(42),
			databasetest.Result{},
			databasetest.Result{},
		)
		paymentID := uuid.New()
		date := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)

		err := NewBillingCashLedger(NewRepository(db)).Record(context.Background(), BillingCash{
			PaymentID: paymentID, Incoming: tc.incoming, ProjectID: &projectID, Date: date, Amount: 2_500_000, Note: "UJI- " + tc.name,
		})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}

		queries := recorder.Queries()
		if args := queries[0].Args; len(args) == 0 || args[0] != tc.wantCode {
			t.Errorf("%s: counter account must be looked up by code %s, got %v", tc.name, tc.wantCode, args)
		}
		cash := insertedRows(t, onlyQuery(t, queries, `INSERT INTO "cash_transaction"`))[0]
		if cash["type"] != tc.wantType || cash["account_id"] != counterID.String() || cash["amount"] != int64(2_500_000) {
			t.Errorf("%s: got cash %+v", tc.name, cash)
		}
		if cash["billing_payment_id"] != paymentID.String() || cash["project_id"] != projectID.String() {
			t.Errorf("%s: cash must point at payment %s and project %s, got %+v", tc.name, paymentID, projectID, cash)
		}

		entry := insertedRows(t, onlyQuery(t, queries, `INSERT INTO "journal_entry"`))[0]
		want := cashLineWant{entryID: entry["id"].(string), debited: counterID, credited: cashAccountID, amount: 2_500_000}
		if tc.cashDebit {
			want.debited, want.credited = cashAccountID, counterID
		}
		assertCashLines(t, insertedRows(t, onlyQuery(t, queries, `INSERT INTO "journal_line"`)), want)
	}
}

func TestBillingCashNeedsTheReceivableAccount(t *testing.T) {
	db, _ := databasetest.Open(t, databasetest.Result{})

	err := NewBillingCashLedger(NewRepository(db)).Record(context.Background(), BillingCash{
		PaymentID: uuid.New(), Incoming: true, Date: recordedAt, Amount: 1_000,
	})
	if !errors.Is(err, errBillingAccountMissing) {
		t.Fatalf("got %v, want errBillingAccountMissing", err)
	}
}

func TestCashFromBillingCannotBeChangedFromFinance(t *testing.T) {
	transactionID, paymentID, counterID := uuid.New(), uuid.New(), uuid.New()
	billingRow := func() databasetest.Result {
		row := append(cashTransactionRow(transactionID, cashIn, counterID, 2_500_000, recordedAt), paymentID.String())
		return databasetest.Result{
			Columns: append(append([]string{}, cashTransactionColumns...), "billing_payment_id"),
			Rows:    [][]driver.Value{row},
		}
	}

	db, recorder := databasetest.Open(t, billingRow())
	_, err := NewService(NewRepository(db)).UpdateCashTransaction(context.Background(), transactionID, CashTransactionRequest{
		Date: recordedAt, Type: cashIn, AccountID: counterID, Amount: 1,
	})
	if !errors.Is(err, errCashFromBilling) {
		t.Fatalf("update: got %v, want errCashFromBilling", err)
	}
	assertRolledBack(t, recorder.Queries())

	db, recorder = databasetest.Open(t, billingRow())
	if err := NewService(NewRepository(db)).DeleteCashTransaction(context.Background(), transactionID); !errors.Is(err, errCashFromBilling) {
		t.Fatalf("delete: got %v, want errCashFromBilling", err)
	}
	assertRolledBack(t, recorder.Queries())
}
