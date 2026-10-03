package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	settledStatus = "paid"

	invoiceColumns = "invoice.*, project.name AS project_name, " +
		"receivable.id AS receivable_id, receivable.reference AS receivable_reference, " +
		"COALESCE(customer.name, vendor.name) AS party_name, " +
		"(SELECT MAX(billing_payment.paid_at) FROM billing_payment WHERE billing_payment.invoice_id = invoice.id) AS last_paid_at"
	receivableColumns = "receivable.*, customer.name AS customer_name, project.name AS project_name, " +
		"contract.number AS contract_number, invoice.number AS invoice_number, " +
		"(SELECT MAX(billing_payment.paid_at) FROM billing_payment WHERE billing_payment.receivable_id = receivable.id) AS last_paid_at"
	payableColumns = "payable.*, vendor.name AS vendor_name, project.name AS project_name, " +
		"purchase_order.number AS purchase_order_number, " +
		"(SELECT MAX(billing_payment.paid_at) FROM billing_payment WHERE billing_payment.payable_id = payable.id) AS last_paid_at"
	paymentColumns = "billing_payment.*, invoice.number AS invoice_number, " +
		"receivable.reference AS receivable_reference, users.name AS created_by_name"

	invoicePaymentColumn    = "invoice_id"
	receivablePaymentColumn = "receivable_id"
	payablePaymentColumn    = "payable_id"
)

var partyTables = map[string]string{
	"customer": "customer",
	"vendor":   "vendor",
}

type InvoiceRow struct {
	Invoice
	ProjectName         *string
	ReceivableID        *uuid.UUID
	ReceivableReference *string
	PartyName           *string
	LastPaidAt          *time.Time
}

type ReceivableRow struct {
	Receivable
	CustomerName   string
	ProjectName    *string
	ContractNumber *string
	InvoiceNumber  *string
	LastPaidAt     *time.Time
}

type PayableRow struct {
	Payable
	VendorName          string
	ProjectName         *string
	PurchaseOrderNumber *string
	LastPaidAt          *time.Time
}

type PaymentRow struct {
	Payment
	InvoiceNumber       *string
	ReceivableReference *string
	CreatedByName       *string
}

type PaymentFilter struct {
	Column   string
	ParentID uuid.UUID
	Offset   int
	Limit    int
}

type ContractRef struct {
	ID         uuid.UUID
	Number     string
	CustomerID uuid.UUID
}

type PurchaseOrderRef struct {
	ID       uuid.UUID
	Number   string
	VendorID uuid.UUID
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
	Recorded  *bool
	Offset    int
	Limit     int
}

type ReceivableFilter struct {
	DueFilter
	Search     string
	CustomerID *uuid.UUID
	ProjectID  *uuid.UUID
	ContractID *uuid.UUID
	Offset     int
	Limit      int
}

type PayableFilter struct {
	DueFilter
	Search          string
	VendorID        *uuid.UUID
	ProjectID       *uuid.UUID
	PurchaseOrderID *uuid.UUID
	Offset          int
	Limit           int
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListInvoices(ctx context.Context, filter InvoiceFilter) ([]InvoiceRow, int64, error)
	FindInvoiceRow(ctx context.Context, id uuid.UUID) (InvoiceRow, error)
	FindInvoice(ctx context.Context, id uuid.UUID) (Invoice, error)
	LockInvoice(ctx context.Context, id uuid.UUID) (Invoice, error)
	CreateInvoice(ctx context.Context, invoice *Invoice) error
	SaveInvoice(ctx context.Context, invoice *Invoice) error
	DeleteInvoice(ctx context.Context, id uuid.UUID) error
	PartyExists(ctx context.Context, partyType string, id uuid.UUID) (bool, error)

	ListReceivables(ctx context.Context, filter ReceivableFilter) ([]ReceivableRow, int64, error)
	FindReceivableRow(ctx context.Context, id uuid.UUID) (ReceivableRow, error)
	FindReceivable(ctx context.Context, id uuid.UUID) (Receivable, error)
	FindReceivableByInvoice(ctx context.Context, invoiceID uuid.UUID) (Receivable, error)
	LockReceivable(ctx context.Context, id uuid.UUID) (Receivable, error)
	CreateReceivable(ctx context.Context, receivable *Receivable) error
	SaveReceivable(ctx context.Context, receivable *Receivable) error
	DeleteReceivable(ctx context.Context, id uuid.UUID) error

	ListPayables(ctx context.Context, filter PayableFilter) ([]PayableRow, int64, error)
	FindPayableRow(ctx context.Context, id uuid.UUID) (PayableRow, error)
	LockPayable(ctx context.Context, id uuid.UUID) (Payable, error)
	CreatePayable(ctx context.Context, payable *Payable) error
	SavePayable(ctx context.Context, payable *Payable) error
	DeletePayable(ctx context.Context, id uuid.UUID) error

	ProjectExists(ctx context.Context, id uuid.UUID) (bool, error)
	FindContract(ctx context.Context, id uuid.UUID) (ContractRef, error)
	FindPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrderRef, error)

	CreatePayment(ctx context.Context, payment *Payment) error
	ListPayments(ctx context.Context, filter PaymentFilter) ([]PaymentRow, int64, error)
	LinkPayments(ctx context.Context, invoiceID, receivableID uuid.UUID) error
	BackfillPayments(ctx context.Context) (int64, error)
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

func (r *gormRepository) ListInvoices(ctx context.Context, filter InvoiceFilter) ([]InvoiceRow, int64, error) {
	return listRows[InvoiceRow](r.invoiceRows(ctx), filter.apply, invoiceColumns, "invoice", filter.Offset, filter.Limit)
}

func (r *gormRepository) FindInvoiceRow(ctx context.Context, id uuid.UUID) (InvoiceRow, error) {
	var row InvoiceRow
	err := r.invoiceRows(ctx).Select(invoiceColumns).Where("invoice.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
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
	return r.exists(ctx, table, id)
}

func (r *gormRepository) ListReceivables(ctx context.Context, filter ReceivableFilter) ([]ReceivableRow, int64, error) {
	return listRows[ReceivableRow](r.receivableRows(ctx), filter.apply, receivableColumns, "receivable", filter.Offset, filter.Limit)
}

func (r *gormRepository) FindReceivableRow(ctx context.Context, id uuid.UUID) (ReceivableRow, error) {
	var row ReceivableRow
	err := r.receivableRows(ctx).Select(receivableColumns).Where("receivable.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) FindReceivable(ctx context.Context, id uuid.UUID) (Receivable, error) {
	var receivable Receivable
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&receivable).Error
	return receivable, database.Translate(err)
}

func (r *gormRepository) FindReceivableByInvoice(ctx context.Context, invoiceID uuid.UUID) (Receivable, error) {
	var receivable Receivable
	err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).Take(&receivable).Error
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

func (r *gormRepository) ListPayables(ctx context.Context, filter PayableFilter) ([]PayableRow, int64, error) {
	return listRows[PayableRow](r.payableRows(ctx), filter.apply, payableColumns, "payable", filter.Offset, filter.Limit)
}

func (r *gormRepository) FindPayableRow(ctx context.Context, id uuid.UUID) (PayableRow, error) {
	var row PayableRow
	err := r.payableRows(ctx).Select(payableColumns).Where("payable.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
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

func (r *gormRepository) ProjectExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.exists(ctx, "project", id)
}

func (r *gormRepository) FindContract(ctx context.Context, id uuid.UUID) (ContractRef, error) {
	var contract ContractRef
	err := r.db.WithContext(ctx).Table("contract").Select("id, number, customer_id").Where("id = ?", id).Take(&contract).Error
	return contract, database.Translate(err)
}

func (r *gormRepository) FindPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrderRef, error) {
	var order PurchaseOrderRef
	err := r.db.WithContext(ctx).Table("purchase_order").Select("id, number, vendor_id").Where("id = ?", id).Take(&order).Error
	return order, database.Translate(err)
}

func (r *gormRepository) CreatePayment(ctx context.Context, payment *Payment) error {
	return database.Translate(r.db.WithContext(ctx).Create(payment).Error)
}

func (r *gormRepository) ListPayments(ctx context.Context, filter PaymentFilter) ([]PaymentRow, int64, error) {
	base := r.db.WithContext(ctx).
		Model(&Payment{}).
		Where("billing_payment."+filter.Column+" = ?", filter.ParentID)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]PaymentRow, 0, filter.Limit)
	err := base.Session(&gorm.Session{}).
		Joins("LEFT JOIN invoice ON invoice.id = billing_payment.invoice_id").
		Joins("LEFT JOIN receivable ON receivable.id = billing_payment.receivable_id").
		Joins("LEFT JOIN users ON users.id = billing_payment.created_by").
		Select(paymentColumns).
		Order("billing_payment.paid_at DESC, billing_payment.created_at DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) LinkPayments(ctx context.Context, invoiceID, receivableID uuid.UUID) error {
	err := r.db.WithContext(ctx).Exec(linkPaymentsStatement, invoiceID, receivableID, invoiceID, receivableID).Error
	return database.Translate(err)
}

func (r *gormRepository) BackfillPayments(ctx context.Context) (int64, error) {
	var inserted int64
	for _, statement := range backfillStatements {
		result := r.db.WithContext(ctx).Exec(statement, map[string]any{"method": backfillMethod, "note": backfillNote})
		if result.Error != nil {
			return inserted, database.Translate(result.Error)
		}
		inserted += result.RowsAffected
	}
	return inserted, nil
}

func (r *gormRepository) invoiceRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&Invoice{}).
		Joins("LEFT JOIN project ON project.id = invoice.project_id").
		Joins("LEFT JOIN receivable ON receivable.invoice_id = invoice.id").
		Joins("LEFT JOIN customer ON invoice.party_type = 'customer' AND customer.id = invoice.party_id").
		Joins("LEFT JOIN vendor ON invoice.party_type = 'vendor' AND vendor.id = invoice.party_id")
}

func (r *gormRepository) receivableRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&Receivable{}).
		Joins("JOIN customer ON customer.id = receivable.customer_id").
		Joins("LEFT JOIN project ON project.id = receivable.project_id").
		Joins("LEFT JOIN contract ON contract.id = receivable.contract_id").
		Joins("LEFT JOIN invoice ON invoice.id = receivable.invoice_id")
}

func (r *gormRepository) payableRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&Payable{}).
		Joins("JOIN vendor ON vendor.id = payable.vendor_id").
		Joins("LEFT JOIN project ON project.id = payable.project_id").
		Joins("LEFT JOIN purchase_order ON purchase_order.id = payable.purchase_order_id")
}

func listRows[T any](base *gorm.DB, filter func(*gorm.DB) *gorm.DB, columns, table string, offset, limit int) ([]T, int64, error) {
	var total int64
	if err := base.Session(&gorm.Session{}).Scopes(filter).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]T, 0, limit)
	err := base.Session(&gorm.Session{}).
		Scopes(filter).
		Select(columns).
		Order(table + ".due_date ASC, " + table + ".created_at ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) exists(ctx context.Context, table string, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS (?)", r.db.Table(table).Select("1").Where("id = ?", id)).
		Scan(&exists).Error
	return exists, database.Translate(err)
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

func (f DueFilter) scope(table string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.Settled != nil && *f.Settled {
			db = db.Where(table+".status = ?", settledStatus)
		}
		if f.Settled != nil && !*f.Settled {
			db = db.Where(table+".status <> ?", settledStatus)
		}
		if f.DueFrom != nil {
			db = db.Where(table+".due_date >= ?", *f.DueFrom)
		}
		if f.DueBefore != nil {
			db = db.Where(table+".due_date < ?", *f.DueBefore)
		}
		return db
	}
}

func (f InvoiceFilter) apply(db *gorm.DB) *gorm.DB {
	db = f.DueFilter.scope("invoice")(db)
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(invoice.number ILIKE ? OR invoice.note ILIKE ?)", pattern, pattern)
	}
	if f.PartyType != "" {
		db = db.Where("invoice.party_type = ?", f.PartyType)
	}
	if f.PartyID != nil {
		db = db.Where("invoice.party_id = ?", *f.PartyID)
	}
	if f.ProjectID != nil {
		db = db.Where("invoice.project_id = ?", *f.ProjectID)
	}
	if f.Recorded != nil && *f.Recorded {
		db = db.Where("receivable.id IS NOT NULL")
	}
	if f.Recorded != nil && !*f.Recorded {
		db = db.Where("receivable.id IS NULL")
	}
	return db
}

func (f ReceivableFilter) apply(db *gorm.DB) *gorm.DB {
	db = f.DueFilter.scope("receivable")(db)
	if f.Search != "" {
		db = db.Where("receivable.reference ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.CustomerID != nil {
		db = db.Where("receivable.customer_id = ?", *f.CustomerID)
	}
	if f.ProjectID != nil {
		db = db.Where("receivable.project_id = ?", *f.ProjectID)
	}
	if f.ContractID != nil {
		db = db.Where("receivable.contract_id = ?", *f.ContractID)
	}
	return db
}

func (f PayableFilter) apply(db *gorm.DB) *gorm.DB {
	db = f.DueFilter.scope("payable")(db)
	if f.Search != "" {
		db = db.Where("payable.reference ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.VendorID != nil {
		db = db.Where("payable.vendor_id = ?", *f.VendorID)
	}
	if f.ProjectID != nil {
		db = db.Where("payable.project_id = ?", *f.ProjectID)
	}
	if f.PurchaseOrderID != nil {
		db = db.Where("payable.purchase_order_id = ?", *f.PurchaseOrderID)
	}
	return db
}
