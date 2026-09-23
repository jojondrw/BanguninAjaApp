package finance

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	cashOut = "out"
	cashIn  = "in"

	budgetColumns = "budget.id, budget.project_id, budget.year, budget.value, budget.note, budget.created_at, budget.updated_at, " +
		"COALESCE((SELECT SUM(cash_transaction.amount) FROM cash_transaction " +
		"WHERE cash_transaction.project_id = budget.project_id AND cash_transaction.type = ? " +
		"AND cash_transaction.date >= make_date(budget.year, 1, 1) AND cash_transaction.date < make_date(budget.year + 1, 1, 1)), 0) AS realized"
	journalColumns = "journal_entry.id, journal_entry.number, journal_entry.date, journal_entry.note, journal_entry.source, journal_entry.created_at, " +
		"(SELECT COALESCE(SUM(journal_line.debit), 0) FROM journal_line WHERE journal_line.journal_entry_id = journal_entry.id) AS total"
	ledgerColumns = "journal_entry.id AS journal_entry_id, journal_entry.number, journal_entry.date, journal_entry.note, " +
		"journal_line.id AS line_id, journal_line.debit, journal_line.kredit AS credit, " +
		"SUM(journal_line.debit - journal_line.kredit) OVER (ORDER BY journal_entry.date, journal_entry.number, journal_line.id) AS running_balance"
)

type BudgetFilter struct {
	ProjectID *uuid.UUID
	Year      int
	Offset    int
	Limit     int
}

type CashTransactionFilter struct {
	Type      string
	AccountID *uuid.UUID
	ProjectID *uuid.UUID
	DateFrom  *time.Time
	DateTo    *time.Time
	Offset    int
	Limit     int
}

type CashFlowFilter struct {
	ProjectID *uuid.UUID
	DateFrom  time.Time
	DateTo    time.Time
}

type JournalEntryFilter struct {
	Search    string
	AccountID *uuid.UUID
	DateFrom  *time.Time
	DateTo    *time.Time
	Offset    int
	Limit     int
}

type LedgerFilter struct {
	AccountID uuid.UUID
	DateFrom  *time.Time
	DateTo    *time.Time
	Offset    int
	Limit     int
}

type BudgetRow struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Year      int
	Value     int64
	Note      string
	Realized  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CashFlowRow struct {
	Period  time.Time
	CashIn  int64
	CashOut int64
}

type JournalEntryRow struct {
	ID        uuid.UUID
	Number    string
	Date      time.Time
	Note      string
	Source    string
	Total     int64
	CreatedAt time.Time
}

type LedgerRow struct {
	JournalEntryID uuid.UUID
	Number         string
	Date           time.Time
	Note           string
	Debit          int64
	Credit         int64
	RunningBalance int64
}

type LedgerTotals struct {
	Debit  int64
	Credit int64
}

type Repository interface {
	Transaction(ctx context.Context, work func(Repository) error) error

	ListBudgets(ctx context.Context, filter BudgetFilter) ([]BudgetRow, int64, error)
	FindBudgetRow(ctx context.Context, id uuid.UUID) (BudgetRow, error)
	FindBudget(ctx context.Context, id uuid.UUID) (Budget, error)
	CreateBudget(ctx context.Context, budget *Budget) error
	SaveBudget(ctx context.Context, budget *Budget) error
	DeleteBudget(ctx context.Context, id uuid.UUID) error

	ListCashTransactions(ctx context.Context, filter CashTransactionFilter) ([]CashTransaction, int64, error)
	FindCashTransaction(ctx context.Context, id uuid.UUID) (CashTransaction, error)
	CreateCashTransaction(ctx context.Context, transaction *CashTransaction) error
	SaveCashTransaction(ctx context.Context, transaction *CashTransaction) error
	DeleteCashTransaction(ctx context.Context, id uuid.UUID) error
	SumCashFlowByMonth(ctx context.Context, filter CashFlowFilter) ([]CashFlowRow, error)
	CashBalance(ctx context.Context, projectID *uuid.UUID, until time.Time) (int64, error)

	ListJournalEntries(ctx context.Context, filter JournalEntryFilter) ([]JournalEntryRow, int64, error)
	FindJournalEntryRow(ctx context.Context, id uuid.UUID) (JournalEntryRow, error)
	ListJournalLines(ctx context.Context, journalEntryID uuid.UUID) ([]JournalLine, error)
	CreateJournalEntry(ctx context.Context, entry *JournalEntry) error
	CreateJournalLines(ctx context.Context, lines []JournalLine) error

	ListLedgerLines(ctx context.Context, filter LedgerFilter) ([]LedgerRow, int64, error)
	LedgerBalanceBefore(ctx context.Context, accountID uuid.UUID, before time.Time) (int64, error)
	SumLedger(ctx context.Context, filter LedgerFilter) (LedgerTotals, error)
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

func (r *gormRepository) ListBudgets(ctx context.Context, filter BudgetFilter) ([]BudgetRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Budget{}).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]BudgetRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).
		Model(&Budget{}).
		Scopes(filter.apply).
		Select(budgetColumns, cashOut).
		Order("budget.year DESC, budget.created_at DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindBudgetRow(ctx context.Context, id uuid.UUID) (BudgetRow, error) {
	var row BudgetRow
	err := r.db.WithContext(ctx).Model(&Budget{}).Select(budgetColumns, cashOut).Where("budget.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) FindBudget(ctx context.Context, id uuid.UUID) (Budget, error) {
	var budget Budget
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&budget).Error
	return budget, database.Translate(err)
}

func (r *gormRepository) CreateBudget(ctx context.Context, budget *Budget) error {
	return database.Translate(r.db.WithContext(ctx).Create(budget).Error)
}

func (r *gormRepository) SaveBudget(ctx context.Context, budget *Budget) error {
	return database.Translate(r.db.WithContext(ctx).Save(budget).Error)
}

func (r *gormRepository) DeleteBudget(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Budget{}, id)
}

func (r *gormRepository) ListCashTransactions(ctx context.Context, filter CashTransactionFilter) ([]CashTransaction, int64, error) {
	return database.FindPage[CashTransaction](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "date DESC, created_at DESC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) FindCashTransaction(ctx context.Context, id uuid.UUID) (CashTransaction, error) {
	var transaction CashTransaction
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&transaction).Error
	return transaction, database.Translate(err)
}

func (r *gormRepository) CreateCashTransaction(ctx context.Context, transaction *CashTransaction) error {
	return database.Translate(r.db.WithContext(ctx).Create(transaction).Error)
}

func (r *gormRepository) SaveCashTransaction(ctx context.Context, transaction *CashTransaction) error {
	return database.Translate(r.db.WithContext(ctx).Save(transaction).Error)
}

func (r *gormRepository) DeleteCashTransaction(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &CashTransaction{}, id)
}

func (r *gormRepository) SumCashFlowByMonth(ctx context.Context, filter CashFlowFilter) ([]CashFlowRow, error) {
	var rows []CashFlowRow
	err := r.db.WithContext(ctx).
		Model(&CashTransaction{}).
		Scopes(filter.apply).
		Select("date_trunc('month', date)::date AS period, "+
			"COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) AS cash_in, "+
			"COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) AS cash_out", cashIn, cashOut).
		Group("period").
		Order("period ASC").
		Scan(&rows).Error
	return rows, database.Translate(err)
}

func (r *gormRepository) CashBalance(ctx context.Context, projectID *uuid.UUID, until time.Time) (int64, error) {
	var balance int64
	query := r.db.WithContext(ctx).
		Model(&CashTransaction{}).
		Select("COALESCE(SUM(CASE WHEN type = ? THEN amount ELSE -amount END), 0)", cashIn).
		Where("date <= ?", until)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	err := query.Scan(&balance).Error
	return balance, database.Translate(err)
}

func (r *gormRepository) ListJournalEntries(ctx context.Context, filter JournalEntryFilter) ([]JournalEntryRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&JournalEntry{}).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]JournalEntryRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).
		Model(&JournalEntry{}).
		Scopes(filter.apply).
		Select(journalColumns).
		Order("journal_entry.date DESC, journal_entry.number DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindJournalEntryRow(ctx context.Context, id uuid.UUID) (JournalEntryRow, error) {
	var row JournalEntryRow
	err := r.db.WithContext(ctx).Model(&JournalEntry{}).Select(journalColumns).Where("journal_entry.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) ListJournalLines(ctx context.Context, journalEntryID uuid.UUID) ([]JournalLine, error) {
	var lines []JournalLine
	err := r.db.WithContext(ctx).
		Where("journal_entry_id = ?", journalEntryID).
		Order("created_at ASC, id ASC").
		Find(&lines).Error
	return lines, database.Translate(err)
}

func (r *gormRepository) CreateJournalEntry(ctx context.Context, entry *JournalEntry) error {
	return database.Translate(r.db.WithContext(ctx).Create(entry).Error)
}

func (r *gormRepository) CreateJournalLines(ctx context.Context, lines []JournalLine) error {
	return database.Translate(r.db.WithContext(ctx).Create(&lines).Error)
}

func (r *gormRepository) ListLedgerLines(ctx context.Context, filter LedgerFilter) ([]LedgerRow, int64, error) {
	var total int64
	if err := r.ledgerLines(ctx, filter).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]LedgerRow, 0, filter.Limit)
	err := r.db.WithContext(ctx).
		Table("(?) AS ledger", r.ledgerLines(ctx, filter).Select(ledgerColumns)).
		Order("ledger.date ASC, ledger.number ASC, ledger.line_id ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) LedgerBalanceBefore(ctx context.Context, accountID uuid.UUID, before time.Time) (int64, error) {
	var balance int64
	err := r.db.WithContext(ctx).
		Table("journal_line").
		Joins("JOIN journal_entry ON journal_entry.id = journal_line.journal_entry_id").
		Where("journal_line.account_id = ? AND journal_entry.date < ?", accountID, before).
		Select("COALESCE(SUM(journal_line.debit - journal_line.kredit), 0)").
		Scan(&balance).Error
	return balance, database.Translate(err)
}

func (r *gormRepository) SumLedger(ctx context.Context, filter LedgerFilter) (LedgerTotals, error) {
	var totals LedgerTotals
	err := r.ledgerLines(ctx, filter).
		Select("COALESCE(SUM(journal_line.debit), 0) AS debit, COALESCE(SUM(journal_line.kredit), 0) AS credit").
		Scan(&totals).Error
	return totals, database.Translate(err)
}

func (r *gormRepository) ledgerLines(ctx context.Context, filter LedgerFilter) *gorm.DB {
	query := r.db.WithContext(ctx).
		Table("journal_line").
		Joins("JOIN journal_entry ON journal_entry.id = journal_line.journal_entry_id").
		Where("journal_line.account_id = ?", filter.AccountID)
	if filter.DateFrom != nil {
		query = query.Where("journal_entry.date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("journal_entry.date <= ?", *filter.DateTo)
	}
	return query
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

func (f BudgetFilter) apply(db *gorm.DB) *gorm.DB {
	if f.ProjectID != nil {
		db = db.Where("budget.project_id = ?", *f.ProjectID)
	}
	if f.Year != 0 {
		db = db.Where("budget.year = ?", f.Year)
	}
	return db
}

func (f CashTransactionFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Type != "" {
		db = db.Where("type = ?", f.Type)
	}
	if f.AccountID != nil {
		db = db.Where("account_id = ?", *f.AccountID)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	if f.DateFrom != nil {
		db = db.Where("date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("date <= ?", *f.DateTo)
	}
	return db
}

func (f CashFlowFilter) apply(db *gorm.DB) *gorm.DB {
	db = db.Where("date >= ? AND date <= ?", f.DateFrom, f.DateTo)
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	return db
}

func (f JournalEntryFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(journal_entry.number ILIKE ? OR journal_entry.note ILIKE ?)", pattern, pattern)
	}
	if f.AccountID != nil {
		db = db.Where("EXISTS (SELECT 1 FROM journal_line WHERE journal_line.journal_entry_id = journal_entry.id AND journal_line.account_id = ?)", *f.AccountID)
	}
	if f.DateFrom != nil {
		db = db.Where("journal_entry.date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("journal_entry.date <= ?", *f.DateTo)
	}
	return db
}
