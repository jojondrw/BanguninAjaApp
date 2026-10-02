package hr

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var payrollTotalsResult = databasetest.Result{
	Columns: []string{"count", "gross_pay", "net_pay", "paid_net_pay", "unpaid_net_pay", "unpaid_count"},
	Rows:    [][]driver.Value{{int64(3), "21000000", "20300000", "6850000", "13450000", int64(2)}},
}

func TestSummarizePayrollsAggregatesInDatabase(t *testing.T) {
	db, recorder := databasetest.Open(t, payrollTotalsResult)
	projectID := uuid.New()

	totals, err := NewRepository(db).SummarizePayrolls(context.Background(), PayrollFilter{ProjectID: &projectID, Period: "2026-09"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := PayrollTotals{Count: 3, GrossPay: 21000000, NetPay: 20300000, PaidNetPay: 6850000, UnpaidNetPay: 13450000, UnpaidCount: 2}
	if totals != want {
		t.Fatalf("got %+v, want %+v", totals, want)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want a single aggregate query", len(queries))
	}
	for _, fragment := range []string{
		"COALESCE(SUM(payroll.basic_pay + payroll.allowance), 0) AS gross_pay",
		"COALESCE(SUM(payroll.net_pay) FILTER (WHERE payroll.paid), 0) AS paid_net_pay",
		"COUNT(*) FILTER (WHERE NOT payroll.paid) AS unpaid_count",
		"JOIN employee ON employee.id = payroll.employee_id",
		"employee.project_id = $1",
		"payroll.period = $2",
	} {
		if !strings.Contains(queries[0].SQL, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, queries[0].SQL)
		}
	}
	if strings.Contains(queries[0].SQL, "LIMIT") {
		t.Errorf("summary must cover every payroll, got a limited query:\n%s", queries[0].SQL)
	}
	if len(queries[0].Args) != 2 || queries[0].Args[1] != "2026-09" {
		t.Errorf("got args %v", queries[0].Args)
	}
}

func TestSummarizePayrollsWithoutFilterCoversAllPeriods(t *testing.T) {
	db, recorder := databasetest.Open(t, payrollTotalsResult)

	if _, err := NewRepository(db).SummarizePayrolls(context.Background(), PayrollFilter{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := recorder.Queries()[0]
	if strings.Contains(query.SQL, "payroll.period") || strings.Contains(query.SQL, "employee.project_id") || len(query.Args) != 0 {
		t.Fatalf("unfiltered summary must not narrow rows, got %s %v", query.SQL, query.Args)
	}
}
