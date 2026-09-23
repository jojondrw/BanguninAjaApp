package inventory

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type MaterialRequest struct {
	Code            string    `json:"code" binding:"required,max=20"`
	Name            string    `json:"name" binding:"required,max=160"`
	Category        string    `json:"category" binding:"max=60"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId" binding:"required"`
	MinimumStock    float64   `json:"minimumStock" binding:"min=0"`
	LastPrice       int64     `json:"lastPrice" binding:"min=0"`
}

type MaterialQuery struct {
	pagination.Query
	Search   string `form:"search" binding:"omitempty,max=160"`
	Category string `form:"category" binding:"omitempty,max=60"`
}

type MaterialResponse struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Category        string    `json:"category"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId"`
	MinimumStock    float64   `json:"minimumStock"`
	LastPrice       int64     `json:"lastPrice"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type LowStockResponse struct {
	MaterialID      uuid.UUID `json:"materialId"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId"`
	MinimumStock    float64   `json:"minimumStock"`
	TotalQuantity   float64   `json:"totalQuantity"`
}

type WarehouseRequest struct {
	Code      string     `json:"code" binding:"required,max=20"`
	Name      string     `json:"name" binding:"required,max=120"`
	ProjectID *uuid.UUID `json:"projectId"`
}

type WarehouseQuery struct {
	pagination.Query
	Search    string     `form:"search" binding:"omitempty,max=120"`
	ProjectID *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
}

type WarehouseResponse struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	ProjectID *uuid.UUID `json:"projectId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type StockQuery struct {
	pagination.Query
	MaterialID  *uuid.UUID `form:"materialId,parser=encoding.TextUnmarshaler"`
	WarehouseID *uuid.UUID `form:"warehouseId,parser=encoding.TextUnmarshaler"`
}

type StockResponse struct {
	ID            uuid.UUID `json:"id"`
	MaterialID    uuid.UUID `json:"materialId"`
	MaterialCode  string    `json:"materialCode"`
	MaterialName  string    `json:"materialName"`
	WarehouseID   uuid.UUID `json:"warehouseId"`
	WarehouseCode string    `json:"warehouseCode"`
	WarehouseName string    `json:"warehouseName"`
	Quantity      float64   `json:"quantity"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type StockMovementRequest struct {
	Date              time.Time  `json:"date" binding:"required"`
	Type              string     `json:"type" binding:"required,oneof=in out transfer adjustment"`
	MaterialID        uuid.UUID  `json:"materialId" binding:"required"`
	Quantity          float64    `json:"quantity" binding:"required,gt=0"`
	SourceWarehouseID *uuid.UUID `json:"sourceWarehouseId"`
	TargetWarehouseID *uuid.UUID `json:"targetWarehouseId"`
	Reference         string     `json:"reference" binding:"max=60"`
}

type StockMovementQuery struct {
	pagination.Query
	MaterialID  *uuid.UUID `form:"materialId,parser=encoding.TextUnmarshaler"`
	WarehouseID *uuid.UUID `form:"warehouseId,parser=encoding.TextUnmarshaler"`
	Type        string     `form:"type" binding:"omitempty,oneof=in out transfer adjustment"`
	DateFrom    *time.Time `form:"dateFrom" time_format:"2006-01-02"`
	DateTo      *time.Time `form:"dateTo" time_format:"2006-01-02"`
}

type StockMovementResponse struct {
	ID                uuid.UUID  `json:"id"`
	Date              time.Time  `json:"date"`
	Type              string     `json:"type"`
	MaterialID        uuid.UUID  `json:"materialId"`
	Quantity          float64    `json:"quantity"`
	SourceWarehouseID *uuid.UUID `json:"sourceWarehouseId"`
	TargetWarehouseID *uuid.UUID `json:"targetWarehouseId"`
	Reference         string     `json:"reference"`
	CreatedAt         time.Time  `json:"createdAt"`
}

func newMaterialResponse(material Material) MaterialResponse {
	return MaterialResponse{
		ID:              material.ID,
		Code:            material.Code,
		Name:            material.Name,
		Category:        material.Category,
		UnitOfMeasureID: material.UnitOfMeasureID,
		MinimumStock:    material.MinimumStock,
		LastPrice:       material.LastPrice,
		CreatedAt:       material.CreatedAt,
		UpdatedAt:       material.UpdatedAt,
	}
}

func newLowStockResponse(level StockLevel) LowStockResponse {
	return LowStockResponse{
		MaterialID:      level.MaterialID,
		Code:            level.Code,
		Name:            level.Name,
		UnitOfMeasureID: level.UnitOfMeasureID,
		MinimumStock:    level.MinimumStock,
		TotalQuantity:   level.TotalQuantity,
	}
}

func newWarehouseResponse(warehouse Warehouse) WarehouseResponse {
	return WarehouseResponse{
		ID:        warehouse.ID,
		Code:      warehouse.Code,
		Name:      warehouse.Name,
		ProjectID: warehouse.ProjectID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}
}

func newStockResponse(row StockRow) StockResponse {
	return StockResponse{
		ID:            row.ID,
		MaterialID:    row.MaterialID,
		MaterialCode:  row.MaterialCode,
		MaterialName:  row.MaterialName,
		WarehouseID:   row.WarehouseID,
		WarehouseCode: row.WarehouseCode,
		WarehouseName: row.WarehouseName,
		Quantity:      row.Quantity,
		UpdatedAt:     row.UpdatedAt,
	}
}

func newStockMovementResponse(movement StockMovement) StockMovementResponse {
	return StockMovementResponse{
		ID:                movement.ID,
		Date:              movement.Date,
		Type:              movement.Type,
		MaterialID:        movement.MaterialID,
		Quantity:          movement.Quantity,
		SourceWarehouseID: movement.SourceWarehouseID,
		TargetWarehouseID: movement.TargetWarehouseID,
		Reference:         movement.Reference,
		CreatedAt:         movement.CreatedAt,
	}
}
