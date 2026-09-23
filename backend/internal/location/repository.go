package location

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const wgs84 = 4326

type Bounds struct {
	MinLongitude float64
	MinLatitude  float64
	MaxLongitude float64
	MaxLatitude  float64
}

type SavedLocationFilter struct {
	UserID            uuid.UUID
	Search            string
	RegionID          *uuid.UUID
	BuildingProfileID *uuid.UUID
	MinScore          *int
	Bounds            *Bounds
	Offset            int
	Limit             int
}

type ComparisonRow struct {
	ID            uuid.UUID
	Name          string
	LocationCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListSavedLocations(ctx context.Context, filter SavedLocationFilter) ([]SavedLocation, int64, error)
	FindSavedLocation(ctx context.Context, userID, id uuid.UUID) (SavedLocation, error)
	FindSavedLocations(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) ([]SavedLocation, error)
	CreateSavedLocation(ctx context.Context, location *SavedLocation) error
	SaveSavedLocation(ctx context.Context, location *SavedLocation) error
	DeleteSavedLocation(ctx context.Context, userID, id uuid.UUID) error

	ListDimensionScores(ctx context.Context, locationIDs []uuid.UUID) ([]DimensionScore, error)
	ReplaceDimensionScores(ctx context.Context, locationID uuid.UUID, scores []DimensionScore) error

	ListComparisons(ctx context.Context, userID uuid.UUID, offset, limit int) ([]ComparisonRow, int64, error)
	FindComparison(ctx context.Context, userID, id uuid.UUID) (Comparison, error)
	CreateComparison(ctx context.Context, comparison *Comparison) error
	SaveComparison(ctx context.Context, comparison *Comparison) error
	DeleteComparison(ctx context.Context, userID, id uuid.UUID) error

	ListComparisonItems(ctx context.Context, comparisonID uuid.UUID) ([]ComparisonItem, error)
	ReplaceComparisonItems(ctx context.Context, comparisonID uuid.UUID, items []ComparisonItem) error
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

func (r *gormRepository) ListSavedLocations(ctx context.Context, filter SavedLocationFilter) ([]SavedLocation, int64, error) {
	return database.FindPage[SavedLocation](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "score DESC, saved_at DESC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindSavedLocation(ctx context.Context, userID, id uuid.UUID) (SavedLocation, error) {
	var location SavedLocation
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Take(&location).Error
	return location, database.Translate(err)
}

func (r *gormRepository) FindSavedLocations(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) ([]SavedLocation, error) {
	var locations []SavedLocation
	err := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Find(&locations).Error
	return locations, database.Translate(err)
}

func (r *gormRepository) CreateSavedLocation(ctx context.Context, location *SavedLocation) error {
	return database.Translate(r.db.WithContext(ctx).Create(location).Error)
}

func (r *gormRepository) SaveSavedLocation(ctx context.Context, location *SavedLocation) error {
	return database.Translate(r.db.WithContext(ctx).Save(location).Error)
}

func (r *gormRepository) DeleteSavedLocation(ctx context.Context, userID, id uuid.UUID) error {
	return r.deleteOwned(ctx, &SavedLocation{}, userID, id)
}

func (r *gormRepository) ListDimensionScores(ctx context.Context, locationIDs []uuid.UUID) ([]DimensionScore, error) {
	var scores []DimensionScore
	err := r.db.WithContext(ctx).
		Where("saved_location_id IN ?", locationIDs).
		Order("saved_location_id ASC, created_at ASC").
		Find(&scores).Error
	return scores, database.Translate(err)
}

func (r *gormRepository) ReplaceDimensionScores(ctx context.Context, locationID uuid.UUID, scores []DimensionScore) error {
	if err := r.db.WithContext(ctx).Where("saved_location_id = ?", locationID).Delete(&DimensionScore{}).Error; err != nil {
		return database.Translate(err)
	}
	if len(scores) == 0 {
		return nil
	}
	return database.Translate(r.db.WithContext(ctx).Create(&scores).Error)
}

func (r *gormRepository) ListComparisons(ctx context.Context, userID uuid.UUID, offset, limit int) ([]ComparisonRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Comparison{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]ComparisonRow, 0, limit)
	err := r.db.WithContext(ctx).
		Model(&Comparison{}).
		Select("comparison.id, comparison.name, comparison.created_at, comparison.updated_at, "+
			"(SELECT COUNT(*) FROM comparison_item WHERE comparison_item.comparison_id = comparison.id) AS location_count").
		Where("comparison.user_id = ?", userID).
		Order("comparison.created_at DESC, comparison.id ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindComparison(ctx context.Context, userID, id uuid.UUID) (Comparison, error) {
	var comparison Comparison
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Take(&comparison).Error
	return comparison, database.Translate(err)
}

func (r *gormRepository) CreateComparison(ctx context.Context, comparison *Comparison) error {
	return database.Translate(r.db.WithContext(ctx).Create(comparison).Error)
}

func (r *gormRepository) SaveComparison(ctx context.Context, comparison *Comparison) error {
	return database.Translate(r.db.WithContext(ctx).Save(comparison).Error)
}

func (r *gormRepository) DeleteComparison(ctx context.Context, userID, id uuid.UUID) error {
	return r.deleteOwned(ctx, &Comparison{}, userID, id)
}

func (r *gormRepository) ListComparisonItems(ctx context.Context, comparisonID uuid.UUID) ([]ComparisonItem, error) {
	var items []ComparisonItem
	err := r.db.WithContext(ctx).
		Where("comparison_id = ?", comparisonID).
		Order("sort_order ASC").
		Find(&items).Error
	return items, database.Translate(err)
}

func (r *gormRepository) ReplaceComparisonItems(ctx context.Context, comparisonID uuid.UUID, items []ComparisonItem) error {
	if err := r.db.WithContext(ctx).Where("comparison_id = ?", comparisonID).Delete(&ComparisonItem{}).Error; err != nil {
		return database.Translate(err)
	}
	return database.Translate(r.db.WithContext(ctx).Create(&items).Error)
}

func (r *gormRepository) deleteOwned(ctx context.Context, model any, userID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(model)
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (f SavedLocationFilter) apply(db *gorm.DB) *gorm.DB {
	db = db.Where("user_id = ?", f.UserID)
	if f.Search != "" {
		db = db.Where("name ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.RegionID != nil {
		db = db.Where("region_id = ?", *f.RegionID)
	}
	if f.BuildingProfileID != nil {
		db = db.Where("building_profile_id = ?", *f.BuildingProfileID)
	}
	if f.MinScore != nil {
		db = db.Where("score >= ?", *f.MinScore)
	}
	if f.Bounds != nil {
		db = db.Where("point && ST_MakeEnvelope(?, ?, ?, ?, ?)",
			f.Bounds.MinLongitude, f.Bounds.MinLatitude, f.Bounds.MaxLongitude, f.Bounds.MaxLatitude, wgs84)
	}
	return db
}
