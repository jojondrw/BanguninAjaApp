package master

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	regionColumns = "region.id, region.code, region.name, region.type, region.parent_id, region.created_at, region.updated_at, " +
		"(SELECT parent.code FROM region AS parent WHERE parent.id = region.parent_id) AS parent_code, " +
		"(SELECT parent.name FROM region AS parent WHERE parent.id = region.parent_id) AS parent_name"
	accountHasChildren = "EXISTS (SELECT 1 FROM account AS child WHERE child.parent_id = account.id)"
	accountColumns     = "account.id, account.code, account.name, account.type, account.parent_id, account.created_at, account.updated_at, " +
		"(SELECT parent.code FROM account AS parent WHERE parent.id = account.parent_id) AS parent_code, " +
		"(SELECT parent.name FROM account AS parent WHERE parent.id = account.parent_id) AS parent_name, " +
		"NOT " + accountHasChildren + " AS postable"
)

type usage struct {
	label string
	query string
}

var (
	regionUsages = []usage{
		{label: "wilayah turunan", query: "SELECT EXISTS (SELECT 1 FROM region WHERE parent_id = ?)"},
		{label: "proyek", query: "SELECT EXISTS (SELECT 1 FROM project WHERE region_id = ?)"},
		{label: "lokasi tersimpan", query: "SELECT EXISTS (SELECT 1 FROM saved_location WHERE region_id = ?)"},
	}
	unitOfMeasureUsages = []usage{
		{label: "material", query: "SELECT EXISTS (SELECT 1 FROM material WHERE unit_of_measure_id = ?)"},
		{label: "RAB proyek", query: "SELECT EXISTS (SELECT 1 FROM budget_item WHERE unit_of_measure_id = ?)"},
		{label: "permintaan pembelian", query: "SELECT EXISTS (SELECT 1 FROM purchase_request_item WHERE unit_of_measure_id = ?)"},
		{label: "pesanan pembelian", query: "SELECT EXISTS (SELECT 1 FROM purchase_order_item WHERE unit_of_measure_id = ?)"},
	}
	accountUsages = []usage{
		{label: "akun turunan", query: "SELECT EXISTS (SELECT 1 FROM account WHERE parent_id = ?)"},
		{label: "transaksi kas", query: "SELECT EXISTS (SELECT 1 FROM cash_transaction WHERE account_id = ?)"},
		{label: "jurnal", query: "SELECT EXISTS (SELECT 1 FROM journal_line WHERE account_id = ?)"},
	}
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
	Postable *bool
	Offset   int
	Limit    int
}

type RegionRow struct {
	ID         uuid.UUID
	Code       string
	Name       string
	Type       string
	ParentID   *uuid.UUID
	ParentCode *string
	ParentName *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AccountRow struct {
	ID         uuid.UUID
	Code       string
	Name       string
	Type       string
	ParentID   *uuid.UUID
	ParentCode *string
	ParentName *string
	Postable   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Repository interface {
	ListRegions(ctx context.Context, filter RegionFilter) ([]RegionRow, int64, error)
	FindRegion(ctx context.Context, id uuid.UUID) (Region, error)
	FindRegionRow(ctx context.Context, id uuid.UUID) (RegionRow, error)
	CountRegionChildren(ctx context.Context, id uuid.UUID) (int64, error)
	CreateRegion(ctx context.Context, region *Region) error
	SaveRegion(ctx context.Context, region *Region) error
	DeleteRegion(ctx context.Context, id uuid.UUID) error
	ListRegionUsages(ctx context.Context, id uuid.UUID) ([]string, error)

	ListUnitsOfMeasure(ctx context.Context, filter UnitOfMeasureFilter) ([]UnitOfMeasure, int64, error)
	FindUnitOfMeasure(ctx context.Context, id uuid.UUID) (UnitOfMeasure, error)
	CreateUnitOfMeasure(ctx context.Context, unit *UnitOfMeasure) error
	SaveUnitOfMeasure(ctx context.Context, unit *UnitOfMeasure) error
	DeleteUnitOfMeasure(ctx context.Context, id uuid.UUID) error
	ListUnitOfMeasureUsages(ctx context.Context, id uuid.UUID) ([]string, error)

	ListAccounts(ctx context.Context, filter AccountFilter) ([]AccountRow, int64, error)
	FindAccount(ctx context.Context, id uuid.UUID) (Account, error)
	FindAccountRow(ctx context.Context, id uuid.UUID) (AccountRow, error)
	CountAccountChildren(ctx context.Context, id uuid.UUID) (int64, error)
	AccountDescendsFrom(ctx context.Context, candidateID, ancestorID uuid.UUID) (bool, error)
	CreateAccount(ctx context.Context, account *Account) error
	SaveAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	ListAccountUsages(ctx context.Context, id uuid.UUID) ([]string, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListRegions(ctx context.Context, filter RegionFilter) ([]RegionRow, int64, error) {
	return findRows[RegionRow](ctx, r.db, &Region{}, regionColumns, database.Listing{
		Filter: filter.apply,
		Order:  "region.name ASC, region.id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindRegion(ctx context.Context, id uuid.UUID) (Region, error) {
	var region Region
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&region).Error
	return region, database.Translate(err)
}

func (r *gormRepository) FindRegionRow(ctx context.Context, id uuid.UUID) (RegionRow, error) {
	var row RegionRow
	err := r.db.WithContext(ctx).Model(&Region{}).Select(regionColumns).Where("region.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
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

func (r *gormRepository) ListRegionUsages(ctx context.Context, id uuid.UUID) ([]string, error) {
	return r.listUsages(ctx, regionUsages, id)
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

func (r *gormRepository) ListUnitOfMeasureUsages(ctx context.Context, id uuid.UUID) ([]string, error) {
	return r.listUsages(ctx, unitOfMeasureUsages, id)
}

func (r *gormRepository) ListAccounts(ctx context.Context, filter AccountFilter) ([]AccountRow, int64, error) {
	return findRows[AccountRow](ctx, r.db, &Account{}, accountColumns, database.Listing{
		Filter: filter.apply,
		Order:  "account.code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindAccount(ctx context.Context, id uuid.UUID) (Account, error) {
	var account Account
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&account).Error
	return account, database.Translate(err)
}

func (r *gormRepository) FindAccountRow(ctx context.Context, id uuid.UUID) (AccountRow, error) {
	var row AccountRow
	err := r.db.WithContext(ctx).Model(&Account{}).Select(accountColumns).Where("account.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
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

func (r *gormRepository) ListAccountUsages(ctx context.Context, id uuid.UUID) ([]string, error) {
	return r.listUsages(ctx, accountUsages, id)
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

func (r *gormRepository) listUsages(ctx context.Context, usages []usage, id uuid.UUID) ([]string, error) {
	labels := make([]string, 0, len(usages))
	for _, candidate := range usages {
		var used bool
		if err := r.db.WithContext(ctx).Raw(candidate.query, id).Scan(&used).Error; err != nil {
			return nil, database.Translate(err)
		}
		if used {
			labels = append(labels, candidate.label)
		}
	}
	return labels, nil
}

func findRows[T any](ctx context.Context, db *gorm.DB, model any, columns string, listing database.Listing) ([]T, int64, error) {
	var total int64
	if err := db.WithContext(ctx).Model(model).Scopes(listing.Filter).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]T, 0, listing.Limit)
	err := db.WithContext(ctx).
		Model(model).
		Scopes(listing.Filter).
		Select(columns).
		Order(listing.Order).
		Offset(listing.Offset).
		Limit(listing.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
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
	return filterPostable(db, f.Postable)
}

func filterPostable(db *gorm.DB, postable *bool) *gorm.DB {
	if postable == nil {
		return db
	}
	if *postable {
		return db.Where("NOT " + accountHasChildren)
	}
	return db.Where(accountHasChildren)
}
