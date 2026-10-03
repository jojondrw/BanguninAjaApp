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
	Recorded  *bool      `form:"recorded"`
}

type InvoiceResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Number              string     `json:"number"`
	Note                string     `json:"note"`
	PartyType           string     `json:"partyType"`
	PartyID             uuid.UUID  `json:"partyId"`
	PartyName           *string    `json:"partyName"`
	ProjectID           *uuid.UUID `json:"projectId"`
	ProjectName         *string    `json:"projectName"`
	ReceivableID        *uuid.UUID `json:"receivableId"`
	ReceivableReference *string    `json:"receivableReference"`
	DueDate             time.Time  `json:"dueDate"`
	Amount              int64      `json:"amount"`
	PaidAmount          int64      `json:"paidAmount"`
	Outstanding         int64      `json:"outstanding"`
	Status              string     `json:"status"`
	DaysOverdue         int        `json:"daysOverdue"`
	LastPaidAt          *time.Time `json:"lastPaidAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type ReceivableRequest struct {
	CustomerID uuid.UUID  `json:"customerId" binding:"required"`
	ProjectID  *uuid.UUID `json:"projectId"`
	ContractID *uuid.UUID `json:"contractId"`
	InvoiceID  *uuid.UUID `json:"invoiceId"`
	Reference  string     `json:"reference" binding:"required,max=60"`
	DueDate    time.Time  `json:"dueDate" binding:"required"`
	Amount     int64      `json:"amount" binding:"required,gt=0"`
}

type ReceivableQuery struct {
	pagination.Query
	Search     string     `form:"search" binding:"omitempty,max=60"`
	Status     string     `form:"status" binding:"omitempty,oneof=not_due due paid overdue"`
	CustomerID *uuid.UUID `form:"customerId,parser=encoding.TextUnmarshaler"`
	ProjectID  *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	ContractID *uuid.UUID `form:"contractId,parser=encoding.TextUnmarshaler"`
}

type ReceivableResponse struct {
	ID             uuid.UUID  `json:"id"`
	CustomerID     uuid.UUID  `json:"customerId"`
	CustomerName   string     `json:"customerName"`
	ProjectID      *uuid.UUID `json:"projectId"`
	ProjectName    *string    `json:"projectName"`
	ContractID     *uuid.UUID `json:"contractId"`
	ContractNumber *string    `json:"contractNumber"`
	InvoiceID      *uuid.UUID `json:"invoiceId"`
	InvoiceNumber  *string    `json:"invoiceNumber"`
	Reference      string     `json:"reference"`
	DueDate        time.Time  `json:"dueDate"`
	Amount         int64      `json:"amount"`
	PaidAmount     int64      `json:"paidAmount"`
	Outstanding    int64      `json:"outstanding"`
	Status         string     `json:"status"`
	DaysOverdue    int        `json:"daysOverdue"`
	LastPaidAt     *time.Time `json:"lastPaidAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type PayableRequest struct {
	VendorID        uuid.UUID  `json:"vendorId" binding:"required"`
	ProjectID       *uuid.UUID `json:"projectId"`
	PurchaseOrderID *uuid.UUID `json:"purchaseOrderId"`
	Reference       string     `json:"reference" binding:"required,max=60"`
	DueDate         time.Time  `json:"dueDate" binding:"required"`
	Amount          int64      `json:"amount" binding:"required,gt=0"`
}

type PayableQuery struct {
	pagination.Query
	Search          string     `form:"search" binding:"omitempty,max=60"`
	Status          string     `form:"status" binding:"omitempty,oneof=not_due due paid overdue"`
	VendorID        *uuid.UUID `form:"vendorId,parser=encoding.TextUnmarshaler"`
	ProjectID       *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	PurchaseOrderID *uuid.UUID `form:"purchaseOrderId,parser=encoding.TextUnmarshaler"`
}

type PayableResponse struct {
	ID                  uuid.UUID  `json:"id"`
	VendorID            uuid.UUID  `json:"vendorId"`
	VendorName          string     `json:"vendorName"`
	ProjectID           *uuid.UUID `json:"projectId"`
	ProjectName         *string    `json:"projectName"`
	PurchaseOrderID     *uuid.UUID `json:"purchaseOrderId"`
	PurchaseOrderNumber *string    `json:"purchaseOrderNumber"`
	Reference           string     `json:"reference"`
	DueDate             time.Time  `json:"dueDate"`
	Amount              int64      `json:"amount"`
	PaidAmount          int64      `json:"paidAmount"`
	Outstanding         int64      `json:"outstanding"`
	Status              string     `json:"status"`
	DaysOverdue         int        `json:"daysOverdue"`
	LastPaidAt          *time.Time `json:"lastPaidAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type PaymentRequest struct {
	Amount    int64      `json:"amount" binding:"required,gt=0"`
	PaidAt    *time.Time `json:"paidAt"`
	Method    string     `json:"method" binding:"omitempty,oneof=transfer tunai cek lainnya"`
	Reference string     `json:"reference" binding:"max=60"`
	Note      string     `json:"note" binding:"max=200"`
}

type PaymentQuery struct {
	pagination.Query
}

type PaymentResponse struct {
	ID                  uuid.UUID  `json:"id"`
	InvoiceID           *uuid.UUID `json:"invoiceId"`
	InvoiceNumber       *string    `json:"invoiceNumber"`
	ReceivableID        *uuid.UUID `json:"receivableId"`
	ReceivableReference *string    `json:"receivableReference"`
	PayableID           *uuid.UUID `json:"payableId"`
	Amount              int64      `json:"amount"`
	PaidAt              time.Time  `json:"paidAt"`
	Method              string     `json:"method"`
	Reference           string     `json:"reference"`
	Note                string     `json:"note"`
	CreatedBy           *uuid.UUID `json:"createdBy"`
	CreatedByName       *string    `json:"createdByName"`
	CreatedAt           time.Time  `json:"createdAt"`
}

func newPaymentResponse(row PaymentRow) PaymentResponse {
	return PaymentResponse{
		ID:                  row.ID,
		InvoiceID:           row.InvoiceID,
		InvoiceNumber:       row.InvoiceNumber,
		ReceivableID:        row.ReceivableID,
		ReceivableReference: row.ReceivableReference,
		PayableID:           row.PayableID,
		Amount:              row.Amount,
		PaidAt:              row.PaidAt,
		Method:              row.Method,
		Reference:           row.Reference,
		Note:                row.Note,
		CreatedBy:           row.CreatedBy,
		CreatedByName:       row.CreatedByName,
		CreatedAt:           row.CreatedAt,
	}
}

func newInvoiceResponse(row InvoiceRow, today time.Time) InvoiceResponse {
	invoice := row.Invoice
	settled := settledAmount(invoice.Amount, invoice.PaidAmount)
	return InvoiceResponse{
		ID:                  invoice.ID,
		Number:              invoice.Number,
		Note:                invoice.Note,
		PartyType:           invoice.PartyType,
		PartyID:             invoice.PartyID,
		PartyName:           row.PartyName,
		ProjectID:           invoice.ProjectID,
		ProjectName:         row.ProjectName,
		ReceivableID:        row.ReceivableID,
		ReceivableReference: row.ReceivableReference,
		DueDate:             invoice.DueDate,
		Amount:              invoice.Amount,
		PaidAmount:          invoice.PaidAmount,
		Outstanding:         invoice.Amount - invoice.PaidAmount,
		Status:              duedate.Status(invoice.DueDate, settled, today),
		DaysOverdue:         duedate.DaysOverdue(invoice.DueDate, settled, today),
		LastPaidAt:          row.LastPaidAt,
		CreatedAt:           invoice.CreatedAt,
		UpdatedAt:           invoice.UpdatedAt,
	}
}

func newReceivableResponse(row ReceivableRow, today time.Time) ReceivableResponse {
	receivable := row.Receivable
	settled := settledAmount(receivable.Amount, receivable.PaidAmount)
	return ReceivableResponse{
		ID:             receivable.ID,
		CustomerID:     receivable.CustomerID,
		CustomerName:   row.CustomerName,
		ProjectID:      receivable.ProjectID,
		ProjectName:    row.ProjectName,
		ContractID:     receivable.ContractID,
		ContractNumber: row.ContractNumber,
		InvoiceID:      receivable.InvoiceID,
		InvoiceNumber:  row.InvoiceNumber,
		Reference:      receivable.Reference,
		DueDate:        receivable.DueDate,
		Amount:         receivable.Amount,
		PaidAmount:     receivable.PaidAmount,
		Outstanding:    receivable.Amount - receivable.PaidAmount,
		Status:         duedate.Status(receivable.DueDate, settled, today),
		DaysOverdue:    duedate.DaysOverdue(receivable.DueDate, settled, today),
		LastPaidAt:     row.LastPaidAt,
		CreatedAt:      receivable.CreatedAt,
		UpdatedAt:      receivable.UpdatedAt,
	}
}

func newPayableResponse(row PayableRow, today time.Time) PayableResponse {
	payable := row.Payable
	settled := settledAmount(payable.Amount, payable.PaidAmount)
	return PayableResponse{
		ID:                  payable.ID,
		VendorID:            payable.VendorID,
		VendorName:          row.VendorName,
		ProjectID:           payable.ProjectID,
		ProjectName:         row.ProjectName,
		PurchaseOrderID:     payable.PurchaseOrderID,
		PurchaseOrderNumber: row.PurchaseOrderNumber,
		Reference:           payable.Reference,
		DueDate:             payable.DueDate,
		Amount:              payable.Amount,
		PaidAmount:          payable.PaidAmount,
		Outstanding:         payable.Amount - payable.PaidAmount,
		Status:              duedate.Status(payable.DueDate, settled, today),
		DaysOverdue:         duedate.DaysOverdue(payable.DueDate, settled, today),
		LastPaidAt:          row.LastPaidAt,
		CreatedAt:           payable.CreatedAt,
		UpdatedAt:           payable.UpdatedAt,
	}
}
