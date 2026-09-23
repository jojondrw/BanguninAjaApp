package hr

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type EmployeeRequest struct {
	IdentityNumber string     `json:"identityNumber" binding:"required,max=20"`
	Name           string     `json:"name" binding:"required,max=160"`
	Position       string     `json:"position" binding:"required,max=80"`
	ProjectID      *uuid.UUID `json:"projectId"`
	UserID         *uuid.UUID `json:"userId"`
	EmploymentType string     `json:"employmentType" binding:"required,oneof=permanent contract daily subcontractor"`
	JoinedDate     time.Time  `json:"joinedDate" binding:"required"`
	LeftDate       *time.Time `json:"leftDate"`
	BaseSalary     int64      `json:"baseSalary" binding:"min=0"`
}

type EmployeeQuery struct {
	pagination.Query
	Search         string     `form:"search" binding:"omitempty,max=160"`
	EmploymentType string     `form:"employmentType" binding:"omitempty,oneof=permanent contract daily subcontractor"`
	ProjectID      *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Active         *bool      `form:"active"`
}

type EmployeeResponse struct {
	ID             uuid.UUID  `json:"id"`
	IdentityNumber string     `json:"identityNumber"`
	Name           string     `json:"name"`
	Position       string     `json:"position"`
	ProjectID      *uuid.UUID `json:"projectId"`
	UserID         *uuid.UUID `json:"userId"`
	EmploymentType string     `json:"employmentType"`
	JoinedDate     time.Time  `json:"joinedDate"`
	LeftDate       *time.Time `json:"leftDate"`
	BaseSalary     int64      `json:"baseSalary"`
	Active         bool       `json:"active"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type EmploymentCountResponse struct {
	EmploymentType string `json:"employmentType"`
	Total          int64  `json:"total"`
}

type AttendanceRequest struct {
	EmployeeID   uuid.UUID `json:"employeeId" binding:"required"`
	Date         time.Time `json:"date" binding:"required"`
	CheckInTime  *string   `json:"checkInTime" binding:"omitempty,datetime=15:04"`
	CheckOutTime *string   `json:"checkOutTime" binding:"omitempty,datetime=15:04"`
	Status       string    `json:"status" binding:"required,oneof=present permitted sick absent holiday"`
}

type AttendanceQuery struct {
	pagination.Query
	EmployeeID *uuid.UUID `form:"employeeId,parser=encoding.TextUnmarshaler"`
	ProjectID  *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Status     string     `form:"status" binding:"omitempty,oneof=present permitted sick absent holiday"`
	DateFrom   *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo     *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type AttendanceResponse struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   uuid.UUID `json:"employeeId"`
	EmployeeName string    `json:"employeeName"`
	Date         time.Time `json:"date"`
	CheckInTime  *string   `json:"checkInTime"`
	CheckOutTime *string   `json:"checkOutTime"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type PayrollRequest struct {
	EmployeeID uuid.UUID `json:"employeeId" binding:"required"`
	Period     string    `json:"period" binding:"required,datetime=2006-01"`
	Allowance  int64     `json:"allowance" binding:"min=0"`
	Deduction  int64     `json:"deduction" binding:"min=0"`
}

type PayrollQuery struct {
	pagination.Query
	EmployeeID *uuid.UUID `form:"employeeId,parser=encoding.TextUnmarshaler"`
	Period     string     `form:"period" binding:"omitempty,datetime=2006-01"`
	Paid       *bool      `form:"paid"`
}

type PayrollResponse struct {
	ID           uuid.UUID `json:"id"`
	EmployeeID   uuid.UUID `json:"employeeId"`
	EmployeeName string    `json:"employeeName"`
	Period       string    `json:"period"`
	BasicPay     int64     `json:"basicPay"`
	Allowance    int64     `json:"allowance"`
	Deduction    int64     `json:"deduction"`
	NetPay       int64     `json:"netPay"`
	Paid         bool      `json:"paid"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func newEmployeeResponse(employee Employee, today time.Time) EmployeeResponse {
	return EmployeeResponse{
		ID:             employee.ID,
		IdentityNumber: employee.IdentityNumber,
		Name:           employee.Name,
		Position:       employee.Position,
		ProjectID:      employee.ProjectID,
		UserID:         employee.UserID,
		EmploymentType: employee.EmploymentType,
		JoinedDate:     employee.JoinedDate,
		LeftDate:       employee.LeftDate,
		BaseSalary:     employee.BaseSalary,
		Active:         employedOn(employee, today),
		CreatedAt:      employee.CreatedAt,
		UpdatedAt:      employee.UpdatedAt,
	}
}

func newEmploymentCountResponse(count EmploymentCount) EmploymentCountResponse {
	return EmploymentCountResponse{EmploymentType: count.EmploymentType, Total: count.Total}
}

func newAttendanceResponse(row AttendanceRow) AttendanceResponse {
	return AttendanceResponse{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		EmployeeName: row.EmployeeName,
		Date:         row.Date,
		CheckInTime:  clockText(row.CheckInTime),
		CheckOutTime: clockText(row.CheckOutTime),
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func newPayrollResponse(row PayrollRow) PayrollResponse {
	return PayrollResponse{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		EmployeeName: row.EmployeeName,
		Period:       row.Period,
		BasicPay:     row.BasicPay,
		Allowance:    row.Allowance,
		Deduction:    row.Deduction,
		NetPay:       row.NetPay,
		Paid:         row.Paid,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
