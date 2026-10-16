package billing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

type PayableEntry struct {
	VendorID  uuid.UUID
	ProjectID uuid.UUID
	Reference string
	DueDate   time.Time
	Amount    int64
}

type PayableLedger struct {
	repository Repository
	now        func() time.Time
}

func NewPayableLedger(repository Repository) PayableLedger {
	return PayableLedger{repository: repository, now: time.Now}
}

func (l PayableLedger) Record(ctx context.Context, entry PayableEntry) (uuid.UUID, error) {
	request := PayableRequest{
		VendorID:  entry.VendorID,
		ProjectID: &entry.ProjectID,
		Reference: entry.Reference,
		DueDate:   entry.DueDate,
		Amount:    entry.Amount,
	}
	if err := ensurePayableLinks(ctx, l.repository, request); err != nil {
		return uuid.Nil, err
	}

	var payable Payable
	applyPayableRequest(&payable, request, duedate.Today(l.now()))
	if err := l.repository.CreatePayable(ctx, &payable); err != nil {
		return uuid.Nil, payableWriteErrors.Resolve(err)
	}
	return payable.ID, nil
}
