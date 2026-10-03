package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/site"
)

func scoringServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/health" {
			writer.WriteHeader(status)
			return
		}
		writer.WriteHeader(status)
		_ = json.NewEncoder(writer).Encode(site.ScoreResult{OverallScore: 81, DimensionScores: []site.DimensionScore{{DimensionCode: "infrastruktur", Value: 81}}})
	}))
	t.Cleanup(server.Close)
	return server
}

func TestStrictScoreClientReturnsGenuineScore(t *testing.T) {
	server := scoringServer(t, http.StatusOK)
	client, err := newStrictScoreClient(config.Score{BaseURL: server.URL + "/", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if !client.reachable(context.Background()) {
		t.Fatal("healthy service reported unreachable")
	}
	result, err := client.Score(context.Background(), site.ScoreInput{Latitude: -6.9, Longitude: 107.6, BuildingProfileCode: "housing"})
	if err != nil || result.OverallScore != 81 {
		t.Fatalf("score = %+v, %v; want overall 81", result, err)
	}
}

func TestStrictScoreClientNeverFallsBackToStub(t *testing.T) {
	server := scoringServer(t, http.StatusServiceUnavailable)
	client, err := newStrictScoreClient(config.Score{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if client.reachable(context.Background()) {
		t.Fatal("failing service reported reachable")
	}
	if _, err := client.Score(context.Background(), site.ScoreInput{BuildingProfileCode: "housing"}); err == nil {
		t.Fatal("expected an error instead of stub scores")
	}
}

func TestStrictScoreClientNeedsAURL(t *testing.T) {
	if _, err := newStrictScoreClient(config.Score{BaseURL: "  "}); err == nil {
		t.Fatal("expected an error for an empty score service url")
	}
}
