package reporting

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	cashIn          = "in"
	cashOut         = "out"
	paidStatus      = "paid"
	maximumRowCount = 10000

	cashLineColumns = "cash_transaction.date, cash_transaction.type, cash_transaction.amount, cash_transaction.note, " +
		"account.code AS account_code, account.name AS account_name, project.code AS project_code, project.name AS project_name"
	budgetColumns = "budget.year, budget.value, project.code AS project_code, project.name AS project_name, " +
		"COALESCE((SELECT SUM(cash_transaction.amount) FROM cash_transaction WHERE cash_transaction.project_id = budget.project_id " +
		"AND cash_transaction.type = ? AND cash_transaction.date >= make_date(budget.year::int, 1, 1) " +
		"AND cash_transaction.date < make_date((budget.year + 1)::int, 1, 1)), 0) AS realized"
	projectColumns = "project.code, project.name, project.type, project.status, project.progress, project.start_date, " +
		"project.target_end_date, project.contract_value, " +
		"COALESCE((SELECT SUM(budget_item.total) FROM budget_item WHERE budget_item.project_id = project.id), 0) AS budget_item_total"
	stockColumns = "warehouse.code AS warehouse_code, warehouse.name AS warehouse_name, material.code AS material_code, " +
		"material.name AS material_name, material.category, unit_of_measure.code AS unit_code, stock.quantity"
	lowStockColumns = "material.code AS material_code, material.name AS material_name, unit_of_measure.code AS unit_code, " +
		"material.minimum_stock, COALESCE(totals.quantity, 0) AS total_quantity"
	payrollColumns = "payroll.period, employee.identity_number, employee.name AS employee_name, employee.position, " +
		"project.code AS project_code, project.name AS project_name, payroll.basic_pay, payroll.allowance, payroll.deduction, " +
		"payroll.net_pay, payroll.paid"
	contractColumns = "contract.number, contract.date, contract.type, contract.status, contract.value, customer.name AS customer_name, " +
		"unit.code AS unit_code, project.code AS project_code, project.name AS project_name, " +
		"COALESCE(schedule.paid_total, 0) AS paid_total, COALESCE(schedule.installment_count, 0) AS installment_count, " +
		"COALESCE(schedule.paid_count, 0) AS paid_count, COALESCE(schedule.overdue_count, 0) AS overdue_count"
	contractScheduleJoin = "LEFT JOIN LATERAL (SELECT SUM(installment.amount) FILTER (WHERE installment.status = ?) AS paid_total, " +
		"COUNT(*) AS installment_count, COUNT(*) FILTER (WHERE installment.status = ?) AS paid_count, " +
		"COUNT(*) FILTER (WHERE installment.status <> ? AND installment.due_date < ?) AS overdue_count " +
		"FROM installment WHERE installment.contract_id = contract.id) AS schedule ON TRUE"
	installmentColumns = "contract.number AS contract_number, customer.name AS customer_name, installment.installment_number, " +
		"installment.due_date, installment.amount, installment.paid_date, installment.status"
)

type DatasetFilter struct {
	ProjectID *uuid.UUID
	From      *time.Time
	To        *time.Time
}

type ProjectLabel struct {
	Code string
	Name string
}

type CashMonthRow struct {
	Period  time.Time
	CashIn  int64
	CashOut int64
}

type CashLineRow struct {
	Date        time.Time
	Type        string
	Amount      int64
	Note        string
	AccountCode string
	AccountName string
	ProjectCode string
	ProjectName string
}

type BudgetRow struct {
	ProjectCode string
	ProjectName string
	Year        int
	Value       int64
	Realized    int64
}

type ProjectRow struct {
	Code            string
	Name            string
	Type            string
	Status          string
	Progress        int
	StartDate       *time.Time
	TargetEndDate   *time.Time
	ContractValue   int64
	BudgetItemTotal int64
}

type OpenItemRow struct {
	Reference   string
	PartyName   string
	ProjectCode string
	ProjectName string
	DueDate     time.Time
	Amount      int64
	PaidAmount  int64
}

type StockRow struct {
	WarehouseCode string
	WarehouseName string
	MaterialCode  string
	MaterialName  string
	Category      string
	UnitCode      string
	Quantity      float64
}

type LowStockRow struct {
	MaterialCode  string
	MaterialName  string
	UnitCode      string
	MinimumStock  float64
	TotalQuantity float64
}

type PayrollRow struct {
	Period         string
	IdentityNumber string
	EmployeeName   string
	Position       string
	ProjectCode    string
	ProjectName    string
	BasicPay       int64
	Allowance      int64
	Deduction      int64
	NetPay         int64
	Paid           bool
}

type ContractRow struct {
	Number           string
	Date             time.Time
	Type             string
	Status           string
	Value            int64
	CustomerName     string
	UnitCode         string
	ProjectCode      string
	ProjectName      string
	PaidTotal        int64
	InstallmentCount int64
	PaidCount        int64
	OverdueCount     int64
}

type InstallmentRow struct {
	ContractNumber    string
	CustomerName      string
	InstallmentNumber int
	DueDate           time.Time
	Amount            int64
	PaidDate          *time.Time
	Status            string
}

type openItemSource struct {
	table       string
	columns     string
	partyJoin   string
	projectJoin string
	open        string
	project     string
	dueFrom     string
	dueTo       string
	order       string
}

var (
	receivableSource = openItemSource{
		table: "receivable",
		columns: "receivable.reference, customer.name AS party_name, project.code AS project_code, project.name AS project_name, " +
			"receivable.due_date, receivable.amount, receivable.paid_amount",
		partyJoin:   "JOIN customer ON customer.id = receivable.customer_id",
		projectJoin: "LEFT JOIN project ON project.id = receivable.project_id",
		open:        "receivable.paid_amount < receivable.amount",
		project:     "receivable.project_id = ?",
		dueFrom:     "receivable.due_date >= ?",
		dueTo:       "receivable.due_date <= ?",
		order:       "receivable.due_date ASC, receivable.reference ASC",
	}
	payableSource = openItemSource{
		table: "payable",
		columns: "payable.reference, vendor.name AS party_name, project.code AS project_code, project.name AS project_name, " +
			"payable.due_date, payable.amount, payable.paid_amount",
		partyJoin:   "JOIN vendor ON vendor.id = payable.vendor_id",
		projectJoin: "LEFT JOIN project ON project.id = payable.project_id",
		open:        "payable.paid_amount < payable.amount",
		project:     "payable.project_id = ?",
		dueFrom:     "payable.due_date >= ?",
		dueTo:       "payable.due_date <= ?",
		order:       "payable.due_date ASC, payable.reference ASC",
	}
)

func (r *gormRepository) FindProjectLabel(ctx context.Context, id uuid.UUID) (ProjectLabel, error) {
	var label ProjectLabel
	err := r.db.WithContext(ctx).Table("project").Select("project.code, project.name").Where("project.id = ?", id).Take(&label).Error
	return label, database.Translate(err)
}

func (r *gormRepository) SumCashByMonth(ctx context.Context, filter DatasetFilter) ([]CashMonthRow, error) {
	var rows []CashMonthRow
	err := r.db.WithContext(ctx).Table("cash_transaction").
		Select("date_trunc('month', cash_transaction.date)::date AS period, "+
			"COALESCE(SUM(cash_transaction.amount) FILTER (WHERE cash_transaction.type = ?), 0) AS cash_in, "+
			"COALESCE(SUM(cash_transaction.amount) FILTER (WHERE cash_transaction.type = ?), 0) AS cash_out", cashIn, cashOut).
		Scopes(filter.cash).
		Group("period").
		Order("period ASC").
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) CashBalanceBefore(ctx context.Context, projectID *uuid.UUID, before time.Time) (int64, error) {
	var balance int64
	query := r.db.WithContext(ctx).Table("cash_transaction").
		Select("COALESCE(SUM(CASE WHEN cash_transaction.type = ? THEN cash_transaction.amount ELSE -cash_transaction.amount END), 0)", cashIn).
		Where("cash_transaction.date < ?", before)
	if projectID != nil {
		query = query.Where("cash_transaction.project_id = ?", *projectID)
	}
	err := query.Scan(&balance).Error
	return balance, database.Translate(err)
}

func (r *gormRepository) ListCashLines(ctx context.Context, filter DatasetFilter) ([]CashLineRow, error) {
	var rows []CashLineRow
	err := r.db.WithContext(ctx).Table("cash_transaction").
		Select(cashLineColumns).
		Joins("JOIN account ON account.id = cash_transaction.account_id").
		Joins("LEFT JOIN project ON project.id = cash_transaction.project_id").
		Scopes(filter.cash).
		Order("cash_transaction.date ASC, cash_transaction.created_at ASC").
		Limit(maximumRowCount).
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListBudgetRealization(ctx context.Context, filter DatasetFilter) ([]BudgetRow, error) {
	var rows []BudgetRow
	query := r.db.WithContext(ctx).Table("budget").
		Select(budgetColumns, cashOut).
		Joins("JOIN project ON project.id = budget.project_id")
	if filter.ProjectID != nil {
		query = query.Where("budget.project_id = ?", *filter.ProjectID)
	}
	if filter.From != nil {
		query = query.Where("budget.year >= ?", filter.From.Year())
	}
	if filter.To != nil {
		query = query.Where("budget.year <= ?", filter.To.Year())
	}
	err := query.Order("budget.year DESC, project.code ASC").Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListProjectProgress(ctx context.Context, filter DatasetFilter) ([]ProjectRow, error) {
	var rows []ProjectRow
	query := r.db.WithContext(ctx).Table("project").Select(projectColumns)
	if filter.ProjectID != nil {
		query = query.Where("project.id = ?", *filter.ProjectID)
	}
	if filter.To != nil {
		query = query.Where("(project.start_date IS NULL OR project.start_date <= ?)", *filter.To)
	}
	if filter.From != nil {
		query = query.Where("(project.target_end_date IS NULL OR project.target_end_date >= ?)", *filter.From)
	}
	err := query.Order("project.code ASC").Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListOpenReceivables(ctx context.Context, filter DatasetFilter) ([]OpenItemRow, error) {
	return r.listOpenItems(ctx, receivableSource, filter)
}

func (r *gormRepository) ListOpenPayables(ctx context.Context, filter DatasetFilter) ([]OpenItemRow, error) {
	return r.listOpenItems(ctx, payableSource, filter)
}

func (r *gormRepository) listOpenItems(ctx context.Context, source openItemSource, filter DatasetFilter) ([]OpenItemRow, error) {
	var rows []OpenItemRow
	query := r.db.WithContext(ctx).Table(source.table).
		Select(source.columns).
		Joins(source.partyJoin).
		Joins(source.projectJoin).
		Where(source.open)
	if filter.ProjectID != nil {
		query = query.Where(source.project, *filter.ProjectID)
	}
	if filter.From != nil {
		query = query.Where(source.dueFrom, *filter.From)
	}
	if filter.To != nil {
		query = query.Where(source.dueTo, *filter.To)
	}
	err := query.Order(source.order).Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListStockLevels(ctx context.Context, filter DatasetFilter) ([]StockRow, error) {
	var rows []StockRow
	query := r.db.WithContext(ctx).Table("stock").
		Select(stockColumns).
		Joins("JOIN warehouse ON warehouse.id = stock.warehouse_id").
		Joins("JOIN material ON material.id = stock.material_id").
		Joins("JOIN unit_of_measure ON unit_of_measure.id = material.unit_of_measure_id")
	if filter.ProjectID != nil {
		query = query.Where("warehouse.project_id = ?", *filter.ProjectID)
	}
	err := query.Order("warehouse.code ASC, material.code ASC").Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListLowStock(ctx context.Context, filter DatasetFilter) ([]LowStockRow, error) {
	totals := r.db.Table("stock").
		Select("stock.material_id, SUM(stock.quantity) AS quantity").
		Joins("JOIN warehouse ON warehouse.id = stock.warehouse_id").
		Group("stock.material_id")
	if filter.ProjectID != nil {
		totals = totals.Where("warehouse.project_id = ?", *filter.ProjectID)
	}
	var rows []LowStockRow
	err := r.db.WithContext(ctx).Table("material").
		Select(lowStockColumns).
		Joins("JOIN unit_of_measure ON unit_of_measure.id = material.unit_of_measure_id").
		Joins("LEFT JOIN (?) AS totals ON totals.material_id = material.id", totals).
		Where("material.minimum_stock > 0 AND COALESCE(totals.quantity, 0) < material.minimum_stock").
		Order("material.code ASC").
		Limit(maximumRowCount).
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListPayrolls(ctx context.Context, filter DatasetFilter) ([]PayrollRow, error) {
	var rows []PayrollRow
	query := r.db.WithContext(ctx).Table("payroll").
		Select(payrollColumns).
		Joins("JOIN employee ON employee.id = payroll.employee_id").
		Joins("LEFT JOIN project ON project.id = employee.project_id")
	if filter.ProjectID != nil {
		query = query.Where("employee.project_id = ?", *filter.ProjectID)
	}
	if filter.From != nil {
		query = query.Where("payroll.period >= ?", filter.From.Format(monthPeriod))
	}
	if filter.To != nil {
		query = query.Where("payroll.period <= ?", filter.To.Format(monthPeriod))
	}
	err := query.Order("payroll.period ASC, employee.name ASC").Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListContracts(ctx context.Context, filter DatasetFilter, today time.Time) ([]ContractRow, error) {
	var rows []ContractRow
	query := r.db.WithContext(ctx).Table("contract").
		Select(contractColumns).
		Joins("JOIN customer ON customer.id = contract.customer_id").
		Joins("JOIN unit ON unit.id = contract.unit_id").
		Joins("JOIN project ON project.id = unit.project_id").
		Joins(contractScheduleJoin, paidStatus, paidStatus, paidStatus, today).
		Scopes(filter.salesProject)
	if filter.From != nil {
		query = query.Where("contract.date >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("contract.date <= ?", *filter.To)
	}
	err := query.Order("contract.date ASC, contract.number ASC").Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) ListInstallments(ctx context.Context, filter DatasetFilter) ([]InstallmentRow, error) {
	var rows []InstallmentRow
	query := r.db.WithContext(ctx).Table("installment").
		Select(installmentColumns).
		Joins("JOIN contract ON contract.id = installment.contract_id").
		Joins("JOIN customer ON customer.id = contract.customer_id").
		Joins("JOIN unit ON unit.id = contract.unit_id").
		Scopes(filter.salesProject)
	if filter.From != nil {
		query = query.Where("installment.due_date >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("installment.due_date <= ?", *filter.To)
	}
	err := query.Order("installment.due_date ASC, contract.number ASC, installment.installment_number ASC").
		Limit(maximumRowCount).Scan(&rows).Error
	return rows, database.Translate(err)
}

func (f DatasetFilter) cash(db *gorm.DB) *gorm.DB {
	if f.ProjectID != nil {
		db = db.Where("cash_transaction.project_id = ?", *f.ProjectID)
	}
	if f.From != nil {
		db = db.Where("cash_transaction.date >= ?", *f.From)
	}
	if f.To != nil {
		db = db.Where("cash_transaction.date <= ?", *f.To)
	}
	return db
}

func (f DatasetFilter) salesProject(db *gorm.DB) *gorm.DB {
	if f.ProjectID != nil {
		db = db.Where("unit.project_id = ?", *f.ProjectID)
	}
	return db
}
