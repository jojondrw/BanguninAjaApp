package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/news"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/site"
)

var errScoreServiceMissing = errors.New("score service url is empty")

type strictScoreClient struct {
	baseURL string
	client  *http.Client
}

func newStrictScoreClient(cfg config.Score) (strictScoreClient, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return strictScoreClient{}, errScoreServiceMissing
	}
	return strictScoreClient{baseURL: baseURL, client: &http.Client{Timeout: cfg.Timeout}}, nil
}

func (c strictScoreClient) Score(ctx context.Context, input site.ScoreInput) (site.ScoreResult, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return site.ScoreResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/score", bytes.NewReader(body))
	if err != nil {
		return site.ScoreResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return site.ScoreResult{}, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return site.ScoreResult{}, fmt.Errorf("score service returned status %d", response.StatusCode)
	}
	var result site.ScoreResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return site.ScoreResult{}, err
	}
	return result, nil
}

func (c strictScoreClient) reachable(ctx context.Context) bool {
	probe, cancel := context.WithTimeout(ctx, c.client.Timeout+time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(probe, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return false
	}
	response, err := c.client.Do(request)
	if err != nil {
		return false
	}
	_ = response.Body.Close()
	return response.StatusCode == http.StatusOK
}

type silentRegulation struct{}

func (silentRegulation) GetByPoint(context.Context, float64, float64) (*regulation.Response, error) {
	return nil, nil
}

type silentNews struct{}

func (silentNews) FetchNews(context.Context, news.NewsQuery) ([]news.NewsResponse, error) {
	return nil, nil
}

func newSiteService(ctx context.Context, db *gorm.DB, cfg config.Score) site.Service {
	client, err := newStrictScoreClient(cfg)
	if err != nil || !client.reachable(ctx) {
		return nil
	}
	return site.NewService(site.NewRepository(db), client, silentRegulation{}, silentNews{})
}

type gormLocationPatcher struct {
	db *gorm.DB
}

func (p gormLocationPatcher) describeLocation(ctx context.Context, id uuid.UUID, details locationDetails) error {
	return p.db.WithContext(ctx).Model(&location.SavedLocation{}).Where("id = ?", id).Updates(map[string]any{
		"note":      details.note,
		"region_id": details.regionID,
		"area_sqm":  details.areaSqm,
	}).Error
}

func (p gormLocationPatcher) markSimulated(ctx context.Context, id uuid.UUID) error {
	return p.db.WithContext(ctx).Model(&location.SavedLocation{}).Where("id = ?", id).Update("score_simulated", true).Error
}
