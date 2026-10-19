package finance

import (
	"context"
	"errors"
	"testing"
	"time"
)

type accountRepository struct {
	Repository
	filter CashFlowFilter
	rows   []CashAccountRow
}

func (r *accountRepository) SumCashByAccount(_ context.Context, filter CashFlowFilter) ([]CashAccountRow, error) {
	r.filter = filter
	return r.rows, nil
}

func TestCashByAccountKeepsServerOrderAndUsesCashFlowRange(t *testing.T) {
	repository := &accountRepository{rows: []CashAccountRow{
		{AccountCode: "5100", AccountName: "Beban Konstruksi", Type: "out", Total: 900},
		{AccountCode: "4100", AccountName: "Penjualan Unit", Type: "in", Total: 400},
	}}
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	accounts, err := NewService(repository).SummarizeCashByAccount(context.Background(), CashFlowQuery{DateFrom: &from, DateTo: &to})

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(accounts) != 2 || accounts[0].AccountCode != "5100" || accounts[0].Total != 900 || accounts[1].Type != "in" {
		t.Errorf("accounts = %+v", accounts)
	}
	if !repository.filter.DateFrom.Equal(from) || !repository.filter.DateTo.Equal(to) {
		t.Errorf("filter = %+v, want the requested range", repository.filter)
	}
}

func TestCashByAccountRejectsReversedRange(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := NewService(&accountRepository{}).SummarizeCashByAccount(context.Background(), CashFlowQuery{DateFrom: &from, DateTo: &to})

	if !errors.Is(err, errInvalidDateRange) {
		t.Errorf("err = %v, want errInvalidDateRange", err)
	}
}
