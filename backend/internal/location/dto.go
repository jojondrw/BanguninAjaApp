package location

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type DimensionScoreRequest struct {
	DimensionID uuid.UUID `json:"dimensionId" binding:"required"`
	Value       int       `json:"value" binding:"min=0,max=100"`
}

type SavedLocationRequest struct {
	Name              string                  `json:"name" binding:"required,max=160"`
	RegionID          *uuid.UUID              `json:"regionId"`
	BuildingProfileID *uuid.UUID              `json:"buildingProfileId"`
	Latitude          *float64                `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude         *float64                `json:"longitude" binding:"required,min=-180,max=180"`
	AreaSqm           float64                 `json:"areaSqm" binding:"min=0"`
	LandPricePerSqm   int64                   `json:"landPricePerSqm" binding:"min=0"`
	Score             int                     `json:"score" binding:"min=0,max=100"`
	FloodIndex        float64                 `json:"floodIndex" binding:"min=0,max=1"`
	EarthquakeIndex   float64                 `json:"earthquakeIndex" binding:"min=0,max=1"`
	Note              string                  `json:"note" binding:"max=2000"`
	DimensionScores   []DimensionScoreRequest `json:"dimensionScores" binding:"dive"`
}

type SavedLocationQuery struct {
	pagination.Query
	Search            string     `form:"search" binding:"omitempty,max=160"`
	RegionID          *uuid.UUID `form:"regionId,parser=encoding.TextUnmarshaler"`
	BuildingProfileID *uuid.UUID `form:"buildingProfileId,parser=encoding.TextUnmarshaler"`
	MinScore          *int       `form:"minScore" binding:"omitempty,min=0,max=100"`
	MinLongitude      *float64   `form:"minLongitude" binding:"omitempty,min=-180,max=180"`
	MinLatitude       *float64   `form:"minLatitude" binding:"omitempty,min=-90,max=90"`
	MaxLongitude      *float64   `form:"maxLongitude" binding:"omitempty,min=-180,max=180"`
	MaxLatitude       *float64   `form:"maxLatitude" binding:"omitempty,min=-90,max=90"`
}

type DimensionScoreResponse struct {
	DimensionID uuid.UUID `json:"dimensionId"`
	Value       int       `json:"value"`
}

type SavedLocationResponse struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	RegionID          *uuid.UUID `json:"regionId"`
	BuildingProfileID *uuid.UUID `json:"buildingProfileId"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	AreaSqm           float64    `json:"areaSqm"`
	LandPricePerSqm   int64      `json:"landPricePerSqm"`
	Score             int        `json:"score"`
	FloodIndex        float64    `json:"floodIndex"`
	EarthquakeIndex   float64    `json:"earthquakeIndex"`
	Note              string     `json:"note"`
	SavedAt           time.Time  `json:"savedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type SavedLocationDetailResponse struct {
	SavedLocationResponse
	DimensionScores []DimensionScoreResponse `json:"dimensionScores"`
}

type ComparisonRequest struct {
	Name             string      `json:"name" binding:"required,max=160"`
	SavedLocationIDs []uuid.UUID `json:"savedLocationIds" binding:"required,min=2,max=5"`
}

type ComparisonResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	LocationCount int       `json:"locationCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ComparisonItemResponse struct {
	SavedLocationID uuid.UUID                `json:"savedLocationId"`
	SortOrder       int                      `json:"sortOrder"`
	Name            string                   `json:"name"`
	Score           int                      `json:"score"`
	AreaSqm         float64                  `json:"areaSqm"`
	LandPricePerSqm int64                    `json:"landPricePerSqm"`
	FloodIndex      float64                  `json:"floodIndex"`
	EarthquakeIndex float64                  `json:"earthquakeIndex"`
	DimensionScores []DimensionScoreResponse `json:"dimensionScores"`
}

type ComparisonDetailResponse struct {
	ID        uuid.UUID                `json:"id"`
	Name      string                   `json:"name"`
	Items     []ComparisonItemResponse `json:"items"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
}

func newSavedLocationResponse(location SavedLocation) SavedLocationResponse {
	return SavedLocationResponse{
		ID:                location.ID,
		Name:              location.Name,
		RegionID:          location.RegionID,
		BuildingProfileID: location.BuildingProfileID,
		Latitude:          location.Point.Lat,
		Longitude:         location.Point.Lon,
		AreaSqm:           location.AreaSqm,
		LandPricePerSqm:   location.LandPricePerSqm,
		Score:             location.Score,
		FloodIndex:        location.FloodIndex,
		EarthquakeIndex:   location.EarthquakeIndex,
		Note:              location.Note,
		SavedAt:           location.SavedAt,
		CreatedAt:         location.CreatedAt,
		UpdatedAt:         location.UpdatedAt,
	}
}

func newDimensionScoreResponse(score DimensionScore) DimensionScoreResponse {
	return DimensionScoreResponse{DimensionID: score.DimensionID, Value: score.Value}
}

func newComparisonResponse(row ComparisonRow) ComparisonResponse {
	return ComparisonResponse{
		ID:            row.ID,
		Name:          row.Name,
		LocationCount: row.LocationCount,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func newComparisonItemResponse(item ComparisonItem, location SavedLocation, scores []DimensionScoreResponse) ComparisonItemResponse {
	return ComparisonItemResponse{
		SavedLocationID: item.SavedLocationID,
		SortOrder:       item.SortOrder,
		Name:            location.Name,
		Score:           location.Score,
		AreaSqm:         location.AreaSqm,
		LandPricePerSqm: location.LandPricePerSqm,
		FloodIndex:      location.FloodIndex,
		EarthquakeIndex: location.EarthquakeIndex,
		DimensionScores: scores,
	}
}
