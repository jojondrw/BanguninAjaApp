package procurement

import (
	"context"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	requestDraft     = "draft"
	requestSubmitted = "submitted"
	requestApproved  = "approved"
	requestRejected  = "rejected"
	requestCompleted = "completed"

	orderDraft             = "draft"
	orderSent              = "sent"
	orderPartiallyReceived = "partially_received"
	orderCompleted         = "completed"
	orderCancelled         = "cancelled"

	quantityPrecision = 100
)

var requestTransitions = map[string][]string{
	requestDraft:     {requestSubmitted},
	requestSubmitted: {requestApproved, requestRejected},
	requestApproved:  {requestCompleted},
}

var orderTransitions = map[string][]string{
	orderDraft: {orderSent, orderCancelled},
	orderSent:  {orderCancelled},
}

var (
	errVendorNotFound = apperror.NotFound("vendor_not_found", "Vendor tidak ditemukan")
	errVendorCodeUsed = apperror.Conflict("vendor_code_used", "Kode vendor sudah dipakai")
	errVendorInUse    = apperror.Conflict("vendor_in_use", "Vendor masih dipakai oleh pesanan atau hutang")
	errVendorUnknown  = apperror.Unprocessable("vendor_not_found", "Vendor tidak ditemukan")
	errVendorInactive = apperror.Unprocessable("vendor_inactive", "Vendor sudah tidak aktif")

	errRequestNotFound   = apperror.NotFound("purchase_request_not_found", "Permintaan pembelian tidak ditemukan")
	errRequestNumberUsed = apperror.Conflict("purchase_request_number_used", "Nomor permintaan pembelian sudah dipakai")
	errRequestReference  = apperror.Unprocessable("purchase_request_reference_not_found", "Proyek, material, atau satuan tidak ditemukan")
	errRequestLocked     = apperror.Conflict("purchase_request_locked", "Permintaan pembelian hanya bisa diubah atau dihapus selama masih draf")
	errRequestTransition = apperror.Unprocessable("purchase_request_status_transition_invalid", "Perubahan status permintaan pembelian tidak diizinkan")
	errRequestUnknown    = apperror.Unprocessable("purchase_request_not_found", "Permintaan pembelian tidak ditemukan")
	errRequestUnapproved = apperror.Unprocessable("purchase_request_not_approved", "Pesanan hanya bisa dibuat dari permintaan yang sudah disetujui")
	errRequestProject    = apperror.Unprocessable("purchase_request_project_mismatch", "Proyek pesanan harus sama dengan proyek permintaan")

	errOrderNotFound     = apperror.NotFound("purchase_order_not_found", "Pesanan pembelian tidak ditemukan")
	errOrderNumberUsed   = apperror.Conflict("purchase_order_number_used", "Nomor pesanan pembelian sudah dipakai")
	errOrderReference    = apperror.Unprocessable("purchase_order_reference_not_found", "Proyek, material, atau satuan tidak ditemukan")
	errOrderLocked       = apperror.Conflict("purchase_order_locked", "Pesanan pembelian hanya bisa diubah atau dihapus selama masih draf")
	errOrderTransition   = apperror.Unprocessable("purchase_order_status_transition_invalid", "Perubahan status pesanan pembelian tidak diizinkan")
	errOrderNotReceiving = apperror.Conflict("purchase_order_not_receivable", "Barang hanya bisa diterima untuk pesanan yang sudah dikirim ke vendor")
	errMaterialDuplicate = apperror.Unprocessable("material_duplicate", "Satu material hanya boleh muncul sekali dalam satu dokumen")
	errInvalidDateRange  = apperror.Unprocessable("invalid_date_range", "Tanggal jatuh tempo tidak boleh lebih awal dari tanggal pesanan")

	errReceiptNotFound      = apperror.NotFound("goods_receipt_not_found", "Penerimaan barang tidak ditemukan")
	errReceiptNumberUsed    = apperror.Conflict("goods_receipt_number_used", "Nomor penerimaan barang sudah dipakai")
	errReceiptWarehouse     = apperror.Unprocessable("warehouse_not_found", "Gudang tidak ditemukan")
	errReceiptOrderUnknown  = apperror.Unprocessable("purchase_order_not_found", "Pesanan pembelian tidak ditemukan")
	errReceiptItemInvalid   = apperror.Unprocessable("goods_receipt_item_invalid", "Ada barang yang bukan bagian dari pesanan ini")
	errReceiptItemDuplicate = apperror.Unprocessable("goods_receipt_item_duplicate", "Satu baris pesanan hanya boleh muncul sekali dalam satu penerimaan")
	errReceiptQuantityEmpty = apperror.Unprocessable("goods_receipt_quantity_invalid", "Setiap baris penerimaan harus punya jumlah diterima atau ditolak")
	errReceiptExceedsOrder  = apperror.Unprocessable("goods_receipt_exceeds_order", "Jumlah yang diterima melebihi sisa pesanan")
)

var (
	vendorReadErrors   = database.ErrorMap{NotFound: errVendorNotFound}
	vendorWriteErrors  = database.ErrorMap{NotFound: errVendorNotFound, Duplicate: errVendorCodeUsed}
	vendorDeleteErrors = database.ErrorMap{NotFound: errVendorNotFound, Referenced: errVendorInUse}
	vendorLookupErrors = database.ErrorMap{NotFound: errVendorUnknown}

	requestReadErrors   = database.ErrorMap{NotFound: errRequestNotFound}
	requestWriteErrors  = database.ErrorMap{NotFound: errRequestNotFound, Duplicate: errRequestNumberUsed, Referenced: errRequestReference}
	requestLookupErrors = database.ErrorMap{NotFound: errRequestUnknown}
	requestItemErrors   = database.ErrorMap{Duplicate: errMaterialDuplicate, Referenced: errRequestReference}

	orderReadErrors   = database.ErrorMap{NotFound: errOrderNotFound}
	orderWriteErrors  = database.ErrorMap{NotFound: errOrderNotFound, Duplicate: errOrderNumberUsed, Referenced: errOrderReference}
	orderLookupErrors = database.ErrorMap{NotFound: errReceiptOrderUnknown}
	orderItemErrors   = database.ErrorMap{Duplicate: errMaterialDuplicate, Referenced: errOrderReference}

	receiptReadErrors  = database.ErrorMap{NotFound: errReceiptNotFound}
	receiptWriteErrors = database.ErrorMap{Duplicate: errReceiptNumberUsed, Referenced: errReceiptWarehouse}
	receiptItemErrors  = database.ErrorMap{Referenced: errReceiptItemInvalid}
)

type Service interface {
	ListVendors(ctx context.Context, query VendorQuery) (pagination.Page[VendorResponse], error)
	GetVendor(ctx context.Context, id uuid.UUID) (VendorResponse, error)
	CreateVendor(ctx context.Context, request VendorRequest) (VendorResponse, error)
	UpdateVendor(ctx context.Context, id uuid.UUID, request VendorRequest) (VendorResponse, error)
	DeleteVendor(ctx context.Context, id uuid.UUID) error

	ListPurchaseRequests(ctx context.Context, query PurchaseRequestQuery) (pagination.Page[PurchaseRequestResponse], error)
	GetPurchaseRequest(ctx context.Context, id uuid.UUID) (PurchaseRequestDetailResponse, error)
	CreatePurchaseRequest(ctx context.Context, requesterID uuid.UUID, request PurchaseRequestRequest) (PurchaseRequestDetailResponse, error)
	UpdatePurchaseRequest(ctx context.Context, id uuid.UUID, request PurchaseRequestRequest) (PurchaseRequestDetailResponse, error)
	UpdatePurchaseRequestStatus(ctx context.Context, id uuid.UUID, request PurchaseRequestStatusRequest) (PurchaseRequestDetailResponse, error)
	DeletePurchaseRequest(ctx context.Context, id uuid.UUID) error

	ListPurchaseOrders(ctx context.Context, query PurchaseOrderQuery) (pagination.Page[PurchaseOrderResponse], error)
	GetPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrderDetailResponse, error)
	CreatePurchaseOrder(ctx context.Context, request PurchaseOrderRequest) (PurchaseOrderDetailResponse, error)
	UpdatePurchaseOrder(ctx context.Context, id uuid.UUID, request PurchaseOrderRequest) (PurchaseOrderDetailResponse, error)
	UpdatePurchaseOrderStatus(ctx context.Context, id uuid.UUID, request PurchaseOrderStatusRequest) (PurchaseOrderDetailResponse, error)
	DeletePurchaseOrder(ctx context.Context, id uuid.UUID) error

	ListGoodsReceipts(ctx context.Context, query GoodsReceiptQuery) (pagination.Page[GoodsReceiptResponse], error)
	GetGoodsReceipt(ctx context.Context, id uuid.UUID) (GoodsReceiptDetailResponse, error)
	RecordGoodsReceipt(ctx context.Context, request GoodsReceiptRequest) (GoodsReceiptDetailResponse, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListVendors(ctx context.Context, query VendorQuery) (pagination.Page[VendorResponse], error) {
	vendors, total, err := s.repository.ListVendors(ctx, VendorFilter{
		Search:   query.Search,
		Category: strings.TrimSpace(query.Category),
		Rating:   query.Rating,
		Active:   query.Active,
		Offset:   query.Offset(),
		Limit:    query.Size(),
	})
	if err != nil {
		return pagination.Page[VendorResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(vendors, newVendorResponse), query.Query, total), nil
}

func (s *service) GetVendor(ctx context.Context, id uuid.UUID) (VendorResponse, error) {
	vendor, err := s.repository.FindVendor(ctx, id)
	if err != nil {
		return VendorResponse{}, vendorReadErrors.Resolve(err)
	}
	return newVendorResponse(vendor), nil
}

func (s *service) CreateVendor(ctx context.Context, request VendorRequest) (VendorResponse, error) {
	var vendor Vendor
	applyVendorRequest(&vendor, request)
	if err := s.repository.CreateVendor(ctx, &vendor); err != nil {
		return VendorResponse{}, vendorWriteErrors.Resolve(err)
	}
	return newVendorResponse(vendor), nil
}

func (s *service) UpdateVendor(ctx context.Context, id uuid.UUID, request VendorRequest) (VendorResponse, error) {
	vendor, err := s.repository.FindVendor(ctx, id)
	if err != nil {
		return VendorResponse{}, vendorReadErrors.Resolve(err)
	}

	applyVendorRequest(&vendor, request)
	if err := s.repository.SaveVendor(ctx, &vendor); err != nil {
		return VendorResponse{}, vendorWriteErrors.Resolve(err)
	}
	return newVendorResponse(vendor), nil
}

func (s *service) DeleteVendor(ctx context.Context, id uuid.UUID) error {
	return vendorDeleteErrors.Resolve(s.repository.DeleteVendor(ctx, id))
}

func (s *service) ListPurchaseRequests(ctx context.Context, query PurchaseRequestQuery) (pagination.Page[PurchaseRequestResponse], error) {
	rows, total, err := s.repository.ListPurchaseRequests(ctx, PurchaseRequestFilter{
		Search:    query.Search,
		ProjectID: query.ProjectID,
		Status:    query.Status,
		DateFrom:  query.DateFrom,
		DateTo:    query.DateTo,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[PurchaseRequestResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newPurchaseRequestRowResponse), query.Query, total), nil
}

func (s *service) GetPurchaseRequest(ctx context.Context, id uuid.UUID) (PurchaseRequestDetailResponse, error) {
	request, err := s.repository.FindPurchaseRequest(ctx, id)
	if err != nil {
		return PurchaseRequestDetailResponse{}, requestReadErrors.Resolve(err)
	}

	items, err := s.repository.ListPurchaseRequestItems(ctx, id)
	if err != nil {
		return PurchaseRequestDetailResponse{}, apperror.Internal(err)
	}
	return PurchaseRequestDetailResponse{
		PurchaseRequestResponse: newPurchaseRequestResponse(request, len(items)),
		Items:                   pagination.Map(items, newPurchaseRequestItemResponse),
	}, nil
}

func (s *service) CreatePurchaseRequest(ctx context.Context, requesterID uuid.UUID, request PurchaseRequestRequest) (PurchaseRequestDetailResponse, error) {
	if err := ensureDistinctMaterials(requestMaterials(request.Items)); err != nil {
		return PurchaseRequestDetailResponse{}, err
	}

	document := PurchaseRequest{RequesterID: requesterID, Status: requestDraft}
	applyPurchaseRequestRequest(&document, request)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createPurchaseRequest(ctx, repository, &document, request.Items)
	})
	if err != nil {
		return PurchaseRequestDetailResponse{}, apperror.From(err)
	}
	return s.GetPurchaseRequest(ctx, document.ID)
}

func (s *service) UpdatePurchaseRequest(ctx context.Context, id uuid.UUID, request PurchaseRequestRequest) (PurchaseRequestDetailResponse, error) {
	if err := ensureDistinctMaterials(requestMaterials(request.Items)); err != nil {
		return PurchaseRequestDetailResponse{}, err
	}

	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return updatePurchaseRequest(ctx, repository, id, request)
	})
	if err != nil {
		return PurchaseRequestDetailResponse{}, apperror.From(err)
	}
	return s.GetPurchaseRequest(ctx, id)
}

func (s *service) UpdatePurchaseRequestStatus(ctx context.Context, id uuid.UUID, request PurchaseRequestStatusRequest) (PurchaseRequestDetailResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return changePurchaseRequestStatus(ctx, repository, id, request.Status)
	})
	if err != nil {
		return PurchaseRequestDetailResponse{}, apperror.From(err)
	}
	return s.GetPurchaseRequest(ctx, id)
}

func (s *service) DeletePurchaseRequest(ctx context.Context, id uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return deletePurchaseRequest(ctx, repository, id)
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) ListPurchaseOrders(ctx context.Context, query PurchaseOrderQuery) (pagination.Page[PurchaseOrderResponse], error) {
	orders, total, err := s.repository.ListPurchaseOrders(ctx, PurchaseOrderFilter{
		Search:    query.Search,
		VendorID:  query.VendorID,
		ProjectID: query.ProjectID,
		Status:    query.Status,
		DateFrom:  query.DateFrom,
		DateTo:    query.DateTo,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[PurchaseOrderResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(orders, newPurchaseOrderResponse), query.Query, total), nil
}

func (s *service) GetPurchaseOrder(ctx context.Context, id uuid.UUID) (PurchaseOrderDetailResponse, error) {
	order, err := s.repository.FindPurchaseOrder(ctx, id)
	if err != nil {
		return PurchaseOrderDetailResponse{}, orderReadErrors.Resolve(err)
	}

	items, err := s.repository.ListPurchaseOrderItems(ctx, id)
	if err != nil {
		return PurchaseOrderDetailResponse{}, apperror.Internal(err)
	}

	received, err := s.repository.SumAcceptedQuantities(ctx, id)
	if err != nil {
		return PurchaseOrderDetailResponse{}, apperror.Internal(err)
	}
	return newPurchaseOrderDetail(order, items, receivedByItem(received)), nil
}

func (s *service) CreatePurchaseOrder(ctx context.Context, request PurchaseOrderRequest) (PurchaseOrderDetailResponse, error) {
	if err := validatePurchaseOrder(request); err != nil {
		return PurchaseOrderDetailResponse{}, err
	}

	order := PurchaseOrder{Status: orderDraft}
	applyPurchaseOrderRequest(&order, request)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return createPurchaseOrder(ctx, repository, &order, request.Items)
	})
	if err != nil {
		return PurchaseOrderDetailResponse{}, apperror.From(err)
	}
	return s.GetPurchaseOrder(ctx, order.ID)
}

func (s *service) UpdatePurchaseOrder(ctx context.Context, id uuid.UUID, request PurchaseOrderRequest) (PurchaseOrderDetailResponse, error) {
	if err := validatePurchaseOrder(request); err != nil {
		return PurchaseOrderDetailResponse{}, err
	}

	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return updatePurchaseOrder(ctx, repository, id, request)
	})
	if err != nil {
		return PurchaseOrderDetailResponse{}, apperror.From(err)
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *service) UpdatePurchaseOrderStatus(ctx context.Context, id uuid.UUID, request PurchaseOrderStatusRequest) (PurchaseOrderDetailResponse, error) {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return changePurchaseOrderStatus(ctx, repository, id, request.Status)
	})
	if err != nil {
		return PurchaseOrderDetailResponse{}, apperror.From(err)
	}
	return s.GetPurchaseOrder(ctx, id)
}

func (s *service) DeletePurchaseOrder(ctx context.Context, id uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return deletePurchaseOrder(ctx, repository, id)
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) ListGoodsReceipts(ctx context.Context, query GoodsReceiptQuery) (pagination.Page[GoodsReceiptResponse], error) {
	rows, total, err := s.repository.ListGoodsReceipts(ctx, GoodsReceiptFilter{
		Search:          query.Search,
		PurchaseOrderID: query.PurchaseOrderID,
		WarehouseID:     query.WarehouseID,
		DateFrom:        query.DateFrom,
		DateTo:          query.DateTo,
		Offset:          query.Offset(),
		Limit:           query.Size(),
	})
	if err != nil {
		return pagination.Page[GoodsReceiptResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newGoodsReceiptRowResponse), query.Query, total), nil
}

func (s *service) GetGoodsReceipt(ctx context.Context, id uuid.UUID) (GoodsReceiptDetailResponse, error) {
	receipt, err := s.repository.FindGoodsReceipt(ctx, id)
	if err != nil {
		return GoodsReceiptDetailResponse{}, receiptReadErrors.Resolve(err)
	}

	items, err := s.repository.ListGoodsReceiptItems(ctx, id)
	if err != nil {
		return GoodsReceiptDetailResponse{}, apperror.Internal(err)
	}
	return GoodsReceiptDetailResponse{
		GoodsReceiptResponse: newGoodsReceiptResponse(receipt, len(items)),
		Items:                pagination.Map(items, newGoodsReceiptItemResponse),
	}, nil
}

func (s *service) RecordGoodsReceipt(ctx context.Context, request GoodsReceiptRequest) (GoodsReceiptDetailResponse, error) {
	if err := validateReceiptItems(request.Items); err != nil {
		return GoodsReceiptDetailResponse{}, err
	}

	receipt := newGoodsReceipt(request)
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return receiveGoods(ctx, repository, &receipt, request.Items)
	})
	if err != nil {
		return GoodsReceiptDetailResponse{}, apperror.From(err)
	}
	return s.GetGoodsReceipt(ctx, receipt.ID)
}

func createPurchaseRequest(ctx context.Context, repository Repository, document *PurchaseRequest, items []PurchaseRequestItemRequest) error {
	if err := repository.CreatePurchaseRequest(ctx, document); err != nil {
		return requestWriteErrors.Resolve(err)
	}
	return storePurchaseRequestItems(ctx, repository, document.ID, items)
}

func deletePurchaseRequest(ctx context.Context, repository Repository, id uuid.UUID) error {
	if _, err := lockDraftRequest(ctx, repository, id); err != nil {
		return err
	}
	return requestReadErrors.Resolve(repository.DeletePurchaseRequest(ctx, id))
}

func deletePurchaseOrder(ctx context.Context, repository Repository, id uuid.UUID) error {
	if _, err := lockDraftOrder(ctx, repository, id); err != nil {
		return err
	}
	return orderReadErrors.Resolve(repository.DeletePurchaseOrder(ctx, id))
}

func updatePurchaseRequest(ctx context.Context, repository Repository, id uuid.UUID, request PurchaseRequestRequest) error {
	document, err := lockDraftRequest(ctx, repository, id)
	if err != nil {
		return err
	}

	applyPurchaseRequestRequest(&document, request)
	if err := repository.SavePurchaseRequest(ctx, &document); err != nil {
		return requestWriteErrors.Resolve(err)
	}
	return storePurchaseRequestItems(ctx, repository, id, request.Items)
}

func changePurchaseRequestStatus(ctx context.Context, repository Repository, id uuid.UUID, next string) error {
	document, err := repository.LockPurchaseRequest(ctx, id)
	if err != nil {
		return requestReadErrors.Resolve(err)
	}
	if document.Status == next {
		return nil
	}
	if !slices.Contains(requestTransitions[document.Status], next) {
		return errRequestTransition
	}

	document.Status = next
	return requestWriteErrors.Resolve(repository.SavePurchaseRequest(ctx, &document))
}

func lockDraftRequest(ctx context.Context, repository Repository, id uuid.UUID) (PurchaseRequest, error) {
	document, err := repository.LockPurchaseRequest(ctx, id)
	if err != nil {
		return PurchaseRequest{}, requestReadErrors.Resolve(err)
	}
	if document.Status != requestDraft {
		return PurchaseRequest{}, errRequestLocked
	}
	return document, nil
}

func storePurchaseRequestItems(ctx context.Context, repository Repository, requestID uuid.UUID, requests []PurchaseRequestItemRequest) error {
	items := make([]PurchaseRequestItem, 0, len(requests))
	for _, request := range requests {
		items = append(items, PurchaseRequestItem{
			PurchaseRequestID: requestID,
			MaterialID:        request.MaterialID,
			Quantity:          roundQuantity(request.Quantity),
			UnitOfMeasureID:   request.UnitOfMeasureID,
		})
	}
	return requestItemErrors.Resolve(repository.ReplacePurchaseRequestItems(ctx, requestID, items))
}

func createPurchaseOrder(ctx context.Context, repository Repository, order *PurchaseOrder, items []PurchaseOrderItemRequest) error {
	if err := ensureOrderSources(ctx, repository, *order); err != nil {
		return err
	}
	if err := repository.CreatePurchaseOrder(ctx, order); err != nil {
		return orderWriteErrors.Resolve(err)
	}
	return storePurchaseOrderItems(ctx, repository, order.ID, items)
}

func updatePurchaseOrder(ctx context.Context, repository Repository, id uuid.UUID, request PurchaseOrderRequest) error {
	order, err := lockDraftOrder(ctx, repository, id)
	if err != nil {
		return err
	}

	applyPurchaseOrderRequest(&order, request)
	if err := ensureOrderSources(ctx, repository, order); err != nil {
		return err
	}
	if err := repository.SavePurchaseOrder(ctx, &order); err != nil {
		return orderWriteErrors.Resolve(err)
	}
	return storePurchaseOrderItems(ctx, repository, id, request.Items)
}

func changePurchaseOrderStatus(ctx context.Context, repository Repository, id uuid.UUID, next string) error {
	order, err := repository.LockPurchaseOrder(ctx, id)
	if err != nil {
		return orderReadErrors.Resolve(err)
	}
	if order.Status == next {
		return nil
	}
	if !slices.Contains(orderTransitions[order.Status], next) {
		return errOrderTransition
	}

	order.Status = next
	return orderWriteErrors.Resolve(repository.SavePurchaseOrder(ctx, &order))
}

func lockDraftOrder(ctx context.Context, repository Repository, id uuid.UUID) (PurchaseOrder, error) {
	order, err := repository.LockPurchaseOrder(ctx, id)
	if err != nil {
		return PurchaseOrder{}, orderReadErrors.Resolve(err)
	}
	if order.Status != orderDraft {
		return PurchaseOrder{}, errOrderLocked
	}
	return order, nil
}

func ensureOrderSources(ctx context.Context, repository Repository, order PurchaseOrder) error {
	vendor, err := repository.FindVendor(ctx, order.VendorID)
	if err != nil {
		return vendorLookupErrors.Resolve(err)
	}
	if !vendor.Active {
		return errVendorInactive
	}
	if order.PurchaseRequestID == nil {
		return nil
	}

	source, err := repository.FindPurchaseRequest(ctx, *order.PurchaseRequestID)
	if err != nil {
		return requestLookupErrors.Resolve(err)
	}
	return ensureOrderableRequest(source, order.ProjectID)
}

func ensureOrderableRequest(source PurchaseRequest, projectID uuid.UUID) error {
	if source.Status != requestApproved && source.Status != requestCompleted {
		return errRequestUnapproved
	}
	if source.ProjectID != projectID {
		return errRequestProject
	}
	return nil
}

func storePurchaseOrderItems(ctx context.Context, repository Repository, orderID uuid.UUID, requests []PurchaseOrderItemRequest) error {
	items := make([]PurchaseOrderItem, 0, len(requests))
	for _, request := range requests {
		items = append(items, PurchaseOrderItem{
			PurchaseOrderID: orderID,
			MaterialID:      request.MaterialID,
			Quantity:        roundQuantity(request.Quantity),
			UnitOfMeasureID: request.UnitOfMeasureID,
			UnitPrice:       request.UnitPrice,
		})
	}
	return orderItemErrors.Resolve(repository.ReplacePurchaseOrderItems(ctx, orderID, items))
}

func receiveGoods(ctx context.Context, repository Repository, receipt *GoodsReceipt, requests []GoodsReceiptItemRequest) error {
	order, err := repository.LockPurchaseOrder(ctx, receipt.PurchaseOrderID)
	if err != nil {
		return orderLookupErrors.Resolve(err)
	}
	if order.Status != orderSent && order.Status != orderPartiallyReceived {
		return errOrderNotReceiving
	}

	ordered, received, err := orderProgress(ctx, repository, order.ID)
	if err != nil {
		return err
	}
	updated, err := applyReceipt(ordered, received, requests)
	if err != nil {
		return err
	}
	if err := storeGoodsReceipt(ctx, repository, receipt, requests); err != nil {
		return err
	}

	order.Status = statusAfterReceipt(ordered, updated)
	return orderWriteErrors.Resolve(repository.SavePurchaseOrder(ctx, &order))
}

func orderProgress(ctx context.Context, repository Repository, orderID uuid.UUID) (map[uuid.UUID]float64, map[uuid.UUID]float64, error) {
	items, err := repository.ListPurchaseOrderItems(ctx, orderID)
	if err != nil {
		return nil, nil, apperror.Internal(err)
	}

	received, err := repository.SumAcceptedQuantities(ctx, orderID)
	if err != nil {
		return nil, nil, apperror.Internal(err)
	}

	ordered := make(map[uuid.UUID]float64, len(items))
	for _, item := range items {
		ordered[item.ID] = item.Quantity
	}
	return ordered, receivedByItem(received), nil
}

func applyReceipt(ordered, received map[uuid.UUID]float64, requests []GoodsReceiptItemRequest) (map[uuid.UUID]float64, error) {
	updated := maps.Clone(received)
	for _, request := range requests {
		limit, belongs := ordered[request.PurchaseOrderItemID]
		if !belongs {
			return nil, errReceiptItemInvalid
		}
		total := roundQuantity(updated[request.PurchaseOrderItemID] + request.AcceptedQuantity)
		if total > limit {
			return nil, errReceiptExceedsOrder
		}
		updated[request.PurchaseOrderItemID] = total
	}
	return updated, nil
}

func statusAfterReceipt(ordered, received map[uuid.UUID]float64) string {
	for itemID, quantity := range ordered {
		if received[itemID] < quantity {
			return orderPartiallyReceived
		}
	}
	return orderCompleted
}

func storeGoodsReceipt(ctx context.Context, repository Repository, receipt *GoodsReceipt, requests []GoodsReceiptItemRequest) error {
	if err := repository.CreateGoodsReceipt(ctx, receipt); err != nil {
		return receiptWriteErrors.Resolve(err)
	}

	items := make([]GoodsReceiptItem, 0, len(requests))
	for _, request := range requests {
		items = append(items, GoodsReceiptItem{
			GoodsReceiptID:      receipt.ID,
			PurchaseOrderItemID: request.PurchaseOrderItemID,
			AcceptedQuantity:    roundQuantity(request.AcceptedQuantity),
			RejectedQuantity:    roundQuantity(request.RejectedQuantity),
		})
	}
	return receiptItemErrors.Resolve(repository.CreateGoodsReceiptItems(ctx, items))
}

func validatePurchaseOrder(request PurchaseOrderRequest) error {
	if request.DueDate != nil && request.DueDate.Before(request.Date) {
		return errInvalidDateRange
	}
	return ensureDistinctMaterials(orderMaterials(request.Items))
}

func validateReceiptItems(requests []GoodsReceiptItemRequest) error {
	seen := make(map[uuid.UUID]bool, len(requests))
	for _, request := range requests {
		if seen[request.PurchaseOrderItemID] {
			return errReceiptItemDuplicate
		}
		if request.AcceptedQuantity+request.RejectedQuantity <= 0 {
			return errReceiptQuantityEmpty
		}
		seen[request.PurchaseOrderItemID] = true
	}
	return nil
}

func ensureDistinctMaterials(materials []uuid.UUID) error {
	seen := make(map[uuid.UUID]bool, len(materials))
	for _, material := range materials {
		if seen[material] {
			return errMaterialDuplicate
		}
		seen[material] = true
	}
	return nil
}

func requestMaterials(items []PurchaseRequestItemRequest) []uuid.UUID {
	return pagination.Map(items, func(item PurchaseRequestItemRequest) uuid.UUID { return item.MaterialID })
}

func orderMaterials(items []PurchaseOrderItemRequest) []uuid.UUID {
	return pagination.Map(items, func(item PurchaseOrderItemRequest) uuid.UUID { return item.MaterialID })
}

func receivedByItem(quantities []ReceivedQuantity) map[uuid.UUID]float64 {
	received := make(map[uuid.UUID]float64, len(quantities))
	for _, quantity := range quantities {
		received[quantity.PurchaseOrderItemID] = roundQuantity(quantity.Quantity)
	}
	return received
}

func newPurchaseOrderDetail(order PurchaseOrder, items []PurchaseOrderItem, received map[uuid.UUID]float64) PurchaseOrderDetailResponse {
	responses := make([]PurchaseOrderItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, newPurchaseOrderItemResponse(item, received[item.ID]))
	}
	return PurchaseOrderDetailResponse{PurchaseOrderResponse: newPurchaseOrderResponse(order), Items: responses}
}

func orderValue(items []PurchaseOrderItemRequest) int64 {
	var value int64
	for _, item := range items {
		value += lineTotal(roundQuantity(item.Quantity), item.UnitPrice)
	}
	return value
}

func lineTotal(quantity float64, unitPrice int64) int64 {
	return int64(math.Round(quantity * float64(unitPrice)))
}

func roundQuantity(quantity float64) float64 {
	return math.Round(quantity*quantityPrecision) / quantityPrecision
}

func newGoodsReceipt(request GoodsReceiptRequest) GoodsReceipt {
	return GoodsReceipt{
		Number:          strings.TrimSpace(request.Number),
		PurchaseOrderID: request.PurchaseOrderID,
		WarehouseID:     request.WarehouseID,
		Date:            request.Date,
		Condition:       request.Condition,
		Note:            strings.TrimSpace(request.Note),
	}
}

func applyVendorRequest(vendor *Vendor, request VendorRequest) {
	vendor.Code = strings.TrimSpace(request.Code)
	vendor.Name = strings.TrimSpace(request.Name)
	vendor.Category = strings.TrimSpace(request.Category)
	vendor.Contact = strings.TrimSpace(request.Contact)
	vendor.TaxNumber = strings.TrimSpace(request.TaxNumber)
	vendor.PaymentTermDays = request.PaymentTermDays
	vendor.Rating = request.Rating
	vendor.Active = *request.Active
}

func applyPurchaseRequestRequest(document *PurchaseRequest, request PurchaseRequestRequest) {
	document.Number = strings.TrimSpace(request.Number)
	document.ProjectID = request.ProjectID
	document.Date = request.Date
	document.Note = strings.TrimSpace(request.Note)
}

func applyPurchaseOrderRequest(order *PurchaseOrder, request PurchaseOrderRequest) {
	order.Number = strings.TrimSpace(request.Number)
	order.VendorID = request.VendorID
	order.ProjectID = request.ProjectID
	order.PurchaseRequestID = request.PurchaseRequestID
	order.Date = request.Date
	order.DueDate = request.DueDate
	order.Value = orderValue(request.Items)
}
