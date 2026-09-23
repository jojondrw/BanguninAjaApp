package hr

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	employee    Employee
	presentDays int64
	payroll     Payroll
	saved       *Payroll
}

func (f *fakeRepository) FindEmployee(context.Context, uuid.UUID) (Employee, error) {
	return f.employee, nil
}

func (f *fakeRepository) CountPresentDays(context.Context, uuid.UUID, time.Time, time.Time) (int64, error) {
	return f.presentDays, nil
}

func (f *fakeRepository) CreatePayroll(_ context.Context, payroll *Payroll) error {
	f.saved = payroll
	return nil
}

func (f *fakeRepository) FindPayroll(context.Context, uuid.UUID) (Payroll, error) {
	return f.payroll, nil
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func clock(value string) *string {
	return &value
}

func TestDailyWorkerIsPaidPerPresentDay(t *testing.T) {
	repository := &fakeRepository{
		employee:    Employee{EmploymentType: employmentDaily, BaseSalary: 150000, JoinedDate: date(2026, 1, 1)},
		presentDays: 22,
	}

	payroll, err := NewService(repository).CreatePayroll(context.Background(), PayrollRequest{
		Period: "2026-08", Allowance: 100000, Deduction: 50000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payroll.BasicPay != 3300000 || payroll.NetPay != 3350000 {
		t.Fatalf("got basic %d net %d", payroll.BasicPay, payroll.NetPay)
	}
}

func TestPermanentEmployeeGetsBaseSalary(t *testing.T) {
	repository := &fakeRepository{
		employee:    Employee{EmploymentType: "permanent", BaseSalary: 9000000, JoinedDate: date(2025, 3, 1)},
		presentDays: 3,
	}

	payroll, err := NewService(repository).CreatePayroll(context.Background(), PayrollRequest{Period: "2026-08"})
	if err != nil || payroll.BasicPay != 9000000 {
		t.Fatalf("got basic %d err %v", payroll.BasicPay, err)
	}
}

func TestPayrollRejectsNegativeNetPay(t *testing.T) {
	repository := &fakeRepository{employee: Employee{BaseSalary: 1000, JoinedDate: date(2025, 1, 1)}}

	_, err := NewService(repository).CreatePayroll(context.Background(), PayrollRequest{Period: "2026-08", Deduction: 5000})
	if !errors.Is(err, errPayrollNetNegative) {
		t.Fatalf("got %v, want errPayrollNetNegative", err)
	}
}

func TestPayrollOutsideEmploymentIsRejected(t *testing.T) {
	left := date(2026, 6, 30)
	repository := &fakeRepository{employee: Employee{BaseSalary: 1000, JoinedDate: date(2025, 1, 1), LeftDate: &left}}

	_, err := NewService(repository).CreatePayroll(context.Background(), PayrollRequest{Period: "2026-08"})
	if !errors.Is(err, errPayrollOutsideEmployment) {
		t.Fatalf("got %v, want errPayrollOutsideEmployment", err)
	}
}

func TestPaidPayrollIsLocked(t *testing.T) {
	repository := &fakeRepository{payroll: Payroll{Paid: true}}

	err := NewService(repository).DeletePayroll(context.Background(), uuid.New())
	if !errors.Is(err, errPayrollPaid) {
		t.Fatalf("got %v, want errPayrollPaid", err)
	}
}

func TestValidateAttendanceTimes(t *testing.T) {
	cases := []struct {
		name    string
		request AttendanceRequest
		want    error
	}{
		{"present with times", AttendanceRequest{Status: attendancePresent, CheckInTime: clock("07:30"), CheckOutTime: clock("16:00")}, nil},
		{"present without check in", AttendanceRequest{Status: attendancePresent}, errAttendanceCheckInRequired},
		{"check out before check in", AttendanceRequest{Status: attendancePresent, CheckInTime: clock("08:00"), CheckOutTime: clock("07:00")}, errAttendanceTimeOrder},
		{"sick with times", AttendanceRequest{Status: "sick", CheckInTime: clock("08:00")}, errAttendanceTimeNotAllowed},
		{"absent without times", AttendanceRequest{Status: "absent"}, nil},
	}
	for _, tc := range cases {
		if got := validateAttendanceTimes(tc.request); !errors.Is(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestEmployedOn(t *testing.T) {
	left := date(2026, 6, 30)
	employee := Employee{JoinedDate: date(2026, 1, 1), LeftDate: &left}
	if employedOn(employee, date(2025, 12, 31)) || employedOn(employee, date(2026, 7, 1)) {
		t.Fatal("dates outside employment must be rejected")
	}
	if !employedOn(employee, date(2026, 1, 1)) || !employedOn(employee, left) {
		t.Fatal("first and last day must count")
	}
}

func TestClockTextDropsSeconds(t *testing.T) {
	if got := clockText(clock("07:30:00")); got == nil || *got != "07:30" {
		t.Fatalf("got %v", got)
	}
	if clockText(nil) != nil {
		t.Fatal("nil must stay nil")
	}
}
