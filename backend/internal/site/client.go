package site

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"
)

// ScoreInput is the internal /score request payload (contract §1.1).
type ScoreInput struct {
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	BuildingProfileCode string  `json:"building_profile_code"`
}

// ScoreResult is the internal /score response payload (contract §1.2).
type ScoreResult struct {
	OverallScore    int              `json:"overall_score"`
	DimensionScores []DimensionScore `json:"dimension_scores"`
	RiskFlags       []RiskFlag       `json:"risk_flags"`
}

// errUnknownProfile marks the internal service rejecting the profile code (§1.3).
// The service maps this to 502 (not 400) because the user sent a valid UUID and the
// mismatch is a server-side config problem between the DB and the scoring service.
var errUnknownProfile = errors.New("unknown_building_profile")

// ScoreClient calls the internal scoring service (Moses's Python /score, §1).
type ScoreClient interface {
	Score(ctx context.Context, input ScoreInput) (ScoreResult, error)
}

// scoredDimensions are the five real-data dimension codes the predictive score covers
// (contract §1.4). Regulasi & Zonasi is deliberately excluded (§4a).
var scoredDimensions = []string{
	"fisik_lingkungan",
	"infrastruktur",
	"demografi_sosial",
	"pasar_kompetisi",
	"finansial_proyek",
}

// httpScoreClient calls the real Python service over HTTP, and falls back to a
// deterministic stub whenever the service is not configured or unreachable.
type httpScoreClient struct {
	baseURL string
	client  *http.Client
	stub    ScoreClient
}

// NewScoreClient builds the score client. An empty baseURL means the service is not
// wired yet, so every call resolves through the stub. When a baseURL is configured,
// network/5xx failures still degrade to the stub so the demo path never hard-fails,
// while a genuine unknown-profile answer surfaces as an error the service maps to 502.
func NewScoreClient(baseURL string, timeout time.Duration) ScoreClient {
	stub := stubScoreClient{}
	if strings.TrimSpace(baseURL) == "" {
		return stub
	}
	return &httpScoreClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
		stub:    stub,
	}
}

func (c *httpScoreClient) Score(ctx context.Context, input ScoreInput) (ScoreResult, error) {
	result, err := c.call(ctx, input)
	if err == nil {
		return result, nil
	}
	// A recognised business error (unknown profile) must propagate; everything else
	// (timeouts, connection refused, 5xx, decode errors) degrades to the stub so the
	// endpoint keeps working while Moses's service is still coming online.
	if errors.Is(err, errUnknownProfile) {
		return ScoreResult{}, err
	}
	return c.stub.Score(ctx, input)
}

func (c *httpScoreClient) call(ctx context.Context, input ScoreInput) (ScoreResult, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return ScoreResult{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/score", bytes.NewReader(body))
	if err != nil {
		return ScoreResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return ScoreResult{}, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode == http.StatusOK {
		var result ScoreResult
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return ScoreResult{}, err
		}
		return result, nil
	}

	return ScoreResult{}, decodeScoreError(response)
}

func decodeScoreError(response *http.Response) error {
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(response.Body).Decode(&payload)
	if payload.Error.Code == "unknown_building_profile" {
		return errUnknownProfile
	}
	return fmt.Errorf("score service returned status %d", response.StatusCode)
}

// stubScoreClient produces a deterministic, plausible predictive score without any GIS
// data. It exists so T4/T16/T17 can be built and demoed before Moses's /score is live;
// point the SCORE_SERVICE_URL env var at his service to use real scores instead.
type stubScoreClient struct{}

const (
	stubScoreFloor = 55
	stubScoreSpan  = 40
)

var stubExplanations = map[string]string{
	"fisik_lingkungan": "Estimasi risiko banjir & gempa pada radius sekitar 1 km (data simulasi).",
	"infrastruktur":    "Estimasi akses jalan dan fasilitas kesehatan terdekat (data simulasi).",
	"demografi_sosial": "Estimasi kepadatan dan profil demografi sekitar lokasi (data simulasi).",
	"pasar_kompetisi":  "Estimasi permintaan belum terlayani dan kompetisi sejenis (data simulasi).",
	"finansial_proyek": "Estimasi nilai tanah (ZNT) dan indeks biaya konstruksi (data simulasi).",
}

func (stubScoreClient) Score(_ context.Context, input ScoreInput) (ScoreResult, error) {
	if input.BuildingProfileCode == "" {
		return ScoreResult{}, errUnknownProfile
	}

	dimensions := make([]DimensionScore, 0, len(scoredDimensions))
	total := 0
	for _, code := range scoredDimensions {
		value := stubValue(input, code)
		total += value
		dimensions = append(dimensions, DimensionScore{
			DimensionCode: code,
			Value:         value,
			Explanation:   stubExplanations[code],
		})
	}

	return ScoreResult{
		OverallScore:    total / len(scoredDimensions),
		DimensionScores: dimensions,
		RiskFlags:       []RiskFlag{},
	}, nil
}

// stubValue derives a stable pseudo-score in [floor, floor+span] from the coordinates,
// profile and dimension, so the same input always yields the same output.
func stubValue(input ScoreInput, dimensionCode string) int {
	hasher := fnv.New32a()
	_, _ = fmt.Fprintf(hasher, "%.4f:%.4f:%s:%s", input.Latitude, input.Longitude, input.BuildingProfileCode, dimensionCode)
	return stubScoreFloor + int(hasher.Sum32()%uint32(stubScoreSpan+1))
}
