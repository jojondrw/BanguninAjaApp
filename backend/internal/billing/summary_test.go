package billing

import "testing"

func TestBalanceSummaryFillsEveryBucketAndSumsOverdue(t *testing.T) {
	summary := balanceSummary(
		[]AgingRow{
			{Bucket: "current", Count: 3, Amount: 300},
			{Bucket: "days_60", Count: 1, Amount: 50},
			{Bucket: "days_over_90", Count: 2, Amount: 20},
		},
		[]PartyBalanceRow{{Name: "PT Contoh", Count: 2, Amount: 250}},
	)

	if len(summary.Buckets) != len(agingBuckets) {
		t.Fatalf("buckets = %d, want %d", len(summary.Buckets), len(agingBuckets))
	}
	if summary.Buckets[1].Bucket != "days_30" || summary.Buckets[1].Amount != 0 {
		t.Errorf("empty bucket = %+v, want zero days_30", summary.Buckets[1])
	}
	if summary.Outstanding != 370 || summary.Overdue != 70 || summary.OpenCount != 6 {
		t.Errorf("totals = outstanding %d overdue %d open %d", summary.Outstanding, summary.Overdue, summary.OpenCount)
	}
	if len(summary.Largest) != 1 || summary.Largest[0].Name != "PT Contoh" {
		t.Errorf("largest = %+v", summary.Largest)
	}
}

func TestBalanceSummaryWithoutBalancesKeepsEmptyLists(t *testing.T) {
	summary := balanceSummary(nil, nil)
	if summary.Largest == nil || summary.Outstanding != 0 || len(summary.Buckets) != len(agingBuckets) {
		t.Errorf("summary = %+v", summary)
	}
}
