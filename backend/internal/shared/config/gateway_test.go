package config

import (
	"testing"
	"time"
)

var gatewayKeys = []string{"PAYMENT_GATEWAY", "XENDIT_SECRET_KEY", "XENDIT_CALLBACK_TOKEN", "XENDIT_API_URL", "GATEWAY_REQUEST_TTL"}

func clearGatewayEnv(t *testing.T) {
	t.Helper()
	for _, key := range gatewayKeys {
		t.Setenv(key, "")
	}
}

func TestGatewayIsDisabledByDefault(t *testing.T) {
	clearGatewayEnv(t)

	gateway, err := loadGateway()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gateway.Enabled() || gateway.RequestTTL != 30*time.Minute || gateway.XenditAPIURL != "https://api.xendit.co" {
		t.Fatalf("got %+v", gateway)
	}
}

func TestGatewayPicksXenditFromItsKeyAndDetectsTestMode(t *testing.T) {
	clearGatewayEnv(t)
	t.Setenv("XENDIT_SECRET_KEY", "xnd_development_abc")
	t.Setenv("XENDIT_CALLBACK_TOKEN", "callback")

	gateway, err := loadGateway()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gateway.Provider != GatewayXendit || !gateway.TestMode() {
		t.Fatalf("got %+v", gateway)
	}

	t.Setenv("XENDIT_SECRET_KEY", "xnd_production_abc")
	gateway, err = loadGateway()
	if err != nil || gateway.TestMode() {
		t.Fatalf("a production key must not allow simulated payments, got %+v %v", gateway, err)
	}
}

func TestGatewayRejectsIncompleteOrUnknownSettings(t *testing.T) {
	cases := map[string]map[string]string{
		"xendit without callback token": {"PAYMENT_GATEWAY": "xendit", "XENDIT_SECRET_KEY": "xnd_development_abc"},
		"xendit without key":            {"PAYMENT_GATEWAY": "xendit", "XENDIT_CALLBACK_TOKEN": "callback"},
		"unknown provider":              {"PAYMENT_GATEWAY": "midtrans"},
		"ttl too short":                 {"PAYMENT_GATEWAY": "fake", "GATEWAY_REQUEST_TTL": "10s"},
	}
	for name, values := range cases {
		clearGatewayEnv(t)
		for key, value := range values {
			t.Setenv(key, value)
		}
		if _, err := loadGateway(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestFakeGatewayIsAlwaysTestMode(t *testing.T) {
	clearGatewayEnv(t)
	t.Setenv("PAYMENT_GATEWAY", "fake")

	gateway, err := loadGateway()
	if err != nil || !gateway.Enabled() || !gateway.TestMode() {
		t.Fatalf("got %+v %v", gateway, err)
	}
}
