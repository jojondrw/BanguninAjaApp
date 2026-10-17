package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CashEntry struct {
	PaymentID uuid.UUID
	Incoming  bool
	ProjectID *uuid.UUID
	Date      time.Time
	Amount    int64
	Note      string
}

type CashBook interface {
	Record(ctx context.Context, entry CashEntry) error
}

type CashBookFactory func(db *gorm.DB) CashBook

func recordCash(ctx context.Context, repository Repository, parents paymentParents, payment Payment) error {
	book := repository.CashBook()
	if book == nil {
		return nil
	}
	return book.Record(ctx, parents.cashEntry(payment))
}

func (p paymentParents) cashEntry(payment Payment) CashEntry {
	entry := CashEntry{PaymentID: payment.ID, Incoming: true, Date: payment.PaidAt, Amount: payment.Amount}
	switch {
	case p.payable != nil:
		entry.Incoming = false
		entry.ProjectID = p.payable.ProjectID
		entry.Note = "Pembayaran hutang " + p.payable.Reference
	case p.invoice != nil:
		entry.ProjectID = p.invoice.ProjectID
		entry.Note = "Pembayaran faktur " + p.invoice.Number
	case p.receivable != nil:
		entry.ProjectID = p.receivable.ProjectID
		entry.Note = "Pembayaran piutang " + p.receivable.Reference
	}
	if entry.ProjectID == nil && p.receivable != nil {
		entry.ProjectID = p.receivable.ProjectID
	}
	return entry
}
