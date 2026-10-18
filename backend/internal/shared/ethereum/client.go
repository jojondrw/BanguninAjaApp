package ethereum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	jsonRPCVersion   = "2.0"
	defaultTimeout   = 10 * time.Second
	maxResponseBytes = 4 << 20
)

var (
	ErrUnavailable = errors.New("ethereum node unavailable")
	ErrMalformed   = errors.New("ethereum node sent a malformed response")

	nullResult = []byte("null")
)

type Transaction struct {
	Hash        string
	From        string
	To          string
	Value       *big.Int
	ChainID     *big.Int
	BlockNumber *uint64
	BlockTime   *time.Time
	Input       string
}

type Receipt struct {
	Status      uint64
	BlockNumber uint64
}

type Client struct {
	url        string
	httpClient *http.Client
	nextID     atomic.Int64
}

func NewClient(url string) *Client {
	return &Client{url: url, httpClient: &http.Client{Timeout: defaultTimeout}}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rawTransaction struct {
	Hash           string  `json:"hash"`
	From           string  `json:"from"`
	To             *string `json:"to"`
	Value          string  `json:"value"`
	ChainID        *string `json:"chainId"`
	BlockNumber    *string `json:"blockNumber"`
	BlockTimestamp *string `json:"blockTimestamp"`
	Input          string  `json:"input"`
}

type rawReceipt struct {
	Status      string `json:"status"`
	BlockNumber string `json:"blockNumber"`
}

type rawBlock struct {
	Timestamp string `json:"timestamp"`
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var raw string
	if _, err := c.call(ctx, "eth_chainId", []any{}, &raw); err != nil {
		return nil, err
	}
	return malformedOnError(ParseQuantity(raw))
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var raw string
	if _, err := c.call(ctx, "eth_blockNumber", []any{}, &raw); err != nil {
		return 0, err
	}
	return malformedOnError(ParseUint64Quantity(raw))
}

func (c *Client) TransactionByHash(ctx context.Context, hash string) (Transaction, bool, error) {
	var raw rawTransaction
	found, err := c.call(ctx, "eth_getTransactionByHash", []any{hash}, &raw)
	if err != nil || !found {
		return Transaction{}, false, err
	}
	transaction, err := raw.decode()
	if err != nil {
		return Transaction{}, false, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	return transaction, true, nil
}

func (c *Client) TransactionReceipt(ctx context.Context, hash string) (Receipt, bool, error) {
	var raw rawReceipt
	found, err := c.call(ctx, "eth_getTransactionReceipt", []any{hash}, &raw)
	if err != nil || !found {
		return Receipt{}, false, err
	}
	receipt, err := raw.decode()
	if err != nil {
		return Receipt{}, false, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	return receipt, true, nil
}

func (c *Client) BlockTime(ctx context.Context, number uint64) (time.Time, error) {
	var raw rawBlock
	found, err := c.call(ctx, "eth_getBlockByNumber", []any{EncodeQuantity(number), false}, &raw)
	if err != nil {
		return time.Time{}, err
	}
	if !found {
		return time.Time{}, fmt.Errorf("%w: block %d not found", ErrMalformed, number)
	}
	return malformedOnError(unixQuantity(raw.Timestamp))
}

func (c *Client) call(ctx context.Context, method string, params []any, result any) (bool, error) {
	body, err := json.Marshal(rpcRequest{JSONRPC: jsonRPCVersion, ID: c.nextID.Add(1), Method: method, Params: params})
	if err != nil {
		return false, fmt.Errorf("encode %s: %w", method, err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return false, fmt.Errorf("%w: %s: %w", ErrUnavailable, method, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: %s answered HTTP %d", ErrUnavailable, method, response.StatusCode)
	}
	return decodeResult(response.Body, method, result)
}

func decodeResult(body io.Reader, method string, result any) (bool, error) {
	var envelope rpcResponse
	if err := json.NewDecoder(io.LimitReader(body, maxResponseBytes)).Decode(&envelope); err != nil {
		return false, fmt.Errorf("%w: %s: %w", ErrMalformed, method, err)
	}
	if envelope.Error != nil {
		return false, fmt.Errorf("%w: %s: %d %s", ErrUnavailable, method, envelope.Error.Code, envelope.Error.Message)
	}
	if len(envelope.Result) == 0 || bytes.Equal(envelope.Result, nullResult) {
		return false, nil
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return false, fmt.Errorf("%w: %s: %w", ErrMalformed, method, err)
	}
	return true, nil
}

func (r rawTransaction) decode() (Transaction, error) {
	value, err := ParseQuantity(r.Value)
	if err != nil {
		return Transaction{}, err
	}
	transaction := Transaction{Hash: r.Hash, From: r.From, Value: value, Input: strings.ToLower(r.Input)}
	if r.To != nil {
		transaction.To = *r.To
	}
	if transaction.ChainID, err = optionalQuantity(r.ChainID); err != nil {
		return Transaction{}, err
	}
	if transaction.BlockNumber, err = optionalUint64(r.BlockNumber); err != nil {
		return Transaction{}, err
	}
	if r.BlockTimestamp != nil {
		blockTime, err := unixQuantity(*r.BlockTimestamp)
		if err != nil {
			return Transaction{}, err
		}
		transaction.BlockTime = &blockTime
	}
	return transaction, nil
}

func (r rawReceipt) decode() (Receipt, error) {
	status, err := ParseUint64Quantity(r.Status)
	if err != nil {
		return Receipt{}, err
	}
	blockNumber, err := ParseUint64Quantity(r.BlockNumber)
	if err != nil {
		return Receipt{}, err
	}
	return Receipt{Status: status, BlockNumber: blockNumber}, nil
}

func optionalQuantity(raw *string) (*big.Int, error) {
	if raw == nil {
		return nil, nil
	}
	return ParseQuantity(*raw)
}

func optionalUint64(raw *string) (*uint64, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := ParseUint64Quantity(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func unixQuantity(raw string) (time.Time, error) {
	seconds, err := ParseUint64Quantity(raw)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(seconds), 0).UTC(), nil
}

func malformedOnError[T any](value T, err error) (T, error) {
	if err != nil {
		return value, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	return value, nil
}
