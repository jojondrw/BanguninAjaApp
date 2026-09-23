package sales

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type CustomerRequest struct {
	Name           string `json:"name" binding:"required,max=160"`
	Contact        string `json:"contact" binding:"max=60"`
	Email          string `json:"email" binding:"omitempty,email,max=160"`
	IdentityNumber string `json:"identityNumber" binding:"required,max=20"`
	Address        string `json:"address" binding:"max=2000"`
}

type CustomerQuery struct {
	pagination.Query
	Search string `form:"search" binding:"omitempty,max=160"`
}

type CustomerResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Contact        string    `json:"contact"`
	Email          string    `json:"email"`
	IdentityNumber string    `json:"identityNumber"`
	Address        string    `json:"address"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type UnitRequest struct {
	Code      string    `json:"code" binding:"required,max=20"`
	ProjectID uuid.UUID `json:"projectId" binding:"required"`
	UnitType  string    `json:"unitType" binding:"required,max=60"`
	AreaSqm   float64   `json:"areaSqm" binding:"required,gt=0"`
	Price     int64     `json:"price" binding:"min=0"`
	Status    string    `json:"status" binding:"required,oneof=available on_hold"`
}

type UnitQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=60"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Status    string     `form:"status" binding:"omitempty,oneof=available reserved sold on_hold"`
}

type UnitSummaryQuery struct {
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
}

type UnitResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	ProjectID uuid.UUID `json:"projectId"`
	UnitType  string    `json:"unitType"`
	AreaSqm   float64   `json:"areaSqm"`
	Price     int64     `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UnitStatusCountResponse struct {
	Status string `json:"status"`
	Total  int64  `json:"total"`
}

type LeadRequest struct {
	Name            string     `json:"name" binding:"required,max=160"`
	Contact         string     `json:"contact" binding:"max=60"`
	ProjectID       *uuid.UUID `json:"projectId"`
	Source          string     `json:"source" binding:"max=60"`
	Stage           string     `json:"stage" binding:"required,oneof=new interested negotiating won cancelled"`
	LastContactedAt *time.Time `json:"lastContactedAt"`
}

type LeadQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=160"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Stage     string     `form:"stage" binding:"omitempty,oneof=new interested negotiating won cancelled"`
}

type LeadResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Contact         string     `json:"contact"`
	ProjectID       *uuid.UUID `json:"projectId"`
	Source          string     `json:"source"`
	Stage           string     `json:"stage"`
	LastContactedAt *time.Time `json:"lastContactedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type ContractRequest struct {
	Number     string    `json:"number" binding:"required,max=40"`
	CustomerID uuid.UUID `json:"customerId" binding:"required"`
	UnitID     uuid.UUID `json:"unitId" binding:"required"`
	Type       string    `json:"type" binding:"required,oneof=cash mortgage installment"`
	Value      int64     `json:"value" binding:"required,gt=0"`
	Date       time.Time `json:"date" binding:"required"`
}

type ContractStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active paid cancelled"`
}

type ContractQuery struct {
	pagination.Query
	Search     string     `form:"search" binding:"omitempty,max=40"`
	CustomerID *uuid.UUID `form:"customerId,parser=encoding.TextUnmarshaler"`
	UnitID     *uuid.UUID `form:"unitId,parser=encoding.TextUnmarshaler"`
	Status     string     `form:"status" binding:"omitempty,oneof=draft active paid cancelled"`
	Type       string     `form:"type" binding:"omitempty,oneof=cash mortgage installment"`
}

type ContractResponse struct {
	ID           uuid.UUID `json:"id"`
	Number       string    `json:"number"`
	CustomerID   uuid.UUID `json:"customerId"`
	CustomerName string    `json:"customerName"`
	UnitID       uuid.UUID `json:"unitId"`
	UnitCode     string    `json:"unitCode"`
	Type         string    `json:"type"`
	Value        int64     `json:"value"`
	Date         time.Time `json:"date"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type InstallmentRequest struct {
	InstallmentNumber int       `json:"installmentNumber" binding:"required,min=1"`
	DueDate           time.Time `json:"dueDate" binding:"required"`
	Amount            int64     `json:"amount" binding:"required,gt=0"`
}

type InstallmentPaymentRequest struct {
	PaidDate time.Time `json:"paidDate" binding:"required"`
}

type InstallmentQuery struct {
	pagination.Query
	ContractID *uuid.UUID `form:"contractId,parser=encoding.TextUnmarshaler"`
	CustomerID *uuid.UUID `form:"customerId,parser=encoding.TextUnmarshaler"`
	Status     string     `form:"status" binding:"omitempty,oneof=not_due due paid overdue"`
}

type InstallmentResponse struct {
	ID                uuid.UUID  `json:"id"`
	ContractID        uuid.UUID  `json:"contractId"`
	ContractNumber    string     `json:"contractNumber"`
	CustomerID        uuid.UUID  `json:"customerId"`
	CustomerName      string     `json:"customerName"`
	InstallmentNumber int        `json:"installmentNumber"`
	DueDate           time.Time  `json:"dueDate"`
	Amount            int64      `json:"amount"`
	PaidDate          *time.Time `json:"paidDate"`
	Status            string     `json:"status"`
	DaysOverdue       int        `json:"daysOverdue"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func newCustomerResponse(customer Customer) CustomerResponse {
	return CustomerResponse{
		ID:             customer.ID,
		Name:           customer.Name,
		Contact:        customer.Contact,
		Email:          customer.Email,
		IdentityNumber: customer.IdentityNumber,
		Address:        customer.Address,
		CreatedAt:      customer.CreatedAt,
		UpdatedAt:      customer.UpdatedAt,
	}
}

func newUnitResponse(unit Unit) UnitResponse {
	return UnitResponse{
		ID:        unit.ID,
		Code:      unit.Code,
		ProjectID: unit.ProjectID,
		UnitType:  unit.UnitType,
		AreaSqm:   unit.AreaSqm,
		Price:     unit.Price,
		Status:    unit.Status,
		CreatedAt: unit.CreatedAt,
		UpdatedAt: unit.UpdatedAt,
	}
}

func newUnitStatusCountResponse(count UnitStatusCount) UnitStatusCountResponse {
	return UnitStatusCountResponse{Status: count.Status, Total: count.Total}
}

func newLeadResponse(lead Lead) LeadResponse {
	return LeadResponse{
		ID:              lead.ID,
		Name:            lead.Name,
		Contact:         lead.Contact,
		ProjectID:       lead.ProjectID,
		Source:          lead.Source,
		Stage:           lead.Stage,
		LastContactedAt: lead.LastContactedAt,
		CreatedAt:       lead.CreatedAt,
		UpdatedAt:       lead.UpdatedAt,
	}
}

func newContractResponse(row ContractRow) ContractResponse {
	return ContractResponse{
		ID:           row.ID,
		Number:       row.Number,
		CustomerID:   row.CustomerID,
		CustomerName: row.CustomerName,
		UnitID:       row.UnitID,
		UnitCode:     row.UnitCode,
		Type:         row.Type,
		Value:        row.Value,
		Date:         row.Date,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func newInstallmentResponse(row InstallmentRow, today time.Time) InstallmentResponse {
	settled := row.Status == duedate.Paid
	return InstallmentResponse{
		ID:                row.ID,
		ContractID:        row.ContractID,
		ContractNumber:    row.ContractNumber,
		CustomerID:        row.CustomerID,
		CustomerName:      row.CustomerName,
		InstallmentNumber: row.InstallmentNumber,
		DueDate:           row.DueDate,
		Amount:            row.Amount,
		PaidDate:          row.PaidDate,
		Status:            duedate.Status(row.DueDate, settled, today),
		DaysOverdue:       duedate.DaysOverdue(row.DueDate, settled, today),
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
