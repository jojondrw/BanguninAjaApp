package reporting

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

const samplesDirectoryVariable = "REPORT_SAMPLES_DIR"

type sampleRepository struct {
	Repository
}

func sampleDate(month time.Month, day int) time.Time {
	return time.Date(2026, month, day, 0, 0, 0, 0, time.UTC)
}

func (sampleRepository) FindProjectLabel(context.Context, uuid.UUID) (ProjectLabel, error) {
	return ProjectLabel{Code: "DEMO-01", Name: "Contoh Cibubur Residence"}, nil
}

func (sampleRepository) SumCashByMonth(context.Context, DatasetFilter) ([]CashMonthRow, error) {
	return []CashMonthRow{
		{Period: sampleDate(7, 1), CashIn: 850000000, CashOut: 610000000},
		{Period: sampleDate(8, 1), CashIn: 420000000, CashOut: 515000000},
		{Period: sampleDate(9, 1), CashIn: 990000000, CashOut: 700500000},
	}, nil
}

func (sampleRepository) CashBalanceBefore(context.Context, *uuid.UUID, time.Time) (int64, error) {
	return 1250000000, nil
}

func (sampleRepository) ListCashLines(context.Context, DatasetFilter) ([]CashLineRow, error) {
	lines := make([]CashLineRow, 0, 120)
	for index := 0; index < 120; index++ {
		line := CashLineRow{
			Date: sampleDate(time.Month(7+index/40), 1+index%28), Type: cashOut, Amount: int64(2500000 + index*175000),
			Note: fmt.Sprintf("Contoh pembelian material batch %03d (semen, besi, pasir)", index+1), AccountCode: "5110", AccountName: "Beban Material",
			ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence",
		}
		if index%4 == 0 {
			line.Type, line.AccountCode, line.AccountName, line.Note = cashIn, "4110", "Pendapatan Penjualan", fmt.Sprintf("Contoh termin pembayaran unit A-%02d", index/4+1)
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func (sampleRepository) ListBudgetRealization(context.Context, DatasetFilter) ([]BudgetRow, error) {
	return []BudgetRow{
		{ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", Year: 2026, Value: 12000000000, Realized: 7350000000},
		{ProjectCode: "DEMO-02", ProjectName: "Contoh Apartemen Depok", Year: 2026, Value: 25000000000, Realized: 4100000000},
		{ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", Year: 2025, Value: 8000000000, Realized: 8450000000},
	}, nil
}

func (sampleRepository) ListProjectProgress(context.Context, DatasetFilter) ([]ProjectRow, error) {
	start, target := sampleDate(2, 1), time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)
	return []ProjectRow{
		{Code: "DEMO-01", Name: "Contoh Cibubur Residence", Type: "Perumahan", Status: "ongoing", Progress: 62, StartDate: &start, TargetEndDate: &target, ContractValue: 18500000000, BudgetItemTotal: 12100000000},
		{Code: "DEMO-02", Name: "Contoh Apartemen Depok", Type: "Apartemen", Status: "planning", Progress: 5, ContractValue: 42000000000, BudgetItemTotal: 0},
	}, nil
}

func (sampleRepository) ListOpenReceivables(context.Context, DatasetFilter) ([]OpenItemRow, error) {
	return []OpenItemRow{
		{Reference: "DEMO-INV-001", PartyName: "Contoh Budi Santoso", ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", DueDate: sampleDate(6, 15), Amount: 150000000, PaidAmount: 50000000},
		{Reference: "DEMO-INV-002", PartyName: "Contoh Sari Dewi", ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", DueDate: sampleDate(9, 20), Amount: 85000000},
		{Reference: "DEMO-INV-003", PartyName: "Contoh PT Maju", DueDate: sampleDate(10, 25), Amount: 230000000},
	}, nil
}

func (sampleRepository) ListOpenPayables(context.Context, DatasetFilter) ([]OpenItemRow, error) {
	return []OpenItemRow{
		{Reference: "DEMO-PO-014", PartyName: "Contoh CV Semen Jaya", ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", DueDate: sampleDate(8, 10), Amount: 98000000, PaidAmount: 40000000},
		{Reference: "DEMO-PO-015", PartyName: "Contoh PT Baja Prima", DueDate: sampleDate(10, 30), Amount: 176000000},
	}, nil
}

func (sampleRepository) ListStockLevels(context.Context, DatasetFilter) ([]StockRow, error) {
	return []StockRow{
		{WarehouseCode: "GDG-01", WarehouseName: "Contoh Gudang Cibubur", MaterialCode: "MAT-001", MaterialName: "Semen Portland 50 kg", Category: "Struktur", UnitCode: "sak", Quantity: 120},
		{WarehouseCode: "GDG-01", WarehouseName: "Contoh Gudang Cibubur", MaterialCode: "MAT-002", MaterialName: "Besi beton 10 mm", Category: "Struktur", UnitCode: "btg", Quantity: 45.5},
		{WarehouseCode: "GDG-02", WarehouseName: "Contoh Gudang Depok", MaterialCode: "MAT-003", MaterialName: "Pasir beton", Category: "Agregat", UnitCode: "m3", Quantity: 12.25},
	}, nil
}

func (sampleRepository) ListLowStock(context.Context, DatasetFilter) ([]LowStockRow, error) {
	return []LowStockRow{{MaterialCode: "MAT-002", MaterialName: "Besi beton 10 mm", UnitCode: "btg", MinimumStock: 200, TotalQuantity: 45.5}}, nil
}

func (sampleRepository) ListPayrolls(context.Context, DatasetFilter) ([]PayrollRow, error) {
	return []PayrollRow{
		{Period: "2026-10", IdentityNumber: "DEMO-0001", EmployeeName: "Contoh Andi Pratama", Position: "Mandor", ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", BasicPay: 6500000, Allowance: 750000, Deduction: 150000, NetPay: 7100000, Paid: true},
		{Period: "2026-10", IdentityNumber: "DEMO-0002", EmployeeName: "Contoh Rina Lestari", Position: "Staf Keuangan", BasicPay: 7200000, Allowance: 500000, Deduction: 200000, NetPay: 7500000},
	}, nil
}

func (sampleRepository) ListContracts(context.Context, DatasetFilter, time.Time) ([]ContractRow, error) {
	return []ContractRow{
		{Number: "DEMO-SPK-001", Date: sampleDate(10, 2), Type: "installment", Status: "active", Value: 950000000, CustomerName: "Contoh Budi Santoso", UnitCode: "A-01", ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence", PaidTotal: 190000000, InstallmentCount: 24, PaidCount: 4, OverdueCount: 1},
		{Number: "DEMO-SPK-002", Date: sampleDate(10, 3), Type: "cash", Status: "paid", Value: 780000000, CustomerName: "Contoh Sari Dewi", UnitCode: "A-02", ProjectCode: "DEMO-01", ProjectName: "Contoh Cibubur Residence"},
	}, nil
}

func (sampleRepository) ListInstallments(context.Context, DatasetFilter) ([]InstallmentRow, error) {
	paid := sampleDate(10, 1)
	return []InstallmentRow{
		{ContractNumber: "DEMO-SPK-001", CustomerName: "Contoh Budi Santoso", InstallmentNumber: 4, DueDate: sampleDate(10, 1), Amount: 39500000, PaidDate: &paid, Status: "paid"},
		{ContractNumber: "DEMO-SPK-001", CustomerName: "Contoh Budi Santoso", InstallmentNumber: 5, DueDate: sampleDate(10, 31), Amount: 39500000, Status: "not_due"},
	}, nil
}

func TestEveryReportTypeRendersInEveryFormat(t *testing.T) {
	directory := os.Getenv(samplesDirectoryVariable)
	reports := &service{repository: sampleRepository{}, now: func() time.Time { return time.Date(2026, 10, 3, 9, 15, 0, 0, time.UTC) }}
	projectID := uuid.New()
	for key := range reportKinds {
		start, end := samplePeriod(key)
		spec, err := newReportSpec(key, &projectID, start, end, reports.today())
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		document, project, err := reports.buildDocument(context.Background(), spec)
		if err != nil || len(document.Sections) == 0 {
			t.Fatalf("%s: got %d sections, %v", key, len(document.Sections), err)
		}
		for name, format := range fileFormats {
			var content bytes.Buffer
			if err := format.write(&content, document); err != nil || content.Len() == 0 {
				t.Fatalf("%s %s: %v", key, name, err)
			}
			if directory != "" {
				path := filepath.Join(directory, spec.fileName(project.Code, name, reports.today()))
				if err := os.WriteFile(path, content.Bytes(), 0o644); err != nil {
					t.Fatalf("write sample: %v", err)
				}
			}
		}
	}
}

func samplePeriod(key string) (*time.Time, *time.Time) {
	if key != typeCashFlow {
		return nil, nil
	}
	start, end := sampleDate(7, 1), sampleDate(9, 30)
	return &start, &end
}
