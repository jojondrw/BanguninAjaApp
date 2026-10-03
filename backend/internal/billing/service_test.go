package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

type fakeRepository struct {
	Repository
	invoice             Invoice
	paidMeanwhile       *Invoice
	saved               *Invoice
	recordedAs          *Receivable
	partyExists         bool
	deleteCalled        bool
	inTransaction       bool
	lockedInTransaction bool
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	f.inTransaction = true
	defer func() { f.inTransaction = false }()
	return work(f)
}

func (f *fakeRepository) LockInvoice(context.Context, uuid.UUID) (Invoice, error) {
	f.lockedInTransaction = f.inTransaction
	if f.paidMeanwhile != nil {
		return *f.paidMeanwhile, nil
	}
	return f.invoice, nil
}

func (f *fakeRepository) FindInvoice(context.Context, uuid.UUID) (Invoice, error) {
	return f.invoice, nil
}

func (f *fakeRepository) FindInvoiceRow(context.Context, uuid.UUID) (InvoiceRow, error) {
	if f.saved != nil {
		return InvoiceRow{Invoice: *f.saved}, nil
	}
	return InvoiceRow{Invoice: f.invoice}, nil
}

func (f *fakeRepository) FindReceivableByInvoice(context.Context, uuid.UUID) (Receivable, error) {
	if f.recordedAs == nil {
		return Receivable{}, database.ErrNotFound
	}
	return *f.recordedAs, nil
}

func (f *fakeRepository) SaveInvoice(_ context.Context, invoice *Invoice) error {
	f.saved = invoice
	return nil
}

func (f *fakeRepository) CreatePayment(context.Context, *Payment) error {
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

	invoice, err := fixedService(repository, today).PayInvoice(context.Background(), uuid.New(), PaymentRequest{Amount: 600}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.Status != duedate.Paid || invoice.Outstanding != 0 || repository.saved.Status != duedate.Paid {
		t.Fatalf("got status %s outstanding %d", invoice.Status, invoice.Outstanding)
	}
}

func TestPaymentCannotExceedOutstanding(t *testing.T) {
	repository := &fakeRepository{invoice: Invoice{Amount: 1000, PaidAmount: 900}}

	_, err := fixedService(repository, time.Now()).PayInvoice(context.Background(), uuid.New(), PaymentRequest{Amount: 200}, nil)
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

func TestInvoiceUpdateKeepsPaymentRecordedMeanwhile(t *testing.T) {
	repository := &fakeRepository{
		invoice:       Invoice{Amount: 1000},
		paidMeanwhile: &Invoice{Amount: 1000, PaidAmount: 400},
		partyExists:   true,
	}

	_, err := fixedService(repository, time.Now()).UpdateInvoice(context.Background(), uuid.New(), InvoiceRequest{PartyType: "customer", Amount: 1200})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repository.lockedInTransaction {
		t.Fatal("invoice must be locked inside the update transaction")
	}
	if repository.saved == nil || repository.saved.PaidAmount != 400 {
		t.Fatalf("update overwrote the payment, saved %+v", repository.saved)
	}
}

func TestInvoiceUpdateBelowPaymentRecordedMeanwhileIsRejected(t *testing.T) {
	repository := &fakeRepository{
		invoice:       Invoice{Amount: 1000},
		paidMeanwhile: &Invoice{Amount: 1000, PaidAmount: 900},
		partyExists:   true,
	}

	_, err := fixedService(repository, time.Now()).UpdateInvoice(context.Background(), uuid.New(), InvoiceRequest{PartyType: "customer", Amount: 500})
	if !errors.Is(err, errAmountBelowPaid) || repository.saved != nil {
		t.Fatalf("got %v saved %+v", err, repository.saved)
	}
}

func TestInvoiceDeleteRefusesPaymentRecordedMeanwhile(t *testing.T) {
	repository := &fakeRepository{
		invoice:       Invoice{Amount: 1000},
		paidMeanwhile: &Invoice{Amount: 1000, PaidAmount: 1},
	}

	err := fixedService(repository, time.Now()).DeleteInvoice(context.Background(), uuid.New())
	if !errors.Is(err, errInvoiceHasPayment) || repository.deleteCalled {
		t.Fatalf("got %v deleted %v", err, repository.deleteCalled)
	}
	if !repository.lockedInTransaction {
		t.Fatal("invoice must be locked inside the delete transaction")
	}
}

func TestUnpaidInvoiceIsDeleted(t *testing.T) {
	repository := &fakeRepository{invoice: Invoice{Amount: 1000}}

	if err := fixedService(repository, time.Now()).DeleteInvoice(context.Background(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repository.deleteCalled {
		t.Fatal("unpaid invoice must be deleted")
	}
}

type ledgerRepository struct {
	Repository
	receivable          Receivable
	payable             Payable
	savedReceivable     *Receivable
	payableDeleted      bool
	inTransaction       bool
	lockedInTransaction bool
}

func (f *ledgerRepository) Transaction(_ context.Context, work func(Repository) error) error {
	f.inTransaction = true
	defer func() { f.inTransaction = false }()
	return work(f)
}

func (f *ledgerRepository) LockReceivable(context.Context, uuid.UUID) (Receivable, error) {
	f.lockedInTransaction = f.inTransaction
	return f.receivable, nil
}

func (f *ledgerRepository) SaveReceivable(_ context.Context, receivable *Receivable) error {
	f.savedReceivable = receivable
	return nil
}

func (f *ledgerRepository) FindReceivableRow(context.Context, uuid.UUID) (ReceivableRow, error) {
	return ReceivableRow{Receivable: *f.savedReceivable}, nil
}

func (f *ledgerRepository) PartyExists(context.Context, string, uuid.UUID) (bool, error) {
	return true, nil
}

func (f *ledgerRepository) LockPayable(context.Context, uuid.UUID) (Payable, error) {
	f.lockedInTransaction = f.inTransaction
	return f.payable, nil
}

func (f *ledgerRepository) DeletePayable(context.Context, uuid.UUID) error {
	f.payableDeleted = true
	return nil
}

func TestReceivableUpdateKeepsLockedPayment(t *testing.T) {
	repository := &ledgerRepository{receivable: Receivable{Amount: 1000, PaidAmount: 250}}

	response, err := fixedService(repository, time.Now()).UpdateReceivable(context.Background(), uuid.New(), ReceivableRequest{Reference: "KTR-1", Amount: 1500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repository.lockedInTransaction || repository.savedReceivable.PaidAmount != 250 || response.Outstanding != 1250 {
		t.Fatalf("locked %v saved %+v outstanding %d", repository.lockedInTransaction, repository.savedReceivable, response.Outstanding)
	}
}

func TestPayableDeleteRefusesLockedPayment(t *testing.T) {
	repository := &ledgerRepository{payable: Payable{Amount: 1000, PaidAmount: 1}}

	err := fixedService(repository, time.Now()).DeletePayable(context.Background(), uuid.New())
	if !errors.Is(err, errPayableHasPayment) || repository.payableDeleted || !repository.lockedInTransaction {
		t.Fatalf("got %v deleted %v locked %v", err, repository.payableDeleted, repository.lockedInTransaction)
	}
}

func TestStatusIsRecomputedWhenRead(t *testing.T) {
	today := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	stale := Invoice{Amount: 1000, DueDate: today.AddDate(0, 0, -40), Status: duedate.NotDue}

	response := newInvoiceResponse(InvoiceRow{Invoice: stale}, today)
	if response.Status != duedate.Overdue || response.DaysOverdue != 40 {
		t.Fatalf("got status %s days %d", response.Status, response.DaysOverdue)
	}
}
