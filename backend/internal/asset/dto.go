package asset

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type AssetRequest struct {
	Code                    string     `json:"code" binding:"required,max=20"`
	Name                    string     `json:"name" binding:"required,max=160"`
	Category                string     `json:"category" binding:"required,max=60"`
	AcquisitionDate         *time.Time `json:"acquisitionDate"`
	AcquisitionValue        int64      `json:"acquisitionValue" binding:"min=0"`
	AccumulatedDepreciation int64      `json:"accumulatedDepreciation" binding:"min=0"`
	UsefulLifeMonths        int        `json:"usefulLifeMonths" binding:"min=0"`
	Status                  string     `json:"status" binding:"required,oneof=in_use maintenance for_sale disposed"`
}

type AssetQuery struct {
	pagination.Query
	Search   string `form:"search" binding:"omitempty,max=160"`
	Category string `form:"category" binding:"omitempty,max=60"`
	Status   string `form:"status" binding:"omitempty,oneof=in_use maintenance for_sale disposed"`
}

type AssetResponse struct {
	ID                      uuid.UUID  `json:"id"`
	Code                    string     `json:"code"`
	Name                    string     `json:"name"`
	Category                string     `json:"category"`
	AcquisitionDate         *time.Time `json:"acquisitionDate"`
	AcquisitionValue        int64      `json:"acquisitionValue"`
	AccumulatedDepreciation int64      `json:"accumulatedDepreciation"`
	BookValue               int64      `json:"bookValue"`
	UsefulLifeMonths        int        `json:"usefulLifeMonths"`
	MonthlyDepreciation     int64      `json:"monthlyDepreciation"`
	Status                  string     `json:"status"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

type AssetPage struct {
	pagination.Page[AssetResponse]
	TotalAcquisitionValue        int64 `json:"totalAcquisitionValue"`
	TotalAccumulatedDepreciation int64 `json:"totalAccumulatedDepreciation"`
	TotalBookValue               int64 `json:"totalBookValue"`
}

type EquipmentRequest struct {
	Code            string     `json:"code" binding:"required,max=20"`
	Name            string     `json:"name" binding:"required,max=160"`
	AssetID         *uuid.UUID `json:"assetId"`
	ProjectID       *uuid.UUID `json:"projectId"`
	OperatingHours  int        `json:"operatingHours" binding:"min=0"`
	NextServiceDate *time.Time `json:"nextServiceDate"`
	Status          string     `json:"status" binding:"required,oneof=operating idle maintenance broken"`
}

type EquipmentQuery struct {
	pagination.Query
	Search           string     `form:"search" binding:"omitempty,max=160"`
	Status           string     `form:"status" binding:"omitempty,oneof=operating idle maintenance broken"`
	ProjectID        *uuid.UUID `form:"projectId,parser=encoding.TextUnmarshaler"`
	AssetID          *uuid.UUID `form:"assetId,parser=encoding.TextUnmarshaler"`
	ServiceDueBefore *time.Time `form:"serviceDueBefore" time_format:"2006-01-02"`
}

type EquipmentResponse struct {
	ID              uuid.UUID  `json:"id"`
	Code            string     `json:"code"`
	Name            string     `json:"name"`
	AssetID         *uuid.UUID `json:"assetId"`
	ProjectID       *uuid.UUID `json:"projectId"`
	OperatingHours  int        `json:"operatingHours"`
	NextServiceDate *time.Time `json:"nextServiceDate"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func newAssetResponse(asset Asset) AssetResponse {
	return AssetResponse{
		ID:                      asset.ID,
		Code:                    asset.Code,
		Name:                    asset.Name,
		Category:                asset.Category,
		AcquisitionDate:         asset.AcquisitionDate,
		AcquisitionValue:        asset.AcquisitionValue,
		AccumulatedDepreciation: asset.AccumulatedDepreciation,
		BookValue:               bookValue(asset.AcquisitionValue, asset.AccumulatedDepreciation),
		UsefulLifeMonths:        asset.UsefulLifeMonths,
		MonthlyDepreciation:     monthlyDepreciation(asset.AcquisitionValue, asset.UsefulLifeMonths),
		Status:                  asset.Status,
		CreatedAt:               asset.CreatedAt,
		UpdatedAt:               asset.UpdatedAt,
	}
}

func newEquipmentResponse(equipment Equipment) EquipmentResponse {
	return EquipmentResponse{
		ID:              equipment.ID,
		Code:            equipment.Code,
		Name:            equipment.Name,
		AssetID:         equipment.AssetID,
		ProjectID:       equipment.ProjectID,
		OperatingHours:  equipment.OperatingHours,
		NextServiceDate: equipment.NextServiceDate,
		Status:          equipment.Status,
		CreatedAt:       equipment.CreatedAt,
		UpdatedAt:       equipment.UpdatedAt,
	}
}
