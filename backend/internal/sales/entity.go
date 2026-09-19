package sales

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Customer struct {
	entity.Base
	Name           string `gorm:"type:varchar(160);not null"`
	Contact        string `gorm:"type:varchar(60)"`
	Email          string `gorm:"type:varchar(160)"`
	IdentityNumber string `gorm:"type:varchar(20);uniqueIndex:uq_customer_identity_number"`
	Address        string `gorm:"type:text"`
}

func (Customer) TableName() string {
	return "customer"
}

type Unit struct {
	entity.Base
	Code      string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_unit_code"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index:idx_unit_project"`
	UnitType  string    `gorm:"type:varchar(60);not null"`
	AreaSqm   float64   `gorm:"type:numeric(10,2);not null;default:0"`
	Price     int64     `gorm:"not null;default:0"`
	Status    string    `gorm:"type:varchar(20);not null;index:idx_unit_status"`
}

func (Unit) TableName() string {
	return "unit"
}

type Lead struct {
	entity.Base
	Name            string     `gorm:"type:varchar(160);not null"`
	Contact         string     `gorm:"type:varchar(60)"`
	ProjectID       *uuid.UUID `gorm:"type:uuid;index:idx_lead_project"`
	Source          string     `gorm:"type:varchar(60)"`
	Stage           string     `gorm:"type:varchar(20);not null;index:idx_lead_stage"`
	LastContactedAt *time.Time `gorm:"type:date"`
}

func (Lead) TableName() string {
	return "lead"
}

type Contract struct {
	entity.Base
	Number     string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_contract_number"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index:idx_contract_customer"`
	UnitID     uuid.UUID `gorm:"type:uuid;not null;index:idx_contract_unit"`
	Type       string    `gorm:"type:varchar(20);not null"`
	Value      int64     `gorm:"not null;default:0"`
	Date       time.Time `gorm:"type:date;not null"`
	Status     string    `gorm:"type:varchar(20);not null;index:idx_contract_status"`
}

func (Contract) TableName() string {
	return "contract"
}

type Installment struct {
	entity.Base
	ContractID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_installment_contract"`
	InstallmentNumber int        `gorm:"not null"`
	DueDate           time.Time  `gorm:"type:date;not null;index:idx_installment_due_date"`
	Amount            int64      `gorm:"not null;default:0"`
	PaidDate          *time.Time `gorm:"type:date"`
	Status            string     `gorm:"type:varchar(20);not null;index:idx_installment_status"`
}

func (Installment) TableName() string {
	return "installment"
}

func Entities() []any {
	return []any{&Customer{}, &Unit{}, &Lead{}, &Contract{}, &Installment{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_customer_name_trgm ON customer USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_lead_name_trgm ON lead USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_unit_project_status ON unit (project_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_installment_unpaid ON installment (due_date) WHERE status <> 'paid'`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("unit", "project_id", "project", database.DeleteRestrict),
		database.ForeignKey("lead", "project_id", "project", database.DeleteSetNull),
		database.ForeignKey("contract", "customer_id", "customer", database.DeleteRestrict),
		database.ForeignKey("contract", "unit_id", "unit", database.DeleteRestrict),
		database.ForeignKey("installment", "contract_id", "contract", database.DeleteCascade),
		database.Check("unit", "status", "status IN ('available','reserved','sold','on_hold')"),
		database.Check("unit", "price", "price >= 0"),
		database.Check("unit", "area", "area_sqm > 0"),
		database.Check("lead", "stage", "stage IN ('new','interested','negotiating','won','cancelled')"),
		database.Check("contract", "type", "type IN ('cash','mortgage','installment')"),
		database.Check("contract", "status", "status IN ('draft','active','paid','cancelled')"),
		database.Check("contract", "value", "value >= 0"),
		database.Check("installment", "installment_number", "installment_number > 0"),
		database.Check("installment", "amount", "amount >= 0"),
		database.Check("installment", "status", "status IN ('not_due','due','paid','overdue')"),
		database.Unique("contract", "unit_number", "unit_id, number"),
		database.Unique("installment", "contract_number", "contract_id, installment_number"),
	}
}
