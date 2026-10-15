package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	created *Report
}

func (f *fakeRepository) CreateReport(_ context.Context, report *Report) error {
	f.created = report
	return nil
}

func TestCreateReportRecordsCreator(t *testing.T) {
	repository := &fakeRepository{}
	userID := uuid.New()

	report, err := NewService(repository).CreateReport(context.Background(), userID, ReportRequest{
		Name: "  Arus kas konsolidasi ", Type: "cash_flow", Format: "pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.CreatedBy != userID || report.Name != "Arus kas konsolidasi" {
		t.Fatalf("got %+v", report)
	}
}

func TestReportPeriodMustBeOrdered(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, -1)
	repository := &fakeRepository{}

	_, err := NewService(repository).CreateReport(context.Background(), uuid.New(), ReportRequest{Type: typeBudget, PeriodStart: &start, PeriodEnd: &end})
	if !errors.Is(err, errInvalidDateRange) || repository.created != nil {
		t.Fatalf("got %v created %v", err, repository.created)
	}
}

func TestCreateReportStoresCanonicalTypeAndDefaultMonth(t *testing.T) {
	repository := &fakeRepository{}
	reports := &service{repository: repository, now: func() time.Time { return time.Date(2026, 10, 3, 8, 0, 0, 0, time.UTC) }}

	_, err := reports.CreateReport(context.Background(), uuid.New(), ReportRequest{Name: "UJI kas", Type: "cash_flow", Format: "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	created := repository.created
	if created.Type != typeCashFlow || created.PeriodStart.Format(isoDate) != "2026-10-01" || created.PeriodEnd.Format(isoDate) != "2026-10-31" {
		t.Fatalf("got %+v", created)
	}
}

func TestCreateReportRejectsUnknownTypeAndPeriodForStock(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		request ReportRequest
		want    error
	}{
		{ReportRequest{Name: "UJI", Type: "neraca", Format: "pdf"}, errReportTypeInvalid},
		{ReportRequest{Name: "UJI", Type: typeStock, Format: "pdf", PeriodStart: &start}, errReportPeriodUnsupported},
	}
	for _, tc := range cases {
		repository := &fakeRepository{}
		_, err := NewService(repository).CreateReport(context.Background(), uuid.New(), tc.request)
		if !errors.Is(err, tc.want) || repository.created != nil {
			t.Fatalf("%s: got %v", tc.request.Type, err)
		}
	}
}

func TestReportFileNames(t *testing.T) {
	today := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	october := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	octoberEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		spec    reportSpec
		project string
		want    string
	}{
		{reportSpec{kind: reportKinds[typeCashFlow], periodStart: &october, periodEnd: &octoberEnd}, "", "laporan-arus-kas-2026-10.pdf"},
		{reportSpec{kind: reportKinds[typeBudget], periodStart: &yearStart, periodEnd: &yearEnd}, "PRJ 001", "laporan-anggaran-prj-001-2026.pdf"},
		{reportSpec{kind: reportKinds[typeAging], periodStart: &october, periodEnd: &yearEnd}, "", "laporan-piutang-utang-2026-10-01_2026-12-31.pdf"},
		{reportSpec{kind: reportKinds[typeStock]}, "", "laporan-stok-2026-10-03.pdf"},
	}
	for _, tc := range cases {
		if got := tc.spec.fileName(tc.project, "pdf", today); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestCancelledContractsLeaveTheTotals(t *testing.T) {
	section := contractSection([]ContractRow{
		{Number: "UJI-1", Status: "active", Value: 1000, PaidTotal: 400},
		{Number: "UJI-2", Status: cancelledStatus, Value: 500},
		{Number: "UJI-3", Status: paidStatus, Value: 300},
	})
	if section.Rows[1][9] != int64(0) {
		t.Fatalf("cancelled contract must not show a remaining balance, got %v", section.Rows[1][9])
	}
	if section.Total[7] != int64(1300) || section.Total[8] != int64(700) || section.Total[9] != int64(600) {
		t.Fatalf("got total %v", section.Total)
	}
}
