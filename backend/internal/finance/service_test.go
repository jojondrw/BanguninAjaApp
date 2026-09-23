package finance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	entries []JournalEntry
	lines   []JournalLine
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) CreateJournalEntry(_ context.Context, entry *JournalEntry) error {
	f.entries = append(f.entries, *entry)
	return nil
}

func (f *fakeRepository) CreateJournalLines(_ context.Context, lines []JournalLine) error {
	f.lines = append(f.lines, lines...)
	return nil
}

func TestValidateJournalLines(t *testing.T) {
	cash, revenue := uuid.New(), uuid.New()
	cases := []struct {
		name  string
		lines []JournalLineRequest
		want  error
	}{
		{"balanced", []JournalLineRequest{{AccountID: cash, Debit: 500}, {AccountID: revenue, Credit: 500}}, nil},
		{"unbalanced", []JournalLineRequest{{AccountID: cash, Debit: 500}, {AccountID: revenue, Credit: 400}}, errJournalUnbalanced},
		{"both sides", []JournalLineRequest{{AccountID: cash, Debit: 500, Credit: 500}, {AccountID: revenue}}, errJournalSingleSided},
		{"empty line", []JournalLineRequest{{AccountID: cash}, {AccountID: revenue}}, errJournalSingleSided},
	}
	for _, tc := range cases {
		if got := validateJournalLines(tc.lines); !errors.Is(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestUnbalancedJournalIsNotStored(t *testing.T) {
	repository := &fakeRepository{}
	_, err := NewService(repository).RecordJournalEntry(context.Background(), JournalEntryRequest{
		Number: "JU-001",
		Lines:  []JournalLineRequest{{AccountID: uuid.New(), Debit: 100}, {AccountID: uuid.New(), Credit: 90}},
	})
	if !errors.Is(err, errJournalUnbalanced) {
		t.Fatalf("got %v, want errJournalUnbalanced", err)
	}
	if len(repository.entries) != 0 || len(repository.lines) != 0 {
		t.Fatal("nothing must be stored")
	}
}

func TestNewJournalLinesMapsCreditToKredit(t *testing.T) {
	entryID := uuid.New()
	lines := newJournalLines(entryID, []JournalLineRequest{{AccountID: uuid.New(), Credit: 750}})
	if lines[0].Kredit != 750 || lines[0].Debit != 0 || lines[0].JournalEntryID != entryID {
		t.Fatalf("got %+v", lines[0])
	}
}

func TestCashFlowFillsEmptyMonths(t *testing.T) {
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	rows := []CashFlowRow{{Period: from, CashIn: 1000, CashOut: 400}, {Period: from.AddDate(0, 2, 0), CashIn: 200, CashOut: 900}}

	response := newCashFlowResponse(from, to, rows, 5000)
	if len(response.Periods) != 3 || response.Periods[1].Period != "2026-05" || response.Periods[1].CashIn != 0 {
		t.Fatalf("got %+v", response.Periods)
	}
	if response.TotalIn != 1200 || response.TotalOut != 1300 || response.Net != -100 || response.Balance != 5000 {
		t.Fatalf("got totals %+v", response)
	}
}

func TestCashFlowRangeDefaultsToSixMonths(t *testing.T) {
	today := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	from, to := cashFlowRange(CashFlowQuery{}, today)
	if !to.Equal(today) || !from.Equal(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("got %s to %s", from, to)
	}
}

func TestLedgerBalanceStartsFromOpening(t *testing.T) {
	lines := ledgerLines([]LedgerRow{{Debit: 300, RunningBalance: 300}, {Credit: 100, RunningBalance: 200}}, 1000)
	if lines[0].Balance != 1300 || lines[1].Balance != 1200 {
		t.Fatalf("got %+v", lines)
	}
}

func TestAbsorption(t *testing.T) {
	if got := absorption(625, 1000); got != 63 {
		t.Fatalf("got %d, want 63", got)
	}
	if got := absorption(100, 0); got != 0 {
		t.Fatalf("zero budget must give 0, got %d", got)
	}
}
