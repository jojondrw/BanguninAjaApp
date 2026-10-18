package wageproof

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Proof struct {
	entity.Base
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_wage_proof_period,priority:1"`
	Period      string     `gorm:"type:varchar(7);not null;uniqueIndex:uq_wage_proof_period,priority:2"`
	RegionCode  string     `gorm:"type:varchar(20);not null"`
	RegionName  string     `gorm:"type:varchar(120);not null"`
	MinimumWage int64      `gorm:"not null"`
	WageSource  string     `gorm:"type:varchar(200);not null"`
	WorkerCount int        `gorm:"not null"`
	Root        string     `gorm:"type:varchar(70);not null"`
	Proof       []byte     `gorm:"type:bytea;not null"`
	Token       string     `gorm:"type:varchar(40);not null;uniqueIndex:uq_wage_proof_token"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid;index:idx_wage_proof_created_by"`
}

func (Proof) TableName() string {
	return "wage_proof"
}

type Receipt struct {
	entity.Base
	ProofID    uuid.UUID `gorm:"type:uuid;not null;index:idx_wage_receipt_proof"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index:idx_wage_receipt_employee"`
	Name       string    `gorm:"type:varchar(160);not null"`
	LeafIndex  int       `gorm:"not null"`
	Pay        int64     `gorm:"not null"`
	Days       int       `gorm:"not null"`
	Salt       string    `gorm:"type:varchar(70);not null"`
	Token      string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_wage_receipt_token"`
}

func (Receipt) TableName() string {
	return "wage_receipt"
}

type CircuitKey struct {
	entity.Base
	Name      string `gorm:"type:varchar(40);not null;uniqueIndex:uq_wage_circuit_key"`
	Proving   []byte `gorm:"type:bytea;not null"`
	Verifying []byte `gorm:"type:bytea;not null"`
}

func (CircuitKey) TableName() string {
	return "wage_circuit_key"
}

func Entities() []any {
	return []any{&Proof{}, &Receipt{}, &CircuitKey{}}
}

func Indexes() []string {
	return nil
}

func Constraints() []string {
	return []string{
		database.ForeignKey("wage_proof", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("wage_proof", "created_by", "users", database.DeleteSetNull),
		database.ForeignKey("wage_receipt", "proof_id", "wage_proof", database.DeleteCascade),
		database.ForeignKey("wage_receipt", "employee_id", "employee", database.DeleteCascade),
	}
}
