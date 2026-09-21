package scoring

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Dimension struct {
	entity.Base
	Code        string `gorm:"type:varchar(30);not null;uniqueIndex:uq_dimension_code"`
	Name        string `gorm:"type:varchar(120);not null"`
	Description string `gorm:"type:text"`
	SortOrder   int    `gorm:"not null;default:1"`
}

func (Dimension) TableName() string {
	return "dimension"
}

type BuildingProfile struct {
	entity.Base
	Code string `gorm:"type:varchar(30);not null;uniqueIndex:uq_building_profile_code"`
	Name string `gorm:"type:varchar(120);not null"`
}

func (BuildingProfile) TableName() string {
	return "building_profile"
}

type Weight struct {
	entity.Base
	BuildingProfileID uuid.UUID `gorm:"type:uuid;not null;index:idx_weight_profile"`
	DimensionID       uuid.UUID `gorm:"type:uuid;not null;index:idx_weight_dimension"`
	Percent           int       `gorm:"not null;default:0"`
}

func (Weight) TableName() string {
	return "weight"
}

func Entities() []any {
	return []any{&Dimension{}, &BuildingProfile{}, &Weight{}}
}

func Indexes() []string {
	return []string{}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("weight", "building_profile_id", "building_profile", database.DeleteCascade),
		database.ForeignKey("weight", "dimension_id", "dimension", database.DeleteCascade),
		database.Check("weight", "percent", "percent BETWEEN 0 AND 100"),
		database.Unique("weight", "profile_dimension", "building_profile_id, dimension_id"),
	}
}
