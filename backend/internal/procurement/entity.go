package procurement

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Vendor struct {
	entity.Base
	Code            string `gorm:"type:varchar(20);not null;uniqueIndex:uq_vendor_code"`
	Name            string `gorm:"type:varchar(160);not null"`
	Category        string `gorm:"type:varchar(60);not null;index:idx_vendor_category"`
	Contact         string `gorm:"type:varchar(60)"`
	TaxNumber       string `gorm:"type:varchar(25)"`
	PaymentTermDays int    `gorm:"not null;default:0"`
	Rating          string `gorm:"type:varchar(20);not null;default:'new'"`
	Active          bool   `gorm:"not null;default:true"`
}

func (Vendor) TableName() string {
	return "vendor"
}

type PurchaseRequest struct {
	entity.Base
	Number      string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_purchase_request_number"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index:idx_purchase_request_project"`
	RequesterID uuid.UUID `gorm:"type:uuid;not null;index:idx_purchase_request_pemohon"`
	Date        time.Time `gorm:"type:date;not null;index:idx_purchase_request_date"`
	Status      string    `gorm:"type:varchar(20);not null;index:idx_purchase_request_status"`
	Note        string    `gorm:"type:text"`
}

func (PurchaseRequest) TableName() string {
	return "purchase_request"
}

type PurchaseRequestItem struct {
	entity.Base
	PurchaseRequestID uuid.UUID `gorm:"type:uuid;not null;index:idx_purchase_request_item_purchase_request"`
	MaterialID        uuid.UUID `gorm:"type:uuid;not null;index:idx_purchase_request_item_material"`
	Quantity          float64   `gorm:"type:numeric(14,2);not null;default:0"`
	UnitOfMeasureID   uuid.UUID `gorm:"type:uuid;not null"`
}

func (PurchaseRequestItem) TableName() string {
	return "purchase_request_item"
}

type PurchaseOrder struct {
	entity.Base
	Number            string     `gorm:"type:varchar(40);not null;uniqueIndex:uq_purchase_order_number"`
	VendorID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_purchase_order_vendor"`
	ProjectID         uuid.UUID  `gorm:"type:uuid;not null;index:idx_purchase_order_project"`
	PurchaseRequestID *uuid.UUID `gorm:"type:uuid;index:idx_purchase_order_purchase_request"`
	Date              time.Time  `gorm:"type:date;not null"`
	DueDate           *time.Time `gorm:"type:date"`
	Value             int64      `gorm:"not null;default:0"`
	Status            string     `gorm:"type:varchar(20);not null"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_order"
}

type PurchaseOrderItem struct {
	entity.Base
	PurchaseOrderID uuid.UUID `gorm:"type:uuid;not null;index:idx_purchase_order_item_purchase_order"`
	MaterialID      uuid.UUID `gorm:"type:uuid;not null;index:idx_purchase_order_item_material"`
	Quantity        float64   `gorm:"type:numeric(14,2);not null;default:0"`
	UnitOfMeasureID uuid.UUID `gorm:"type:uuid;not null"`
	UnitPrice       int64     `gorm:"not null;default:0"`
}

func (PurchaseOrderItem) TableName() string {
	return "purchase_order_item"
}

type GoodsReceipt struct {
	entity.Base
	Number          string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_goods_receipt_number"`
	PurchaseOrderID uuid.UUID `gorm:"type:uuid;not null;index:idx_goods_receipt_purchase_order"`
	WarehouseID     uuid.UUID `gorm:"type:uuid;not null;index:idx_goods_receipt_warehouse"`
	Date            time.Time `gorm:"type:date;not null;index:idx_goods_receipt_date"`
	Condition       string    `gorm:"type:varchar(20);not null"`
	Note            string    `gorm:"type:text"`
}

func (GoodsReceipt) TableName() string {
	return "goods_receipt"
}

type GoodsReceiptItem struct {
	entity.Base
	GoodsReceiptID      uuid.UUID `gorm:"type:uuid;not null;index:idx_goods_receipt_item_goods_receipt"`
	PurchaseOrderItemID uuid.UUID `gorm:"type:uuid;not null;index:idx_goods_receipt_item_purchase_order_item"`
	AcceptedQuantity    float64   `gorm:"type:numeric(14,2);not null;default:0"`
	RejectedQuantity    float64   `gorm:"type:numeric(14,2);not null;default:0"`
}

func (GoodsReceiptItem) TableName() string {
	return "goods_receipt_item"
}

func Entities() []any {
	return []any{
		&Vendor{}, &PurchaseRequest{}, &PurchaseRequestItem{},
		&PurchaseOrder{}, &PurchaseOrderItem{}, &GoodsReceipt{}, &GoodsReceiptItem{},
	}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_vendor_name_trgm ON vendor USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_purchase_order_status_date ON purchase_order (status, date DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_purchase_request_status_date ON purchase_request (status, date DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("purchase_request", "project_id", "project", database.DeleteRestrict),
		database.ForeignKey("purchase_request", "requester_id", "users", database.DeleteRestrict),
		database.ForeignKey("purchase_request_item", "purchase_request_id", "purchase_request", database.DeleteCascade),
		database.ForeignKey("purchase_request_item", "material_id", "material", database.DeleteRestrict),
		database.ForeignKey("purchase_request_item", "unit_of_measure_id", "unit_of_measure", database.DeleteRestrict),
		database.ForeignKey("purchase_order", "vendor_id", "vendor", database.DeleteRestrict),
		database.ForeignKey("purchase_order", "project_id", "project", database.DeleteRestrict),
		database.ForeignKey("purchase_order", "purchase_request_id", "purchase_request", database.DeleteSetNull),
		database.ForeignKey("purchase_order_item", "purchase_order_id", "purchase_order", database.DeleteCascade),
		database.ForeignKey("purchase_order_item", "material_id", "material", database.DeleteRestrict),
		database.ForeignKey("purchase_order_item", "unit_of_measure_id", "unit_of_measure", database.DeleteRestrict),
		database.ForeignKey("goods_receipt", "purchase_order_id", "purchase_order", database.DeleteRestrict),
		database.ForeignKey("goods_receipt", "warehouse_id", "warehouse", database.DeleteRestrict),
		database.ForeignKey("goods_receipt_item", "goods_receipt_id", "goods_receipt", database.DeleteCascade),
		database.ForeignKey("goods_receipt_item", "purchase_order_item_id", "purchase_order_item", database.DeleteRestrict),
		database.Check("vendor", "payment_term_days", "payment_term_days >= 0"),
		database.Check("vendor", "rating", "rating IN ('new','good','fair','poor')"),
		database.Check("purchase_request", "status", "status IN ('draft','submitted','approved','rejected','completed')"),
		database.Check("purchase_order", "status", "status IN ('draft','sent','partially_received','completed','cancelled')"),
		database.Check("purchase_order", "value", "value >= 0"),
		database.Check("goods_receipt", "condition", "condition IN ('good','partially_damaged','broken')"),
		database.Check("purchase_request_item", "quantity", "quantity > 0"),
		database.Check("purchase_order_item", "quantity", "quantity > 0"),
		database.Check("goods_receipt_item", "quantity", "accepted_quantity >= 0 AND rejected_quantity >= 0"),
		database.Unique("purchase_request_item", "request_material", "purchase_request_id, material_id"),
		database.Unique("purchase_order_item", "order_material", "purchase_order_id, material_id"),
	}
}
