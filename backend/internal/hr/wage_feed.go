package hr

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const fullMonthDays = 25

type WageRow struct {
	EmployeeID     uuid.UUID
	Name           string
	EmploymentType string
	Gross          int64
	Days           int
}

type WageFeed struct {
	db *gorm.DB
}

func NewWageFeed(db *gorm.DB) WageFeed {
	return WageFeed{db: db}
}

func (f WageFeed) PaidWages(ctx context.Context, projectID uuid.UUID, period string) ([]WageRow, error) {
	var rows []struct {
		EmployeeID     uuid.UUID
		Name           string
		EmploymentType string
		Gross          int64
		Present        int
	}
	err := f.db.WithContext(ctx).Table("payroll AS p").
		Select(`p.employee_id, e.name, e.employment_type, p.basic_pay + p.allowance AS gross,
			(SELECT COUNT(*) FROM attendance a WHERE a.employee_id = p.employee_id
				AND to_char(a.date, 'YYYY-MM') = p.period AND a.status = ?) AS present`, presentStatus).
		Joins("JOIN employee e ON e.id = p.employee_id").
		Where("e.project_id = ? AND p.period = ? AND p.paid", projectID, period).
		Order("e.name, p.employee_id").
		Scan(&rows).Error
	if err != nil {
		return nil, database.Translate(err)
	}
	wages := make([]WageRow, len(rows))
	for index, row := range rows {
		days := fullMonthDays
		if row.EmploymentType == employmentDaily {
			days = row.Present
		}
		wages[index] = WageRow{EmployeeID: row.EmployeeID, Name: row.Name, EmploymentType: row.EmploymentType, Gross: row.Gross, Days: days}
	}
	return wages, nil
}
