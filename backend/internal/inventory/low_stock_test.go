package inventory

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func TestListLowStockMaterialsCarriesTheUnitCode(t *testing.T) {
	materialID, unitID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t, countResult, databasetest.Result{
		Columns: []string{"material_id", "code", "name", "unit_of_measure_id", "unit_of_measure_code", "minimum_stock", "total_quantity"},
		Rows:    [][]driver.Value{{materialID.String(), "UJI-SMN", "Semen", unitID.String(), "sak", float64(50), float64(12)}},
	})

	levels, total, err := NewRepository(db).ListLowStockMaterials(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(levels) != 1 {
		t.Fatalf("got total %d, levels %+v", total, levels)
	}
	if levels[0].UnitOfMeasureCode != "sak" || levels[0].UnitOfMeasureID != unitID {
		t.Fatalf("got %+v", levels[0])
	}

	queries := recorder.Queries()
	assertContains(t, queries[1].SQL,
		"LEFT JOIN unit_of_measure u ON u.id = m.unit_of_measure_id",
		"COALESCE(u.code, '') AS unit_of_measure_code",
		lowStockCondition,
	)

	response := newLowStockResponse(levels[0])
	if response.UnitOfMeasureCode != "sak" || response.TotalQuantity != 12 {
		t.Fatalf("got response %+v", response)
	}
}
