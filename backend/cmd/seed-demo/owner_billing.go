package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

const hospitalOwner = "DEMO-YKIA-0001"

type ownerSeed struct {
	project   string
	customer  sales.CustomerRequest
	number    string
	value     int64
	retention float64
	days      int
	terms     []termSeed
}

func demoOwnerContracts() []ownerSeed {
	return []ownerSeed{
		{projectHospital, sales.CustomerRequest{
			Name: "Yayasan Kesehatan Ibu dan Anak Cirebon", Contact: "0231-820-4410",
			Email: "keuangan@ykia-cirebon.contoh.id", IdentityNumber: hospitalOwner, Address: "Jl. Kartini No. 12, Cirebon",
		}, "DEMO-KTR-RSIA-01", 67_000_000_000, 5, 30, []termSeed{
			{0, 10, true}, {25, 20, true}, {50, 20, false}, {75, 25, false}, {100, 25, false},
		}},
	}
}

func (s *seeder) seedOwnerBilling() error {
	for _, seed := range demoOwnerContracts() {
		if err := s.seedOwnerContract(seed); err != nil {
			return fmt.Errorf("owner contract %s: %w", seed.number, err)
		}
	}
	return nil
}

func (s *seeder) seedOwnerContract(seed ownerSeed) error {
	projectID, known := s.ids["project:"+seed.project]
	if !known {
		return nil
	}
	customerID, _, err := s.ensure("customer", map[string]any{"identity_number": seed.customer.IdentityNumber}, func() (uuid.UUID, error) {
		response, err := s.services.sales.CreateCustomer(s.ctx, seed.customer)
		return response.ID, err
	})
	if err != nil {
		return err
	}

	var terms []project.OwnerTermResponse
	_, created, err := s.ensure("owner_contract", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		retention, days := seed.retention, seed.days
		if _, err := s.services.projects.SaveOwnerContract(s.ctx, projectID, project.OwnerContractRequest{
			CustomerID: customerID, Number: seed.number, Value: seed.value, RetentionPercent: &retention, PaymentTermDays: &days,
		}); err != nil {
			return uuid.Nil, err
		}
		for _, term := range seed.terms {
			required := term.required
			schedule, err := s.services.projects.CreateOwnerTerm(s.ctx, projectID, project.PaymentTermRequest{RequiredProgress: &required, Percent: term.percent})
			if err != nil {
				return uuid.Nil, err
			}
			terms = schedule.Terms
		}
		return uuid.Nil, nil
	})
	if err != nil || !created {
		return err
	}

	for i, term := range seed.terms {
		if !term.claim {
			continue
		}
		if err := s.claimOwnerTerm(projectID, seed, terms[i].ID, i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) claimOwnerTerm(projectID uuid.UUID, seed ownerSeed, termID uuid.UUID, position int) error {
	schedule, err := s.services.projects.ClaimOwnerTerm(s.ctx, projectID, termID, s.owner)
	if err != nil {
		return fmt.Errorf("bill term %d: %w", position, err)
	}
	s.tally.add("owner_term claim")

	claimedOn, progress, err := s.claimMoment(projectID, seed.terms[position-1].required)
	if err != nil {
		return err
	}
	due := claimedOn.AddDate(0, 0, seed.days)
	if err := s.terms.backdateOwnerClaim(s.ctx, termID, claimedOn, progress, due, duedate.Status(due, false, s.today)); err != nil {
		return err
	}
	if !due.Before(s.today) {
		return nil
	}

	for _, item := range schedule.Terms {
		if item.ID != termID || item.ReceivableID == nil {
			continue
		}
		paidAt := due
		_, err := s.services.billing.PayReceivable(s.ctx, *item.ReceivableID, billing.PaymentRequest{
			Amount:    item.NetAmount,
			PaidAt:    &paidAt,
			Method:    "transfer",
			Reference: fmt.Sprintf("TRF-YKIA-%d", position),
			Note:      fmt.Sprintf("Pembayaran tagihan termin %d dari pemberi kerja setelah dipotong retensi", position),
		}, &s.owner)
		if err != nil {
			return fmt.Errorf("receive term %d: %w", position, err)
		}
		s.tally.add("billing_payment")
	}
	return nil
}

func (w gormTermWriter) backdateOwnerClaim(ctx context.Context, termID uuid.UUID, claimedOn time.Time, progress int, due time.Time, status string) error {
	db := w.db.WithContext(ctx)
	err := db.Model(&project.OwnerTerm{}).
		Where("id = ?", termID).
		Updates(map[string]any{"claimed_on": claimedOn, "claimed_progress": progress}).Error
	if err != nil {
		return fmt.Errorf("backdate owner term claim: %w", err)
	}
	err = db.Table("receivable").
		Where("id = (SELECT receivable_id FROM owner_term WHERE id = ?)", termID).
		Updates(map[string]any{"due_date": due, "status": status}).Error
	if err != nil {
		return fmt.Errorf("backdate owner term receivable: %w", err)
	}
	return nil
}
