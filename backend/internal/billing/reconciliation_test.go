package billing

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

func reconciliationResult(invoiceID, receivableID uuid.UUID) databasetest.Result {
	return databasetest.Result{
		Columns: []string{
			"invoice_id", "invoice_number", "invoice_amount", "invoice_paid_amount",
			"receivable_id", "receivable_reference", "receivable_amount", "receivable_paid_amount", "customer_name",
		},
		Rows: [][]driver.Value{{
			invoiceID.String(), "UJI-INV-7", int64(10000000), int64(6000000),
			receivableID.String(), "UJI-PTG-7", int64(10000000), int64(4000000), "Budi",
		}},
	}
}

func TestReconciliationRouteListsLinkedPairsWhosePaidAmountsDiffer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	invoiceID, receivableID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, countResult, reconciliationResult(invoiceID, receivableID))
	tokens := token.NewManager(config.Token{AccessSecret: "billing-module-test", AccessTTL: time.Minute, Issuer: "billing-module-test"})
	access, err := tokens.IssueAccess(uuid.New())
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	NewModule(db, tokens).RegisterRoutes(router.Group("/api"))
	request := httptest.NewRequest(http.MethodGet, "/api/billing/reconciliation?pageSize=5", nil)
	request.Header.Set("Authorization", "Bearer "+access.Value)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Data struct {
			Items      []ReconciliationResponse `json:"items"`
			TotalItems int64                    `json:"totalItems"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.TotalItems != 1 || len(body.Data.Items) != 1 {
		t.Fatalf("got %s", response.Body.String())
	}
	item := body.Data.Items[0]
	if item.InvoiceID != invoiceID || item.ReceivableID != receivableID || item.InvoiceNumber != "UJI-INV-7" {
		t.Fatalf("got %+v", item)
	}
	if item.InvoicePaidAmount != 6000000 || item.ReceivablePaidAmount != 4000000 || item.Difference != 2000000 {
		t.Fatalf("got amounts %+v", item)
	}
	if item.CustomerName == nil || *item.CustomerName != "Budi" {
		t.Fatalf("got customer %v", item.CustomerName)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want count and rows", len(queries))
	}
	for _, query := range queries {
		assertContains(t, query.SQL,
			"JOIN invoice ON invoice.id = receivable.invoice_id",
			"invoice.paid_amount <> receivable.paid_amount",
		)
	}
	assertContains(t, queries[1].SQL, "LEFT JOIN customer ON customer.id = receivable.customer_id", "ORDER BY invoice.number ASC", "LIMIT $1")
}

func TestReconciliationRouteNeedsAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, recorder := databasetest.Open(t)
	tokens := token.NewManager(config.Token{AccessSecret: "billing-module-test", AccessTTL: time.Minute, Issuer: "billing-module-test"})

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	NewModule(db, tokens).RegisterRoutes(router.Group("/api"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/billing/reconciliation", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", response.Code)
	}
	if len(recorder.Queries()) != 0 {
		t.Fatalf("an anonymous request must not reach the database")
	}
}
