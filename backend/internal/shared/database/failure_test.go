package database

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

func TestTranslate(t *testing.T) {
	cases := []struct {
		input error
		want  error
	}{
		{gorm.ErrRecordNotFound, ErrNotFound},
		{fmt.Errorf("%w: pg", gorm.ErrDuplicatedKey), ErrDuplicate},
		{fmt.Errorf("%w: pg", gorm.ErrForeignKeyViolated), ErrReferenceViolated},
		{fmt.Errorf("%w: pg", gorm.ErrCheckConstraintViolated), ErrCheckViolated},
		{&pgconn.PgError{Code: restrictViolationCode}, ErrReferenceViolated},
	}
	for _, tc := range cases {
		if got := Translate(tc.input); !errors.Is(got, tc.want) {
			t.Errorf("Translate(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
	if Translate(nil) != nil {
		t.Error("nil must stay nil")
	}
}

func TestErrorMapResolve(t *testing.T) {
	notFound := apperror.NotFound("thing_not_found", "Tidak ada")
	mapping := ErrorMap{NotFound: notFound}

	if err := mapping.Resolve(nil); err != nil {
		t.Fatalf("nil must resolve to nil, got %v", err)
	}
	if got := apperror.From(mapping.Resolve(ErrNotFound)); got.Code != notFound.Code {
		t.Fatalf("got code %s", got.Code)
	}
	if got := apperror.From(mapping.Resolve(ErrDuplicate)); got.Status != http.StatusInternalServerError {
		t.Fatalf("unmapped error must become internal, got %d", got.Status)
	}
	conflict := apperror.Conflict("busy", "Sibuk")
	if got := apperror.From(mapping.Resolve(conflict)); got.Code != "busy" {
		t.Fatalf("existing app error must pass through, got %s", got.Code)
	}
}
