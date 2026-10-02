package site

import (
	"github.com/google/uuid"
)

type EvaluateRequest struct {
	ProjectID         *uuid.UUID `json:"project_id"`
	Latitude          *float64   `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude         *float64   `json:"longitude" binding:"required,min=-180,max=180"`
	BuildingProfileID uuid.UUID  `json:"building_profile_id" binding:"required"`
	Name              string     `json:"name" binding:"required,max=160"`
}

type EvaluateResponse struct {
	SavedLocationID uuid.UUID    `json:"saved_location_id"`
	Predictive      Predictive   `json:"predictive"`
	Descriptive     *Descriptive `json:"descriptive"`
}

type Predictive struct {
	OverallScore    int              `json:"overall_score"`
	DimensionScores []DimensionScore `json:"dimension_scores"`
	RiskFlags       []RiskFlag       `json:"risk_flags"`
}

type DimensionScore struct {
	DimensionCode string `json:"dimension_code"`
	Value         int    `json:"value"`
	Explanation   string `json:"explanation"`
}

type RiskFlag struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type Descriptive struct {
	Regulasi *Regulasi `json:"regulasi"`
	News     []News    `json:"news"`
}

type Regulasi struct {
	KDB         float64 `json:"kdb"`
	KLB         float64 `json:"klb"`
	Zona        string  `json:"zona"`
	IsSimulated bool    `json:"is_simulated"`
}

type News struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Source      string `json:"source"`
	PublishedAt string `json:"published_at"`
}
