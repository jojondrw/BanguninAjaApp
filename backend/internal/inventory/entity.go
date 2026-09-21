package inventory

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Material struct {
	entity.Base
	Code            string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_material_code"`
	Name            string    `gorm:"type:varchar(160);not null"`
	Category        string    `gorm:"type:varchar(60);index:idx_material_category"`
	UnitOfMeasureID uuid.UUID `gorm:"type:uuid;not null;index:idx_material_unit_of_measure"`
	MinimumStock    float64   `gorm:"type:numeric(14,2);not null;default:0"`
	LastPrice       int64     `gorm:"not null;default:0"`
}

func (Material) TableName() string {
	return "material"
}

type Warehouse struct {
	entity.Base
	Code      string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_warehouse_code"`
	Name      string     `gorm:"type:varchar(120);not null"`
	ProjectID *uuid.UUID `gorm:"type:uuid;index:idx_warehouse_project"`
}

func (Warehouse) TableName() string {
	return "warehouse"
}

type Stock struct {
	entity.Base
	MaterialID  uuid.UUID `gorm:"type:uuid;not null;index:idx_stock_material"`
	WarehouseID uuid.UUID `gorm:"type:uuid;not null;index:idx_stock_warehouse"`
	Quantity    float64   `gorm:"type:numeric(14,2);not null;default:0"`
}

func (Stock) TableName() string {
	return "stock"
}

type StockMovement struct {
	entity.Base
	Date              time.Time  `gorm:"type:date;not null;index:idx_movement_stock_date"`
	Type              string     `gorm:"type:varchar(20);not null;index:idx_movement_stock_type"`
	MaterialID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_movement_stock_material"`
	Quantity          float64    `gorm:"type:numeric(14,2);not null"`
	SourceWarehouseID *uuid.UUID `gorm:"type:uuid;index:idx_movement_stock_warehouse_asal"`
	TargetWarehouseID *uuid.UUID `gorm:"type:uuid;index:idx_movement_stock_warehouse_tujuan"`
	Reference         string     `gorm:"type:varchar(60)"`
}

func (StockMovement) TableName() string {
	return "stock_movement"
}

func Entities() []any {
	return []any{&Material{}, &Warehouse{}, &Stock{}, &StockMovement{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_material_name_trgm ON material USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_movement_stock_material_date ON stock_movement (material_id, date DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_low ON stock (material_id) WHERE quantity <= 0`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("material", "unit_of_measure_id", "unit_of_measure", database.DeleteRestrict),
		database.ForeignKey("warehouse", "project_id", "project", database.DeleteSetNull),
		database.ForeignKey("stock", "material_id", "material", database.DeleteCascade),
		database.ForeignKey("stock", "warehouse_id", "warehouse", database.DeleteCascade),
		database.ForeignKey("stock_movement", "material_id", "material", database.DeleteRestrict),
		database.ForeignKey("stock_movement", "source_warehouse_id", "warehouse", database.DeleteRestrict),
		database.ForeignKey("stock_movement", "target_warehouse_id", "warehouse", database.DeleteRestrict),
		database.Check("material", "minimum_stock", "minimum_stock >= 0"),
		database.Check("stock", "quantity", "quantity >= 0"),
		database.Check("stock_movement", "quantity", "quantity > 0"),
		database.Check("stock_movement", "type", "type IN ('in','out','transfer','adjustment')"),
		database.Check("stock_movement", "direction", "source_warehouse_id IS NOT NULL OR target_warehouse_id IS NOT NULL"),
		database.Unique("stock", "material_warehouse", "material_id, warehouse_id"),
	}
}
