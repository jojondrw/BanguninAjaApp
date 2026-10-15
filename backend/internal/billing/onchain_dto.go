package billing

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type OnchainRequestCreate struct {
	Amount int64 `json:"amount" binding:"omitempty,gt=0"`
}

type OnchainRequestQuery struct {
	pagination.Query
}

type OnchainConfirmRequest struct {
	TxHash string `json:"txHash" binding:"required,len=66,startswith=0x,hexadecimal"`
}

type OnchainConfigResponse struct {
	Enabled           bool   `json:"enabled"`
	ChainID           int64  `json:"chainId"`
	Receiver          string `json:"receiver"`
	IDRPerETH         int64  `json:"idrPerEth"`
	ExplorerURL       string `json:"explorerUrl"`
	MinConfirmations  uint64 `json:"minConfirmations"`
	RequestTTLMinutes int    `json:"requestTtlMinutes"`
}

type OnchainRequestResponse struct {
	ID            uuid.UUID  `json:"id"`
	TargetKind    string     `json:"targetKind"`
	TargetID      uuid.UUID  `json:"targetId"`
	AmountIDR     int64      `json:"amountIdr"`
	AmountWei     string     `json:"amountWei"`
	AmountETH     string     `json:"amountEth"`
	IDRPerETH     int64      `json:"idrPerEth"`
	Receiver      string     `json:"receiver"`
	ChainID       int64      `json:"chainId"`
	Token         string     `json:"token"`
	PaymentURI    string     `json:"paymentUri"`
	Status        string     `json:"status"`
	TxHash        *string    `json:"txHash"`
	TxURL         *string    `json:"txUrl"`
	PayerAddress  *string    `json:"payerAddress"`
	FailureReason string     `json:"failureReason"`
	CreatedBy     *uuid.UUID `json:"createdBy"`
	CreatedByName *string    `json:"createdByName"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	PaidAt        *time.Time `json:"paidAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type PublicPaymentResponse struct {
	TargetKind    string     `json:"targetKind"`
	TargetNumber  string     `json:"targetNumber"`
	ProjectName   *string    `json:"projectName"`
	AmountIDR     int64      `json:"amountIdr"`
	AmountWei     string     `json:"amountWei"`
	AmountETH     string     `json:"amountEth"`
	IDRPerETH     int64      `json:"idrPerEth"`
	Receiver      string     `json:"receiver"`
	ChainID       int64      `json:"chainId"`
	PaymentURI    string     `json:"paymentUri"`
	ExplorerURL   string     `json:"explorerUrl"`
	Status        string     `json:"status"`
	TxHash        *string    `json:"txHash"`
	TxURL         *string    `json:"txUrl"`
	PayerAddress  *string    `json:"payerAddress"`
	FailureReason string     `json:"failureReason"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	PaidAt        *time.Time `json:"paidAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type OnchainConfirmResponse struct {
	PublicPaymentResponse
	Verification  string `json:"verification"`
	Message       string `json:"message"`
	Confirmations uint64 `json:"confirmations"`
}

func newOnchainRequestResponse(row OnchainRequestRow, now time.Time, explorerURL string) OnchainRequestResponse {
	request := row.OnchainRequest
	return OnchainRequestResponse{
		ID:            request.ID,
		TargetKind:    request.TargetKind,
		TargetID:      request.TargetID,
		AmountIDR:     request.AmountIDR,
		AmountWei:     request.AmountWei,
		AmountETH:     etherAmount(request.AmountWei),
		IDRPerETH:     request.IDRPerETH,
		Receiver:      request.Receiver,
		ChainID:       request.ChainID,
		Token:         request.Token,
		PaymentURI:    paymentURI(request),
		Status:        effectiveStatus(request, now),
		TxHash:        request.TxHash,
		TxURL:         transactionURL(explorerURL, request.TxHash),
		PayerAddress:  request.PayerAddress,
		FailureReason: request.FailureReason,
		CreatedBy:     request.CreatedBy,
		CreatedByName: row.CreatedByName,
		ExpiresAt:     request.ExpiresAt,
		PaidAt:        request.PaidAt,
		CreatedAt:     request.CreatedAt,
	}
}

func newPublicPaymentResponse(request OnchainRequest, label targetLabel, now time.Time, explorerURL string) PublicPaymentResponse {
	return PublicPaymentResponse{
		TargetKind:    request.TargetKind,
		TargetNumber:  label.number,
		ProjectName:   label.projectName,
		AmountIDR:     request.AmountIDR,
		AmountWei:     request.AmountWei,
		AmountETH:     etherAmount(request.AmountWei),
		IDRPerETH:     request.IDRPerETH,
		Receiver:      request.Receiver,
		ChainID:       request.ChainID,
		PaymentURI:    paymentURI(request),
		ExplorerURL:   explorerURL,
		Status:        effectiveStatus(request, now),
		TxHash:        request.TxHash,
		TxURL:         transactionURL(explorerURL, request.TxHash),
		PayerAddress:  request.PayerAddress,
		FailureReason: request.FailureReason,
		ExpiresAt:     request.ExpiresAt,
		PaidAt:        request.PaidAt,
		CreatedAt:     request.CreatedAt,
	}
}
