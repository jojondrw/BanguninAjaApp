package inventory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	lowStockCondition = "m.minimum_stock > 0 AND COALESCE(s.total, 0) < m.minimum_stock"
	lowStockSource    = "material m LEFT JOIN (SELECT material_id, SUM(quantity) AS total FROM stock GROUP BY material_id) s ON s.material_id = m.id"
)

type MaterialFilter struct {
	Search   string
	Category string
	Offset   int
	Limit    int
}

type WarehouseFilter struct {
	Search    string
	ProjectID *uuid.UUID
	Offset    int
	Limit     int
}

type StockFilter struct {
	MaterialID  *uuid.UUID
	WarehouseID *uuid.UUID
	Offset      int
	Limit       int
}

type StockMovementFilter struct {
	MaterialID  *uuid.UUID
	WarehouseID *uuid.UUID
	Type        string
	DateFrom    *time.Time
	DateTo      *time.Time
	Offset      int
	Limit       int
}

type StockLevel struct {
	MaterialID      uuid.UUID
	Code            string
	Name            string
	UnitOfMeasureID uuid.UUID
	MinimumStock    float64
	TotalQuantity   float64
}

type StockRow struct {
	ID            uuid.UUID
	MaterialID    uuid.UUID
	MaterialCode  string
	MaterialName  string
	WarehouseID   uuid.UUID
	WarehouseCode string
	WarehouseName string
	Quantity      float64
	UpdatedAt     time.Time
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListMaterials(ctx context.Context, filter MaterialFilter) ([]Material, int64, error)
	ListLowStockMaterials(ctx context.Context, offset, limit int) ([]StockLevel, int64, error)
	FindMaterial(ctx context.Context, id uuid.UUID) (Material, error)
	CreateMaterial(ctx context.Context, material *Material) error
	SaveMaterial(ctx context.Context, material *Material) error
	DeleteMaterial(ctx context.Context, id uuid.UUID) error

	ListWarehouses(ctx context.Context, filter WarehouseFilter) ([]Warehouse, int64, error)
	FindWarehouse(ctx context.Context, id uuid.UUID) (Warehouse, error)
	CreateWarehouse(ctx context.Context, warehouse *Warehouse) error
	SaveWarehouse(ctx context.Context, warehouse *Warehouse) error
	DeleteWarehouse(ctx context.Context, id uuid.UUID) error

	ListStocks(ctx context.Context, filter StockFilter) ([]StockRow, int64, error)
	IncreaseStock(ctx context.Context, materialID, warehouseID uuid.UUID, quantity float64) error
	DecreaseStock(ctx context.Context, materialID, warehouseID uuid.UUID, quantity float64) (bool, error)

	ListStockMovements(ctx context.Context, filter StockMovementFilter) ([]StockMovement, int64, error)
	FindStockMovement(ctx context.Context, id uuid.UUID) (StockMovement, error)
	CreateStockMovement(ctx context.Context, movement *StockMovement) error
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

func (r *gormRepository) ListMaterials(ctx context.Context, filter MaterialFilter) ([]Material, int64, error) {
	return database.FindPage[Material](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) ListLowStockMaterials(ctx context.Context, offset, limit int) ([]StockLevel, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Table(lowStockSource).Where(lowStockCondition).Count(&total).Error
	if err != nil {
		return nil, 0, database.Translate(err)
	}

	levels := make([]StockLevel, 0, limit)
	err = r.db.WithContext(ctx).
		Table(lowStockSource).
		Select("m.id AS material_id, m.code, m.name, m.unit_of_measure_id, m.minimum_stock, COALESCE(s.total, 0) AS total_quantity").
		Where(lowStockCondition).
		Order("m.code ASC").
		Offset(offset).
		Limit(limit).
		Scan(&levels).Error
	return levels, total, database.Translate(err)
}

func (r *gormRepository) FindMaterial(ctx context.Context, id uuid.UUID) (Material, error) {
	var material Material
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&material).Error
	return material, database.Translate(err)
}

func (r *gormRepository) CreateMaterial(ctx context.Context, material *Material) error {
	return database.Translate(r.db.WithContext(ctx).Create(material).Error)
}

func (r *gormRepository) SaveMaterial(ctx context.Context, material *Material) error {
	return database.Translate(r.db.WithContext(ctx).Save(material).Error)
}

func (r *gormRepository) DeleteMaterial(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Material{}, id)
}

func (r *gormRepository) ListWarehouses(ctx context.Context, filter WarehouseFilter) ([]Warehouse, int64, error) {
	return database.FindPage[Warehouse](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindWarehouse(ctx context.Context, id uuid.UUID) (Warehouse, error) {
	var warehouse Warehouse
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&warehouse).Error
	return warehouse, database.Translate(err)
}

func (r *gormRepository) CreateWarehouse(ctx context.Context, warehouse *Warehouse) error {
	return database.Translate(r.db.WithContext(ctx).Create(warehouse).Error)
}

func (r *gormRepository) SaveWarehouse(ctx context.Context, warehouse *Warehouse) error {
	return database.Translate(r.db.WithContext(ctx).Save(warehouse).Error)
}

func (r *gormRepository) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Warehouse{}, id)
}

func (r *gormRepository) ListStocks(ctx context.Context, filter StockFilter) ([]StockRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Stock{}).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]StockRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).
		Model(&Stock{}).
		Scopes(filter.apply).
		Select("stock.id, stock.material_id, material.code AS material_code, material.name AS material_name, " +
			"stock.warehouse_id, warehouse.code AS warehouse_code, warehouse.name AS warehouse_name, stock.quantity, stock.updated_at").
		Joins("JOIN material ON material.id = stock.material_id").
		Joins("JOIN warehouse ON warehouse.id = stock.warehouse_id").
		Order("material.code ASC, warehouse.code ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) IncreaseStock(ctx context.Context, materialID, warehouseID uuid.UUID, quantity float64) error {
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO stock (id, material_id, warehouse_id, quantity, created_at, updated_at)
		VALUES (?, ?, ?, ?, now(), now())
		ON CONFLICT (material_id, warehouse_id)
		DO UPDATE SET quantity = stock.quantity + EXCLUDED.quantity, updated_at = now()`,
		uuid.New(), materialID, warehouseID, quantity).Error
	return database.Translate(err)
}

func (r *gormRepository) DecreaseStock(ctx context.Context, materialID, warehouseID uuid.UUID, quantity float64) (bool, error) {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE stock SET quantity = quantity - ?, updated_at = now()
		WHERE material_id = ? AND warehouse_id = ? AND quantity >= ?`,
		quantity, materialID, warehouseID, quantity)
	if result.Error != nil {
		return false, database.Translate(result.Error)
	}
	return result.RowsAffected > 0, nil
}

func (r *gormRepository) ListStockMovements(ctx context.Context, filter StockMovementFilter) ([]StockMovement, int64, error) {
	return database.FindPage[StockMovement](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "date DESC, created_at DESC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindStockMovement(ctx context.Context, id uuid.UUID) (StockMovement, error) {
	var movement StockMovement
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&movement).Error
	return movement, database.Translate(err)
}

func (r *gormRepository) CreateStockMovement(ctx context.Context, movement *StockMovement) error {
	return database.Translate(r.db.WithContext(ctx).Create(movement).Error)
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

func (f MaterialFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
	}
	return db
}

func (f WarehouseFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	return db
}

func (f StockFilter) apply(db *gorm.DB) *gorm.DB {
	if f.MaterialID != nil {
		db = db.Where("stock.material_id = ?", *f.MaterialID)
	}
	if f.WarehouseID != nil {
		db = db.Where("stock.warehouse_id = ?", *f.WarehouseID)
	}
	return db
}

func (f StockMovementFilter) apply(db *gorm.DB) *gorm.DB {
	if f.MaterialID != nil {
		db = db.Where("material_id = ?", *f.MaterialID)
	}
	if f.WarehouseID != nil {
		db = db.Where("(source_warehouse_id = ? OR target_warehouse_id = ?)", *f.WarehouseID, *f.WarehouseID)
	}
	if f.Type != "" {
		db = db.Where("type = ?", f.Type)
	}
	if f.DateFrom != nil {
		db = db.Where("date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("date <= ?", *f.DateTo)
	}
	return db
}
