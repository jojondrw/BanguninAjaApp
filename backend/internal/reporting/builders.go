package reporting

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting/export"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/privacy"
)

const (
	fullPercent     = 100
	hoursPerDay     = 24
	agingBucketDay  = 30
	noProjectLabel  = "Kantor pusat"
	totalLabel      = "Total"
	cancelledStatus = "cancelled"
)

type sectionBuilder func(ctx context.Context, repository Repository, filter DatasetFilter, today time.Time) ([]export.Section, error)

var sectionBuilders = map[string]sectionBuilder{
	typeCashFlow: buildCashFlow,
	typeBudget:   buildBudget,
	typeProjects: buildProjects,
	typeAging:    buildAging,
	typeStock:    buildStock,
	typePayroll:  buildPayroll,
	typeSales:    buildSales,
}

var projectStatusLabels = map[string]string{
	"planning": "Perencanaan", "ongoing": "Berjalan", "on_hold": "Tertunda", "completed": "Selesai", "cancelled": "Batal",
}

var contractTypeLabels = map[string]string{"cash": "Tunai", "mortgage": "KPR", "installment": "Angsuran"}

var contractStatusLabels = map[string]string{"draft": "Draf", "active": "Aktif", "paid": "Lunas", "cancelled": "Batal"}

var dueStatusLabels = map[string]string{
	duedate.NotDue: "Belum jatuh tempo", duedate.Due: "Jatuh tempo 7 hari", duedate.Paid: "Lunas", duedate.Overdue: "Terlambat",
}

var agingBuckets = []string{"0–30 hari", "31–60 hari", "61–90 hari", "> 90 hari"}

func buildCashFlow(ctx context.Context, repository Repository, filter DatasetFilter, _ time.Time) ([]export.Section, error) {
	months, err := repository.SumCashByMonth(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("sum cash by month: %w", err)
	}
	opening, err := repository.CashBalanceBefore(ctx, filter.ProjectID, *filter.From)
	if err != nil {
		return nil, fmt.Errorf("cash opening balance: %w", err)
	}
	lines, err := repository.ListCashLines(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list cash lines: %w", err)
	}
	return []export.Section{cashMonthSection(months, opening, *filter.From, *filter.To), cashLineSection(lines)}, nil
}

func cashMonthSection(months []CashMonthRow, opening int64, from, to time.Time) export.Section {
	byMonth := map[time.Time]CashMonthRow{}
	for _, month := range months {
		byMonth[firstOfMonth(month.Period)] = month
	}
	section := export.Section{
		Title: "Arus kas per bulan",
		Columns: []export.Column{
			{Header: "Bulan", Kind: export.Text}, {Header: "Kas masuk (Rp)", Kind: export.Money},
			{Header: "Kas keluar (Rp)", Kind: export.Money}, {Header: "Arus bersih (Rp)", Kind: export.Money},
			{Header: "Saldo akhir (Rp)", Kind: export.Money},
		},
		Rows: [][]any{{"Saldo awal", nil, nil, nil, opening}},
	}
	balance, totalIn, totalOut := opening, int64(0), int64(0)
	for month := firstOfMonth(from); !month.After(to); month = month.AddDate(0, 1, 0) {
		row := byMonth[month]
		balance += row.CashIn - row.CashOut
		totalIn, totalOut = totalIn+row.CashIn, totalOut+row.CashOut
		section.Rows = append(section.Rows, []any{monthLabel(month), row.CashIn, row.CashOut, row.CashIn - row.CashOut, balance})
	}
	section.Total = []any{totalLabel, totalIn, totalOut, totalIn - totalOut, balance}
	return section
}

func cashLineSection(lines []CashLineRow) export.Section {
	section := export.Section{
		Title: "Transaksi kas",
		Columns: []export.Column{
			{Header: "Tanggal", Kind: export.Date}, {Header: "Akun", Kind: export.Text}, {Header: "Proyek", Kind: export.Text},
			{Header: "Keterangan", Kind: export.Text}, {Header: "Masuk (Rp)", Kind: export.Money}, {Header: "Keluar (Rp)", Kind: export.Money},
		},
		Rows: make([][]any, 0, len(lines)),
	}
	totalIn, totalOut := int64(0), int64(0)
	for _, line := range lines {
		var cashInValue, cashOutValue any
		if line.Type == cashIn {
			cashInValue, totalIn = line.Amount, totalIn+line.Amount
		} else {
			cashOutValue, totalOut = line.Amount, totalOut+line.Amount
		}
		section.Rows = append(section.Rows, []any{
			line.Date, line.AccountCode + " " + line.AccountName, projectText(line.ProjectCode, line.ProjectName), line.Note, cashInValue, cashOutValue,
		})
	}
	section.Total = []any{nil, totalLabel, nil, nil, totalIn, totalOut}
	return section
}

func buildBudget(ctx context.Context, repository Repository, filter DatasetFilter, _ time.Time) ([]export.Section, error) {
	budgets, err := repository.ListBudgetRealization(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list budget realization: %w", err)
	}
	section := export.Section{
		Title: "Anggaran dan realisasi",
		Columns: []export.Column{
			{Header: "Proyek", Kind: export.Text}, {Header: "Tahun", Kind: export.Text}, {Header: "Anggaran (Rp)", Kind: export.Money},
			{Header: "Realisasi (Rp)", Kind: export.Money}, {Header: "Sisa (Rp)", Kind: export.Money}, {Header: "Serapan", Kind: export.Percent},
		},
		Rows: make([][]any, 0, len(budgets)),
	}
	totalValue, totalRealized := int64(0), int64(0)
	for _, budget := range budgets {
		totalValue, totalRealized = totalValue+budget.Value, totalRealized+budget.Realized
		section.Rows = append(section.Rows, []any{
			projectText(budget.ProjectCode, budget.ProjectName), strconv.Itoa(budget.Year), budget.Value, budget.Realized,
			budget.Value - budget.Realized, absorption(budget.Realized, budget.Value),
		})
	}
	section.Total = []any{totalLabel, nil, totalValue, totalRealized, totalValue - totalRealized, absorption(totalRealized, totalValue)}
	return []export.Section{section}, nil
}

func buildProjects(ctx context.Context, repository Repository, filter DatasetFilter, _ time.Time) ([]export.Section, error) {
	projects, err := repository.ListProjectProgress(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list project progress: %w", err)
	}
	section := export.Section{
		Title: "Daftar proyek",
		Columns: []export.Column{
			{Header: "Kode", Kind: export.Text}, {Header: "Nama", Kind: export.Text}, {Header: "Jenis", Kind: export.Text},
			{Header: "Status", Kind: export.Text}, {Header: "Progres", Kind: export.Percent}, {Header: "Mulai", Kind: export.Date},
			{Header: "Target selesai", Kind: export.Date}, {Header: "Nilai kontrak (Rp)", Kind: export.Money}, {Header: "Total RAB (Rp)", Kind: export.Money},
		},
		Rows: make([][]any, 0, len(projects)),
	}
	totalContract, totalBudget := int64(0), int64(0)
	for _, project := range projects {
		totalContract, totalBudget = totalContract+project.ContractValue, totalBudget+project.BudgetItemTotal
		section.Rows = append(section.Rows, []any{
			project.Code, project.Name, project.Type, labelOf(projectStatusLabels, project.Status), int64(project.Progress),
			dateValue(project.StartDate), dateValue(project.TargetEndDate), project.ContractValue, project.BudgetItemTotal,
		})
	}
	section.Total = []any{totalLabel, nil, nil, nil, nil, nil, nil, totalContract, totalBudget}
	return []export.Section{section}, nil
}

func buildAging(ctx context.Context, repository Repository, filter DatasetFilter, today time.Time) ([]export.Section, error) {
	receivables, err := repository.ListOpenReceivables(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list open receivables: %w", err)
	}
	payables, err := repository.ListOpenPayables(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list open payables: %w", err)
	}
	return []export.Section{
		agingSummarySection(receivables, payables, today),
		openItemSection("Piutang terbuka", "Pelanggan", receivables, today),
		openItemSection("Utang terbuka", "Vendor", payables, today),
	}, nil
}

func agingSummarySection(receivables, payables []OpenItemRow, today time.Time) export.Section {
	receivableTotals, payableTotals := bucketTotals(receivables, today), bucketTotals(payables, today)
	section := export.Section{
		Title: "Ringkasan umur",
		Columns: []export.Column{
			{Header: "Kelompok umur", Kind: export.Text}, {Header: "Piutang (Rp)", Kind: export.Money}, {Header: "Utang (Rp)", Kind: export.Money},
		},
	}
	sumReceivable, sumPayable := int64(0), int64(0)
	for index, bucket := range agingBuckets {
		sumReceivable, sumPayable = sumReceivable+receivableTotals[index], sumPayable+payableTotals[index]
		section.Rows = append(section.Rows, []any{bucket, receivableTotals[index], payableTotals[index]})
	}
	section.Total = []any{totalLabel, sumReceivable, sumPayable}
	return section
}

func bucketTotals(items []OpenItemRow, today time.Time) []int64 {
	totals := make([]int64, len(agingBuckets))
	for _, item := range items {
		totals[agingBucket(daysLate(item.DueDate, today))] += item.Amount - item.PaidAmount
	}
	return totals
}

func openItemSection(title, partyHeader string, items []OpenItemRow, today time.Time) export.Section {
	section := export.Section{
		Title: title,
		Columns: []export.Column{
			{Header: "Referensi", Kind: export.Text}, {Header: partyHeader, Kind: export.Text}, {Header: "Proyek", Kind: export.Text},
			{Header: "Jatuh tempo", Kind: export.Date}, {Header: "Nilai (Rp)", Kind: export.Money}, {Header: "Dibayar (Rp)", Kind: export.Money},
			{Header: "Sisa (Rp)", Kind: export.Money}, {Header: "Hari terlambat", Kind: export.Integer}, {Header: "Umur", Kind: export.Text},
		},
		Rows: make([][]any, 0, len(items)),
	}
	totalAmount, totalPaid := int64(0), int64(0)
	for _, item := range items {
		late := daysLate(item.DueDate, today)
		totalAmount, totalPaid = totalAmount+item.Amount, totalPaid+item.PaidAmount
		section.Rows = append(section.Rows, []any{
			item.Reference, item.PartyName, projectText(item.ProjectCode, item.ProjectName), item.DueDate, item.Amount, item.PaidAmount,
			item.Amount - item.PaidAmount, int64(late), agingBuckets[agingBucket(late)],
		})
	}
	section.Total = []any{totalLabel, nil, nil, nil, totalAmount, totalPaid, totalAmount - totalPaid, nil, nil}
	return section
}

func daysLate(dueDate, today time.Time) int {
	due := duedate.Today(dueDate)
	if !due.Before(today) {
		return 0
	}
	return int(today.Sub(due).Hours() / hoursPerDay)
}

func agingBucket(days int) int {
	if days <= 0 {
		return 0
	}
	return min((days-1)/agingBucketDay, len(agingBuckets)-1)
}

func buildStock(ctx context.Context, repository Repository, filter DatasetFilter, _ time.Time) ([]export.Section, error) {
	levels, err := repository.ListStockLevels(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list stock levels: %w", err)
	}
	low, err := repository.ListLowStock(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list low stock: %w", err)
	}
	return []export.Section{stockLevelSection(levels), lowStockSection(low)}, nil
}

func stockLevelSection(levels []StockRow) export.Section {
	section := export.Section{
		Title: "Stok per gudang",
		Columns: []export.Column{
			{Header: "Gudang", Kind: export.Text}, {Header: "Kode material", Kind: export.Text}, {Header: "Material", Kind: export.Text},
			{Header: "Kategori", Kind: export.Text}, {Header: "Satuan", Kind: export.Text}, {Header: "Jumlah", Kind: export.Decimal},
		},
		Rows: make([][]any, 0, len(levels)),
	}
	for _, level := range levels {
		section.Rows = append(section.Rows, []any{
			level.WarehouseCode + " " + level.WarehouseName, level.MaterialCode, level.MaterialName, level.Category, level.UnitCode, level.Quantity,
		})
	}
	return section
}

func lowStockSection(low []LowStockRow) export.Section {
	section := export.Section{
		Title: "Stok menipis",
		Columns: []export.Column{
			{Header: "Kode material", Kind: export.Text}, {Header: "Material", Kind: export.Text}, {Header: "Satuan", Kind: export.Text},
			{Header: "Stok total", Kind: export.Decimal}, {Header: "Stok minimum", Kind: export.Decimal}, {Header: "Kekurangan", Kind: export.Decimal},
		},
		Rows: make([][]any, 0, len(low)),
	}
	for _, material := range low {
		section.Rows = append(section.Rows, []any{
			material.MaterialCode, material.MaterialName, material.UnitCode, material.TotalQuantity, material.MinimumStock,
			material.MinimumStock - material.TotalQuantity,
		})
	}
	return section
}

func buildPayroll(ctx context.Context, repository Repository, filter DatasetFilter, _ time.Time) ([]export.Section, error) {
	payrolls, err := repository.ListPayrolls(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list payrolls: %w", err)
	}
	return []export.Section{payrollSummarySection(payrolls), payrollDetailSection(payrolls)}, nil
}

type payrollTotals struct {
	count     int64
	basicPay  int64
	allowance int64
	deduction int64
	netPay    int64
	unpaid    int64
}

func (t payrollTotals) add(payroll PayrollRow) payrollTotals {
	t.count++
	t.basicPay += payroll.BasicPay
	t.allowance += payroll.Allowance
	t.deduction += payroll.Deduction
	t.netPay += payroll.NetPay
	if !payroll.Paid {
		t.unpaid += payroll.NetPay
	}
	return t
}

func (t payrollTotals) row(label string) []any {
	return []any{label, t.count, t.basicPay, t.allowance, t.deduction, t.netPay, t.unpaid}
}

func payrollSummarySection(payrolls []PayrollRow) export.Section {
	section := export.Section{
		Title: "Ringkasan per periode",
		Columns: []export.Column{
			{Header: "Periode", Kind: export.Text}, {Header: "Slip", Kind: export.Integer}, {Header: "Gaji pokok (Rp)", Kind: export.Money},
			{Header: "Tunjangan (Rp)", Kind: export.Money}, {Header: "Potongan (Rp)", Kind: export.Money},
			{Header: "Gaji bersih (Rp)", Kind: export.Money}, {Header: "Belum dibayar (Rp)", Kind: export.Money},
		},
	}
	var periods []string
	byPeriod := map[string]payrollTotals{}
	overall := payrollTotals{}
	for _, payroll := range payrolls {
		if _, seen := byPeriod[payroll.Period]; !seen {
			periods = append(periods, payroll.Period)
		}
		byPeriod[payroll.Period] = byPeriod[payroll.Period].add(payroll)
		overall = overall.add(payroll)
	}
	for _, period := range periods {
		section.Rows = append(section.Rows, byPeriod[period].row(periodLabel(period)))
	}
	section.Total = overall.row(totalLabel)
	return section
}

func payrollDetailSection(payrolls []PayrollRow) export.Section {
	section := export.Section{
		Title: "Rincian slip gaji",
		Columns: []export.Column{
			{Header: "Periode", Kind: export.Text}, {Header: "No. identitas", Kind: export.Text}, {Header: "Nama", Kind: export.Text},
			{Header: "Jabatan", Kind: export.Text}, {Header: "Proyek", Kind: export.Text}, {Header: "Gaji pokok (Rp)", Kind: export.Money},
			{Header: "Tunjangan (Rp)", Kind: export.Money}, {Header: "Potongan (Rp)", Kind: export.Money},
			{Header: "Gaji bersih (Rp)", Kind: export.Money}, {Header: "Status", Kind: export.Text},
		},
		Rows: make([][]any, 0, len(payrolls)),
	}
	totals := payrollTotals{}
	for _, payroll := range payrolls {
		totals = totals.add(payroll)
		section.Rows = append(section.Rows, []any{
			periodLabel(payroll.Period), privacy.MaskIdentity(payroll.IdentityNumber), payroll.EmployeeName, payroll.Position,
			projectText(payroll.ProjectCode, payroll.ProjectName), payroll.BasicPay, payroll.Allowance, payroll.Deduction, payroll.NetPay,
			paidLabel(payroll.Paid),
		})
	}
	section.Total = []any{totalLabel, nil, nil, nil, nil, totals.basicPay, totals.allowance, totals.deduction, totals.netPay, nil}
	return section
}

func buildSales(ctx context.Context, repository Repository, filter DatasetFilter, today time.Time) ([]export.Section, error) {
	contracts, err := repository.ListContracts(ctx, filter, today)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	installments, err := repository.ListInstallments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list installments: %w", err)
	}
	return []export.Section{contractSection(contracts), installmentSection(installments, today)}, nil
}

func contractSection(contracts []ContractRow) export.Section {
	section := export.Section{
		Title: "Kontrak penjualan",
		Columns: []export.Column{
			{Header: "Nomor", Kind: export.Text}, {Header: "Tanggal", Kind: export.Date}, {Header: "Pelanggan", Kind: export.Text},
			{Header: "Unit", Kind: export.Text}, {Header: "Proyek", Kind: export.Text}, {Header: "Jenis", Kind: export.Text},
			{Header: "Status", Kind: export.Text}, {Header: "Nilai (Rp)", Kind: export.Money}, {Header: "Terbayar (Rp)", Kind: export.Money},
			{Header: "Sisa (Rp)", Kind: export.Money}, {Header: "Cicilan lunas", Kind: export.Text}, {Header: "Terlambat", Kind: export.Integer},
		},
		Rows: make([][]any, 0, len(contracts)),
	}
	totalValue, totalPaid, totalRemaining := int64(0), int64(0), int64(0)
	for _, contract := range contracts {
		paid, remaining := contractPaid(contract), contractRemaining(contract)
		if contract.Status != cancelledStatus {
			totalValue, totalPaid, totalRemaining = totalValue+contract.Value, totalPaid+paid, totalRemaining+remaining
		}
		section.Rows = append(section.Rows, []any{
			contract.Number, contract.Date, contract.CustomerName, contract.UnitCode, projectText(contract.ProjectCode, contract.ProjectName),
			labelOf(contractTypeLabels, contract.Type), labelOf(contractStatusLabels, contract.Status), contract.Value, paid,
			remaining, fmt.Sprintf("%d dari %d", contract.PaidCount, contract.InstallmentCount), contract.OverdueCount,
		})
	}
	section.Total = []any{totalLabel, nil, nil, nil, nil, nil, nil, totalValue, totalPaid, totalRemaining, nil, nil}
	return section
}

func contractRemaining(contract ContractRow) int64 {
	if contract.Status == cancelledStatus {
		return 0
	}
	return contract.Value - contractPaid(contract)
}

func contractPaid(contract ContractRow) int64 {
	if contract.Status == paidStatus {
		return contract.Value
	}
	return min(contract.PaidTotal, contract.Value)
}

func installmentSection(installments []InstallmentRow, today time.Time) export.Section {
	section := export.Section{
		Title: "Status cicilan",
		Columns: []export.Column{
			{Header: "Kontrak", Kind: export.Text}, {Header: "Pelanggan", Kind: export.Text}, {Header: "Cicilan ke", Kind: export.Integer},
			{Header: "Jatuh tempo", Kind: export.Date}, {Header: "Nilai (Rp)", Kind: export.Money}, {Header: "Tanggal bayar", Kind: export.Date},
			{Header: "Status", Kind: export.Text}, {Header: "Hari terlambat", Kind: export.Integer},
		},
		Rows: make([][]any, 0, len(installments)),
	}
	totalAmount := int64(0)
	for _, installment := range installments {
		settled := installment.Status == duedate.Paid
		totalAmount += installment.Amount
		section.Rows = append(section.Rows, []any{
			installment.ContractNumber, installment.CustomerName, int64(installment.InstallmentNumber), installment.DueDate,
			installment.Amount, dateValue(installment.PaidDate), labelOf(dueStatusLabels, duedate.Status(installment.DueDate, settled, today)),
			int64(duedate.DaysOverdue(installment.DueDate, settled, today)),
		})
	}
	section.Total = []any{totalLabel, nil, nil, nil, totalAmount, nil, nil, nil}
	return section
}

func projectText(code, name string) string {
	if code == "" && name == "" {
		return noProjectLabel
	}
	return code + " " + name
}

func dateValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func labelOf(labels map[string]string, value string) string {
	if label, ok := labels[value]; ok {
		return label
	}
	return value
}

func paidLabel(paid bool) string {
	if paid {
		return "Dibayar"
	}
	return "Belum dibayar"
}

func absorption(realized, value int64) int64 {
	if value <= 0 {
		return 0
	}
	return int64(math.Round(float64(realized) * fullPercent / float64(value)))
}
