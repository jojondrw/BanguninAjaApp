package procurement

import "testing"

func TestProcurementOverviewMergesStatusesAndKeepsVendorOrder(t *testing.T) {
	overview := newProcurementOverview(
		[]StatusCount{{Status: "approved", Count: 3}, {Status: "draft", Count: 1}},
		PurchaseOrderTotals{Count: 5, Sent: 2, PartiallyReceived: 1, OpenValue: 900, TotalValue: 1_500},
		LateOrders{Count: 2, Value: 400},
		ReceiptHealth{Total: 8, Damaged: 2},
		[]StatusCount{{Status: "good", Count: 4}, {Status: "poor", Count: 1}},
		[]VendorValue{{Name: "PT A", Rating: "good", Orders: 3, Value: 800}, {Name: "CV B", Rating: "fair", Orders: 1, Value: 200}},
	)

	if overview.Requests["approved"] != 3 || overview.Requests["draft"] != 1 {
		t.Errorf("requests = %v", overview.Requests)
	}
	if overview.LateOrders != 2 || overview.LateValue != 400 || overview.DamagedRecent != 2 || overview.ReceiptDays != recentReceiptDays {
		t.Errorf("overview = %+v", overview)
	}
	if overview.Orders.OpenValue != 900 || overview.VendorRatings["poor"] != 1 {
		t.Errorf("orders or ratings wrong: %+v %v", overview.Orders, overview.VendorRatings)
	}
	if len(overview.LargestVendors) != 2 || overview.LargestVendors[0].Name != "PT A" {
		t.Errorf("vendors = %+v", overview.LargestVendors)
	}
}
