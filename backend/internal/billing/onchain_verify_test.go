package billing

import (
	"context"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const (
	testReceiver = "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"
	testPayer    = "0x690c39adabdea83322bf8e90626cd40eeb456a95"
	testHash     = "0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f"
	otherHash    = "0x509d12ae3c3b226132673ab1bc507b55c967f0c43321ce97b71bee6b5cf4777b"
	testBlock    = uint64(11921168)
)

var onchainNow = time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)

type fakeChain struct {
	chainID      int64
	head         uint64
	transactions map[string]ethereum.Transaction
	receipts     map[string]ethereum.Receipt
	blockTimes   map[uint64]time.Time
	failure      error
}

func newFakeChain() *fakeChain {
	return &fakeChain{
		chainID:      sepoliaChainID,
		head:         testBlock + 4,
		transactions: map[string]ethereum.Transaction{},
		receipts:     map[string]ethereum.Receipt{},
		blockTimes:   map[uint64]time.Time{},
	}
}

func (f *fakeChain) ChainID(context.Context) (*big.Int, error) {
	return big.NewInt(f.chainID), f.failure
}

func (f *fakeChain) BlockNumber(context.Context) (uint64, error) {
	return f.head, f.failure
}

func (f *fakeChain) TransactionByHash(_ context.Context, hash string) (ethereum.Transaction, bool, error) {
	transaction, found := f.transactions[hash]
	return transaction, found, f.failure
}

func (f *fakeChain) TransactionReceipt(_ context.Context, hash string) (ethereum.Receipt, bool, error) {
	receipt, found := f.receipts[hash]
	return receipt, found, f.failure
}

func (f *fakeChain) BlockTime(_ context.Context, number uint64) (time.Time, error) {
	return f.blockTimes[number], f.failure
}

func (f *fakeChain) minedTransfer(hash, to, wei string, minedAt time.Time) {
	value, _ := new(big.Int).SetString(wei, 10)
	block := testBlock
	f.transactions[hash] = ethereum.Transaction{
		Hash: hash, From: testPayer, To: to, Value: value, ChainID: big.NewInt(sepoliaChainID), BlockNumber: &block,
	}
	f.receipts[hash] = ethereum.Receipt{Status: successfulReceipt, BlockNumber: block}
	f.blockTimes[block] = minedAt
}

func pendingRequest(amountWei string) OnchainRequest {
	request := OnchainRequest{
		TargetKind: onchainInvoiceTarget,
		TargetID:   uuid.New(),
		AmountIDR:  50_000_000,
		AmountWei:  amountWei,
		IDRPerETH:  1_000_000_000_000,
		Receiver:   testReceiver,
		ChainID:    sepoliaChainID,
		Status:     onchainPending,
		ExpiresAt:  onchainNow.Add(30 * time.Minute),
	}
	request.ID = uuid.New()
	request.CreatedAt = onchainNow
	return request
}

func verifierAt(chain ChainReader, now time.Time) transferVerifier {
	return transferVerifier{chain: chain, minConfirmations: 2, now: now}
}

func TestVerifierAcceptsAConfirmedTransferToTheReceiver(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, strings.ToLower(testReceiver), "60000000000000", onchainNow.Add(time.Minute))

	check, err := verifierAt(chain, onchainNow.Add(2*time.Minute)).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !check.verified || check.payer != testPayer || check.confirmations != 5 {
		t.Fatalf("got %+v", check)
	}
}

func TestVerifierRejectsTransferToAnotherAddress(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, "0xc2c8ea7679b5119fe934bc20262ac99d9497517c", "50000000000000", onchainNow.Add(time.Minute))

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_wrong_receiver")
}

func TestVerifierRejectsContractCreation(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, "", "50000000000000", onchainNow.Add(time.Minute))

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_wrong_receiver")
}

func TestVerifierRejectsTooLittleValueWithBothAmounts(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "49999999999999", onchainNow.Add(time.Minute))

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	appError := assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_value_too_low")
	if !strings.Contains(appError.Message, "terkirim 0,000049999999999999 SepoliaETH") || !strings.Contains(appError.Message, "minimal 0,00005 SepoliaETH") {
		t.Fatalf("got %q", appError.Message)
	}
}

func TestVerifierRejectsFailedReceipt(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(time.Minute))
	chain.receipts[testHash] = ethereum.Receipt{Status: 0, BlockNumber: testBlock}

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_tx_reverted")
}

func TestVerifierWaitsWhileTheTransferIsPending(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(time.Minute))
	pending := chain.transactions[testHash]
	pending.BlockNumber = nil
	chain.transactions[testHash] = pending
	delete(chain.receipts, testHash)

	check, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	if err != nil || check.verified || check.message != waitingForBlock {
		t.Fatalf("got %+v %v", check, err)
	}
}

func TestVerifierWaitsForTransactionsTheNodeHasNotSeen(t *testing.T) {
	check, err := verifierAt(newFakeChain(), onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	if err != nil || check.verified || check.message != waitingNotSeen {
		t.Fatalf("got %+v %v", check, err)
	}
}

func TestVerifierWaitsForEnoughConfirmations(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(time.Minute))
	chain.head = testBlock

	check, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	if err != nil || check.verified || check.confirmations != 1 || !strings.Contains(check.message, "(1 dari 2)") {
		t.Fatalf("got %+v %v", check, err)
	}
}

func TestVerifierRejectsTransferFromAnotherChain(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(time.Minute))
	mainnet := chain.transactions[testHash]
	mainnet.ChainID = big.NewInt(1)
	chain.transactions[testHash] = mainnet

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_wrong_chain")
}

func TestVerifierRefusesAnRpcNodeOnAnotherChain(t *testing.T) {
	chain := newFakeChain()
	chain.chainID = 1

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusBadGateway, "chain_mismatch")
}

func TestVerifierRejectsTransferMinedBeforeTheRequest(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(-time.Hour))

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_tx_before_request")
}

func TestVerifierRejectsTransferMinedAfterExpiry(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(40*time.Minute))

	_, err := verifierAt(chain, onchainNow.Add(41*time.Minute)).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_tx_after_expiry")
}

func TestVerifierStillAcceptsATransferMinedInTimeAfterExpiry(t *testing.T) {
	chain := newFakeChain()
	chain.minedTransfer(testHash, testReceiver, "50000000000000", onchainNow.Add(29*time.Minute))
	blockTime := onchainNow.Add(29 * time.Minute)
	withTimestamp := chain.transactions[testHash]
	withTimestamp.BlockTime = &blockTime
	chain.transactions[testHash] = withTimestamp
	chain.blockTimes = map[uint64]time.Time{}

	check, err := verifierAt(chain, onchainNow.Add(3*time.Hour)).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	if err != nil || !check.verified {
		t.Fatalf("got %+v %v", check, err)
	}
}

func TestVerifierStopsWaitingOnceTheRequestExpired(t *testing.T) {
	_, err := verifierAt(newFakeChain(), onchainNow.Add(time.Hour)).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_request_expired")
}

func TestVerifierReportsAnUnreachableNode(t *testing.T) {
	chain := newFakeChain()
	chain.failure = ethereum.ErrUnavailable

	_, err := verifierAt(chain, onchainNow).verify(context.Background(), pendingRequest("50000000000000"), testHash)
	appError := assertAppError(t, err, http.StatusBadGateway, "chain_unavailable")
	if !errors.Is(appError, ethereum.ErrUnavailable) {
		t.Fatal("the cause must be kept for the server log")
	}
}
