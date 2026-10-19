package sales

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectUnitSummariesPivotStatusesPerProject(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	counts := []ProjectUnitCount{
		{ProjectID: first, Status: "available", Total: 4, Value: 2_000},
		{ProjectID: first, Status: "on_hold", Total: 1, Value: 500},
		{ProjectID: first, Status: "reserved", Total: 2, Value: 900},
		{ProjectID: first, Status: "sold", Total: 3, Value: 1_500},
		{ProjectID: second, Status: "sold", Total: 6, Value: 3_000},
	}

	summaries := projectUnitSummaries(counts)

	if len(summaries) != 2 {
		t.Fatalf("summaries = %d, want 2", len(summaries))
	}
	got := summaries[0]
	if got.ProjectID != first || got.Total != 10 || got.Available != 4 || got.OnHold != 1 || got.Reserved != 2 || got.Sold != 3 {
		t.Errorf("first project counts = %+v", got)
	}
	if got.StockValue != 2_500 || got.SoldValue != 1_500 {
		t.Errorf("first project values = stock %d sold %d, want 2500 and 1500", got.StockValue, got.SoldValue)
	}
	if summaries[1].Sold != 6 || summaries[1].Available != 0 || summaries[1].SoldValue != 3_000 {
		t.Errorf("second project = %+v", summaries[1])
	}
}

func TestProjectUnitSummariesOfNoUnitsIsEmptyList(t *testing.T) {
	summaries := projectUnitSummaries(nil)
	if summaries == nil || len(summaries) != 0 {
		t.Errorf("summaries = %#v, want empty non-nil list", summaries)
	}
}
