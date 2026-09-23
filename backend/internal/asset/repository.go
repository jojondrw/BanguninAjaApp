package asset

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type AssetFilter struct {
	Search   string
	Category string
	Status   string
	Offset   int
	Limit    int
}

type EquipmentFilter struct {
	Search           string
	Status           string
	ProjectID        *uuid.UUID
	AssetID          *uuid.UUID
	ServiceDueBefore *time.Time
	Offset           int
	Limit            int
}

type AssetTotals struct {
	AcquisitionValue        int64
	AccumulatedDepreciation int64
}

type Repository interface {
	ListAssets(ctx context.Context, filter AssetFilter) ([]Asset, int64, error)
	SumAssets(ctx context.Context, filter AssetFilter) (AssetTotals, error)
	FindAsset(ctx context.Context, id uuid.UUID) (Asset, error)
	CreateAsset(ctx context.Context, asset *Asset) error
	SaveAsset(ctx context.Context, asset *Asset) error
	DeleteAsset(ctx context.Context, id uuid.UUID) error

	ListEquipment(ctx context.Context, filter EquipmentFilter) ([]Equipment, int64, error)
	FindEquipment(ctx context.Context, id uuid.UUID) (Equipment, error)
	CreateEquipment(ctx context.Context, equipment *Equipment) error
	SaveEquipment(ctx context.Context, equipment *Equipment) error
	DeleteEquipment(ctx context.Context, id uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListAssets(ctx context.Context, filter AssetFilter) ([]Asset, int64, error) {
	return database.FindPage[Asset](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) SumAssets(ctx context.Context, filter AssetFilter) (AssetTotals, error) {
	var totals AssetTotals
	err := r.db.WithContext(ctx).
		Model(&Asset{}).
		Scopes(filter.apply).
		Select("COALESCE(SUM(acquisition_value), 0) AS acquisition_value, " +
			"COALESCE(SUM(accumulated_depreciation), 0) AS accumulated_depreciation").
		Scan(&totals).Error
	return totals, database.Translate(err)
}

func (r *gormRepository) FindAsset(ctx context.Context, id uuid.UUID) (Asset, error) {
	var asset Asset
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&asset).Error
	return asset, database.Translate(err)
}

func (r *gormRepository) CreateAsset(ctx context.Context, asset *Asset) error {
	return database.Translate(r.db.WithContext(ctx).Create(asset).Error)
}

func (r *gormRepository) SaveAsset(ctx context.Context, asset *Asset) error {
	return database.Translate(r.db.WithContext(ctx).Save(asset).Error)
}

func (r *gormRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Asset{}, id)
}

func (r *gormRepository) ListEquipment(ctx context.Context, filter EquipmentFilter) ([]Equipment, int64, error) {
	return database.FindPage[Equipment](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindEquipment(ctx context.Context, id uuid.UUID) (Equipment, error) {
	var equipment Equipment
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&equipment).Error
	return equipment, database.Translate(err)
}

func (r *gormRepository) CreateEquipment(ctx context.Context, equipment *Equipment) error {
	return database.Translate(r.db.WithContext(ctx).Create(equipment).Error)
}

func (r *gormRepository) SaveEquipment(ctx context.Context, equipment *Equipment) error {
	return database.Translate(r.db.WithContext(ctx).Save(equipment).Error)
}

func (r *gormRepository) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Equipment{}, id)
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

func (f AssetFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	return db
}

func (f EquipmentFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	if f.AssetID != nil {
		db = db.Where("asset_id = ?", *f.AssetID)
	}
	if f.ServiceDueBefore != nil {
		db = db.Where("next_service_date <= ?", *f.ServiceDueBefore)
	}
	return db
}
