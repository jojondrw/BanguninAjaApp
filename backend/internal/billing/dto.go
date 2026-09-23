package billing

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type InvoiceRequest struct {
	Number    string     `json:"number" binding:"required,max=40"`
	Note      string     `json:"note" binding:"max=200"`
	PartyType string     `json:"partyType" binding:"required,oneof=customer vendor"`
	PartyID   uuid.UUID  `json:"partyId" binding:"required"`
	ProjectID *uuid.UUID `json:"projectId"`
	DueDate   time.Time  `json:"dueDate" binding:"required"`
	Amount    int64      `json:"amount" binding:"required,gt=0"`
}

type InvoiceQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=200"`
	Status    string     `form:"status" binding:"omitempty,oneof=not_due due paid overdue"`
	PartyType string     `form:"partyType" binding:"omitempty,oneof=customer vendor"`
	PartyID   *uuid.UUID `form:"partyId,parser=encoding.TextUnmarshaler"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
}

type InvoiceResponse struct {
	ID          uuid.UUID  `json:"id"`
	Number      string     `json:"number"`
	Note        string     `json:"note"`
	PartyType   string     `json:"partyType"`
	PartyID     uuid.UUID  `json:"partyId"`
	ProjectID   *uuid.UUID `json:"projectId"`
	DueDate     time.Time  `json:"dueDate"`
	Amount      int64      `json:"amount"`
	PaidAmount  int64      `json:"paidAmount"`
	Outstanding int64      `json:"outstanding"`
	Status      string     `json:"status"`
	DaysOverdue int        `json:"daysOverdue"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type ReceivableRequest struct {
	CustomerID uuid.UUID `json:"customerId" binding:"required"`
	Reference  string    `json:"reference" binding:"required,max=60"`
	DueDate    time.Time `json:"dueDate" binding:"required"`
	Amount     int64     `json:"amount" binding:"required,gt=0"`
}

type ReceivableQuery struct {
	pagination.Query
	Search     string     `form:"search" binding:"omitempty,max=60"`
	Status     string     `form:"status" binding:"omitempty,oneof=not_due due paid overdue"`
	CustomerID *uuid.UUID `form:"customerId,parser=encoding.TextUnmarshaler"`
}

type ReceivableResponse struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customerId"`
	Reference   string    `json:"reference"`
	DueDate     time.Time `json:"dueDate"`
	Amount      int64     `json:"amount"`
	PaidAmount  int64     `json:"paidAmount"`
	Outstanding int64     `json:"outstanding"`
	Status      string    `json:"status"`
	DaysOverdue int       `json:"daysOverdue"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PayableRequest struct {
	VendorID  uuid.UUID `json:"vendorId" binding:"required"`
	Reference string    `json:"reference" binding:"required,max=60"`
	DueDate   time.Time `json:"dueDate" binding:"required"`
	Amount    int64     `json:"amount" binding:"required,gt=0"`
}

type PayableQuery struct {
	pagination.Query
	Search   string     `form:"search" binding:"omitempty,max=60"`
	Status   string     `form:"status" binding:"omitempty,oneof=not_due due paid overdue"`
	VendorID *uuid.UUID `form:"vendorId,parser=encoding.TextUnmarshaler"`
}

type PayableResponse struct {
	ID          uuid.UUID `json:"id"`
	VendorID    uuid.UUID `json:"vendorId"`
	Reference   string    `json:"reference"`
	DueDate     time.Time `json:"dueDate"`
	Amount      int64     `json:"amount"`
	PaidAmount  int64     `json:"paidAmount"`
	Outstanding int64     `json:"outstanding"`
	Status      string    `json:"status"`
	DaysOverdue int       `json:"daysOverdue"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PaymentRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

func newInvoiceResponse(invoice Invoice, today time.Time) InvoiceResponse {
	settled := settledAmount(invoice.Amount, invoice.PaidAmount)
	return InvoiceResponse{
		ID:          invoice.ID,
		Number:      invoice.Number,
		Note:        invoice.Note,
		PartyType:   invoice.PartyType,
		PartyID:     invoice.PartyID,
		ProjectID:   invoice.ProjectID,
		DueDate:     invoice.DueDate,
		Amount:      invoice.Amount,
		PaidAmount:  invoice.PaidAmount,
		Outstanding: invoice.Amount - invoice.PaidAmount,
		Status:      duedate.Status(invoice.DueDate, settled, today),
		DaysOverdue: duedate.DaysOverdue(invoice.DueDate, settled, today),
		CreatedAt:   invoice.CreatedAt,
		UpdatedAt:   invoice.UpdatedAt,
	}
}

func newReceivableResponse(receivable Receivable, today time.Time) ReceivableResponse {
	settled := settledAmount(receivable.Amount, receivable.PaidAmount)
	return ReceivableResponse{
		ID:          receivable.ID,
		CustomerID:  receivable.CustomerID,
		Reference:   receivable.Reference,
		DueDate:     receivable.DueDate,
		Amount:      receivable.Amount,
		PaidAmount:  receivable.PaidAmount,
		Outstanding: receivable.Amount - receivable.PaidAmount,
		Status:      duedate.Status(receivable.DueDate, settled, today),
		DaysOverdue: duedate.DaysOverdue(receivable.DueDate, settled, today),
		CreatedAt:   receivable.CreatedAt,
		UpdatedAt:   receivable.UpdatedAt,
	}
}

func newPayableResponse(payable Payable, today time.Time) PayableResponse {
	settled := settledAmount(payable.Amount, payable.PaidAmount)
	return PayableResponse{
		ID:          payable.ID,
		VendorID:    payable.VendorID,
		Reference:   payable.Reference,
		DueDate:     payable.DueDate,
		Amount:      payable.Amount,
		PaidAmount:  payable.PaidAmount,
		Outstanding: payable.Amount - payable.PaidAmount,
		Status:      duedate.Status(payable.DueDate, settled, today),
		DaysOverdue: duedate.DaysOverdue(payable.DueDate, settled, today),
		CreatedAt:   payable.CreatedAt,
		UpdatedAt:   payable.UpdatedAt,
	}
}
