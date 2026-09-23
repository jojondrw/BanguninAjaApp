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

	_, err := NewService(repository).CreateReport(context.Background(), uuid.New(), ReportRequest{PeriodStart: &start, PeriodEnd: &end})
	if !errors.Is(err, errInvalidDateRange) || repository.created != nil {
		t.Fatalf("got %v created %v", err, repository.created)
	}
}
