package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const contractValue = 30_000_000_000

type termRepository struct {
	Repository
	project   Project
	contract  *ConstructionContract
	vendor    ContractVendor
	terms     []PaymentTerm
	paid      map[uuid.UUID]int64
	recorded  []PayableEntry
	createdAt time.Time
}

func newTermRepository(status string, progress int) *termRepository {
	return &termRepository{
		project:   Project{Code: "DEMO-RSIA", Status: status, Progress: progress},
		contract:  &ConstructionContract{VendorID: uuid.New(), Number: "SPK-01", Value: contractValue, RetentionPercent: 5},
		vendor:    ContractVendor{Name: "PT Karya Beton", PaymentTermDays: 14, Active: true},
		paid:      map[uuid.UUID]int64{},
		createdAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func (f *termRepository) addTerm(required int, percent float64) *PaymentTerm {
	f.createdAt = f.createdAt.Add(time.Hour)
	term := PaymentTerm{ProjectID: f.project.ID, Kind: termProgress, RequiredProgress: required, Percent: percent}
	term.ID = uuid.New()
	term.CreatedAt = f.createdAt
	f.terms = append(f.terms, term)
	return &f.terms[len(f.terms)-1]
}

func (f *termRepository) claim(term *PaymentTerm) {
	gross, retention := termAmounts(contractValue, term.Percent, 5)
	payable := uuid.New()
	term.GrossAmount, term.RetentionAmount, term.PayableID = gross, retention, &payable
}

func (f *termRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *termRepository) LockProject(context.Context, uuid.UUID) (Project, error) {
	return f.project, nil
}

func (f *termRepository) FindProject(context.Context, uuid.UUID) (Project, error) {
	return f.project, nil
}

func (f *termRepository) FindContract(context.Context, uuid.UUID) (ConstructionContract, error) {
	if f.contract == nil {
		return ConstructionContract{}, database.ErrNotFound
	}
	return *f.contract, nil
}

func (f *termRepository) CreateContract(_ context.Context, contract *ConstructionContract) error {
	f.contract = contract
	return nil
}

func (f *termRepository) SaveContract(_ context.Context, contract *ConstructionContract) error {
	f.contract = contract
	return nil
}

func (f *termRepository) FindContractVendor(context.Context, uuid.UUID) (ContractVendor, error) {
	return f.vendor, nil
}

func (f *termRepository) ListPaymentTerms(context.Context, uuid.UUID) ([]PaymentTermRow, error) {
	rows := make([]PaymentTermRow, 0, len(f.terms))
	for _, term := range f.terms {
		row := PaymentTermRow{
			ID:               term.ID,
			Kind:             term.Kind,
			RequiredProgress: term.RequiredProgress,
			Percent:          term.Percent,
			GrossAmount:      term.GrossAmount,
			RetentionAmount:  term.RetentionAmount,
			PayableID:        term.PayableID,
			ClaimedOn:        term.ClaimedOn,
			ClaimedProgress:  term.ClaimedProgress,
			CreatedAt:        term.CreatedAt,
		}
		if term.PayableID != nil {
			row.PayableAmount = term.GrossAmount - term.RetentionAmount
			row.PayablePaid = f.paid[term.ID]
		}
		rows = append([]PaymentTermRow{row}, rows...)
	}
	return rows, nil
}

func (f *termRepository) FindPaymentTerm(_ context.Context, _, id uuid.UUID) (PaymentTerm, error) {
	for _, term := range f.terms {
		if term.ID == id {
			return term, nil
		}
	}
	return PaymentTerm{}, database.ErrNotFound
}

func (f *termRepository) CreatePaymentTerm(_ context.Context, term *PaymentTerm) error {
	f.createdAt = f.createdAt.Add(time.Hour)
	term.ID = uuid.New()
	term.CreatedAt = f.createdAt
	f.terms = append(f.terms, *term)
	return nil
}

func (f *termRepository) SavePaymentTerm(_ context.Context, term *PaymentTerm) error {
	for i := range f.terms {
		if f.terms[i].ID == term.ID {
			f.terms[i] = *term
			return nil
		}
	}
	return database.ErrNotFound
}

func (f *termRepository) DeletePaymentTerm(_ context.Context, _, id uuid.UUID) error {
	for i := range f.terms {
		if f.terms[i].ID == id {
			f.terms = append(f.terms[:i], f.terms[i+1:]...)
			return nil
		}
	}
	return database.ErrNotFound
}

func (f *termRepository) Payables() PayableLedger {
	return f
}

func (f *termRepository) Record(_ context.Context, entry PayableEntry) (uuid.UUID, error) {
	f.recorded = append(f.recorded, entry)
	return uuid.New(), nil
}

func termService(repository *termRepository) *service {
	evaluator := NewService(repository).(*service)
	evaluator.now = func() time.Time { return time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC) }
	return evaluator
}

func termRequest(required int, percent float64) PaymentTermRequest {
	return PaymentTermRequest{RequiredProgress: &required, Percent: percent}
}

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	var appError *apperror.Error
	if !errors.As(err, &appError) || appError.Code != code {
		t.Fatalf("err = %v, want %s", err, code)
	}
}

func TestTermAmountsTakeRetentionFromTheGrossValue(t *testing.T) {
	gross, retention := termAmounts(contractValue, 20, 5)

	if gross != 6_000_000_000 || retention != 300_000_000 {
		t.Fatalf("gross %d retention %d, want 6.000.000.000 and 300.000.000", gross, retention)
	}
}

func TestClaimRecordsThePayableNetOfRetentionDueAfterVendorTerms(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	first := repository.addTerm(20, 20)
	repository.claim(first)
	second := repository.addTerm(40, 25)
	actor := uuid.New()

	schedule, err := termService(repository).ClaimPaymentTerm(context.Background(), uuid.New(), second.ID, actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repository.recorded) != 1 {
		t.Fatalf("recorded %d payables, want 1", len(repository.recorded))
	}
	entry := repository.recorded[0]
	if entry.Amount != 7_125_000_000 {
		t.Fatalf("payable amount = %d, want 7.500.000.000 minus 5%% retention", entry.Amount)
	}
	if !entry.DueDate.Equal(date(2026, 10, 31)) {
		t.Fatalf("due date = %s, want 14 vendor days after today", entry.DueDate)
	}
	if entry.Reference != "Termin 2 DEMO-RSIA" || entry.VendorID != repository.contract.VendorID {
		t.Fatalf("payable = %+v, want reference Termin 2 DEMO-RSIA for the contract vendor", entry)
	}

	claimed := repository.terms[1]
	if claimed.ClaimedProgress == nil || *claimed.ClaimedProgress != 46 || claimed.ClaimedBy == nil || *claimed.ClaimedBy != actor {
		t.Fatalf("claim audit not stored: %+v", claimed)
	}
	if schedule.Terms[1].State != termClaimed || schedule.RetentionHeld != 300_000_000+375_000_000 {
		t.Fatalf("schedule = %+v", schedule)
	}
}

func TestClaimIsRefusedBeforeTheWeightedProgressIsReached(t *testing.T) {
	repository := newTermRepository(statusOngoing, 38)
	term := repository.addTerm(40, 25)

	_, err := termService(repository).ClaimPaymentTerm(context.Background(), uuid.New(), term.ID, uuid.New())

	requireCode(t, err, "payment_term_progress_not_reached")
	if len(repository.recorded) != 0 {
		t.Fatal("a payable was recorded for unfinished work")
	}
}

func TestClaimMustFollowTheTermOrder(t *testing.T) {
	repository := newTermRepository(statusOngoing, 80)
	repository.addTerm(20, 20)
	later := repository.addTerm(50, 30)

	_, err := termService(repository).ClaimPaymentTerm(context.Background(), uuid.New(), later.ID, uuid.New())

	if !errors.Is(err, errTermOutOfOrder) {
		t.Fatalf("err = %v, want errTermOutOfOrder", err)
	}
}

func TestClaimIsRefusedTwiceAndOnCancelledProjects(t *testing.T) {
	repository := newTermRepository(statusOngoing, 80)
	term := repository.addTerm(20, 20)
	repository.claim(term)

	if _, err := termService(repository).ClaimPaymentTerm(context.Background(), uuid.New(), term.ID, uuid.New()); !errors.Is(err, errTermClaimed) {
		t.Fatalf("second claim err = %v, want errTermClaimed", err)
	}

	cancelled := newTermRepository(statusCancelled, 80)
	open := cancelled.addTerm(20, 20)
	if _, err := termService(cancelled).ClaimPaymentTerm(context.Background(), uuid.New(), open.ID, uuid.New()); !errors.Is(err, errTermProjectCancelled) {
		t.Fatalf("cancelled project err = %v, want errTermProjectCancelled", err)
	}
}

func TestNewTermsCannotExceedTheContractOrGoBehindAClaim(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	first := repository.addTerm(30, 30)
	repository.claim(first)
	repository.addTerm(60, 60)

	cases := []struct {
		name    string
		request PaymentTermRequest
		want    error
	}{
		{"over 100 percent", termRequest(90, 10.01), errTermPercentExceeded},
		{"before a claimed term", termRequest(20, 5), errTermBeforeClaimed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := termService(repository).CreatePaymentTerm(context.Background(), uuid.New(), tc.request); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}

	schedule, err := termService(repository).CreatePaymentTerm(context.Background(), uuid.New(), termRequest(100, 10))
	if err != nil {
		t.Fatalf("exactly 100 percent refused: %v", err)
	}
	if schedule.PlannedPercent != 100 || len(schedule.Terms) != 3 {
		t.Fatalf("schedule = %+v, want three terms covering 100%%", schedule)
	}
}

func TestTermsNeedAContractAndAnOpenProject(t *testing.T) {
	repository := newTermRepository(statusOngoing, 0)
	repository.contract = nil
	if _, err := termService(repository).CreatePaymentTerm(context.Background(), uuid.New(), termRequest(10, 10)); !errors.Is(err, errContractNotFound) {
		t.Fatalf("err = %v, want errContractNotFound", err)
	}

	completed := newTermRepository(statusCompleted, 100)
	if _, err := termService(completed).CreatePaymentTerm(context.Background(), uuid.New(), termRequest(10, 10)); !errors.Is(err, errProjectClosed) {
		t.Fatalf("err = %v, want errProjectClosed", err)
	}
}

func TestClaimedTermsCannotBeDeleted(t *testing.T) {
	repository := newTermRepository(statusOngoing, 50)
	claimed := repository.addTerm(20, 20)
	repository.claim(claimed)
	planned := repository.addTerm(60, 30)

	if _, err := termService(repository).DeletePaymentTerm(context.Background(), uuid.New(), claimed.ID); !errors.Is(err, errTermClaimed) {
		t.Fatalf("err = %v, want errTermClaimed", err)
	}
	schedule, err := termService(repository).DeletePaymentTerm(context.Background(), uuid.New(), planned.ID)
	if err != nil || len(schedule.Terms) != 1 {
		t.Fatalf("delete planned: err %v, terms %d", err, len(schedule.Terms))
	}
}

func TestContractIsLockedOnceATermIsClaimedExceptItsNumber(t *testing.T) {
	repository := newTermRepository(statusOngoing, 50)
	repository.claim(repository.addTerm(20, 20))
	retention := 5.0
	request := ContractRequest{VendorID: repository.contract.VendorID, Number: " SPK-01/ADD-1 ", Value: contractValue, RetentionPercent: &retention}

	schedule, err := termService(repository).SaveContract(context.Background(), uuid.New(), request)
	if err != nil {
		t.Fatalf("number change refused: %v", err)
	}
	if schedule.Contract.Number != "SPK-01/ADD-1" || !schedule.ContractLocked {
		t.Fatalf("contract = %+v locked %v", schedule.Contract, schedule.ContractLocked)
	}

	request.Value = contractValue + 1
	if _, err := termService(repository).SaveContract(context.Background(), uuid.New(), request); !errors.Is(err, errContractLocked) {
		t.Fatalf("value change err = %v, want errContractLocked", err)
	}
}

func TestNewContractRejectsAnInactiveVendor(t *testing.T) {
	repository := newTermRepository(statusOngoing, 0)
	repository.contract = nil
	repository.vendor.Active = false
	retention := 5.0

	_, err := termService(repository).SaveContract(context.Background(), uuid.New(),
		ContractRequest{VendorID: uuid.New(), Number: "SPK-02", Value: 1_000_000, RetentionPercent: &retention})

	if !errors.Is(err, errContractVendorInactive) {
		t.Fatalf("err = %v, want errContractVendorInactive", err)
	}
}

func TestScheduleMarksOnlyTheNextReachableTermReady(t *testing.T) {
	repository := newTermRepository(statusOngoing, 55)
	paid := repository.addTerm(0, 10)
	repository.claim(paid)
	repository.paid[paid.ID] = paid.GrossAmount - paid.RetentionAmount
	repository.addTerm(50, 30)
	repository.addTerm(50, 30)
	repository.addTerm(90, 30)

	schedule, err := termService(repository).PaymentSchedule(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	states := []string{}
	for _, term := range schedule.Terms {
		states = append(states, term.State)
	}
	want := []string{termPaid, termReady, termPlanned, termPlanned}
	for i := range want {
		if states[i] != want[i] {
			t.Fatalf("states = %v, want %v", states, want)
		}
	}
	if schedule.Terms[1].Label != "Termin 2" || schedule.Terms[1].NetAmount != 8_550_000_000 {
		t.Fatalf("planned term = %+v, want Termin 2 worth 9 M minus retention", schedule.Terms[1])
	}
	if schedule.CanReleaseRetention {
		t.Fatal("retention releasable on an ongoing project")
	}
}

func TestRetentionIsReleasedOnceAfterCompletion(t *testing.T) {
	ongoing := newTermRepository(statusOngoing, 100)
	ongoing.claim(ongoing.addTerm(50, 100))
	if _, err := termService(ongoing).ReleaseRetention(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, errRetentionNotDue) {
		t.Fatalf("ongoing err = %v, want errRetentionNotDue", err)
	}

	repository := newTermRepository(statusCompleted, 100)
	repository.claim(repository.addTerm(50, 60))
	repository.claim(repository.addTerm(100, 40))

	schedule, err := termService(repository).ReleaseRetention(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.recorded) != 1 || repository.recorded[0].Amount != 1_500_000_000 || repository.recorded[0].Reference != "Retensi DEMO-RSIA" {
		t.Fatalf("recorded = %+v, want one payable of 5%% of 30 M", repository.recorded)
	}
	if !schedule.RetentionReleased || schedule.CanReleaseRetention || schedule.Terms[2].Label != "Pelepasan retensi" {
		t.Fatalf("schedule = %+v", schedule)
	}

	if _, err := termService(repository).ReleaseRetention(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, errRetentionReleased) {
		t.Fatalf("second release err = %v, want errRetentionReleased", err)
	}
}

func TestRetentionWaitsForEveryProgressTerm(t *testing.T) {
	repository := newTermRepository(statusCompleted, 100)
	repository.claim(repository.addTerm(50, 60))
	repository.addTerm(100, 40)

	if _, err := termService(repository).ReleaseRetention(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, errRetentionTermsOpen) {
		t.Fatalf("err = %v, want errRetentionTermsOpen", err)
	}
}

func TestRetentionReleaseReusesARowWhosePayableWasDeleted(t *testing.T) {
	repository := newTermRepository(statusCompleted, 100)
	repository.claim(repository.addTerm(100, 100))
	orphan := repository.addTerm(100, 0)
	orphan.Kind = termRetention

	if _, err := termService(repository).ReleaseRetention(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.terms) != 2 || repository.terms[1].PayableID == nil || repository.terms[1].GrossAmount != 1_500_000_000 {
		t.Fatalf("terms = %+v, want the orphan retention row reused", repository.terms)
	}
}
