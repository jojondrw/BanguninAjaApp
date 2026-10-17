//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
)

func TestOwnerTermBecomesAReceivableWhosePaymentIsCashIn(t *testing.T) {
	ctx := context.Background()
	ensureAccount(t, "1110", "Kas", "asset")
	receivableAccount := ensureAccount(t, "1200", "Piutang Usaha", "asset")
	projects := project.NewService(project.NewRepository(db, project.NewBillingLedger))
	created := newProject(t, projects)
	owner := sales.Customer{Name: "Pemberi kerja uji " + uniqueCode("")}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	retention, days := 5.0, 30

	if _, err := projects.SaveOwnerContract(ctx, created.ID, project.OwnerContractRequest{
		CustomerID: owner.ID, Number: "KTR-UJI-01", Value: 4_000_000_000, RetentionPercent: &retention, PaymentTermDays: &days,
	}); err != nil {
		t.Fatalf("save owner contract: %v", err)
	}
	required := 20
	schedule, err := projects.CreateOwnerTerm(ctx, created.ID, project.PaymentTermRequest{RequiredProgress: &required, Percent: 25})
	if err != nil {
		t.Fatalf("create owner term: %v", err)
	}
	if err := db.Exec("UPDATE project SET progress = 30 WHERE id = ?", created.ID).Error; err != nil {
		t.Fatalf("set progress: %v", err)
	}
	schedule, err = projects.ClaimOwnerTerm(ctx, created.ID, schedule.Terms[0].ID, newUser(t))
	if err != nil {
		t.Fatalf("claim owner term: %v", err)
	}
	receivableID := schedule.Terms[0].ReceivableID
	if receivableID == nil {
		t.Fatal("claim did not link a receivable")
	}

	stored := count(t, "SELECT COUNT(*) FROM receivable r JOIN owner_term t ON t.receivable_id = r.id "+
		"WHERE r.id = ? AND r.amount = 950000000 AND r.customer_id = ? AND r.project_id = ? "+
		"AND r.reference = ? AND r.due_date = t.claimed_on + 30",
		*receivableID, owner.ID, created.ID, "Tagihan termin 1 "+created.Code)
	if stored != 1 {
		t.Fatal("receivable is not 25% of the owner contract minus 5% retention, due 30 days after the claim")
	}

	bills := billing.NewService(billing.NewRepository(db, billing.NewFinanceCashBook))
	if _, err := bills.PayReceivable(ctx, *receivableID, billing.PaymentRequest{Amount: 950_000_000}, nil); err != nil {
		t.Fatalf("pay receivable: %v", err)
	}
	cash := count(t, "SELECT COUNT(*) FROM cash_transaction c JOIN billing_payment p ON p.id = c.billing_payment_id "+
		"WHERE p.receivable_id = ? AND c.type = 'in' AND c.amount = 950000000 AND c.account_id = ? AND c.project_id = ?",
		*receivableID, receivableAccount, created.ID)
	if cash != 1 {
		t.Fatal("paying the owner's receivable must record cash-in against Piutang Usaha for the project")
	}

	schedule, err = projects.OwnerSchedule(ctx, created.ID)
	if err != nil {
		t.Fatalf("owner schedule: %v", err)
	}
	if schedule.Terms[0].State != "paid" || schedule.PaidAmount != 950_000_000 {
		t.Fatalf("term after payment = %+v, paid %d", schedule.Terms[0], schedule.PaidAmount)
	}
}
