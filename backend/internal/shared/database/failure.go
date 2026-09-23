package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

const restrictViolationCode = "23001"

var (
	ErrNotFound          = errors.New("record not found")
	ErrDuplicate         = errors.New("duplicate record")
	ErrReferenceViolated = errors.New("reference violated")
	ErrCheckViolated     = errors.New("check constraint violated")
)

type ErrorMap struct {
	NotFound   *apperror.Error
	Duplicate  *apperror.Error
	Referenced *apperror.Error
	Invalid    *apperror.Error
}

func Translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return fmt.Errorf("%w: %w", ErrDuplicate, err)
	case errors.Is(err, gorm.ErrForeignKeyViolated), isRestrictViolation(err):
		return fmt.Errorf("%w: %w", ErrReferenceViolated, err)
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return fmt.Errorf("%w: %w", ErrCheckViolated, err)
	default:
		return err
	}
}

func isRestrictViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == restrictViolationCode
}

func (m ErrorMap) Resolve(err error) error {
	if err == nil {
		return nil
	}
	if mapped := m.lookup(err); mapped != nil {
		return mapped
	}
	return apperror.From(err)
}

func (m ErrorMap) lookup(err error) *apperror.Error {
	switch {
	case errors.Is(err, ErrNotFound):
		return m.NotFound
	case errors.Is(err, ErrDuplicate):
		return m.Duplicate
	case errors.Is(err, ErrReferenceViolated):
		return m.Referenced
	case errors.Is(err, ErrCheckViolated):
		return m.Invalid
	default:
		return nil
	}
}
