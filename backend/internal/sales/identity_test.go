package sales

import (
	"context"
	"testing"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type customerListRepository struct {
	Repository
	filter CustomerFilter
}

func (f *customerListRepository) ListCustomers(_ context.Context, filter CustomerFilter) ([]Customer, int64, error) {
	f.filter = filter
	return []Customer{{Name: "Budi Santoso", IdentityNumber: "3273011504900003"}}, 1, nil
}

func TestCustomerIdentityIsMaskedAndUnsearchableForReadOnlyRoles(t *testing.T) {
	repository := &customerListRepository{}
	page, err := NewService(repository).ListCustomers(context.Background(), CustomerQuery{
		Query: pagination.Query{Page: 1, PageSize: 10}, Search: "3273",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := page.Items[0].IdentityNumber; got != "••••••••••••0003" {
		t.Fatalf("identity = %q, want only the last four digits", got)
	}
	if repository.filter.SearchIdentity {
		t.Fatal("a read-only role must not be able to find customers by identity number")
	}
}

func TestCustomerIdentityStaysWholeForWriters(t *testing.T) {
	repository := &customerListRepository{}
	page, err := NewService(repository).ListCustomers(context.Background(), CustomerQuery{
		Query: pagination.Query{Page: 1, PageSize: 10}, Search: "3273", IdentityVisible: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := page.Items[0].IdentityNumber; got != "3273011504900003" || !repository.filter.SearchIdentity {
		t.Fatalf("identity = %q, search by identity = %v", got, repository.filter.SearchIdentity)
	}
}
