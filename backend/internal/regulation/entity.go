package regulation

import (
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
)

type Regulation struct {
	entity.Base

	RTRID      string `gorm:"type:varchar(20);not null"`
	RegionCode string `gorm:"type:varchar(20);not null;index"`

	ZoneName    string `gorm:"type:varchar(150);not null"`
	ZoneCode    string `gorm:"type:varchar(50)"`
	SubZoneName string `gorm:"type:varchar(150)"`
	SubZoneCode string `gorm:"type:varchar(50)"`

	District string `gorm:"type:varchar(150)"`
	Village  string `gorm:"type:varchar(150)"`

	Point geo.Point

	KDB float64 `gorm:"not null"`
	KLB float64 `gorm:"not null"`
	KDH float64 `gorm:"not null"`

	GSB       string `gorm:"type:text"`
	MaxHeight string `gorm:"type:text"`

	RegulationNote string `gorm:"type:text"`
	Source         string `gorm:"type:text"`

	IsSimulated bool `gorm:"not null;default:false"`
}

func (Regulation) TableName() string {
	return "regulation"
}

func Entities() []any {
	return []any{&Regulation{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_regulation_region_code ON regulation (region_code)`,
		`CREATE INDEX IF NOT EXISTS idx_regulation_point ON regulation USING GIST (point)`,
	}
}

func Constraints() []string {
	return nil
}
