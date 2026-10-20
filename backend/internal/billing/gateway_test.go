package billing

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const testCallbackToken = "uji-callback-token"

type gatewayStore struct {
	*onchainStore
	gatewayRequests map[uuid.UUID]GatewayRequest
}

func newGatewayStore() *gatewayStore {
	return &gatewayStore{onchainStore: newOnchainStore(), gatewayRequests: map[uuid.UUID]GatewayRequest{}}
}

func (f *gatewayStore) Transaction(ctx context.Context, work func(Repository) error) error {
	snapshot := maps.Clone(f.gatewayRequests)
	err := f.onchainStore.Transaction(ctx, func(Repository) error { return work(f) })
	if err != nil {
		f.gatewayRequests = snapshot
	}
	return err
}

func (f *gatewayStore) CreateGatewayRequest(_ context.Context, request *GatewayRequest) error {
	f.gatewayRequests[request.ID] = *request
	return nil
}

func (f *gatewayStore) FindGatewayRequestRow(_ context.Context, id uuid.UUID) (GatewayRequestRow, error) {
	request, found := f.gatewayRequests[id]
	if !found {
		return GatewayRequestRow{}, database.ErrNotFound
	}
	return GatewayRequestRow{GatewayRequest: request}, nil
}

func (f *gatewayStore) FindGatewayRequestByToken(_ context.Context, token string) (GatewayRequest, error) {
	for _, request := range f.gatewayRequests {
		if request.Token == token {
			return request, nil
		}
	}
	return GatewayRequest{}, database.ErrNotFound
}

func (f *gatewayStore) FindGatewayRequestByProviderID(_ context.Context, providerID string) (GatewayRequest, error) {
	for _, request := range f.gatewayRequests {
		if request.ProviderID == providerID {
			return request, nil
		}
	}
	return GatewayRequest{}, database.ErrNotFound
}

func (f *gatewayStore) LockGatewayRequest(_ context.Context, id uuid.UUID) (GatewayRequest, error) {
	f.lock("gateway", id)
	request, found := f.gatewayRequests[id]
	if !found {
		return GatewayRequest{}, database.ErrNotFound
	}
	return request, nil
}

func (f *gatewayStore) SaveGatewayRequest(_ context.Context, request *GatewayRequest) error {
	f.gatewayRequests[request.ID] = *request
	return nil
}

func (f *gatewayStore) ListGatewayRequests(_ context.Context, filter OnchainRequestFilter) ([]GatewayRequestRow, int64, error) {
	rows := []GatewayRequestRow{}
	for _, request := range f.gatewayRequests {
		if filter.Targets.match(OnchainRequest{TargetKind: request.TargetKind, TargetID: request.TargetID}) {
			rows = append(rows, GatewayRequestRow{GatewayRequest: request})
		}
	}
	return rows, int64(len(rows)), nil
}

func (f *gatewayStore) ActiveGatewayRequestExists(_ context.Context, targets OnchainTargets, now time.Time) (bool, error) {
	for _, request := range f.gatewayRequests {
		match := targets.match(OnchainRequest{TargetKind: request.TargetKind, TargetID: request.TargetID})
		if match && request.Status == onchainPending && request.ExpiresAt.After(now) {
			return true, nil
		}
	}
	return false, nil
}

type countingGateway struct {
	*FakeGateway
	created   int
	cancelled int
	amount    int64
	statusErr error
}

func (g *countingGateway) CreateQRIS(ctx context.Context, charge GatewayCharge) (GatewayQR, error) {
	g.created++
	return g.FakeGateway.CreateQRIS(ctx, charge)
}

func (g *countingGateway) Cancel(ctx context.Context, providerID string) error {
	g.cancelled++
	return g.FakeGateway.Cancel(ctx, providerID)
}

func (g *countingGateway) Status(ctx context.Context, providerID string) (GatewayStatus, error) {
	if g.statusErr != nil {
		return GatewayStatus{}, g.statusErr
	}
	status, err := g.FakeGateway.Status(ctx, providerID)
	if g.amount != 0 {
		status.Amount = g.amount
	}
	return status, err
}

type gatewayHarness struct {
	store   *gatewayStore
	gateway *countingGateway
	now     time.Time
	service *gatewayService
}

func newGatewayHarness(settings config.Gateway) *gatewayHarness {
	harness := &gatewayHarness{store: newGatewayStore(), gateway: &countingGateway{FakeGateway: NewFakeGateway()}, now: onchainNow}
	harness.service = &gatewayService{
		repository: harness.store,
		gateway:    harness.gateway,
		settings:   settings,
		now:        func() time.Time { return harness.now },
		newToken:   func() (string, error) { return testToken, nil },
	}
	return harness
}

func testGatewaySettings() config.Gateway {
	return config.Gateway{Provider: config.GatewayFake, XenditCallbackToken: testCallbackToken, RequestTTL: 30 * time.Minute}
}

func (h *gatewayHarness) customerInvoice(amount, paid int64) Invoice {
	invoice := h.store.addInvoice("UJI-INV-QR", uuid.New(), amount, paid)
	invoice.DueDate = onchainNow.AddDate(0, 0, 14)
	h.store.invoices[invoice.ID] = invoice
	return invoice
}

func (h *gatewayHarness) request(t *testing.T, invoice Invoice) GatewayRequestResponse {
	t.Helper()
	created, err := h.service.CreateInvoiceRequest(context.Background(), invoice.ID, GatewayRequestCreate{}, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	return created
}

func qrisPayments(store *gatewayStore) []Payment {
	var payments []Payment
	for _, payment := range store.payments {
		if payment.Method == qrisMethod {
			payments = append(payments, payment)
		}
	}
	return payments
}

func TestGatewayRequestDefaultsToOutstandingAndShowsTheQR(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(8_000_000, 3_000_000)
	actor := uuid.New()

	created, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, GatewayRequestCreate{}, &actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.AmountIDR != 5_000_000 || created.Status != onchainPending || created.Token != testToken || created.QRString == "" {
		t.Fatalf("got %+v", created)
	}
	if created.Channel != qrisChannel || created.Provider != config.GatewayFake || *created.CreatedBy != actor {
		t.Fatalf("got %+v", created)
	}
	if !created.ExpiresAt.Equal(onchainNow.Add(30 * time.Minute)) {
		t.Fatalf("got expiry %s", created.ExpiresAt)
	}
}

func TestGatewayRequestRejectsLimitVendorAndDuplicate(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())

	large := harness.customerInvoice(25_000_000, 0)
	_, err := harness.service.CreateInvoiceRequest(context.Background(), large.ID, GatewayRequestCreate{}, nil)
	assertAppError(t, err, http.StatusUnprocessableEntity, "gateway_amount_over_limit")
	if _, err := harness.service.CreateInvoiceRequest(context.Background(), large.ID, GatewayRequestCreate{Amount: 10_000_000}, nil); err != nil {
		t.Fatalf("a part payment within the limit must be accepted: %v", err)
	}
	_, err = harness.service.CreateInvoiceRequest(context.Background(), large.ID, GatewayRequestCreate{Amount: 1_000_000}, nil)
	assertAppError(t, err, http.StatusConflict, "gateway_request_active")

	vendor := harness.customerInvoice(1_000_000, 0)
	vendor.PartyType = vendorParty
	harness.store.invoices[vendor.ID] = vendor
	_, err = harness.service.CreateInvoiceRequest(context.Background(), vendor.ID, GatewayRequestCreate{}, nil)
	assertAppError(t, err, http.StatusUnprocessableEntity, "gateway_vendor_invoice")

	if harness.gateway.created != 1 {
		t.Fatalf("rejected requests must not reach the gateway, got %d charges", harness.gateway.created)
	}
}

func TestGatewayDisabledRefusesNewRequests(t *testing.T) {
	harness := newGatewayHarness(config.Gateway{})
	harness.service.gateway = nil
	invoice := harness.customerInvoice(1_000_000, 0)

	_, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, GatewayRequestCreate{}, nil)
	assertAppError(t, err, http.StatusServiceUnavailable, "gateway_disabled")
	if harness.service.Config().Enabled {
		t.Fatal("config must report the gateway disabled")
	}
}

func TestGatewayRefreshWaitsUntilTheGatewaySaysPaid(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(2_000_000, 0)
	harness.request(t, invoice)

	waiting, err := harness.service.RefreshPayment(context.Background(), testToken)
	if err != nil || waiting.Status != onchainPending || len(qrisPayments(harness.store)) != 0 {
		t.Fatalf("got %+v %v", waiting, err)
	}

	paid, err := harness.service.SimulatePayment(context.Background(), testToken)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if paid.Status != onchainPaid || paid.QRString != "" || paid.PaidAt == nil {
		t.Fatalf("got %+v", paid)
	}
	payments := qrisPayments(harness.store)
	if len(payments) != 1 || payments[0].Amount != 2_000_000 || payments[0].Reference == "" {
		t.Fatalf("got %+v", payments)
	}
	if harness.store.invoices[invoice.ID].PaidAmount != 2_000_000 {
		t.Fatalf("the invoice must be settled, got %+v", harness.store.invoices[invoice.ID])
	}

	if _, err := harness.service.RefreshPayment(context.Background(), testToken); err != nil {
		t.Fatalf("refresh after payment: %v", err)
	}
	if len(qrisPayments(harness.store)) != 1 {
		t.Fatal("a repeated refresh must not record the payment twice")
	}
}

func TestGatewayWebhookChecksTheTokenAndAsksTheGateway(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(1_500_000, 0)
	created := harness.request(t, invoice)
	body := GatewayWebhook{Event: "payment.capture", Data: GatewayWebhookData{PaymentRequestID: created.ProviderID}}

	err := harness.service.HandleWebhook(context.Background(), "palsu", body)
	assertAppError(t, err, http.StatusUnauthorized, "gateway_callback_rejected")

	if err := harness.service.HandleWebhook(context.Background(), testCallbackToken, body); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if len(qrisPayments(harness.store)) != 0 {
		t.Fatal("a webhook alone must not mark the bill paid while the gateway still says waiting")
	}

	if err := harness.gateway.FakeGateway.Simulate(context.Background(), created.ProviderID, 0); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := harness.service.HandleWebhook(context.Background(), testCallbackToken, body); err != nil {
			t.Fatalf("webhook: %v", err)
		}
	}
	if len(qrisPayments(harness.store)) != 1 {
		t.Fatalf("duplicate webhooks must record one payment, got %d", len(qrisPayments(harness.store)))
	}

	unknown := GatewayWebhook{Data: GatewayWebhookData{PaymentRequestID: "pr-tidak-dikenal"}}
	if err := harness.service.HandleWebhook(context.Background(), testCallbackToken, unknown); err != nil {
		t.Fatalf("an unknown payment request must be acknowledged, got %v", err)
	}
}

func TestGatewayAmountMismatchFailsTheRequest(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(3_000_000, 0)
	harness.request(t, invoice)
	harness.gateway.amount = 2_999_000

	_, err := harness.service.SimulatePayment(context.Background(), testToken)
	assertAppError(t, err, http.StatusUnprocessableEntity, "gateway_settlement_failed")

	public, _ := harness.service.PublicPayment(context.Background(), testToken)
	if public.Status != onchainFailed || public.FailureReason == "" || len(qrisPayments(harness.store)) != 0 {
		t.Fatalf("got %+v", public)
	}
}

func TestGatewayCancelReachesTheGateway(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(1_000_000, 0)
	created := harness.request(t, invoice)

	cancelled, err := harness.service.CancelRequest(context.Background(), created.ID)
	if err != nil || cancelled.Status != onchainCancelled || harness.gateway.cancelled != 1 {
		t.Fatalf("got %+v %v", cancelled, err)
	}
	_, err = harness.service.SimulatePayment(context.Background(), testToken)
	assertAppError(t, err, http.StatusConflict, "gateway_request_not_pending")
	if _, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, GatewayRequestCreate{}, nil); err != nil {
		t.Fatalf("a new request after cancelling must be allowed: %v", err)
	}
}

func TestGatewaySimulationNeedsTestMode(t *testing.T) {
	harness := newGatewayHarness(config.Gateway{Provider: config.GatewayXendit, XenditSecretKey: "xnd_production_x", XenditCallbackToken: "t", RequestTTL: time.Hour})
	invoice := harness.customerInvoice(1_000_000, 0)
	harness.request(t, invoice)

	_, err := harness.service.SimulatePayment(context.Background(), testToken)
	assertAppError(t, err, http.StatusForbidden, "gateway_simulation_disabled")
	public, _ := harness.service.PublicPayment(context.Background(), testToken)
	if public.TestMode {
		t.Fatal("the public page must not offer simulation with a production key")
	}
}

func TestGatewayExpiryAndForgottenRequests(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(1_000_000, 0)
	harness.request(t, invoice)

	harness.now = onchainNow.Add(31 * time.Minute)
	public, _ := harness.service.PublicPayment(context.Background(), testToken)
	if public.Status != onchainExpired || public.QRString != "" {
		t.Fatalf("an expired request must hide its QR, got %+v", public)
	}

	harness.gateway.statusErr = errGatewayUnknownRequest
	refreshed, err := harness.service.RefreshPayment(context.Background(), testToken)
	if err != nil || refreshed.Status != onchainFailed {
		t.Fatalf("got %+v %v", refreshed, err)
	}

	harness.gateway.statusErr = errors.New("timeout")
	other := harness.customerInvoice(1_000_000, 0)
	harness.service.newToken = func() (string, error) { return "ZyXwVuTsRqPoNmLkJiHgFeDcBa987654", nil }
	if _, err := harness.service.CreateInvoiceRequest(context.Background(), other.ID, GatewayRequestCreate{}, nil); err != nil {
		t.Fatal(err)
	}
	_, err = harness.service.RefreshPayment(context.Background(), "ZyXwVuTsRqPoNmLkJiHgFeDcBa987654")
	assertAppError(t, err, http.StatusBadGateway, "gateway_unavailable")
}

func TestGatewayPaymentNoteNamesTheProvider(t *testing.T) {
	harness := newGatewayHarness(testGatewaySettings())
	invoice := harness.customerInvoice(1_000_000, 0)
	harness.request(t, invoice)
	if _, err := harness.service.SimulatePayment(context.Background(), testToken); err != nil {
		t.Fatal(err)
	}
	payments := qrisPayments(harness.store)
	if len(payments) != 1 || payments[0].Note != "Dibayar lewat QRIS (simulator lokal), dikonfirmasi otomatis oleh gerbang pembayaran" {
		t.Fatalf("got %+v", payments)
	}
	if providerName(config.GatewayXendit) != "Xendit" {
		t.Fatal("xendit must be named")
	}
}

func TestGatewayLogsEveryStepWithoutSecrets(t *testing.T) {
	var output bytes.Buffer
	harness := newGatewayHarness(testGatewaySettings())
	harness.service.logger = slog.New(slog.NewTextHandler(&output, nil))
	invoice := harness.customerInvoice(1_000_000, 0)
	created := harness.request(t, invoice)

	_ = harness.service.HandleWebhook(context.Background(), "palsu", GatewayWebhook{Event: "payment.capture"})
	if _, err := harness.service.SimulatePayment(context.Background(), testToken); err != nil {
		t.Fatal(err)
	}
	logs := output.String()
	for _, want := range []string{"qris request created", "qris webhook rejected", "qris payment simulated", "qris payment confirmed", created.ProviderID} {
		if !strings.Contains(logs, want) {
			t.Errorf("logs are missing %q:\n%s", want, logs)
		}
	}
	for _, secret := range []string{testCallbackToken, testToken, "palsu"} {
		if strings.Contains(logs, secret) {
			t.Errorf("logs must not contain %q", secret)
		}
	}
}
