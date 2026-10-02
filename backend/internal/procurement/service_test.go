package procurement

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

type fakeLedger struct {
	recorded []IncomingStock
	err      error
}

func (l *fakeLedger) RecordIncoming(_ context.Context, stock IncomingStock) (uuid.UUID, error) {
	if l.err != nil {
		return uuid.Nil, l.err
	}
	l.recorded = append(l.recorded, stock)
	return uuid.New(), nil
}

type fakeRepository struct {
	Repository
	vendor     Vendor
	source     PurchaseRequest
	order      PurchaseOrder
	orderItems []PurchaseOrderItem
	received   []ReceivedQuantity
	materials  []MaterialStockUnit
	unitCodes  map[uuid.UUID]string
	ledger     fakeLedger
	savedOrder *PurchaseOrder
	receipt    GoodsReceipt
	items      []GoodsReceiptItem
	receipts   int
	rolledBack bool
}

func (f *fakeRepository) Transaction(_ context.Context, work func(Repository) error) error {
	err := work(f)
	f.rolledBack = err != nil
	return err
}

func (f *fakeRepository) StockLedger() StockLedger {
	return &f.ledger
}

func (f *fakeRepository) ListMaterialStockUnits(context.Context, []uuid.UUID) ([]MaterialStockUnit, error) {
	return f.materials, nil
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
	f.receipt = *receipt
	f.receipts++
	return nil
}

func (f *fakeRepository) CreateGoodsReceiptItems(_ context.Context, items []GoodsReceiptItem) error {
	for index := range items {
		items[index].ID = uuid.New()
	}
	f.items = append(f.items, items...)
	return nil
}

func (f *fakeRepository) FindGoodsReceipt(context.Context, uuid.UUID) (GoodsReceipt, error) {
	return f.receipt, nil
}

func (f *fakeRepository) ListGoodsReceiptLines(context.Context, uuid.UUID) ([]GoodsReceiptLine, error) {
	lines := make([]GoodsReceiptLine, 0, len(f.items))
	for _, item := range f.items {
		line := GoodsReceiptLine{GoodsReceiptItem: item}
		for _, orderLine := range f.orderItems {
			if orderLine.ID == item.PurchaseOrderItemID {
				line.MaterialID, line.UnitOfMeasureID = orderLine.MaterialID, orderLine.UnitOfMeasureID
			}
		}
		for _, material := range f.materials {
			if material.ID == line.MaterialID {
				line.MaterialName, line.StockUnitOfMeasureID = material.Name, material.UnitOfMeasureID
			}
		}
		line.UnitOfMeasureCode = f.unitCodes[line.UnitOfMeasureID]
		line.StockUnitOfMeasureCode = f.unitCodes[line.StockUnitOfMeasureID]
		lines = append(lines, line)
	}
	return lines, nil
}

func orderItem(quantity float64) PurchaseOrderItem {
	item := PurchaseOrderItem{Quantity: quantity}
	item.ID = uuid.New()
	return item
}

func orderLineOf(quantity float64, materialID, unitID uuid.UUID) PurchaseOrderItem {
	item := orderItem(quantity)
	item.MaterialID, item.UnitOfMeasureID = materialID, unitID
	return item
}

func storedItemFor(t *testing.T, repository *fakeRepository, line PurchaseOrderItem) GoodsReceiptItem {
	t.Helper()
	for _, item := range repository.items {
		if item.PurchaseOrderItemID == line.ID {
			return item
		}
	}
	t.Fatalf("no stored receipt item for order line %s", line.ID)
	return GoodsReceiptItem{}
}

func responseItemFor(t *testing.T, response GoodsReceiptDetailResponse, line PurchaseOrderItem) GoodsReceiptItemResponse {
	t.Helper()
	for _, item := range response.Items {
		if item.PurchaseOrderItemID == line.ID {
			return item
		}
	}
	t.Fatalf("no response item for order line %s", line.ID)
	return GoodsReceiptItemResponse{}
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

func TestReceiptPostsStockForMatchedLines(t *testing.T) {
	sak := uuid.New()
	semen, pasir := orderLineOf(10, uuid.New(), sak), orderLineOf(5, uuid.New(), sak)
	warehouse := uuid.New()
	date := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	repository := &fakeRepository{
		order:      PurchaseOrder{Status: orderSent},
		orderItems: []PurchaseOrderItem{semen, pasir},
		materials: []MaterialStockUnit{
			{ID: semen.MaterialID, Name: "Semen Portland 50 kg", UnitOfMeasureID: sak},
			{ID: pasir.MaterialID, Name: "Pasir cor", UnitOfMeasureID: sak},
		},
	}

	response, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Number: " UJI-BPB-001 ", WarehouseID: warehouse, Date: date,
		Items: []GoodsReceiptItemRequest{
			{PurchaseOrderItemID: semen.ID, AcceptedQuantity: 10},
			{PurchaseOrderItemID: pasir.ID, AcceptedQuantity: 3.5, RejectedQuantity: 1.5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []IncomingStock{
		{Date: date, MaterialID: semen.MaterialID, WarehouseID: warehouse, Quantity: 10, Reference: "UJI-BPB-001"},
		{Date: date, MaterialID: pasir.MaterialID, WarehouseID: warehouse, Quantity: 3.5, Reference: "UJI-BPB-001"},
	}
	if len(repository.ledger.recorded) != len(want) {
		t.Fatalf("got %d stock postings, want %d", len(repository.ledger.recorded), len(want))
	}
	for index, posting := range repository.ledger.recorded {
		if posting != want[index] {
			t.Errorf("posting %d: got %+v, want %+v", index, posting, want[index])
		}
	}
	for _, line := range []PurchaseOrderItem{semen, pasir} {
		if stored := storedItemFor(t, repository, line); stored.StockMovementID == nil || stored.StockSkipReason != "" {
			t.Errorf("stored item %+v must carry its stock movement", stored)
		}
		if item := responseItemFor(t, response, line); item.StockStatus != stockPosted || !item.StockPosted || item.StockMovementID == nil {
			t.Errorf("response item %+v must report posted stock", item)
		}
	}
	if len(response.UnpostedLines) != 0 || repository.savedOrder.Status != orderPartiallyReceived {
		t.Fatalf("got unposted %+v order %s", response.UnpostedLines, repository.savedOrder.Status)
	}
}

func TestReceiptDoesNotPostRejectedOnlyLines(t *testing.T) {
	sak := uuid.New()
	semen := orderLineOf(10, uuid.New(), sak)
	repository := &fakeRepository{
		order:      PurchaseOrder{Status: orderSent},
		orderItems: []PurchaseOrderItem{semen},
		materials:  []MaterialStockUnit{{ID: semen.MaterialID, Name: "Semen Portland 50 kg", UnitOfMeasureID: sak}},
	}

	response, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Items: []GoodsReceiptItemRequest{{PurchaseOrderItemID: semen.ID, RejectedQuantity: 4}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.ledger.recorded) != 0 {
		t.Fatalf("rejected goods must not enter stock, got %+v", repository.ledger.recorded)
	}
	if stored := storedItemFor(t, repository, semen); stored.StockMovementID != nil || stored.StockSkipReason != "" {
		t.Fatalf("got stored %+v", stored)
	}
	if item := responseItemFor(t, response, semen); item.StockStatus != stockNone || item.StockPosted || len(response.UnpostedLines) != 0 {
		t.Fatalf("got item %+v unposted %+v", item, response.UnpostedLines)
	}
}

func TestReceiptSkipsUnitMismatchAndReportsIt(t *testing.T) {
	sak, kilogram := uuid.New(), uuid.New()
	semen, pasir := orderLineOf(10, uuid.New(), sak), orderLineOf(5, uuid.New(), sak)
	repository := &fakeRepository{
		order:      PurchaseOrder{Status: orderSent},
		orderItems: []PurchaseOrderItem{semen, pasir},
		materials: []MaterialStockUnit{
			{ID: semen.MaterialID, Name: "Semen Portland 50 kg", UnitOfMeasureID: kilogram},
			{ID: pasir.MaterialID, Name: "Pasir cor", UnitOfMeasureID: sak},
		},
		unitCodes: map[uuid.UUID]string{sak: "sak", kilogram: "kg"},
	}

	response, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
		Number: "UJI-BPB-002",
		Items: []GoodsReceiptItemRequest{
			{PurchaseOrderItemID: semen.ID, AcceptedQuantity: 10},
			{PurchaseOrderItemID: pasir.ID, AcceptedQuantity: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.ledger.recorded) != 1 || repository.ledger.recorded[0].MaterialID != pasir.MaterialID {
		t.Fatalf("only the matching line may be posted, got %+v", repository.ledger.recorded)
	}

	stored := storedItemFor(t, repository, semen)
	if stored.StockMovementID != nil || stored.StockSkipReason != stockSkipUnitMismatch {
		t.Fatalf("got stored %+v", stored)
	}
	if item := responseItemFor(t, response, semen); item.StockStatus != stockSkipped || item.StockPosted {
		t.Fatalf("got item %+v", item)
	}
	if len(response.UnpostedLines) != 1 {
		t.Fatalf("got unposted %+v", response.UnpostedLines)
	}
	unposted := response.UnpostedLines[0]
	if unposted.LineID != stored.ID || unposted.MaterialID != semen.MaterialID || unposted.MaterialName != "Semen Portland 50 kg" {
		t.Fatalf("got unposted %+v", unposted)
	}
	if !strings.Contains(unposted.Reason, "(sak)") || !strings.Contains(unposted.Reason, "(kg)") {
		t.Fatalf("reason must name both units, got %q", unposted.Reason)
	}
}

func TestReceiptFailsWhenStockPostingFails(t *testing.T) {
	cases := []struct {
		name    string
		cause   error
		status  int
		message string
	}{
		{"rejected by inventory", apperror.Unprocessable("stock_movement_reference_not_found", "Material atau gudang tidak ditemukan"), http.StatusUnprocessableEntity, "Material atau gudang tidak ditemukan"},
		{"database failure", errors.New("connection reset by peer"), http.StatusInternalServerError, "Terjadi kesalahan pada server"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sak := uuid.New()
			semen := orderLineOf(10, uuid.New(), sak)
			repository := &fakeRepository{
				order:      PurchaseOrder{Status: orderSent},
				orderItems: []PurchaseOrderItem{semen},
				materials:  []MaterialStockUnit{{ID: semen.MaterialID, Name: "Semen Portland 50 kg", UnitOfMeasureID: sak}},
				ledger:     fakeLedger{err: tc.cause},
			}

			_, err := NewService(repository).RecordGoodsReceipt(context.Background(), GoodsReceiptRequest{
				Items: []GoodsReceiptItemRequest{{PurchaseOrderItemID: semen.ID, AcceptedQuantity: 10}},
			})
			failure := apperror.From(err)
			if failure.Code != receiptStockFailedCode || failure.Status != tc.status {
				t.Fatalf("got %d %s: %v", failure.Status, failure.Code, err)
			}
			if !strings.Contains(failure.Message, "dibatalkan") || !strings.Contains(failure.Message, "Semen Portland 50 kg") || !strings.Contains(failure.Message, tc.message) {
				t.Fatalf("got message %q", failure.Message)
			}
			if !errors.Is(err, tc.cause) {
				t.Fatalf("the cause must stay attached for logging, got %v", err)
			}
			if !repository.rolledBack || repository.savedOrder != nil || len(repository.items) != 0 {
				t.Fatalf("rolled back %v order %+v items %d", repository.rolledBack, repository.savedOrder, len(repository.items))
			}
		})
	}
}

func TestReceiptDetailReportsLegacyLinesWithoutUnpostedReason(t *testing.T) {
	item := GoodsReceiptItem{AcceptedQuantity: 4}
	item.ID = uuid.New()

	detail := newGoodsReceiptDetail(GoodsReceipt{}, []GoodsReceiptLine{{GoodsReceiptItem: item, MaterialName: "Besi beton 10 mm"}})
	if len(detail.Items) != 1 || detail.Items[0].StockStatus != stockLegacy || detail.Items[0].StockPosted || detail.Items[0].UnpostedReason != "" {
		t.Fatalf("got %+v", detail.Items)
	}
	if detail.UnpostedLines == nil || len(detail.UnpostedLines) != 0 {
		t.Fatalf("legacy lines are not skipped lines, got %+v", detail.UnpostedLines)
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
