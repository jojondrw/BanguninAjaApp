package master

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type RegionFilter struct {
	Search   string
	Type     string
	ParentID *uuid.UUID
	Offset   int
	Limit    int
}

type UnitOfMeasureFilter struct {
	Search string
	Offset int
	Limit  int
}

type AccountFilter struct {
	Search   string
	Type     string
	ParentID *uuid.UUID
	Offset   int
	Limit    int
}

type Repository interface {
	ListRegions(ctx context.Context, filter RegionFilter) ([]Region, int64, error)
	FindRegion(ctx context.Context, id uuid.UUID) (Region, error)
	CountRegionChildren(ctx context.Context, id uuid.UUID) (int64, error)
	CreateRegion(ctx context.Context, region *Region) error
	SaveRegion(ctx context.Context, region *Region) error
	DeleteRegion(ctx context.Context, id uuid.UUID) error

	ListUnitsOfMeasure(ctx context.Context, filter UnitOfMeasureFilter) ([]UnitOfMeasure, int64, error)
	FindUnitOfMeasure(ctx context.Context, id uuid.UUID) (UnitOfMeasure, error)
	CreateUnitOfMeasure(ctx context.Context, unit *UnitOfMeasure) error
	SaveUnitOfMeasure(ctx context.Context, unit *UnitOfMeasure) error
	DeleteUnitOfMeasure(ctx context.Context, id uuid.UUID) error

	ListAccounts(ctx context.Context, filter AccountFilter) ([]Account, int64, error)
	FindAccount(ctx context.Context, id uuid.UUID) (Account, error)
	CountAccountChildren(ctx context.Context, id uuid.UUID) (int64, error)
	AccountDescendsFrom(ctx context.Context, candidateID, ancestorID uuid.UUID) (bool, error)
	CreateAccount(ctx context.Context, account *Account) error
	SaveAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListRegions(ctx context.Context, filter RegionFilter) ([]Region, int64, error) {
	return database.FindPage[Region](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "name ASC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindRegion(ctx context.Context, id uuid.UUID) (Region, error) {
	var region Region
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&region).Error
	return region, database.Translate(err)
}

func (r *gormRepository) CountRegionChildren(ctx context.Context, id uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&Region{}).Where("parent_id = ?", id).Count(&total).Error
	return total, database.Translate(err)
}

func (r *gormRepository) CreateRegion(ctx context.Context, region *Region) error {
	return database.Translate(r.db.WithContext(ctx).Create(region).Error)
}

func (r *gormRepository) SaveRegion(ctx context.Context, region *Region) error {
	return database.Translate(r.db.WithContext(ctx).Save(region).Error)
}

func (r *gormRepository) DeleteRegion(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Region{}, id)
}

func (r *gormRepository) ListUnitsOfMeasure(ctx context.Context, filter UnitOfMeasureFilter) ([]UnitOfMeasure, int64, error) {
	return database.FindPage[UnitOfMeasure](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindUnitOfMeasure(ctx context.Context, id uuid.UUID) (UnitOfMeasure, error) {
	var unit UnitOfMeasure
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&unit).Error
	return unit, database.Translate(err)
}

func (r *gormRepository) CreateUnitOfMeasure(ctx context.Context, unit *UnitOfMeasure) error {
	return database.Translate(r.db.WithContext(ctx).Create(unit).Error)
}

func (r *gormRepository) SaveUnitOfMeasure(ctx context.Context, unit *UnitOfMeasure) error {
	return database.Translate(r.db.WithContext(ctx).Save(unit).Error)
}

func (r *gormRepository) DeleteUnitOfMeasure(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &UnitOfMeasure{}, id)
}

func (r *gormRepository) ListAccounts(ctx context.Context, filter AccountFilter) ([]Account, int64, error) {
	return database.FindPage[Account](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindAccount(ctx context.Context, id uuid.UUID) (Account, error) {
	var account Account
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&account).Error
	return account, database.Translate(err)
}

func (r *gormRepository) CountAccountChildren(ctx context.Context, id uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&Account{}).Where("parent_id = ?", id).Count(&total).Error
	return total, database.Translate(err)
}

func (r *gormRepository) AccountDescendsFrom(ctx context.Context, candidateID, ancestorID uuid.UUID) (bool, error) {
	var found bool
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE lineage AS (
			SELECT id, parent_id FROM account WHERE id = ?
			UNION
			SELECT a.id, a.parent_id FROM account a JOIN lineage l ON a.id = l.parent_id
		)
		SELECT EXISTS (SELECT 1 FROM lineage WHERE id = ?)`, candidateID, ancestorID).
		Scan(&found).Error
	return found, database.Translate(err)
}

func (r *gormRepository) CreateAccount(ctx context.Context, account *Account) error {
	return database.Translate(r.db.WithContext(ctx).Create(account).Error)
}

func (r *gormRepository) SaveAccount(ctx context.Context, account *Account) error {
	return database.Translate(r.db.WithContext(ctx).Save(account).Error)
}

func (r *gormRepository) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Account{}, id)
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

func searchNameOrCode(db *gorm.DB, term string) *gorm.DB {
	if term == "" {
		return db
	}
	pattern := database.ContainsPattern(term)
	return db.Where("(name ILIKE ? OR code ILIKE ?)", pattern, pattern)
}

func (f RegionFilter) apply(db *gorm.DB) *gorm.DB {
	db = searchNameOrCode(db, f.Search)
	if f.Type != "" {
		db = db.Where("type = ?", f.Type)
	}
	if f.ParentID != nil {
		db = db.Where("parent_id = ?", *f.ParentID)
	}
	return db
}

func (f UnitOfMeasureFilter) apply(db *gorm.DB) *gorm.DB {
	return searchNameOrCode(db, f.Search)
}

func (f AccountFilter) apply(db *gorm.DB) *gorm.DB {
	db = searchNameOrCode(db, f.Search)
	if f.Type != "" {
		db = db.Where("type = ?", f.Type)
	}
	if f.ParentID != nil {
		db = db.Where("parent_id = ?", *f.ParentID)
	}
	return db
}
