package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const (
	defaultOnchainRPCURL      = "https://ethereum-sepolia-rpc.publicnode.com"
	defaultOnchainExplorerURL = "https://sepolia.etherscan.io"
	defaultOnchainChainID     = 11155111
	defaultOnchainIDRPerETH   = 1_000_000_000_000
	defaultOnchainRequestTTL  = 30 * time.Minute
	defaultMinConfirmations   = 1
	maxOnchainIDRPerETH       = 1e15
	minOnchainRequestTTL      = time.Minute
)

func loadOnchain() (Onchain, error) {
	receiver, err := receiverAddressEnv("ONCHAIN_RECEIVER_ADDRESS")
	if err != nil {
		return Onchain{}, err
	}
	chainID, err := positiveIntEnv("ONCHAIN_CHAIN_ID", defaultOnchainChainID)
	if err != nil {
		return Onchain{}, err
	}
	rate, err := wholeNumberEnv("ONCHAIN_IDR_PER_ETH", defaultOnchainIDRPerETH)
	if err != nil {
		return Onchain{}, err
	}
	confirmations, err := positiveIntEnv("ONCHAIN_MIN_CONFIRMATIONS", defaultMinConfirmations)
	if err != nil {
		return Onchain{}, err
	}
	ttl, err := durationEnv("ONCHAIN_REQUEST_TTL", defaultOnchainRequestTTL)
	if err != nil {
		return Onchain{}, err
	}
	if ttl < minOnchainRequestTTL {
		return Onchain{}, fmt.Errorf("environment variable ONCHAIN_REQUEST_TTL must be at least %s", minOnchainRequestTTL)
	}

	return Onchain{
		ReceiverAddress:  receiver,
		RPCURL:           env("ONCHAIN_RPC_URL", defaultOnchainRPCURL),
		ChainID:          chainID,
		IDRPerETH:        rate,
		MinConfirmations: uint64(confirmations),
		RequestTTL:       ttl,
		ExplorerURL:      strings.TrimRight(env("ONCHAIN_EXPLORER_URL", defaultOnchainExplorerURL), "/"),
	}, nil
}

func receiverAddressEnv(key string) (string, error) {
	raw := env(key, "")
	if raw == "" {
		return "", nil
	}
	address, err := ethereum.ChecksumAddress(raw)
	if err != nil {
		return "", fmt.Errorf("environment variable %s is not a valid address: %w", key, err)
	}
	return address, nil
}

func positiveIntEnv(key string, fallback int64) (int64, error) {
	raw := env(key, "")
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("environment variable %s must be a positive whole number", key)
	}
	return parsed, nil
}

func wholeNumberEnv(key string, fallback int64) (int64, error) {
	raw := env(key, "")
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil || parsed < 1 || parsed > maxOnchainIDRPerETH || parsed != math.Trunc(parsed) {
		return 0, fmt.Errorf("environment variable %s must be a whole number between 1 and 1e15, such as 1e12", key)
	}
	return int64(parsed), nil
}
