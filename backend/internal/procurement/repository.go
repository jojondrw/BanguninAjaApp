package procurement

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	purchaseRequestColumns = "purchase_request.*, " +
		"(SELECT COUNT(*) FROM purchase_request_item WHERE purchase_request_item.purchase_request_id = purchase_request.id) AS item_count"
	goodsReceiptColumns = "goods_receipt.*, " +
		"(SELECT COUNT(*) FROM goods_receipt_item WHERE goods_receipt_item.goods_receipt_id = goods_receipt.id) AS item_count"
)

type VendorFilter struct {
	Search   string
	Category string
	Rating   string
	Active   *bool
	Offset   int
	Limit    int
}

type PurchaseRequestFilter struct {
	Search    string
	ProjectID *uuid.UUID
	Status    string
	DateFrom  *time.Time
	DateTo    *time.Time
	Offset    int
	Limit     int
}

type PurchaseOrderFilter struct {
	Search    string
	VendorID  *uuid.UUID
	ProjectID *uuid.UUID
	Status    string
	DateFrom  *time.Time
	DateTo    *time.Time
	Offset    int
	Limit     int
}

type GoodsReceiptFilter struct {
	Search          string
	PurchaseOrderID *uuid.UUID
	WarehouseID     *uuid.UUID
	DateFrom        *time.Time
	DateTo          *time.Time
	Offset          int
	Limit           int
}

type PurchaseRequestRow struct {
	PurchaseRequest
	ItemCount int
}

type GoodsReceiptRow struct {
	GoodsReceipt
	ItemCount int
}

type ReceivedQuantity struct {
	PurchaseOrderItemID uuid.UUID
	Quantity            float64
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListVendors(ctx context.Context, filter VendorFilter) ([]Vendor, int64, error)
	FindVendor(ctx context.Context, id uuid.UUID) (Vendor, error)
	CreateVendor(ctx context.Context, vendor *Vendor) error
	SaveVendor(ctx context.Context, vendor *Vendor) error
	DeleteVendor(ctx context.Context, id uuid.UUID) error

	ListPurchaseRequests(ctx context.Context, filter PurchaseRequestFilter) ([]PurchaseRequestRow, int64, error)
	FindPurchaseRequest(ctx context.Context, id uuid.UUID) (PurchaseRequest, error)
	LockPurchaseRequest(ctx context.Context, id uuid.UUID) (PurchaseRequest, error)
	CreatePurchaseRequest(ctx context.Context, request *PurchaseRequest) error
	SavePurchaseRequest(ctx context.Context, request *PurchaseRequest) error
	DeletePurchaseRequest(ctx context.Context, id uuid.UUID) error
	ListPurchaseRequestItems(ctx context.Context, requestID uuid.UUID) ([]PurchaseRequestItem, error)
	ReplacePurchaseRequestItems(ctx context.Context, requestID uuid.UUID, items []PurchaseRequestItem) error

	ListPurchaseOrders(ctx context.Context, filter PurchaseOrderFilter) ([]PurchaseOrder, int64, error)
	FindPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrder, error)
	LockPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrder, error)
	CreatePurchaseOrder(ctx context.Context, order *PurchaseOrder) error
	SavePurchaseOrder(ctx context.Context, order *PurchaseOrder) error
	DeletePurchaseOrder(ctx context.Context, id uuid.UUID) error
	ListPurchaseOrderItems(ctx context.Context, orderID uuid.UUID) ([]PurchaseOrderItem, error)
	ReplacePurchaseOrderItems(ctx context.Context, orderID uuid.UUID, items []PurchaseOrderItem) error
	SumAcceptedQuantities(ctx context.Context, orderID uuid.UUID) ([]ReceivedQuantity, error)

	ListGoodsReceipts(ctx context.Context, filter GoodsReceiptFilter) ([]GoodsReceiptRow, int64, error)
	FindGoodsReceipt(ctx context.Context, id uuid.UUID) (GoodsReceipt, error)
	CreateGoodsReceipt(ctx context.Context, receipt *GoodsReceipt) error
	ListGoodsReceiptItems(ctx context.Context, receiptID uuid.UUID) ([]GoodsReceiptItem, error)
	CreateGoodsReceiptItems(ctx context.Context, items []GoodsReceiptItem) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Transaction(ctx context.Context, work func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(&gormRepository{db: tx})
	})
}

func (r *gormRepository) ListVendors(ctx context.Context, filter VendorFilter) ([]Vendor, int64, error) {
	return database.FindPage[Vendor](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "name ASC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindVendor(ctx context.Context, id uuid.UUID) (Vendor, error) {
	var vendor Vendor
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&vendor).Error
	return vendor, database.Translate(err)
}

func (r *gormRepository) CreateVendor(ctx context.Context, vendor *Vendor) error {
	return database.Translate(r.db.WithContext(ctx).Create(vendor).Error)
}

func (r *gormRepository) SaveVendor(ctx context.Context, vendor *Vendor) error {
	return database.Translate(r.db.WithContext(ctx).Save(vendor).Error)
}

func (r *gormRepository) DeleteVendor(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Vendor{}, id)
}

func (r *gormRepository) ListPurchaseRequests(ctx context.Context, filter PurchaseRequestFilter) ([]PurchaseRequestRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&PurchaseRequest{}).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]PurchaseRequestRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).
		Model(&PurchaseRequest{}).
		Scopes(filter.apply).
		Select(purchaseRequestColumns).
		Order("purchase_request.date DESC, purchase_request.number DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindPurchaseRequest(ctx context.Context, id uuid.UUID) (PurchaseRequest, error) {
	var request PurchaseRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&request).Error
	return request, database.Translate(err)
}

func (r *gormRepository) LockPurchaseRequest(ctx context.Context, id uuid.UUID) (PurchaseRequest, error) {
	var request PurchaseRequest
	err := r.locked(ctx).Where("id = ?", id).Take(&request).Error
	return request, database.Translate(err)
}

func (r *gormRepository) CreatePurchaseRequest(ctx context.Context, request *PurchaseRequest) error {
	return database.Translate(r.db.WithContext(ctx).Create(request).Error)
}

func (r *gormRepository) SavePurchaseRequest(ctx context.Context, request *PurchaseRequest) error {
	return database.Translate(r.db.WithContext(ctx).Save(request).Error)
}

func (r *gormRepository) DeletePurchaseRequest(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &PurchaseRequest{}, id)
}

func (r *gormRepository) ListPurchaseRequestItems(ctx context.Context, requestID uuid.UUID) ([]PurchaseRequestItem, error) {
	var items []PurchaseRequestItem
	err := r.db.WithContext(ctx).Where("purchase_request_id = ?", requestID).Order("created_at ASC, id ASC").Find(&items).Error
	return items, database.Translate(err)
}

func (r *gormRepository) ReplacePurchaseRequestItems(ctx context.Context, requestID uuid.UUID, items []PurchaseRequestItem) error {
	if err := r.db.WithContext(ctx).Where("purchase_request_id = ?", requestID).Delete(&PurchaseRequestItem{}).Error; err != nil {
		return database.Translate(err)
	}
	return database.Translate(r.db.WithContext(ctx).Create(&items).Error)
}

func (r *gormRepository) ListPurchaseOrders(ctx context.Context, filter PurchaseOrderFilter) ([]PurchaseOrder, int64, error) {
	return database.FindPage[PurchaseOrder](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "date DESC, number DESC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	var order PurchaseOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&order).Error
	return order, database.Translate(err)
}

func (r *gormRepository) LockPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	var order PurchaseOrder
	err := r.locked(ctx).Where("id = ?", id).Take(&order).Error
	return order, database.Translate(err)
}

func (r *gormRepository) CreatePurchaseOrder(ctx context.Context, order *PurchaseOrder) error {
	return database.Translate(r.db.WithContext(ctx).Create(order).Error)
}

func (r *gormRepository) SavePurchaseOrder(ctx context.Context, order *PurchaseOrder) error {
	return database.Translate(r.db.WithContext(ctx).Save(order).Error)
}

func (r *gormRepository) DeletePurchaseOrder(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &PurchaseOrder{}, id)
}

func (r *gormRepository) ListPurchaseOrderItems(ctx context.Context, orderID uuid.UUID) ([]PurchaseOrderItem, error) {
	var items []PurchaseOrderItem
	err := r.db.WithContext(ctx).Where("purchase_order_id = ?", orderID).Order("created_at ASC, id ASC").Find(&items).Error
	return items, database.Translate(err)
}

func (r *gormRepository) ReplacePurchaseOrderItems(ctx context.Context, orderID uuid.UUID, items []PurchaseOrderItem) error {
	if err := r.db.WithContext(ctx).Where("purchase_order_id = ?", orderID).Delete(&PurchaseOrderItem{}).Error; err != nil {
		return database.Translate(err)
	}
	return database.Translate(r.db.WithContext(ctx).Create(&items).Error)
}

func (r *gormRepository) SumAcceptedQuantities(ctx context.Context, orderID uuid.UUID) ([]ReceivedQuantity, error) {
	var quantities []ReceivedQuantity
	err := r.db.WithContext(ctx).
		Model(&GoodsReceiptItem{}).
		Select("goods_receipt_item.purchase_order_item_id, SUM(goods_receipt_item.accepted_quantity) AS quantity").
		Joins("JOIN goods_receipt ON goods_receipt.id = goods_receipt_item.goods_receipt_id").
		Where("goods_receipt.purchase_order_id = ?", orderID).
		Group("goods_receipt_item.purchase_order_item_id").
		Scan(&quantities).Error
	return quantities, database.Translate(err)
}

func (r *gormRepository) ListGoodsReceipts(ctx context.Context, filter GoodsReceiptFilter) ([]GoodsReceiptRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&GoodsReceipt{}).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]GoodsReceiptRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).
		Model(&GoodsReceipt{}).
		Scopes(filter.apply).
		Select(goodsReceiptColumns).
		Order("goods_receipt.date DESC, goods_receipt.number DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindGoodsReceipt(ctx context.Context, id uuid.UUID) (GoodsReceipt, error) {
	var receipt GoodsReceipt
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&receipt).Error
	return receipt, database.Translate(err)
}

func (r *gormRepository) CreateGoodsReceipt(ctx context.Context, receipt *GoodsReceipt) error {
	return database.Translate(r.db.WithContext(ctx).Create(receipt).Error)
}

func (r *gormRepository) ListGoodsReceiptItems(ctx context.Context, receiptID uuid.UUID) ([]GoodsReceiptItem, error) {
	var items []GoodsReceiptItem
	err := r.db.WithContext(ctx).Where("goods_receipt_id = ?", receiptID).Order("created_at ASC, id ASC").Find(&items).Error
	return items, database.Translate(err)
}

func (r *gormRepository) CreateGoodsReceiptItems(ctx context.Context, items []GoodsReceiptItem) error {
	return database.Translate(r.db.WithContext(ctx).Create(&items).Error)
}

func (r *gormRepository) locked(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
}

func (r *gormRepository) deleteByID(ctx context.Context, model any, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(model)
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (f VendorFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
	}
	if f.Rating != "" {
		db = db.Where("rating = ?", f.Rating)
	}
	if f.Active != nil {
		db = db.Where("active = ?", *f.Active)
	}
	return db
}

func (f PurchaseRequestFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		db = db.Where("purchase_request.number ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.ProjectID != nil {
		db = db.Where("purchase_request.project_id = ?", *f.ProjectID)
	}
	if f.Status != "" {
		db = db.Where("purchase_request.status = ?", f.Status)
	}
	if f.DateFrom != nil {
		db = db.Where("purchase_request.date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("purchase_request.date <= ?", *f.DateTo)
	}
	return db
}

func (f PurchaseOrderFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		db = db.Where("number ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.VendorID != nil {
		db = db.Where("vendor_id = ?", *f.VendorID)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if f.DateFrom != nil {
		db = db.Where("date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("date <= ?", *f.DateTo)
	}
	return db
}

func (f GoodsReceiptFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		db = db.Where("goods_receipt.number ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.PurchaseOrderID != nil {
		db = db.Where("goods_receipt.purchase_order_id = ?", *f.PurchaseOrderID)
	}
	if f.WarehouseID != nil {
		db = db.Where("goods_receipt.warehouse_id = ?", *f.WarehouseID)
	}
	if f.DateFrom != nil {
		db = db.Where("goods_receipt.date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("goods_receipt.date <= ?", *f.DateTo)
	}
	return db
}
