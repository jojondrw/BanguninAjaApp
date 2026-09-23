package pagination

import "testing"

func TestQueryDefaults(t *testing.T) {
	var query Query
	if query.Number() != 1 || query.Size() != defaultPageSize || query.Offset() != 0 {
		t.Fatalf("got page %d size %d offset %d", query.Number(), query.Size(), query.Offset())
	}
}

func TestQueryOffset(t *testing.T) {
	query := Query{Page: 3, PageSize: 25}
	if query.Offset() != 50 {
		t.Fatalf("got offset %d, want 50", query.Offset())
	}
}

func TestNewPageCountsTotalPages(t *testing.T) {
	page := New([]int{1, 2}, Query{Page: 2, PageSize: 10}, 21)
	if page.TotalPages != 3 || page.Page != 2 || page.PageSize != 10 {
		t.Fatalf("got %+v", page)
	}
	if empty := New([]int{}, Query{}, 0); empty.TotalPages != 0 || empty.Items == nil {
		t.Fatalf("got %+v", empty)
	}
}
