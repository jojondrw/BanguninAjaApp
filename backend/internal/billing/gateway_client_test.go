package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

func xenditTestServer(t *testing.T, handler func(t *testing.T, request *http.Request, body map[string]any) (int, string)) *xenditGateway {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user, password, ok := request.BasicAuth()
		if !ok || user != "xnd_development_uji" || password != "" {
			t.Errorf("the secret key must be sent as basic auth user, got %q %q %v", user, password, ok)
		}
		if request.Header.Get("api-version") != xenditAPIVersion {
			t.Errorf("missing api-version header, got %q", request.Header.Get("api-version"))
		}
		body := map[string]any{}
		if request.Body != nil && request.ContentLength != 0 {
			_ = json.NewDecoder(request.Body).Decode(&body)
		}
		status, response := handler(t, request, body)
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	gateway := NewPaymentGateway(config.Gateway{Provider: config.GatewayXendit, XenditSecretKey: "xnd_development_uji", XenditAPIURL: server.URL})
	return gateway.(*xenditGateway)
}

func TestXenditCreatesADynamicQRISPaymentRequest(t *testing.T) {
	expires := time.Date(2026, 10, 19, 10, 30, 0, 0, time.UTC)
	gateway := xenditTestServer(t, func(t *testing.T, request *http.Request, body map[string]any) (int, string) {
		if request.Method != http.MethodPost || request.URL.Path != "/v3/payment_requests" {
			t.Errorf("got %s %s", request.Method, request.URL.Path)
		}
		if body["channel_code"] != "QRIS" || body["currency"] != "IDR" || body["type"] != "PAY" || body["request_amount"] != float64(2_500_000) {
			t.Errorf("got body %v", body)
		}
		properties, _ := body["channel_properties"].(map[string]any)
		if properties["expires_at"] != "2026-10-19T10:30:00Z" || body["reference_id"] != "ref-1" {
			t.Errorf("got body %v", body)
		}
		return http.StatusCreated, `{"payment_request_id":"pr-uji","status":"REQUIRES_ACTION","request_amount":2500000,
			"actions":[{"type":"PRESENT_TO_CUSTOMER","descriptor":"QR_STRING","value":"00020101QR"}]}`
	})

	qr, err := gateway.CreateQRIS(context.Background(), GatewayCharge{ReferenceID: "ref-1", Amount: 2_500_000, ExpiresAt: expires})
	if err != nil || qr.ProviderID != "pr-uji" || qr.QRString != "00020101QR" {
		t.Fatalf("got %+v %v", qr, err)
	}
}

func TestXenditStatusMapsEveryOutcome(t *testing.T) {
	answers := map[string]string{
		"SUCCEEDED": gatewaySucceeded, "EXPIRED": gatewayExpired, "CANCELED": gatewayCancelled,
		"FAILED": gatewayFailed, "REQUIRES_ACTION": gatewayWaiting, "ACCEPTING_PAYMENTS": gatewayWaiting,
	}
	for xendit, want := range answers {
		gateway := xenditTestServer(t, func(t *testing.T, request *http.Request, _ map[string]any) (int, string) {
			if request.Method != http.MethodGet || request.URL.Path != "/v3/payment_requests/pr-uji" {
				t.Errorf("got %s %s", request.Method, request.URL.Path)
			}
			return http.StatusOK, `{"payment_request_id":"pr-uji","status":"` + xendit + `","request_amount":1500000}`
		})
		status, err := gateway.Status(context.Background(), "pr-uji")
		if err != nil || status.State != want || status.Amount != 1_500_000 {
			t.Errorf("%s: got %+v %v", xendit, status, err)
		}
	}
}

func TestXenditErrorsAreReported(t *testing.T) {
	gateway := xenditTestServer(t, func(*testing.T, *http.Request, map[string]any) (int, string) {
		return http.StatusNotFound, `{"error_code":"DATA_NOT_FOUND"}`
	})
	if _, err := gateway.Status(context.Background(), "pr-hilang"); !errors.Is(err, errGatewayUnknownRequest) {
		t.Fatalf("got %v", err)
	}

	gateway = xenditTestServer(t, func(*testing.T, *http.Request, map[string]any) (int, string) {
		return http.StatusBadRequest, `{"error_code":"API_VALIDATION_ERROR"}`
	})
	if err := gateway.Simulate(context.Background(), "pr-uji", 1000); err == nil {
		t.Fatal("a 400 answer must be an error")
	}
}

func TestXenditSimulateAndCancelUseTheirEndpoints(t *testing.T) {
	var paths []string
	gateway := xenditTestServer(t, func(t *testing.T, request *http.Request, body map[string]any) (int, string) {
		paths = append(paths, request.Method+" "+request.URL.Path)
		if request.URL.Path == "/v3/payment_requests/pr-uji/simulate" && body["amount"] != float64(1000) {
			t.Errorf("got body %v", body)
		}
		return http.StatusOK, `{"status":"PENDING"}`
	})
	if err := gateway.Simulate(context.Background(), "pr-uji", 1000); err != nil {
		t.Fatal(err)
	}
	if err := gateway.Cancel(context.Background(), "pr-uji"); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0] != "POST /v3/payment_requests/pr-uji/simulate" || paths[1] != "POST /v3/payment_requests/pr-uji/cancel" {
		t.Fatalf("got %v", paths)
	}
}

func TestNoGatewayWithoutProvider(t *testing.T) {
	if NewPaymentGateway(config.Gateway{}) != nil {
		t.Fatal("an empty provider must leave the gateway off")
	}
}
