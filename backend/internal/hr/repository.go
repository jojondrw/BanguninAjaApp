package hr

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	attendanceColumns = "attendance.id, attendance.employee_id, employee.name AS employee_name, attendance.date, " +
		"attendance.check_in_time, attendance.check_out_time, attendance.status, attendance.created_at, attendance.updated_at"
	payrollColumns = "payroll.id, payroll.employee_id, employee.name AS employee_name, payroll.period, payroll.basic_pay, " +
		"payroll.allowance, payroll.deduction, payroll.net_pay, payroll.paid, payroll.created_at, payroll.updated_at"
	presentStatus = "present"
)

type EmployeeFilter struct {
	Search         string
	EmploymentType string
	ProjectID      *uuid.UUID
	Active         *bool
	Today          time.Time
	Offset         int
	Limit          int
}

type AttendanceFilter struct {
	EmployeeID *uuid.UUID
	ProjectID  *uuid.UUID
	Status     string
	DateFrom   *time.Time
	DateTo     *time.Time
	Offset     int
	Limit      int
}

type PayrollFilter struct {
	EmployeeID *uuid.UUID
	Period     string
	Paid       *bool
	Offset     int
	Limit      int
}

type EmploymentCount struct {
	EmploymentType string
	Total          int64
}

type AttendanceRow struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	EmployeeName string
	Date         time.Time
	CheckInTime  *string
	CheckOutTime *string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PayrollRow struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	EmployeeName string
	Period       string
	BasicPay     int64
	Allowance    int64
	Deduction    int64
	NetPay       int64
	Paid         bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository interface {
	ListEmployees(ctx context.Context, filter EmployeeFilter) ([]Employee, int64, error)
	CountActiveEmployees(ctx context.Context, today time.Time) ([]EmploymentCount, error)
	FindEmployee(ctx context.Context, id uuid.UUID) (Employee, error)
	CreateEmployee(ctx context.Context, employee *Employee) error
	SaveEmployee(ctx context.Context, employee *Employee) error
	DeleteEmployee(ctx context.Context, id uuid.UUID) error
	EmployeeHasRecords(ctx context.Context, id uuid.UUID) (bool, error)

	ListAttendances(ctx context.Context, filter AttendanceFilter) ([]AttendanceRow, int64, error)
	FindAttendanceRow(ctx context.Context, id uuid.UUID) (AttendanceRow, error)
	FindAttendance(ctx context.Context, id uuid.UUID) (Attendance, error)
	CreateAttendance(ctx context.Context, attendance *Attendance) error
	SaveAttendance(ctx context.Context, attendance *Attendance) error
	DeleteAttendance(ctx context.Context, id uuid.UUID) error
	CountPresentDays(ctx context.Context, employeeID uuid.UUID, from, before time.Time) (int64, error)

	ListPayrolls(ctx context.Context, filter PayrollFilter) ([]PayrollRow, int64, error)
	FindPayrollRow(ctx context.Context, id uuid.UUID) (PayrollRow, error)
	FindPayroll(ctx context.Context, id uuid.UUID) (Payroll, error)
	CreatePayroll(ctx context.Context, payroll *Payroll) error
	SavePayroll(ctx context.Context, payroll *Payroll) error
	DeletePayroll(ctx context.Context, id uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListEmployees(ctx context.Context, filter EmployeeFilter) ([]Employee, int64, error) {
	return database.FindPage[Employee](ctx, r.db, database.Listing{
		Filter: filter.apply,
		Order:  "name ASC, id ASC",
		Offset: filter.Offset,
		Limit:  filter.Limit,
	})
}

func (r *gormRepository) CountActiveEmployees(ctx context.Context, today time.Time) ([]EmploymentCount, error) {
	var counts []EmploymentCount
	err := r.db.WithContext(ctx).
		Model(&Employee{}).
		Select("employment_type, COUNT(*) AS total").
		Where("left_date IS NULL OR left_date >= ?", today).
		Group("employment_type").
		Order("employment_type ASC").
		Scan(&counts).Error
	return counts, database.Translate(err)
}

func (r *gormRepository) FindEmployee(ctx context.Context, id uuid.UUID) (Employee, error) {
	var employee Employee
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&employee).Error
	return employee, database.Translate(err)
}

func (r *gormRepository) CreateEmployee(ctx context.Context, employee *Employee) error {
	return database.Translate(r.db.WithContext(ctx).Create(employee).Error)
}

func (r *gormRepository) SaveEmployee(ctx context.Context, employee *Employee) error {
	return database.Translate(r.db.WithContext(ctx).Save(employee).Error)
}

func (r *gormRepository) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Employee{}, id)
}

func (r *gormRepository) EmployeeHasRecords(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS (SELECT 1 FROM attendance WHERE employee_id = ?) OR EXISTS (SELECT 1 FROM payroll WHERE employee_id = ?)", id, id).
		Scan(&exists).Error
	return exists, database.Translate(err)
}

func (r *gormRepository) ListAttendances(ctx context.Context, filter AttendanceFilter) ([]AttendanceRow, int64, error) {
	var total int64
	if err := r.attendanceRows(ctx).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]AttendanceRow, 0, filter.Limit)
	err := r.attendanceRows(ctx).
		Scopes(filter.apply).
		Select(attendanceColumns).
		Order("attendance.date DESC, employee.name ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindAttendanceRow(ctx context.Context, id uuid.UUID) (AttendanceRow, error) {
	var row AttendanceRow
	err := r.attendanceRows(ctx).Select(attendanceColumns).Where("attendance.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) FindAttendance(ctx context.Context, id uuid.UUID) (Attendance, error) {
	var attendance Attendance
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&attendance).Error
	return attendance, database.Translate(err)
}

func (r *gormRepository) CreateAttendance(ctx context.Context, attendance *Attendance) error {
	return database.Translate(r.db.WithContext(ctx).Create(attendance).Error)
}

func (r *gormRepository) SaveAttendance(ctx context.Context, attendance *Attendance) error {
	return database.Translate(r.db.WithContext(ctx).Save(attendance).Error)
}

func (r *gormRepository) DeleteAttendance(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Attendance{}, id)
}

func (r *gormRepository) CountPresentDays(ctx context.Context, employeeID uuid.UUID, from, before time.Time) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Attendance{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date < ?", employeeID, presentStatus, from, before).
		Count(&total).Error
	return total, database.Translate(err)
}

func (r *gormRepository) ListPayrolls(ctx context.Context, filter PayrollFilter) ([]PayrollRow, int64, error) {
	var total int64
	if err := r.payrollRows(ctx).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, database.Translate(err)
	}

	rows := make([]PayrollRow, 0, filter.Limit)
	err := r.payrollRows(ctx).
		Scopes(filter.apply).
		Select(payrollColumns).
		Order("payroll.period DESC, employee.name ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, database.Translate(err)
}

func (r *gormRepository) FindPayrollRow(ctx context.Context, id uuid.UUID) (PayrollRow, error) {
	var row PayrollRow
	err := r.payrollRows(ctx).Select(payrollColumns).Where("payroll.id = ?", id).Take(&row).Error
	return row, database.Translate(err)
}

func (r *gormRepository) FindPayroll(ctx context.Context, id uuid.UUID) (Payroll, error) {
	var payroll Payroll
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&payroll).Error
	return payroll, database.Translate(err)
}

func (r *gormRepository) CreatePayroll(ctx context.Context, payroll *Payroll) error {
	return database.Translate(r.db.WithContext(ctx).Create(payroll).Error)
}

func (r *gormRepository) SavePayroll(ctx context.Context, payroll *Payroll) error {
	return database.Translate(r.db.WithContext(ctx).Save(payroll).Error)
}

func (r *gormRepository) DeletePayroll(ctx context.Context, id uuid.UUID) error {
	return r.deleteByID(ctx, &Payroll{}, id)
}

func (r *gormRepository) attendanceRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&Attendance{}).Joins("JOIN employee ON employee.id = attendance.employee_id")
}

func (r *gormRepository) payrollRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&Payroll{}).Joins("JOIN employee ON employee.id = payroll.employee_id")
}

func (r *gormRepository) deleteByID(ctx context.Context, model any, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(model)
	if result.Error != nil {
		return database.Translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (f EmployeeFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("(name ILIKE ? OR identity_number ILIKE ?)", pattern, pattern)
	}
	if f.EmploymentType != "" {
		db = db.Where("employment_type = ?", f.EmploymentType)
	}
	if f.ProjectID != nil {
		db = db.Where("project_id = ?", *f.ProjectID)
	}
	if f.Active != nil && *f.Active {
		db = db.Where("(left_date IS NULL OR left_date >= ?)", f.Today)
	}
	if f.Active != nil && !*f.Active {
		db = db.Where("left_date < ?", f.Today)
	}
	return db
}

func (f AttendanceFilter) apply(db *gorm.DB) *gorm.DB {
	if f.EmployeeID != nil {
		db = db.Where("attendance.employee_id = ?", *f.EmployeeID)
	}
	if f.ProjectID != nil {
		db = db.Where("employee.project_id = ?", *f.ProjectID)
	}
	if f.Status != "" {
		db = db.Where("attendance.status = ?", f.Status)
	}
	if f.DateFrom != nil {
		db = db.Where("attendance.date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		db = db.Where("attendance.date <= ?", *f.DateTo)
	}
	return db
}

func (f PayrollFilter) apply(db *gorm.DB) *gorm.DB {
	if f.EmployeeID != nil {
		db = db.Where("payroll.employee_id = ?", *f.EmployeeID)
	}
	if f.Period != "" {
		db = db.Where("payroll.period = ?", f.Period)
	}
	if f.Paid != nil {
		db = db.Where("payroll.paid = ?", *f.Paid)
	}
	return db
}
