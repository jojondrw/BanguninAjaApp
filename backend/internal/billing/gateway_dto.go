package billing

import (
	"time"

	"github.com/google/uuid"
)

type GatewayRequestCreate struct {
	Amount int64 `json:"amount" binding:"omitempty,gt=0"`
}

type GatewayWebhook struct {
	Event string             `json:"event"`
	Data  GatewayWebhookData `json:"data"`
}

type GatewayWebhookData struct {
	PaymentRequestID string `json:"payment_request_id"`
}

type GatewayConfigResponse struct {
	Enabled           bool   `json:"enabled"`
	Provider          string `json:"provider"`
	TestMode          bool   `json:"testMode"`
	Channel           string `json:"channel"`
	MaxAmount         int64  `json:"maxAmount"`
	RequestTTLMinutes int    `json:"requestTtlMinutes"`
}

type GatewayRequestResponse struct {
	ID            uuid.UUID  `json:"id"`
	TargetKind    string     `json:"targetKind"`
	TargetID      uuid.UUID  `json:"targetId"`
	AmountIDR     int64      `json:"amountIdr"`
	Provider      string     `json:"provider"`
	Channel       string     `json:"channel"`
	ProviderID    string     `json:"providerId"`
	QRString      string     `json:"qrString"`
	Token         string     `json:"token"`
	Status        string     `json:"status"`
	FailureReason string     `json:"failureReason"`
	CreatedBy     *uuid.UUID `json:"createdBy"`
	CreatedByName *string    `json:"createdByName"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	PaidAt        *time.Time `json:"paidAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type GatewayPublicResponse struct {
	TargetKind    string     `json:"targetKind"`
	TargetNumber  string     `json:"targetNumber"`
	ProjectName   *string    `json:"projectName"`
	AmountIDR     int64      `json:"amountIdr"`
	Provider      string     `json:"provider"`
	Channel       string     `json:"channel"`
	QRString      string     `json:"qrString"`
	Status        string     `json:"status"`
	FailureReason string     `json:"failureReason"`
	TestMode      bool       `json:"testMode"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	PaidAt        *time.Time `json:"paidAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func newGatewayRequestResponse(row GatewayRequestRow, now time.Time) GatewayRequestResponse {
	request := row.GatewayRequest
	status := gatewayEffectiveStatus(request, now)
	return GatewayRequestResponse{
		ID:            request.ID,
		TargetKind:    request.TargetKind,
		TargetID:      request.TargetID,
		AmountIDR:     request.AmountIDR,
		Provider:      request.Provider,
		Channel:       request.Channel,
		ProviderID:    request.ProviderID,
		QRString:      visibleQR(request, status),
		Token:         request.Token,
		Status:        status,
		FailureReason: request.FailureReason,
		CreatedBy:     request.CreatedBy,
		CreatedByName: row.CreatedByName,
		ExpiresAt:     request.ExpiresAt,
		PaidAt:        request.PaidAt,
		CreatedAt:     request.CreatedAt,
	}
}

func newGatewayPublicResponse(request GatewayRequest, label targetLabel, now time.Time, testMode bool) GatewayPublicResponse {
	status := gatewayEffectiveStatus(request, now)
	return GatewayPublicResponse{
		TargetKind:    request.TargetKind,
		TargetNumber:  label.number,
		ProjectName:   label.projectName,
		AmountIDR:     request.AmountIDR,
		Provider:      request.Provider,
		Channel:       request.Channel,
		QRString:      visibleQR(request, status),
		Status:        status,
		FailureReason: request.FailureReason,
		TestMode:      testMode && status == onchainPending,
		ExpiresAt:     request.ExpiresAt,
		PaidAt:        request.PaidAt,
		CreatedAt:     request.CreatedAt,
	}
}

func visibleQR(request GatewayRequest, status string) string {
	if status != onchainPending {
		return ""
	}
	return request.QRString
}
