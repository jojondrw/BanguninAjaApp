package finance

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type BudgetRequest struct {
	ProjectID uuid.UUID `json:"projectId" binding:"required"`
	Year      int       `json:"year" binding:"required,min=2000,max=2100"`
	Value     int64     `json:"value" binding:"min=0"`
	Note      string    `json:"note" binding:"max=200"`
}

type BudgetQuery struct {
	pagination.Query
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Year      int        `form:"year" binding:"omitempty,min=2000,max=2100"`
}

type BudgetResponse struct {
	ID         uuid.UUID `json:"id"`
	ProjectID  uuid.UUID `json:"projectId"`
	Year       int       `json:"year"`
	Value      int64     `json:"value"`
	Note       string    `json:"note"`
	Realized   int64     `json:"realized"`
	Remaining  int64     `json:"remaining"`
	Absorption int       `json:"absorption"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CashTransactionRequest struct {
	Date      time.Time  `json:"date" binding:"required"`
	Type      string     `json:"type" binding:"required,oneof=in out"`
	AccountID uuid.UUID  `json:"accountId" binding:"required"`
	ProjectID *uuid.UUID `json:"projectId"`
	Amount    int64      `json:"amount" binding:"required,gt=0"`
	Note      string     `json:"note" binding:"max=200"`
}

type CashTransactionQuery struct {
	pagination.Query
	Type      string     `form:"type" binding:"omitempty,oneof=in out"`
	AccountID *uuid.UUID `form:"accountId,parser=encoding.TextUnmarshaler"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	DateFrom  *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo    *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type CashTransactionResponse struct {
	ID        uuid.UUID  `json:"id"`
	Date      time.Time  `json:"date"`
	Type      string     `json:"type"`
	AccountID uuid.UUID  `json:"accountId"`
	ProjectID *uuid.UUID `json:"projectId"`
	Amount    int64      `json:"amount"`
	Note      string     `json:"note"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type CashFlowQuery struct {
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	DateFrom  *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo    *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type CashFlowPeriodResponse struct {
	Period  string `json:"period"`
	CashIn  int64  `json:"cashIn"`
	CashOut int64  `json:"cashOut"`
	Net     int64  `json:"net"`
}

type CashFlowResponse struct {
	DateFrom time.Time                `json:"dateFrom"`
	DateTo   time.Time                `json:"dateTo"`
	Periods  []CashFlowPeriodResponse `json:"periods"`
	TotalIn  int64                    `json:"totalIn"`
	TotalOut int64                    `json:"totalOut"`
	Net      int64                    `json:"net"`
	Balance  int64                    `json:"balance"`
}

type JournalLineRequest struct {
	AccountID uuid.UUID `json:"accountId" binding:"required"`
	Debit     int64     `json:"debit" binding:"min=0"`
	Credit    int64     `json:"credit" binding:"min=0"`
}

type JournalEntryRequest struct {
	Number string               `json:"number" binding:"required,max=40"`
	Date   time.Time            `json:"date" binding:"required"`
	Note   string               `json:"note" binding:"max=200"`
	Source string               `json:"source" binding:"max=40"`
	Lines  []JournalLineRequest `json:"lines" binding:"required,min=2,dive"`
}

type JournalEntryQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=200"`
	AccountID *uuid.UUID `form:"accountId,parser=encoding.TextUnmarshaler"`
	DateFrom  *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo    *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type JournalEntryResponse struct {
	ID        uuid.UUID `json:"id"`
	Number    string    `json:"number"`
	Date      time.Time `json:"date"`
	Note      string    `json:"note"`
	Source    string    `json:"source"`
	Total     int64     `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
}

type JournalLineResponse struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"accountId"`
	Debit     int64     `json:"debit"`
	Credit    int64     `json:"credit"`
}

type JournalEntryDetailResponse struct {
	JournalEntryResponse
	Lines []JournalLineResponse `json:"lines"`
}

type LedgerQuery struct {
	pagination.Query
	AccountID *uuid.UUID `form:"accountId,parser=encoding.TextUnmarshaler" binding:"required"`
	DateFrom  *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo    *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type LedgerLineResponse struct {
	JournalEntryID uuid.UUID `json:"journalEntryId"`
	Number         string    `json:"number"`
	Date           time.Time `json:"date"`
	Note           string    `json:"note"`
	Debit          int64     `json:"debit"`
	Credit         int64     `json:"credit"`
	Balance        int64     `json:"balance"`
}

type LedgerPage struct {
	pagination.Page[LedgerLineResponse]
	OpeningBalance int64 `json:"openingBalance"`
	TotalDebit     int64 `json:"totalDebit"`
	TotalCredit    int64 `json:"totalCredit"`
	ClosingBalance int64 `json:"closingBalance"`
}

func newBudgetResponse(row BudgetRow) BudgetResponse {
	return BudgetResponse{
		ID:         row.ID,
		ProjectID:  row.ProjectID,
		Year:       row.Year,
		Value:      row.Value,
		Note:       row.Note,
		Realized:   row.Realized,
		Remaining:  row.Value - row.Realized,
		Absorption: absorption(row.Realized, row.Value),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func newCashTransactionResponse(transaction CashTransaction) CashTransactionResponse {
	return CashTransactionResponse{
		ID:        transaction.ID,
		Date:      transaction.Date,
		Type:      transaction.Type,
		AccountID: transaction.AccountID,
		ProjectID: transaction.ProjectID,
		Amount:    transaction.Amount,
		Note:      transaction.Note,
		CreatedAt: transaction.CreatedAt,
		UpdatedAt: transaction.UpdatedAt,
	}
}

func newJournalEntryResponse(row JournalEntryRow) JournalEntryResponse {
	return JournalEntryResponse{
		ID:        row.ID,
		Number:    row.Number,
		Date:      row.Date,
		Note:      row.Note,
		Source:    row.Source,
		Total:     row.Total,
		CreatedAt: row.CreatedAt,
	}
}

func newJournalLineResponse(line JournalLine) JournalLineResponse {
	return JournalLineResponse{ID: line.ID, AccountID: line.AccountID, Debit: line.Debit, Credit: line.Kredit}
}
