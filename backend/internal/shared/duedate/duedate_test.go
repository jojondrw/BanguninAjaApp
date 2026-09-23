package duedate

import (
	"testing"
	"time"
)

func TestStatus(t *testing.T) {
	today := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		dueDate time.Time
		settled bool
		want    string
	}{
		{today.AddDate(0, 0, -1), false, Overdue},
		{today, false, Due},
		{today.AddDate(0, 0, dueSoonDays), false, Due},
		{today.AddDate(0, 0, dueSoonDays+1), false, NotDue},
		{today.AddDate(0, 0, -30), true, Paid},
	}
	for _, tc := range cases {
		if got := Status(tc.dueDate, tc.settled, today); got != tc.want {
			t.Errorf("due %s settled %v: got %s, want %s", tc.dueDate.Format(time.DateOnly), tc.settled, got, tc.want)
		}
	}
}

func TestDaysOverdue(t *testing.T) {
	today := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	if got := DaysOverdue(today.AddDate(0, 0, -45), false, today); got != 45 {
		t.Fatalf("got %d, want 45", got)
	}
	if got := DaysOverdue(today.AddDate(0, 0, -45), true, today); got != 0 {
		t.Fatalf("settled must not age, got %d", got)
	}
	if got := DaysOverdue(today, false, today); got != 0 {
		t.Fatalf("due today is not overdue, got %d", got)
	}
}

func TestRangeOfMatchesStatus(t *testing.T) {
	today := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	for _, status := range []string{NotDue, Due, Overdue} {
		window := RangeOf(status, today)
		if window.Settled == nil || *window.Settled {
			t.Fatalf("%s must select unsettled rows", status)
		}
	}
	if window := RangeOf(Paid, today); window.Settled == nil || !*window.Settled || window.From != nil {
		t.Fatalf("paid must select settled rows only, got %+v", window)
	}
	if window := RangeOf("", today); window.Settled != nil {
		t.Fatalf("empty status must not filter, got %+v", window)
	}
}
