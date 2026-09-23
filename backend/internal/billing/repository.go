package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const settledStatus = "paid"

var partyTables = map[string]string{
	"customer": "customer",
	"vendor":   "vendor",
}

type DueFilter struct {
	Settled   *bool
	DueFrom   *time.Time
	DueBefore *time.Time
}

type InvoiceFilter struct {
	DueFilter
	Search    string
	PartyType string
	PartyID   *uuid.UUID
	ProjectID *uuid.UUID
	Offset    int
	Limit     int
}

type ReceivableFilter struct {
	DueFilter
	Search     string
	CustomerID *uuid.UUID
	Offset     int
	Limit      int
}

type PayableFilter struct {
	DueFilter
	Search   string
	VendorID *uuid.UUID
	Offset   int
	Limit    int
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListInvoices(ctx context.Context, filter InvoiceFilter) ([]Invoice, int64, error)
	FindInvoice(ctx context.Context, id uuid.UUID) (Invoice, error)
	LockInvoice(ctx context.Context, id uuid.UUID) (Invoice, error)
	CreateInvoice(ctx context.Context, invoice *Invoice) error
	SaveInvoice(ctx context.Context, invoice *Invoice) error
	DeleteInvoice(ctx context.Context, id uuid.UUID) error
	PartyExists(ctx context.Context, partyType string, id uuid.UUID) (bool, error)

	ListReceivables(ctx context.Context, filter ReceivableFilter) ([]Receivable, int64, error)
	FindReceivable(ctx context.Context, id uuid.UUID) (Receivable, error)
	LockReceivable(ctx context.Context, id uuid.UUID) (Receivable, error)
	CreateReceivable(ctx context.Context, receivable *Receivable) error
	SaveReceivable(ctx context.Context, receivable *Receivable) error
	DeleteReceivable(ctx context.Context, id uuid.UUID) error

	ListPayables(ctx context.Context, filter PayableFilter) ([]Payable, int64, error)
	FindPayable(ctx context.Context, id uuid.UUID) (Payable, error)
	LockPayable(ctx context.Context, id uuid.UUID) (Payable, error)
	CreatePayable(ctx context.Context, payable *Payable) error
	SavePayable(ctx context.Context, payable *Payable) error
	DeletePayable(ctx context.Context, id uuid.UUID) error
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

func (r *gormRepository) ListInvoices(ctx context.Context, filter InvoiceFilter) ([]Invoice, int64, error) {
	return database.FindPage[Invoice](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "due_date ASC, created_at ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindInvoice(ctx context.Context, id uuid.UUID) (Invoice, error) {
	var invoice Invoice
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&invoice).Error
	return invoice, database.Translate(err)
}

func (r *gormRepository) LockInvoice(ctx context.Context, id uuid.UUID) (Invoice, error) {
	var invoice Invoice
	err := r.locked(ctx).Where("id = ?", id).Take(&invoice).Error
	return invoice, database.Translate(err)
}

func (r *gormRepository) CreateInvoice(ctx context.Context, invoice *Invoice) error {
	return database.Translate(r.db.WithContext(ctx).Create(invoice).Error)
}

func (r *gormRepository) SaveInvoice(ctx context.Context, invoice *Invoice) error {
	return database.Translate(r.db.WithContext(ctx).Save(invoice).Error)
}

func (r *gormRepository) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Invoice{}, id)
}

func (r *gormRepository) PartyExists(ctx context.Context, partyType string, id uuid.UUID) (bool, error) {
	table, known := partyTables[partyType]
	if !known {
		return false, nil
	}

	var exists bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS (?)", r.db.Table(table).Select("1").Where("id = ?", id)).
		Scan(&exists).Error
	return exists, database.Translate(err)
}

func (r *gormRepository) ListReceivables(ctx context.Context, filter ReceivableFilter) ([]Receivable, int64, error) {
	return database.FindPage[Receivable](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "due_date ASC, created_at ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindReceivable(ctx context.Context, id uuid.UUID) (Receivable, error) {
	var receivable Receivable
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&receivable).Error
	return receivable, database.Translate(err)
}

func (r *gormRepository) LockReceivable(ctx context.Context, id uuid.UUID) (Receivable, error) {
	var receivable Receivable
	err := r.locked(ctx).Where("id = ?", id).Take(&receivable).Error
	return receivable, database.Translate(err)
}

func (r *gormRepository) CreateReceivable(ctx context.Context, receivable *Receivable) error {
	return database.Translate(r.db.WithContext(ctx).Create(receivable).Error)
}

func (r *gormRepository) SaveReceivable(ctx context.Context, receivable *Receivable) error {
	return database.Translate(r.db.WithContext(ctx).Save(receivable).Error)
}

func (r *gormRepository) DeleteReceivable(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Receivable{}, id)
}

func (r *gormRepository) ListPayables(ctx context.Context, filter PayableFilter) ([]Payable, int64, error) {
	return database.FindPage[Payable](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "due_date ASC, created_at ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindPayable(ctx context.Context, id uuid.UUID) (Payable, error) {
	var payable Payable
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&payable).Error
	return payable, database.Translate(err)
}

func (r *gormRepository) LockPayable(ctx context.Context, id uuid.UUID) (Payable, error) {
	var payable Payable
	err := r.locked(ctx).Where("id = ?", id).Take(&payable).Error
	return payable, database.Translate(err)
}

func (r *gormRepository) CreatePayable(ctx context.Context, payable *Payable) error {
	return database.Translate(r.db.WithContext(ctx).Create(payable).Error)
}

func (r *gormRepository) SavePayable(ctx context.Context, payable *Payable) error {
	return database.Translate(r.db.WithContext(ctx).Save(payable).Error)
}

func (r *gormRepository) DeletePayable(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Payable{}, id)
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

func (f DueFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Settled != nil && *f.Settled {
		db = db.Where("status = ?", settledStatus)
	}
	if f.Settled != nil && !*f.Settled {
		db = db.Where("status <> ?", settledStatus)
	}
	if f.DueFrom != nil {
		db = db.Where("due_date >= ?", *f.DueFrom)
	}
	if f.DueBefore != nil {
		db = db.Where("due_date < ?", *f.DueBefore)
	}
	return db
}

func (f InvoiceFilter) apply(db *gorm.DB) *gorm.DB {
	db = f.DueFilter.apply(db)
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(number ILIKE ? OR note ILIKE ?)", pattern, pattern)
	}
	if f.PartyType != "" {
		db = db.Where("party_type = ?", f.PartyType)
	}
	if f.PartyID != nil {
		db = db.Where("party_id = ?", *f.PartyID)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	return db
}

func (f ReceivableFilter) apply(db *gorm.DB) *gorm.DB {
	db = f.DueFilter.apply(db)
	if f.Search != "" {
		db = db.Where("reference ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.CustomerID != nil {
		db = db.Where("customer_id = ?", *f.CustomerID)
	}
	return db
}

func (f PayableFilter) apply(db *gorm.DB) *gorm.DB {
	db = f.DueFilter.apply(db)
	if f.Search != "" {
		db = db.Where("reference ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.VendorID != nil {
		db = db.Where("vendor_id = ?", *f.VendorID)
	}
	return db
}
