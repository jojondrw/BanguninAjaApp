package asset

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Asset struct {
	entity.Base
	Code                    string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_asset_code"`
	Name                    string     `gorm:"type:varchar(160);not null"`
	Category                string     `gorm:"type:varchar(60);not null;index:idx_asset_category"`
	AcquisitionDate         *time.Time `gorm:"type:date"`
	AcquisitionValue        int64      `gorm:"not null;default:0"`
	AccumulatedDepreciation int64      `gorm:"not null;default:0"`
	UsefulLifeMonths        int        `gorm:"not null;default:0"`
	Status                  string     `gorm:"type:varchar(20);not null;default:'in_use'"`
}

func (Asset) TableName() string {
	return "asset"
}

type Equipment struct {
	entity.Base
	Code            string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_equipment_code"`
	Name            string     `gorm:"type:varchar(160);not null"`
	AssetID         *uuid.UUID `gorm:"type:uuid;index:idx_equipment_asset"`
	ProjectID       *uuid.UUID `gorm:"type:uuid;index:idx_equipment_project"`
	OperatingHours  int        `gorm:"not null;default:0"`
	NextServiceDate *time.Time `gorm:"type:date;index:idx_equipment_service"`
	Status          string     `gorm:"type:varchar(20);not null;index:idx_equipment_status"`
}

func (Equipment) TableName() string {
	return "equipment"
}

func Entities() []any {
	return []any{&Asset{}, &Equipment{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_asset_name_trgm ON asset USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_equipment_name_trgm ON equipment USING gin (name gin_trgm_ops)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("equipment", "asset_id", "asset", database.DeleteSetNull),
		database.ForeignKey("equipment", "project_id", "project", database.DeleteSetNull),
		database.Check("asset", "value", "acquisition_value >= 0 AND accumulated_depreciation >= 0"),
		database.Check("asset", "depreciation_within_cost", "accumulated_depreciation <= acquisition_value"),
		database.Check("asset", "status", "status IN ('in_use','maintenance','for_sale','disposed')"),
		database.Check("equipment", "operating_hours", "operating_hours >= 0"),
		database.Check("equipment", "status", "status IN ('operating','idle','maintenance','broken')"),
	}
}
