package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPayablePaymentIsRecordedAsCashOutForItsProject(t *testing.T) {
	store := newLedgerStore()
	projectID := uuid.New()
	payable := Payable{Reference: "UJI-Termin 2", ProjectID: &projectID, Amount: 5000, DueDate: paymentNow.AddDate(0, 1, 0)}
	payable.ID = uuid.New()
	store.payables[payable.ID] = payable
	paidAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	if _, err := fixedService(store, paymentNow).PayPayable(context.Background(), payable.ID, PaymentRequest{Amount: 4000, PaidAt: &paidAt}, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.cash) != 1 {
		t.Fatalf("want one cash entry, got %+v", store.cash)
	}
	entry := store.cash[0]
	if entry.Incoming || entry.Amount != 4000 || !entry.Date.Equal(paidAt) || entry.PaymentID != store.payments[0].ID {
		t.Fatalf("got %+v", entry)
	}
	if entry.ProjectID == nil || *entry.ProjectID != projectID || entry.Note != "Pembayaran hutang UJI-Termin 2" {
		t.Fatalf("got %+v", entry)
	}
}

func TestLinkedInvoicePaymentIsRecordedAsOneCashIn(t *testing.T) {
	store := newLedgerStore()
	invoice, receivable := linkedPair(store, 1000, 1000, 0, 0)
	projectID := uuid.New()
	receivable.ProjectID = &projectID
	store.receivables[receivable.ID] = receivable

	if _, err := fixedService(store, paymentNow).PayInvoice(context.Background(), invoice.ID, PaymentRequest{Amount: 300}, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.cash) != 1 {
		t.Fatalf("a payment that settles both sides is one cash receipt, got %+v", store.cash)
	}
	entry := store.cash[0]
	if !entry.Incoming || entry.Amount != 300 || entry.Note != "Pembayaran faktur "+invoice.Number {
		t.Fatalf("got %+v", entry)
	}
	if entry.ProjectID == nil || *entry.ProjectID != projectID {
		t.Fatalf("the project comes from the linked receivable when the invoice has none, got %+v", entry.ProjectID)
	}
}

func TestPaymentFailsWhenItsCashCannotBeRecorded(t *testing.T) {
	store := newLedgerStore()
	payable := Payable{Amount: 5000}
	payable.ID = uuid.New()
	store.payables[payable.ID] = payable
	store.cashFailure = errors.New("akun 2100 tidak ada")

	if _, err := fixedService(store, paymentNow).PayPayable(context.Background(), payable.ID, PaymentRequest{Amount: 100}, nil); err == nil {
		t.Fatal("a payment whose cash entry fails must fail too, so the transaction rolls back")
	}
}

func TestRepositoryWithoutCashBookSkipsCash(t *testing.T) {
	if err := recordCash(context.Background(), &fakeRepository{}, paymentParents{}, Payment{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
