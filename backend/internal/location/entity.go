package location

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
)

type SavedLocation struct {
	entity.Base
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index:idx_saved_location_user"`
	Name              string     `gorm:"type:varchar(160);not null"`
	RegionID          *uuid.UUID `gorm:"type:uuid;index:idx_saved_location_region"`
	BuildingProfileID *uuid.UUID `gorm:"type:uuid;index:idx_saved_location_building_profile"`
	Point             geo.Point  `gorm:"not null"`
	AreaSqm           float64    `gorm:"type:numeric(14,2);not null;default:0"`
	LandPricePerSqm   int64      `gorm:"not null;default:0"`
	Score             int        `gorm:"not null;default:0;index:idx_saved_location_score"`
	FloodIndex        float64    `gorm:"type:numeric(4,3);not null;default:0"`
	EarthquakeIndex   float64    `gorm:"type:numeric(4,3);not null;default:0"`
	Note              string     `gorm:"type:text"`
	SavedAt           time.Time  `gorm:"not null;default:now()"`
}

func (SavedLocation) TableName() string {
	return "saved_location"
}

type DimensionScore struct {
	entity.Base
	SavedLocationID uuid.UUID `gorm:"type:uuid;not null;index:idx_dimension_score_saved_location"`
	DimensionID     uuid.UUID `gorm:"type:uuid;not null;index:idx_dimension_score_dimension"`
	Value           int       `gorm:"not null;default:0"`
}

func (DimensionScore) TableName() string {
	return "dimension_score"
}

type Comparison struct {
	entity.Base
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_comparison_user"`
	Name   string    `gorm:"type:varchar(160);not null"`
}

func (Comparison) TableName() string {
	return "comparison"
}

type ComparisonItem struct {
	entity.Base
	ComparisonID    uuid.UUID `gorm:"type:uuid;not null;index:idx_comparison_item_comparison"`
	SavedLocationID uuid.UUID `gorm:"type:uuid;not null;index:idx_comparison_item_saved_location"`
	SortOrder       int       `gorm:"not null;default:1"`
}

func (ComparisonItem) TableName() string {
	return "comparison_item"
}

func Entities() []any {
	return []any{&SavedLocation{}, &DimensionScore{}, &Comparison{}, &ComparisonItem{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_saved_location_point ON saved_location USING gist (point)`,
		`CREATE INDEX IF NOT EXISTS idx_saved_location_name_trgm ON saved_location USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_saved_location_user_score ON saved_location (user_id, score DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("saved_location", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("saved_location", "region_id", "region", database.DeleteRestrict),
		database.ForeignKey("saved_location", "building_profile_id", "building_profile", database.DeleteSetNull),
		database.ForeignKey("dimension_score", "saved_location_id", "saved_location", database.DeleteCascade),
		database.ForeignKey("dimension_score", "dimension_id", "dimension", database.DeleteRestrict),
		database.ForeignKey("comparison", "user_id", "users", database.DeleteCascade),
		database.ForeignKey("comparison_item", "comparison_id", "comparison", database.DeleteCascade),
		database.ForeignKey("comparison_item", "saved_location_id", "saved_location", database.DeleteCascade),
		database.Check("saved_location", "score", "score BETWEEN 0 AND 100"),
		database.Check("saved_location", "index", "flood_index BETWEEN 0 AND 1 AND earthquake_index BETWEEN 0 AND 1"),
		database.Check("saved_location", "area", "area_sqm >= 0"),
		database.Check("saved_location", "price", "land_price_per_sqm >= 0"),
		database.Check("dimension_score", "value", "value BETWEEN 0 AND 100"),
		database.Unique("dimension_score", "location_dimension", "saved_location_id, dimension_id"),
		database.Unique("comparison_item", "comparison_location", "comparison_id, saved_location_id"),
	}
}
