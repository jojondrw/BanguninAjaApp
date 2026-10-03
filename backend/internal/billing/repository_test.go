package billing

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

var countResult = databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}}

func assertContains(t *testing.T, query string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(query, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query)
		}
	}
}

func receivableResult(id, customerID, projectID, contractID, invoiceID uuid.UUID, at time.Time) databasetest.Result {
	return databasetest.Result{
		Columns: []string{
			"id", "created_at", "updated_at", "customer_id", "project_id", "contract_id", "invoice_id", "reference",
			"due_date", "amount", "paid_amount", "status", "customer_name", "project_name", "contract_number", "invoice_number",
		},
		Rows: [][]driver.Value{{
			id.String(), at, at, customerID.String(), projectID.String(), contractID.String(), invoiceID.String(), "UJI-PTG",
			at, int64(1000), int64(0), "not_due", "Budi", "Griya Asri", "KTR-1", "INV-1",
		}},
	}
}

func TestListReceivablesJoinsLinkedNamesAndFiltersByProjectAndContract(t *testing.T) {
	id, customerID, projectID, contractID, invoiceID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	at := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	settled := false
	db, recorder := databasetest.Open(t, countResult, receivableResult(id, customerID, projectID, contractID, invoiceID, at))

	rows, total, err := NewRepository(db).ListReceivables(context.Background(), ReceivableFilter{
		DueFilter:  DueFilter{Settled: &settled},
		ProjectID:  &projectID,
		ContractID: &contractID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("got total %d rows %d", total, len(rows))
	}
	row := rows[0]
	if row.ID != id || *row.ProjectID != projectID || *row.ContractID != contractID || *row.InvoiceID != invoiceID {
		t.Fatalf("got %+v", row.Receivable)
	}
	if row.CustomerName != "Budi" || *row.ProjectName != "Griya Asri" || *row.ContractNumber != "KTR-1" || *row.InvoiceNumber != "INV-1" {
		t.Fatalf("got names %+v", row)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	for _, query := range queries {
		assertContains(t, query.SQL,
			"JOIN customer ON customer.id = receivable.customer_id",
			"LEFT JOIN project ON project.id = receivable.project_id",
			"LEFT JOIN contract ON contract.id = receivable.contract_id",
			"LEFT JOIN invoice ON invoice.id = receivable.invoice_id",
			"receivable.status <>",
			"receivable.project_id =",
			"receivable.contract_id =",
		)
	}
	assertContains(t, queries[1].SQL, "contract.number AS contract_number", "invoice.number AS invoice_number", "ORDER BY receivable.due_date ASC")
}

func TestListInvoicesCanHideInvoicesAlreadyRecordedAsReceivables(t *testing.T) {
	recorded := false
	db, recorder := databasetest.Open(t, countResult)

	if _, _, err := NewRepository(db).ListInvoices(context.Background(), InvoiceFilter{Recorded: &recorded, PartyType: customerParty, Limit: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[0].SQL, "LEFT JOIN receivable ON receivable.invoice_id = invoice.id", "receivable.id IS NULL", "invoice.party_type =")
	assertContains(t, queries[1].SQL, "receivable.id AS receivable_id", "receivable.reference AS receivable_reference", "project.name AS project_name")
}

func TestFindInvoiceRowReportsTheReceivableItWasRecordedAs(t *testing.T) {
	id, customerID, receivableID := uuid.New(), uuid.New(), uuid.New()
	at := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	db, _ := databasetest.Open(t, databasetest.Result{
		Columns: []string{
			"id", "created_at", "updated_at", "number", "note", "party_type", "party_id", "project_id", "due_date",
			"amount", "paid_amount", "status", "project_name", "receivable_id", "receivable_reference",
		},
		Rows: [][]driver.Value{{
			id.String(), at, at, "INV-9", "", customerParty, customerID.String(), nil, at,
			int64(1000), int64(0), "not_due", nil, receivableID.String(), "PTG-9",
		}},
	})

	row, err := NewRepository(db).FindInvoiceRow(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response := newInvoiceResponse(row, at)
	if response.ReceivableID == nil || *response.ReceivableID != receivableID || *response.ReceivableReference != "PTG-9" {
		t.Fatalf("got %+v", response)
	}
	if response.ProjectID != nil || response.ProjectName != nil {
		t.Fatalf("an invoice without project must not report one, got %+v", response)
	}
}

func TestListPayablesJoinsPurchaseOrderAndFiltersByIt(t *testing.T) {
	orderID, projectID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, countResult)

	if _, _, err := NewRepository(db).ListPayables(context.Background(), PayableFilter{PurchaseOrderID: &orderID, ProjectID: &projectID, Limit: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[0].SQL,
		"JOIN vendor ON vendor.id = payable.vendor_id",
		"LEFT JOIN purchase_order ON purchase_order.id = payable.purchase_order_id",
		"payable.purchase_order_id =",
		"payable.project_id =",
	)
	assertContains(t, queries[1].SQL, "purchase_order.number AS purchase_order_number", "vendor.name AS vendor_name")
}

func TestReceivableListRouteBindsProjectAndContractFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id, customerID, projectID, contractID, invoiceID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	at := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, countResult, receivableResult(id, customerID, projectID, contractID, invoiceID, at))
	tokens := token.NewManager(config.Token{AccessSecret: "billing-module-test", AccessTTL: time.Minute, Issuer: "billing-module-test"})
	access, err := tokens.IssueAccess(uuid.New())
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	NewModule(db, tokens).RegisterRoutes(router.Group("/api"))
	request := httptest.NewRequest(http.MethodGet, "/api/billing/receivables?projectId="+projectID.String()+"&contractId="+contractID.String(), nil)
	request.Header.Set("Authorization", "Bearer "+access.Value)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Data struct {
			Items []ReceivableResponse `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data.Items) != 1 || *body.Data.Items[0].ContractNumber != "KTR-1" || *body.Data.Items[0].ProjectName != "Griya Asri" {
		t.Fatalf("got %s", response.Body.String())
	}
	args := recorder.Queries()[0].Args
	if len(args) != 2 || args[0] != projectID.String() || args[1] != contractID.String() {
		t.Fatalf("got args %v", args)
	}
}
