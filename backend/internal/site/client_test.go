package site

import (
	"context"
	"errors"
	"testing"
)

func TestStubScoreClientProducesFiveScoredDimensions(t *testing.T) {
	client := stubScoreClient{}

	result, err := client.Score(context.Background(), ScoreInput{
		Latitude:            -6.914744,
		Longitude:           107.609810,
		BuildingProfileCode: "housing",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.DimensionScores) != len(scoredDimensions) {
		t.Fatalf("expected %d dimensions, got %d", len(scoredDimensions), len(result.DimensionScores))
	}

	for _, dimension := range result.DimensionScores {
		if dimension.Value < 0 || dimension.Value > 100 {
			t.Errorf("dimension %s value %d out of range 0..100", dimension.DimensionCode, dimension.Value)
		}
		if dimension.Explanation == "" {
			t.Errorf("dimension %s missing explanation", dimension.DimensionCode)
		}
	}

	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("overall score %d out of range 0..100", result.OverallScore)
	}

	// risk_flags must never be nil per the contract (§1.2).
	if result.RiskFlags == nil {
		t.Error("risk_flags must be a non-nil slice")
	}
}

func TestStubScoreClientIsDeterministic(t *testing.T) {
	client := stubScoreClient{}
	input := ScoreInput{Latitude: -6.2, Longitude: 106.8, BuildingProfileCode: "mall"}

	first, err := client.Score(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := client.Score(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.OverallScore != second.OverallScore {
		t.Errorf("same input gave different scores: %d vs %d", first.OverallScore, second.OverallScore)
	}
}

func TestStubScoreClientRejectsEmptyProfile(t *testing.T) {
	client := stubScoreClient{}

	_, err := client.Score(context.Background(), ScoreInput{Latitude: 1, Longitude: 1})
	if !errors.Is(err, errUnknownProfile) {
		t.Fatalf("expected errUnknownProfile, got %v", err)
	}
}

func TestNewScoreClientFallsBackToStubWhenUnconfigured(t *testing.T) {
	client := NewScoreClient("", 0)

	if _, ok := client.(stubScoreClient); !ok {
		t.Fatalf("expected stubScoreClient when no base URL configured, got %T", client)
	}
}
