package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	project        Project
	phases         []ProjectPhase
	savedProject   *Project
	storedProgress *int
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) LockProject(context.Context, uuid.UUID) (Project, error) {
	return f.project, nil
}

func (f *fakeRepository) FindProject(context.Context, uuid.UUID) (Project, error) {
	return f.project, nil
}

func (f *fakeRepository) ListPhases(context.Context, uuid.UUID) ([]ProjectPhase, error) {
	return f.phases, nil
}

func (f *fakeRepository) SaveProject(_ context.Context, project *Project) error {
	f.savedProject = project
	return nil
}

func (f *fakeRepository) CreatePhase(_ context.Context, phase *ProjectPhase) error {
	f.phases = append(f.phases, *phase)
	return nil
}

func (f *fakeRepository) UpdateProjectProgress(_ context.Context, _ uuid.UUID, progress int) error {
	f.storedProgress = &progress
	return nil
}

func TestCanTransitionProject(t *testing.T) {
	cases := []struct {
		current, next string
		allowed       bool
	}{
		{statusPlanning, statusOngoing, true},
		{statusPlanning, statusCompleted, false},
		{statusOngoing, statusOnHold, true},
		{statusOnHold, statusOngoing, true},
		{statusOnHold, statusCompleted, false},
		{statusCompleted, statusOngoing, false},
		{statusCancelled, statusPlanning, false},
	}
	for _, tc := range cases {
		if got := canTransitionProject(tc.current, tc.next); got != tc.allowed {
			t.Errorf("%s -> %s: got %v, want %v", tc.current, tc.next, got, tc.allowed)
		}
	}
}

func TestAverageProgress(t *testing.T) {
	phases := []ProjectPhase{{Progress: 100}, {Progress: 50}, {Progress: 0}}
	if got := averageProgress(phases); got != 50 {
		t.Fatalf("got %d, want 50", got)
	}
	if got := averageProgress(nil); got != 0 {
		t.Fatalf("got %d for no phases, want 0", got)
	}
	if got := averageProgress([]ProjectPhase{{Progress: 33}, {Progress: 34}}); got != 34 {
		t.Fatalf("got %d, want rounded 34", got)
	}
}

func TestPhaseStatus(t *testing.T) {
	cases := map[int]string{0: phaseNotStarted, 1: phaseInProgress, 99: phaseInProgress, 100: phaseCompleted}
	for progress, want := range cases {
		if got := phaseStatus(progress); got != want {
			t.Errorf("progress %d: got %s, want %s", progress, got, want)
		}
	}
}

func TestBudgetTotalRoundsToWholeRupiah(t *testing.T) {
	if got := budgetTotal(roundVolume(12.345), 15000); got != 185250 {
		t.Fatalf("got %d, want 185250", got)
	}
}

func TestDateRangeValid(t *testing.T) {
	start := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	before := start.AddDate(0, 0, -1)
	if dateRangeValid(&start, &before) {
		t.Fatal("end before start must be rejected")
	}
	if !dateRangeValid(&start, &start) || !dateRangeValid(nil, &before) {
		t.Fatal("same day and open ranges must be accepted")
	}
}

func TestValidatePermitRequiresIssuedDate(t *testing.T) {
	err := validatePermit(PermitRequest{Status: permitIssued})
	if !errors.Is(err, errPermitIssuedDateRequired) {
		t.Fatalf("got %v, want errPermitIssuedDateRequired", err)
	}
}

func TestCompletingProjectWithUnfinishedPhaseIsRejected(t *testing.T) {
	repository := &fakeRepository{
		project: Project{Status: statusOngoing, Progress: 60},
		phases:  []ProjectPhase{{Progress: 100}, {Progress: 20}},
	}
	service := NewService(repository)

	_, err := service.UpdateProjectStatus(context.Background(), uuid.New(), ProjectStatusRequest{Status: statusCompleted})
	if !errors.Is(err, errProjectUnfinished) {
		t.Fatalf("got %v, want errProjectUnfinished", err)
	}
	if repository.savedProject != nil {
		t.Fatal("project must not be saved")
	}
}

func TestCompletingProjectSetsFullProgress(t *testing.T) {
	repository := &fakeRepository{
		project: Project{Status: statusOngoing, Progress: 90},
		phases:  []ProjectPhase{{Progress: 100}},
	}
	service := NewService(repository)

	response, err := service.UpdateProjectStatus(context.Background(), uuid.New(), ProjectStatusRequest{Status: statusCompleted})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Progress != fullProgress || response.Status != statusCompleted {
		t.Fatalf("got progress %d status %s", response.Progress, response.Status)
	}
}

func TestCreatePhaseRefreshesProjectProgress(t *testing.T) {
	repository := &fakeRepository{
		project: Project{Status: statusOngoing},
		phases:  []ProjectPhase{{Progress: 100}},
	}
	service := NewService(repository)

	phase, err := service.CreatePhase(context.Background(), uuid.New(), PhaseRequest{Name: "Struktur", SortOrder: 2, Progress: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if phase.Status != phaseInProgress {
		t.Fatalf("got phase status %s", phase.Status)
	}
	if repository.storedProgress == nil || *repository.storedProgress != 75 {
		t.Fatalf("got stored progress %v, want 75", repository.storedProgress)
	}
}

func TestClosedProjectRejectsNewPhase(t *testing.T) {
	repository := &fakeRepository{project: Project{Status: statusCancelled}}
	service := NewService(repository)

	_, err := service.CreatePhase(context.Background(), uuid.New(), PhaseRequest{Name: "Finishing", SortOrder: 1})
	if !errors.Is(err, errProjectClosed) {
		t.Fatalf("got %v, want errProjectClosed", err)
	}
}
