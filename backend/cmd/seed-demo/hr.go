package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/hr"
)

const (
	employmentPermanent = "permanent"
	employmentContract  = "contract"
	employmentDaily     = "daily"

	attendanceDays     = 14
	attendanceCycle    = 23
	checkInSpread      = 45
	checkOutSpread     = 50
	socialSecurityRate = 50
)

type employeeSeed struct {
	identity   string
	name       string
	position   string
	project    string
	employment string
	joined     int
	salary     int64
	allowance  int64
}

func demoEmployees() []employeeSeed {
	return []employeeSeed{
		{"DEMO-NIK-001", "Agus Setiawan", "Manajer Proyek", projectGriyaAsri, employmentPermanent, -38, 18_500_000, 3_000_000},
		{"DEMO-NIK-002", "Rizky Ramadhan", "Site Engineer", projectGriyaAsri, employmentPermanent, -26, 11_000_000, 1_500_000},
		{"DEMO-NIK-003", "Bambang Sutrisno", "Pengawas Lapangan", projectHospital, employmentPermanent, -30, 10_500_000, 1_500_000},
		{"DEMO-NIK-004", "Intan Permata", "Staf Keuangan", "", employmentPermanent, -44, 9_000_000, 1_000_000},
		{"DEMO-NIK-005", "Yoga Pratama", "Staf Pengadaan", "", employmentPermanent, -20, 8_500_000, 1_000_000},
		{"DEMO-NIK-006", "Melati Anggraini", "Staf Pemasaran", projectGriyaAsri, employmentPermanent, -15, 7_500_000, 1_250_000},
		{"DEMO-NIK-007", "Faisal Akbar", "Drafter", projectRuko, employmentContract, -6, 7_000_000, 750_000},
		{"DEMO-NIK-008", "Nurul Hidayah", "Quantity Surveyor", projectHospital, employmentContract, -9, 9_500_000, 750_000},
		{"DEMO-NIK-009", "Eko Saputro", "Admin Gudang", projectGriyaAsri, employmentContract, -10, 5_200_000, 500_000},
		{"DEMO-NIK-010", "Slamet Riyadi", "Mandor", projectGriyaAsri, employmentDaily, -11, 225_000, 0},
		{"DEMO-NIK-011", "Joko Susilo", "Tukang Besi", projectHospital, employmentDaily, -8, 175_000, 0},
		{"DEMO-NIK-012", "Darto Wibowo", "Tukang Batu", projectGriyaAsri, employmentDaily, -10, 165_000, 0},
	}
}

func (s *seeder) seedHR() error {
	for index, seed := range demoEmployees() {
		id, err := s.seedEmployee(seed)
		if err != nil {
			return err
		}
		if err := s.seedAttendance(id, index, seed); err != nil {
			return err
		}
		if err := s.seedPayrolls(id, seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedEmployee(seed employeeSeed) (uuid.UUID, error) {
	projectID, err := s.knownPointer("project", seed.project)
	if err != nil {
		return uuid.Nil, err
	}
	id, _, err := s.ensure("employee", map[string]any{"identity_number": seed.identity}, func() (uuid.UUID, error) {
		response, err := s.services.hr.CreateEmployee(s.ctx, hr.EmployeeRequest{
			IdentityNumber: seed.identity, Name: seed.name, Position: seed.position, ProjectID: projectID,
			EmploymentType: seed.employment, JoinedDate: *s.monthOffset(seed.joined), BaseSalary: seed.salary,
		})
		return response.ID, err
	})
	return id, err
}

func (s *seeder) attendanceStart(seed employeeSeed) time.Time {
	start := s.day(1 - attendanceDays)
	if seed.employment != employmentDaily {
		return start
	}
	previousMonth := s.monthDay(1, 1)
	if previousMonth.Before(start) {
		return previousMonth
	}
	return start
}

func (s *seeder) seedAttendance(employeeID uuid.UUID, index int, seed employeeSeed) error {
	for date := s.attendanceStart(seed); !date.After(s.today); date = date.AddDate(0, 0, 1) {
		if !worksOn(date, seed.employment) {
			continue
		}
		request := attendanceFor(employeeID, index, date)
		keys := map[string]any{"employee_id": employeeID, "date": date.Format(time.DateOnly)}
		_, _, err := s.ensure("attendance", keys, func() (uuid.UUID, error) {
			response, err := s.services.hr.CreateAttendance(s.ctx, request)
			return response.ID, err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func worksOn(date time.Time, employment string) bool {
	switch date.Weekday() {
	case time.Sunday:
		return false
	case time.Saturday:
		return employment == employmentDaily
	default:
		return true
	}
}

func attendanceFor(employeeID uuid.UUID, index int, date time.Time) hr.AttendanceRequest {
	request := hr.AttendanceRequest{EmployeeID: employeeID, Date: date, Status: "present"}
	switch (index*7 + date.YearDay()*3) % attendanceCycle {
	case 0:
		request.Status = "sick"
		return request
	case 1:
		request.Status = "permitted"
		return request
	case 2:
		request.Status = "absent"
		return request
	}
	checkIn := fmt.Sprintf("07:%02d", (index*11+date.YearDay()*7)%checkInSpread+10)
	checkOut := fmt.Sprintf("16:%02d", (index*13+date.YearDay()*5)%checkOutSpread)
	request.CheckInTime = &checkIn
	request.CheckOutTime = &checkOut
	return request
}

func (s *seeder) seedPayrolls(employeeID uuid.UUID, seed employeeSeed) error {
	for monthsAgo := 1; monthsAgo >= 0; monthsAgo-- {
		period := s.period(monthsAgo)
		id, created, err := s.ensure("payroll", map[string]any{"employee_id": employeeID, "period": period}, func() (uuid.UUID, error) {
			response, err := s.services.hr.CreatePayroll(s.ctx, hr.PayrollRequest{
				EmployeeID: employeeID, Period: period, Allowance: seed.allowance, Deduction: deductionOf(seed),
			})
			return response.ID, err
		})
		if err != nil {
			return err
		}
		if created && monthsAgo == 1 {
			if _, err := s.services.hr.PayPayroll(s.ctx, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func deductionOf(seed employeeSeed) int64 {
	if seed.employment == employmentDaily {
		return 0
	}
	return seed.salary / socialSecurityRate
}
