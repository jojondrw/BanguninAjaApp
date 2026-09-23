package procurement

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	Repository
	vendor     Vendor
	source     PurchaseRequest
	order      PurchaseOrder
	orderItems []PurchaseOrderItem
	received   []ReceivedQuantity
	savedOrder *PurchaseOrder
	receipts   int
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *fakeRepository) FindVendor(context.Context, uuid.UUID) (Vendor, error) {
	return f.vendor, nil
}

func (f *fakeRepository) FindPurchaseRequest(context.Context, uuid.UUID) (PurchaseRequest, error) {
	return f.source, nil
}

func (f *fakeRepository) LockPurchaseOrder(context.Context, uuid.UUID) (PurchaseOrder, error) {
	return f.order, nil
}

func (f *fakeRepository) SavePurchaseOrder(_ context.Context, order *PurchaseOrder) error {
	f.savedOrder = order
	return nil
}

func (f *fakeRepository) ListPurchaseOrderItems(context.Context, uuid.UUID) ([]PurchaseOrderItem, error) {
	return f.orderItems, nil
}

func (f *fakeRepository) SumAcceptedQuantities(context.Context, uuid.UUID) ([]ReceivedQuantity, error) {
	return f.received, nil
}

func (f *fakeRepository) CreateGoodsReceipt(_ context.Context, receipt *GoodsReceipt) error {
	receipt.ID = uuid.New()
	f.receipts++
	return nil
}

func (f *fakeRepository) CreateGoodsReceiptItems(context.Context, []GoodsReceiptItem) error {
	return nil
}

func (f *fakeRepository) FindGoodsReceipt(context.Context, uuid.UUID) (GoodsReceipt, error) {
	return GoodsReceipt{}, nil
}

func (f *fakeRepository) ListGoodsReceiptItems(context.Context, uuid.UUID) ([]GoodsReceiptItem, error) {
	return nil, nil
}

func orderItem(quantity float64) PurchaseOrderItem {
	item := PurchaseOrderItem{Quantity: quantity}
	item.ID = uuid.New()
	return item
}

func TestOrderValueSumsRoundedLines(t *testing.T) {
	value := orderValue([]PurchaseOrderItemRequest{
		{Quantity: 12.345, UnitPrice: 15000},
		{Quantity: 2, UnitPrice: 1250},
	})
	if value != 187750 {
		t.Fatalf("got %d, want 187750", value)
	}
}

func TestOrderFromUnapprovedRequestIsRejected(t *testing.T) {
	projectID := uuid.New()
	cases := []struct {
		source PurchaseRequest
		want   error
	}{
		{PurchaseRequest{Status: requestSubmitted, ProjectID: projectID}, errRequestUnapproved},
		{PurchaseRequest{Status: requestApproved, ProjectID: uuid.New()}, errRequestProject},
		{PurchaseRequest{Status: requestApproved, ProjectID: projectID}, nil},
	}
	for _, tc := range cases {
		if got := ensureOrderableRequest(tc.source, projectID); !errors.Is(got, tc.want) {
			t.Errorf("status %s: got %v, want %v", tc.source.Status, got, tc.want)
		}
	}
}

func TestOrderFromInactiveVendorIsRejected(t *testing.T) {
	repository := &fakeRepository{vendor: Vendor{Active: false}}
	err := ensureOrderSources(context.Background(), repository, PurchaseOrder{})
	if !errors.Is(err, errVendorInactive) {
		t.Fatalf("got %v, want errVendorInactive", err)
	}
}

func TestPartialReceiptMarksOrderPartiallyReceived(t *testing.T) {
	first, second := orderItem(10), orderItem(5)
	repository := &fakeRepository{order: PurchaseOrder{Status: orderSent}, orderItems: []PurchaseOrderItem{first, second}}

	_, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Items: []GoodsReceiptItemRequest{{PurchaseOrderItemID: first.ID, AcceptedQuantity: 10}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.savedOrder.Status != orderPartiallyReceived {
		t.Fatalf("got %s", repository.savedOrder.Status)
	}
}

func TestFinalReceiptCompletesOrder(t *testing.T) {
	first := orderItem(10)
	repository := &fakeRepository{
		order:      PurchaseOrder{Status: orderPartiallyReceived},
		orderItems: []PurchaseOrderItem{first},
		received:   []ReceivedQuantity{{PurchaseOrderItemID: first.ID, Quantity: 6.4}},
	}

	_, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Items: []GoodsReceiptItemRequest{{PurchaseOrderItemID: first.ID, AcceptedQuantity: 3.6, RejectedQuantity: 1}},
	})
	if err != nil || repository.savedOrder.Status != orderCompleted {
		t.Fatalf("got err %v order %+v", err, repository.savedOrder)
	}
}

func TestReceiptCannotExceedOrder(t *testing.T) {
	first := orderItem(10)
	repository := &fakeRepository{
		order:      PurchaseOrder{Status: orderSent},
		orderItems: []PurchaseOrderItem{first},
		received:   []ReceivedQuantity{{PurchaseOrderItemID: first.ID, Quantity: 8}},
	}

	_, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Items: []GoodsReceiptItemRequest{{PurchaseOrderItemID: first.ID, AcceptedQuantity: 3}},
	})
	if !errors.Is(err, errReceiptExceedsOrder) || repository.receipts != 0 {
		t.Fatalf("got %v receipts %d", err, repository.receipts)
	}
}

func TestReceiptForDraftOrderIsRejected(t *testing.T) {
	repository := &fakeRepository{order: PurchaseOrder{Status: orderDraft}}

	_, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Items: []GoodsReceiptItemRequest{{PurchaseOrderItemID: uuid.New(), AcceptedQuantity: 1}},
	})
	if !errors.Is(err, errOrderNotReceiving) {
		t.Fatalf("got %v, want errOrderNotReceiving", err)
	}
}

func TestValidateReceiptItems(t *testing.T) {
	line := uuid.New()
	if err := validateReceiptItems([]GoodsReceiptItemRequest{{PurchaseOrderItemID: line}}); !errors.Is(err, errReceiptQuantityEmpty) {
		t.Fatalf("empty line: got %v", err)
	}
	duplicate := []GoodsReceiptItemRequest{{PurchaseOrderItemID: line, AcceptedQuantity: 1}, {PurchaseOrderItemID: line, AcceptedQuantity: 1}}
	if err := validateReceiptItems(duplicate); !errors.Is(err, errReceiptItemDuplicate) {
		t.Fatalf("duplicate line: got %v", err)
	}
}
