package inventory

import (
	"testing"
	"time"
)

func TestInventoryOverviewSumsWarehouseValuesAndFormatsWeeks(t *testing.T) {
	overview := newInventoryOverview(
		[]WarehouseValue{{Name: "Gudang Pusat", Value: 700, Items: 9}, {Name: "Gudang Proyek", Value: 300, Items: 4}},
		3,
		[]WeeklyMovement{{Week: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), InValue: 100, OutValue: 40}},
	)

	if overview.StockValue != 1_000 || overview.LowStock != 3 || len(overview.Warehouses) != 2 {
		t.Errorf("overview = %+v", overview)
	}
	if overview.Weeks[0].Week != "2026-10-12" || overview.Weeks[0].OutValue != 40 {
		t.Errorf("weeks = %+v", overview.Weeks)
	}
}
