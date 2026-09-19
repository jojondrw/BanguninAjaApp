package hr

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

type Employee struct {
	entity.Base
	IdentityNumber string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_employee_identity_number"`
	Name           string     `gorm:"type:varchar(160);not null"`
	Position       string     `gorm:"type:varchar(80);not null"`
	ProjectID      *uuid.UUID `gorm:"type:uuid;index:idx_employee_project"`
	UserID         *uuid.UUID `gorm:"type:uuid;index:idx_employee_user"`
	EmploymentType string     `gorm:"type:varchar(20);not null;index:idx_employee_status"`
	JoinedDate     time.Time  `gorm:"type:date;not null"`
	LeftDate       *time.Time `gorm:"type:date"`
	BaseSalary     int64      `gorm:"not null;default:0"`
}

func (Employee) TableName() string {
	return "employee"
}

type Attendance struct {
	entity.Base
	EmployeeID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_attendance_employee"`
	Date         time.Time  `gorm:"type:date;not null;index:idx_attendance_date"`
	CheckInTime  *time.Time `gorm:"type:time"`
	CheckOutTime *time.Time `gorm:"type:time"`
	Status       string     `gorm:"type:varchar(20);not null;index:idx_attendance_status"`
}

func (Attendance) TableName() string {
	return "attendance"
}

type Payroll struct {
	entity.Base
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index:idx_payroll_employee"`
	Period     string    `gorm:"type:varchar(7);not null;index:idx_payroll_period"`
	BasicPay   int64     `gorm:"not null;default:0"`
	Allowance  int64     `gorm:"not null;default:0"`
	Deduction  int64     `gorm:"not null;default:0"`
	NetPay     int64     `gorm:"not null;default:0"`
	Paid       bool      `gorm:"not null;default:false"`
}

func (Payroll) TableName() string {
	return "payroll"
}

func Entities() []any {
	return []any{&Employee{}, &Attendance{}, &Payroll{}}
}

func Indexes() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_employee_name_trgm ON employee USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_attendance_employee_date ON attendance (employee_id, date DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_employee_active ON employee (project_id) WHERE left_date IS NULL`,
	}
}

func Constraints() []string {
	return []string{
		database.ForeignKey("employee", "project_id", "project", database.DeleteSetNull),
		database.ForeignKey("employee", "user_id", "users", database.DeleteSetNull),
		database.ForeignKey("attendance", "employee_id", "employee", database.DeleteCascade),
		database.ForeignKey("payroll", "employee_id", "employee", database.DeleteCascade),
		database.Check("employee", "employment_type", "employment_type IN ('permanent','contract','daily','subcontractor')"),
		database.Check("employee", "employment_period", "left_date IS NULL OR left_date >= joined_date"),
		database.Check("employee", "base_salary", "base_salary >= 0"),
		database.Check("attendance", "status", "status IN ('present','permitted','sick','absent','holiday')"),
		database.Check("attendance", "jam", "check_out_time IS NULL OR check_in_time IS NULL OR check_out_time >= check_in_time"),
		database.Check("payroll", "period", "period ~ '^[0-9]{4}-[0-9]{2}$'"),
		database.Check("payroll", "value", "basic_pay >= 0 AND allowance >= 0 AND deduction >= 0 AND net_pay >= 0"),
		database.Unique("attendance", "employee_date", "employee_id, date"),
		database.Unique("payroll", "employee_period", "employee_id, period"),
	}
}
