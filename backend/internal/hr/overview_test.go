package hr

import (
	"testing"
	"time"
)

func TestRecentPeriodsEndWithThisMonth(t *testing.T) {
	periods := recentPeriods(time.Date(2026, 2, 19, 0, 0, 0, 0, time.UTC), 3)
	if len(periods) != 3 || periods[0] != "2025-12" || periods[2] != "2026-02" {
		t.Errorf("periods = %v", periods)
	}
}

func TestHROverviewFillsMissingPayrollMonthsWithZero(t *testing.T) {
	overview := newHROverview(
		[]EmploymentCount{{EmploymentType: "permanent", Total: 7}, {EmploymentType: "daily", Total: 5}},
		[]ProjectHeadcount{{Total: 3}},
		[]DailyAttendance{{Date: time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC), Present: 9, Absent: 1}},
		[]string{"2026-09", "2026-10"},
		[]MonthlyPayroll{{Period: "2026-10", NetPay: 500, Unpaid: 500, Slips: 4}},
	)

	if overview.Headcount != 12 || overview.Days[0].Date != "2026-10-19" || overview.Days[0].Present != 9 {
		t.Errorf("overview = %+v", overview)
	}
	if overview.Payrolls[0].Period != "2026-09" || overview.Payrolls[0].NetPay != 0 || overview.Payrolls[1].Unpaid != 500 {
		t.Errorf("payrolls = %+v", overview.Payrolls)
	}
}
