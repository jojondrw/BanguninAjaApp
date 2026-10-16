package main

import (
	"testing"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
)

func curveWith(actual int, planned float64, points ...project.CurvePointResponse) project.ProgressCurveResponse {
	return project.ProgressCurveResponse{
		Points: points,
		Today:  project.CurveTodayResponse{Date: "2026-03-01", Actual: actual, Planned: &planned},
	}
}

func TestDemoHistoryFollowsThePlanScaledToTodaysProgress(t *testing.T) {
	curve := curveWith(40, 80,
		project.CurvePointResponse{Date: "2026-01-01", Planned: 0},
		project.CurvePointResponse{Date: "2026-01-15", Planned: 30},
		project.CurvePointResponse{Date: "2026-02-01", Planned: 60},
		project.CurvePointResponse{Date: "2026-03-01", Planned: 80},
		project.CurvePointResponse{Date: "2026-04-01", Planned: 100},
	)

	history := demoHistory(curve, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))

	want := []int{0, 15, 30}
	if len(history) != len(want) {
		t.Fatalf("got %d points, want %d before today", len(history), len(want))
	}
	for i, point := range history {
		if point.progress != want[i] {
			t.Fatalf("point %d = %d, want %d (half of the plan, because the project is at 40 of a planned 80)", i, point.progress, want[i])
		}
	}
}

func TestDemoHistoryNeverDropsOrPassesTodaysProgress(t *testing.T) {
	curve := curveWith(50, 40,
		project.CurvePointResponse{Date: "2026-01-01", Planned: 20},
		project.CurvePointResponse{Date: "2026-01-15", Planned: 10},
		project.CurvePointResponse{Date: "2026-02-01", Planned: 39},
	)

	history := demoHistory(curve, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))

	previous := 0
	for _, point := range history {
		if point.progress < previous || point.progress > 50 {
			t.Fatalf("history %+v drops or passes today's 50", history)
		}
		previous = point.progress
	}
}

func TestDemoHistoryIsEmptyForProjectsThatHaveNotStarted(t *testing.T) {
	curve := curveWith(0, 20, project.CurvePointResponse{Date: "2026-01-01", Planned: 10})
	if history := demoHistory(curve, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)); len(history) != 0 {
		t.Fatalf("got %+v, want no history while actual progress is 0", history)
	}

	undated := project.ProgressCurveResponse{Today: project.CurveTodayResponse{Actual: 30}}
	if history := demoHistory(undated, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)); len(history) != 0 {
		t.Fatalf("got %+v, want no history without a plan", history)
	}
}
