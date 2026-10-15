package sales

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func TestSummarizeInstallmentsAggregatesEveryInstallmentOfTheContract(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{
		Columns: []string{"count", "paid_count", "total", "paid_total", "overdue_count", "last_number"},
		Rows:    [][]driver.Value{{int64(240), int64(100), int64(1200000000), int64(500000000), int64(3), int64(240)}},
	})
	contractID := uuid.New()
	today := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)

	totals, err := NewRepository(db).SummarizeInstallments(context.Background(), contractID, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := InstallmentTotals{Count: 240, PaidCount: 100, Total: 1200000000, PaidTotal: 500000000, OverdueCount: 3, LastNumber: 240}
	if totals != want {
		t.Fatalf("got %+v, want %+v", totals, want)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want a single aggregate query", len(queries))
	}
	query := queries[0]
	for _, fragment := range []string{
		"COUNT(*) FILTER (WHERE status = $1) AS paid_count",
		"COALESCE(SUM(amount) FILTER (WHERE status = $2), 0) AS paid_total",
		"COUNT(*) FILTER (WHERE status <> $3 AND due_date < $4) AS overdue_count",
		"COALESCE(MAX(installment_number), 0) AS last_number",
		"contract_id = $5",
	} {
		if !strings.Contains(query.SQL, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, query.SQL)
		}
	}
	if strings.Contains(query.SQL, "LIMIT") {
		t.Errorf("summary must cover every installment, got a limited query:\n%s", query.SQL)
	}
	if len(query.Args) != 5 || query.Args[0] != settledStatus || query.Args[3] != today || query.Args[4] != contractID.String() {
		t.Fatalf("got args %v", query.Args)
	}
}

type summaryRepository struct {
	Repository
	findErr error
	totals  InstallmentTotals
	asked   time.Time
}

func (f *summaryRepository) FindContractRow(context.Context, uuid.UUID) (ContractRow, error) {
	return ContractRow{}, f.findErr
}

func (f *summaryRepository) SummarizeInstallments(_ context.Context, _ uuid.UUID, today time.Time) (InstallmentTotals, error) {
	f.asked = today
	return f.totals, nil
}

func TestSummarizeInstallmentsUsesTodayAndMapsTotals(t *testing.T) {
	repository := &summaryRepository{totals: InstallmentTotals{Count: 12, PaidCount: 5, Total: 120, PaidTotal: 50, OverdueCount: 2, LastNumber: 12}}
	service := &service{repository: repository, now: func() time.Time { return time.Date(2026, 10, 3, 15, 30, 0, 0, time.UTC) }}

	summary, err := service.SummarizeInstallments(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := InstallmentSummaryResponse{Count: 12, PaidCount: 5, Total: 120, PaidTotal: 50, OverdueCount: 2, LastNumber: 12}
	if summary != want {
		t.Fatalf("got %+v, want %+v", summary, want)
	}
	if !repository.asked.Equal(time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("overdue must be counted against the start of today, got %v", repository.asked)
	}
}

func TestSummarizeInstallmentsOfUnknownContractIsNotFound(t *testing.T) {
	service := &service{repository: &summaryRepository{findErr: database.ErrNotFound}, now: time.Now}

	_, err := service.SummarizeInstallments(context.Background(), uuid.New())
	if !errors.Is(err, errContractNotFound) {
		t.Fatalf("got %v, want contract_not_found", err)
	}
}
