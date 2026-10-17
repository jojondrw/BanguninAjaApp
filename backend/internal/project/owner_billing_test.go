package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const ownerContractValue = 20_000_000_000

type ownerRepository struct {
	Repository
	project   Project
	contract  *OwnerContract
	customers map[uuid.UUID]string
	terms     []OwnerTerm
	paid      map[uuid.UUID]int64
	billed    []ReceivableEntry
	createdAt time.Time
}

func newOwnerRepository(status string, progress int) *ownerRepository {
	customerID := uuid.New()
	project := Project{Code: "DEMO-RSIA", Status: status, Progress: progress}
	project.ID = uuid.New()
	return &ownerRepository{
		project:   project,
		contract:  &OwnerContract{ProjectID: project.ID, CustomerID: customerID, Number: "KTR-RSIA-01", Value: ownerContractValue, RetentionPercent: 5, PaymentTermDays: 30},
		customers: map[uuid.UUID]string{customerID: "Yayasan RSIA Cirebon"},
		paid:      map[uuid.UUID]int64{},
		createdAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func (f *ownerRepository) addTerm(required int, percent float64) OwnerTerm {
	f.createdAt = f.createdAt.Add(time.Hour)
	term := OwnerTerm{ProjectID: f.project.ID, Kind: termProgress, RequiredProgress: required, Percent: percent}
	term.ID = uuid.New()
	term.CreatedAt = f.createdAt
	f.terms = append(f.terms, term)
	return term
}

func (f *ownerRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *ownerRepository) LockProject(context.Context, uuid.UUID) (Project, error) {
	return f.project, nil
}

func (f *ownerRepository) FindProject(context.Context, uuid.UUID) (Project, error) {
	return f.project, nil
}

func (f *ownerRepository) FindOwnerContract(context.Context, uuid.UUID) (OwnerContract, error) {
	if f.contract == nil {
		return OwnerContract{}, database.ErrNotFound
	}
	return *f.contract, nil
}

func (f *ownerRepository) CreateOwnerContract(_ context.Context, contract *OwnerContract) error {
	f.contract = contract
	return nil
}

func (f *ownerRepository) SaveOwnerContract(_ context.Context, contract *OwnerContract) error {
	f.contract = contract
	return nil
}

func (f *ownerRepository) FindOwnerCustomerName(_ context.Context, id uuid.UUID) (string, error) {
	name, found := f.customers[id]
	if !found {
		return "", database.ErrNotFound
	}
	return name, nil
}

func (f *ownerRepository) ListOwnerTerms(context.Context, uuid.UUID) ([]PaymentTermRow, error) {
	rows := make([]PaymentTermRow, 0, len(f.terms))
	for _, term := range f.terms {
		row := PaymentTermRow{
			ID: term.ID, Kind: term.Kind, RequiredProgress: term.RequiredProgress, Percent: term.Percent,
			GrossAmount: term.GrossAmount, RetentionAmount: term.RetentionAmount, PayableID: term.ReceivableID,
			ClaimedOn: term.ClaimedOn, ClaimedProgress: term.ClaimedProgress, CreatedAt: term.CreatedAt,
		}
		if term.ReceivableID != nil {
			row.PayableAmount = term.GrossAmount - term.RetentionAmount
			row.PayablePaid = f.paid[term.ID]
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (f *ownerRepository) FindOwnerTerm(_ context.Context, _, id uuid.UUID) (OwnerTerm, error) {
	for _, term := range f.terms {
		if term.ID == id {
			return term, nil
		}
	}
	return OwnerTerm{}, database.ErrNotFound
}

func (f *ownerRepository) CreateOwnerTerm(_ context.Context, term *OwnerTerm) error {
	f.createdAt = f.createdAt.Add(time.Hour)
	term.ID = uuid.New()
	term.CreatedAt = f.createdAt
	f.terms = append(f.terms, *term)
	return nil
}

func (f *ownerRepository) SaveOwnerTerm(_ context.Context, term *OwnerTerm) error {
	for i := range f.terms {
		if f.terms[i].ID == term.ID {
			f.terms[i] = *term
			return nil
		}
	}
	return database.ErrNotFound
}

func (f *ownerRepository) DeleteOwnerTerm(_ context.Context, _, id uuid.UUID) error {
	for i := range f.terms {
		if f.terms[i].ID == id {
			f.terms = append(f.terms[:i], f.terms[i+1:]...)
			return nil
		}
	}
	return database.ErrNotFound
}

func (f *ownerRepository) Billing() BillingLedger {
	return f
}

func (f *ownerRepository) RecordPayable(context.Context, PayableEntry) (uuid.UUID, error) {
	return uuid.Nil, errors.New("owner billing never records a payable")
}

func (f *ownerRepository) RecordReceivable(_ context.Context, entry ReceivableEntry) (uuid.UUID, error) {
	f.billed = append(f.billed, entry)
	return uuid.New(), nil
}

func ownerService(repository *ownerRepository) *service {
	evaluator := NewService(repository).(*service)
	evaluator.now = func() time.Time { return time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC) }
	return evaluator
}

func TestOwnerTermClaimBillsTheOwnerAsAReceivable(t *testing.T) {
	repository := newOwnerRepository(statusOngoing, 45)
	term := repository.addTerm(40, 40)
	actor := uuid.New()

	schedule, err := ownerService(repository).ClaimOwnerTerm(context.Background(), repository.project.ID, term.ID, actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.billed) != 1 {
		t.Fatalf("want one receivable, got %+v", repository.billed)
	}
	billed := repository.billed[0]
	if billed.CustomerID != repository.contract.CustomerID || billed.ProjectID != repository.project.ID {
		t.Fatalf("receivable must bill the contract's owner for this project, got %+v", billed)
	}
	if billed.Amount != 7_600_000_000 || billed.Reference != "Tagihan termin 1 DEMO-RSIA" {
		t.Fatalf("want 40%% of Rp 20 M minus 5%% retention, got %+v", billed)
	}
	if !billed.DueDate.Equal(time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("due date = %s, want 30 days after the claim", billed.DueDate)
	}

	claimed := schedule.Terms[0]
	if claimed.State != termClaimed || claimed.ReceivableID == nil || claimed.NetAmount != 7_600_000_000 || claimed.RetentionAmount != 400_000_000 {
		t.Fatalf("got term %+v", claimed)
	}
	if !schedule.ContractLocked || schedule.ClaimedAmount != 8_000_000_000 || schedule.RetentionHeld != 400_000_000 {
		t.Fatalf("got schedule %+v", schedule)
	}
	if schedule.Contract == nil || schedule.Contract.CustomerName != "Yayasan RSIA Cirebon" {
		t.Fatalf("got contract %+v", schedule.Contract)
	}
}

func TestOwnerTermWaitsForTheRequiredProgressAndTheTermBeforeIt(t *testing.T) {
	repository := newOwnerRepository(statusOngoing, 30)
	first := repository.addTerm(20, 20)
	second := repository.addTerm(30, 20)
	service := ownerService(repository)

	_, err := service.ClaimOwnerTerm(context.Background(), repository.project.ID, second.ID, uuid.New())
	requireCode(t, err, "payment_term_out_of_order")

	repository.project.Progress = 10
	_, err = service.ClaimOwnerTerm(context.Background(), repository.project.ID, first.ID, uuid.New())
	requireCode(t, err, "payment_term_progress_not_reached")
	if len(repository.billed) != 0 {
		t.Fatal("a refused claim must not bill the owner")
	}
}

func TestOwnerContractLocksItsTermsOnceATermIsBilled(t *testing.T) {
	repository := newOwnerRepository(statusOngoing, 50)
	term := repository.addTerm(20, 30)
	service := ownerService(repository)
	if _, err := service.ClaimOwnerTerm(context.Background(), repository.project.ID, term.ID, uuid.New()); err != nil {
		t.Fatalf("claim: %v", err)
	}

	retention, days := 5.0, 45
	request := OwnerContractRequest{CustomerID: repository.contract.CustomerID, Number: "KTR-RSIA-01", Value: ownerContractValue, RetentionPercent: &retention, PaymentTermDays: &days}
	_, err := service.SaveOwnerContract(context.Background(), repository.project.ID, request)
	requireCode(t, err, "owner_contract_locked")

	days = 30
	request.Number = "KTR-RSIA-01/REV"
	if _, err := service.SaveOwnerContract(context.Background(), repository.project.ID, request); err != nil {
		t.Fatalf("renaming the contract number stays allowed: %v", err)
	}
}

func TestOwnerContractNeedsAKnownCustomer(t *testing.T) {
	repository := newOwnerRepository(statusOngoing, 0)
	repository.contract = nil
	retention, days := 5.0, 30

	_, err := ownerService(repository).SaveOwnerContract(context.Background(), repository.project.ID, OwnerContractRequest{
		CustomerID: uuid.New(), Number: "KTR-01", Value: 1_000_000, RetentionPercent: &retention, PaymentTermDays: &days,
	})
	requireCode(t, err, "customer_not_found")
}

func TestOwnerRetentionIsBilledOnceTheProjectIsCompleted(t *testing.T) {
	repository := newOwnerRepository(statusOngoing, 100)
	first := repository.addTerm(50, 60)
	second := repository.addTerm(100, 40)
	service := ownerService(repository)
	for _, term := range []OwnerTerm{first, second} {
		if _, err := service.ClaimOwnerTerm(context.Background(), repository.project.ID, term.ID, uuid.New()); err != nil {
			t.Fatalf("claim: %v", err)
		}
	}

	_, err := service.ReleaseOwnerRetention(context.Background(), repository.project.ID, uuid.New())
	requireCode(t, err, "retention_not_due")

	repository.project.Status = statusCompleted
	schedule, err := service.ReleaseOwnerRetention(context.Background(), repository.project.ID, uuid.New())
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	last := repository.billed[len(repository.billed)-1]
	if last.Amount != 1_000_000_000 || last.Reference != "Tagihan retensi DEMO-RSIA" {
		t.Fatalf("want the whole 5%% retention billed, got %+v", last)
	}
	if !schedule.RetentionReleased || schedule.CanReleaseRetention {
		t.Fatalf("got schedule %+v", schedule)
	}

	_, err = service.ReleaseOwnerRetention(context.Background(), repository.project.ID, uuid.New())
	requireCode(t, err, "retention_released")
}

func TestOwnerScheduleWithoutContractStillListsTheTerms(t *testing.T) {
	repository := newOwnerRepository(statusOngoing, 0)
	repository.contract = nil

	schedule, err := ownerService(repository).OwnerSchedule(context.Background(), repository.project.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schedule.Contract != nil || len(schedule.Terms) != 0 || schedule.CanReleaseRetention {
		t.Fatalf("got %+v", schedule)
	}

	_, err = ownerService(repository).CreateOwnerTerm(context.Background(), repository.project.ID, termRequest(10, 10))
	requireCode(t, err, "owner_contract_not_found")
}
