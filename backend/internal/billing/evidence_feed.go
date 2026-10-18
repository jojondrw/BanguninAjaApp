package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type PaymentFact struct {
	ID        uuid.UUID
	ProjectID *uuid.UUID
	Target    string
	Parent    string
	Amount    int64
	PaidAt    time.Time
	Method    string
	Reference string
}

type PaymentFeed struct {
	db *gorm.DB
}

func NewPaymentFeed(db *gorm.DB) PaymentFeed {
	return PaymentFeed{db: db}
}

func (f PaymentFeed) Payments(ctx context.Context, projectID *uuid.UUID) ([]PaymentFact, error) {
	query := f.db.WithContext(ctx).Table("billing_payment AS p").
		Select(`p.id, p.amount, p.paid_at, p.method, p.reference,
			COALESCE(i.project_id, r.project_id, b.project_id) AS project_id,
			CASE WHEN p.invoice_id IS NOT NULL THEN 'invoice' WHEN p.receivable_id IS NOT NULL THEN 'receivable' ELSE 'payable' END AS target,
			COALESCE(i.number, r.reference, b.reference, '') AS parent`).
		Joins("LEFT JOIN invoice i ON i.id = p.invoice_id").
		Joins("LEFT JOIN receivable r ON r.id = p.receivable_id").
		Joins("LEFT JOIN payable b ON b.id = p.payable_id").
		Order("p.paid_at, p.id")
	if projectID != nil {
		query = query.Where("COALESCE(i.project_id, r.project_id, b.project_id) = ?", *projectID)
	}
	var rows []PaymentFact
	if err := query.Scan(&rows).Error; err != nil {
		return nil, database.Translate(err)
	}
	return rows, nil
}
