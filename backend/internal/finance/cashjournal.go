package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	cashAccountCode         = "1110"
	cashJournalSource       = "kas"
	cashJournalNumberPrefix = "KAS-"
	cashJournalNumberFormat = cashJournalNumberPrefix + "%04d-%06d"
)

var cashJournalDefaultNotes = map[string]string{
	cashIn:  "Kas masuk",
	cashOut: "Kas keluar",
}

func PostMissingCashJournals(ctx context.Context, repository Repository) (int, error) {
	var posted int
	err := repository.Transaction(ctx, func(repository Repository) error {
		transactions, err := repository.ListCashTransactionsWithoutJournal(ctx)
		if err != nil {
			return apperror.Internal(err)
		}
		posted = len(transactions)
		return postCashJournals(ctx, repository, transactions)
	})
	if err != nil {
		return 0, fmt.Errorf("post missing cash journals: %w", err)
	}
	return posted, nil
}

func postCashJournals(ctx context.Context, repository Repository, transactions []CashTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	cashAccountID, err := findCashAccount(ctx, repository)
	if err != nil {
		return err
	}
	for _, transaction := range transactions {
		if _, err := postCashJournal(ctx, repository, cashAccountID, transaction); err != nil {
			return err
		}
	}
	return nil
}

func recordCashTransaction(ctx context.Context, repository Repository, request CashTransactionRequest) (CashTransactionRow, error) {
	var transaction CashTransaction
	applyCashTransactionRequest(&transaction, request)

	cashAccountID, err := resolveCashAccount(ctx, repository, transaction)
	if err != nil {
		return CashTransactionRow{}, err
	}
	if err := repository.CreateCashTransaction(ctx, &transaction); err != nil {
		return CashTransactionRow{}, cashWriteErrors.Resolve(err)
	}

	entry, err := postCashJournal(ctx, repository, cashAccountID, transaction)
	if err != nil {
		return CashTransactionRow{}, err
	}
	return newCashTransactionRow(transaction, entry.ID), nil
}

func rewriteCashTransaction(ctx context.Context, repository Repository, id uuid.UUID, request CashTransactionRequest) (CashTransactionRow, error) {
	transaction, err := repository.FindCashTransaction(ctx, id)
	if err != nil {
		return CashTransactionRow{}, cashReadErrors.Resolve(err)
	}
	applyCashTransactionRequest(&transaction, request)

	cashAccountID, err := resolveCashAccount(ctx, repository, transaction)
	if err != nil {
		return CashTransactionRow{}, err
	}
	if err := repository.SaveCashTransaction(ctx, &transaction); err != nil {
		return CashTransactionRow{}, cashWriteErrors.Resolve(err)
	}

	entry, err := syncCashJournal(ctx, repository, cashAccountID, transaction)
	if err != nil {
		return CashTransactionRow{}, err
	}
	return newCashTransactionRow(transaction, entry.ID), nil
}

func removeCashTransaction(ctx context.Context, repository Repository, id uuid.UUID) error {
	if err := repository.DeleteCashJournal(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return cashReadErrors.Resolve(repository.DeleteCashTransaction(ctx, id))
}

func findCashAccount(ctx context.Context, repository Repository) (uuid.UUID, error) {
	id, err := repository.FindAccountIDByCode(ctx, cashAccountCode)
	if err != nil {
		return uuid.Nil, cashAccountErrors.Resolve(err)
	}
	return id, nil
}

func resolveCashAccount(ctx context.Context, repository Repository, transaction CashTransaction) (uuid.UUID, error) {
	cashAccountID, err := findCashAccount(ctx, repository)
	if err != nil {
		return uuid.Nil, err
	}
	if transaction.AccountID == cashAccountID {
		return uuid.Nil, errCashCounterIsCash
	}
	return cashAccountID, nil
}

func postCashJournal(ctx context.Context, repository Repository, cashAccountID uuid.UUID, transaction CashTransaction) (JournalEntry, error) {
	sequence, err := repository.NextCashJournalSequence(ctx)
	if err != nil {
		return JournalEntry{}, apperror.Internal(err)
	}

	entry := JournalEntry{
		Number:            cashJournalNumber(transaction.Date, sequence),
		Source:            cashJournalSource,
		CashTransactionID: &transaction.ID,
	}
	applyCashJournal(&entry, transaction)
	if err := repository.CreateJournalEntry(ctx, &entry); err != nil {
		return JournalEntry{}, cashJournalErrors.Resolve(err)
	}
	return entry, writeCashJournalLines(ctx, repository, entry.ID, cashAccountID, transaction)
}

func syncCashJournal(ctx context.Context, repository Repository, cashAccountID uuid.UUID, transaction CashTransaction) (JournalEntry, error) {
	entry, err := repository.FindCashJournal(ctx, transaction.ID)
	if errors.Is(err, database.ErrNotFound) {
		return postCashJournal(ctx, repository, cashAccountID, transaction)
	}
	if err != nil {
		return JournalEntry{}, apperror.Internal(err)
	}

	applyCashJournal(&entry, transaction)
	if err := repository.SaveJournalEntry(ctx, &entry); err != nil {
		return JournalEntry{}, cashJournalErrors.Resolve(err)
	}
	if err := repository.DeleteJournalLines(ctx, entry.ID); err != nil {
		return JournalEntry{}, apperror.Internal(err)
	}
	return entry, writeCashJournalLines(ctx, repository, entry.ID, cashAccountID, transaction)
}

func writeCashJournalLines(ctx context.Context, repository Repository, entryID, cashAccountID uuid.UUID, transaction CashTransaction) error {
	lines := cashJournalLines(entryID, cashAccountID, transaction)
	return cashJournalErrors.Resolve(repository.CreateJournalLines(ctx, lines))
}

func cashJournalLines(entryID, cashAccountID uuid.UUID, transaction CashTransaction) []JournalLine {
	debited, credited := transaction.AccountID, cashAccountID
	if transaction.Type == cashIn {
		debited, credited = cashAccountID, transaction.AccountID
	}
	return []JournalLine{
		{JournalEntryID: entryID, AccountID: debited, Debit: transaction.Amount},
		{JournalEntryID: entryID, AccountID: credited, Kredit: transaction.Amount},
	}
}

func applyCashJournal(entry *JournalEntry, transaction CashTransaction) {
	entry.Date = transaction.Date
	entry.Note = cashJournalNote(transaction)
}

func cashJournalNote(transaction CashTransaction) string {
	if transaction.Note != "" {
		return transaction.Note
	}
	return cashJournalDefaultNotes[transaction.Type]
}

func cashJournalNumber(date time.Time, sequence int64) string {
	return fmt.Sprintf(cashJournalNumberFormat, date.Year(), sequence)
}

func newCashTransactionRow(transaction CashTransaction, journalEntryID uuid.UUID) CashTransactionRow {
	return CashTransactionRow{
		ID:             transaction.ID,
		Date:           transaction.Date,
		Type:           transaction.Type,
		AccountID:      transaction.AccountID,
		ProjectID:      transaction.ProjectID,
		Amount:         transaction.Amount,
		Note:           transaction.Note,
		JournalEntryID: &journalEntryID,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}
}
