package scoring

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListDimensions(ctx context.Context) ([]Dimension, error)
	FindDimension(ctx context.Context, id uuid.UUID) (Dimension, error)
	CreateDimension(ctx context.Context, dimension *Dimension) error
	SaveDimension(ctx context.Context, dimension *Dimension) error
	DeleteDimension(ctx context.Context, id uuid.UUID) error
	DimensionWeighted(ctx context.Context, id uuid.UUID) (bool, error)

	ListBuildingProfiles(ctx context.Context) ([]BuildingProfile, error)
	FindBuildingProfile(ctx context.Context, id uuid.UUID) (BuildingProfile, error)
	CreateBuildingProfile(ctx context.Context, profile *BuildingProfile) error
	SaveBuildingProfile(ctx context.Context, profile *BuildingProfile) error
	DeleteBuildingProfile(ctx context.Context, id uuid.UUID) error

	ListWeights(ctx context.Context, profileID uuid.UUID) ([]Weight, error)
	ReplaceWeights(ctx context.Context, profileID uuid.UUID, weights []Weight) error
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

func (r *gormRepository) ListDimensions(ctx context.Context) ([]Dimension, error) {
	var dimensions []Dimension
	err := r.db.WithContext(ctx).Order("sort_order ASC, code ASC").Find(&dimensions).Error
	return dimensions, database.Translate(err)
}

func (r *gormRepository) FindDimension(ctx context.Context, id uuid.UUID) (Dimension, error) {
	var dimension Dimension
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&dimension).Error
	return dimension, database.Translate(err)
}

func (r *gormRepository) CreateDimension(ctx context.Context, dimension *Dimension) error {
	return database.Translate(r.db.WithContext(ctx).Create(dimension).Error)
}

func (r *gormRepository) SaveDimension(ctx context.Context, dimension *Dimension) error {
	return database.Translate(r.db.WithContext(ctx).Save(dimension).Error)
}

func (r *gormRepository) DeleteDimension(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Dimension{}, id)
}

func (r *gormRepository) DimensionWeighted(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS (SELECT 1 FROM weight WHERE dimension_id = ?)", id).
		Scan(&exists).Error
	return exists, database.Translate(err)
}

func (r *gormRepository) ListBuildingProfiles(ctx context.Context) ([]BuildingProfile, error) {
	var profiles []BuildingProfile
	err := r.db.WithContext(ctx).Order("code ASC").Find(&profiles).Error
	return profiles, database.Translate(err)
}

func (r *gormRepository) FindBuildingProfile(ctx context.Context, id uuid.UUID) (BuildingProfile, error) {
	var profile BuildingProfile
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&profile).Error
	return profile, database.Translate(err)
}

func (r *gormRepository) CreateBuildingProfile(ctx context.Context, profile *BuildingProfile) error {
	return database.Translate(r.db.WithContext(ctx).Create(profile).Error)
}

func (r *gormRepository) SaveBuildingProfile(ctx context.Context, profile *BuildingProfile) error {
	return database.Translate(r.db.WithContext(ctx).Save(profile).Error)
}

func (r *gormRepository) DeleteBuildingProfile(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &BuildingProfile{}, id)
}

func (r *gormRepository) ListWeights(ctx context.Context, profileID uuid.UUID) ([]Weight, error) {
	var weights []Weight
	err := r.db.WithContext(ctx).
		Table("weight").
		Select("weight.*").
		Joins("JOIN dimension ON dimension.id = weight.dimension_id").
		Where("weight.building_profile_id = ?", profileID).
		Order("dimension.sort_order ASC, dimension.code ASC").
		Scan(&weights).Error
	return weights, database.Translate(err)
}

func (r *gormRepository) ReplaceWeights(ctx context.Context, profileID uuid.UUID, weights []Weight) error {
	if err := r.db.WithContext(ctx).Where("building_profile_id = ?", profileID).Delete(&Weight{}).Error; err != nil {
		return database.Translate(err)
	}
	return database.Translate(r.db.WithContext(ctx).Create(&weights).Error)
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
