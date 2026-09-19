package master

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Region struct {
	entity.Base
	Code     string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_region_code"`
	Name     string     `gorm:"type:varchar(120);not null;index:idx_region_name"`
	Type     string     `gorm:"type:varchar(20);not null;index:idx_region_type"`
	ParentID *uuid.UUID `gorm:"type:uuid;index:idx_region_parent"`
}

func (Region) TableName() string {
	return "region"
}

type UnitOfMeasure struct {
	entity.Base
	Code string `gorm:"type:varchar(20);not null;uniqueIndex:uq_unit_of_measure_code"`
	Name string `gorm:"type:varchar(60);not null"`
}

func (UnitOfMeasure) TableName() string {
	return "unit_of_measure"
}

type Account struct {
	entity.Base
	Code     string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_account_code"`
	Name     string     `gorm:"type:varchar(120);not null;index:idx_account_name"`
	Type     string     `gorm:"type:varchar(20);not null;index:idx_account_type"`
	ParentID *uuid.UUID `gorm:"type:uuid;index:idx_account_parent"`
}

func (Account) TableName() string {
	return "account"
}

func Entities() []any {
	return []any{&Region{}, &UnitOfMeasure{}, &Account{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_region_name_trgm ON region USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_account_name_trgm ON account USING gin (name gin_trgm_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("region", "parent_id", "region", database.DeleteRestrict),
		database.ForeignKey("account", "parent_id", "account", database.DeleteRestrict),
		database.Check("region", "type", "type IN ('province','city','regency','district')"),
		database.Check("account", "type", "type IN ('asset','liability','equity','revenue','expense')"),
	}
}
