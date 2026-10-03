package procurement

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var countResult = databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}}

func TestListPurchaseOrdersReturnsVendorAndProjectNames(t *testing.T) {
	orderID, vendorID, projectID := uuid.New(), uuid.New(), uuid.New()
	date := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "number", "vendor_id", "project_id", "date", "value", "status", "vendor_name", "project_name"},
		Rows: [][]driver.Value{{
			orderID.String(), "UJI-PO-001", vendorID.String(), projectID.String(), date, int64(1500000), "sent", "UJI Vendor Beton", "UJI Proyek Cibubur",
		}},
	})

	rows, total, err := NewRepository(db, newInventoryLedger).ListPurchaseOrders(context.Background(), PurchaseOrderFilter{VendorID: &vendorID, Status: "sent", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("got total %d rows %d", total, len(rows))
	}
	if rows[0].ID != orderID || rows[0].VendorName != "UJI Vendor Beton" || rows[0].ProjectName != "UJI Proyek Cibubur" {
		t.Fatalf("got %+v", rows[0])
	}

	response := newPurchaseOrderRowResponse(rows[0])
	if response.VendorName != "UJI Vendor Beton" || response.ProjectName != "UJI Proyek Cibubur" || response.Number != "UJI-PO-001" {
		t.Fatalf("got response %+v", response)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want count and page", len(queries))
	}
	assertContains(t, queries[1].SQL,
		"(SELECT vendor.name FROM vendor WHERE vendor.id = purchase_order.vendor_id) AS vendor_name",
		"(SELECT project.name FROM project WHERE project.id = purchase_order.project_id) AS project_name",
		"purchase_order.vendor_id = $1",
		"purchase_order.status = $2",
		"ORDER BY purchase_order.date DESC, purchase_order.number DESC",
	)
}

func TestListPurchaseRequestsReturnsProjectName(t *testing.T) {
	requestID, projectID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "number", "project_id", "status", "item_count", "project_name"},
		Rows:    [][]driver.Value{{requestID.String(), "UJI-PR-001", projectID.String(), "draft", int64(2), "UJI Proyek Cibubur"}},
	})

	rows, _, err := NewRepository(db, newInventoryLedger).ListPurchaseRequests(context.Background(), PurchaseRequestFilter{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].ProjectName != "UJI Proyek Cibubur" || rows[0].ItemCount != 2 {
		t.Fatalf("got %+v", rows)
	}
	if response := newPurchaseRequestRowResponse(rows[0]); response.ProjectName != "UJI Proyek Cibubur" {
		t.Fatalf("got response %+v", response)
	}
	assertContains(t, recorder.Queries()[1].SQL, "(SELECT project.name FROM project WHERE project.id = purchase_request.project_id) AS project_name")
}

func TestListGoodsReceiptsReturnsWarehouseNameAndOrderNumber(t *testing.T) {
	receiptID, orderID, warehouseID := uuid.New(), uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "number", "purchase_order_id", "warehouse_id", "condition", "item_count", "warehouse_name", "purchase_order_number"},
		Rows: [][]driver.Value{{
			receiptID.String(), "UJI-GRN-001", orderID.String(), warehouseID.String(), "good", int64(1), "UJI Gudang Pusat", "UJI-PO-001",
		}},
	})

	rows, _, err := NewRepository(db, newInventoryLedger).ListGoodsReceipts(context.Background(), GoodsReceiptFilter{WarehouseID: &warehouseID, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].WarehouseName != "UJI Gudang Pusat" || rows[0].PurchaseOrderNumber != "UJI-PO-001" {
		t.Fatalf("got %+v", rows)
	}
	response := newGoodsReceiptRowResponse(rows[0])
	if response.WarehouseName != "UJI Gudang Pusat" || response.PurchaseOrderNumber != "UJI-PO-001" {
		t.Fatalf("got response %+v", response)
	}
	assertContains(t, recorder.Queries()[1].SQL,
		"(SELECT warehouse.name FROM warehouse WHERE warehouse.id = goods_receipt.warehouse_id) AS warehouse_name",
		"(SELECT purchase_order.number FROM purchase_order WHERE purchase_order.id = goods_receipt.purchase_order_id) AS purchase_order_number",
		"goods_receipt.warehouse_id = $1",
	)
}
