package billing

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
	invoice      Invoice
	saved        *Invoice
	partyExists  bool
	deleteCalled bool
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) LockInvoice(context.Context, uuid.UUID) (Invoice, error) {
	return f.invoice, nil
}

func (f *fakeRepository) FindInvoice(context.Context, uuid.UUID) (Invoice, error) {
	return f.invoice, nil
}

func (f *fakeRepository) SaveInvoice(_ context.Context, invoice *Invoice) error {
	f.saved = invoice
	return nil
}

func (f *fakeRepository) DeleteInvoice(context.Context, uuid.UUID) error {
	f.deleteCalled = true
	return nil
}

func (f *fakeRepository) PartyExists(context.Context, string, uuid.UUID) (bool, error) {
	return f.partyExists, nil
}

func fixedService(repository Repository, today time.Time) *service {
	return &service{repository: repository, now: func() time.Time { return today }}
}

func TestPaymentSettlesInvoice(t *testing.T) {
	today := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	repository := &fakeRepository{invoice: Invoice{Amount: 1000, PaidAmount: 400, DueDate: today.AddDate(0, 0, -3)}}

	invoice, err := fixedService(repository, today).PayInvoice(context.Background(), uuid.New(), PaymentRequest{Amount: 600})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.Status != duedate.Paid || invoice.Outstanding != 0 || repository.saved.Status != duedate.Paid {
		t.Fatalf("got status %s outstanding %d", invoice.Status, invoice.Outstanding)
	}
}

func TestPaymentCannotExceedOutstanding(t *testing.T) {
	repository := &fakeRepository{invoice: Invoice{Amount: 1000, PaidAmount: 900}}

	_, err := fixedService(repository, time.Now()).PayInvoice(context.Background(), uuid.New(), PaymentRequest{Amount: 200})
	if !errors.Is(err, errPaymentOverdrawn) {
		t.Fatalf("got %v, want errPaymentOverdrawn", err)
	}
	if repository.saved != nil {
		t.Fatal("invoice must not be saved")
	}
}

func TestPaidInvoiceCannotBeDeleted(t *testing.T) {
	repository := &fakeRepository{invoice: Invoice{Amount: 1000, PaidAmount: 1}}

	err := fixedService(repository, time.Now()).DeleteInvoice(context.Background(), uuid.New())
	if !errors.Is(err, errInvoiceHasPayment) || repository.deleteCalled {
		t.Fatalf("got %v deleted %v", err, repository.deleteCalled)
	}
}

func TestInvoiceAmountCannotDropBelowPaid(t *testing.T) {
	repository := &fakeRepository{invoice: Invoice{Amount: 1000, PaidAmount: 700}, partyExists: true}

	_, err := fixedService(repository, time.Now()).UpdateInvoice(context.Background(), uuid.New(), InvoiceRequest{Amount: 500})
	if !errors.Is(err, errAmountBelowPaid) {
		t.Fatalf("got %v, want errAmountBelowPaid", err)
	}
}

func TestInvoiceForUnknownPartyIsRejected(t *testing.T) {
	repository := &fakeRepository{partyExists: false}

	_, err := fixedService(repository, time.Now()).CreateInvoice(context.Background(), InvoiceRequest{PartyType: "vendor", Amount: 10})
	if !errors.Is(err, errInvoicePartyMissing) {
		t.Fatalf("got %v, want errInvoicePartyMissing", err)
	}
}

func TestStatusIsRecomputedWhenRead(t *testing.T) {
	today := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	stale := Invoice{Amount: 1000, DueDate: today.AddDate(0, 0, -40), Status: duedate.NotDue}

	response := newInvoiceResponse(stale, today)
	if response.Status != duedate.Overdue || response.DaysOverdue != 40 {
		t.Fatalf("got status %s days %d", response.Status, response.DaysOverdue)
	}
}
