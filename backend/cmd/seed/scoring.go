package main

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/scoring"
)

const fullWeight = 100

type scoringSeed struct {
	Dimensions []dimensionSeed           `json:"dimensions"`
	Profiles   []profileSeed             `json:"profiles"`
	Weights    map[string]map[string]int `json:"weights"`
}

type dimensionSeed struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

type profileSeed struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func seedScoring(tx *gorm.DB) error {
	var seed scoringSeed
	if err := readJSON("scoring.json", &seed); err != nil {
		return err
	}

	if err := upsertDimensions(tx, seed.Dimensions); err != nil {
		return err
	}

	if err := upsertProfiles(tx, seed.Profiles); err != nil {
		return err
	}

	return upsertWeights(tx, seed.Weights)
}

func upsertDimensions(tx *gorm.DB, dimensions []dimensionSeed) error {
	for _, item := range dimensions {
		dimension := &scoring.Dimension{
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			SortOrder:   item.SortOrder,
		}
		if err := upsertByCode(tx, dimension, "name", "description", "sort_order"); err != nil {
			return fmt.Errorf("dimension %s: %w", item.Code, err)
		}
	}

	return nil
}

func upsertProfiles(tx *gorm.DB, profiles []profileSeed) error {
	for _, item := range profiles {
		profile := &scoring.BuildingProfile{Code: item.Code, Name: item.Name}
		if err := upsertByCode(tx, profile, "name"); err != nil {
			return fmt.Errorf("building profile %s: %w", item.Code, err)
		}
	}

	return nil
}

func upsertWeights(tx *gorm.DB, weights map[string]map[string]int) error {
	profiles, err := loadCodeIDs(tx, "building_profile")
	if err != nil {
		return err
	}

	dimensions, err := loadCodeIDs(tx, "dimension")
	if err != nil {
		return err
	}

	for profileCode, percents := range weights {
		if err := upsertProfileWeights(tx, profileCode, percents, profiles, dimensions); err != nil {
			return err
		}
	}

	return nil
}

func upsertProfileWeights(tx *gorm.DB, profileCode string, percents map[string]int, profiles, dimensions map[string]uuid.UUID) error {
	profileID, found := profiles[profileCode]
	if !found {
		return fmt.Errorf("weights reference unknown building profile %s", profileCode)
	}

	if total := sumPercents(percents); total != fullWeight {
		return fmt.Errorf("weights for %s sum to %d, expected %d", profileCode, total, fullWeight)
	}

	for dimensionCode, percent := range percents {
		dimensionID, found := dimensions[dimensionCode]
		if !found {
			return fmt.Errorf("weights reference unknown dimension %s", dimensionCode)
		}
		if err := upsertWeight(tx, profileID, dimensionID, percent); err != nil {
			return fmt.Errorf("weight %s/%s: %w", profileCode, dimensionCode, err)
		}
	}

	return nil
}

func upsertWeight(tx *gorm.DB, profileID, dimensionID uuid.UUID, percent int) error {
	weight := &scoring.Weight{BuildingProfileID: profileID, DimensionID: dimensionID, Percent: percent}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "building_profile_id"}, {Name: "dimension_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"percent", "updated_at"}),
	}).Create(weight).Error
}

func sumPercents(percents map[string]int) int {
	total := 0
	for _, percent := range percents {
		total += percent
	}

	return total
}
