package regulation

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	GetByPoint(ctx context.Context, latitude, longitude float64) (*Regulation, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) GetByPoint(ctx context.Context, latitude, longitude float64) (*Regulation, error) {
	var regulation Regulation

	point := "ST_SetSRID(ST_MakePoint(?, ?), 4326)"

	err := r.db.WithContext(ctx).
		Where(
			"ST_DWithin(point::geography, "+point+"::geography, ?)",
			longitude,
			latitude,
			150,
		).
		Order(gorm.Expr(
			"ST_Distance(point::geography, "+point+"::geography)",
			longitude,
			latitude,
		)).
		First(&regulation).Error

	if err != nil {
		return nil, err
	}

	return &regulation, nil
}
