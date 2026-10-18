package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
)

const (
	cashIn              = "in"
	cashOut             = "out"
	ledgerMonths        = 18
	resetLookbackMonths = 36

	openingCapitalNote = "Setoran modal awal pemegang saham"
	bankLoanNote       = "Pencairan kredit konstruksi bank untuk Griya Asri"
	salaryNote         = "Pembayaran gaji karyawan — "
	officeNote         = "Biaya operasional kantor — "
)

type cashSeed struct {
	monthsAgo int
	day       int
	kind      string
	account   string
	project   string
	amount    int64
	note      string
}

type budgetSeed struct {
	project    string
	yearOffset int
	value      int64
	note       string
}

func companyCashNotes(today time.Time) []string {
	notes := []string{openingCapitalNote}
	for _, seed := range portfolioOneOffCash() {
		if seed.project == "" {
			notes = append(notes, seed.note)
		}
	}
	for monthsAgo := 0; monthsAgo <= resetLookbackMonths; monthsAgo++ {
		label := monthLabelOf(today, monthsAgo)
		notes = append(notes, salaryNote+label, officeNote+label)
	}
	return notes
}

func (s *seeder) cashSchedule() []cashSeed {
	schedule := []cashSeed{
		{ledgerMonths, 2, cashIn, "3100", "", 25_000_000_000, openingCapitalNote},
		{ledgerMonths - 1, 10, cashIn, "2300", projectGriyaAsri, 10_000_000_000, bankLoanNote},
		{6, 18, cashOut, "5200", projectRuko, 4_500_000_000, "Pelunasan pembelian lahan Ruko Sentra Makassar"},
		{3, 22, cashOut, "5300", projectRuko, 185_000_000, "Biaya pengurusan PKKPR Ruko Sentra Makassar"},
		{1, 14, cashOut, "5300", projectRuko, 142_500_000, "Biaya retribusi PBG Ruko Sentra Makassar"},
	}
	schedule = append(schedule, portfolioOneOffCash()...)
	for monthsAgo := ledgerMonths - 1; monthsAgo >= 0; monthsAgo-- {
		schedule = append(schedule, s.monthlyCash(monthsAgo)...)
		schedule = append(schedule, s.portfolioMonthlyCash(monthsAgo)...)
	}
	return schedule
}

func (s *seeder) monthlyCash(monthsAgo int) []cashSeed {
	label := s.monthLabel(monthsAgo)
	step := int64(ledgerMonths - monthsAgo)
	entries := []cashSeed{
		{monthsAgo, 5, cashOut, "5100", projectGriyaAsri, 1_350_000_000 + step*42_000_000, "Termin kontraktor Griya Asri — " + label},
		{monthsAgo, 12, cashIn, "4100", projectGriyaAsri, 220_000_000 + step*31_000_000, "Penerimaan DP & cicilan unit Griya Asri — " + label},
		{monthsAgo, 8, cashOut, "5100", projectHospital, 980_000_000 + step*36_000_000, "Termin kontraktor RSIA Cirebon — " + label},
		{monthsAgo, 25, cashOut, "5500", "", 112_000_000 + step*1_500_000, salaryNote + label},
		{monthsAgo, 20, cashOut, "5400", projectGriyaAsri, 28_000_000 + (step%3)*6_500_000, "Biaya pemasaran & pameran Griya Asri — " + label},
		{monthsAgo, 3, cashOut, "5600", "", 18_500_000 + (step%2)*2_250_000, officeNote + label},
	}
	if monthsAgo >= ledgerMonths-4 {
		entries = append(entries, cashSeed{monthsAgo, 17, cashIn, "4100", projectCluster, 420_000_000 + step*15_000_000, "Pelunasan unit Cluster Lama Depok — " + label})
	}
	return entries
}

func (s *seeder) demoBudgets() []budgetSeed {
	return append([]budgetSeed{
		{projectGriyaAsri, -1, 14_500_000_000, "Anggaran tahap pematangan lahan & infrastruktur"},
		{projectGriyaAsri, 0, 21_800_000_000, "Anggaran pembangunan unit Blok A & B"},
		{projectHospital, -1, 9_500_000_000, "Anggaran pekerjaan persiapan & pondasi"},
		{projectHospital, 0, 31_000_000_000, "Anggaran struktur & MEP"},
		{projectRuko, 0, 6_200_000_000, "Anggaran pembelian lahan & perizinan"},
		{projectRuko, 1, 15_400_000_000, "Anggaran konstruksi ruko"},
		{projectCluster, -1, 7_800_000_000, "Anggaran penyelesaian unit & serah terima"},
	}, portfolioBudgets()...)
}

func (s *seeder) seedFinance() error {
	for _, seed := range s.cashSchedule() {
		if err := s.seedCash(seed); err != nil {
			return err
		}
	}
	for _, seed := range s.demoBudgets() {
		if err := s.seedBudget(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedCash(seed cashSeed) error {
	accountID, err := s.reference("account", seed.account)
	if err != nil {
		return err
	}
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return err
	}
	_, _, err = s.ensure("cash_transaction", map[string]any{"note": seed.note}, func() (uuid.UUID, error) {
		response, err := s.services.finance.CreateCashTransaction(s.ctx, finance.CashTransactionRequest{
			Date: s.monthDay(seed.monthsAgo, seed.day), Type: seed.kind, AccountID: accountID,
			ProjectID: projectID, Amount: seed.amount, Note: seed.note,
		})
		return response.ID, err
	})
	return err
}

func (s *seeder) seedBudget(seed budgetSeed) error {
	projectID, err := s.known("project", seed.project)
	if err != nil {
		return err
	}
	year := s.today.Year() + seed.yearOffset
	_, _, err = s.ensure("budget", map[string]any{"project_id": projectID, "year": year}, func() (uuid.UUID, error) {
		response, err := s.services.finance.CreateBudget(s.ctx, finance.BudgetRequest{
			ProjectID: projectID, Year: year, Value: seed.value, Note: fmt.Sprintf("%s %d", seed.note, year),
		})
		return response.ID, err
	})
	return err
}
