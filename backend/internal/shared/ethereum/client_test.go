package ethereum

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	sepoliaHash = "0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f"

	sepoliaChainIDResponse     = `{"jsonrpc":"2.0","result":"0xaa36a7","id":1}`
	sepoliaBlockNumberResponse = `{"jsonrpc":"2.0","result":"0xb5e712","id":2}`

	sepoliaTransactionResponse = `{"jsonrpc":"2.0","id":5,"result":{"type":"0x2","chainId":"0xaa36a7","nonce":"0x8693","gas":"0x3d090",` +
		`"maxFeePerGas":"0x165a0bc16","maxPriorityFeePerGas":"0x165a0bc00","to":"0xc2c8ea7679b5119fe934bc20262ac99d9497517c",` +
		`"value":"0x38d7ea4c68000","accessList":[],"input":"0x","r":"0x4d069b0d8344489ad06722761aae04d829017b21dff7261b4bda8d7302b55761",` +
		`"s":"0x191edfaf85ac89bee0bf77b9e8ae51021c95505daaf6a21ccdc047c6523155be","yParity":"0x0","v":"0x0",` +
		`"hash":"0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f",` +
		`"blockHash":"0xe1c6b5656bc6af0522f0c22956d1451e78de6e857bc554b9413dcb4d11ef5f21","blockNumber":"0xb5e710",` +
		`"transactionIndex":"0x0","from":"0x690c39adabdea83322bf8e90626cd40eeb456a95","gasPrice":"0x165a0bc10","blockTimestamp":"0x6ad0e9d4"}}`

	sepoliaReceiptResponse = `{"jsonrpc":"2.0","id":6,"result":{"type":"0x2","status":"0x1","cumulativeGasUsed":"0x5208",` +
		`"logs":[{"address":"0xfffffffffffffffffffffffffffffffffffffffe","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",` +
		`"0x000000000000000000000000690c39adabdea83322bf8e90626cd40eeb456a95","0x000000000000000000000000c2c8ea7679b5119fe934bc20262ac99d9497517c"],` +
		`"data":"0x00000000000000000000000000000000000000000000000000038d7ea4c68000",` +
		`"blockHash":"0xe1c6b5656bc6af0522f0c22956d1451e78de6e857bc554b9413dcb4d11ef5f21","blockNumber":"0xb5e710","blockTimestamp":"0x6ad0e9d4",` +
		`"transactionHash":"0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f","transactionIndex":"0x0","logIndex":"0x0","removed":false}],` +
		`"logsBloom":"0x00","transactionHash":"0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f","transactionIndex":"0x0",` +
		`"blockHash":"0xe1c6b5656bc6af0522f0c22956d1451e78de6e857bc554b9413dcb4d11ef5f21","blockNumber":"0xb5e710","gasUsed":"0x5208",` +
		`"effectiveGasPrice":"0x165a0bc10","from":"0x690c39adabdea83322bf8e90626cd40eeb456a95","to":"0xc2c8ea7679b5119fe934bc20262ac99d9497517c",` +
		`"contractAddress":null}}`

	sepoliaBlockResponse = `{"jsonrpc":"2.0","id":3,"result":{"hash":"0xe1c6b5656bc6af0522f0c22956d1451e78de6e857bc554b9413dcb4d11ef5f21",` +
		`"number":"0xb5e710","timestamp":"0x6ad0e9d4","miner":"0x0c67c0f33baba3d1be81a69242bc18c51da27b03","baseFeePerGas":"0x10",` +
		`"transactions":["0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f"]}}`

	pendingTransactionResponse = `{"jsonrpc":"2.0","id":5,"result":{"type":"0x0","nonce":"0x1","gas":"0x5208",` +
		`"to":null,"value":"0x0","input":"0x6080","hash":"0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f",` +
		`"blockHash":null,"blockNumber":null,"transactionIndex":null,"from":"0x690c39adabdea83322bf8e90626cd40eeb456a95","gasPrice":"0x1"}}`

	notFoundResponse     = `{"jsonrpc":"2.0","id":7,"result":null}`
	invalidParamResponse = `{"jsonrpc":"2.0","id":9,"error":{"code":-32602,"message":"Invalid params","data":"invalid string length at line 1 column 6"}}`
)

type recordedCall struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

func fakeNode(t *testing.T, answers map[string]string) (*Client, *[]recordedCall) {
	t.Helper()
	calls := []recordedCall{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		var call recordedCall
		if err := json.Unmarshal(body, &call); err != nil {
			t.Errorf("decode request %s: %v", body, err)
		}
		calls = append(calls, call)
		answer, ok := answers[call.Method]
		if !ok {
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = writer.Write([]byte(answer))
	}))
	t.Cleanup(server.Close)
	return NewClient(server.URL), &calls
}

func TestClientReadsARealSepoliaTransfer(t *testing.T) {
	client, calls := fakeNode(t, map[string]string{"eth_getTransactionByHash": sepoliaTransactionResponse})

	transaction, found, err := client.TransactionByHash(context.Background(), sepoliaHash)
	if err != nil || !found {
		t.Fatalf("got found=%v err=%v", found, err)
	}
	if transaction.From != "0x690c39adabdea83322bf8e90626cd40eeb456a95" || transaction.To != "0xc2c8ea7679b5119fe934bc20262ac99d9497517c" {
		t.Fatalf("got %+v", transaction)
	}
	if transaction.Value.String() != "1000000000000000" || transaction.ChainID.Int64() != 11155111 {
		t.Fatalf("got value %s chain %s", transaction.Value, transaction.ChainID)
	}
	if transaction.BlockNumber == nil || *transaction.BlockNumber != 0xb5e710 {
		t.Fatalf("got block %v", transaction.BlockNumber)
	}
	if transaction.BlockTime == nil || !transaction.BlockTime.Equal(time.Unix(0x6ad0e9d4, 0)) {
		t.Fatalf("got block time %v", transaction.BlockTime)
	}
	if len(*calls) != 1 || (*calls)[0].Params[0] != sepoliaHash {
		t.Fatalf("got calls %+v", *calls)
	}
}

func TestClientReadsARealSepoliaReceiptBlockAndHead(t *testing.T) {
	client, calls := fakeNode(t, map[string]string{
		"eth_getTransactionReceipt": sepoliaReceiptResponse,
		"eth_getBlockByNumber":      sepoliaBlockResponse,
		"eth_blockNumber":           sepoliaBlockNumberResponse,
		"eth_chainId":               sepoliaChainIDResponse,
	})
	ctx := context.Background()

	receipt, found, err := client.TransactionReceipt(ctx, sepoliaHash)
	if err != nil || !found || receipt.Status != 1 || receipt.BlockNumber != 0xb5e710 {
		t.Fatalf("got %+v found=%v err=%v", receipt, found, err)
	}
	blockTime, err := client.BlockTime(ctx, receipt.BlockNumber)
	if err != nil || !blockTime.Equal(time.Unix(0x6ad0e9d4, 0)) {
		t.Fatalf("got %v %v", blockTime, err)
	}
	head, err := client.BlockNumber(ctx)
	if err != nil || head != 0xb5e712 {
		t.Fatalf("got %d %v", head, err)
	}
	chain, err := client.ChainID(ctx)
	if err != nil || chain.Int64() != 11155111 {
		t.Fatalf("got %v %v", chain, err)
	}
	blockCall := (*calls)[1]
	if blockCall.Params[0] != "0xb5e710" || blockCall.Params[1] != false {
		t.Fatalf("block must be asked by hex number without full transactions, got %+v", blockCall)
	}
}

func TestClientReportsPendingAndContractCreationTransactions(t *testing.T) {
	client, _ := fakeNode(t, map[string]string{"eth_getTransactionByHash": pendingTransactionResponse})

	transaction, found, err := client.TransactionByHash(context.Background(), sepoliaHash)
	if err != nil || !found {
		t.Fatalf("got found=%v err=%v", found, err)
	}
	if transaction.BlockNumber != nil || transaction.BlockTime != nil || transaction.ChainID != nil || transaction.To != "" {
		t.Fatalf("got %+v", transaction)
	}
}

func TestClientTreatsNullResultAsNotFound(t *testing.T) {
	client, _ := fakeNode(t, map[string]string{
		"eth_getTransactionByHash":  notFoundResponse,
		"eth_getTransactionReceipt": notFoundResponse,
	})

	if _, found, err := client.TransactionByHash(context.Background(), sepoliaHash); found || err != nil {
		t.Fatalf("got found=%v err=%v", found, err)
	}
	if _, found, err := client.TransactionReceipt(context.Background(), sepoliaHash); found || err != nil {
		t.Fatalf("got found=%v err=%v", found, err)
	}
}

func TestClientReportsNodeErrorsAsUnavailable(t *testing.T) {
	client, _ := fakeNode(t, map[string]string{"eth_getTransactionByHash": invalidParamResponse})

	_, _, err := client.TransactionByHash(context.Background(), "0x12")
	if !errors.Is(err, ErrUnavailable) || !strings.Contains(err.Error(), "Invalid params") {
		t.Fatalf("got %v", err)
	}
	if _, err := client.BlockNumber(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("an HTTP error must be unavailable, got %v", err)
	}
}

func TestClientRejectsMalformedQuantities(t *testing.T) {
	client, _ := fakeNode(t, map[string]string{"eth_blockNumber": `{"jsonrpc":"2.0","id":1,"result":"latest"}`})

	if _, err := client.BlockNumber(context.Background()); !errors.Is(err, ErrMalformed) {
		t.Fatalf("got %v", err)
	}
}

func TestClientReportsAnUnreachableNode(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")

	if _, err := client.ChainID(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("got %v", err)
	}
}
