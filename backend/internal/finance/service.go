package finance

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	periodLayout          = "2006-01"
	defaultCashFlowMonths = 6
	fullPercent           = 100
)

var (
	errBudgetNotFound       = apperror.NotFound("budget_not_found", "Anggaran tidak ditemukan")
	errBudgetYearUsed       = apperror.Conflict("budget_year_used", "Proyek ini sudah punya anggaran untuk tahun tersebut")
	errBudgetProjectMissing = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")

	errCashTransactionNotFound  = apperror.NotFound("cash_transaction_not_found", "Transaksi kas tidak ditemukan")
	errCashTransactionReference = apperror.Unprocessable("cash_transaction_reference_not_found", "Akun atau proyek tidak ditemukan")
	errInvalidDateRange         = apperror.Unprocessable("invalid_date_range", "Tanggal selesai tidak boleh lebih awal dari tanggal mulai")

	errJournalNotFound     = apperror.NotFound("journal_entry_not_found", "Jurnal tidak ditemukan")
	errJournalNumberUsed   = apperror.Conflict("journal_number_used", "Nomor jurnal sudah dipakai")
	errJournalAccount      = apperror.Unprocessable("account_not_found", "Akun tidak ditemukan")
	errJournalSingleSided  = apperror.Unprocessable("journal_line_single_sided", "Setiap baris jurnal hanya boleh berisi debit atau kredit, dan nilainya lebih dari nol")
	errJournalUnbalanced   = apperror.Unprocessable("journal_unbalanced", "Total debit dan total kredit jurnal harus sama")
	errJournalLineRejected = apperror.Unprocessable("journal_line_invalid", "Baris jurnal ditolak karena nilainya tidak valid")
)

var (
	budgetReadErrors  = database.ErrorMap{NotFound: errBudgetNotFound}
	budgetWriteErrors = database.ErrorMap{NotFound: errBudgetNotFound, Duplicate: errBudgetYearUsed, Referenced: errBudgetProjectMissing}

	cashReadErrors  = database.ErrorMap{NotFound: errCashTransactionNotFound}
	cashWriteErrors = database.ErrorMap{NotFound: errCashTransactionNotFound, Referenced: errCashTransactionReference}

	journalReadErrors  = database.ErrorMap{NotFound: errJournalNotFound}
	journalWriteErrors = database.ErrorMap{Duplicate: errJournalNumberUsed, Referenced: errJournalAccount, Invalid: errJournalLineRejected}
)

type Service interface {
	ListBudgets(ctx context.Context, query BudgetQuery) (pagination.Page[BudgetResponse], error)
	GetBudget(ctx context.Context, id uuid.UUID) (BudgetResponse, error)
	CreateBudget(ctx context.Context, request BudgetRequest) (BudgetResponse, error)
	UpdateBudget(ctx context.Context, id uuid.UUID, request BudgetRequest) (BudgetResponse, error)
	DeleteBudget(ctx context.Context, id uuid.UUID) error

	ListCashTransactions(ctx context.Context, query CashTransactionQuery) (pagination.Page[CashTransactionResponse], error)
	GetCashTransaction(ctx context.Context, id uuid.UUID) (CashTransactionResponse, error)
	CreateCashTransaction(ctx context.Context, request CashTransactionRequest) (CashTransactionResponse, error)
	UpdateCashTransaction(ctx context.Context, id uuid.UUID, request CashTransactionRequest) (CashTransactionResponse, error)
	DeleteCashTransaction(ctx context.Context, id uuid.UUID) error
	SummarizeCashFlow(ctx context.Context, query CashFlowQuery) (CashFlowResponse, error)

	ListJournalEntries(ctx context.Context, query JournalEntryQuery) (pagination.Page[JournalEntryResponse], error)
	GetJournalEntry(ctx context.Context, id uuid.UUID) (JournalEntryDetailResponse, error)
	RecordJournalEntry(ctx context.Context, request JournalEntryRequest) (JournalEntryDetailResponse, error)

	ReadLedger(ctx context.Context, query LedgerQuery) (LedgerPage, error)
}

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) ListBudgets(ctx context.Context, query BudgetQuery) (pagination.Page[BudgetResponse], error) {
	rows, total, err := s.repository.ListBudgets(ctx, BudgetFilter{
		ProjectID: query.ProjectID,
		Year:      query.Year,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[BudgetResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newBudgetResponse), query.Query, total), nil
}

func (s *service) GetBudget(ctx context.Context, id uuid.UUID) (BudgetResponse, error) {
	row, err := s.repository.FindBudgetRow(ctx, id)
	if err != nil {
		return BudgetResponse{}, budgetReadErrors.Resolve(err)
	}
	return newBudgetResponse(row), nil
}

func (s *service) CreateBudget(ctx context.Context, request BudgetRequest) (BudgetResponse, error) {
	var budget Budget
	applyBudgetRequest(&budget, request)
	if err := s.repository.CreateBudget(ctx, &budget); err != nil {
		return BudgetResponse{}, budgetWriteErrors.Resolve(err)
	}
	return s.GetBudget(ctx, budget.ID)
}

func (s *service) UpdateBudget(ctx context.Context, id uuid.UUID, request BudgetRequest) (BudgetResponse, error) {
	budget, err := s.repository.FindBudget(ctx, id)
	if err != nil {
		return BudgetResponse{}, budgetReadErrors.Resolve(err)
	}

	applyBudgetRequest(&budget, request)
	if err := s.repository.SaveBudget(ctx, &budget); err != nil {
		return BudgetResponse{}, budgetWriteErrors.Resolve(err)
	}
	return s.GetBudget(ctx, id)
}

func (s *service) DeleteBudget(ctx context.Context, id uuid.UUID) error {
	return budgetReadErrors.Resolve(s.repository.DeleteBudget(ctx, id))
}

func (s *service) ListCashTransactions(ctx context.Context, query CashTransactionQuery) (pagination.Page[CashTransactionResponse], error) {
	transactions, total, err := s.repository.ListCashTransactions(ctx, CashTransactionFilter{
		Type:      query.Type,
		AccountID: query.AccountID,
		ProjectID: query.ProjectID,
		DateFrom:  query.DateFrom,
		DateTo:    query.DateTo,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[CashTransactionResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(transactions, newCashTransactionResponse), query.Query, total), nil
}

func (s *service) GetCashTransaction(ctx context.Context, id uuid.UUID) (CashTransactionResponse, error) {
	transaction, err := s.repository.FindCashTransaction(ctx, id)
	if err != nil {
		return CashTransactionResponse{}, cashReadErrors.Resolve(err)
	}
	return newCashTransactionResponse(transaction), nil
}

func (s *service) CreateCashTransaction(ctx context.Context, request CashTransactionRequest) (CashTransactionResponse, error) {
	var transaction CashTransaction
	applyCashTransactionRequest(&transaction, request)
	if err := s.repository.CreateCashTransaction(ctx, &transaction); err != nil {
		return CashTransactionResponse{}, cashWriteErrors.Resolve(err)
	}
	return newCashTransactionResponse(transaction), nil
}

func (s *service) UpdateCashTransaction(ctx context.Context, id uuid.UUID, request CashTransactionRequest) (CashTransactionResponse, error) {
	transaction, err := s.repository.FindCashTransaction(ctx, id)
	if err != nil {
		return CashTransactionResponse{}, cashReadErrors.Resolve(err)
	}

	applyCashTransactionRequest(&transaction, request)
	if err := s.repository.SaveCashTransaction(ctx, &transaction); err != nil {
		return CashTransactionResponse{}, cashWriteErrors.Resolve(err)
	}
	return newCashTransactionResponse(transaction), nil
}

func (s *service) DeleteCashTransaction(ctx context.Context, id uuid.UUID) error {
	return cashReadErrors.Resolve(s.repository.DeleteCashTransaction(ctx, id))
}

func (s *service) SummarizeCashFlow(ctx context.Context, query CashFlowQuery) (CashFlowResponse, error) {
	from, to := cashFlowRange(query, s.today())
	if to.Before(from) {
		return CashFlowResponse{}, errInvalidDateRange
	}

	rows, err := s.repository.SumCashFlowByMonth(ctx, CashFlowFilter{ProjectID: query.ProjectID, DateFrom: from, DateTo: to})
	if err != nil {
		return CashFlowResponse{}, apperror.Internal(err)
	}

	balance, err := s.repository.CashBalance(ctx, query.ProjectID, to)
	if err != nil {
		return CashFlowResponse{}, apperror.Internal(err)
	}
	return newCashFlowResponse(from, to, rows, balance), nil
}

func (s *service) ListJournalEntries(ctx context.Context, query JournalEntryQuery) (pagination.Page[JournalEntryResponse], error) {
	rows, total, err := s.repository.ListJournalEntries(ctx, JournalEntryFilter{
		Search:    query.Search,
		AccountID: query.AccountID,
		DateFrom:  query.DateFrom,
		DateTo:    query.DateTo,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[JournalEntryResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newJournalEntryResponse), query.Query, total), nil
}

func (s *service) GetJournalEntry(ctx context.Context, id uuid.UUID) (JournalEntryDetailResponse, error) {
	row, err := s.repository.FindJournalEntryRow(ctx, id)
	if err != nil {
		return JournalEntryDetailResponse{}, journalReadErrors.Resolve(err)
	}

	lines, err := s.repository.ListJournalLines(ctx, id)
	if err != nil {
		return JournalEntryDetailResponse{}, apperror.Internal(err)
	}
	return JournalEntryDetailResponse{
		JournalEntryResponse: newJournalEntryResponse(row),
		Lines:                pagination.Map(lines, newJournalLineResponse),
	}, nil
}

func (s *service) RecordJournalEntry(ctx context.Context, request JournalEntryRequest) (JournalEntryDetailResponse, error) {
	if err := validateJournalLines(request.Lines); err != nil {
		return JournalEntryDetailResponse{}, err
	}

	entry := newJournalEntry(request)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return recordJournalEntry(ctx, repository, &entry, request.Lines)
	})
	if err != nil {
		return JournalEntryDetailResponse{}, apperror.From(err)
	}
	return s.GetJournalEntry(ctx, entry.ID)
}

func (s *service) ReadLedger(ctx context.Context, query LedgerQuery) (LedgerPage, error) {
	if query.DateFrom != nil && query.DateTo != nil && query.DateTo.Before(*query.DateFrom) {
		return LedgerPage{}, errInvalidDateRange
	}

	filter := LedgerFilter{AccountID: *query.AccountID, DateFrom: query.DateFrom, DateTo: query.DateTo, Offset: query.Offset(), Limit: query.Size()}
	opening, err := s.openingBalance(ctx, filter)
	if err != nil {
		return LedgerPage{}, err
	}

	rows, total, err := s.repository.ListLedgerLines(ctx, filter)
	if err != nil {
		return LedgerPage{}, apperror.Internal(err)
	}

	totals, err := s.repository.SumLedger(ctx, filter)
	if err != nil {
		return LedgerPage{}, apperror.Internal(err)
	}
	return LedgerPage{
		Page:           pagination.New(ledgerLines(rows, opening), query.Query, total),
		OpeningBalance: opening,
		TotalDebit:     totals.Debit,
		TotalCredit:    totals.Credit,
		ClosingBalance: opening + totals.Debit - totals.Credit,
	}, nil
}

func (s *service) openingBalance(ctx context.Context, filter LedgerFilter) (int64, error) {
	if filter.DateFrom == nil {
		return 0, nil
	}

	balance, err := s.repository.LedgerBalanceBefore(ctx, filter.AccountID, *filter.DateFrom)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return balance, nil
}

func (s *service) today() time.Time {
	year, month, day := s.now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func recordJournalEntry(ctx context.Context, repository Repository, entry *JournalEntry, requests []JournalLineRequest) error {
	if err := repository.CreateJournalEntry(ctx, entry); err != nil {
		return journalWriteErrors.Resolve(err)
	}
	return journalWriteErrors.Resolve(repository.CreateJournalLines(ctx, newJournalLines(entry.ID, requests)))
}

func validateJournalLines(lines []JournalLineRequest) error {
	var debit, credit int64
	for _, line := range lines {
		if !lineSingleSided(line) {
			return errJournalSingleSided
		}
		debit += line.Debit
		credit += line.Credit
	}
	if debit != credit {
		return errJournalUnbalanced
	}
	return nil
}

func lineSingleSided(line JournalLineRequest) bool {
	return (line.Debit > 0) != (line.Credit > 0)
}

func cashFlowRange(query CashFlowQuery, today time.Time) (time.Time, time.Time) {
	to := today
	if query.DateTo != nil {
		to = *query.DateTo
	}
	from := firstOfMonth(to).AddDate(0, 1-defaultCashFlowMonths, 0)
	if query.DateFrom != nil {
		from = *query.DateFrom
	}
	return from, to
}

func firstOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func newCashFlowResponse(from, to time.Time, rows []CashFlowRow, balance int64) CashFlowResponse {
	byPeriod := make(map[string]CashFlowRow, len(rows))
	for _, row := range rows {
		byPeriod[row.Period.Format(periodLayout)] = row
	}

	response := CashFlowResponse{DateFrom: from, DateTo: to, Balance: balance}
	for month := firstOfMonth(from); !month.After(to); month = month.AddDate(0, 1, 0) {
		row := byPeriod[month.Format(periodLayout)]
		response.Periods = append(response.Periods, CashFlowPeriodResponse{
			Period:  month.Format(periodLayout),
			CashIn:  row.CashIn,
			CashOut: row.CashOut,
			Net:     row.CashIn - row.CashOut,
		})
		response.TotalIn += row.CashIn
		response.TotalOut += row.CashOut
	}
	response.Net = response.TotalIn - response.TotalOut
	return response
}

func ledgerLines(rows []LedgerRow, opening int64) []LedgerLineResponse {
	return pagination.Map(rows, func(row LedgerRow) LedgerLineResponse {
		return LedgerLineResponse{
			JournalEntryID: row.JournalEntryID,
			Number:         row.Number,
			Date:           row.Date,
			Note:           row.Note,
			Debit:          row.Debit,
			Credit:         row.Credit,
			Balance:        opening + row.RunningBalance,
		}
	})
}

func absorption(realized, value int64) int {
	if value <= 0 {
		return 0
	}
	return int(math.Round(float64(realized) * fullPercent / float64(value)))
}

func newJournalEntry(request JournalEntryRequest) JournalEntry {
	return JournalEntry{
		Number: strings.TrimSpace(request.Number),
		Date:   request.Date,
		Note:   strings.TrimSpace(request.Note),
		Source: strings.TrimSpace(request.Source),
	}
}

func newJournalLines(entryID uuid.UUID, requests []JournalLineRequest) []JournalLine {
	lines := make([]JournalLine, 0, len(requests))
	for _, request := range requests {
		lines = append(lines, JournalLine{
			JournalEntryID: entryID,
			AccountID:      request.AccountID,
			Debit:          request.Debit,
			Kredit:         request.Credit,
		})
	}
	return lines
}

func applyBudgetRequest(budget *Budget, request BudgetRequest) {
	budget.ProjectID = request.ProjectID
	budget.Year = request.Year
	budget.Value = request.Value
	budget.Note = strings.TrimSpace(request.Note)
}

func applyCashTransactionRequest(transaction *CashTransaction, request CashTransactionRequest) {
	transaction.Date = request.Date
	transaction.Type = request.Type
	transaction.AccountID = request.AccountID
	transaction.ProjectID = request.ProjectID
	transaction.Amount = request.Amount
	transaction.Note = strings.TrimSpace(request.Note)
}
