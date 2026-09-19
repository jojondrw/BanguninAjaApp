package finance

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Budget struct {
	entity.Base
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index:idx_budget_project"`
	Year      int       `gorm:"not null;index:idx_budget_tahun"`
	Value     int64     `gorm:"not null;default:0"`
	Note      string    `gorm:"type:varchar(200)"`
}

func (Budget) TableName() string {
	return "budget"
}

type CashTransaction struct {
	entity.Base
	Date      time.Time  `gorm:"type:date;not null;index:idx_cash_transaction_date"`
	Type      string     `gorm:"type:varchar(20);not null;index:idx_cash_transaction_type"`
	AccountID uuid.UUID  `gorm:"type:uuid;not null;index:idx_cash_transaction_account"`
	ProjectID *uuid.UUID `gorm:"type:uuid;index:idx_cash_transaction_project"`
	Amount    int64      `gorm:"not null;default:0"`
	Note      string     `gorm:"type:varchar(200)"`
}

func (CashTransaction) TableName() string {
	return "cash_transaction"
}

type JournalEntry struct {
	entity.Base
	Number string    `gorm:"type:varchar(40);not null;uniqueIndex:uq_journal_number"`
	Date   time.Time `gorm:"type:date;not null;index:idx_journal_date"`
	Note   string    `gorm:"type:varchar(200)"`
	Source string    `gorm:"type:varchar(40)"`
}

func (JournalEntry) TableName() string {
	return "journal_entry"
}

type JournalLine struct {
	entity.Base
	JournalEntryID uuid.UUID `gorm:"type:uuid;not null;index:idx_journal_detail_journal"`
	AccountID      uuid.UUID `gorm:"type:uuid;not null;index:idx_journal_detail_account"`
	Debit          int64     `gorm:"not null;default:0"`
	Kredit         int64     `gorm:"not null;default:0"`
}

func (JournalLine) TableName() string {
	return "journal_line"
}

func Entities() []any {
	return []any{&Budget{}, &CashTransaction{}, &JournalEntry{}, &JournalLine{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_journal_detail_account_journal ON journal_line (account_id, journal_entry_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cash_transaction_project_date ON cash_transaction (project_id, date DESC)`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("budget", "project_id", "project", database.DeleteCascade),
		database.ForeignKey("cash_transaction", "account_id", "account", database.DeleteRestrict),
		database.ForeignKey("cash_transaction", "project_id", "project", database.DeleteSetNull),
		database.ForeignKey("journal_line", "journal_entry_id", "journal_entry", database.DeleteCascade),
		database.ForeignKey("journal_line", "account_id", "account", database.DeleteRestrict),
		database.Check("budget", "value", "value >= 0"),
		database.Check("budget", "year", "year BETWEEN 2000 AND 2100"),
		database.Check("cash_transaction", "type", "type IN ('in','out')"),
		database.Check("cash_transaction", "amount", "amount > 0"),
		database.Check("journal_line", "value", "debit >= 0 AND kredit >= 0"),
		database.Check("journal_line", "single_sided", "(debit = 0) <> (kredit = 0)"),
		database.Unique("budget", "project_year", "project_id, year"),
	}
}
