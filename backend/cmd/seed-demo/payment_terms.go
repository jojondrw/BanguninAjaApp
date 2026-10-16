package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

const claimPaperworkDays = 3

type termSeed struct {
	required int
	percent  float64
	claim    bool
}

type constructionSeed struct {
	project   string
	vendor    string
	number    string
	value     int64
	retention float64
	terms     []termSeed
}

func demoConstructionContracts() []constructionSeed {
	return []constructionSeed{
		{projectHospital, vendorStructure, "DEMO-SPK-RSIA-01", 23_500_000_000, 5, []termSeed{
			{0, 10, true}, {20, 20, true}, {40, 20, true}, {70, 25, false}, {100, 25, false},
		}},
		{projectGriyaAsri, vendorHousing, "DEMO-SPK-GAB-01", 22_750_000_000, 5, []termSeed{
			{25, 25, true}, {50, 25, true}, {75, 25, false}, {100, 25, false},
		}},
		{projectCluster, vendorHousing, "DEMO-SPK-CLD-01", 6_850_000_000, 5, []termSeed{
			{30, 30, true}, {60, 30, true}, {100, 40, true},
		}},
	}
}

type termWriter interface {
	createSchedule(ctx context.Context, contract *project.ConstructionContract, terms []project.PaymentTerm) error
	reachedOn(ctx context.Context, projectID uuid.UUID, progress int) (time.Time, int, bool, error)
	backdateClaim(ctx context.Context, termID uuid.UUID, claimedOn time.Time, progress int, due time.Time, status string) error
}

func (s *seeder) seedPaymentTerms() error {
	for _, seed := range demoConstructionContracts() {
		if err := s.seedConstructionContract(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedConstructionContract(seed constructionSeed) error {
	projectID, known := s.ids["project:"+seed.project]
	if !known {
		return nil
	}
	vendorID, err := s.known("vendor", seed.vendor)
	if err != nil {
		return err
	}

	contract := project.ConstructionContract{
		ProjectID: projectID, VendorID: vendorID, Number: seed.number, Value: seed.value, RetentionPercent: seed.retention,
	}
	terms := make([]project.PaymentTerm, 0, len(seed.terms))
	for _, term := range seed.terms {
		terms = append(terms, project.PaymentTerm{
			ProjectID: projectID, Kind: "progress", RequiredProgress: term.required, Percent: term.percent,
		})
	}
	_, created, err := s.ensure("construction_contract", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		err := s.terms.createSchedule(s.ctx, &contract, terms)
		return contract.ID, err
	})
	if err != nil || !created {
		return err
	}

	for i, term := range seed.terms {
		if !term.claim {
			continue
		}
		if err := s.claimTerm(projectID, seed, terms[i].ID, i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) claimTerm(projectID uuid.UUID, seed constructionSeed, termID uuid.UUID, position int) error {
	schedule, err := s.services.projects.ClaimPaymentTerm(s.ctx, projectID, termID, s.owner)
	if err != nil {
		return fmt.Errorf("claim term %d of %s: %w", position, seed.number, err)
	}
	s.tally.add("payment_term claim")

	claimedOn, progress, err := s.claimMoment(projectID, seed.terms[position-1].required)
	if err != nil {
		return err
	}
	due := claimedOn.AddDate(0, 0, schedule.Contract.PaymentTermDays)
	if err := s.terms.backdateClaim(s.ctx, termID, claimedOn, progress, due, duedate.Status(due, false, s.today)); err != nil {
		return err
	}
	if !due.Before(s.today) {
		return nil
	}

	for _, item := range schedule.Terms {
		if item.ID != termID || item.PayableID == nil {
			continue
		}
		paidAt := due
		_, err := s.services.billing.PayPayable(s.ctx, *item.PayableID, billing.PaymentRequest{
			Amount:    item.NetAmount,
			PaidAt:    &paidAt,
			Method:    "transfer",
			Reference: fmt.Sprintf("TRF-%s-%d", strings.TrimPrefix(seed.number, "DEMO-SPK-"), position),
			Note:      fmt.Sprintf("Pembayaran termin %d setelah dipotong retensi", position),
		}, &s.owner)
		if err != nil {
			return fmt.Errorf("pay term %d of %s: %w", position, seed.number, err)
		}
		s.tally.add("billing_payment")
	}
	return nil
}

func (s *seeder) claimMoment(projectID uuid.UUID, required int) (time.Time, int, error) {
	reached, progress, found, err := s.terms.reachedOn(s.ctx, projectID, required)
	if err != nil {
		return time.Time{}, 0, err
	}
	if !found {
		return s.today, required, nil
	}
	return pastOrToday(reached.AddDate(0, 0, claimPaperworkDays), s.today), progress, nil
}

type gormTermWriter struct {
	db *gorm.DB
}

func (w gormTermWriter) createSchedule(ctx context.Context, contract *project.ConstructionContract, terms []project.PaymentTerm) error {
	if err := w.db.WithContext(ctx).Create(contract).Error; err != nil {
		return fmt.Errorf("create construction contract: %w", err)
	}
	if err := w.db.WithContext(ctx).Create(&terms).Error; err != nil {
		return fmt.Errorf("create payment terms: %w", err)
	}
	return nil
}

func (w gormTermWriter) reachedOn(ctx context.Context, projectID uuid.UUID, progress int) (time.Time, int, bool, error) {
	var snapshots []project.ProgressSnapshot
	err := w.db.WithContext(ctx).
		Where("project_id = ? AND progress >= ?", projectID, progress).
		Order("recorded_on").
		Limit(1).
		Find(&snapshots).Error
	if err != nil {
		return time.Time{}, 0, false, fmt.Errorf("find progress milestone: %w", err)
	}
	if len(snapshots) == 0 {
		return time.Time{}, 0, false, nil
	}
	return snapshots[0].RecordedOn, snapshots[0].Progress, true, nil
}

func (w gormTermWriter) backdateClaim(ctx context.Context, termID uuid.UUID, claimedOn time.Time, progress int, due time.Time, status string) error {
	db := w.db.WithContext(ctx)
	err := db.Model(&project.PaymentTerm{}).
		Where("id = ?", termID).
		Updates(map[string]any{"claimed_on": claimedOn, "claimed_progress": progress}).Error
	if err != nil {
		return fmt.Errorf("backdate term claim: %w", err)
	}
	err = db.Table("payable").
		Where("id = (SELECT payable_id FROM payment_term WHERE id = ?)", termID).
		Updates(map[string]any{"due_date": due, "status": status}).Error
	if err != nil {
		return fmt.Errorf("backdate term payable: %w", err)
	}
	return nil
}
