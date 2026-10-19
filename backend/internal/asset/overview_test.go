package asset

import "testing"

func TestAssetOverviewSumsBookValueAndMergesStatuses(t *testing.T) {
	overview := newAssetOverview(
		[]CategoryValue{{Category: "Alat berat", Count: 3, BookValue: 900}, {Category: "Kendaraan", Count: 2, BookValue: 100}},
		[]EquipmentHealth{{Status: "operating", Count: 5}, {Status: "broken", Count: 1}},
		ServiceDue{Overdue: 2, Soon: 3},
	)

	if overview.BookValue != 1_000 || len(overview.Categories) != 2 || overview.Categories[0].Category != "Alat berat" {
		t.Errorf("categories = %+v", overview)
	}
	if overview.EquipmentStatus["operating"] != 5 || overview.EquipmentStatus["broken"] != 1 {
		t.Errorf("status = %v", overview.EquipmentStatus)
	}
	if overview.ServiceOverdue != 2 || overview.ServiceSoon != 3 || overview.ServiceWindow != serviceWindowDays {
		t.Errorf("service = %+v", overview)
	}
}
