package hr

import (
	"context"
	"testing"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type employeeListRepository struct {
	Repository
	filter EmployeeFilter
}

func (f *employeeListRepository) ListEmployees(_ context.Context, filter EmployeeFilter) ([]Employee, int64, error) {
	f.filter = filter
	return []Employee{{Name: "Andi Pratama", IdentityNumber: "3204122003880001"}}, 1, nil
}

func TestEmployeeIdentityIsMaskedForReadOnlyRolesOnly(t *testing.T) {
	cases := []struct {
		visible bool
		want    string
	}{
		{visible: false, want: "••••••••••••0001"},
		{visible: true, want: "3204122003880001"},
	}
	for _, tc := range cases {
		repository := &employeeListRepository{}
		page, err := NewService(repository).ListEmployees(context.Background(), EmployeeQuery{
			Query: pagination.Query{Page: 1, PageSize: 10}, Search: "3204", IdentityVisible: tc.visible,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := page.Items[0].IdentityNumber; got != tc.want {
			t.Errorf("visible %v: identity = %q, want %q", tc.visible, got, tc.want)
		}
		if repository.filter.SearchIdentity != tc.visible {
			t.Errorf("visible %v: search by identity = %v", tc.visible, repository.filter.SearchIdentity)
		}
	}
}
