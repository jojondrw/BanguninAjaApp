package config

import (
	"strings"
	"testing"
	"time"
)

var onchainKeys = []string{
	"ONCHAIN_RECEIVER_ADDRESS", "ONCHAIN_RPC_URL", "ONCHAIN_CHAIN_ID", "ONCHAIN_IDR_PER_ETH",
	"ONCHAIN_MIN_CONFIRMATIONS", "ONCHAIN_REQUEST_TTL", "ONCHAIN_EXPLORER_URL",
	"EVIDENCE_REGISTRY_ADDRESS", "EVIDENCE_SIGNER_KEY", "EVIDENCE_AUTO_ANCHOR_INTERVAL",
}

func clearOnchainEnv(t *testing.T) {
	t.Helper()
	for _, key := range onchainKeys {
		t.Setenv(key, "")
	}
}

func TestOnchainIsDisabledWithoutReceiverAndUsesSepoliaDefaults(t *testing.T) {
	clearOnchainEnv(t)

	onchain, err := loadOnchain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if onchain.Enabled() {
		t.Fatal("without a receiver address the feature must report itself disabled")
	}
	if onchain.ChainID != 11155111 || onchain.IDRPerETH != 1_000_000_000_000 || onchain.MinConfirmations != 1 {
		t.Fatalf("got %+v", onchain)
	}
	if onchain.RequestTTL != 30*time.Minute || onchain.RPCURL != defaultOnchainRPCURL || onchain.ExplorerURL != defaultOnchainExplorerURL {
		t.Fatalf("got %+v", onchain)
	}
}

func TestOnchainNormalisesTheReceiverToItsChecksum(t *testing.T) {
	clearOnchainEnv(t)
	t.Setenv("ONCHAIN_RECEIVER_ADDRESS", " 0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed ")
	t.Setenv("ONCHAIN_IDR_PER_ETH", "5e7")
	t.Setenv("ONCHAIN_EXPLORER_URL", "https://sepolia.otterscan.io/")

	onchain, err := loadOnchain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !onchain.Enabled() || onchain.ReceiverAddress != "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed" {
		t.Fatalf("got %q", onchain.ReceiverAddress)
	}
	if onchain.IDRPerETH != 50_000_000 || onchain.ExplorerURL != "https://sepolia.otterscan.io" {
		t.Fatalf("got %+v", onchain)
	}
}

func TestOnchainRejectsInvalidSettings(t *testing.T) {
	cases := map[string]string{
		"ONCHAIN_RECEIVER_ADDRESS":  "0x5AAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"ONCHAIN_CHAIN_ID":          "sepolia",
		"ONCHAIN_IDR_PER_ETH":       "0.5",
		"ONCHAIN_MIN_CONFIRMATIONS": "0",
		"ONCHAIN_REQUEST_TTL":       "10s",
	}
	for key, value := range cases {
		t.Run(key, func(t *testing.T) {
			clearOnchainEnv(t)
			t.Setenv(key, value)
			if _, err := loadOnchain(); err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestEvidenceRegistryAndSignerSettings(t *testing.T) {
	clearOnchainEnv(t)
	t.Setenv("EVIDENCE_REGISTRY_ADDRESS", "0x05788638f4e23932884765f666d85654547aecb4")
	t.Setenv("EVIDENCE_SIGNER_KEY", "0x0000000000000000000000000000000000000000000000000000000000000001")
	t.Setenv("EVIDENCE_AUTO_ANCHOR_INTERVAL", "24h")

	onchain, err := loadOnchain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if onchain.RegistryAddress != "0x05788638f4e23932884765f666D85654547AecB4" || !onchain.HasSigner() || onchain.AutoAnchorInterval != 24*time.Hour {
		t.Fatalf("got registry %s, signer %v, interval %s", onchain.RegistryAddress, onchain.HasSigner(), onchain.AutoAnchorInterval)
	}
}

func TestEvidenceSettingsRejectBadCombinations(t *testing.T) {
	cases := map[string]map[string]string{
		"bad key":              {"EVIDENCE_SIGNER_KEY": "0x1234"},
		"interval without key": {"EVIDENCE_AUTO_ANCHOR_INTERVAL": "24h"},
		"interval too short":   {"EVIDENCE_SIGNER_KEY": "0x" + strings.Repeat("0", 63) + "1", "EVIDENCE_AUTO_ANCHOR_INTERVAL": "1m"},
		"bad registry address": {"EVIDENCE_REGISTRY_ADDRESS": "0x1234"},
	}
	for name, values := range cases {
		clearOnchainEnv(t)
		for key, value := range values {
			t.Setenv(key, value)
		}
		if _, err := loadOnchain(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
