package site

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/scoring"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

// SavedRecord is what a single evaluation persists: the saved location plus the
// per-dimension breakdown returned by the scoring service.
type SavedRecord struct {
	Location location.SavedLocation
	Scores   []location.DimensionScore
}

// Repository owns the reads/writes T4 needs across the location, scoring and
// regulation tables. It reuses the existing entity structs so the schema stays
// the single source of truth.
type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	FindBuildingProfile(ctx context.Context, id uuid.UUID) (scoring.BuildingProfile, error)
	DimensionIDByCode(ctx context.Context) (map[string]uuid.UUID, error)
	ProjectExists(ctx context.Context, id uuid.UUID) (bool, error)

	CreateSavedLocation(ctx context.Context, saved *SavedRecord) error
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

func (r *gormRepository) FindBuildingProfile(ctx context.Context, id uuid.UUID) (scoring.BuildingProfile, error) {
	var profile scoring.BuildingProfile
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&profile).Error
	return profile, database.Translate(err)
}

func (r *gormRepository) DimensionIDByCode(ctx context.Context) (map[string]uuid.UUID, error) {
	var dimensions []scoring.Dimension
	if err := r.db.WithContext(ctx).Find(&dimensions).Error; err != nil {
		return nil, database.Translate(err)
	}

	byCode := make(map[string]uuid.UUID, len(dimensions))
	for _, dimension := range dimensions {
		byCode[dimension.Code] = dimension.ID
	}
	return byCode, nil
}

func (r *gormRepository) ProjectExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS (SELECT 1 FROM project WHERE id = ?)", id).
		Scan(&exists).Error
	return exists, database.Translate(err)
}

func (r *gormRepository) CreateSavedLocation(ctx context.Context, saved *SavedRecord) error {
	if saved.Location.SavedAt.IsZero() {
		saved.Location.SavedAt = time.Now()
	}
	if err := r.db.WithContext(ctx).Create(&saved.Location).Error; err != nil {
		return database.Translate(err)
	}

	if len(saved.Scores) == 0 {
		return nil
	}
	for index := range saved.Scores {
		saved.Scores[index].SavedLocationID = saved.Location.ID
	}
	return database.Translate(r.db.WithContext(ctx).Create(&saved.Scores).Error)
}
