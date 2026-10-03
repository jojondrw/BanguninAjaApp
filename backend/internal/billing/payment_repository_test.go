package billing

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func TestListPaymentsIsNewestFirstForOneParent(t *testing.T) {
	invoiceID := uuid.New()
	db, recorder := databasetest.Open(t, countResult)

	if _, _, err := NewRepository(db).ListPayments(context.Background(), PaymentFilter{Column: invoicePaymentColumn, ParentID: invoiceID, Limit: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[0].SQL, "billing_payment.invoice_id =")
	assertContains(t, queries[1].SQL,
		"billing_payment.invoice_id =",
		"invoice.number AS invoice_number",
		"receivable.reference AS receivable_reference",
		"users.name AS created_by_name",
		"ORDER BY billing_payment.paid_at DESC, billing_payment.created_at DESC",
	)
	if queries[1].Args[0] != invoiceID.String() {
		t.Fatalf("got args %v", queries[1].Args)
	}
}

func TestBillingRowsReportPartyNameAndLastPaymentDate(t *testing.T) {
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{}, countResult, databasetest.Result{}, countResult)
	repository := NewRepository(db)

	if _, _, err := repository.ListInvoices(context.Background(), InvoiceFilter{Limit: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := repository.ListReceivables(context.Background(), ReceivableFilter{Limit: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := repository.ListPayables(context.Background(), PayableFilter{Limit: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	queries := recorder.Queries()
	assertContains(t, queries[1].SQL,
		"LEFT JOIN customer ON invoice.party_type = 'customer'",
		"LEFT JOIN vendor ON invoice.party_type = 'vendor'",
		"COALESCE(customer.name, vendor.name) AS party_name",
		"WHERE billing_payment.invoice_id = invoice.id) AS last_paid_at",
	)
	assertContains(t, queries[3].SQL, "WHERE billing_payment.receivable_id = receivable.id) AS last_paid_at")
	assertContains(t, queries[5].SQL, "WHERE billing_payment.payable_id = payable.id) AS last_paid_at")
}

func TestBackfillInsertsOneRowPerUnrecordedBalanceInOneTransaction(t *testing.T) {
	db, recorder := databasetest.Open(t,
		databasetest.Result{Affected: 2},
		databasetest.Result{Affected: 3},
		databasetest.Result{Affected: 1},
		databasetest.Result{Affected: 4},
	)

	inserted, err := BackfillPayments(context.Background(), NewRepository(db))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted != 10 {
		t.Fatalf("got %d inserted", inserted)
	}
	queries := recorder.Queries()
	if len(queries) != 6 || queries[0].SQL != databasetest.Begin || queries[5].SQL != databasetest.Commit {
		t.Fatalf("backfill must run in one transaction, got %v", queries)
	}
	pair := queries[1]
	assertContains(t, pair.SQL,
		"invoice_id, receivable_id, amount",
		"JOIN invoice ON invoice.id = receivable.invoice_id",
		"invoice.paid_amount = receivable.paid_amount",
		"billing_payment.invoice_id = invoice.id OR billing_payment.receivable_id = receivable.id",
	)
	for index, column := range []string{"invoice_id", "receivable_id", "payable_id"} {
		statement := queries[index+2].SQL
		assertContains(t, statement, "paid_amount > 0", "NOT EXISTS (SELECT 1 FROM billing_payment WHERE billing_payment."+column)
	}
	for _, query := range queries[1:5] {
		assertContains(t, query.SQL, "::date")
		if len(query.Args) != 2 || query.Args[0] != backfillMethod || query.Args[1] != backfillNote {
			t.Fatalf("got args %v", query.Args)
		}
	}
	if !strings.Contains(queries[1].SQL, "receivable_id") || strings.Contains(queries[4].SQL, "receivable") {
		t.Fatal("linked pairs are filled first and payables never touch receivables")
	}
}

func TestSecondBackfillRunInsertsNothing(t *testing.T) {
	db, recorder := databasetest.Open(t)

	inserted, err := BackfillPayments(context.Background(), NewRepository(db))
	if err != nil || inserted != 0 {
		t.Fatalf("got %d %v", inserted, err)
	}
	for _, query := range recorder.Queries()[1:5] {
		assertContains(t, query.SQL, "AND NOT EXISTS (")
	}
}

func receivableLockResult(id, customerID, invoiceID uuid.UUID, at time.Time, paid int64) databasetest.Result {
	return databasetest.Result{
		Columns: []string{"id", "created_at", "updated_at", "customer_id", "invoice_id", "reference", "due_date", "amount", "paid_amount", "status"},
		Rows:    [][]driver.Value{{id.String(), at, at, customerID.String(), invoiceID.String(), "PTG-1", at, int64(1000), paid, "not_due"}},
	}
}

func TestReceivablePaymentLocksInvoiceBeforeReceivableInSQL(t *testing.T) {
	receivableID, customerID, invoiceID := uuid.New(), uuid.New(), uuid.New()
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	invoice := databasetest.Result{
		Columns: []string{"id", "created_at", "updated_at", "number", "party_type", "party_id", "due_date", "amount", "paid_amount", "status"},
		Rows:    [][]driver.Value{{invoiceID.String(), at, at, "INV-1", customerParty, customerID.String(), at, int64(1000), int64(0), "not_due"}},
	}
	db, recorder := databasetest.Open(t,
		receivableLockResult(receivableID, customerID, invoiceID, at, 0),
		invoice,
		receivableLockResult(receivableID, customerID, invoiceID, at, 0),
		databasetest.Result{Affected: 1},
		databasetest.Result{Affected: 1},
		databasetest.Result{Affected: 1},
	)

	_, _ = fixedService(NewRepository(db), at).PayReceivable(context.Background(), receivableID, PaymentRequest{Amount: 250}, nil)

	var statements []string
	for _, query := range recorder.Queries() {
		statements = append(statements, query.SQL)
	}
	if statements[0] != databasetest.Begin {
		t.Fatalf("payment must start a transaction, got %v", statements)
	}
	assertContains(t, statements[2], `FROM "invoice"`, "FOR UPDATE")
	assertContains(t, statements[3], `FROM "receivable"`, "FOR UPDATE")
	assertContains(t, statements[4], `INSERT INTO "billing_payment"`, `"invoice_id"`, `"receivable_id"`)
	assertContains(t, statements[5], `UPDATE "invoice"`)
	assertContains(t, statements[6], `UPDATE "receivable"`)
}

func payableLockResult(id, vendorID uuid.UUID, at time.Time) databasetest.Result {
	return databasetest.Result{
		Columns: []string{"id", "created_at", "updated_at", "vendor_id", "reference", "due_date", "amount", "paid_amount", "status"},
		Rows:    [][]driver.Value{{id.String(), at, at, vendorID.String(), "UJI-UTG", at, int64(5000), int64(0), "not_due"}},
	}
}

func servePaymentRoute(t *testing.T, db *gorm.DB, method, path, body string) (*httptest.ResponseRecorder, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tokens := token.NewManager(config.Token{AccessSecret: "billing-payment-test", AccessTTL: time.Minute, Issuer: "billing-payment-test"})
	userID := uuid.New()
	access, err := tokens.IssueAccess(userID)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	NewModule(db, tokens).RegisterRoutes(router.Group("/api"))
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+access.Value)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response, userID
}

func TestPaymentRouteRecordsWhoPaidAndTheDetails(t *testing.T) {
	payableID, vendorID := uuid.New(), uuid.New()
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t,
		payableLockResult(payableID, vendorID, at),
		databasetest.Result{Affected: 1},
		databasetest.Result{Affected: 1},
		payableLockResult(payableID, vendorID, at),
	)

	response, userID := servePaymentRoute(t, db, http.MethodPost, "/api/billing/payables/"+payableID.String()+"/payments",
		`{"amount":1500,"paidAt":"2026-09-30T00:00:00Z","method":"tunai","reference":"KW-01","note":"DP"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	var insert databasetest.Query
	for _, query := range recorder.Queries() {
		if strings.HasPrefix(query.SQL, `INSERT INTO "billing_payment"`) {
			insert = query
		}
	}
	joined := fmt.Sprint(insert.Args)
	for _, want := range []string{payableID.String(), userID.String(), "tunai", "KW-01", "DP", "1500"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("insert args %s are missing %s", joined, want)
		}
	}
}

func TestPaymentRouteRejectsUnknownMethod(t *testing.T) {
	db, recorder := databasetest.Open(t)

	response, _ := servePaymentRoute(t, db, http.MethodPost, "/api/billing/invoices/"+uuid.NewString()+"/payments", `{"amount":10,"method":"bitcoin"}`)
	if response.Code != http.StatusBadRequest || len(recorder.Queries()) != 0 {
		t.Fatalf("got status %d with %d queries", response.Code, len(recorder.Queries()))
	}
}

func TestPaymentHistoryRouteIsPaged(t *testing.T) {
	payableID, vendorID := uuid.New(), uuid.New()
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, payableLockResult(payableID, vendorID, at), countResult, databasetest.Result{
		Columns: []string{"id", "created_at", "updated_at", "payable_id", "amount", "paid_at", "method", "reference", "note"},
		Rows:    [][]driver.Value{{uuid.NewString(), at, at, payableID.String(), int64(1500), at, "tunai", "KW-01", "DP"}},
	})

	response, _ := servePaymentRoute(t, db, http.MethodGet, "/api/billing/payables/"+payableID.String()+"/payments?page=1&pageSize=5", "")
	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Data struct {
			Items      []PaymentResponse `json:"items"`
			PageSize   int               `json:"pageSize"`
			TotalItems int64             `json:"totalItems"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data.Items) != 1 || body.Data.PageSize != 5 || body.Data.Items[0].Method != "tunai" || *body.Data.Items[0].PayableID != payableID {
		t.Fatalf("got %s", response.Body.String())
	}
	assertContains(t, recorder.Queries()[2].SQL, "billing_payment.payable_id =", "LIMIT $2")
	if args := recorder.Queries()[2].Args; args[len(args)-1] != int64(5) {
		t.Fatalf("got args %v", args)
	}
}
