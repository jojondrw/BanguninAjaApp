package inventory

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
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

func TestListStockMovementsJoinsMaterialAndWarehouseNames(t *testing.T) {
	movementID, materialID, targetID := uuid.New(), uuid.New(), uuid.New()
	recordedAt := time.Date(2026, 10, 2, 9, 30, 0, 0, time.UTC)
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{
			"id", "date", "type", "material_id", "material_code", "material_name", "unit_of_measure_code", "quantity",
			"source_warehouse_id", "source_warehouse_code", "source_warehouse_name",
			"target_warehouse_id", "target_warehouse_code", "target_warehouse_name", "reference", "created_at",
		},
		Rows: [][]driver.Value{{
			movementID.String(), recordedAt, movementIn, materialID.String(), "MAT-SMN", "Semen Portland 50 kg", "sak", "12.50",
			nil, nil, nil,
			targetID.String(), "GDG-01", "Gudang Pusat", "UJI-REF", recordedAt,
		}},
	})

	rows, total, err := NewRepository(db).ListStockMovements(context.Background(), StockMovementFilter{MaterialID: &materialID, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("got total %d rows %d", total, len(rows))
	}

	row := rows[0]
	if row.ID != movementID || row.MaterialName != "Semen Portland 50 kg" || row.MaterialCode != "MAT-SMN" || row.UnitOfMeasureCode != "sak" || row.Quantity != 12.5 {
		t.Fatalf("got %+v", row)
	}
	if row.SourceWarehouseID != nil || row.SourceWarehouseName != nil {
		t.Fatalf("an incoming movement has no source warehouse, got %+v", row)
	}
	if row.TargetWarehouseID == nil || *row.TargetWarehouseID != targetID || row.TargetWarehouseName == nil || *row.TargetWarehouseName != "Gudang Pusat" {
		t.Fatalf("got target %v %v", row.TargetWarehouseID, row.TargetWarehouseName)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want count and page", len(queries))
	}
	assertContains(t, queries[0].SQL, "stock_movement.material_id = $1")
	assertContains(t, queries[1].SQL,
		"JOIN material ON material.id = stock_movement.material_id",
		"JOIN unit_of_measure ON unit_of_measure.id = material.unit_of_measure_id",
		"LEFT JOIN warehouse AS source_warehouse ON source_warehouse.id = stock_movement.source_warehouse_id",
		"LEFT JOIN warehouse AS target_warehouse ON target_warehouse.id = stock_movement.target_warehouse_id",
		"stock_movement.material_id = $1",
		"ORDER BY stock_movement.date DESC, stock_movement.created_at DESC",
	)
}

func TestListStockMovementsQualifiesWarehouseFilter(t *testing.T) {
	warehouseID := uuid.New()
	db, recorder := databasetest.Open(t, countResult)

	_, _, err := NewRepository(db).ListStockMovements(context.Background(), StockMovementFilter{WarehouseID: &warehouseID, Type: movementTransfer, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, query := range recorder.Queries() {
		assertContains(t, query.SQL,
			"(stock_movement.source_warehouse_id = $1 OR stock_movement.target_warehouse_id = $2)",
			"stock_movement.type = $3",
		)
	}
}

func TestListStocksCarriesUnitOfMeasure(t *testing.T) {
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"id", "material_id", "material_code", "material_name", "unit_of_measure_code", "warehouse_id", "warehouse_code", "warehouse_name", "quantity", "updated_at"},
		Rows:    [][]driver.Value{{uuid.NewString(), uuid.NewString(), "MAT-BSI10", "Besi beton 10 mm", "btg", uuid.NewString(), "GDG-MKS", "Gudang Lapangan Makassar", "30.00", time.Now()}},
	})

	rows, _, err := NewRepository(db).ListStocks(context.Background(), StockFilter{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].UnitOfMeasureCode != "btg" || rows[0].WarehouseName != "Gudang Lapangan Makassar" || rows[0].Quantity != 30 {
		t.Fatalf("got %+v", rows)
	}
	assertContains(t, recorder.Queries()[1].SQL,
		"unit_of_measure.code AS unit_of_measure_code",
		"JOIN unit_of_measure ON unit_of_measure.id = material.unit_of_measure_id",
	)
}

func TestStockMovementResponseKeepsNames(t *testing.T) {
	source := "Gudang Pusat"
	response := newStockMovementResponse(StockMovementRow{MaterialCode: "MAT-SMN", MaterialName: "Semen", UnitOfMeasureCode: "sak", SourceWarehouseName: &source})
	if response.MaterialName != "Semen" || response.MaterialCode != "MAT-SMN" || response.UnitOfMeasureCode != "sak" {
		t.Fatalf("got %+v", response)
	}
	if response.SourceWarehouseName == nil || *response.SourceWarehouseName != source || response.TargetWarehouseName != nil {
		t.Fatalf("got source %v target %v", response.SourceWarehouseName, response.TargetWarehouseName)
	}
}

func TestListStockMovementsFiltersByExactReference(t *testing.T) {
	warehouseID := uuid.New()
	db, recorder := databasetest.Open(t, countResult)

	_, _, err := NewRepository(db).ListStockMovements(context.Background(), StockMovementFilter{WarehouseID: &warehouseID, Type: movementIn, Reference: "UJI-BPB-001", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, query := range recorder.Queries() {
		assertContains(t, query.SQL, "stock_movement.type = $3", "stock_movement.reference = $4")
	}
	if args := recorder.Queries()[0].Args; len(args) != 4 || args[3] != "UJI-BPB-001" {
		t.Fatalf("got args %v", args)
	}
}
