package config

import (
	"fmt"
	"strings"
	"time"
)

const (
	GatewayXendit = "xendit"
	GatewayFake   = "fake"

	defaultXenditAPIURL      = "https://api.xendit.co"
	defaultGatewayRequestTTL = 30 * time.Minute
	minGatewayRequestTTL     = time.Minute
	maxGatewayRequestTTL     = 48 * time.Hour
	xenditTestKeyPrefix      = "xnd_development_"
)

type Gateway struct {
	Provider            string
	XenditSecretKey     string
	XenditCallbackToken string
	XenditAPIURL        string
	RequestTTL          time.Duration
}

func (g Gateway) Enabled() bool {
	return g.Provider != ""
}

func (g Gateway) TestMode() bool {
	return g.Provider == GatewayFake || strings.HasPrefix(g.XenditSecretKey, xenditTestKeyPrefix)
}

func loadGateway() (Gateway, error) {
	secret := env("XENDIT_SECRET_KEY", "")
	provider := strings.ToLower(env("PAYMENT_GATEWAY", ""))
	if provider == "" && secret != "" {
		provider = GatewayXendit
	}
	ttl, err := durationEnv("GATEWAY_REQUEST_TTL", defaultGatewayRequestTTL)
	if err != nil {
		return Gateway{}, err
	}
	if ttl < minGatewayRequestTTL || ttl > maxGatewayRequestTTL {
		return Gateway{}, fmt.Errorf("environment variable GATEWAY_REQUEST_TTL must be between %s and %s", minGatewayRequestTTL, maxGatewayRequestTTL)
	}
	gateway := Gateway{
		Provider:            provider,
		XenditSecretKey:     secret,
		XenditCallbackToken: env("XENDIT_CALLBACK_TOKEN", ""),
		XenditAPIURL:        strings.TrimRight(env("XENDIT_API_URL", defaultXenditAPIURL), "/"),
		RequestTTL:          ttl,
	}
	switch provider {
	case "", GatewayFake:
		return gateway, nil
	case GatewayXendit:
		if secret == "" || gateway.XenditCallbackToken == "" {
			return Gateway{}, fmt.Errorf("PAYMENT_GATEWAY=xendit needs XENDIT_SECRET_KEY and XENDIT_CALLBACK_TOKEN")
		}
		return gateway, nil
	default:
		return Gateway{}, fmt.Errorf("environment variable PAYMENT_GATEWAY must be xendit, fake, or empty")
	}
}
