package procurement

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var purchaseOrderTotalsResult = databasetest.Result{
	Columns: []string{"count", "draft", "sent", "partially_received", "completed", "cancelled", "open_value", "total_value"},
	Rows:    [][]driver.Value{{int64(160), int64(20), int64(30), int64(15), int64(85), int64(10), "650000000", "4100000000"}},
}

func TestSummarizePurchaseOrdersAggregatesInDatabase(t *testing.T) {
	db, recorder := databasetest.Open(t, purchaseOrderTotalsResult)
	projectID := uuid.New()

	totals, err := NewRepository(db, newInventoryLedger).SummarizePurchaseOrders(context.Background(), PurchaseOrderFilter{ProjectID: &projectID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := PurchaseOrderTotals{Count: 160, Draft: 20, Sent: 30, PartiallyReceived: 15, Completed: 85, Cancelled: 10, OpenValue: 650000000, TotalValue: 4100000000}
	if totals != want {
		t.Fatalf("got %+v, want %+v", totals, want)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want a single aggregate query", len(queries))
	}
	query := queries[0]
	for _, fragment := range []string{
		"COUNT(*) FILTER (WHERE purchase_order.status = $3) AS partially_received",
		"COALESCE(SUM(purchase_order.value) FILTER (WHERE purchase_order.status IN ($6,$7,$8)), 0) AS open_value",
		"COALESCE(SUM(purchase_order.value) FILTER (WHERE purchase_order.status <> $9), 0) AS total_value",
		"purchase_order.project_id = $10",
	} {
		if !strings.Contains(query.SQL, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query.SQL)
		}
	}
	if strings.Contains(query.SQL, "LIMIT") {
		t.Errorf("summary must cover every purchase order, got a limited query:\n%s", query.SQL)
	}
	wantArgs := []any{orderDraft, orderSent, orderPartiallyReceived, orderCompleted, orderCancelled, orderDraft, orderSent, orderPartiallyReceived, orderCancelled, projectID}
	if len(query.Args) != len(wantArgs) {
		t.Fatalf("got args %v, want %v", query.Args, wantArgs)
	}
	for index, arg := range wantArgs[:len(wantArgs)-1] {
		if query.Args[index] != arg {
			t.Errorf("arg %d: got %v, want %v", index, query.Args[index], arg)
		}
	}
}

func TestSummarizePurchaseOrdersWithoutProjectCoversAllOrders(t *testing.T) {
	db, recorder := databasetest.Open(t, purchaseOrderTotalsResult)

	if _, err := NewRepository(db, newInventoryLedger).SummarizePurchaseOrders(context.Background(), PurchaseOrderFilter{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := recorder.Queries()[0]
	if strings.Contains(query.SQL, "project_id") || len(query.Args) != 9 {
		t.Fatalf("unfiltered summary must not narrow rows, got %s %v", query.SQL, query.Args)
	}
}

func TestPurchaseOrderSummaryOpenCountAddsOpenStatuses(t *testing.T) {
	response := newPurchaseOrderSummaryResponse(PurchaseOrderTotals{Count: 9, Draft: 1, Sent: 2, PartiallyReceived: 3, Completed: 2, Cancelled: 1, OpenValue: 700, TotalValue: 900})

	if response.OpenCount != 6 {
		t.Fatalf("got open count %d, want 6", response.OpenCount)
	}
	if response.OpenValue != 700 || response.TotalValue != 900 || response.Count != 9 {
		t.Fatalf("got %+v", response)
	}
	if response.ByStatus.PartiallyReceived != 3 || response.ByStatus.Cancelled != 1 {
		t.Fatalf("got by status %+v", response.ByStatus)
	}
}

func TestGoodsReceiptItemResponseCarriesUnitCodes(t *testing.T) {
	response := newGoodsReceiptItemResponse(GoodsReceiptLine{UnitOfMeasureCode: "sak", StockUnitOfMeasureCode: "kg"})

	if response.UnitOfMeasureCode != "sak" || response.StockUnitOfMeasureCode != "kg" {
		t.Fatalf("got order unit %q and stock unit %q", response.UnitOfMeasureCode, response.StockUnitOfMeasureCode)
	}
}
