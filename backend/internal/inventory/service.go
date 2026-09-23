package inventory

import (
	"context"
	"math"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	movementIn         = "in"
	movementOut        = "out"
	movementTransfer   = "transfer"
	movementAdjustment = "adjustment"

	quantityPrecision = 100
)

var (
	errMaterialNotFound      = apperror.NotFound("material_not_found", "Material tidak ditemukan")
	errMaterialCodeUsed      = apperror.Conflict("material_code_used", "Kode material sudah dipakai")
	errMaterialInUse         = apperror.Conflict("material_in_use", "Material masih dipakai oleh data lain")
	errUnitOfMeasureNotFound = apperror.Unprocessable("unit_of_measure_not_found", "Satuan tidak ditemukan")

	errWarehouseNotFound        = apperror.NotFound("warehouse_not_found", "Gudang tidak ditemukan")
	errWarehouseCodeUsed        = apperror.Conflict("warehouse_code_used", "Kode gudang sudah dipakai")
	errWarehouseInUse           = apperror.Conflict("warehouse_in_use", "Gudang masih dipakai oleh data lain")
	errWarehouseProjectNotFound = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")

	errMovementNotFound   = apperror.NotFound("stock_movement_not_found", "Mutasi stok tidak ditemukan")
	errMovementDirection  = apperror.Unprocessable("stock_movement_direction_invalid", "Gudang asal dan tujuan tidak sesuai dengan jenis mutasi")
	errMovementReference  = apperror.Unprocessable("stock_movement_reference_not_found", "Material atau gudang tidak ditemukan")
	errMovementQuantity   = apperror.Unprocessable("stock_movement_quantity_invalid", "Jumlah mutasi harus lebih dari nol")
	errInsufficientStock  = apperror.Unprocessable("insufficient_stock", "Stok di gudang asal tidak mencukupi")
	errMovementSameTarget = apperror.Unprocessable("stock_movement_same_warehouse", "Gudang asal dan tujuan tidak boleh sama")
)

var (
	materialReadErrors   = database.ErrorMap{NotFound: errMaterialNotFound}
	materialWriteErrors  = database.ErrorMap{NotFound: errMaterialNotFound, Duplicate: errMaterialCodeUsed, Referenced: errUnitOfMeasureNotFound}
	materialDeleteErrors = database.ErrorMap{NotFound: errMaterialNotFound, Referenced: errMaterialInUse}

	warehouseReadErrors   = database.ErrorMap{NotFound: errWarehouseNotFound}
	warehouseWriteErrors  = database.ErrorMap{NotFound: errWarehouseNotFound, Duplicate: errWarehouseCodeUsed, Referenced: errWarehouseProjectNotFound}
	warehouseDeleteErrors = database.ErrorMap{NotFound: errWarehouseNotFound, Referenced: errWarehouseInUse}

	movementReadErrors  = database.ErrorMap{NotFound: errMovementNotFound}
	movementWriteErrors = database.ErrorMap{Referenced: errMovementReference, Invalid: errInsufficientStock}
)

type Service interface {
	ListMaterials(ctx context.Context, query MaterialQuery) (pagination.Page[MaterialResponse], error)
	ListLowStockMaterials(ctx context.Context, query pagination.Query) (pagination.Page[LowStockResponse], error)
	GetMaterial(ctx context.Context, id uuid.UUID) (MaterialResponse, error)
	CreateMaterial(ctx context.Context, request MaterialRequest) (MaterialResponse, error)
	UpdateMaterial(ctx context.Context, id uuid.UUID, request MaterialRequest) (MaterialResponse, error)
	DeleteMaterial(ctx context.Context, id uuid.UUID) error

	ListWarehouses(ctx context.Context, query WarehouseQuery) (pagination.Page[WarehouseResponse], error)
	GetWarehouse(ctx context.Context, id uuid.UUID) (WarehouseResponse, error)
	CreateWarehouse(ctx context.Context, request WarehouseRequest) (WarehouseResponse, error)
	UpdateWarehouse(ctx context.Context, id uuid.UUID, request WarehouseRequest) (WarehouseResponse, error)
	DeleteWarehouse(ctx context.Context, id uuid.UUID) error

	ListStocks(ctx context.Context, query StockQuery) (pagination.Page[StockResponse], error)

	ListStockMovements(ctx context.Context, query StockMovementQuery) (pagination.Page[StockMovementResponse], error)
	GetStockMovement(ctx context.Context, id uuid.UUID) (StockMovementResponse, error)
	RecordStockMovement(ctx context.Context, request StockMovementRequest) (StockMovementResponse, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListMaterials(ctx context.Context, query MaterialQuery) (pagination.Page[MaterialResponse], error) {
	materials, total, err := s.repository.ListMaterials(ctx, MaterialFilter{
		Search:   query.Search,
		Category: strings.TrimSpace(query.Category),
		Offset:   query.Offset(),
		Limit:    query.Size(),
	})
	if err != nil {
		return pagination.Page[MaterialResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(materials, newMaterialResponse), query.Query, total), nil
}

func (s *service) ListLowStockMaterials(ctx context.Context, query pagination.Query) (pagination.Page[LowStockResponse], error) {
	levels, total, err := s.repository.ListLowStockMaterials(ctx, query.Offset(), query.Size())
	if err != nil {
		return pagination.Page[LowStockResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(levels, newLowStockResponse), query, total), nil
}

func (s *service) GetMaterial(ctx context.Context, id uuid.UUID) (MaterialResponse, error) {
	material, err := s.repository.FindMaterial(ctx, id)
	if err != nil {
		return MaterialResponse{}, materialReadErrors.Resolve(err)
	}
	return newMaterialResponse(material), nil
}

func (s *service) CreateMaterial(ctx context.Context, request MaterialRequest) (MaterialResponse, error) {
	var material Material
	applyMaterialRequest(&material, request)
	if err := s.repository.CreateMaterial(ctx, &material); err != nil {
		return MaterialResponse{}, materialWriteErrors.Resolve(err)
	}
	return newMaterialResponse(material), nil
}

func (s *service) UpdateMaterial(ctx context.Context, id uuid.UUID, request MaterialRequest) (MaterialResponse, error) {
	material, err := s.repository.FindMaterial(ctx, id)
	if err != nil {
		return MaterialResponse{}, materialReadErrors.Resolve(err)
	}

	applyMaterialRequest(&material, request)
	if err := s.repository.SaveMaterial(ctx, &material); err != nil {
		return MaterialResponse{}, materialWriteErrors.Resolve(err)
	}
	return newMaterialResponse(material), nil
}

func (s *service) DeleteMaterial(ctx context.Context, id uuid.UUID) error {
	return materialDeleteErrors.Resolve(s.repository.DeleteMaterial(ctx, id))
}

func (s *service) ListWarehouses(ctx context.Context, query WarehouseQuery) (pagination.Page[WarehouseResponse], error) {
	warehouses, total, err := s.repository.ListWarehouses(ctx, WarehouseFilter{
		Search:    query.Search,
		ProjectID: query.ProjectID,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[WarehouseResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(warehouses, newWarehouseResponse), query.Query, total), nil
}

func (s *service) GetWarehouse(ctx context.Context, id uuid.UUID) (WarehouseResponse, error) {
	warehouse, err := s.repository.FindWarehouse(ctx, id)
	if err != nil {
		return WarehouseResponse{}, warehouseReadErrors.Resolve(err)
	}
	return newWarehouseResponse(warehouse), nil
}

func (s *service) CreateWarehouse(ctx context.Context, request WarehouseRequest) (WarehouseResponse, error) {
	var warehouse Warehouse
	applyWarehouseRequest(&warehouse, request)
	if err := s.repository.CreateWarehouse(ctx, &warehouse); err != nil {
		return WarehouseResponse{}, warehouseWriteErrors.Resolve(err)
	}
	return newWarehouseResponse(warehouse), nil
}

func (s *service) UpdateWarehouse(ctx context.Context, id uuid.UUID, request WarehouseRequest) (WarehouseResponse, error) {
	warehouse, err := s.repository.FindWarehouse(ctx, id)
	if err != nil {
		return WarehouseResponse{}, warehouseReadErrors.Resolve(err)
	}

	applyWarehouseRequest(&warehouse, request)
	if err := s.repository.SaveWarehouse(ctx, &warehouse); err != nil {
		return WarehouseResponse{}, warehouseWriteErrors.Resolve(err)
	}
	return newWarehouseResponse(warehouse), nil
}

func (s *service) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	return warehouseDeleteErrors.Resolve(s.repository.DeleteWarehouse(ctx, id))
}

func (s *service) ListStocks(ctx context.Context, query StockQuery) (pagination.Page[StockResponse], error) {
	rows, total, err := s.repository.ListStocks(ctx, StockFilter{
		MaterialID:  query.MaterialID,
		WarehouseID: query.WarehouseID,
		Offset:      query.Offset(),
		Limit:       query.Size(),
	})
	if err != nil {
		return pagination.Page[StockResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newStockResponse), query.Query, total), nil
}

func (s *service) ListStockMovements(ctx context.Context, query StockMovementQuery) (pagination.Page[StockMovementResponse], error) {
	movements, total, err := s.repository.ListStockMovements(ctx, StockMovementFilter{
		MaterialID:  query.MaterialID,
		WarehouseID: query.WarehouseID,
		Type:        query.Type,
		DateFrom:    query.DateFrom,
		DateTo:      query.DateTo,
		Offset:      query.Offset(),
		Limit:       query.Size(),
	})
	if err != nil {
		return pagination.Page[StockMovementResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(movements, newStockMovementResponse), query.Query, total), nil
}

func (s *service) GetStockMovement(ctx context.Context, id uuid.UUID) (StockMovementResponse, error) {
	movement, err := s.repository.FindStockMovement(ctx, id)
	if err != nil {
		return StockMovementResponse{}, movementReadErrors.Resolve(err)
	}
	return newStockMovementResponse(movement), nil
}

func (s *service) RecordStockMovement(ctx context.Context, request StockMovementRequest) (StockMovementResponse, error) {
	movement := newStockMovement(request)
	if err := validateMovement(movement); err != nil {
		return StockMovementResponse{}, err
	}

	err := s.repository.Transaction(ctx, func(repository Repository) error {
		return applyMovement(ctx, repository, &movement)
	})
	if err != nil {
		return StockMovementResponse{}, apperror.From(err)
	}
	return newStockMovementResponse(movement), nil
}

func applyMovement(ctx context.Context, repository Repository, movement *StockMovement) error {
	if err := repository.CreateStockMovement(ctx, movement); err != nil {
		return movementWriteErrors.Resolve(err)
	}
	if err := takeFromSource(ctx, repository, *movement); err != nil {
		return err
	}
	if movement.TargetWarehouseID == nil {
		return nil
	}

	err := repository.IncreaseStock(ctx, movement.MaterialID, *movement.TargetWarehouseID, movement.Quantity)
	return movementWriteErrors.Resolve(err)
}

func takeFromSource(ctx context.Context, repository Repository, movement StockMovement) error {
	if movement.SourceWarehouseID == nil {
		return nil
	}

	decreased, err := repository.DecreaseStock(ctx, movement.MaterialID, *movement.SourceWarehouseID, movement.Quantity)
	if err != nil {
		return movementWriteErrors.Resolve(err)
	}
	if !decreased {
		return errInsufficientStock
	}
	return nil
}

func validateMovement(movement StockMovement) error {
	if movement.Quantity <= 0 {
		return errMovementQuantity
	}
	if !directionMatchesType(movement.Type, movement.SourceWarehouseID != nil, movement.TargetWarehouseID != nil) {
		return errMovementDirection
	}
	if movement.Type == movementTransfer && *movement.SourceWarehouseID == *movement.TargetWarehouseID {
		return errMovementSameTarget
	}
	return nil
}

func directionMatchesType(movementType string, hasSource, hasTarget bool) bool {
	switch movementType {
	case movementIn:
		return hasTarget && !hasSource
	case movementOut:
		return hasSource && !hasTarget
	case movementTransfer:
		return hasSource && hasTarget
	case movementAdjustment:
		return hasSource != hasTarget
	default:
		return false
	}
}

func newStockMovement(request StockMovementRequest) StockMovement {
	return StockMovement{
		Date:              request.Date,
		Type:              request.Type,
		MaterialID:        request.MaterialID,
		Quantity:          roundQuantity(request.Quantity),
		SourceWarehouseID: request.SourceWarehouseID,
		TargetWarehouseID: request.TargetWarehouseID,
		Reference:         strings.TrimSpace(request.Reference),
	}
}

func roundQuantity(quantity float64) float64 {
	return math.Round(quantity*quantityPrecision) / quantityPrecision
}

func applyMaterialRequest(material *Material, request MaterialRequest) {
	material.Code = strings.TrimSpace(request.Code)
	material.Name = strings.TrimSpace(request.Name)
	material.Category = strings.TrimSpace(request.Category)
	material.UnitOfMeasureID = request.UnitOfMeasureID
	material.MinimumStock = roundQuantity(request.MinimumStock)
	material.LastPrice = request.LastPrice
}

func applyWarehouseRequest(warehouse *Warehouse, request WarehouseRequest) {
	warehouse.Code = strings.TrimSpace(request.Code)
	warehouse.Name = strings.TrimSpace(request.Name)
	warehouse.ProjectID = request.ProjectID
}
