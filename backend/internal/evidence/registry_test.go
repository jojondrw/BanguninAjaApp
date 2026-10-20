package evidence

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const testRegistry = "0x05788638f4e23932884765f666D85654547AecB4"

func registryConfig() config.Onchain {
	settings := testConfig
	settings.RegistryAddress = testRegistry
	return settings
}

type stubWriter struct {
	nonce    uint64
	sent     []string
	estimate ethereum.CallRequest
	sendErr  error
}

func (w *stubWriter) PendingNonce(context.Context, string) (uint64, error) { return w.nonce, nil }
func (w *stubWriter) SuggestFees(context.Context) (ethereum.Fees, error) {
	return ethereum.Fees{MaxPriorityFeePerGas: big.NewInt(1_000_000_000), MaxFeePerGas: big.NewInt(20_000_000_000)}, nil
}
func (w *stubWriter) EstimateGas(_ context.Context, request ethereum.CallRequest) (uint64, error) {
	w.estimate = request
	return 48_000, nil
}
func (w *stubWriter) SendRawTransaction(_ context.Context, raw string) (string, error) {
	if w.sendErr != nil {
		return "", w.sendErr
	}
	w.sent = append(w.sent, raw)
	return "0x" + hex.EncodeToString(ethereum.Keccak256(mustHex(raw))), nil
}

func mustHex(value string) []byte {
	raw, _ := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	return raw
}

func TestRegistryCalldataEncodesAnchorCall(t *testing.T) {
	root := "0x" + strings.Repeat("ab", 32)
	calldata := RegistryCalldata(root, 2464)
	selector := hex.EncodeToString(ethereum.Keccak256([]byte("anchor(bytes32,uint32)"))[:4])
	if !strings.HasPrefix(calldata, "0x"+selector) || len(calldata) != 2+8+64+64 {
		t.Fatalf("got %s", calldata)
	}
	if !strings.HasSuffix(calldata, strings.Repeat("ab", 32)+strings.Repeat("0", 61)+"9a0") {
		t.Fatalf("root or count encoded wrongly: %s", calldata)
	}
}

func TestBatchPointsTheWalletAtTheRegistry(t *testing.T) {
	service := NewService(newMemory(), &stubSources{facts: []Fact{reportFact(uuid.New(), uuid.New(), "Galian")}}, &stubChain{}, registryConfig())
	batch, err := service.CreateBatch(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if batch.AnchorTo != testRegistry || batch.Calldata != RegistryCalldata(batch.Root, batch.RecordCount) {
		t.Fatalf("got %+v", batch)
	}

	legacy := NewService(newMemory(), &stubSources{facts: []Fact{reportFact(uuid.New(), uuid.New(), "Galian")}}, &stubChain{}, testConfig)
	plain, _ := legacy.CreateBatch(context.Background(), uuid.New())
	if plain.AnchorTo != "" || plain.Calldata != Calldata(plain.Root) {
		t.Fatalf("without a registry the old self-transaction format must stay, got %+v", plain)
	}
}

func TestAnchorRequiresTheRegistryAsRecipient(t *testing.T) {
	chain := &stubChain{}
	service := NewService(newMemory(), &stubSources{facts: []Fact{reportFact(uuid.New(), uuid.New(), "Galian")}}, chain, registryConfig())
	batch, err := service.CreateBatch(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	chain.seen = true
	chain.receipt = &ethereum.Receipt{Status: 1, BlockNumber: 9}
	chain.head = 9
	chain.transaction = ethereum.Transaction{ChainID: big.NewInt(11155111), To: "0x0000000000000000000000000000000000000001", Input: batch.Calldata}
	if _, err := service.Anchor(context.Background(), batch.ID, anchorHash); !errors.Is(err, errTxWrongTarget) {
		t.Fatalf("wrong recipient: got %v", err)
	}

	chain.transaction.To = strings.ToLower(testRegistry)
	result, err := service.Anchor(context.Background(), batch.ID, anchorHash)
	if err != nil || result.Status != anchorConfirmed {
		t.Fatalf("got %+v %v", result, err)
	}
}

func TestServerAnchorSignsACallToTheRegistry(t *testing.T) {
	writer := &stubWriter{nonce: 3}
	chain := &stubChain{}
	service := NewService(newMemory(), &stubSources{facts: []Fact{reportFact(uuid.New(), uuid.New(), "Galian")}}, chain, registryConfig())

	if _, err := service.AnchorWithServer(context.Background(), nil); !errors.Is(err, errNoServerSigner) {
		t.Fatalf("without a signer: got %v", err)
	}
	if service.Config().ServerSigner {
		t.Fatal("config must report no server signer")
	}

	signer, err := ethereum.NewSigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	if err != nil {
		t.Fatal(err)
	}
	service.UseServerSigner(writer, signer)
	sent, err := service.AnchorWithServer(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(writer.sent) != 1 || sent.TxHash == "" || !strings.HasPrefix(sent.TxURL, "https://sepolia.etherscan.io/tx/") {
		t.Fatalf("got %+v", sent)
	}
	if writer.estimate.To != testRegistry || writer.estimate.From != signer.Address() || writer.estimate.Data != sent.Batch.Calldata {
		t.Fatalf("gas was estimated for the wrong call: %+v", writer.estimate)
	}
	raw := mustHex(writer.sent[0])
	if raw[0] != 0x02 || !strings.Contains(hex.EncodeToString(raw), strings.TrimPrefix(strings.ToLower(testRegistry), "0x")) {
		t.Fatal("the signed transaction must be an EIP-1559 call to the registry")
	}
	settings := service.Config()
	if !settings.ServerSigner || settings.SignerAddress != signer.Address() || settings.RegistryURL != "https://sepolia.etherscan.io/address/"+testRegistry {
		t.Fatalf("got %+v", settings)
	}

	again, err := service.AnchorWithServer(context.Background(), nil)
	if err != nil || again.TxHash != sent.TxHash || len(writer.sent) != 1 {
		t.Fatalf("a run while the first transaction is unconfirmed must not send again, got %+v %v (%d sent)", again, err, len(writer.sent))
	}

	chain.seen = true
	chain.transaction = ethereum.Transaction{ChainID: big.NewInt(11155111), To: testRegistry, Input: sent.Batch.Calldata}
	chain.receipt = &ethereum.Receipt{Status: 1, BlockNumber: 12}
	chain.head = 12
	if _, err := service.AnchorWithServer(context.Background(), nil); !errors.Is(err, errNothingToAnchor) {
		t.Fatalf("after confirmation with nothing new there is nothing to send, got %v", err)
	}
	if len(writer.sent) != 1 {
		t.Fatalf("got %d transactions", len(writer.sent))
	}
}

func TestServerAnchorReportsSendFailures(t *testing.T) {
	writer := &stubWriter{sendErr: errors.New("insufficient funds for gas")}
	service := NewService(newMemory(), &stubSources{facts: []Fact{reportFact(uuid.New(), uuid.New(), "Galian")}}, &stubChain{}, testConfig)
	signer, _ := ethereum.NewSigner("0x0000000000000000000000000000000000000000000000000000000000000001")
	service.UseServerSigner(writer, signer)

	if _, err := service.AnchorWithServer(context.Background(), nil); apperror.From(err).Code != "evidence_anchor_send_failed" {
		t.Fatalf("got %v", err)
	}
	if writer.estimate.To != signer.Address() {
		t.Fatalf("without a registry the server sends to itself, estimated %+v", writer.estimate)
	}
}
