package billing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

type ReceivableEntry struct {
	CustomerID uuid.UUID
	ProjectID  uuid.UUID
	Reference  string
	DueDate    time.Time
	Amount     int64
}

type ReceivableLedger struct {
	repository Repository
	now        func() time.Time
}

func NewReceivableLedger(repository Repository) ReceivableLedger {
	return ReceivableLedger{repository: repository, now: time.Now}
}

func (l ReceivableLedger) Record(ctx context.Context, entry ReceivableEntry) (uuid.UUID, error) {
	request := ReceivableRequest{
		CustomerID: entry.CustomerID,
		ProjectID:  &entry.ProjectID,
		Reference:  entry.Reference,
		DueDate:    entry.DueDate,
		Amount:     entry.Amount,
	}
	if err := ensureReceivableLinks(ctx, l.repository, request, uuid.Nil); err != nil {
		return uuid.Nil, err
	}

	var receivable Receivable
	applyReceivableRequest(&receivable, request, duedate.Today(l.now()))
	if err := l.repository.CreateReceivable(ctx, &receivable); err != nil {
		return uuid.Nil, receivableWriteErrors.Resolve(err)
	}
	return receivable.ID, nil
}
