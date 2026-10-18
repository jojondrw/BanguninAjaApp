package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

const claimPaperworkDays = 3

type termSeed struct {
	required int
	percent  float64
	claim    bool
}

type changeSeed struct {
	description string
	value       int64
	days        int
	decision    string
	proposed    int
	decided     int
}

type constructionSeed struct {
	project   string
	vendor    string
	number    string
	value     int64
	retention float64
	terms     []termSeed
	changes   []changeSeed
}

func demoConstructionContracts() []constructionSeed {
	return append(append(coreConstructionContracts(), portfolioConstructionContracts()...), generatedConstructionContracts()...)
}

func coreConstructionContracts() []constructionSeed {
	return []constructionSeed{
		{projectHospital, vendorStructure, "DEMO-SPK-RSIA-01", 23_500_000_000, 5, []termSeed{
			{0, 10, true}, {20, 20, true}, {40, 20, true}, {70, 25, false}, {100, 25, false},
		}, []changeSeed{
			{"Penambahan ruang NICU dan instalasi gas medis di lantai 3", 1_250_000_000, 30, "approved", -16, -9},
			{"Pengurangan pekerjaan lansekap halaman depan karena area parkir diperluas", -350_000_000, 0, "", -3, 0},
		}},
		{projectGriyaAsri, vendorHousing, "DEMO-SPK-GAB-01", 22_750_000_000, 5, []termSeed{
			{25, 25, true}, {50, 25, true}, {75, 25, false}, {100, 25, false},
		}, []changeSeed{
			{"Penggantian atap metal ke genteng keramik untuk blok B", 420_000_000, 0, "rejected", -30, -26},
		}},
		{projectCluster, vendorHousing, "DEMO-SPK-CLD-01", 6_850_000_000, 5, []termSeed{
			{30, 30, true}, {60, 30, true}, {100, 40, true},
		}, nil},
	}
}

type termWriter interface {
	createSchedule(ctx context.Context, contract *project.ConstructionContract, terms []project.PaymentTerm) error
	reachedOn(ctx context.Context, projectID uuid.UUID, progress int) (time.Time, int, bool, error)
	backdateClaim(ctx context.Context, termID uuid.UUID, claimedOn time.Time, progress int, due time.Time, status string) error
	backdateOwnerClaim(ctx context.Context, termID uuid.UUID, claimedOn time.Time, progress int, due time.Time, status string) error
	backdateChange(ctx context.Context, changeID uuid.UUID, proposedOn time.Time, decidedOn *time.Time) error
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
	if err != nil {
		return err
	}

	for i, term := range seed.terms {
		if !created || !term.claim {
			continue
		}
		if err := s.claimTerm(projectID, seed, terms[i].ID, i+1); err != nil {
			return err
		}
	}
	for _, change := range seed.changes {
		if err := s.seedChange(projectID, change); err != nil {
			return fmt.Errorf("change order of %s: %w", seed.number, err)
		}
	}
	return nil
}

func (s *seeder) seedChange(projectID uuid.UUID, seed changeSeed) error {
	keys := map[string]any{"project_id": projectID, "description": seed.description}
	_, _, err := s.ensure("contract_change", keys, func() (uuid.UUID, error) {
		return s.proposeChange(projectID, seed)
	})
	return err
}

func (s *seeder) proposeChange(projectID uuid.UUID, seed changeSeed) (uuid.UUID, error) {
	schedule, err := s.services.projects.ProposeContractChange(s.ctx, projectID, s.owner, project.ContractChangeRequest{
		Description: seed.description, ValueChange: seed.value, DaysChange: seed.days,
	})
	if err != nil {
		return uuid.Nil, err
	}
	changeID := schedule.Changes[len(schedule.Changes)-1].ID

	var decidedOn *time.Time
	if seed.decision != "" {
		_, err := s.services.projects.DecideContractChange(s.ctx, projectID, changeID, s.owner, access.RoleAdmin,
			project.ContractChangeDecision{Status: seed.decision})
		if err != nil {
			return uuid.Nil, err
		}
		decidedOn = s.dayPointer(seed.decided)
	}
	return changeID, s.terms.backdateChange(s.ctx, changeID, s.day(seed.proposed), decidedOn)
}

func (s *seeder) claimTerm(projectID uuid.UUID, seed constructionSeed, termID uuid.UUID, position int) error {
	schedule, err := s.services.projects.ClaimPaymentTerm(s.ctx, projectID, termID, s.owner)
	if progressNotReached(err) {
		return nil
	}
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

func progressNotReached(err error) bool {
	var appError *apperror.Error
	return errors.As(err, &appError) && appError.Code == "payment_term_progress_not_reached"
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

func (w gormTermWriter) backdateChange(ctx context.Context, changeID uuid.UUID, proposedOn time.Time, decidedOn *time.Time) error {
	updates := map[string]any{"created_at": proposedOn, "updated_at": proposedOn}
	if decidedOn != nil {
		updates["decided_on"] = *decidedOn
		updates["updated_at"] = *decidedOn
	}
	err := w.db.WithContext(ctx).Model(&project.ContractChange{}).Where("id = ?", changeID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("backdate change order: %w", err)
	}
	return nil
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
