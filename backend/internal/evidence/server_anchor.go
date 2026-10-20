package evidence

import (
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const (
	gasBufferPercent  = 125
	minAnchorGas      = 30_000
	confirmationPoll  = 15 * time.Second
	confirmationTries = 40
	inflightLimit     = 30 * time.Minute
)

var (
	errNoServerSigner = apperror.New(http.StatusServiceUnavailable, "evidence_no_server_signer", "Dompet server belum dipasang. Isi EVIDENCE_SIGNER_KEY di server atau catat lewat MetaMask")
	errAnchorSend     = apperror.New(http.StatusBadGateway, "evidence_anchor_send_failed", "Transaksi bukti gagal dikirim ke jaringan. Pastikan dompet server punya SepoliaETH lalu coba lagi")
)

type ChainWriter interface {
	PendingNonce(ctx context.Context, address string) (uint64, error)
	SuggestFees(ctx context.Context) (ethereum.Fees, error)
	EstimateGas(ctx context.Context, request ethereum.CallRequest) (uint64, error)
	SendRawTransaction(ctx context.Context, raw string) (string, error)
}

type serverAnchor struct {
	writer   ChainWriter
	signer   *ethereum.Signer
	mu       sync.Mutex
	inflight *ServerAnchorResult
	sentAt   time.Time
}

type ServerAnchorResult struct {
	Batch  BatchResponse `json:"batch"`
	TxHash string        `json:"txHash"`
	TxURL  string        `json:"txUrl"`
}

type EvidenceConfig struct {
	ChainID         int64  `json:"chainId"`
	ExplorerURL     string `json:"explorerUrl"`
	RegistryAddress string `json:"registryAddress"`
	RegistryURL     string `json:"registryUrl"`
	ServerSigner    bool   `json:"serverSigner"`
	SignerAddress   string `json:"signerAddress"`
	AutoAnchorHours int    `json:"autoAnchorHours"`
}

func (s *Service) Config() EvidenceConfig {
	settings := EvidenceConfig{
		ChainID:         s.config.ChainID,
		ExplorerURL:     s.config.ExplorerURL,
		RegistryAddress: s.config.RegistryAddress,
		RegistryURL:     s.registryURL(),
		AutoAnchorHours: int(s.config.AutoAnchorInterval / time.Hour),
	}
	if s.server != nil {
		settings.ServerSigner = true
		settings.SignerAddress = s.server.signer.Address()
	}
	return settings
}

func (s *Service) AnchorWithServer(ctx context.Context, actorID *uuid.UUID) (ServerAnchorResult, error) {
	if s.server == nil {
		return ServerAnchorResult{}, errNoServerSigner
	}
	s.server.mu.Lock()
	defer s.server.mu.Unlock()

	if waiting, ok := s.stillInFlight(ctx); ok {
		return waiting, nil
	}
	batch, err := s.createBatch(ctx, actorID)
	if err != nil {
		return ServerAnchorResult{}, err
	}
	hash, err := s.send(ctx, batch)
	if err != nil {
		slog.Warn("evidence anchor send failed", slog.String("batch_id", batch.ID.String()), slog.String("error", err.Error()))
		return ServerAnchorResult{}, errAnchorSend.WithCause(err)
	}
	slog.Info("evidence anchor sent", slog.String("batch_id", batch.ID.String()), slog.String("root", batch.Root),
		slog.Int("records", batch.RecordCount), slog.String("tx_hash", hash), slog.String("from", s.server.signer.Address()))
	sent := ServerAnchorResult{Batch: s.batchResponse(batch), TxHash: hash, TxURL: s.txURL(hash)}
	s.server.inflight = &sent
	s.server.sentAt = s.now()
	return sent, nil
}

func (s *Service) stillInFlight(ctx context.Context) (ServerAnchorResult, bool) {
	inflight := s.server.inflight
	if inflight == nil {
		return ServerAnchorResult{}, false
	}
	result, err := s.Anchor(ctx, inflight.Batch.ID, inflight.TxHash)
	if err == nil && result.Status == anchorWaiting && s.now().Sub(s.server.sentAt) < inflightLimit {
		return *inflight, true
	}
	s.server.inflight = nil
	return ServerAnchorResult{}, false
}

func (s *Service) send(ctx context.Context, batch Batch) (string, error) {
	signer := s.server.signer
	target := s.target(batch)
	to := target.to
	if to == "" {
		to = signer.Address()
	}
	data, err := hex.DecodeString(strings.TrimPrefix(target.calldata, "0x"))
	if err != nil {
		return "", err
	}
	nonce, err := s.server.writer.PendingNonce(ctx, signer.Address())
	if err != nil {
		return "", err
	}
	fees, err := s.server.writer.SuggestFees(ctx)
	if err != nil {
		return "", err
	}
	gas, err := s.server.writer.EstimateGas(ctx, ethereum.CallRequest{From: signer.Address(), To: to, Data: target.calldata})
	if err != nil {
		return "", err
	}
	gas = max(minAnchorGas, gas*gasBufferPercent/100)
	signed, err := signer.Sign(ethereum.DynamicFeeTx{
		ChainID:              big.NewInt(batch.ChainID),
		Nonce:                nonce,
		MaxPriorityFeePerGas: fees.MaxPriorityFeePerGas,
		MaxFeePerGas:         fees.MaxFeePerGas,
		Gas:                  gas,
		To:                   to,
		Value:                big.NewInt(0),
		Data:                 data,
	})
	if err != nil {
		return "", err
	}
	hash, err := s.server.writer.SendRawTransaction(ctx, signed.Raw)
	if err != nil {
		return "", err
	}
	if hash != signed.Hash {
		slog.Warn("evidence anchor hash differs from the node", slog.String("signed", signed.Hash), slog.String("node", hash))
	}
	return hash, nil
}

func (s *Service) RunAutoAnchor(ctx context.Context) {
	if s.server == nil || s.config.AutoAnchorInterval <= 0 {
		return
	}
	slog.Info("evidence auto anchor scheduled", slog.String("every", s.config.AutoAnchorInterval.String()),
		slog.String("signer", s.server.signer.Address()))
	ticker := time.NewTicker(s.config.AutoAnchorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.autoAnchorOnce(ctx)
		}
	}
}

func (s *Service) autoAnchorOnce(ctx context.Context) {
	sent, err := s.AnchorWithServer(ctx, nil)
	if errors.Is(err, errNothingToAnchor) {
		slog.Info("evidence auto anchor skipped", slog.String("reason", "nothing new"))
		return
	}
	if err != nil {
		slog.Warn("evidence auto anchor failed", slog.String("error", err.Error()))
		return
	}
	for attempt := 0; attempt < confirmationTries; attempt++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(confirmationPoll):
		}
		result, err := s.Anchor(ctx, sent.Batch.ID, sent.TxHash)
		if err != nil {
			slog.Warn("evidence auto anchor not confirmed", slog.String("tx_hash", sent.TxHash), slog.String("error", err.Error()))
			return
		}
		if result.Status == anchorConfirmed {
			slog.Info("evidence auto anchor confirmed", slog.String("batch_id", sent.Batch.ID.String()), slog.String("tx_hash", sent.TxHash))
			return
		}
	}
	slog.Warn("evidence auto anchor still waiting", slog.String("tx_hash", sent.TxHash))
}
