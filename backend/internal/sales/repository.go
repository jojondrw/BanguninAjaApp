package sales

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

	contractColumns = "contract.id, contract.number, contract.customer_id, customer.name AS customer_name, " +
		"contract.unit_id, unit.code AS unit_code, contract.type, contract.value, contract.date, contract.status, " +
		"contract.created_at, contract.updated_at"
	installmentColumns = "installment.id, installment.contract_id, contract.number AS contract_number, " +
		"contract.customer_id, customer.name AS customer_name, installment.installment_number, installment.due_date, " +
		"installment.amount, installment.paid_date, installment.status, installment.created_at, installment.updated_at"
)

type CustomerFilter struct {
	Search string
	Offset int
	Limit  int
}

type UnitFilter struct {
	Search    string
	ProjectID *uuid.UUID
	Status    string
	Offset    int
	Limit     int
}

type LeadFilter struct {
	Search    string
	ProjectID *uuid.UUID
	Stage     string
	Offset    int
	Limit     int
}

type ContractFilter struct {
	Search     string
	CustomerID *uuid.UUID
	UnitID     *uuid.UUID
	Status     string
	Type       string
	Offset     int
	Limit      int
}

type InstallmentFilter struct {
	ContractID *uuid.UUID
	CustomerID *uuid.UUID
	Settled    *bool
	DueFrom    *time.Time
	DueBefore  *time.Time
	Offset     int
	Limit      int
}

type UnitStatusCount struct {
	Status string
	Total  int64
}

type ContractRow struct {
	ID           uuid.UUID
	Number       string
	CustomerID   uuid.UUID
	CustomerName string
	UnitID       uuid.UUID
	UnitCode     string
	Type         string
	Value        int64
	Date         time.Time
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type InstallmentRow struct {
	ID                uuid.UUID
	ContractID        uuid.UUID
	ContractNumber    string
	CustomerID        uuid.UUID
	CustomerName      string
	InstallmentNumber int
	DueDate           time.Time
	Amount            int64
	PaidDate          *time.Time
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListCustomers(ctx context.Context, filter CustomerFilter) ([]Customer, int64, error)
	FindCustomer(ctx context.Context, id uuid.UUID) (Customer, error)
	CreateCustomer(ctx context.Context, customer *Customer) error
	SaveCustomer(ctx context.Context, customer *Customer) error
	DeleteCustomer(ctx context.Context, id uuid.UUID) error

	ListUnits(ctx context.Context, filter UnitFilter) ([]Unit, int64, error)
	CountUnitsByStatus(ctx context.Context, projectID *uuid.UUID) ([]UnitStatusCount, error)
	FindUnit(ctx context.Context, id uuid.UUID) (Unit, error)
	LockUnit(ctx context.Context, id uuid.UUID) (Unit, error)
	CreateUnit(ctx context.Context, unit *Unit) error
	SaveUnit(ctx context.Context, unit *Unit) error
	DeleteUnit(ctx context.Context, id uuid.UUID) error

	ListLeads(ctx context.Context, filter LeadFilter) ([]Lead, int64, error)
	FindLead(ctx context.Context, id uuid.UUID) (Lead, error)
	CreateLead(ctx context.Context, lead *Lead) error
	SaveLead(ctx context.Context, lead *Lead) error
	DeleteLead(ctx context.Context, id uuid.UUID) error

	ListContracts(ctx context.Context, filter ContractFilter) ([]ContractRow, int64, error)
	FindContractRow(ctx context.Context, id uuid.UUID) (ContractRow, error)
	LockContract(ctx context.Context, id uuid.UUID) (Contract, error)
	CreateContract(ctx context.Context, contract *Contract) error
	SaveContract(ctx context.Context, contract *Contract) error
	DeleteContract(ctx context.Context, id uuid.UUID) error

	ListInstallments(ctx context.Context, filter InstallmentFilter) ([]InstallmentRow, int64, error)
	FindInstallmentRow(ctx context.Context, id uuid.UUID) (InstallmentRow, error)
	FindInstallment(ctx context.Context, contractID, id uuid.UUID) (Installment, error)
	CreateInstallment(ctx context.Context, installment *Installment) error
	SaveInstallment(ctx context.Context, installment *Installment) error
	DeleteInstallment(ctx context.Context, contractID, id uuid.UUID) error
	SumInstallments(ctx context.Context, contractID uuid.UUID, excludeID *uuid.UUID) (int64, error)
	CountUnpaidInstallments(ctx context.Context, contractID uuid.UUID) (int64, error)
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

func (r *gormRepository) ListCustomers(ctx context.Context, filter CustomerFilter) ([]Customer, int64, error) {
	return database.FindPage[Customer](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "name ASC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindCustomer(ctx context.Context, id uuid.UUID) (Customer, error) {
	var customer Customer
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&customer).Error
	return customer, database.Translate(err)
}

func (r *gormRepository) CreateCustomer(ctx context.Context, customer *Customer) error {
	return database.Translate(r.db.WithContext(ctx).Create(customer).Error)
}

func (r *gormRepository) SaveCustomer(ctx context.Context, customer *Customer) error {
	return database.Translate(r.db.WithContext(ctx).Save(customer).Error)
}

func (r *gormRepository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Customer{}, "id = ?", id)
}

func (r *gormRepository) ListUnits(ctx context.Context, filter UnitFilter) ([]Unit, int64, error) {
	return database.FindPage[Unit](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "code ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) CountUnitsByStatus(ctx context.Context, projectID *uuid.UUID) ([]UnitStatusCount, error) {
	var counts []UnitStatusCount
	query := r.db.WithContext(ctx).Model(&Unit{}).Select("status, COUNT(*) AS total")
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	err := query.Group("status").Order("status ASC").Scan(&counts).Error
	return counts, database.Translate(err)
}

func (r *gormRepository) FindUnit(ctx context.Context, id uuid.UUID) (Unit, error) {
	var unit Unit
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&unit).Error
	return unit, database.Translate(err)
}

func (r *gormRepository) LockUnit(ctx context.Context, id uuid.UUID) (Unit, error) {
	var unit Unit
	err := r.locked(ctx).Where("id = ?", id).Take(&unit).Error
	return unit, database.Translate(err)
}

func (r *gormRepository) CreateUnit(ctx context.Context, unit *Unit) error {
	return database.Translate(r.db.WithContext(ctx).Create(unit).Error)
}

func (r *gormRepository) SaveUnit(ctx context.Context, unit *Unit) error {
	return database.Translate(r.db.WithContext(ctx).Save(unit).Error)
}

func (r *gormRepository) DeleteUnit(ctx context.Context, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Unit{}, "id = ?", id)
}

func (r *gormRepository) ListLeads(ctx context.Context, filter LeadFilter) ([]Lead, int64, error) {
	return database.FindPage[Lead](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "last_contacted_at DESC NULLS LAST, created_at DESC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindLead(ctx context.Context, id uuid.UUID) (Lead, error) {
	var lead Lead
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&lead).Error
	return lead, database.Translate(err)
}

func (r *gormRepository) CreateLead(ctx context.Context, lead *Lead) error {
	return database.Translate(r.db.WithContext(ctx).Create(lead).Error)
}

func (r *gormRepository) SaveLead(ctx context.Context, lead *Lead) error {
	return database.Translate(r.db.WithContext(ctx).Save(lead).Error)
}

func (r *gormRepository) DeleteLead(ctx context.Context, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Lead{}, "id = ?", id)
}

func (r *gormRepository) ListContracts(ctx context.Context, filter ContractFilter) ([]ContractRow, int64, error) {
	var total int64
	if err := r.contractRows(ctx).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]ContractRow, 0, filter.Limit)
	err := r.contractRows(ctx).
		Scopes(filter.apply).
		Select(contractColumns).
		Order("contract.date DESC, contract.number ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindContractRow(ctx context.Context, id uuid.UUID) (ContractRow, error) {
	var row ContractRow
	err := r.contractRows(ctx).Select(contractColumns).Where("contract.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) LockContract(ctx context.Context, id uuid.UUID) (Contract, error) {
	var contract Contract
	err := r.locked(ctx).Where("id = ?", id).Take(&contract).Error
	return contract, database.Translate(err)
}

func (r *gormRepository) CreateContract(ctx context.Context, contract *Contract) error {
	return database.Translate(r.db.WithContext(ctx).Create(contract).Error)
}

func (r *gormRepository) SaveContract(ctx context.Context, contract *Contract) error {
	return database.Translate(r.db.WithContext(ctx).Save(contract).Error)
}

func (r *gormRepository) DeleteContract(ctx context.Context, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Contract{}, "id = ?", id)
}

func (r *gormRepository) ListInstallments(ctx context.Context, filter InstallmentFilter) ([]InstallmentRow, int64, error) {
	var total int64
	if err := r.installmentRows(ctx).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]InstallmentRow, 0, filter.Limit)
	err := r.installmentRows(ctx).
		Scopes(filter.apply).
		Select(installmentColumns).
		Order("installment.due_date ASC, contract.number ASC, installment.installment_number ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindInstallmentRow(ctx context.Context, id uuid.UUID) (InstallmentRow, error) {
	var row InstallmentRow
	err := r.installmentRows(ctx).Select(installmentColumns).Where("installment.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) FindInstallment(ctx context.Context, contractID, id uuid.UUID) (Installment, error) {
	var installment Installment
	err := r.db.WithContext(ctx).Where("id = ? AND contract_id = ?", id, contractID).Take(&installment).Error
	return installment, database.Translate(err)
}

func (r *gormRepository) CreateInstallment(ctx context.Context, installment *Installment) error {
	return database.Translate(r.db.WithContext(ctx).Create(installment).Error)
}

func (r *gormRepository) SaveInstallment(ctx context.Context, installment *Installment) error {
	return database.Translate(r.db.WithContext(ctx).Save(installment).Error)
}

func (r *gormRepository) DeleteInstallment(ctx context.Context, contractID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &Installment{}, "id = ? AND contract_id = ?", id, contractID)
}

func (r *gormRepository) SumInstallments(ctx context.Context, contractID uuid.UUID, excludeID *uuid.UUID) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).
		Model(&Installment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("contract_id = ?", contractID)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}
	err := query.Scan(&total).Error
	return total, database.Translate(err)
}

func (r *gormRepository) CountUnpaidInstallments(ctx context.Context, contractID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Installment{}).
		Where("contract_id = ? AND status <> ?", contractID, settledStatus).
		Count(&total).Error
	return total, database.Translate(err)
}

func (r *gormRepository) contractRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&Contract{}).
		Joins("JOIN customer ON customer.id = contract.customer_id").
		Joins("JOIN unit ON unit.id = contract.unit_id")
}

func (r *gormRepository) installmentRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&Installment{}).
		Joins("JOIN contract ON contract.id = installment.contract_id").
		Joins("JOIN customer ON customer.id = contract.customer_id")
}

func (r *gormRepository) locked(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
}

func (r *gormRepository) deleteWhere(ctx context.Context, model any, condition string, args ...any) error {
	result := r.db.WithContext(ctx).Where(condition, args...).Delete(model)
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (f CustomerFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR email ILIKE ? OR identity_number ILIKE ?)", pattern, pattern, pattern)
	}
	return db
}

func (f UnitFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(code ILIKE ? OR unit_type ILIKE ?)", pattern, pattern)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	return db
}

func (f LeadFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR contact ILIKE ?)", pattern, pattern)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	if f.Stage != "" {
		db = db.Where("stage = ?", f.Stage)
	}
	return db
}

func (f ContractFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		db = db.Where("contract.number ILIKE ?", database.ContainsPattern(f.Search))
	}
	if f.CustomerID != nil {
		db = db.Where("contract.customer_id = ?", *f.CustomerID)
	}
	if f.UnitID != nil {
		db = db.Where("contract.unit_id = ?", *f.UnitID)
	}
	if f.Status != "" {
		db = db.Where("contract.status = ?", f.Status)
	}
	if f.Type != "" {
		db = db.Where("contract.type = ?", f.Type)
	}
	return db
}

func (f InstallmentFilter) apply(db *gorm.DB) *gorm.DB {
	if f.ContractID != nil {
		db = db.Where("installment.contract_id = ?", *f.ContractID)
	}
	if f.CustomerID != nil {
		db = db.Where("contract.customer_id = ?", *f.CustomerID)
	}
	if f.Settled != nil && *f.Settled {
		db = db.Where("installment.status = ?", settledStatus)
	}
	if f.Settled != nil && !*f.Settled {
		db = db.Where("installment.status <> ?", settledStatus)
	}
	if f.DueFrom != nil {
		db = db.Where("installment.due_date >= ?", *f.DueFrom)
	}
	if f.DueBefore != nil {
		db = db.Where("installment.due_date < ?", *f.DueBefore)
	}
	return db
}
