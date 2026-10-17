//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/master"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
)

func ensureAccount(t *testing.T, code, name, kind string) uuid.UUID {
	t.Helper()
	account := master.Account{Code: code, Name: name, Type: kind}
	if err := db.Where("code = ?", code).FirstOrCreate(&account).Error; err != nil {
		t.Fatalf("ensure account %s: %v", code, err)
	}
	return account.ID
}

func TestPayingAPayableWritesCashAndAJournalThatFinanceCannotEdit(t *testing.T) {
	ctx := context.Background()
	cashAccount := ensureAccount(t, "1110", "Kas", "asset")
	payableAccount := ensureAccount(t, "2100", "Utang Usaha", "liability")
	projects := project.NewService(project.NewRepository(db, project.NewBillingPayables))
	created := newProject(t, projects)
	bills := billing.NewService(billing.NewRepository(db, billing.NewFinanceCashBook))

	payable, err := bills.CreatePayable(ctx, billing.PayableRequest{
		VendorID: newContractor(t), ProjectID: &created.ID, Reference: uniqueCode("UJI-HUTANG-"),
		DueDate: time.Now().AddDate(0, 0, 30), Amount: 900_000,
	})
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}
	if _, err := bills.PayPayable(ctx, payable.ID, billing.PaymentRequest{Amount: 400_000}, nil); err != nil {
		t.Fatalf("pay payable: %v", err)
	}

	var cashKey string
	if err := db.Raw("SELECT c.id FROM cash_transaction c JOIN billing_payment p ON p.id = c.billing_payment_id "+
		"WHERE p.payable_id = ? AND c.type = 'out' AND c.amount = 400000 AND c.account_id = ? AND c.project_id = ?",
		payable.ID, payableAccount, created.ID).Scan(&cashKey).Error; err != nil || cashKey == "" {
		t.Fatalf("the payment did not become a cash-out against Utang Usaha for its project (err %v)", err)
	}
	cashID := uuid.MustParse(cashKey)
	lines := count(t, "SELECT COUNT(*) FROM journal_line l JOIN journal_entry e ON e.id = l.journal_entry_id "+
		"WHERE e.cash_transaction_id = ? AND ((l.account_id = ? AND l.debit = 400000) OR (l.account_id = ? AND l.kredit = 400000))",
		cashID, payableAccount, cashAccount)
	if lines != 2 {
		t.Fatalf("want the journal to debit Utang Usaha and credit Kas, got %d matching lines", lines)
	}

	cash := finance.NewService(finance.NewRepository(db))
	if err := cash.DeleteCashTransaction(ctx, cashID); appCode(err) != "cash_from_billing" {
		t.Fatalf("deleting billing cash from Keuangan: err = %v", err)
	}
	if count(t, "SELECT COUNT(*) FROM cash_transaction WHERE id = ?", cashID) != 1 {
		t.Fatal("the refused delete must leave the cash transaction in place")
	}
}
