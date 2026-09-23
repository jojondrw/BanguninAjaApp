package sales

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

type fakeRepository struct {
	Repository
	unit        Unit
	contract    Contract
	installment Installment
	unpaid      int64
	scheduled   int64
	created     *Contract
	savedUnit   *Unit
	savedPlan   *Installment
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) FindUnit(context.Context, uuid.UUID) (Unit, error) {
	return f.unit, nil
}

func (f *fakeRepository) LockUnit(context.Context, uuid.UUID) (Unit, error) {
	return f.unit, nil
}

func (f *fakeRepository) SaveUnit(_ context.Context, unit *Unit) error {
	f.savedUnit = unit
	return nil
}

func (f *fakeRepository) LockContract(context.Context, uuid.UUID) (Contract, error) {
	return f.contract, nil
}

func (f *fakeRepository) CreateContract(_ context.Context, contract *Contract) error {
	f.created = contract
	return nil
}

func (f *fakeRepository) SaveContract(_ context.Context, contract *Contract) error {
	f.contract = *contract
	return nil
}

func (f *fakeRepository) FindContractRow(context.Context, uuid.UUID) (ContractRow, error) {
	return ContractRow{Status: f.contract.Status}, nil
}

func (f *fakeRepository) CountUnpaidInstallments(context.Context, uuid.UUID) (int64, error) {
	return f.unpaid, nil
}

func (f *fakeRepository) SumInstallments(context.Context, uuid.UUID, *uuid.UUID) (int64, error) {
	return f.scheduled, nil
}

func (f *fakeRepository) FindInstallment(context.Context, uuid.UUID, uuid.UUID) (Installment, error) {
	return f.installment, nil
}

func (f *fakeRepository) SaveInstallment(_ context.Context, installment *Installment) error {
	f.savedPlan = installment
	return nil
}

func (f *fakeRepository) CreateInstallment(_ context.Context, installment *Installment) error {
	f.savedPlan = installment
	return nil
}

func (f *fakeRepository) FindInstallmentRow(context.Context, uuid.UUID) (InstallmentRow, error) {
	return InstallmentRow{Status: f.savedPlan.Status, DueDate: f.savedPlan.DueDate}, nil
}

func TestCanTransitionContract(t *testing.T) {
	cases := []struct {
		current, next string
		allowed       bool
	}{
		{contractDraft, contractActive, true},
		{contractDraft, contractPaid, false},
		{contractActive, contractPaid, true},
		{contractActive, contractCancelled, true},
		{contractPaid, contractCancelled, false},
		{contractCancelled, contractActive, false},
	}
	for _, tc := range cases {
		if got := canTransitionContract(tc.current, tc.next); got != tc.allowed {
			t.Errorf("%s -> %s: got %v, want %v", tc.current, tc.next, got, tc.allowed)
		}
	}
}

func TestNewContractReservesUnit(t *testing.T) {
	repository := &fakeRepository{unit: Unit{Status: unitAvailable}}

	_, err := NewService(repository).CreateContract(context.Background(), ContractRequest{Number: "KTR-01", Value: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.created == nil || repository.created.Status != contractDraft || repository.savedUnit.Status != unitReserved {
		t.Fatalf("got contract %+v unit %+v", repository.created, repository.savedUnit)
	}
}

func TestReservedUnitCannotGetSecondContract(t *testing.T) {
	repository := &fakeRepository{unit: Unit{Status: unitReserved}}

	_, err := NewService(repository).CreateContract(context.Background(), ContractRequest{Number: "KTR-02", Value: 100})
	if !errors.Is(err, errUnitNotAvailable) || repository.created != nil {
		t.Fatalf("got %v created %v", err, repository.created)
	}
}

func TestPaidContractRequiresSettledInstallments(t *testing.T) {
	repository := &fakeRepository{contract: Contract{Status: contractActive}, unpaid: 2}

	_, err := NewService(repository).UpdateContractStatus(context.Background(), uuid.New(), ContractStatusRequest{Status: contractPaid})
	if !errors.Is(err, errContractUnpaid) {
		t.Fatalf("got %v, want errContractUnpaid", err)
	}
}

func TestPaidContractSellsUnit(t *testing.T) {
	repository := &fakeRepository{contract: Contract{Status: contractActive}, unit: Unit{Status: unitReserved}}

	_, err := NewService(repository).UpdateContractStatus(context.Background(), uuid.New(), ContractStatusRequest{Status: contractPaid})
	if err != nil || repository.savedUnit == nil || repository.savedUnit.Status != unitSold {
		t.Fatalf("got err %v unit %+v", err, repository.savedUnit)
	}
}

func TestCancelledContractFreesUnit(t *testing.T) {
	repository := &fakeRepository{contract: Contract{Status: contractDraft}, unit: Unit{Status: unitReserved}}

	_, err := NewService(repository).UpdateContractStatus(context.Background(), uuid.New(), ContractStatusRequest{Status: contractCancelled})
	if err != nil || repository.savedUnit.Status != unitAvailable {
		t.Fatalf("got err %v unit %+v", err, repository.savedUnit)
	}
}

func TestUnitBoundToContractCannotBeEdited(t *testing.T) {
	repository := &fakeRepository{unit: Unit{Status: unitSold}}

	_, err := NewService(repository).UpdateUnit(context.Background(), uuid.New(), UnitRequest{Status: unitAvailable})
	if !errors.Is(err, errUnitLocked) {
		t.Fatalf("got %v, want errUnitLocked", err)
	}
}

func TestInstallmentsCannotExceedContractValue(t *testing.T) {
	repository := &fakeRepository{contract: Contract{Status: contractActive, Value: 1000}, scheduled: 800}

	_, err := NewService(repository).CreateInstallment(context.Background(), uuid.New(), InstallmentRequest{InstallmentNumber: 3, Amount: 300})
	if !errors.Is(err, errInstallmentExceeds) {
		t.Fatalf("got %v, want errInstallmentExceeds", err)
	}
}

func TestInstallmentPaymentNeedsActiveContract(t *testing.T) {
	repository := &fakeRepository{contract: Contract{Status: contractDraft}}

	_, err := NewService(repository).PayInstallment(context.Background(), uuid.New(), uuid.New(), InstallmentPaymentRequest{PaidDate: time.Now()})
	if !errors.Is(err, errContractNotActive) {
		t.Fatalf("got %v, want errContractNotActive", err)
	}
}

func TestPaidInstallmentIsLocked(t *testing.T) {
	repository := &fakeRepository{contract: Contract{Status: contractActive, Value: 1000}, installment: Installment{Status: duedate.Paid}}

	err := NewService(repository).DeleteInstallment(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, errInstallmentPaid) {
		t.Fatalf("got %v, want errInstallmentPaid", err)
	}
}
