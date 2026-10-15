package billing

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const (
	clockSkewAllowance = 2 * time.Minute
	successfulReceipt  = 1
	sepoliaChainID     = 11155111
	decimalBase        = 10

	waitingNotSeen          = "Transaksi belum terlihat di jaringan. Kalau baru saja dikirim, tunggu beberapa detik"
	waitingForBlock         = "Transaksi sudah terkirim dan sedang menunggu masuk blok"
	waitingForConfirmations = "Transaksi sudah masuk blok, menunggu konfirmasi jaringan (%d dari %d)"
)

var (
	errChainUnavailable = apperror.New(http.StatusBadGateway, "chain_unavailable", "Node jaringan blockchain tidak bisa dihubungi. Coba lagi sebentar")
	errTxReverted       = apperror.Unprocessable("onchain_tx_reverted", "Transaksi ini gagal di jaringan (reverted), jadi tidak ada dana yang masuk")
	errTxWrongReceiver  = apperror.Unprocessable("onchain_wrong_receiver", "Transaksi ini dikirim ke alamat lain, bukan ke alamat penerima perusahaan")
	errTxBeforeRequest  = apperror.Unprocessable("onchain_tx_before_request", "Transaksi ini terjadi sebelum permintaan bayar dibuat, jadi tidak bisa dipakai untuk tagihan ini")
	errTxAfterExpiry    = apperror.Unprocessable("onchain_tx_after_expiry", "Transaksi ini masuk setelah permintaan bayar kedaluwarsa. Hubungi bagian keuangan")
	errOnchainExpired   = apperror.Unprocessable("onchain_request_expired", "Permintaan bayar ini sudah kedaluwarsa. Minta tautan baru ke bagian keuangan")
)

type ChainReader interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionByHash(ctx context.Context, hash string) (ethereum.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, hash string) (ethereum.Receipt, bool, error)
	BlockTime(ctx context.Context, number uint64) (time.Time, error)
}

type transferCheck struct {
	verified      bool
	message       string
	confirmations uint64
	payer         string
}

type transferVerifier struct {
	chain            ChainReader
	minConfirmations uint64
	now              time.Time
}

func networkName(chainID int64) string {
	if chainID == sepoliaChainID {
		return "Sepolia"
	}
	return fmt.Sprintf("chain %d", chainID)
}

func currencyName(chainID int64) string {
	if chainID == sepoliaChainID {
		return "SepoliaETH"
	}
	return "ETH"
}

func chainMismatch(expected int64, actual *big.Int) *apperror.Error {
	return apperror.New(http.StatusBadGateway, "chain_mismatch", fmt.Sprintf(
		"Node RPC server terhubung ke chain %s, bukan %s (chain %d). Periksa ONCHAIN_RPC_URL", actual, networkName(expected), expected,
	))
}

func wrongChain(expected int64, actual *big.Int) *apperror.Error {
	return apperror.Unprocessable("onchain_wrong_chain", fmt.Sprintf(
		"Transaksi ini dari jaringan lain (chain %s), bukan %s (chain %d)", actual, networkName(expected), expected,
	))
}

func valueTooLow(sent, wanted *big.Int, chainID int64) *apperror.Error {
	currency := currencyName(chainID)
	return apperror.Unprocessable("onchain_value_too_low", fmt.Sprintf(
		"Nilai transaksi kurang: terkirim %s %s, seharusnya minimal %s %s",
		ethereum.FormatEther(sent), currency, ethereum.FormatEther(wanted), currency,
	))
}

func (v transferVerifier) verify(ctx context.Context, request OnchainRequest, hash string) (transferCheck, error) {
	if err := v.ensureChain(ctx, request.ChainID); err != nil {
		return transferCheck{}, err
	}
	transaction, found, err := v.chain.TransactionByHash(ctx, hash)
	if err != nil {
		return transferCheck{}, errChainUnavailable.WithCause(err)
	}
	if !found {
		return v.waiting(request, waitingNotSeen)
	}
	if err := matchTransfer(transaction, request); err != nil {
		return transferCheck{}, err
	}
	if transaction.BlockNumber == nil {
		return v.waiting(request, waitingForBlock)
	}
	return v.checkMined(ctx, request, transaction, hash)
}

func (v transferVerifier) ensureChain(ctx context.Context, expected int64) error {
	actual, err := v.chain.ChainID(ctx)
	if err != nil {
		return errChainUnavailable.WithCause(err)
	}
	if !actual.IsInt64() || actual.Int64() != expected {
		return chainMismatch(expected, actual)
	}
	return nil
}

func matchTransfer(transaction ethereum.Transaction, request OnchainRequest) error {
	if transaction.ChainID != nil && (!transaction.ChainID.IsInt64() || transaction.ChainID.Int64() != request.ChainID) {
		return wrongChain(request.ChainID, transaction.ChainID)
	}
	if !ethereum.SameAddress(transaction.To, request.Receiver) {
		return errTxWrongReceiver
	}
	wanted, ok := new(big.Int).SetString(request.AmountWei, decimalBase)
	if !ok {
		return apperror.Internal(fmt.Errorf("stored wei amount %q is not a number", request.AmountWei))
	}
	if transaction.Value.Cmp(wanted) < 0 {
		return valueTooLow(transaction.Value, wanted, request.ChainID)
	}
	return nil
}

func (v transferVerifier) checkMined(ctx context.Context, request OnchainRequest, transaction ethereum.Transaction, hash string) (transferCheck, error) {
	receipt, found, err := v.chain.TransactionReceipt(ctx, hash)
	if err != nil {
		return transferCheck{}, errChainUnavailable.WithCause(err)
	}
	if !found {
		return v.waiting(request, waitingForBlock)
	}
	if receipt.Status != successfulReceipt {
		return transferCheck{}, errTxReverted
	}
	if err := v.ensureMinedInWindow(ctx, request, transaction, receipt.BlockNumber); err != nil {
		return transferCheck{}, err
	}
	head, err := v.chain.BlockNumber(ctx)
	if err != nil {
		return transferCheck{}, errChainUnavailable.WithCause(err)
	}
	confirmations := confirmationsAt(head, receipt.BlockNumber)
	if confirmations < v.minConfirmations {
		message := fmt.Sprintf(waitingForConfirmations, confirmations, v.minConfirmations)
		return transferCheck{message: message, confirmations: confirmations}, nil
	}
	return transferCheck{verified: true, confirmations: confirmations, payer: transaction.From}, nil
}

func (v transferVerifier) ensureMinedInWindow(ctx context.Context, request OnchainRequest, transaction ethereum.Transaction, block uint64) error {
	minedAt, err := v.blockTime(ctx, transaction, block)
	if err != nil {
		return errChainUnavailable.WithCause(err)
	}
	if minedAt.Before(request.CreatedAt.Add(-clockSkewAllowance)) {
		return errTxBeforeRequest
	}
	if minedAt.After(request.ExpiresAt.Add(clockSkewAllowance)) {
		return errTxAfterExpiry
	}
	return nil
}

func (v transferVerifier) blockTime(ctx context.Context, transaction ethereum.Transaction, block uint64) (time.Time, error) {
	if transaction.BlockTime != nil {
		return *transaction.BlockTime, nil
	}
	return v.chain.BlockTime(ctx, block)
}

func (v transferVerifier) waiting(request OnchainRequest, message string) (transferCheck, error) {
	if v.now.After(request.ExpiresAt.Add(clockSkewAllowance)) {
		return transferCheck{}, errOnchainExpired
	}
	return transferCheck{message: message}, nil
}

func confirmationsAt(head, block uint64) uint64 {
	if head < block {
		return 0
	}
	return head - block + 1
}
