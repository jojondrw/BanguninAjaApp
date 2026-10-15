package billing

import (
	"context"
	"maps"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

const testToken = "AbCdEfGhIjKlMnOpQrStUvWxYz012345"

type onchainStore struct {
	*ledgerStore
	requests      map[uuid.UUID]OnchainRequest
	usedReference map[string]bool
}

func newOnchainStore() *onchainStore {
	return &onchainStore{ledgerStore: newLedgerStore(), requests: map[uuid.UUID]OnchainRequest{}, usedReference: map[string]bool{}}
}

func (f *onchainStore) Transaction(_ context.Context, work func(Repository) error) error {
	invoices, receivables := maps.Clone(f.invoices), maps.Clone(f.receivables)
	payments, requests := append([]Payment(nil), f.payments...), maps.Clone(f.requests)
	f.inTransaction = true
	defer func() { f.inTransaction = false }()
	err := work(f)
	if err != nil {
		f.invoices, f.receivables, f.payments, f.requests = invoices, receivables, payments, requests
	}
	return err
}

func (f *onchainStore) CreateOnchainRequest(_ context.Context, request *OnchainRequest) error {
	request.ID = uuid.New()
	f.requests[request.ID] = *request
	return nil
}

func (f *onchainStore) FindOnchainRequestRow(_ context.Context, id uuid.UUID) (OnchainRequestRow, error) {
	request, found := f.requests[id]
	if !found {
		return OnchainRequestRow{}, database.ErrNotFound
	}
	return OnchainRequestRow{OnchainRequest: request}, nil
}

func (f *onchainStore) FindOnchainRequestByToken(_ context.Context, token string) (OnchainRequest, error) {
	for _, request := range f.requests {
		if request.Token == token {
			return request, nil
		}
	}
	return OnchainRequest{}, database.ErrNotFound
}

func (f *onchainStore) LockOnchainRequest(_ context.Context, id uuid.UUID) (OnchainRequest, error) {
	f.lock("onchain", id)
	request, found := f.requests[id]
	if !found {
		return OnchainRequest{}, database.ErrNotFound
	}
	return request, nil
}

func (f *onchainStore) SaveOnchainRequest(_ context.Context, request *OnchainRequest) error {
	f.requests[request.ID] = *request
	return nil
}

func (f *onchainStore) ListOnchainRequests(_ context.Context, filter OnchainRequestFilter) ([]OnchainRequestRow, int64, error) {
	rows := []OnchainRequestRow{}
	for _, request := range f.requests {
		if filter.Targets.match(request) {
			rows = append(rows, OnchainRequestRow{OnchainRequest: request})
		}
	}
	return rows, int64(len(rows)), nil
}

func (f *onchainStore) ActiveOnchainRequestExists(_ context.Context, targets OnchainTargets, now time.Time) (bool, error) {
	for _, request := range f.requests {
		if targets.match(request) && request.Status == onchainPending && request.ExpiresAt.After(now) {
			return true, nil
		}
	}
	return false, nil
}

func (f *onchainStore) OnchainTxHashTaken(_ context.Context, hash string, exceptID uuid.UUID) (bool, error) {
	for _, request := range f.requests {
		if request.ID != exceptID && request.TxHash != nil && *request.TxHash == hash {
			return true, nil
		}
	}
	return f.usedReference[hash], nil
}

func (t OnchainTargets) match(request OnchainRequest) bool {
	return (request.TargetKind == onchainInvoiceTarget && request.TargetID == t.InvoiceID) ||
		(request.TargetKind == onchainReceivableTarget && request.TargetID == t.ReceivableID)
}

func testOnchainSettings() config.Onchain {
	return config.Onchain{
		ReceiverAddress:  testReceiver,
		ChainID:          sepoliaChainID,
		IDRPerETH:        1_000_000_000_000,
		MinConfirmations: 2,
		RequestTTL:       30 * time.Minute,
		ExplorerURL:      "https://sepolia.etherscan.io",
	}
}

type onchainHarness struct {
	store   *onchainStore
	chain   *fakeChain
	now     time.Time
	service *onchainService
}

func newOnchainHarness() *onchainHarness {
	harness := &onchainHarness{store: newOnchainStore(), chain: newFakeChain(), now: onchainNow}
	harness.service = &onchainService{
		repository: harness.store,
		chain:      harness.chain,
		settings:   testOnchainSettings(),
		now:        func() time.Time { return harness.now },
		newToken:   func() (string, error) { return testToken, nil },
	}
	return harness
}

func (h *onchainHarness) customerInvoice(amount, paid int64) Invoice {
	invoice := h.store.addInvoice("UJI-INV-1", uuid.New(), amount, paid)
	invoice.DueDate = onchainNow.AddDate(0, 0, 14)
	h.store.invoices[invoice.ID] = invoice
	return invoice
}

func (h *onchainHarness) requestFor(t *testing.T, invoice Invoice) OnchainRequestResponse {
	t.Helper()
	created, err := h.service.CreateInvoiceRequest(context.Background(), invoice.ID, OnchainRequestCreate{}, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	return created
}

func (h *onchainHarness) confirm(hash string) (OnchainConfirmResponse, error) {
	return h.service.ConfirmPayment(context.Background(), testToken, OnchainConfirmRequest{TxHash: hash})
}

func TestCreateRequestDefaultsToOutstandingAtTheSimulationRate(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(80_000_000, 30_000_000)
	actor := uuid.New()

	created, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, OnchainRequestCreate{}, &actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.AmountIDR != 50_000_000 || created.AmountWei != "50000000000000" || created.AmountETH != "0.00005" {
		t.Fatalf("got %+v", created)
	}
	if created.Status != onchainPending || created.Token != testToken || !created.ExpiresAt.Equal(onchainNow.Add(30*time.Minute)) {
		t.Fatalf("got %+v", created)
	}
	if created.PaymentURI != "ethereum:"+testReceiver+"@11155111?value=50000000000000" || *created.CreatedBy != actor {
		t.Fatalf("got %+v", created)
	}
	assertLockOrder(t, harness.store.ledgerStore, "invoice:"+invoice.ID.String())
}

func TestCreateRequestValidatesTheAmountAndTarget(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(1_000_000, 400_000)

	_, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, OnchainRequestCreate{Amount: 600_001}, nil)
	appError := assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_amount_exceeds_outstanding")
	if !strings.Contains(appError.Message, "Rp 600.000") {
		t.Fatalf("got %q", appError.Message)
	}

	paid := harness.customerInvoice(1_000_000, 1_000_000)
	_, err = harness.service.CreateInvoiceRequest(context.Background(), paid.ID, OnchainRequestCreate{}, nil)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_nothing_outstanding")

	vendor := harness.customerInvoice(1_000_000, 0)
	vendor.PartyType = vendorParty
	harness.store.invoices[vendor.ID] = vendor
	_, err = harness.service.CreateInvoiceRequest(context.Background(), vendor.ID, OnchainRequestCreate{}, nil)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_vendor_invoice")

	if len(harness.store.requests) != 0 {
		t.Fatal("rejected requests must not be stored")
	}
}

func TestCreateRequestCapsAtTheLinkedReceivableAndBlocksASecondActiveRequest(t *testing.T) {
	harness := newOnchainHarness()
	invoice, receivable := linkedPair(harness.store.ledgerStore, 1_000_000, 700_000, 0, 0)

	created, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, OnchainRequestCreate{}, nil)
	if err != nil || created.AmountIDR != 700_000 {
		t.Fatalf("got %+v %v", created, err)
	}
	_, err = harness.service.CreateReceivableRequest(context.Background(), receivable.ID, OnchainRequestCreate{}, nil)
	assertAppError(t, err, http.StatusConflict, "onchain_request_active")

	harness.now = onchainNow.Add(31 * time.Minute)
	if _, err := harness.service.CreateReceivableRequest(context.Background(), receivable.ID, OnchainRequestCreate{}, nil); err != nil {
		t.Fatalf("an expired request must not block a new one: %v", err)
	}
}

func TestCreateRequestNeedsTheFeatureEnabled(t *testing.T) {
	harness := newOnchainHarness()
	harness.service.settings.ReceiverAddress = ""
	invoice := harness.customerInvoice(1_000_000, 0)

	_, err := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, OnchainRequestCreate{}, nil)
	assertAppError(t, err, http.StatusServiceUnavailable, "onchain_disabled")
	if harness.service.Config().Enabled || harness.service.Config().Receiver != "" {
		t.Fatal("config must report the feature disabled")
	}
}

func TestConfirmRecordsOneOnchainPaymentAndMarksTheRequestPaid(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(80_000_000, 30_000_000)
	actor := uuid.New()
	created, _ := harness.service.CreateInvoiceRequest(context.Background(), invoice.ID, OnchainRequestCreate{}, &actor)
	harness.chain.minedTransfer(testHash, strings.ToLower(testReceiver), created.AmountWei, onchainNow.Add(time.Minute))
	harness.now = onchainNow.Add(2 * time.Minute)

	outcome, err := harness.confirm("0x" + strings.ToUpper(testHash[2:]))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcome.Verification != verificationPaid || outcome.Status != onchainPaid || *outcome.TxHash != testHash || *outcome.PayerAddress != testPayer {
		t.Fatalf("got %+v", outcome)
	}
	if *outcome.TxURL != "https://sepolia.etherscan.io/tx/"+testHash || outcome.TargetNumber != "UJI-INV-1" {
		t.Fatalf("got %+v", outcome)
	}
	if len(harness.store.payments) != 1 {
		t.Fatalf("got %d payments", len(harness.store.payments))
	}
	payment := harness.store.payments[0]
	if payment.Method != onchainMethod || payment.Reference != testHash || payment.Amount != 50_000_000 || *payment.CreatedBy != actor {
		t.Fatalf("got %+v", payment)
	}
	if !strings.Contains(payment.Note, "https://sepolia.etherscan.io/tx/"+testHash) || !strings.Contains(payment.Note, testPayer) {
		t.Fatalf("got note %q", payment.Note)
	}
	settledInvoice := harness.store.invoices[invoice.ID]
	if settledInvoice.PaidAmount != 80_000_000 || settledInvoice.Status != duedate.Paid {
		t.Fatalf("got %+v", settledInvoice)
	}
}

func TestConfirmIsIdempotentForTheSameTransaction(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	created := harness.requestFor(t, invoice)
	harness.chain.minedTransfer(testHash, testReceiver, created.AmountWei, onchainNow.Add(time.Minute))

	if _, err := harness.confirm(testHash); err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	again, err := harness.confirm(testHash)
	if err != nil || again.Verification != verificationPaid || again.Status != onchainPaid {
		t.Fatalf("got %+v %v", again, err)
	}
	if len(harness.store.payments) != 1 || harness.store.invoices[invoice.ID].PaidAmount != 50_000_000 {
		t.Fatalf("confirming twice must not pay twice, got %d payments", len(harness.store.payments))
	}
	_, err = harness.confirm(otherHash)
	assertAppError(t, err, http.StatusConflict, "onchain_request_paid")
}

func TestConfirmRejectsAHashAlreadyUsedElsewhere(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	created := harness.requestFor(t, invoice)
	harness.chain.minedTransfer(testHash, testReceiver, created.AmountWei, onchainNow.Add(time.Minute))
	harness.store.usedReference[testHash] = true

	_, err := harness.confirm(testHash)
	assertAppError(t, err, http.StatusConflict, "onchain_tx_used")
	if len(harness.store.payments) != 0 {
		t.Fatal("a reused hash must not record a payment")
	}
}

func TestConfirmReportsWaitingWithoutRecording(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	harness.requestFor(t, invoice)

	outcome, err := harness.confirm(testHash)
	if err != nil || outcome.Verification != verificationWaiting || outcome.Message != waitingNotSeen || outcome.Status != onchainPending {
		t.Fatalf("got %+v %v", outcome, err)
	}
	if len(harness.store.payments) != 0 {
		t.Fatal("a waiting transfer must not record a payment")
	}
}

func TestConfirmRejectsAnExpiredRequestAndRecordsALateTransfer(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	created := harness.requestFor(t, invoice)
	harness.now = onchainNow.Add(time.Hour)

	_, err := harness.confirm(testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_request_expired")

	harness.chain.minedTransfer(testHash, testReceiver, created.AmountWei, onchainNow.Add(50*time.Minute))
	_, err = harness.confirm(testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_tx_after_expiry")
	stored := harness.store.requests[created.ID]
	if stored.Status != onchainExpired || !strings.Contains(stored.FailureReason, testHash) || len(harness.store.payments) != 0 {
		t.Fatalf("got %+v with %d payments", stored, len(harness.store.payments))
	}
	_, err = harness.confirm(testHash)
	assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_request_expired")
}

func TestConfirmMarksTheRequestFailedWhenTheInvoiceWasPaidMeanwhile(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	created := harness.requestFor(t, invoice)
	harness.chain.minedTransfer(testHash, testReceiver, created.AmountWei, onchainNow.Add(time.Minute))
	paidMeanwhile := harness.store.invoices[invoice.ID]
	paidMeanwhile.PaidAmount = 20_000_000
	harness.store.invoices[invoice.ID] = paidMeanwhile

	_, err := harness.confirm(testHash)
	appError := assertAppError(t, err, http.StatusUnprocessableEntity, "onchain_settlement_failed")
	if !strings.Contains(appError.Message, "Pembayaran melebihi sisa") {
		t.Fatalf("got %q", appError.Message)
	}
	stored := harness.store.requests[created.ID]
	if stored.Status != onchainFailed || *stored.TxHash != testHash || *stored.PayerAddress != testPayer || stored.FailureReason == "" {
		t.Fatalf("got %+v", stored)
	}
	if len(harness.store.payments) != 0 || harness.store.invoices[invoice.ID].PaidAmount != 20_000_000 {
		t.Fatal("a failed settlement must not change the invoice")
	}
	again, err := harness.confirm(testHash)
	if err != nil || again.Verification != verificationFailed || again.Status != onchainFailed {
		t.Fatalf("got %+v %v", again, err)
	}
}

func TestConfirmOnLinkedReceivableMovesBothBalances(t *testing.T) {
	harness := newOnchainHarness()
	invoice, receivable := linkedPair(harness.store.ledgerStore, 2_000_000, 2_000_000, 500_000, 500_000)
	created, err := harness.service.CreateReceivableRequest(context.Background(), receivable.ID, OnchainRequestCreate{Amount: 1_000_000}, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	harness.chain.minedTransfer(testHash, testReceiver, created.AmountWei, onchainNow.Add(time.Minute))

	if _, err := harness.confirm(testHash); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if harness.store.invoices[invoice.ID].PaidAmount != 1_500_000 || harness.store.receivables[receivable.ID].PaidAmount != 1_500_000 {
		t.Fatal("the linked invoice and receivable must both move")
	}
	if len(harness.store.payments) != 1 || *harness.store.payments[0].InvoiceID != invoice.ID || *harness.store.payments[0].ReceivableID != receivable.ID {
		t.Fatalf("want one shared payment row, got %+v", harness.store.payments)
	}
	listed, err := harness.service.InvoiceRequests(context.Background(), invoice.ID, OnchainRequestQuery{})
	if err != nil || len(listed.Items) != 1 || listed.Items[0].TargetKind != onchainReceivableTarget {
		t.Fatalf("the invoice must list the request made from its receivable, got %+v %v", listed, err)
	}
}

func TestCancelStopsFurtherConfirmation(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	created := harness.requestFor(t, invoice)

	cancelled, err := harness.service.CancelRequest(context.Background(), created.ID)
	if err != nil || cancelled.Status != onchainCancelled {
		t.Fatalf("got %+v %v", cancelled, err)
	}
	_, err = harness.confirm(testHash)
	assertAppError(t, err, http.StatusConflict, "onchain_request_cancelled")
	_, err = harness.service.CancelRequest(context.Background(), created.ID)
	assertAppError(t, err, http.StatusConflict, "onchain_request_not_pending")
}

func TestPublicPaymentShowsExpiryWithoutStoringIt(t *testing.T) {
	harness := newOnchainHarness()
	invoice := harness.customerInvoice(50_000_000, 0)
	created := harness.requestFor(t, invoice)
	harness.now = onchainNow.Add(31 * time.Minute)

	public, err := harness.service.PublicPayment(context.Background(), testToken)
	if err != nil || public.Status != onchainExpired || public.TargetNumber != "UJI-INV-1" || public.AmountETH != "0.00005" {
		t.Fatalf("got %+v %v", public, err)
	}
	if harness.store.requests[created.ID].Status != onchainPending {
		t.Fatal("reading the page must not change the stored status")
	}
	_, err = harness.service.PublicPayment(context.Background(), "ZZZZEfGhIjKlMnOpQrStUvWxYz012345")
	assertAppError(t, err, http.StatusNotFound, "onchain_request_not_found")
}
