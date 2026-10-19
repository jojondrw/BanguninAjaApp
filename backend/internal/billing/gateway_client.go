package billing

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

const (
	gatewayWaiting   = "waiting"
	gatewaySucceeded = "succeeded"
	gatewayExpired   = "expired"
	gatewayFailed    = "failed"
	gatewayCancelled = "cancelled"

	xenditAPIVersion   = "2024-11-11"
	xenditTimeout      = 15 * time.Second
	xenditErrorPreview = 300
	fakeQRPrefix       = "00020101021226BANGUNINAJA.FAKE.QRIS"
)

var errGatewayUnknownRequest = errors.New("gateway does not know this payment request")

type GatewayCharge struct {
	ReferenceID string
	Amount      int64
	ExpiresAt   time.Time
	Description string
}

type GatewayQR struct {
	ProviderID string
	QRString   string
}

type GatewayStatus struct {
	State  string
	Amount int64
}

type PaymentGateway interface {
	Provider() string
	CreateQRIS(ctx context.Context, charge GatewayCharge) (GatewayQR, error)
	Status(ctx context.Context, providerID string) (GatewayStatus, error)
	Cancel(ctx context.Context, providerID string) error
	Simulate(ctx context.Context, providerID string, amount int64) error
}

func NewPaymentGateway(settings config.Gateway) PaymentGateway {
	switch settings.Provider {
	case config.GatewayXendit:
		return &xenditGateway{baseURL: settings.XenditAPIURL, secretKey: settings.XenditSecretKey, client: &http.Client{Timeout: xenditTimeout}}
	case config.GatewayFake:
		return NewFakeGateway()
	default:
		return nil
	}
}

type xenditGateway struct {
	baseURL   string
	secretKey string
	client    *http.Client
}

type xenditAction struct {
	Type       string `json:"type"`
	Descriptor string `json:"descriptor"`
	Value      string `json:"value"`
}

type xenditPaymentRequest struct {
	PaymentRequestID string         `json:"payment_request_id"`
	Status           string         `json:"status"`
	RequestAmount    float64        `json:"request_amount"`
	Actions          []xenditAction `json:"actions"`
}

func (g *xenditGateway) Provider() string {
	return config.GatewayXendit
}

func (g *xenditGateway) CreateQRIS(ctx context.Context, charge GatewayCharge) (GatewayQR, error) {
	body := map[string]any{
		"reference_id":   charge.ReferenceID,
		"type":           "PAY",
		"country":        "ID",
		"currency":       "IDR",
		"channel_code":   "QRIS",
		"request_amount": charge.Amount,
		"description":    charge.Description,
		"channel_properties": map[string]any{
			"expires_at": charge.ExpiresAt.UTC().Format(time.RFC3339),
		},
	}
	var created xenditPaymentRequest
	if err := g.call(ctx, http.MethodPost, "/v3/payment_requests", body, &created); err != nil {
		return GatewayQR{}, err
	}
	for _, action := range created.Actions {
		if action.Descriptor == "QR_STRING" && action.Value != "" {
			return GatewayQR{ProviderID: created.PaymentRequestID, QRString: action.Value}, nil
		}
	}
	return GatewayQR{}, fmt.Errorf("xendit returned payment request %s without a QR string", created.PaymentRequestID)
}

func (g *xenditGateway) Status(ctx context.Context, providerID string) (GatewayStatus, error) {
	var current xenditPaymentRequest
	if err := g.call(ctx, http.MethodGet, "/v3/payment_requests/"+url.PathEscape(providerID), nil, &current); err != nil {
		return GatewayStatus{}, err
	}
	return GatewayStatus{State: xenditState(current.Status), Amount: int64(math.Round(current.RequestAmount))}, nil
}

func (g *xenditGateway) Cancel(ctx context.Context, providerID string) error {
	return g.call(ctx, http.MethodPost, "/v3/payment_requests/"+url.PathEscape(providerID)+"/cancel", map[string]any{}, nil)
}

func (g *xenditGateway) Simulate(ctx context.Context, providerID string, amount int64) error {
	return g.call(ctx, http.MethodPost, "/v3/payment_requests/"+url.PathEscape(providerID)+"/simulate", map[string]any{"amount": amount}, nil)
}

func xenditState(status string) string {
	switch status {
	case "SUCCEEDED":
		return gatewaySucceeded
	case "EXPIRED":
		return gatewayExpired
	case "CANCELED":
		return gatewayCancelled
	case "FAILED":
		return gatewayFailed
	default:
		return gatewayWaiting
	}
}

func (g *xenditGateway) call(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode xendit request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build xendit request: %w", err)
	}
	request.SetBasicAuth(g.secretKey, "")
	request.Header.Set("api-version", xenditAPIVersion)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := g.client.Do(request)
	if err != nil {
		return fmt.Errorf("call xendit %s %s: %w", method, path, err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read xendit response: %w", err)
	}
	if response.StatusCode == http.StatusNotFound {
		return errGatewayUnknownRequest
	}
	if response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("xendit %s %s answered %d: %s", method, path, response.StatusCode, truncateRunes(strings.TrimSpace(string(payload)), xenditErrorPreview))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode xendit response: %w", err)
	}
	return nil
}

type fakeCharge struct {
	amount int64
	state  string
}

type FakeGateway struct {
	mu      sync.Mutex
	charges map[string]*fakeCharge
}

func NewFakeGateway() *FakeGateway {
	return &FakeGateway{charges: map[string]*fakeCharge{}}
}

func (g *FakeGateway) Provider() string {
	return config.GatewayFake
}

func (g *FakeGateway) CreateQRIS(_ context.Context, charge GatewayCharge) (GatewayQR, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return GatewayQR{}, fmt.Errorf("generate fake payment id: %w", err)
	}
	id := "fake-" + hex.EncodeToString(raw)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.charges[id] = &fakeCharge{amount: charge.Amount, state: gatewayWaiting}
	return GatewayQR{ProviderID: id, QRString: fmt.Sprintf("%s.%s.%d", fakeQRPrefix, id, charge.Amount)}, nil
}

func (g *FakeGateway) Status(_ context.Context, providerID string) (GatewayStatus, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	charge, found := g.charges[providerID]
	if !found {
		return GatewayStatus{}, errGatewayUnknownRequest
	}
	return GatewayStatus{State: charge.state, Amount: charge.amount}, nil
}

func (g *FakeGateway) Cancel(_ context.Context, providerID string) error {
	return g.move(providerID, gatewayCancelled)
}

func (g *FakeGateway) Simulate(_ context.Context, providerID string, _ int64) error {
	return g.move(providerID, gatewaySucceeded)
}

func (g *FakeGateway) move(providerID, state string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	charge, found := g.charges[providerID]
	if !found {
		return errGatewayUnknownRequest
	}
	if charge.state == gatewayWaiting {
		charge.state = state
	}
	return nil
}
