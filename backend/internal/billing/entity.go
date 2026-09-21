package billing

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Invoice struct {
	entity.Base
	Number     string     `gorm:"type:varchar(40);not null;uniqueIndex:uq_invoice_number"`
	Note       string     `gorm:"type:varchar(200)"`
	PartyType  string     `gorm:"type:varchar(20);not null"`
	PartyID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_invoice_pihak"`
	ProjectID  *uuid.UUID `gorm:"type:uuid;index:idx_invoice_project"`
	DueDate    time.Time  `gorm:"type:date;not null;index:idx_invoice_due_date"`
	Amount     int64      `gorm:"not null;default:0"`
	PaidAmount int64      `gorm:"not null;default:0"`
	Status     string     `gorm:"type:varchar(20);not null;index:idx_invoice_status"`
}

func (Invoice) TableName() string {
	return "invoice"
}

type Receivable struct {
	entity.Base
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index:idx_receivable_customer"`
	Reference  string    `gorm:"type:varchar(60);not null"`
	DueDate    time.Time `gorm:"type:date;not null;index:idx_receivable_due_date"`
	Amount     int64     `gorm:"not null;default:0"`
	PaidAmount int64     `gorm:"not null;default:0"`
	Status     string    `gorm:"type:varchar(20);not null;index:idx_receivable_status"`
}

func (Receivable) TableName() string {
	return "receivable"
}

type Payable struct {
	entity.Base
	VendorID   uuid.UUID `gorm:"type:uuid;not null;index:idx_payable_vendor"`
	Reference  string    `gorm:"type:varchar(60);not null"`
	DueDate    time.Time `gorm:"type:date;not null;index:idx_payable_due_date"`
	Amount     int64     `gorm:"not null;default:0"`
	PaidAmount int64     `gorm:"not null;default:0"`
	Status     string    `gorm:"type:varchar(20);not null;index:idx_payable_status"`
}

func (Payable) TableName() string {
	return "payable"
}

func Entities() []any {
	return []any{&Invoice{}, &Receivable{}, &Payable{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_invoice_unpaid ON invoice (due_date) WHERE status <> 'paid'`,
		`CREATE INDEX IF NOT EXISTS idx_receivable_unpaid ON receivable (due_date) WHERE status <> 'paid'`,
		`CREATE INDEX IF NOT EXISTS idx_payable_unpaid ON payable (due_date) WHERE status <> 'paid'`,
	}
}

func Constraints() []string {
	statuses := "status IN ('not_due','due','paid','overdue')"

	return []string{
		database.ForeignKey("invoice", "project_id", "project", database.DeleteSetNull),
		database.ForeignKey("receivable", "customer_id", "customer", database.DeleteRestrict),
		database.ForeignKey("payable", "vendor_id", "vendor", database.DeleteRestrict),
		database.Check("invoice", "party_type", "party_type IN ('customer','vendor')"),
		database.Check("invoice", "status", statuses),
		database.Check("receivable", "status", statuses),
		database.Check("payable", "status", statuses),
		database.Check("invoice", "amount", "amount >= 0 AND paid_amount >= 0 AND paid_amount <= amount"),
		database.Check("receivable", "amount", "amount >= 0 AND paid_amount >= 0 AND paid_amount <= amount"),
		database.Check("payable", "amount", "amount >= 0 AND paid_amount >= 0 AND paid_amount <= amount"),
	}
}
