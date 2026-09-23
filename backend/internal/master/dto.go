package master

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type RegionRequest struct {
	Code     string     `json:"code" binding:"required,max=20"`
	Name     string     `json:"name" binding:"required,max=120"`
	Type     string     `json:"type" binding:"required,oneof=province city regency district"`
	ParentID *uuid.UUID `json:"parentId"`
}

type RegionQuery struct {
	pagination.Query
	Search   string     `form:"search" binding:"omitempty,max=120"`
	Type     string     `form:"type" binding:"omitempty,oneof=province city regency district"`
	ParentID *uuid.UUID `form:"parentId,parser=encoding.TextUnmarshaler"`
}

type RegionResponse struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	ParentID  *uuid.UUID `json:"parentId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type UnitOfMeasureRequest struct {
	Code string `json:"code" binding:"required,max=20"`
	Name string `json:"name" binding:"required,max=60"`
}

type UnitOfMeasureQuery struct {
	pagination.Query
	Search string `form:"search" binding:"omitempty,max=60"`
}

type UnitOfMeasureResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AccountRequest struct {
	Code     string     `json:"code" binding:"required,max=20"`
	Name     string     `json:"name" binding:"required,max=120"`
	Type     string     `json:"type" binding:"required,oneof=asset liability equity revenue expense"`
	ParentID *uuid.UUID `json:"parentId"`
}

type AccountQuery struct {
	pagination.Query
	Search   string     `form:"search" binding:"omitempty,max=120"`
	Type     string     `form:"type" binding:"omitempty,oneof=asset liability equity revenue expense"`
	ParentID *uuid.UUID `form:"parentId,parser=encoding.TextUnmarshaler"`
}

type AccountResponse struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	ParentID  *uuid.UUID `json:"parentId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func newRegionResponse(region Region) RegionResponse {
	return RegionResponse{
		ID:        region.ID,
		Code:      region.Code,
		Name:      region.Name,
		Type:      region.Type,
		ParentID:  region.ParentID,
		CreatedAt: region.CreatedAt,
		UpdatedAt: region.UpdatedAt,
	}
}

func newUnitOfMeasureResponse(unit UnitOfMeasure) UnitOfMeasureResponse {
	return UnitOfMeasureResponse{
		ID:        unit.ID,
		Code:      unit.Code,
		Name:      unit.Name,
		CreatedAt: unit.CreatedAt,
		UpdatedAt: unit.UpdatedAt,
	}
}

func newAccountResponse(account Account) AccountResponse {
	return AccountResponse{
		ID:        account.ID,
		Code:      account.Code,
		Name:      account.Name,
		Type:      account.Type,
		ParentID:  account.ParentID,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}
