package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

func withID(phase ProjectPhase) ProjectPhase {
	phase.ID = uuid.New()
	return phase
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func datedPhase(progress int, start, end time.Time) ProjectPhase {
	return withID(ProjectPhase{Progress: progress, StartDate: &start, TargetEndDate: &end})
}

func TestBudgetWeightsLetCostlyPhasesCountMore(t *testing.T) {
	preparation := withID(ProjectPhase{Progress: 100})
	structure := withID(ProjectPhase{Progress: 0})
	phases := []ProjectPhase{preparation, structure}

	weights := weighPhases(phases, map[uuid.UUID]int64{preparation.ID: 500_000_000, structure.ID: 9_500_000_000})

	if weights.method != weightingBudget {
		t.Fatalf("method = %s, want budget", weights.method)
	}
	if got := weightedProgress(phases, weights); got != 5 {
		t.Fatalf("progress = %d, want 5 because preparation is only 5%% of the RAB", got)
	}
	if got := weightedProgress(phases, weighPhases(phases, nil)); got != 50 {
		t.Fatalf("equal weighting = %d, want the old average of 50", got)
	}
}

func TestPhasesWithoutLinkedBudgetCountForNothingOnceAnyPhaseHasBudget(t *testing.T) {
	linked := withID(ProjectPhase{Progress: 40})
	unlinked := withID(ProjectPhase{Progress: 100})

	weights := weighPhases([]ProjectPhase{linked, unlinked}, map[uuid.UUID]int64{linked.ID: 1_000})

	if weights.fraction[unlinked.ID] != 0 || weights.fraction[linked.ID] != 1 {
		t.Fatalf("fractions = %v", weights.fraction)
	}
	if got := weightedProgress([]ProjectPhase{linked, unlinked}, weights); got != 40 {
		t.Fatalf("progress = %d, want 40", got)
	}
}

func TestPlannedShareGrowsLinearlyFromStartToTargetDay(t *testing.T) {
	phase := datedPhase(0, date(2026, 1, 1), date(2026, 1, 10))
	cases := map[time.Time]float64{
		date(2025, 12, 31): 0,
		date(2026, 1, 1):   0.1,
		date(2026, 1, 5):   0.5,
		date(2026, 1, 10):  1,
		date(2026, 3, 1):   1,
	}
	for on, want := range cases {
		got, dated := plannedShare(phase, on)
		if !dated || got != want {
			t.Fatalf("share on %s = %v (dated %v), want %v", on.Format(time.DateOnly), got, dated, want)
		}
	}

	if _, dated := plannedShare(withID(ProjectPhase{}), date(2026, 1, 5)); dated {
		t.Fatal("a phase without dates must not contribute to the plan")
	}
}

func TestPlannedProgressAddsWeightedPhaseShares(t *testing.T) {
	first := datedPhase(0, date(2026, 1, 1), date(2026, 1, 10))
	second := datedPhase(0, date(2026, 1, 11), date(2026, 1, 20))
	phases := []ProjectPhase{first, second}
	weights := weighPhases(phases, map[uuid.UUID]int64{first.ID: 1, second.ID: 3})

	if got := plannedProgress(phases, weights, date(2026, 1, 10)); got != 25 {
		t.Fatalf("plan at end of first phase = %v, want 25", got)
	}
	if got := plannedProgress(phases, weights, date(2026, 1, 15)); got != 62.5 {
		t.Fatalf("plan halfway through second phase = %v, want 62.5", got)
	}
	if got := plannedProgress(phases, weights, date(2026, 2, 1)); got != 100 {
		t.Fatalf("plan after the last phase = %v, want 100", got)
	}
}

func TestProgressCurveHasWeeklyPlanAndActualOnlyUpToToday(t *testing.T) {
	phase := datedPhase(40, date(2026, 1, 1), date(2026, 2, 28))
	phases := []ProjectPhase{phase}
	weights := weighPhases(phases, nil)
	snapshots := []ProgressSnapshot{
		{RecordedOn: date(2026, 1, 20), Progress: 25},
		{RecordedOn: date(2026, 1, 9), Progress: 10},
	}
	today := date(2026, 1, 25)

	points := progressCurve(Project{}, phases, weights, snapshots, today)

	if !points[0].day.Equal(date(2026, 1, 1)) || !points[len(points)-1].day.Equal(date(2026, 2, 28)) {
		t.Fatalf("curve runs %s to %s", points[0].day.Format(time.DateOnly), points[len(points)-1].day.Format(time.DateOnly))
	}
	byDay := map[string]curvePoint{}
	for _, point := range points {
		byDay[point.day.Format(time.DateOnly)] = point
	}

	if byDay["2026-01-01"].actual != nil {
		t.Fatal("actual progress before the first snapshot must be unknown")
	}
	if got := byDay["2026-01-15"].actual; got == nil || *got != 10 {
		t.Fatalf("actual on 15 Jan = %v, want the 9 Jan snapshot of 10", got)
	}
	if got := byDay["2026-01-25"].actual; got == nil || *got != 25 {
		t.Fatalf("today must be on the curve with the latest snapshot, got %v", got)
	}
	if byDay["2026-02-05"].actual != nil {
		t.Fatal("future points must not have actual progress")
	}
	if byDay["2026-02-28"].planned != 100 {
		t.Fatalf("plan on the target day = %v, want 100", byDay["2026-02-28"].planned)
	}
}

func TestProgressCurveStretchesItsStepForLongProjects(t *testing.T) {
	phase := datedPhase(0, date(2026, 1, 1), date(2031, 1, 1))
	points := progressCurve(Project{}, []ProjectPhase{phase}, weighPhases([]ProjectPhase{phase}, nil), nil, date(2026, 6, 1))

	if len(points) > maxCurvePoints+2 {
		t.Fatalf("a five year project produced %d points, limit is about %d", len(points), maxCurvePoints)
	}
}

func TestProgressCurveIsEmptyWithoutAnyDates(t *testing.T) {
	phases := []ProjectPhase{withID(ProjectPhase{Progress: 30})}
	if points := progressCurve(Project{}, phases, weighPhases(phases, nil), nil, date(2026, 1, 1)); len(points) != 0 {
		t.Fatalf("got %d points, want none", len(points))
	}
}

func TestProgressCurveResponseReportsDeviationFromPlan(t *testing.T) {
	phase := datedPhase(30, date(2026, 1, 1), date(2026, 1, 10))
	repository := &fakeRepository{
		project: Project{Progress: 30},
		phases:  []ProjectPhase{phase},
	}
	evaluator := NewService(repository).(*service)
	evaluator.now = func() time.Time { return time.Date(2026, 1, 5, 14, 0, 0, 0, time.UTC) }

	curve, err := evaluator.ProgressCurve(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if curve.Weighting != weightingEqual || len(curve.Phases) != 1 || curve.Phases[0].WeightPercent != 100 {
		t.Fatalf("weighting %s, phases %+v", curve.Weighting, curve.Phases)
	}
	if curve.Today.Date != "2026-01-05" || curve.Today.Planned == nil || *curve.Today.Planned != 50 {
		t.Fatalf("today = %+v, want plan 50 on 2026-01-05", curve.Today)
	}
	if curve.Today.Deviation == nil || *curve.Today.Deviation != -20 {
		t.Fatalf("deviation = %v, want -20 (30 actual vs 50 planned)", curve.Today.Deviation)
	}
}

func TestBudgetItemCannotPointAtAPhaseOutsideTheProject(t *testing.T) {
	repository := &fakeRepository{project: Project{Status: statusOngoing}, missingPhase: true}
	foreignPhase := uuid.New()

	_, err := NewService(repository).CreateBudgetItem(context.Background(), uuid.New(), BudgetItemRequest{
		PhaseID: &foreignPhase, Code: "1.1", Description: "Galian", Volume: 10, UnitOfMeasureID: uuid.New(), UnitPrice: 1000,
	})

	var appError *apperror.Error
	if !errors.As(err, &appError) || appError.Code != "budget_item_phase_unknown" || appError.Status != 422 {
		t.Fatalf("err = %v, want 422 budget_item_phase_unknown", err)
	}
	if repository.createdItem != nil {
		t.Fatal("the item was created anyway")
	}
}

func TestCreatingABudgetItemRecalculatesWeightedProgress(t *testing.T) {
	done := withID(ProjectPhase{Progress: 100})
	open := withID(ProjectPhase{Progress: 0})
	repository := &fakeRepository{
		project:       Project{Status: statusOngoing},
		phases:        []ProjectPhase{done, open},
		budgetByPhase: map[uuid.UUID]int64{done.ID: 1_000, open.ID: 3_000},
	}
	evaluator := NewService(repository).(*service)
	evaluator.now = func() time.Time { return time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC) }

	item, err := evaluator.CreateBudgetItem(context.Background(), uuid.New(), BudgetItemRequest{
		PhaseID: &open.ID, Code: "2.1", Description: "Struktur", Volume: 3, UnitOfMeasureID: uuid.New(), UnitPrice: 1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.PhaseID == nil || *item.PhaseID != open.ID {
		t.Fatalf("item phase = %v, want %s", item.PhaseID, open.ID)
	}
	if repository.storedProgress == nil || *repository.storedProgress != 25 {
		t.Fatalf("stored progress = %v, want 25", repository.storedProgress)
	}
	if len(repository.snapshots) != 1 || !repository.snapshots[0].RecordedOn.Equal(date(2026, 3, 2)) {
		t.Fatalf("snapshots = %+v, want one on 2026-03-02", repository.snapshots)
	}
}
