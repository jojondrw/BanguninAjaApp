package project

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	weightingBudget = "budget"
	weightingEqual  = "equal"

	curveStepDays  = 7
	maxCurvePoints = 105
	day            = 24 * time.Hour
)

type phaseWeights struct {
	method   string
	fraction map[uuid.UUID]float64
	budget   map[uuid.UUID]int64
}

type curvePoint struct {
	day     time.Time
	planned float64
	actual  *float64
}

func weighPhases(phases []ProjectPhase, budgetByPhase map[uuid.UUID]int64) phaseWeights {
	weights := phaseWeights{fraction: make(map[uuid.UUID]float64, len(phases)), budget: map[uuid.UUID]int64{}}

	var total int64
	for _, phase := range phases {
		if amount := budgetByPhase[phase.ID]; amount > 0 {
			weights.budget[phase.ID] = amount
			total += amount
		}
	}

	if total == 0 {
		weights.method = weightingEqual
		for _, phase := range phases {
			weights.fraction[phase.ID] = 1 / float64(len(phases))
		}
		return weights
	}

	weights.method = weightingBudget
	for _, phase := range phases {
		weights.fraction[phase.ID] = float64(weights.budget[phase.ID]) / float64(total)
	}
	return weights
}

func weightedProgress(phases []ProjectPhase, weights phaseWeights) int {
	sum := 0.0
	for _, phase := range phases {
		sum += weights.fraction[phase.ID] * float64(phase.Progress)
	}
	return int(math.Round(sum))
}

func dateOnly(value time.Time) time.Time {
	year, month, date := value.Date()
	return time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(math.Round(dateOnly(to).Sub(dateOnly(from)).Hours() / 24))
}

func plannedShare(phase ProjectPhase, on time.Time) (float64, bool) {
	if phase.StartDate == nil || phase.TargetEndDate == nil {
		return 0, false
	}
	duration := daysBetween(*phase.StartDate, *phase.TargetEndDate) + 1
	elapsed := daysBetween(*phase.StartDate, on) + 1
	switch {
	case elapsed <= 0:
		return 0, true
	case elapsed >= duration:
		return 1, true
	default:
		return float64(elapsed) / float64(duration), true
	}
}

func plannedProgress(phases []ProjectPhase, weights phaseWeights, on time.Time) float64 {
	sum := 0.0
	for _, phase := range phases {
		if share, dated := plannedShare(phase, on); dated {
			sum += weights.fraction[phase.ID] * share * 100
		}
	}
	return roundTenth(sum)
}

func undatedPhases(phases []ProjectPhase) int {
	count := 0
	for _, phase := range phases {
		if phase.StartDate == nil || phase.TargetEndDate == nil {
			count++
		}
	}
	return count
}

func curveRange(project Project, phases []ProjectPhase) (time.Time, time.Time, bool) {
	var start, end time.Time
	consider := func(from, to *time.Time) {
		if from != nil && (start.IsZero() || dateOnly(*from).Before(start)) {
			start = dateOnly(*from)
		}
		if to != nil && (end.IsZero() || dateOnly(*to).After(end)) {
			end = dateOnly(*to)
		}
	}
	consider(project.StartDate, project.TargetEndDate)
	for _, phase := range phases {
		consider(phase.StartDate, phase.TargetEndDate)
	}
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func curveStep(start, end time.Time) int {
	span := daysBetween(start, end)
	step := curveStepDays
	if span/step+1 > maxCurvePoints {
		step = int(math.Ceil(float64(span) / float64(maxCurvePoints-1)))
	}
	return step
}

func progressCurve(project Project, phases []ProjectPhase, weights phaseWeights, snapshots []ProgressSnapshot, today time.Time) []curvePoint {
	start, end, ok := curveRange(project, phases)
	if !ok {
		return []curvePoint{}
	}
	today = dateOnly(today)

	days := []time.Time{}
	step := curveStep(start, end)
	for current := start; !current.After(end); current = current.Add(time.Duration(step) * day) {
		days = append(days, current)
	}
	if !days[len(days)-1].Equal(end) {
		days = append(days, end)
	}
	if !today.Before(start) && !today.After(end) && !containsDay(days, today) {
		days = append(days, today)
		sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
	}

	ordered := append([]ProgressSnapshot(nil), snapshots...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].RecordedOn.Before(ordered[j].RecordedOn) })

	points := make([]curvePoint, 0, len(days))
	for _, current := range days {
		points = append(points, curvePoint{
			day:     current,
			planned: plannedProgress(phases, weights, current),
			actual:  actualOn(ordered, current, today),
		})
	}
	return points
}

func actualOn(ordered []ProgressSnapshot, on, today time.Time) *float64 {
	if on.After(today) {
		return nil
	}
	var found *float64
	for _, snapshot := range ordered {
		if dateOnly(snapshot.RecordedOn).After(on) {
			break
		}
		value := float64(snapshot.Progress)
		found = &value
	}
	return found
}

func containsDay(days []time.Time, target time.Time) bool {
	for _, current := range days {
		if current.Equal(target) {
			return true
		}
	}
	return false
}

func roundTenth(value float64) float64 {
	return math.Round(value*10) / 10
}

func newProgressCurveResponse(project Project, phases []ProjectPhase, weights phaseWeights, points []curvePoint, today time.Time) ProgressCurveResponse {
	response := ProgressCurveResponse{
		Weighting:     weights.method,
		UndatedPhases: undatedPhases(phases),
		Phases:        make([]PhaseWeightResponse, 0, len(phases)),
		Points:        make([]CurvePointResponse, 0, len(points)),
		Today:         CurveTodayResponse{Date: today.Format(time.DateOnly), Actual: project.Progress},
	}
	for _, phase := range phases {
		response.Phases = append(response.Phases, PhaseWeightResponse{
			ID:            phase.ID,
			Name:          phase.Name,
			WeightPercent: roundTenth(weights.fraction[phase.ID] * 100),
			Budget:        weights.budget[phase.ID],
			Progress:      phase.Progress,
			Dated:         phase.StartDate != nil && phase.TargetEndDate != nil,
		})
	}
	for _, point := range points {
		response.Points = append(response.Points, CurvePointResponse{
			Date:    point.day.Format(time.DateOnly),
			Planned: point.planned,
			Actual:  point.actual,
		})
	}
	if len(points) > 0 {
		planned := plannedProgress(phases, weights, today)
		deviation := roundTenth(float64(project.Progress) - planned)
		response.Today.Planned = &planned
		response.Today.Deviation = &deviation
	}
	return response
}
