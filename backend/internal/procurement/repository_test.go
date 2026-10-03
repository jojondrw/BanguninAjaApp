package procurement

import (
	"context"
	"database/sql/driver"
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

type receiptFixture struct {
	orderID, lineID, materialID, unitID, warehouseID uuid.UUID
	date                                             time.Time
}

func newReceiptFixture() receiptFixture {
	return receiptFixture{
		orderID: uuid.New(), lineID: uuid.New(), materialID: uuid.New(), unitID: uuid.New(), warehouseID: uuid.New(),
		date: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
	}
}

func (f receiptFixture) receipt() GoodsReceipt {
	return GoodsReceipt{Number: "UJI-BPB-010", PurchaseOrderID: f.orderID, WarehouseID: f.warehouseID, Date: f.date, Condition: "good"}
}

func (f receiptFixture) results(stockWrite databasetest.Result) []databasetest.Result {
	return []databasetest.Result{
		{Columns: []string{"id", "status"}, Rows: [][]driver.Value{{f.orderID.String(), orderSent}}},
		{
			Columns: []string{"id", "purchase_order_id", "material_id", "quantity", "unit_of_measure_id", "unit_price"},
			Rows:    [][]driver.Value{{f.lineID.String(), f.orderID.String(), f.materialID.String(), "10.00", f.unitID.String(), int64(65000)}},
		},
		{},
		{},
		{Columns: []string{"id", "name", "unit_of_measure_id"}, Rows: [][]driver.Value{{f.materialID.String(), "Semen Portland 50 kg", f.unitID.String()}}},
		{},
		stockWrite,
		{},
		{Affected: 1},
	}
}

func (f receiptFixture) receive(t *testing.T, stockWrite databasetest.Result) ([]databasetest.Query, error) {
	t.Helper()
	db, recorder := databasetest.Open(t, f.results(stockWrite)...)
	repository := NewRepository(db, newInventoryLedger)
	receipt := f.receipt()
	requests := []GoodsReceiptItemRequest{{PurchaseOrderItemID: f.lineID, AcceptedQuantity: 8, RejectedQuantity: 2}}

	err := repository.Transaction(context.Background(), func(transaction Repository) error {
		return receiveGoods(context.Background(), transaction, &receipt, requests)
	})
	return recorder.Queries(), err
}

func statements(queries []databasetest.Query) []string {
	return pagination.Map(queries, func(query databasetest.Query) string { return query.SQL })
}

func assertContains(t *testing.T, query string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(query, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query)
		}
	}
}

func assertArgs(t *testing.T, query databasetest.Query, wanted ...any) {
	t.Helper()
	for _, want := range wanted {
		if !slices.Contains(query.Args, want) {
			t.Errorf("query is missing argument %v, got %v:\n%s", want, query.Args, query.SQL)
		}
	}
}

func TestStockLedgerWritesInMovementAndUpsertsStock(t *testing.T) {
	materialID, warehouseID := uuid.New(), uuid.New()
	date := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, databasetest.Result{}, databasetest.Result{Affected: 1})

	movementID, err := NewRepository(db, newInventoryLedger).StockLedger().RecordIncoming(context.Background(), IncomingStock{
		Date: date, MaterialID: materialID, WarehouseID: warehouseID, Quantity: 12.5, Reference: "UJI-BPB-010",
	})
	if err != nil || movementID == uuid.Nil {
		t.Fatalf("got id %s error %v", movementID, err)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d statements, want movement and stock: %v", len(queries), statements(queries))
	}
	assertContains(t, queries[0].SQL, `INSERT INTO "stock_movement"`)
	assertArgs(t, queries[0], movementID.String(), "in", materialID.String(), 12.5, warehouseID.String(), "UJI-BPB-010", date)
	assertContains(t, queries[1].SQL,
		"INSERT INTO stock (id, material_id, warehouse_id, quantity, created_at, updated_at)",
		"ON CONFLICT (material_id, warehouse_id)",
		"DO UPDATE SET quantity = stock.quantity + EXCLUDED.quantity",
	)
	assertArgs(t, queries[1], materialID.String(), warehouseID.String(), 12.5)
}

func TestReceiptCommitsReceiptStockAndOrderInOneTransaction(t *testing.T) {
	fixture := newReceiptFixture()

	queries, err := fixture.receive(t, databasetest.Result{Affected: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		databasetest.Begin,
		`FROM "purchase_order"`,
		`FROM "purchase_order_item"`,
		"SUM(goods_receipt_item.accepted_quantity)",
		`INSERT INTO "goods_receipt"`,
		`SELECT id, name, unit_of_measure_id FROM "material"`,
		`INSERT INTO "stock_movement"`,
		"INSERT INTO stock ",
		`INSERT INTO "goods_receipt_item"`,
		`UPDATE "purchase_order"`,
		databasetest.Commit,
	}
	if len(queries) != len(want) {
		t.Fatalf("got %d statements, want %d: %v", len(queries), len(want), statements(queries))
	}
	for index, fragment := range want {
		assertContains(t, queries[index].SQL, fragment)
	}

	movementID, ok := queries[6].Args[0].(string)
	if !ok {
		t.Fatalf("movement id argument is %T", queries[6].Args[0])
	}
	assertArgs(t, queries[6], "in", fixture.materialID.String(), 8.0, fixture.warehouseID.String(), "UJI-BPB-010", fixture.date)
	assertArgs(t, queries[7], fixture.materialID.String(), fixture.warehouseID.String(), 8.0)
	assertArgs(t, queries[8], fixture.lineID.String(), 8.0, 2.0, movementID)
	assertArgs(t, queries[9], orderPartiallyReceived)
}

func TestReceiptRollsBackWhenStockPostingFails(t *testing.T) {
	fixture := newReceiptFixture()

	queries, err := fixture.receive(t, databasetest.Result{Err: &pgconn.PgError{Code: "23503", Message: "insert or update on table \"stock\" violates foreign key constraint"}})

	failure := apperror.From(err)
	if failure.Code != receiptStockFailedCode || failure.Status != http.StatusUnprocessableEntity {
		t.Fatalf("got %d %s: %v", failure.Status, failure.Code, err)
	}
	if failure.Message != "Penerimaan barang dibatalkan karena stok Semen Portland 50 kg gagal ditambahkan ke gudang: Material atau gudang tidak ditemukan" {
		t.Fatalf("got message %q", failure.Message)
	}

	executed := statements(queries)
	if executed[len(executed)-1] != databasetest.Rollback || slices.Contains(executed, databasetest.Commit) {
		t.Fatalf("the receipt must be rolled back, got %v", executed)
	}
	for _, statement := range executed {
		if strings.Contains(statement, `INSERT INTO "goods_receipt_item"`) || strings.Contains(statement, `UPDATE "purchase_order"`) {
			t.Fatalf("nothing may be written after the failed stock posting, got %v", executed)
		}
	}
}

func TestListGoodsReceiptLinesCarriesMaterialAndBothUnits(t *testing.T) {
	receiptID, movementID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, databasetest.Result{
		Columns: []string{
			"id", "goods_receipt_id", "purchase_order_item_id", "accepted_quantity", "rejected_quantity", "stock_movement_id", "stock_skip_reason",
			"material_id", "material_name", "unit_of_measure_id", "unit_of_measure_code", "stock_unit_of_measure_id", "stock_unit_of_measure_code",
		},
		Rows: [][]driver.Value{{
			uuid.NewString(), receiptID.String(), uuid.NewString(), "8.00", "2.00", movementID.String(), "",
			uuid.NewString(), "Semen Portland 50 kg", uuid.NewString(), "sak", uuid.NewString(), "kg",
		}},
	})

	lines, err := NewRepository(db, newInventoryLedger).ListGoodsReceiptLines(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines", len(lines))
	}
	line := lines[0]
	if line.MaterialName != "Semen Portland 50 kg" || line.UnitOfMeasureCode != "sak" || line.StockUnitOfMeasureCode != "kg" || line.AcceptedQuantity != 8 {
		t.Fatalf("got %+v", line)
	}
	if line.StockMovementID == nil || *line.StockMovementID != movementID {
		t.Fatalf("got movement %v", line.StockMovementID)
	}
	assertContains(t, recorder.Queries()[0].SQL,
		"JOIN purchase_order_item ON purchase_order_item.id = goods_receipt_item.purchase_order_item_id",
		"JOIN material ON material.id = purchase_order_item.material_id",
		"JOIN unit_of_measure AS order_unit ON order_unit.id = purchase_order_item.unit_of_measure_id",
		"JOIN unit_of_measure AS stock_unit ON stock_unit.id = material.unit_of_measure_id",
		"goods_receipt_item.goods_receipt_id = $1",
	)
}
