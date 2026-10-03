package billing

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

type ledgerStore struct {
	Repository
	invoices      map[uuid.UUID]Invoice
	receivables   map[uuid.UUID]Receivable
	payables      map[uuid.UUID]Payable
	payments      []Payment
	links         [][2]uuid.UUID
	locks         []string
	inTransaction bool
}

func newLedgerStore() *ledgerStore {
	return &ledgerStore{
		invoices:    map[uuid.UUID]Invoice{},
		receivables: map[uuid.UUID]Receivable{},
		payables:    map[uuid.UUID]Payable{},
	}
}

func (f *ledgerStore) addInvoice(number string, customerID uuid.UUID, amount, paid int64) Invoice {
	invoice := Invoice{Number: number, PartyType: customerParty, PartyID: customerID, Amount: amount, PaidAmount: paid}
	invoice.ID = uuid.New()
	f.invoices[invoice.ID] = invoice
	return invoice
}

func (f *ledgerStore) addReceivable(reference string, customerID uuid.UUID, invoiceID *uuid.UUID, amount, paid int64) Receivable {
	receivable := Receivable{Reference: reference, CustomerID: customerID, InvoiceID: invoiceID, Amount: amount, PaidAmount: paid}
	receivable.ID = uuid.New()
	f.receivables[receivable.ID] = receivable
	return receivable
}

func (f *ledgerStore) Transaction(_ context.Context, work func(Repository) error) error {
	f.inTransaction = true
	defer func() { f.inTransaction = false }()
	return work(f)
}

func (f *ledgerStore) lock(kind string, id uuid.UUID) {
	if !f.inTransaction {
		panic("row locked outside a transaction")
	}
	f.locks = append(f.locks, kind+":"+id.String())
}

func (f *ledgerStore) FindInvoice(_ context.Context, id uuid.UUID) (Invoice, error) {
	invoice, found := f.invoices[id]
	if !found {
		return Invoice{}, database.ErrNotFound
	}
	return invoice, nil
}

func (f *ledgerStore) LockInvoice(ctx context.Context, id uuid.UUID) (Invoice, error) {
	f.lock("invoice", id)
	return f.FindInvoice(ctx, id)
}

func (f *ledgerStore) SaveInvoice(_ context.Context, invoice *Invoice) error {
	f.invoices[invoice.ID] = *invoice
	return nil
}

func (f *ledgerStore) FindInvoiceRow(ctx context.Context, id uuid.UUID) (InvoiceRow, error) {
	invoice, err := f.FindInvoice(ctx, id)
	return InvoiceRow{Invoice: invoice}, err
}

func (f *ledgerStore) FindReceivable(_ context.Context, id uuid.UUID) (Receivable, error) {
	receivable, found := f.receivables[id]
	if !found {
		return Receivable{}, database.ErrNotFound
	}
	return receivable, nil
}

func (f *ledgerStore) FindReceivableByInvoice(_ context.Context, invoiceID uuid.UUID) (Receivable, error) {
	for _, receivable := range f.receivables {
		if receivable.InvoiceID != nil && *receivable.InvoiceID == invoiceID {
			return receivable, nil
		}
	}
	return Receivable{}, database.ErrNotFound
}

func (f *ledgerStore) LockReceivable(ctx context.Context, id uuid.UUID) (Receivable, error) {
	f.lock("receivable", id)
	return f.FindReceivable(ctx, id)
}

func (f *ledgerStore) CreateReceivable(_ context.Context, receivable *Receivable) error {
	receivable.ID = uuid.New()
	f.receivables[receivable.ID] = *receivable
	return nil
}

func (f *ledgerStore) SaveReceivable(_ context.Context, receivable *Receivable) error {
	f.receivables[receivable.ID] = *receivable
	return nil
}

func (f *ledgerStore) FindReceivableRow(ctx context.Context, id uuid.UUID) (ReceivableRow, error) {
	receivable, err := f.FindReceivable(ctx, id)
	return ReceivableRow{Receivable: receivable}, err
}

func (f *ledgerStore) LockPayable(_ context.Context, id uuid.UUID) (Payable, error) {
	f.lock("payable", id)
	return f.payables[id], nil
}

func (f *ledgerStore) SavePayable(_ context.Context, payable *Payable) error {
	f.payables[payable.ID] = *payable
	return nil
}

func (f *ledgerStore) FindPayableRow(_ context.Context, id uuid.UUID) (PayableRow, error) {
	return PayableRow{Payable: f.payables[id]}, nil
}

func (f *ledgerStore) CreatePayment(_ context.Context, payment *Payment) error {
	f.payments = append(f.payments, *payment)
	return nil
}

func (f *ledgerStore) LinkPayments(_ context.Context, invoiceID, receivableID uuid.UUID) error {
	f.links = append(f.links, [2]uuid.UUID{invoiceID, receivableID})
	return nil
}

func (f *ledgerStore) PartyExists(context.Context, string, uuid.UUID) (bool, error) {
	return true, nil
}

func (f *ledgerStore) ProjectExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}

var paymentNow = time.Date(2026, 10, 2, 20, 0, 0, 0, time.UTC)

func linkedPair(store *ledgerStore, invoiceAmount, receivableAmount, invoicePaid, receivablePaid int64) (Invoice, Receivable) {
	customerID := uuid.New()
	invoice := store.addInvoice("INV-1", customerID, invoiceAmount, invoicePaid)
	receivable := store.addReceivable("PTG-1", customerID, &invoice.ID, receivableAmount, receivablePaid)
	return invoice, receivable
}

func assertLockOrder(t *testing.T, store *ledgerStore, expected ...string) {
	t.Helper()
	if strings.Join(store.locks, ",") != strings.Join(expected, ",") {
		t.Fatalf("got locks %v, want %v", store.locks, expected)
	}
}

func TestPaymentRecordsHistoryRowAndRunningTotal(t *testing.T) {
	store := newLedgerStore()
	payable := Payable{Reference: "UJI-UTANG", Amount: 5000, PaidAmount: 1000, DueDate: paymentNow.AddDate(0, 1, 0)}
	payable.ID = uuid.New()
	store.payables[payable.ID] = payable
	actor := uuid.New()
	paidAt := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	response, err := fixedService(store, paymentNow).PayPayable(context.Background(), payable.ID, PaymentRequest{
		Amount: 4000, PaidAt: &paidAt, Method: "cek", Reference: " CEK-778 ", Note: "Pelunasan termin akhir",
	}, &actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.PaidAmount != 5000 || response.Status != duedate.Paid || store.payables[payable.ID].PaidAmount != 5000 {
		t.Fatalf("got response %+v stored %+v", response, store.payables[payable.ID])
	}
	if len(store.payments) != 1 {
		t.Fatalf("got %d payment rows", len(store.payments))
	}
	row := store.payments[0]
	if *row.PayableID != payable.ID || row.InvoiceID != nil || row.ReceivableID != nil || row.Amount != 4000 {
		t.Fatalf("got row %+v", row)
	}
	if !row.PaidAt.Equal(paidAt) || row.Method != "cek" || row.Reference != "CEK-778" || row.Note != "Pelunasan termin akhir" || *row.CreatedBy != actor {
		t.Fatalf("got row %+v", row)
	}
}

func TestPaymentDefaultsToTodayInJakartaAndTransfer(t *testing.T) {
	store := newLedgerStore()
	payable := Payable{Amount: 5000}
	payable.ID = uuid.New()
	store.payables[payable.ID] = payable

	if _, err := fixedService(store, paymentNow).PayPayable(context.Background(), payable.ID, PaymentRequest{Amount: 100}, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	row := store.payments[0]
	if !row.PaidAt.Equal(time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)) || row.Method != "transfer" || row.CreatedBy != nil {
		t.Fatalf("got paidAt %s method %s createdBy %v", row.PaidAt, row.Method, row.CreatedBy)
	}
}

func TestPaymentDateCannotBeAfterToday(t *testing.T) {
	store := newLedgerStore()
	payable := Payable{Amount: 5000}
	payable.ID = uuid.New()
	store.payables[payable.ID] = payable
	tomorrow := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	_, err := fixedService(store, paymentNow).PayPayable(context.Background(), payable.ID, PaymentRequest{Amount: 100, PaidAt: &tomorrow}, nil)
	assertAppError(t, err, http.StatusUnprocessableEntity, "payment_date_in_future")
	if len(store.payments) != 0 || store.payables[payable.ID].PaidAmount != 0 {
		t.Fatal("a rejected payment must not be recorded")
	}
}

func TestInvoicePaymentAlsoPaysItsLinkedReceivable(t *testing.T) {
	store := newLedgerStore()
	invoice, receivable := linkedPair(store, 1000, 1000, 200, 200)

	response, err := fixedService(store, paymentNow).PayInvoice(context.Background(), invoice.ID, PaymentRequest{Amount: 300}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.PaidAmount != 500 || store.receivables[receivable.ID].PaidAmount != 500 {
		t.Fatalf("got invoice %d receivable %d", response.PaidAmount, store.receivables[receivable.ID].PaidAmount)
	}
	if len(store.payments) != 1 || *store.payments[0].InvoiceID != invoice.ID || *store.payments[0].ReceivableID != receivable.ID {
		t.Fatalf("want one shared payment row, got %+v", store.payments)
	}
	assertLockOrder(t, store, "invoice:"+invoice.ID.String(), "receivable:"+receivable.ID.String())
}

func TestReceivablePaymentAlsoPaysItsLinkedInvoice(t *testing.T) {
	store := newLedgerStore()
	invoice, receivable := linkedPair(store, 1500, 1000, 0, 0)

	response, err := fixedService(store, paymentNow).PayReceivable(context.Background(), receivable.ID, PaymentRequest{Amount: 1000}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.PaidAmount != 1000 || response.Status != duedate.Paid || store.invoices[invoice.ID].PaidAmount != 1000 {
		t.Fatalf("got receivable %+v invoice %+v", response, store.invoices[invoice.ID])
	}
	if store.invoices[invoice.ID].Status == duedate.Paid {
		t.Fatal("the invoice is worth more than the receivable, so it is not settled yet")
	}
	if len(store.payments) != 1 || *store.payments[0].InvoiceID != invoice.ID || *store.payments[0].ReceivableID != receivable.ID {
		t.Fatalf("want one shared payment row, got %+v", store.payments)
	}
	assertLockOrder(t, store, "invoice:"+invoice.ID.String(), "receivable:"+receivable.ID.String())
}

func TestLinkedPaymentCannotOverpayEitherSide(t *testing.T) {
	cases := []struct {
		name                                 string
		invoiceAmount, receivableAmount      int64
		invoicePaid, receivablePaid, payment int64
		payReceivable                        bool
		message                              string
	}{
		{name: "receivable side via invoice", invoiceAmount: 1000, receivableAmount: 500, payment: 600, message: "sisa piutang PTG-1 yang tertaut, sisanya tinggal Rp 500"},
		{name: "receivable side via receivable", invoiceAmount: 1000, receivableAmount: 500, payment: 600, payReceivable: true, message: "sisa piutang PTG-1"},
		{name: "invoice side via receivable", invoiceAmount: 1000, receivableAmount: 1000, invoicePaid: 800, payment: 300, payReceivable: true, message: "sisa faktur INV-1 yang tertaut, sisanya tinggal Rp 200"},
		{name: "invoice side via invoice", invoiceAmount: 1000, receivableAmount: 1000, invoicePaid: 800, payment: 300, message: "sisa faktur INV-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := newLedgerStore()
			invoice, receivable := linkedPair(store, tc.invoiceAmount, tc.receivableAmount, tc.invoicePaid, tc.receivablePaid)
			service := fixedService(store, paymentNow)

			var err error
			if tc.payReceivable {
				_, err = service.PayReceivable(context.Background(), receivable.ID, PaymentRequest{Amount: tc.payment}, nil)
			} else {
				_, err = service.PayInvoice(context.Background(), invoice.ID, PaymentRequest{Amount: tc.payment}, nil)
			}
			appError := assertAppError(t, err, http.StatusUnprocessableEntity, "payment_exceeds_outstanding")
			if !strings.Contains(appError.Message, tc.message) {
				t.Fatalf("got message %q, want it to contain %q", appError.Message, tc.message)
			}
			if len(store.payments) != 0 || store.invoices[invoice.ID].PaidAmount != tc.invoicePaid || store.receivables[receivable.ID].PaidAmount != tc.receivablePaid {
				t.Fatal("an overpayment must not change either side")
			}
		})
	}
}

func TestReceivablePaymentRefusesALinkChangedWhileWaitingForTheLock(t *testing.T) {
	store := newLedgerStore()
	invoice, receivable := linkedPair(store, 1000, 1000, 0, 0)
	moved := &relinkingStore{ledgerStore: store, receivableID: receivable.ID}

	_, err := fixedService(moved, paymentNow).PayReceivable(context.Background(), receivable.ID, PaymentRequest{Amount: 100}, nil)
	assertAppError(t, err, http.StatusConflict, "billing_link_changed")
	if len(store.payments) != 0 || store.invoices[invoice.ID].PaidAmount != 0 {
		t.Fatal("nothing may be recorded when the link changed")
	}
}

type relinkingStore struct {
	*ledgerStore
	receivableID uuid.UUID
}

func (f *relinkingStore) Transaction(ctx context.Context, work func(Repository) error) error {
	return f.ledgerStore.Transaction(ctx, func(Repository) error { return work(f) })
}

func (f *relinkingStore) LockReceivable(ctx context.Context, id uuid.UUID) (Receivable, error) {
	receivable, err := f.ledgerStore.LockReceivable(ctx, id)
	receivable.InvoiceID = nil
	return receivable, err
}

func TestLinkingReceivableCopiesThePaidTotalOfItsInvoice(t *testing.T) {
	store := newLedgerStore()
	customerID := uuid.New()
	invoice := store.addInvoice("INV-1", customerID, 1000, 400)

	response, err := fixedService(store, paymentNow).CreateReceivable(context.Background(), ReceivableRequest{
		CustomerID: customerID, InvoiceID: &invoice.ID, Reference: "PTG-1", DueDate: paymentNow, Amount: 1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.PaidAmount != 400 || response.Outstanding != 600 {
		t.Fatalf("got %+v", response)
	}
	if len(store.links) != 1 || store.links[0] != [2]uuid.UUID{invoice.ID, response.ID} {
		t.Fatalf("existing invoice payments must be shared with the receivable, got %v", store.links)
	}
	assertLockOrder(t, store, "invoice:"+invoice.ID.String())
}

func TestLinkingPaidReceivableCopiesItsTotalToTheInvoice(t *testing.T) {
	store := newLedgerStore()
	customerID := uuid.New()
	invoice := store.addInvoice("INV-1", customerID, 1000, 0)
	receivable := store.addReceivable("PTG-1", customerID, nil, 800, 300)

	_, err := fixedService(store, paymentNow).UpdateReceivable(context.Background(), receivable.ID, ReceivableRequest{
		CustomerID: customerID, InvoiceID: &invoice.ID, Reference: "PTG-1", DueDate: paymentNow, Amount: 800,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.invoices[invoice.ID].PaidAmount != 300 || store.receivables[receivable.ID].PaidAmount != 300 {
		t.Fatalf("got invoice %d receivable %d", store.invoices[invoice.ID].PaidAmount, store.receivables[receivable.ID].PaidAmount)
	}
	if len(store.links) != 1 || store.links[0] != [2]uuid.UUID{invoice.ID, receivable.ID} {
		t.Fatalf("got links %v", store.links)
	}
	assertLockOrder(t, store, "invoice:"+invoice.ID.String(), "receivable:"+receivable.ID.String())
}

func TestLinkConsistencyRule(t *testing.T) {
	cases := []struct {
		name             string
		invoicePaid      int64
		receivablePaid   int64
		receivableAmount int64
		code             string
	}{
		{name: "both paid with different totals", invoicePaid: 400, receivablePaid: 300, receivableAmount: 800, code: "link_paid_mismatch"},
		{name: "invoice paid more than the receivable is worth", invoicePaid: 900, receivableAmount: 800, code: "receivable_below_invoice_paid"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := newLedgerStore()
			customerID := uuid.New()
			invoice := store.addInvoice("INV-1", customerID, 1000, tc.invoicePaid)
			receivable := store.addReceivable("PTG-1", customerID, nil, tc.receivableAmount, tc.receivablePaid)

			_, err := fixedService(store, paymentNow).UpdateReceivable(context.Background(), receivable.ID, ReceivableRequest{
				CustomerID: customerID, InvoiceID: &invoice.ID, Reference: "PTG-1", DueDate: paymentNow, Amount: tc.receivableAmount,
			})
			assertAppError(t, err, http.StatusUnprocessableEntity, tc.code)
			if store.receivables[receivable.ID].InvoiceID != nil || len(store.links) != 0 {
				t.Fatal("a rejected link must not be saved")
			}
		})
	}
}

func TestLinkingWithEqualPaidTotalsKeepsBothSides(t *testing.T) {
	store := newLedgerStore()
	customerID := uuid.New()
	invoice := store.addInvoice("INV-1", customerID, 1000, 300)
	receivable := store.addReceivable("PTG-1", customerID, nil, 1000, 300)

	_, err := fixedService(store, paymentNow).UpdateReceivable(context.Background(), receivable.ID, ReceivableRequest{
		CustomerID: customerID, InvoiceID: &invoice.ID, Reference: "PTG-1", DueDate: paymentNow, Amount: 1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.invoices[invoice.ID].PaidAmount != 300 || store.receivables[receivable.ID].PaidAmount != 300 || len(store.links) != 0 {
		t.Fatalf("equal totals must be linked as they are, got links %v", store.links)
	}
}

func TestPaidReceivableCannotLeaveItsInvoice(t *testing.T) {
	store := newLedgerStore()
	invoice, receivable := linkedPair(store, 1000, 1000, 300, 300)
	other := store.addInvoice("INV-2", receivable.CustomerID, 1000, 0)

	for _, target := range []*uuid.UUID{nil, &other.ID} {
		_, err := fixedService(store, paymentNow).UpdateReceivable(context.Background(), receivable.ID, ReceivableRequest{
			CustomerID: receivable.CustomerID, InvoiceID: target, Reference: "PTG-1", DueDate: paymentNow, Amount: 1000,
		})
		assertAppError(t, err, http.StatusUnprocessableEntity, "receivable_link_has_payment")
		if *store.receivables[receivable.ID].InvoiceID != invoice.ID {
			t.Fatal("the receivable must stay linked to its invoice")
		}
	}
}

func TestRupiahGroupsThousands(t *testing.T) {
	cases := map[int64]string{0: "Rp 0", 500: "Rp 500", 1000: "Rp 1.000", 1250000: "Rp 1.250.000"}
	for value, want := range cases {
		if got := rupiah(value); got != want {
			t.Fatalf("rupiah(%d) = %q, want %q", value, got, want)
		}
	}
}
