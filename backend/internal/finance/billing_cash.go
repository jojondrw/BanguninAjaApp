package finance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	receivableAccountCode = "1200"
	payableAccountCode    = "2100"
)

var (
	errCashFromBilling = apperror.Unprocessable("cash_from_billing",
		"Kas ini tercatat otomatis dari pembayaran tagihan, jadi diubah atau dihapus lewat modul Tagihan")
	errBillingAccountMissing = apperror.Unprocessable("billing_cash_account_not_found",
		"Akun 1200 Piutang Usaha atau 2100 Utang Usaha tidak ada di bagan akun, jadi pembayaran tagihan tidak bisa dicatat ke kas. Tambahkan akun itu lewat Master akun")

	billingAccountErrors = database.ErrorMap{NotFound: errBillingAccountMissing}
)

type BillingCash struct {
	PaymentID uuid.UUID
	Incoming  bool
	ProjectID *uuid.UUID
	Date      time.Time
	Amount    int64
	Note      string
}

type BillingCashLedger struct {
	repository Repository
}

func NewBillingCashLedger(repository Repository) BillingCashLedger {
	return BillingCashLedger{repository: repository}
}

func (l BillingCashLedger) Record(ctx context.Context, entry BillingCash) error {
	kind, counterCode := cashOut, payableAccountCode
	if entry.Incoming {
		kind, counterCode = cashIn, receivableAccountCode
	}
	counterID, err := l.repository.FindAccountIDByCode(ctx, counterCode)
	if err != nil {
		return billingAccountErrors.Resolve(err)
	}
	cashAccountID, err := findCashAccount(ctx, l.repository)
	if err != nil {
		return err
	}

	paymentID := entry.PaymentID
	transaction := CashTransaction{
		Date:             entry.Date,
		Type:             kind,
		AccountID:        counterID,
		ProjectID:        entry.ProjectID,
		Amount:           entry.Amount,
		Note:             entry.Note,
		BillingPaymentID: &paymentID,
	}
	if err := l.repository.CreateCashTransaction(ctx, &transaction); err != nil {
		return cashWriteErrors.Resolve(err)
	}
	_, err = postCashJournal(ctx, l.repository, cashAccountID, transaction)
	return err
}

func ensureManualCash(transaction CashTransaction) error {
	if transaction.BillingPaymentID != nil {
		return errCashFromBilling
	}
	return nil
}
