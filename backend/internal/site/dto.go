package site

import (
	"github.com/google/uuid"
)

// EvaluateRequest is the public request body for POST /api/site/evaluate.
// Field names follow the T1 contract (docs/requirements/api/site-evaluate.md §2.1).
type EvaluateRequest struct {
	ProjectID         *uuid.UUID `json:"project_id"`
	Latitude          *float64   `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude         *float64   `json:"longitude" binding:"required,min=-180,max=180"`
	BuildingProfileID uuid.UUID  `json:"building_profile_id" binding:"required"`
	Name              string     `json:"name" binding:"required,max=160"`
}

// EvaluateResponse is the public response `data` object (§2.2).
type EvaluateResponse struct {
	SavedLocationID uuid.UUID    `json:"saved_location_id"`
	Predictive      Predictive   `json:"predictive"`
	Descriptive     *Descriptive `json:"descriptive"`
}

// Predictive is passed through from the internal /score response (§1.2), 5 layers only.
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

// Descriptive is the reserved block (§2.3), owned by Ian's T18. Until then the
// endpoint returns descriptive: null. The shape is defined here so the frontend
// can code against it defensively, but this slice never populates `news`.
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
