package hr

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	employmentDaily   = "daily"
	attendancePresent = "present"

	clockLayout       = "15:04"
	storedClockLayout = "15:04:05"
	periodLayout      = "2006-01"
)

var (
	errEmployeeNotFound       = apperror.NotFound("employee_not_found", "Karyawan tidak ditemukan")
	errEmployeeIdentityUsed   = apperror.Conflict("employee_identity_number_used", "Nomor induk karyawan sudah dipakai")
	errEmployeeReference      = apperror.Unprocessable("employee_reference_not_found", "Proyek atau akun pengguna tidak ditemukan")
	errEmployeeInUse          = apperror.Conflict("employee_in_use", "Karyawan sudah punya absensi atau penggajian. Isi tanggal keluar, jangan dihapus")
	errEmployeeUnknown        = apperror.Unprocessable("employee_not_found", "Karyawan tidak ditemukan")
	errInvalidEmploymentRange = apperror.Unprocessable("invalid_date_range", "Tanggal keluar tidak boleh lebih awal dari tanggal bergabung")

	errAttendanceNotFound          = apperror.NotFound("attendance_not_found", "Absensi tidak ditemukan")
	errAttendanceRecorded          = apperror.Conflict("attendance_already_recorded", "Karyawan ini sudah punya absensi di tanggal tersebut")
	errAttendanceOutsideEmployment = apperror.Unprocessable("attendance_outside_employment", "Tanggal absensi di luar masa kerja karyawan")
	errAttendanceCheckInRequired   = apperror.Unprocessable("attendance_check_in_required", "Karyawan yang hadir wajib punya jam masuk")
	errAttendanceTimeNotAllowed    = apperror.Unprocessable("attendance_time_not_allowed", "Jam masuk dan pulang hanya diisi untuk karyawan yang hadir")
	errAttendanceTimeOrder         = apperror.Unprocessable("attendance_time_invalid", "Jam pulang tidak boleh lebih awal dari jam masuk")

	errPayrollNotFound          = apperror.NotFound("payroll_not_found", "Penggajian tidak ditemukan")
	errPayrollExists            = apperror.Conflict("payroll_already_exists", "Penggajian karyawan ini untuk periode tersebut sudah ada")
	errPayrollPaid              = apperror.Conflict("payroll_paid", "Penggajian yang sudah dibayar tidak bisa diubah atau dihapus")
	errPayrollPeriod            = apperror.Unprocessable("payroll_period_invalid", "Periode penggajian harus berformat YYYY-MM")
	errPayrollOutsideEmployment = apperror.Unprocessable("payroll_outside_employment", "Periode penggajian di luar masa kerja karyawan")
	errPayrollNetNegative       = apperror.Unprocessable("payroll_net_negative", "Potongan tidak boleh melebihi gaji pokok ditambah tunjangan")
)

var (
	employeeReadErrors      = database.ErrorMap{NotFound: errEmployeeNotFound}
	employeeWriteErrors     = database.ErrorMap{NotFound: errEmployeeNotFound, Duplicate: errEmployeeIdentityUsed, Referenced: errEmployeeReference, Invalid: errInvalidEmploymentRange}
	employeeReferenceErrors = database.ErrorMap{NotFound: errEmployeeUnknown}

	attendanceReadErrors  = database.ErrorMap{NotFound: errAttendanceNotFound}
	attendanceWriteErrors = database.ErrorMap{NotFound: errAttendanceNotFound, Duplicate: errAttendanceRecorded, Referenced: errEmployeeUnknown, Invalid: errAttendanceTimeOrder}

	payrollReadErrors  = database.ErrorMap{NotFound: errPayrollNotFound}
	payrollWriteErrors = database.ErrorMap{NotFound: errPayrollNotFound, Duplicate: errPayrollExists, Referenced: errEmployeeUnknown, Invalid: errPayrollNetNegative}
)

type Service interface {
	ListEmployees(ctx context.Context, query EmployeeQuery) (pagination.Page[EmployeeResponse], error)
	SummarizeEmployees(ctx context.Context) ([]EmploymentCountResponse, error)
	GetEmployee(ctx context.Context, id uuid.UUID) (EmployeeResponse, error)
	CreateEmployee(ctx context.Context, request EmployeeRequest) (EmployeeResponse, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, request EmployeeRequest) (EmployeeResponse, error)
	DeleteEmployee(ctx context.Context, id uuid.UUID) error

	ListAttendances(ctx context.Context, query AttendanceQuery) (pagination.Page[AttendanceResponse], error)
	GetAttendance(ctx context.Context, id uuid.UUID) (AttendanceResponse, error)
	CreateAttendance(ctx context.Context, request AttendanceRequest) (AttendanceResponse, error)
	UpdateAttendance(ctx context.Context, id uuid.UUID, request AttendanceRequest) (AttendanceResponse, error)
	DeleteAttendance(ctx context.Context, id uuid.UUID) error

	ListPayrolls(ctx context.Context, query PayrollQuery) (pagination.Page[PayrollResponse], error)
	GetPayroll(ctx context.Context, id uuid.UUID) (PayrollResponse, error)
	CreatePayroll(ctx context.Context, request PayrollRequest) (PayrollResponse, error)
	UpdatePayroll(ctx context.Context, id uuid.UUID, request PayrollRequest) (PayrollResponse, error)
	PayPayroll(ctx context.Context, id uuid.UUID) (PayrollResponse, error)
	DeletePayroll(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) ListEmployees(ctx context.Context, query EmployeeQuery) (pagination.Page[EmployeeResponse], error) {
	today := s.today()
	employees, total, err := s.repository.ListEmployees(ctx, EmployeeFilter{
		Search:         query.Search,
		EmploymentType: query.EmploymentType,
		ProjectID:      query.ProjectID,
		Active:         query.Active,
		Today:          today,
		Offset:         query.Offset(),
		Limit:          query.Size(),
	})
	if err != nil {
		return pagination.Page[EmployeeResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(employees, func(employee Employee) EmployeeResponse {
		return newEmployeeResponse(employee, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) SummarizeEmployees(ctx context.Context) ([]EmploymentCountResponse, error) {
	counts, err := s.repository.CountActiveEmployees(ctx, s.today())
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(counts, newEmploymentCountResponse), nil
}

func (s *service) GetEmployee(ctx context.Context, id uuid.UUID) (EmployeeResponse, error) {
	employee, err := s.repository.FindEmployee(ctx, id)
	if err != nil {
		return EmployeeResponse{}, employeeReadErrors.Resolve(err)
	}
	return newEmployeeResponse(employee, s.today()), nil
}

func (s *service) CreateEmployee(ctx context.Context, request EmployeeRequest) (EmployeeResponse, error) {
	if !employmentRangeValid(request.JoinedDate, request.LeftDate) {
		return EmployeeResponse{}, errInvalidEmploymentRange
	}

	var employee Employee
	applyEmployeeRequest(&employee, request)
	if err := s.repository.CreateEmployee(ctx, &employee); err != nil {
		return EmployeeResponse{}, employeeWriteErrors.Resolve(err)
	}
	return newEmployeeResponse(employee, s.today()), nil
}

func (s *service) UpdateEmployee(ctx context.Context, id uuid.UUID, request EmployeeRequest) (EmployeeResponse, error) {
	if !employmentRangeValid(request.JoinedDate, request.LeftDate) {
		return EmployeeResponse{}, errInvalidEmploymentRange
	}

	employee, err := s.repository.FindEmployee(ctx, id)
	if err != nil {
		return EmployeeResponse{}, employeeReadErrors.Resolve(err)
	}

	applyEmployeeRequest(&employee, request)
	if err := s.repository.SaveEmployee(ctx, &employee); err != nil {
		return EmployeeResponse{}, employeeWriteErrors.Resolve(err)
	}
	return newEmployeeResponse(employee, s.today()), nil
}

func (s *service) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	hasRecords, err := s.repository.EmployeeHasRecords(ctx, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if hasRecords {
		return errEmployeeInUse
	}
	return employeeReadErrors.Resolve(s.repository.DeleteEmployee(ctx, id))
}

func (s *service) ListAttendances(ctx context.Context, query AttendanceQuery) (pagination.Page[AttendanceResponse], error) {
	rows, total, err := s.repository.ListAttendances(ctx, AttendanceFilter{
		EmployeeID: query.EmployeeID,
		ProjectID:  query.ProjectID,
		Status:     query.Status,
		DateFrom:   query.DateFrom,
		DateTo:     query.DateTo,
		Offset:     query.Offset(),
		Limit:      query.Size(),
	})
	if err != nil {
		return pagination.Page[AttendanceResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newAttendanceResponse), query.Query, total), nil
}

func (s *service) GetAttendance(ctx context.Context, id uuid.UUID) (AttendanceResponse, error) {
	row, err := s.repository.FindAttendanceRow(ctx, id)
	if err != nil {
		return AttendanceResponse{}, attendanceReadErrors.Resolve(err)
	}
	return newAttendanceResponse(row), nil
}

func (s *service) CreateAttendance(ctx context.Context, request AttendanceRequest) (AttendanceResponse, error) {
	request = normalizeAttendance(request)
	employee, err := s.attendingEmployee(ctx, request)
	if err != nil {
		return AttendanceResponse{}, err
	}

	var attendance Attendance
	applyAttendanceRequest(&attendance, request)
	if err := s.repository.CreateAttendance(ctx, &attendance); err != nil {
		return AttendanceResponse{}, attendanceWriteErrors.Resolve(err)
	}
	return newAttendanceResponse(attendanceRowOf(attendance, employee)), nil
}

func (s *service) UpdateAttendance(ctx context.Context, id uuid.UUID, request AttendanceRequest) (AttendanceResponse, error) {
	attendance, err := s.repository.FindAttendance(ctx, id)
	if err != nil {
		return AttendanceResponse{}, attendanceReadErrors.Resolve(err)
	}

	request = normalizeAttendance(request)
	employee, err := s.attendingEmployee(ctx, request)
	if err != nil {
		return AttendanceResponse{}, err
	}

	applyAttendanceRequest(&attendance, request)
	if err := s.repository.SaveAttendance(ctx, &attendance); err != nil {
		return AttendanceResponse{}, attendanceWriteErrors.Resolve(err)
	}
	return newAttendanceResponse(attendanceRowOf(attendance, employee)), nil
}

func (s *service) DeleteAttendance(ctx context.Context, id uuid.UUID) error {
	return attendanceReadErrors.Resolve(s.repository.DeleteAttendance(ctx, id))
}

func (s *service) ListPayrolls(ctx context.Context, query PayrollQuery) (pagination.Page[PayrollResponse], error) {
	rows, total, err := s.repository.ListPayrolls(ctx, PayrollFilter{
		EmployeeID: query.EmployeeID,
		Period:     query.Period,
		Paid:       query.Paid,
		Offset:     query.Offset(),
		Limit:      query.Size(),
	})
	if err != nil {
		return pagination.Page[PayrollResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newPayrollResponse), query.Query, total), nil
}

func (s *service) GetPayroll(ctx context.Context, id uuid.UUID) (PayrollResponse, error) {
	row, err := s.repository.FindPayrollRow(ctx, id)
	if err != nil {
		return PayrollResponse{}, payrollReadErrors.Resolve(err)
	}
	return newPayrollResponse(row), nil
}

func (s *service) CreatePayroll(ctx context.Context, request PayrollRequest) (PayrollResponse, error) {
	employee, err := s.repository.FindEmployee(ctx, request.EmployeeID)
	if err != nil {
		return PayrollResponse{}, employeeReferenceErrors.Resolve(err)
	}

	var payroll Payroll
	if err := s.calculatePayroll(ctx, employee, request, &payroll); err != nil {
		return PayrollResponse{}, err
	}
	if err := s.repository.CreatePayroll(ctx, &payroll); err != nil {
		return PayrollResponse{}, payrollWriteErrors.Resolve(err)
	}
	return newPayrollResponse(payrollRowOf(payroll, employee)), nil
}

func (s *service) UpdatePayroll(ctx context.Context, id uuid.UUID, request PayrollRequest) (PayrollResponse, error) {
	payroll, err := s.unpaidPayroll(ctx, id)
	if err != nil {
		return PayrollResponse{}, err
	}

	employee, err := s.repository.FindEmployee(ctx, request.EmployeeID)
	if err != nil {
		return PayrollResponse{}, employeeReferenceErrors.Resolve(err)
	}
	if err := s.calculatePayroll(ctx, employee, request, &payroll); err != nil {
		return PayrollResponse{}, err
	}
	if err := s.repository.SavePayroll(ctx, &payroll); err != nil {
		return PayrollResponse{}, payrollWriteErrors.Resolve(err)
	}
	return newPayrollResponse(payrollRowOf(payroll, employee)), nil
}

func (s *service) PayPayroll(ctx context.Context, id uuid.UUID) (PayrollResponse, error) {
	payroll, err := s.repository.FindPayroll(ctx, id)
	if err != nil {
		return PayrollResponse{}, payrollReadErrors.Resolve(err)
	}
	if payroll.Paid {
		return s.GetPayroll(ctx, id)
	}

	payroll.Paid = true
	if err := s.repository.SavePayroll(ctx, &payroll); err != nil {
		return PayrollResponse{}, payrollWriteErrors.Resolve(err)
	}
	return s.GetPayroll(ctx, id)
}

func (s *service) DeletePayroll(ctx context.Context, id uuid.UUID) error {
	if _, err := s.unpaidPayroll(ctx, id); err != nil {
		return err
	}
	return payrollReadErrors.Resolve(s.repository.DeletePayroll(ctx, id))
}

func (s *service) attendingEmployee(ctx context.Context, request AttendanceRequest) (Employee, error) {
	if err := validateAttendanceTimes(request); err != nil {
		return Employee{}, err
	}

	employee, err := s.repository.FindEmployee(ctx, request.EmployeeID)
	if err != nil {
		return Employee{}, employeeReferenceErrors.Resolve(err)
	}
	if !employedOn(employee, request.Date) {
		return Employee{}, errAttendanceOutsideEmployment
	}
	return employee, nil
}

func (s *service) unpaidPayroll(ctx context.Context, id uuid.UUID) (Payroll, error) {
	payroll, err := s.repository.FindPayroll(ctx, id)
	if err != nil {
		return Payroll{}, payrollReadErrors.Resolve(err)
	}
	if payroll.Paid {
		return Payroll{}, errPayrollPaid
	}
	return payroll, nil
}

func (s *service) calculatePayroll(ctx context.Context, employee Employee, request PayrollRequest, payroll *Payroll) error {
	start, err := time.Parse(periodLayout, request.Period)
	if err != nil {
		return errPayrollPeriod
	}
	end := start.AddDate(0, 1, 0)
	if !employedDuring(employee, start, end) {
		return errPayrollOutsideEmployment
	}

	basicPay, err := s.basicPay(ctx, employee, start, end)
	if err != nil {
		return err
	}
	netPay := basicPay + request.Allowance - request.Deduction
	if netPay < 0 {
		return errPayrollNetNegative
	}

	payroll.EmployeeID = employee.ID
	payroll.Period = request.Period
	payroll.BasicPay = basicPay
	payroll.Allowance = request.Allowance
	payroll.Deduction = request.Deduction
	payroll.NetPay = netPay
	return nil
}

func (s *service) basicPay(ctx context.Context, employee Employee, start, end time.Time) (int64, error) {
	if employee.EmploymentType != employmentDaily {
		return employee.BaseSalary, nil
	}

	presentDays, err := s.repository.CountPresentDays(ctx, employee.ID, start, end)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return employee.BaseSalary * presentDays, nil
}

func (s *service) today() time.Time {
	year, month, day := s.now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func employmentRangeValid(joined time.Time, left *time.Time) bool {
	return left == nil || !left.Before(joined)
}

func employedOn(employee Employee, date time.Time) bool {
	if date.Before(employee.JoinedDate) {
		return false
	}
	return employee.LeftDate == nil || !date.After(*employee.LeftDate)
}

func employedDuring(employee Employee, start, end time.Time) bool {
	if !employee.JoinedDate.Before(end) {
		return false
	}
	return employee.LeftDate == nil || !employee.LeftDate.Before(start)
}

func validateAttendanceTimes(request AttendanceRequest) error {
	hasTimes := request.CheckInTime != nil || request.CheckOutTime != nil
	switch {
	case request.Status != attendancePresent && hasTimes:
		return errAttendanceTimeNotAllowed
	case request.Status != attendancePresent:
		return nil
	case request.CheckInTime == nil:
		return errAttendanceCheckInRequired
	case request.CheckOutTime != nil && *request.CheckOutTime < *request.CheckInTime:
		return errAttendanceTimeOrder
	default:
		return nil
	}
}

func normalizeAttendance(request AttendanceRequest) AttendanceRequest {
	request.CheckInTime = blankToNil(request.CheckInTime)
	request.CheckOutTime = blankToNil(request.CheckOutTime)
	return request
}

func blankToNil(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func clockText(stored *string) *string {
	if stored == nil {
		return nil
	}
	parsed, err := time.Parse(storedClockLayout, *stored)
	if err != nil {
		return stored
	}
	formatted := parsed.Format(clockLayout)
	return &formatted
}

func attendanceRowOf(attendance Attendance, employee Employee) AttendanceRow {
	return AttendanceRow{
		ID:           attendance.ID,
		EmployeeID:   attendance.EmployeeID,
		EmployeeName: employee.Name,
		Date:         attendance.Date,
		CheckInTime:  attendance.CheckInTime,
		CheckOutTime: attendance.CheckOutTime,
		Status:       attendance.Status,
		CreatedAt:    attendance.CreatedAt,
		UpdatedAt:    attendance.UpdatedAt,
	}
}

func payrollRowOf(payroll Payroll, employee Employee) PayrollRow {
	return PayrollRow{
		ID:           payroll.ID,
		EmployeeID:   payroll.EmployeeID,
		EmployeeName: employee.Name,
		Period:       payroll.Period,
		BasicPay:     payroll.BasicPay,
		Allowance:    payroll.Allowance,
		Deduction:    payroll.Deduction,
		NetPay:       payroll.NetPay,
		Paid:         payroll.Paid,
		CreatedAt:    payroll.CreatedAt,
		UpdatedAt:    payroll.UpdatedAt,
	}
}

func applyEmployeeRequest(employee *Employee, request EmployeeRequest) {
	employee.IdentityNumber = strings.TrimSpace(request.IdentityNumber)
	employee.Name = strings.TrimSpace(request.Name)
	employee.Position = strings.TrimSpace(request.Position)
	employee.ProjectID = request.ProjectID
	employee.UserID = request.UserID
	employee.EmploymentType = request.EmploymentType
	employee.JoinedDate = request.JoinedDate
	employee.LeftDate = request.LeftDate
	employee.BaseSalary = request.BaseSalary
}

func applyAttendanceRequest(attendance *Attendance, request AttendanceRequest) {
	attendance.EmployeeID = request.EmployeeID
	attendance.Date = request.Date
	attendance.CheckInTime = request.CheckInTime
	attendance.CheckOutTime = request.CheckOutTime
	attendance.Status = request.Status
}
