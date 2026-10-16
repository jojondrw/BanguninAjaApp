package project

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

func changeRequest(value int64, days int) ContractChangeRequest {
	return ContractChangeRequest{Description: "  Penambahan ruang NICU lantai 3  ", ValueChange: value, DaysChange: days}
}

func proposed(repository *termRepository, value int64) ContractChange {
	change := ContractChange{ProjectID: repository.project.ID, Sequence: len(repository.changes) + 1, Description: "Uji", ValueChange: value, Status: changeProposed}
	change.ID = uuid.New()
	repository.changes = append(repository.changes, change)
	return change
}

func approve(repository *termRepository, changeID uuid.UUID, role string) (PaymentScheduleResponse, error) {
	return termService(repository).DecideContractChange(context.Background(), uuid.New(), changeID, uuid.New(), role,
		ContractChangeDecision{Status: changeApproved})
}

func TestProposedChangeGetsTheNextNumberAndItsProposer(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	proposed(repository, 500_000_000)
	proposer := uuid.New()

	schedule, err := termService(repository).ProposeContractChange(context.Background(), uuid.New(), proposer, changeRequest(1_250_000_000, 30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := repository.changes[1]
	if stored.Sequence != 2 || stored.Status != changeProposed || stored.Description != "Penambahan ruang NICU lantai 3" {
		t.Fatalf("stored = %+v", stored)
	}
	if stored.ProposedBy == nil || *stored.ProposedBy != proposer {
		t.Fatal("proposer not recorded")
	}
	if schedule.Contract.CurrentValue != contractValue || schedule.Changes[1].Label != "Adendum 2" {
		t.Fatalf("a proposal changed the contract value: %+v", schedule.Contract)
	}
}

func TestProposalsMustChangeSomethingAndKeepTheContractPositive(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)

	if _, err := termService(repository).ProposeContractChange(context.Background(), uuid.New(), uuid.New(), changeRequest(0, 0)); !errors.Is(err, errChangeEmpty) {
		t.Fatalf("empty change err = %v, want errChangeEmpty", err)
	}
	if _, err := termService(repository).ProposeContractChange(context.Background(), uuid.New(), uuid.New(), changeRequest(-contractValue, 0)); !errors.Is(err, errChangeNonPositive) {
		t.Fatalf("zero contract err = %v, want errChangeNonPositive", err)
	}

	closed := newTermRepository(statusCompleted, 100)
	if _, err := termService(closed).ProposeContractChange(context.Background(), uuid.New(), uuid.New(), changeRequest(1, 0)); !errors.Is(err, errProjectClosed) {
		t.Fatalf("completed project err = %v, want errProjectClosed", err)
	}
}

func TestOnlyAdminDecidesAChange(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	change := proposed(repository, 1_000_000_000)

	if _, err := approve(repository, change.ID, access.RoleProjectManager); !errors.Is(err, errChangeApprover) {
		t.Fatalf("project manager err = %v, want errChangeApprover", err)
	}
	if repository.changes[0].Status != changeProposed {
		t.Fatal("a project manager decided the change")
	}
}

func TestApprovedChangeRaisesTheRunningValueAndTheRemainingTerms(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	repository.claim(repository.addTerm(20, 20))
	repository.addTerm(100, 80)
	change := proposed(repository, 3_000_000_000)
	change.DaysChange = 30
	repository.changes[0] = change

	schedule, err := approve(repository, change.ID, access.RoleAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decided := repository.changes[0]
	if decided.Status != changeApproved || decided.DecidedBy == nil || decided.DecidedOn == nil || !decided.DecidedOn.Equal(date(2026, 10, 17)) {
		t.Fatalf("decision not recorded: %+v", decided)
	}
	if schedule.Contract.CurrentValue != 33_000_000_000 || schedule.Contract.ApprovedChange != 3_000_000_000 || schedule.Contract.ExtensionDays != 30 {
		t.Fatalf("contract = %+v, want 33 M after the change and 30 extra days", schedule.Contract)
	}
	if schedule.Terms[1].GrossAmount != 27_000_000_000 {
		t.Fatalf("last term = %d, want the running value minus the 6 M already billed", schedule.Terms[1].GrossAmount)
	}
}

func TestChangeCannotCutTheContractBelowWhatWasBilled(t *testing.T) {
	repository := newTermRepository(statusOngoing, 60)
	repository.claim(repository.addTerm(50, 60))
	change := proposed(repository, -15_000_000_000)

	if _, err := approve(repository, change.ID, access.RoleAdmin); !errors.Is(err, errChangeBelowClaimed) {
		t.Fatalf("err = %v, want errChangeBelowClaimed", err)
	}
}

func TestDecidedChangesAreFinal(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	change := proposed(repository, 1_000_000_000)
	if _, err := termService(repository).DecideContractChange(context.Background(), uuid.New(), change.ID, uuid.New(), access.RoleAdmin,
		ContractChangeDecision{Status: changeRejected}); err != nil {
		t.Fatalf("reject: %v", err)
	}

	if _, err := approve(repository, change.ID, access.RoleAdmin); !errors.Is(err, errChangeDecided) {
		t.Fatalf("second decision err = %v, want errChangeDecided", err)
	}
	if _, err := termService(repository).DeleteContractChange(context.Background(), uuid.New(), change.ID); !errors.Is(err, errChangeDecided) {
		t.Fatalf("delete err = %v, want errChangeDecided", err)
	}
	if schedule, _ := termService(repository).PaymentSchedule(context.Background(), uuid.New()); schedule.Contract.CurrentValue != contractValue {
		t.Fatalf("a rejected change moved the contract value to %d", schedule.Contract.CurrentValue)
	}
}

func TestClaimAfterAnApprovedChangeUsesTheRunningValue(t *testing.T) {
	repository := newTermRepository(statusOngoing, 46)
	repository.claim(repository.addTerm(0, 20))
	next := repository.addTerm(40, 25)
	change := proposed(repository, 3_000_000_000)
	change.Status = changeApproved
	repository.changes[0] = change

	if _, err := termService(repository).ClaimPaymentTerm(context.Background(), uuid.New(), next.ID, uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repository.recorded[0].Amount != 8_407_500_000 {
		t.Fatalf("payable = %d, want 45%% of 33 M minus 6 M billed, minus 5%% retention", repository.recorded[0].Amount)
	}
}
