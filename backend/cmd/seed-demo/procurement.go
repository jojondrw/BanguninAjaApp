package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/procurement"
)

const (
	vendorBuilding  = "DEMO-V01"
	vendorSteel     = "DEMO-V02"
	vendorConcrete  = "DEMO-V03"
	vendorElectric  = "DEMO-V04"
	vendorFinish    = "DEMO-V05"
	vendorPaint     = "DEMO-V06"
	vendorStructure = "DEMO-V07"
	vendorHousing   = "DEMO-V08"

	orderCement   = "DEMO-PO-001"
	orderConcrete = "DEMO-PO-002"
	orderFinish   = "DEMO-PO-003"
	orderTruss    = "DEMO-PO-004"
	orderElectric = "DEMO-PO-005"
)

type vendorSeed struct {
	code     string
	name     string
	category string
	contact  string
	tax      string
	term     int
	rating   string
	active   bool
}

type lineSeed struct {
	material string
	quantity float64
	price    int64
}

type requestSeed struct {
	number  string
	project string
	day     int
	status  string
	note    string
	lines   []lineSeed
}

type orderSeed struct {
	number  string
	vendor  string
	project string
	request string
	day     int
	dueDays int
	sent    bool
	lines   []lineSeed
}

type receiptLineSeed struct {
	material string
	accepted float64
	rejected float64
}

type receiptSeed struct {
	number    string
	order     string
	warehouse string
	day       int
	condition string
	note      string
	lines     []receiptLineSeed
}

var requestStatusPath = map[string][]string{
	"draft":     nil,
	"submitted": {"submitted"},
	"approved":  {"submitted", "approved"},
	"rejected":  {"submitted", "rejected"},
	"completed": {"submitted", "approved", "completed"},
}

func demoVendors() []vendorSeed {
	return []vendorSeed{
		{vendorBuilding, "CV Sumber Bangunan Jaya", "Material bangunan", "022-7310-445", "01.234.567.8-421.000", 30, "good", true},
		{vendorSteel, "PT Baja Prima Nusantara", "Besi & baja", "021-8990-1270", "02.345.678.9-431.000", 45, "good", true},
		{vendorConcrete, "PT Beton Readymix Parahyangan", "Beton siap pakai", "022-6120-338", "03.456.789.0-441.000", 30, "fair", true},
		{vendorElectric, "CV Listrik Terang Abadi", "Mekanikal & elektrikal", "0231-220-918", "04.567.890.1-426.000", 14, "new", true},
		{vendorFinish, "UD Keramik Mulia Cirebon", "Finishing", "0231-481-733", "05.678.901.2-426.000", 30, "good", true},
		{vendorPaint, "PT Cat Warna Lestari", "Finishing", "021-5544-092", "06.789.012.3-411.000", 30, "poor", false},
		{vendorStructure, "PT Graha Struktur Nusantara", "Subkontraktor struktur", "0231-884-120", "07.890.123.4-426.000", 30, "good", true},
		{vendorHousing, "CV Karya Bangun Priangan", "Kontraktor perumahan", "022-7781-204", "08.901.234.5-424.000", 21, "good", true},
	}
}

func demoRequests() []requestSeed {
	return []requestSeed{
		{"DEMO-PR-001", projectGriyaAsri, -215, "completed", "Kebutuhan pengecoran Blok A", []lineSeed{{materialCement, 400, 0}, {materialRebar10, 300, 0}}},
		{"DEMO-PR-002", projectHospital, -45, "approved", "Pengecoran pelat lantai 3 RSIA", []lineSeed{{materialReadymix, 120, 0}, {materialRebar13, 200, 0}}},
		{"DEMO-PR-003", projectGriyaAsri, -14, "approved", "Finishing unit Blok A tahap akhir", []lineSeed{{materialTile, 600, 0}, {materialPaint, 400, 0}}},
		{"DEMO-PR-004", projectRuko, -5, "submitted", "Material contoh untuk ruko percontohan", []lineSeed{{materialLightBrick, 80, 0}, {materialTruss, 300, 0}}},
		{"DEMO-PR-005", projectGriyaAsri, -1, "draft", "Tambahan pasir & pipa untuk Blok B", []lineSeed{{materialSand, 40, 0}, {materialPipe, 30, 0}}},
		{"DEMO-PR-006", projectHospital, -20, "rejected", "Ditolak: spesifikasi cat belum sesuai standar ruang rawat", []lineSeed{{materialPaint, 500, 0}}},
	}
}

func demoOrders() []orderSeed {
	return []orderSeed{
		{orderCement, vendorBuilding, projectGriyaAsri, "DEMO-PR-001", -210, 30, true, []lineSeed{{materialCement, 400, 72_000}, {materialRebar10, 300, 98_000}}},
		{orderConcrete, vendorConcrete, projectHospital, "DEMO-PR-002", -40, 30, true, []lineSeed{{materialReadymix, 120, 1_150_000}}},
		{orderFinish, vendorFinish, projectGriyaAsri, "DEMO-PR-003", -10, 21, true, []lineSeed{{materialTile, 600, 68_000}, {materialPaint, 400, 38_000}}},
		{orderTruss, vendorSteel, projectRuko, "", -3, 45, false, []lineSeed{{materialTruss, 300, 92_000}}},
		{orderElectric, vendorElectric, projectHospital, "", -25, 14, true, []lineSeed{{materialCable, 1_000, 9_500}, {materialPipe, 40, 115_000}}},
	}
}

func demoReceipts() []receiptSeed {
	return []receiptSeed{
		{"DEMO-GR-001", orderCement, warehouseGriya, -200, "good", "Pengiriman pertama, lengkap", []receiptLineSeed{{materialCement, 250, 0}, {materialRebar10, 300, 0}}},
		{"DEMO-GR-002", orderCement, warehouseGriya, -185, "good", "Sisa semen diterima", []receiptLineSeed{{materialCement, 150, 0}}},
		{"DEMO-GR-003", orderConcrete, warehouseHospital, -20, "good", "Pengecoran tahap 1 dari 2", []receiptLineSeed{{materialReadymix, 60, 0}}},
		{"DEMO-GR-004", orderElectric, warehouseCentral, -18, "partially_damaged", "50 m kabel terkelupas, dikembalikan ke vendor", []receiptLineSeed{{materialCable, 950, 50}, {materialPipe, 40, 0}}},
		{"DEMO-GR-005", orderElectric, warehouseCentral, -9, "good", "Penggantian kabel yang rusak", []receiptLineSeed{{materialCable, 50, 0}}},
	}
}

type inventoryLedger struct {
	ledger inventory.StockLedger
}

func newInventoryLedger(db *gorm.DB) procurement.StockLedger {
	return inventoryLedger{ledger: inventory.NewStockLedger(inventory.NewRepository(db))}
}

func (l inventoryLedger) RecordIncoming(ctx context.Context, stock procurement.IncomingStock) (uuid.UUID, error) {
	return l.ledger.RecordIncoming(ctx, inventory.IncomingStock{
		Date: stock.Date, MaterialID: stock.MaterialID, WarehouseID: stock.WarehouseID,
		Quantity: stock.Quantity, Reference: stock.Reference,
	})
}

func (s *seeder) seedProcurement() error {
	steps := []func() error{s.seedVendors, s.seedRequests, s.seedOrders, s.seedReceipts}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedVendors() error {
	for _, seed := range demoVendors() {
		active := seed.active
		id, _, err := s.ensure("vendor", map[string]any{"code": seed.code}, func() (uuid.UUID, error) {
			response, err := s.services.procurement.CreateVendor(s.ctx, procurement.VendorRequest{
				Code: seed.code, Name: seed.name, Category: seed.category, Contact: seed.contact,
				TaxNumber: seed.tax, PaymentTermDays: seed.term, Rating: seed.rating, Active: &active,
			})
			return response.ID, err
		})
		if err != nil {
			return err
		}
		s.remember("vendor", seed.code, id)
	}
	return nil
}

func (s *seeder) seedRequests() error {
	for _, seed := range demoRequests() {
		if err := s.seedRequest(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedRequest(seed requestSeed) error {
	projectID, err := s.known("project", seed.project)
	if err != nil {
		return err
	}
	items, err := s.requestItems(seed.lines)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("purchase_request", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		response, err := s.services.procurement.CreatePurchaseRequest(s.ctx, s.owner, procurement.PurchaseRequestRequest{
			Number: seed.number, ProjectID: projectID, Date: s.day(seed.day), Note: seed.note, Items: items,
		})
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("purchase_request", seed.number, id)
	if !created {
		return nil
	}
	for _, next := range requestStatusPath[seed.status] {
		if _, err := s.services.procurement.UpdatePurchaseRequestStatus(s.ctx, id, procurement.PurchaseRequestStatusRequest{Status: next}); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) requestItems(lines []lineSeed) ([]procurement.PurchaseRequestItemRequest, error) {
	items := make([]procurement.PurchaseRequestItemRequest, 0, len(lines))
	for _, line := range lines {
		materialID, unitID, err := s.materialWithUnit(line.material)
		if err != nil {
			return nil, err
		}
		items = append(items, procurement.PurchaseRequestItemRequest{MaterialID: materialID, Quantity: line.quantity, UnitOfMeasureID: unitID})
	}
	return items, nil
}

func (s *seeder) orderItems(lines []lineSeed) ([]procurement.PurchaseOrderItemRequest, error) {
	items := make([]procurement.PurchaseOrderItemRequest, 0, len(lines))
	for _, line := range lines {
		materialID, unitID, err := s.materialWithUnit(line.material)
		if err != nil {
			return nil, err
		}
		items = append(items, procurement.PurchaseOrderItemRequest{
			MaterialID: materialID, Quantity: line.quantity, UnitOfMeasureID: unitID, UnitPrice: line.price,
		})
	}
	return items, nil
}

func (s *seeder) materialWithUnit(code string) (uuid.UUID, uuid.UUID, error) {
	materialID, err := s.known("material", code)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	unitID, err := s.known("material_unit", code)
	return materialID, unitID, err
}

func (s *seeder) seedOrders() error {
	for _, seed := range demoOrders() {
		if err := s.seedOrder(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedOrder(seed orderSeed) error {
	request, err := s.orderRequest(seed)
	if err != nil {
		return err
	}
	id, created, err := s.ensure("purchase_order", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		response, err := s.services.procurement.CreatePurchaseOrder(s.ctx, request)
		return response.ID, err
	})
	if err != nil {
		return err
	}
	s.remember("purchase_order", seed.number, id)
	if !created || !seed.sent {
		return nil
	}
	_, err = s.services.procurement.UpdatePurchaseOrderStatus(s.ctx, id, procurement.PurchaseOrderStatusRequest{Status: "sent"})
	return err
}

func (s *seeder) orderRequest(seed orderSeed) (procurement.PurchaseOrderRequest, error) {
	vendorID, err := s.known("vendor", seed.vendor)
	if err != nil {
		return procurement.PurchaseOrderRequest{}, err
	}
	projectID, err := s.known("project", seed.project)
	if err != nil {
		return procurement.PurchaseOrderRequest{}, err
	}
	requestID, err := s.knownPointer("purchase_request", seed.request)
	if err != nil {
		return procurement.PurchaseOrderRequest{}, err
	}
	items, err := s.orderItems(seed.lines)
	if err != nil {
		return procurement.PurchaseOrderRequest{}, err
	}
	return procurement.PurchaseOrderRequest{
		Number: seed.number, VendorID: vendorID, ProjectID: projectID, PurchaseRequestID: requestID,
		Date: s.day(seed.day), DueDate: s.dayPointer(seed.day + seed.dueDays), Items: items,
	}, nil
}

func (s *seeder) seedReceipts() error {
	for _, seed := range demoReceipts() {
		if err := s.seedReceipt(seed); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedReceipt(seed receiptSeed) error {
	orderID, err := s.known("purchase_order", seed.order)
	if err != nil {
		return err
	}
	warehouseID, err := s.known("warehouse", seed.warehouse)
	if err != nil {
		return err
	}
	_, _, err = s.ensure("goods_receipt", map[string]any{"number": seed.number}, func() (uuid.UUID, error) {
		items, err := s.receiptItems(orderID, seed.lines)
		if err != nil {
			return uuid.Nil, err
		}
		response, err := s.services.procurement.RecordGoodsReceipt(s.ctx, procurement.GoodsReceiptRequest{
			Number: seed.number, PurchaseOrderID: orderID, WarehouseID: warehouseID, Date: s.day(seed.day),
			Condition: seed.condition, Note: seed.note, Items: items,
		})
		return response.ID, err
	})
	return err
}

func (s *seeder) receiptItems(orderID uuid.UUID, lines []receiptLineSeed) ([]procurement.GoodsReceiptItemRequest, error) {
	order, err := s.services.procurement.GetPurchaseOrder(s.ctx, orderID)
	if err != nil {
		return nil, err
	}
	itemByMaterial := map[uuid.UUID]uuid.UUID{}
	for _, item := range order.Items {
		itemByMaterial[item.MaterialID] = item.ID
	}
	items := make([]procurement.GoodsReceiptItemRequest, 0, len(lines))
	for _, line := range lines {
		materialID, err := s.known("material", line.material)
		if err != nil {
			return nil, err
		}
		itemID, found := itemByMaterial[materialID]
		if !found {
			return nil, fmt.Errorf("purchase order %s has no line for %s", orderID, line.material)
		}
		items = append(items, procurement.GoodsReceiptItemRequest{PurchaseOrderItemID: itemID, AcceptedQuantity: line.accepted, RejectedQuantity: line.rejected})
	}
	return items, nil
}
