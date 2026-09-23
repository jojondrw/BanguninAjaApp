package scoring

import (
	"time"

	"github.com/google/uuid"
)

type DimensionRequest struct {
	Code        string `json:"code" binding:"required,max=30"`
	Name        string `json:"name" binding:"required,max=120"`
	Description string `json:"description" binding:"max=2000"`
	SortOrder   int    `json:"sortOrder" binding:"required,min=1"`
}

type DimensionResponse struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type BuildingProfileRequest struct {
	Code string `json:"code" binding:"required,max=30"`
	Name string `json:"name" binding:"required,max=120"`
}

type BuildingProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BuildingProfileDetailResponse struct {
	BuildingProfileResponse
	Weights []WeightResponse `json:"weights"`
}

type WeightItemRequest struct {
	DimensionID uuid.UUID `json:"dimensionId" binding:"required"`
	Percent     int       `json:"percent" binding:"min=0,max=100"`
}

type WeightsRequest struct {
	Weights []WeightItemRequest `json:"weights" binding:"required,min=1,dive"`
}

type WeightResponse struct {
	DimensionID uuid.UUID `json:"dimensionId"`
	Percent     int       `json:"percent"`
}

func newDimensionResponse(dimension Dimension) DimensionResponse {
	return DimensionResponse{
		ID:          dimension.ID,
		Code:        dimension.Code,
		Name:        dimension.Name,
		Description: dimension.Description,
		SortOrder:   dimension.SortOrder,
		CreatedAt:   dimension.CreatedAt,
		UpdatedAt:   dimension.UpdatedAt,
	}
}

func newBuildingProfileResponse(profile BuildingProfile) BuildingProfileResponse {
	return BuildingProfileResponse{
		ID:        profile.ID,
		Code:      profile.Code,
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}
}

func newWeightResponse(weight Weight) WeightResponse {
	return WeightResponse{DimensionID: weight.DimensionID, Percent: weight.Percent}
}
