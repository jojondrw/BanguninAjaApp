package procurement

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type VendorRequest struct {
	Code            string `json:"code" binding:"required,max=20"`
	Name            string `json:"name" binding:"required,max=160"`
	Category        string `json:"category" binding:"required,max=60"`
	Contact         string `json:"contact" binding:"max=60"`
	TaxNumber       string `json:"taxNumber" binding:"max=25"`
	PaymentTermDays int    `json:"paymentTermDays" binding:"min=0"`
	Rating          string `json:"rating" binding:"required,oneof=new good fair poor"`
	Active          *bool  `json:"active" binding:"required"`
}

type VendorQuery struct {
	pagination.Query
	Search   string `form:"search" binding:"omitempty,max=160"`
	Category string `form:"category" binding:"omitempty,max=60"`
	Rating   string `form:"rating" binding:"omitempty,oneof=new good fair poor"`
	Active   *bool  `form:"active"`
}

type VendorResponse struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Category        string    `json:"category"`
	Contact         string    `json:"contact"`
	TaxNumber       string    `json:"taxNumber"`
	PaymentTermDays int       `json:"paymentTermDays"`
	Rating          string    `json:"rating"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type PurchaseRequestItemRequest struct {
	MaterialID      uuid.UUID `json:"materialId" binding:"required"`
	Quantity        float64   `json:"quantity" binding:"required,gt=0"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId" binding:"required"`
}

type PurchaseRequestRequest struct {
	Number    string                       `json:"number" binding:"required,max=40"`
	ProjectID uuid.UUID                    `json:"projectId" binding:"required"`
	Date      time.Time                    `json:"date" binding:"required"`
	Note      string                       `json:"note" binding:"max=2000"`
	Items     []PurchaseRequestItemRequest `json:"items" binding:"required,min=1,dive"`
}

type PurchaseRequestStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=submitted approved rejected completed"`
}

type PurchaseRequestQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=40"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Status    string     `form:"status" binding:"omitempty,oneof=draft submitted approved rejected completed"`
	DateFrom  *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo    *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type PurchaseRequestResponse struct {
	ID          uuid.UUID `json:"id"`
	Number      string    `json:"number"`
	ProjectID   uuid.UUID `json:"projectId"`
	RequesterID uuid.UUID `json:"requesterId"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"`
	Note        string    `json:"note"`
	ItemCount   int       `json:"itemCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PurchaseRequestItemResponse struct {
	ID              uuid.UUID `json:"id"`
	MaterialID      uuid.UUID `json:"materialId"`
	Quantity        float64   `json:"quantity"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId"`
}

type PurchaseRequestDetailResponse struct {
	PurchaseRequestResponse
	Items []PurchaseRequestItemResponse `json:"items"`
}

type PurchaseOrderItemRequest struct {
	MaterialID      uuid.UUID `json:"materialId" binding:"required"`
	Quantity        float64   `json:"quantity" binding:"required,gt=0"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId" binding:"required"`
	UnitPrice       int64     `json:"unitPrice" binding:"min=0"`
}

type PurchaseOrderRequest struct {
	Number            string                     `json:"number" binding:"required,max=40"`
	VendorID          uuid.UUID                  `json:"vendorId" binding:"required"`
	ProjectID         uuid.UUID                  `json:"projectId" binding:"required"`
	PurchaseRequestID *uuid.UUID                 `json:"purchaseRequestId"`
	Date              time.Time                  `json:"date" binding:"required"`
	DueDate           *time.Time                 `json:"dueDate"`
	Items             []PurchaseOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type PurchaseOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=sent cancelled"`
}

type PurchaseOrderQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=40"`
	VendorID  *uuid.UUID `form:"vendorId,parser=encoding.TextUnmarshaler"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	Status    string     `form:"status" binding:"omitempty,oneof=draft sent partially_received completed cancelled"`
	DateFrom  *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo    *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type PurchaseOrderResponse struct {
	ID                uuid.UUID  `json:"id"`
	Number            string     `json:"number"`
	VendorID          uuid.UUID  `json:"vendorId"`
	ProjectID         uuid.UUID  `json:"projectId"`
	PurchaseRequestID *uuid.UUID `json:"purchaseRequestId"`
	Date              time.Time  `json:"date"`
	DueDate           *time.Time `json:"dueDate"`
	Value             int64      `json:"value"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type PurchaseOrderItemResponse struct {
	ID                uuid.UUID `json:"id"`
	MaterialID        uuid.UUID `json:"materialId"`
	Quantity          float64   `json:"quantity"`
	UnitOfMeasureID   uuid.UUID `json:"unitOfMeasureId"`
	UnitPrice         int64     `json:"unitPrice"`
	Total             int64     `json:"total"`
	ReceivedQuantity  float64   `json:"receivedQuantity"`
	RemainingQuantity float64   `json:"remainingQuantity"`
}

type PurchaseOrderDetailResponse struct {
	PurchaseOrderResponse
	Items []PurchaseOrderItemResponse `json:"items"`
}

type GoodsReceiptItemRequest struct {
	PurchaseOrderItemID uuid.UUID `json:"purchaseOrderItemId" binding:"required"`
	AcceptedQuantity    float64   `json:"acceptedQuantity" binding:"min=0"`
	RejectedQuantity    float64   `json:"rejectedQuantity" binding:"min=0"`
}

type GoodsReceiptRequest struct {
	Number          string                    `json:"number" binding:"required,max=40"`
	PurchaseOrderID uuid.UUID                 `json:"purchaseOrderId" binding:"required"`
	WarehouseID     uuid.UUID                 `json:"warehouseId" binding:"required"`
	Date            time.Time                 `json:"date" binding:"required"`
	Condition       string                    `json:"condition" binding:"required,oneof=good partially_damaged broken"`
	Note            string                    `json:"note" binding:"max=2000"`
	Items           []GoodsReceiptItemRequest `json:"items" binding:"required,min=1,dive"`
}

type GoodsReceiptQuery struct {
	pagination.Query
	Search          string     `form:"search" binding:"omitempty,max=40"`
	PurchaseOrderID *uuid.UUID `form:"purchaseOrderId,parser=encoding.TextUnmarshaler"`
	WarehouseID     *uuid.UUID `form:"warehouseId,parser=encoding.TextUnmarshaler"`
	DateFrom        *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo          *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type GoodsReceiptResponse struct {
	ID              uuid.UUID `json:"id"`
	Number          string    `json:"number"`
	PurchaseOrderID uuid.UUID `json:"purchaseOrderId"`
	WarehouseID     uuid.UUID `json:"warehouseId"`
	Date            time.Time `json:"date"`
	Condition       string    `json:"condition"`
	Note            string    `json:"note"`
	ItemCount       int       `json:"itemCount"`
	CreatedAt       time.Time `json:"createdAt"`
}

type GoodsReceiptItemResponse struct {
	ID                  uuid.UUID `json:"id"`
	PurchaseOrderItemID uuid.UUID `json:"purchaseOrderItemId"`
	AcceptedQuantity    float64   `json:"acceptedQuantity"`
	RejectedQuantity    float64   `json:"rejectedQuantity"`
}

type GoodsReceiptDetailResponse struct {
	GoodsReceiptResponse
	Items []GoodsReceiptItemResponse `json:"items"`
}

func newVendorResponse(vendor Vendor) VendorResponse {
	return VendorResponse{
		ID:              vendor.ID,
		Code:            vendor.Code,
		Name:            vendor.Name,
		Category:        vendor.Category,
		Contact:         vendor.Contact,
		TaxNumber:       vendor.TaxNumber,
		PaymentTermDays: vendor.PaymentTermDays,
		Rating:          vendor.Rating,
		Active:          vendor.Active,
		CreatedAt:       vendor.CreatedAt,
		UpdatedAt:       vendor.UpdatedAt,
	}
}

func newPurchaseRequestResponse(request PurchaseRequest, itemCount int) PurchaseRequestResponse {
	return PurchaseRequestResponse{
		ID:          request.ID,
		Number:      request.Number,
		ProjectID:   request.ProjectID,
		RequesterID: request.RequesterID,
		Date:        request.Date,
		Status:      request.Status,
		Note:        request.Note,
		ItemCount:   itemCount,
		CreatedAt:   request.CreatedAt,
		UpdatedAt:   request.UpdatedAt,
	}
}

func newPurchaseRequestRowResponse(row PurchaseRequestRow) PurchaseRequestResponse {
	return newPurchaseRequestResponse(row.PurchaseRequest, row.ItemCount)
}

func newPurchaseRequestItemResponse(item PurchaseRequestItem) PurchaseRequestItemResponse {
	return PurchaseRequestItemResponse{
		ID:              item.ID,
		MaterialID:      item.MaterialID,
		Quantity:        item.Quantity,
		UnitOfMeasureID: item.UnitOfMeasureID,
	}
}

func newPurchaseOrderResponse(order PurchaseOrder) PurchaseOrderResponse {
	return PurchaseOrderResponse{
		ID:                order.ID,
		Number:            order.Number,
		VendorID:          order.VendorID,
		ProjectID:         order.ProjectID,
		PurchaseRequestID: order.PurchaseRequestID,
		Date:              order.Date,
		DueDate:           order.DueDate,
		Value:             order.Value,
		Status:            order.Status,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}
}

func newPurchaseOrderItemResponse(item PurchaseOrderItem, received float64) PurchaseOrderItemResponse {
	return PurchaseOrderItemResponse{
		ID:                item.ID,
		MaterialID:        item.MaterialID,
		Quantity:          item.Quantity,
		UnitOfMeasureID:   item.UnitOfMeasureID,
		UnitPrice:         item.UnitPrice,
		Total:             lineTotal(item.Quantity, item.UnitPrice),
		ReceivedQuantity:  received,
		RemainingQuantity: roundQuantity(item.Quantity - received),
	}
}

func newGoodsReceiptResponse(receipt GoodsReceipt, itemCount int) GoodsReceiptResponse {
	return GoodsReceiptResponse{
		ID:              receipt.ID,
		Number:          receipt.Number,
		PurchaseOrderID: receipt.PurchaseOrderID,
		WarehouseID:     receipt.WarehouseID,
		Date:            receipt.Date,
		Condition:       receipt.Condition,
		Note:            receipt.Note,
		ItemCount:       itemCount,
		CreatedAt:       receipt.CreatedAt,
	}
}

func newGoodsReceiptRowResponse(row GoodsReceiptRow) GoodsReceiptResponse {
	return newGoodsReceiptResponse(row.GoodsReceipt, row.ItemCount)
}

func newGoodsReceiptItemResponse(item GoodsReceiptItem) GoodsReceiptItemResponse {
	return GoodsReceiptItemResponse{
		ID:                  item.ID,
		PurchaseOrderItemID: item.PurchaseOrderItemID,
		AcceptedQuantity:    item.AcceptedQuantity,
		RejectedQuantity:    item.RejectedQuantity,
	}
}
