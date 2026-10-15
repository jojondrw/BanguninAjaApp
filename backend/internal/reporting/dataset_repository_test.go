package reporting

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var (
	fixedNow   = time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	october1   = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	october31  = time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	projectRow = databasetest.Result{Columns: []string{"code", "name"}, Rows: [][]driver.Value{{"PRJ-UJI", "UJI Cibubur"}}}
)

func previewWith(t *testing.T, query PreviewQuery, results ...databasetest.Result) (PreviewResponse, []databasetest.Query) {
	t.Helper()
	db, recorder := databasetest.Open(t, results...)
	reports := &service{repository: NewRepository(db), now: func() time.Time { return fixedNow }}
	preview, err := reports.PreviewReport(context.Background(), access.RoleAdmin, query)
	if err != nil {
		t.Fatalf("preview %s: %v", query.Type, err)
	}
	return preview, recorder.Queries()
}

func assertSQLContains(t *testing.T, sql string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(sql, fragment) {
			t.Errorf("SQL is missing %q:\n%s", fragment, sql)
		}
	}
}

func assertRow(t *testing.T, got []any, want ...any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got row %#v\nwant %#v", got, want)
	}
}

func TestCashFlowReportFillsMonthsAndRunsBalanceFromOpening(t *testing.T) {
	projectID := uuid.New()
	preview, queries := previewWith(t,
		PreviewQuery{Type: typeCashFlow, ProjectID: &projectID, PeriodStart: &october1, PeriodEnd: &october31},
		projectRow,
		databasetest.Result{Columns: []string{"period", "cash_in", "cash_out"}, Rows: [][]driver.Value{{october1, int64(1000), int64(400)}}},
		databasetest.Result{Columns: []string{"balance"}, Rows: [][]driver.Value{{int64(5000)}}},
		databasetest.Result{
			Columns: []string{"date", "type", "amount", "note", "account_code", "account_name", "project_code", "project_name"},
			Rows: [][]driver.Value{
				{october1, "in", int64(1000), "UJI- termin", "1110", "Kas", "PRJ-UJI", "UJI Cibubur"},
				{october1, "out", int64(400), "UJI- semen", "5110", "Beban Material", nil, nil},
			},
		},
	)

	assertSQLContains(t, queries[1].SQL, "date_trunc('month', cash_transaction.date)::date AS period",
		"cash_transaction.project_id = $3", "cash_transaction.date >= $4", "cash_transaction.date <= $5", "GROUP BY")
	assertSQLContains(t, queries[2].SQL, "cash_transaction.date < $2", "cash_transaction.project_id = $3")
	assertSQLContains(t, queries[3].SQL, "JOIN account ON account.id = cash_transaction.account_id",
		"LEFT JOIN project ON project.id = cash_transaction.project_id", "ORDER BY cash_transaction.date ASC")

	months := preview.Sections[0]
	assertRow(t, months.Rows[0], "Saldo awal", nil, nil, nil, int64(5000))
	assertRow(t, months.Rows[1], "Okt 2026", int64(1000), int64(400), int64(600), int64(5600))
	assertRow(t, months.Total, "Total", int64(1000), int64(400), int64(600), int64(5600))
	lines := preview.Sections[1]
	assertRow(t, lines.Rows[1], october1, "5110 Beban Material", "Kantor pusat", "UJI- semen", nil, int64(400))
	if preview.Filters != "Proyek PRJ-UJI UJI Cibubur · Periode 01/10/2026 – 31/10/2026" {
		t.Fatalf("got filters %q", preview.Filters)
	}
}

func TestBudgetReportFiltersYearsAndComputesAbsorption(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	preview, queries := previewWith(t,
		PreviewQuery{Type: typeBudget, PeriodStart: &start, PeriodEnd: &october31},
		databasetest.Result{
			Columns: []string{"year", "value", "project_code", "project_name", "realized"},
			Rows:    [][]driver.Value{{int64(2026), int64(1000), "PRJ-UJI", "UJI Cibubur", int64(250)}},
		},
	)

	assertSQLContains(t, queries[0].SQL, "JOIN project ON project.id = budget.project_id", "cash_transaction.type = $1",
		"budget.year >= $2", "budget.year <= $3")
	if queries[0].Args[0] != cashOut || queries[0].Args[1] != int64(2025) || queries[0].Args[2] != int64(2026) {
		t.Fatalf("got args %v", queries[0].Args)
	}
	assertRow(t, preview.Sections[0].Rows[0], "PRJ-UJI UJI Cibubur", "2026", int64(1000), int64(250), int64(750), int64(25))
	assertRow(t, preview.Sections[0].Total, "Total", nil, int64(1000), int64(250), int64(750), int64(25))
}

func TestProjectReportListsStatusProgressAndBudgetItems(t *testing.T) {
	projectID := uuid.New()
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	preview, queries := previewWith(t,
		PreviewQuery{Type: typeProjects, ProjectID: &projectID},
		projectRow,
		databasetest.Result{
			Columns: []string{"code", "name", "type", "status", "progress", "start_date", "target_end_date", "contract_value", "budget_item_total"},
			Rows:    [][]driver.Value{{"PRJ-UJI", "UJI Cibubur", "Perumahan", "ongoing", int64(40), start, nil, int64(9000), int64(7000)}},
		},
	)

	assertSQLContains(t, queries[1].SQL, "project.id = $1", "SUM(budget_item.total)", "AS budget_item_total")
	assertRow(t, preview.Sections[0].Rows[0], "PRJ-UJI", "UJI Cibubur", "Perumahan", "Berjalan", int64(40), start, nil, int64(9000), int64(7000))
}

func TestAgingReportBucketsOpenItemsByDaysLate(t *testing.T) {
	openItemColumns := []string{"reference", "party_name", "project_code", "project_name", "due_date", "amount", "paid_amount"}
	preview, queries := previewWith(t,
		PreviewQuery{Type: typeAging},
		databasetest.Result{Columns: openItemColumns, Rows: [][]driver.Value{
			{"UJI-AR-1", "UJI Budi", "PRJ-UJI", "UJI Cibubur", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), int64(1000), int64(300)},
		}},
		databasetest.Result{Columns: openItemColumns, Rows: [][]driver.Value{
			{"UJI-AP-1", "UJI Vendor", nil, nil, time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC), int64(500), int64(0)},
		}},
	)

	assertSQLContains(t, queries[0].SQL, "receivable.paid_amount < receivable.amount", "JOIN customer ON customer.id = receivable.customer_id")
	assertSQLContains(t, queries[1].SQL, "payable.paid_amount < payable.amount", "JOIN vendor ON vendor.id = payable.vendor_id")
	summary := preview.Sections[0]
	assertRow(t, summary.Rows[0], "0–30 hari", int64(0), int64(500))
	assertRow(t, summary.Rows[2], "61–90 hari", int64(700), int64(0))
	assertRow(t, summary.Total, "Total", int64(700), int64(500))
	receivable := preview.Sections[1].Rows[0]
	if receivable[7] != int64(63) || receivable[8] != "61–90 hari" {
		t.Fatalf("got receivable %#v", receivable)
	}
	if !strings.Contains(preview.Filters, "Posisi per 03/10/2026") {
		t.Fatalf("got filters %q", preview.Filters)
	}
}

func TestStockReportScopesWarehousesAndLowStockToProject(t *testing.T) {
	projectID := uuid.New()
	preview, queries := previewWith(t,
		PreviewQuery{Type: typeStock, ProjectID: &projectID},
		projectRow,
		databasetest.Result{
			Columns: []string{"warehouse_code", "warehouse_name", "material_code", "material_name", "category", "unit_code", "quantity"},
			Rows:    [][]driver.Value{{"GDG-UJI", "UJI Gudang", "MAT-1", "UJI Semen", "Struktur", "sak", 12.5}},
		},
		databasetest.Result{
			Columns: []string{"material_code", "material_name", "unit_code", "minimum_stock", "total_quantity"},
			Rows:    [][]driver.Value{{"MAT-1", "UJI Semen", "sak", 50.0, 12.5}},
		},
	)

	assertSQLContains(t, queries[1].SQL, "JOIN warehouse ON warehouse.id = stock.warehouse_id", "warehouse.project_id = $1")
	assertSQLContains(t, queries[2].SQL, "SUM(stock.quantity) AS quantity", "warehouse.project_id = $1",
		"AS totals ON totals.material_id = material.id", "material.minimum_stock > 0 AND COALESCE(totals.quantity, 0) < material.minimum_stock")
	assertRow(t, preview.Sections[0].Rows[0], "GDG-UJI UJI Gudang", "MAT-1", "UJI Semen", "Struktur", "sak", 12.5)
	assertRow(t, preview.Sections[1].Rows[0], "MAT-1", "UJI Semen", "sak", 12.5, 50.0, 37.5)
}

func TestStockReportRejectsAPeriod(t *testing.T) {
	db, recorder := databasetest.Open(t)
	reports := &service{repository: NewRepository(db), now: func() time.Time { return fixedNow }}
	_, err := reports.PreviewReport(context.Background(), access.RoleAdmin, PreviewQuery{Type: typeStock, PeriodStart: &october1})
	if !errors.Is(err, errReportPeriodUnsupported) || len(recorder.Queries()) != 0 {
		t.Fatalf("got %v with %d queries", err, len(recorder.Queries()))
	}
}

func TestPayrollReportDefaultsToCurrentMonthAndSummarizesUnpaid(t *testing.T) {
	preview, queries := previewWith(t,
		PreviewQuery{Type: typePayroll},
		databasetest.Result{
			Columns: []string{"period", "identity_number", "employee_name", "position", "project_code", "project_name",
				"basic_pay", "allowance", "deduction", "net_pay", "paid"},
			Rows: [][]driver.Value{
				{"2026-10", "UJI-001", "UJI Andi", "Mandor", "PRJ-UJI", "UJI Cibubur", int64(5000), int64(500), int64(100), int64(5400), true},
				{"2026-10", "UJI-002", "UJI Sari", "Staf", nil, nil, int64(4000), int64(0), int64(0), int64(4000), false},
			},
		},
	)

	assertSQLContains(t, queries[0].SQL, "JOIN employee ON employee.id = payroll.employee_id", "payroll.period >= $1", "payroll.period <= $2")
	if queries[0].Args[0] != "2026-10" || queries[0].Args[1] != "2026-10" {
		t.Fatalf("got args %v", queries[0].Args)
	}
	assertRow(t, preview.Sections[0].Rows[0], "Okt 2026", int64(2), int64(9000), int64(500), int64(100), int64(9400), int64(4000))
	if preview.Sections[1].Rows[1][9] != "Belum dibayar" || preview.Sections[1].Rows[1][4] != "Kantor pusat" {
		t.Fatalf("got payroll row %#v", preview.Sections[1].Rows[1])
	}
	if preview.PeriodStart == nil || !preview.PeriodStart.Equal(october1) || !preview.PeriodEnd.Equal(october31) {
		t.Fatalf("got period %v - %v", preview.PeriodStart, preview.PeriodEnd)
	}
}

func TestSalesReportJoinsInstallmentScheduleAndDerivesStatus(t *testing.T) {
	projectID := uuid.New()
	preview, queries := previewWith(t,
		PreviewQuery{Type: typeSales, ProjectID: &projectID, PeriodStart: &october1, PeriodEnd: &october31},
		projectRow,
		databasetest.Result{
			Columns: []string{"number", "date", "type", "status", "value", "customer_name", "unit_code", "project_code", "project_name",
				"paid_total", "installment_count", "paid_count", "overdue_count"},
			Rows: [][]driver.Value{{"UJI-K-1", october1, "installment", "active", int64(12000), "UJI Budi", "A-01", "PRJ-UJI", "UJI Cibubur",
				int64(3000), int64(12), int64(3), int64(1)}},
		},
		databasetest.Result{
			Columns: []string{"contract_number", "customer_name", "installment_number", "due_date", "amount", "paid_date", "status"},
			Rows:    [][]driver.Value{{"UJI-K-1", "UJI Budi", int64(4), time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), int64(1000), nil, "due"}},
		},
	)

	assertSQLContains(t, queries[1].SQL, "LEFT JOIN LATERAL", "FROM installment WHERE installment.contract_id = contract.id",
		"contract.date >= $5", "contract.date <= $6", "unit.project_id = $7")
	if queries[1].Args[3] != duedateToday() {
		t.Fatalf("overdue count must use today, got %v", queries[1].Args[3])
	}
	assertSQLContains(t, queries[2].SQL, "JOIN contract ON contract.id = installment.contract_id",
		"installment.due_date >= $1", "installment.due_date <= $2", "unit.project_id = $3")
	assertRow(t, preview.Sections[0].Rows[0], "UJI-K-1", october1, "UJI Budi", "A-01", "PRJ-UJI UJI Cibubur", "Angsuran", "Aktif",
		int64(12000), int64(3000), int64(9000), "3 dari 12", int64(1))
	installment := preview.Sections[1].Rows[0]
	if installment[6] != "Terlambat" || installment[7] != int64(32) {
		t.Fatalf("got installment %#v", installment)
	}
}

func duedateToday() time.Time {
	return time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
}

func TestUnknownProjectIsRejectedInPreview(t *testing.T) {
	db, _ := databasetest.Open(t, databasetest.Result{Columns: []string{"code", "name"}})
	reports := &service{repository: NewRepository(db), now: func() time.Time { return fixedNow }}
	projectID := uuid.New()
	_, err := reports.PreviewReport(context.Background(), access.RoleAdmin, PreviewQuery{Type: typeProjects, ProjectID: &projectID})
	if !errors.Is(err, errProjectNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestReportListReturnsProjectAndCreatorNames(t *testing.T) {
	reportID, creatorID := uuid.New(), uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}},
		databasetest.Result{Columns: reportRowColumns, Rows: [][]driver.Value{reportRowValues(reportID, creatorID, "pdf")}},
	)

	page, err := NewService(NewRepository(db)).ListReports(context.Background(), access.RoleAdmin, ReportQuery{Type: typeCashFlow})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := page.Items[0]
	if item.ProjectName != "" || item.CreatedByName != "UJI Admin" || item.Type != typeCashFlow {
		t.Fatalf("got %+v", item)
	}
	assertSQLContains(t, recorder.Queries()[1].SQL, "(SELECT users.name FROM users WHERE users.id = report.created_by) AS created_by_name",
		"report.type = $1", "report.type IN ($2,$3,$4,$5,$6,$7,$8)", "ORDER BY report.created_at DESC")
}

var reportRowColumns = []string{
	"id", "name", "type", "scope", "project_id", "period_start", "period_end", "format", "created_by", "created_at", "updated_at",
	"project_code", "project_name", "created_by_name",
}

func reportRowValues(id, creator uuid.UUID, format string) []driver.Value {
	return []driver.Value{
		id.String(), "UJI Arus kas Oktober", typeCashFlow, "Semua proyek", nil, october1, october31, format, creator.String(),
		fixedNow, fixedNow, nil, nil, "UJI Admin",
	}
}
