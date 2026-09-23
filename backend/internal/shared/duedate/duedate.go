package duedate

import "time"

const (
	NotDue  = "not_due"
	Due     = "due"
	Paid    = "paid"
	Overdue = "overdue"

	dueSoonDays = 7
	hoursPerDay = 24
)

type Range struct {
	Settled *bool
	From    *time.Time
	Before  *time.Time
}

func Today(now time.Time) time.Time {
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func Status(dueDate time.Time, settled bool, today time.Time) string {
	due := Today(dueDate)
	switch {
	case settled:
		return Paid
	case due.Before(today):
		return Overdue
	case due.Before(dueSoonLimit(today)):
		return Due
	default:
		return NotDue
	}
}

func DaysOverdue(dueDate time.Time, settled bool, today time.Time) int {
	due := Today(dueDate)
	if settled || !due.Before(today) {
		return 0
	}
	return int(today.Sub(due).Hours() / hoursPerDay)
}

func RangeOf(status string, today time.Time) Range {
	settled := status == Paid
	limit := dueSoonLimit(today)
	switch status {
	case Paid:
		return Range{Settled: &settled}
	case Overdue:
		return Range{Settled: &settled, Before: &today}
	case Due:
		return Range{Settled: &settled, From: &today, Before: &limit}
	case NotDue:
		return Range{Settled: &settled, From: &limit}
	default:
		return Range{}
	}
}

func dueSoonLimit(today time.Time) time.Time {
	return today.AddDate(0, 0, dueSoonDays+1)
}
