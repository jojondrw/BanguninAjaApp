package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type linkRepository struct {
	Repository
	missingParties  map[uuid.UUID]bool
	projects        map[uuid.UUID]bool
	contracts       map[uuid.UUID]ContractRef
	orders          map[uuid.UUID]PurchaseOrderRef
	invoices        map[uuid.UUID]Invoice
	recorded        map[uuid.UUID]Receivable
	locked          Receivable
	createErr       error
	savedReceivable *Receivable
	savedPayable    *Payable
	savedInvoice    *Invoice
	invoiceDeleted  bool
}

func newLinkRepository() *linkRepository {
	return &linkRepository{
		missingParties: map[uuid.UUID]bool{},
		projects:       map[uuid.UUID]bool{},
		contracts:      map[uuid.UUID]ContractRef{},
		orders:         map[uuid.UUID]PurchaseOrderRef{},
		invoices:       map[uuid.UUID]Invoice{},
		recorded:       map[uuid.UUID]Receivable{},
	}
}

func (f *linkRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *linkRepository) PartyExists(_ context.Context, _ string, id uuid.UUID) (bool, error) {
	return !f.missingParties[id], nil
}

func (f *linkRepository) ProjectExists(_ context.Context, id uuid.UUID) (bool, error) {
	return f.projects[id], nil
}

func (f *linkRepository) FindContract(_ context.Context, id uuid.UUID) (ContractRef, error) {
	contract, found := f.contracts[id]
	if !found {
		return ContractRef{}, database.ErrNotFound
	}
	return contract, nil
}

func (f *linkRepository) FindPurchaseOrder(_ context.Context, id uuid.UUID) (PurchaseOrderRef, error) {
	order, found := f.orders[id]
	if !found {
		return PurchaseOrderRef{}, database.ErrNotFound
	}
	return order, nil
}

func (f *linkRepository) FindInvoice(_ context.Context, id uuid.UUID) (Invoice, error) {
	invoice, found := f.invoices[id]
	if !found {
		return Invoice{}, database.ErrNotFound
	}
	return invoice, nil
}

func (f *linkRepository) LockInvoice(ctx context.Context, id uuid.UUID) (Invoice, error) {
	return f.FindInvoice(ctx, id)
}

func (f *linkRepository) SaveInvoice(_ context.Context, invoice *Invoice) error {
	f.savedInvoice = invoice
	return nil
}

func (f *linkRepository) DeleteInvoice(context.Context, uuid.UUID) error {
	f.invoiceDeleted = true
	return nil
}

func (f *linkRepository) FindReceivableByInvoice(_ context.Context, invoiceID uuid.UUID) (Receivable, error) {
	receivable, found := f.recorded[invoiceID]
	if !found {
		return Receivable{}, database.ErrNotFound
	}
	return receivable, nil
}

func (f *linkRepository) LockReceivable(context.Context, uuid.UUID) (Receivable, error) {
	return f.locked, nil
}

func (f *linkRepository) CreateReceivable(_ context.Context, receivable *Receivable) error {
	if f.createErr != nil {
		return f.createErr
	}
	receivable.ID = uuid.New()
	f.savedReceivable = receivable
	return nil
}

func (f *linkRepository) SaveReceivable(_ context.Context, receivable *Receivable) error {
	f.savedReceivable = receivable
	return nil
}

func (f *linkRepository) FindReceivableRow(context.Context, uuid.UUID) (ReceivableRow, error) {
	return ReceivableRow{Receivable: *f.savedReceivable}, nil
}

func (f *linkRepository) CreatePayable(_ context.Context, payable *Payable) error {
	payable.ID = uuid.New()
	f.savedPayable = payable
	return nil
}

func (f *linkRepository) FindPayableRow(context.Context, uuid.UUID) (PayableRow, error) {
	return PayableRow{Payable: *f.savedPayable}, nil
}

type receivableFixture struct {
	repository *linkRepository
	customerID uuid.UUID
	projectID  uuid.UUID
	contractID uuid.UUID
	invoiceID  uuid.UUID
}

func newReceivableFixture() receivableFixture {
	fixture := receivableFixture{
		repository: newLinkRepository(),
		customerID: uuid.New(),
		projectID:  uuid.New(),
		contractID: uuid.New(),
		invoiceID:  uuid.New(),
	}
	fixture.repository.projects[fixture.projectID] = true
	fixture.repository.contracts[fixture.contractID] = ContractRef{ID: fixture.contractID, Number: "KTR-1", CustomerID: fixture.customerID}
	invoice := Invoice{Number: "INV-1", PartyType: customerParty, PartyID: fixture.customerID, Amount: 1000}
	invoice.ID = fixture.invoiceID
	fixture.repository.invoices[fixture.invoiceID] = invoice
	return fixture
}

func (f receivableFixture) request(amount int64) ReceivableRequest {
	return ReceivableRequest{
		CustomerID: f.customerID,
		ProjectID:  &f.projectID,
		ContractID: &f.contractID,
		InvoiceID:  &f.invoiceID,
		Reference:  "UJI-PIUTANG",
		DueDate:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		Amount:     amount,
	}
}

func assertAppError(t *testing.T, err error, status int, code string) *apperror.Error {
	t.Helper()
	var appError *apperror.Error
	if !errors.As(err, &appError) {
		t.Fatalf("got %v, want an apperror with code %s", err, code)
	}
	if appError.Status != status || appError.Code != code {
		t.Fatalf("got %d %s (%s), want %d %s", appError.Status, appError.Code, appError.Message, status, code)
	}
	return appError
}

func TestReceivableKeepsItsProjectContractAndInvoiceLinks(t *testing.T) {
	fixture := newReceivableFixture()

	response, err := fixedService(fixture.repository, time.Now()).CreateReceivable(context.Background(), fixture.request(1000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	saved := fixture.repository.savedReceivable
	if *saved.ProjectID != fixture.projectID || *saved.ContractID != fixture.contractID || *saved.InvoiceID != fixture.invoiceID {
		t.Fatalf("links were not saved: %+v", saved)
	}
	if *response.InvoiceID != fixture.invoiceID || *response.ContractID != fixture.contractID || *response.ProjectID != fixture.projectID {
		t.Fatalf("links were not returned: %+v", response)
	}
}

func TestReceivableLinkValidation(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(fixture receivableFixture, request *ReceivableRequest)
		status int
		code   string
	}{
		{
			name: "unknown customer",
			mutate: func(fixture receivableFixture, _ *ReceivableRequest) {
				fixture.repository.missingParties[fixture.customerID] = true
			},
			status: http.StatusUnprocessableEntity,
			code:   "customer_not_found",
		},
		{
			name: "unknown project",
			mutate: func(_ receivableFixture, request *ReceivableRequest) {
				unknown := uuid.New()
				request.ProjectID = &unknown
			},
			status: http.StatusUnprocessableEntity,
			code:   "project_not_found",
		},
		{
			name: "unknown contract",
			mutate: func(_ receivableFixture, request *ReceivableRequest) {
				unknown := uuid.New()
				request.ContractID = &unknown
			},
			status: http.StatusUnprocessableEntity,
			code:   "contract_not_found",
		},
		{
			name: "contract of another customer",
			mutate: func(fixture receivableFixture, _ *ReceivableRequest) {
				fixture.repository.contracts[fixture.contractID] = ContractRef{ID: fixture.contractID, CustomerID: uuid.New()}
			},
			status: http.StatusUnprocessableEntity,
			code:   "contract_customer_mismatch",
		},
		{
			name: "unknown invoice",
			mutate: func(_ receivableFixture, request *ReceivableRequest) {
				unknown := uuid.New()
				request.InvoiceID = &unknown
			},
			status: http.StatusUnprocessableEntity,
			code:   "linked_invoice_not_found",
		},
		{
			name: "invoice billed to another customer",
			mutate: func(fixture receivableFixture, _ *ReceivableRequest) {
				invoice := fixture.repository.invoices[fixture.invoiceID]
				invoice.PartyID = uuid.New()
				fixture.repository.invoices[fixture.invoiceID] = invoice
			},
			status: http.StatusUnprocessableEntity,
			code:   "invoice_customer_mismatch",
		},
		{
			name: "invoice billed to a vendor with the same id",
			mutate: func(fixture receivableFixture, _ *ReceivableRequest) {
				invoice := fixture.repository.invoices[fixture.invoiceID]
				invoice.PartyType = vendorParty
				fixture.repository.invoices[fixture.invoiceID] = invoice
			},
			status: http.StatusUnprocessableEntity,
			code:   "invoice_customer_mismatch",
		},
		{
			name: "amount above the invoice",
			mutate: func(_ receivableFixture, request *ReceivableRequest) {
				request.Amount = 1001
			},
			status: http.StatusUnprocessableEntity,
			code:   "receivable_exceeds_invoice",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newReceivableFixture()
			request := fixture.request(1000)
			tc.mutate(fixture, &request)

			_, err := fixedService(fixture.repository, time.Now()).CreateReceivable(context.Background(), request)
			assertAppError(t, err, tc.status, tc.code)
			if fixture.repository.savedReceivable != nil {
				t.Fatal("an invalid receivable must not be saved")
			}
		})
	}
}

func TestInvoiceCannotBeRecordedAsTwoReceivables(t *testing.T) {
	fixture := newReceivableFixture()
	fixture.repository.recorded[fixture.invoiceID] = Receivable{Reference: "PTG-007"}

	_, err := fixedService(fixture.repository, time.Now()).CreateReceivable(context.Background(), fixture.request(500))
	appError := assertAppError(t, err, http.StatusConflict, "invoice_already_recorded")
	if appError.Message != "Faktur ini sudah dicatat sebagai piutang PTG-007" {
		t.Fatalf("got message %q", appError.Message)
	}
	if fixture.repository.savedReceivable != nil {
		t.Fatal("a second receivable for the same invoice must not be saved")
	}
}

func TestConcurrentDuplicateInvoiceLinkIsAConflict(t *testing.T) {
	fixture := newReceivableFixture()
	fixture.repository.createErr = fmt.Errorf("%w: unique violation", database.ErrDuplicate)

	_, err := fixedService(fixture.repository, time.Now()).CreateReceivable(context.Background(), fixture.request(1000))
	assertAppError(t, err, http.StatusConflict, "invoice_already_recorded")
}

func TestReceivableUpdateMayKeepItsOwnInvoice(t *testing.T) {
	fixture := newReceivableFixture()
	ownID := uuid.New()
	own := Receivable{Reference: "PTG-001", Amount: 1000}
	own.ID = ownID
	fixture.repository.recorded[fixture.invoiceID] = own
	fixture.repository.locked = own

	_, err := fixedService(fixture.repository, time.Now()).UpdateReceivable(context.Background(), ownID, fixture.request(900))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fixture.repository.savedReceivable == nil || fixture.repository.savedReceivable.Amount != 900 {
		t.Fatalf("got %+v", fixture.repository.savedReceivable)
	}
}

func TestReceivableUpdateCannotTakeAnotherReceivablesInvoice(t *testing.T) {
	fixture := newReceivableFixture()
	other := Receivable{Reference: "PTG-002"}
	other.ID = uuid.New()
	fixture.repository.recorded[fixture.invoiceID] = other

	_, err := fixedService(fixture.repository, time.Now()).UpdateReceivable(context.Background(), uuid.New(), fixture.request(1000))
	appError := assertAppError(t, err, http.StatusConflict, "invoice_already_recorded")
	if appError.Message != "Faktur ini sudah dicatat sebagai piutang PTG-002" {
		t.Fatalf("got message %q", appError.Message)
	}
}

func TestRecordedInvoiceCannotBeDeleted(t *testing.T) {
	fixture := newReceivableFixture()
	fixture.repository.recorded[fixture.invoiceID] = Receivable{Reference: "PTG-003"}

	err := fixedService(fixture.repository, time.Now()).DeleteInvoice(context.Background(), fixture.invoiceID)
	assertAppError(t, err, http.StatusConflict, "invoice_has_receivable")
	if fixture.repository.invoiceDeleted {
		t.Fatal("an invoice recorded as a receivable must not be deleted")
	}
}

func TestRecordedInvoiceMustStillCoverItsReceivable(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(request *InvoiceRequest)
	}{
		{name: "another customer", mutate: func(request *InvoiceRequest) { request.PartyID = uuid.New() }},
		{name: "billed to a vendor", mutate: func(request *InvoiceRequest) { request.PartyType = vendorParty }},
		{name: "amount below the receivable", mutate: func(request *InvoiceRequest) { request.Amount = 799 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newReceivableFixture()
			fixture.repository.recorded[fixture.invoiceID] = Receivable{Reference: "PTG-004", CustomerID: fixture.customerID, Amount: 800}
			request := InvoiceRequest{Number: "INV-1", PartyType: customerParty, PartyID: fixture.customerID, Amount: 1000}
			tc.mutate(&request)

			_, err := fixedService(fixture.repository, time.Now()).UpdateInvoice(context.Background(), fixture.invoiceID, request)
			assertAppError(t, err, http.StatusUnprocessableEntity, "invoice_receivable_mismatch")
			if fixture.repository.savedInvoice != nil {
				t.Fatal("the invoice must not be saved")
			}
		})
	}
}

type payableFixture struct {
	repository *linkRepository
	vendorID   uuid.UUID
	projectID  uuid.UUID
	orderID    uuid.UUID
}

func newPayableFixture() payableFixture {
	fixture := payableFixture{repository: newLinkRepository(), vendorID: uuid.New(), projectID: uuid.New(), orderID: uuid.New()}
	fixture.repository.projects[fixture.projectID] = true
	fixture.repository.orders[fixture.orderID] = PurchaseOrderRef{ID: fixture.orderID, Number: "PO-1", VendorID: fixture.vendorID}
	return fixture
}

func (f payableFixture) request() PayableRequest {
	return PayableRequest{
		VendorID:        f.vendorID,
		ProjectID:       &f.projectID,
		PurchaseOrderID: &f.orderID,
		Reference:       "UJI-UTANG",
		DueDate:         time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		Amount:          5000,
	}
}

func TestPayableKeepsItsProjectAndPurchaseOrderLinks(t *testing.T) {
	fixture := newPayableFixture()

	response, err := fixedService(fixture.repository, time.Now()).CreatePayable(context.Background(), fixture.request())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *response.ProjectID != fixture.projectID || *response.PurchaseOrderID != fixture.orderID {
		t.Fatalf("links were not returned: %+v", response)
	}
}

func TestPayableLinkValidation(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(fixture payableFixture, request *PayableRequest)
		code   string
	}{
		{
			name: "unknown vendor",
			mutate: func(fixture payableFixture, _ *PayableRequest) {
				fixture.repository.missingParties[fixture.vendorID] = true
			},
			code: "vendor_not_found",
		},
		{
			name: "unknown project",
			mutate: func(_ payableFixture, request *PayableRequest) {
				unknown := uuid.New()
				request.ProjectID = &unknown
			},
			code: "project_not_found",
		},
		{
			name: "unknown purchase order",
			mutate: func(_ payableFixture, request *PayableRequest) {
				unknown := uuid.New()
				request.PurchaseOrderID = &unknown
			},
			code: "purchase_order_not_found",
		},
		{
			name: "purchase order of another vendor",
			mutate: func(fixture payableFixture, _ *PayableRequest) {
				fixture.repository.orders[fixture.orderID] = PurchaseOrderRef{ID: fixture.orderID, VendorID: uuid.New()}
			},
			code: "purchase_order_vendor_mismatch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPayableFixture()
			request := fixture.request()
			tc.mutate(fixture, &request)

			_, err := fixedService(fixture.repository, time.Now()).CreatePayable(context.Background(), request)
			assertAppError(t, err, http.StatusUnprocessableEntity, tc.code)
			if fixture.repository.savedPayable != nil {
				t.Fatal("an invalid payable must not be saved")
			}
		})
	}
}
