package billing

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

var (
	errInvoiceNotFound     = apperror.NotFound("invoice_not_found", "Tagihan tidak ditemukan")
	errInvoiceNumberUsed   = apperror.Conflict("invoice_number_used", "Nomor tagihan sudah dipakai")
	errInvoiceHasPayment   = apperror.Conflict("invoice_has_payment", "Tagihan yang sudah dibayar sebagian atau penuh tidak bisa dihapus")
	errInvoicePartyMissing = apperror.Unprocessable("invoice_party_not_found", "Pelanggan atau vendor yang ditagih tidak ditemukan")
	errInvoiceProject      = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")

	errReceivableNotFound   = apperror.NotFound("receivable_not_found", "Piutang tidak ditemukan")
	errReceivableHasPayment = apperror.Conflict("receivable_has_payment", "Piutang yang sudah dibayar sebagian atau penuh tidak bisa dihapus")
	errReceivableCustomer   = apperror.Unprocessable("customer_not_found", "Pelanggan tidak ditemukan")

	errPayableNotFound   = apperror.NotFound("payable_not_found", "Hutang tidak ditemukan")
	errPayableHasPayment = apperror.Conflict("payable_has_payment", "Hutang yang sudah dibayar sebagian atau penuh tidak bisa dihapus")
	errPayableVendor     = apperror.Unprocessable("vendor_not_found", "Vendor tidak ditemukan")

	errAmountBelowPaid  = apperror.Unprocessable("amount_below_paid", "Jumlah tagihan tidak boleh lebih kecil dari yang sudah dibayar")
	errPaymentOverdrawn = apperror.Unprocessable("payment_exceeds_outstanding", "Pembayaran melebihi sisa yang belum dibayar")
)

var (
	invoiceReadErrors  = database.ErrorMap{NotFound: errInvoiceNotFound}
	invoiceWriteErrors = database.ErrorMap{NotFound: errInvoiceNotFound, Duplicate: errInvoiceNumberUsed, Referenced: errInvoiceProject, Invalid: errPaymentOverdrawn}

	receivableReadErrors  = database.ErrorMap{NotFound: errReceivableNotFound}
	receivableWriteErrors = database.ErrorMap{NotFound: errReceivableNotFound, Referenced: errReceivableCustomer, Invalid: errPaymentOverdrawn}

	payableReadErrors  = database.ErrorMap{NotFound: errPayableNotFound}
	payableWriteErrors = database.ErrorMap{NotFound: errPayableNotFound, Referenced: errPayableVendor, Invalid: errPaymentOverdrawn}
)

type Service interface {
	ListInvoices(ctx context.Context, query InvoiceQuery) (pagination.Page[InvoiceResponse], error)
	GetInvoice(ctx context.Context, id uuid.UUID) (InvoiceResponse, error)
	CreateInvoice(ctx context.Context, request InvoiceRequest) (InvoiceResponse, error)
	UpdateInvoice(ctx context.Context, id uuid.UUID, request InvoiceRequest) (InvoiceResponse, error)
	DeleteInvoice(ctx context.Context, id uuid.UUID) error
	PayInvoice(ctx context.Context, id uuid.UUID, request PaymentRequest) (InvoiceResponse, error)

	ListReceivables(ctx context.Context, query ReceivableQuery) (pagination.Page[ReceivableResponse], error)
	GetReceivable(ctx context.Context, id uuid.UUID) (ReceivableResponse, error)
	CreateReceivable(ctx context.Context, request ReceivableRequest) (ReceivableResponse, error)
	UpdateReceivable(ctx context.Context, id uuid.UUID, request ReceivableRequest) (ReceivableResponse, error)
	DeleteReceivable(ctx context.Context, id uuid.UUID) error
	PayReceivable(ctx context.Context, id uuid.UUID, request PaymentRequest) (ReceivableResponse, error)

	ListPayables(ctx context.Context, query PayableQuery) (pagination.Page[PayableResponse], error)
	GetPayable(ctx context.Context, id uuid.UUID) (PayableResponse, error)
	CreatePayable(ctx context.Context, request PayableRequest) (PayableResponse, error)
	UpdatePayable(ctx context.Context, id uuid.UUID, request PayableRequest) (PayableResponse, error)
	DeletePayable(ctx context.Context, id uuid.UUID) error
	PayPayable(ctx context.Context, id uuid.UUID, request PaymentRequest) (PayableResponse, error)
}

type service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return &service{repository: repository, now: time.Now}
}

func (s *service) ListInvoices(ctx context.Context, query InvoiceQuery) (pagination.Page[InvoiceResponse], error) {
	today := s.today()
	invoices, total, err := s.repository.ListInvoices(ctx, InvoiceFilter{
		DueFilter: dueFilterOf(query.Status, today),
		Search:    query.Search,
		PartyType: query.PartyType,
		PartyID:   query.PartyID,
		ProjectID: query.ProjectID,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[InvoiceResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(invoices, func(invoice Invoice) InvoiceResponse {
		return newInvoiceResponse(invoice, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) GetInvoice(ctx context.Context, id uuid.UUID) (InvoiceResponse, error) {
	invoice, err := s.repository.FindInvoice(ctx, id)
	if err != nil {
		return InvoiceResponse{}, invoiceReadErrors.Resolve(err)
	}
	return newInvoiceResponse(invoice, s.today()), nil
}

func (s *service) CreateInvoice(ctx context.Context, request InvoiceRequest) (InvoiceResponse, error) {
	if err := s.ensureParty(ctx, request.PartyType, request.PartyID); err != nil {
		return InvoiceResponse{}, err
	}

	today := s.today()
	var invoice Invoice
	applyInvoiceRequest(&invoice, request, today)
	if err := s.repository.CreateInvoice(ctx, &invoice); err != nil {
		return InvoiceResponse{}, invoiceWriteErrors.Resolve(err)
	}
	return newInvoiceResponse(invoice, today), nil
}

func (s *service) UpdateInvoice(ctx context.Context, id uuid.UUID, request InvoiceRequest) (InvoiceResponse, error) {
	invoice, err := s.repository.FindInvoice(ctx, id)
	if err != nil {
		return InvoiceResponse{}, invoiceReadErrors.Resolve(err)
	}
	if request.Amount < invoice.PaidAmount {
		return InvoiceResponse{}, errAmountBelowPaid
	}
	if err := s.ensureParty(ctx, request.PartyType, request.PartyID); err != nil {
		return InvoiceResponse{}, err
	}

	today := s.today()
	applyInvoiceRequest(&invoice, request, today)
	if err := s.repository.SaveInvoice(ctx, &invoice); err != nil {
		return InvoiceResponse{}, invoiceWriteErrors.Resolve(err)
	}
	return newInvoiceResponse(invoice, today), nil
}

func (s *service) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	invoice, err := s.repository.FindInvoice(ctx, id)
	if err != nil {
		return invoiceReadErrors.Resolve(err)
	}
	if invoice.PaidAmount > 0 {
		return errInvoiceHasPayment
	}
	return invoiceReadErrors.Resolve(s.repository.DeleteInvoice(ctx, id))
}

func (s *service) PayInvoice(ctx context.Context, id uuid.UUID, request PaymentRequest) (InvoiceResponse, error) {
	today := s.today()
	var invoice Invoice
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		var err error
		invoice, err = payInvoice(ctx, repository, id, request.Amount, today)
		return err
	})
	if err != nil {
		return InvoiceResponse{}, apperror.From(err)
	}
	return newInvoiceResponse(invoice, today), nil
}

func (s *service) ListReceivables(ctx context.Context, query ReceivableQuery) (pagination.Page[ReceivableResponse], error) {
	today := s.today()
	receivables, total, err := s.repository.ListReceivables(ctx, ReceivableFilter{
		DueFilter:  dueFilterOf(query.Status, today),
		Search:     query.Search,
		CustomerID: query.CustomerID,
		Offset:     query.Offset(),
		Limit:      query.Size(),
	})
	if err != nil {
		return pagination.Page[ReceivableResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(receivables, func(receivable Receivable) ReceivableResponse {
		return newReceivableResponse(receivable, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) GetReceivable(ctx context.Context, id uuid.UUID) (ReceivableResponse, error) {
	receivable, err := s.repository.FindReceivable(ctx, id)
	if err != nil {
		return ReceivableResponse{}, receivableReadErrors.Resolve(err)
	}
	return newReceivableResponse(receivable, s.today()), nil
}

func (s *service) CreateReceivable(ctx context.Context, request ReceivableRequest) (ReceivableResponse, error) {
	today := s.today()
	var receivable Receivable
	applyReceivableRequest(&receivable, request, today)
	if err := s.repository.CreateReceivable(ctx, &receivable); err != nil {
		return ReceivableResponse{}, receivableWriteErrors.Resolve(err)
	}
	return newReceivableResponse(receivable, today), nil
}

func (s *service) UpdateReceivable(ctx context.Context, id uuid.UUID, request ReceivableRequest) (ReceivableResponse, error) {
	receivable, err := s.repository.FindReceivable(ctx, id)
	if err != nil {
		return ReceivableResponse{}, receivableReadErrors.Resolve(err)
	}
	if request.Amount < receivable.PaidAmount {
		return ReceivableResponse{}, errAmountBelowPaid
	}

	today := s.today()
	applyReceivableRequest(&receivable, request, today)
	if err := s.repository.SaveReceivable(ctx, &receivable); err != nil {
		return ReceivableResponse{}, receivableWriteErrors.Resolve(err)
	}
	return newReceivableResponse(receivable, today), nil
}

func (s *service) DeleteReceivable(ctx context.Context, id uuid.UUID) error {
	receivable, err := s.repository.FindReceivable(ctx, id)
	if err != nil {
		return receivableReadErrors.Resolve(err)
	}
	if receivable.PaidAmount > 0 {
		return errReceivableHasPayment
	}
	return receivableReadErrors.Resolve(s.repository.DeleteReceivable(ctx, id))
}

func (s *service) PayReceivable(ctx context.Context, id uuid.UUID, request PaymentRequest) (ReceivableResponse, error) {
	today := s.today()
	var receivable Receivable
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		var err error
		receivable, err = payReceivable(ctx, repository, id, request.Amount, today)
		return err
	})
	if err != nil {
		return ReceivableResponse{}, apperror.From(err)
	}
	return newReceivableResponse(receivable, today), nil
}

func (s *service) ListPayables(ctx context.Context, query PayableQuery) (pagination.Page[PayableResponse], error) {
	today := s.today()
	payables, total, err := s.repository.ListPayables(ctx, PayableFilter{
		DueFilter: dueFilterOf(query.Status, today),
		Search:    query.Search,
		VendorID:  query.VendorID,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[PayableResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(payables, func(payable Payable) PayableResponse {
		return newPayableResponse(payable, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) GetPayable(ctx context.Context, id uuid.UUID) (PayableResponse, error) {
	payable, err := s.repository.FindPayable(ctx, id)
	if err != nil {
		return PayableResponse{}, payableReadErrors.Resolve(err)
	}
	return newPayableResponse(payable, s.today()), nil
}

func (s *service) CreatePayable(ctx context.Context, request PayableRequest) (PayableResponse, error) {
	today := s.today()
	var payable Payable
	applyPayableRequest(&payable, request, today)
	if err := s.repository.CreatePayable(ctx, &payable); err != nil {
		return PayableResponse{}, payableWriteErrors.Resolve(err)
	}
	return newPayableResponse(payable, today), nil
}

func (s *service) UpdatePayable(ctx context.Context, id uuid.UUID, request PayableRequest) (PayableResponse, error) {
	payable, err := s.repository.FindPayable(ctx, id)
	if err != nil {
		return PayableResponse{}, payableReadErrors.Resolve(err)
	}
	if request.Amount < payable.PaidAmount {
		return PayableResponse{}, errAmountBelowPaid
	}

	today := s.today()
	applyPayableRequest(&payable, request, today)
	if err := s.repository.SavePayable(ctx, &payable); err != nil {
		return PayableResponse{}, payableWriteErrors.Resolve(err)
	}
	return newPayableResponse(payable, today), nil
}

func (s *service) DeletePayable(ctx context.Context, id uuid.UUID) error {
	payable, err := s.repository.FindPayable(ctx, id)
	if err != nil {
		return payableReadErrors.Resolve(err)
	}
	if payable.PaidAmount > 0 {
		return errPayableHasPayment
	}
	return payableReadErrors.Resolve(s.repository.DeletePayable(ctx, id))
}

func (s *service) PayPayable(ctx context.Context, id uuid.UUID, request PaymentRequest) (PayableResponse, error) {
	today := s.today()
	var payable Payable
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		var err error
		payable, err = payPayable(ctx, repository, id, request.Amount, today)
		return err
	})
	if err != nil {
		return PayableResponse{}, apperror.From(err)
	}
	return newPayableResponse(payable, today), nil
}

func (s *service) ensureParty(ctx context.Context, partyType string, partyID uuid.UUID) error {
	exists, err := s.repository.PartyExists(ctx, partyType, partyID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !exists {
		return errInvoicePartyMissing
	}
	return nil
}

func (s *service) today() time.Time {
	return duedate.Today(s.now())
}

func payInvoice(ctx context.Context, repository Repository, id uuid.UUID, payment int64, today time.Time) (Invoice, error) {
	invoice, err := repository.LockInvoice(ctx, id)
	if err != nil {
		return Invoice{}, invoiceReadErrors.Resolve(err)
	}
	paid, err := addPayment(invoice.Amount, invoice.PaidAmount, payment)
	if err != nil {
		return Invoice{}, err
	}

	invoice.PaidAmount = paid
	invoice.Status = storedStatus(invoice.DueDate, invoice.Amount, paid, today)
	if err := repository.SaveInvoice(ctx, &invoice); err != nil {
		return Invoice{}, invoiceWriteErrors.Resolve(err)
	}
	return invoice, nil
}

func payReceivable(ctx context.Context, repository Repository, id uuid.UUID, payment int64, today time.Time) (Receivable, error) {
	receivable, err := repository.LockReceivable(ctx, id)
	if err != nil {
		return Receivable{}, receivableReadErrors.Resolve(err)
	}
	paid, err := addPayment(receivable.Amount, receivable.PaidAmount, payment)
	if err != nil {
		return Receivable{}, err
	}

	receivable.PaidAmount = paid
	receivable.Status = storedStatus(receivable.DueDate, receivable.Amount, paid, today)
	if err := repository.SaveReceivable(ctx, &receivable); err != nil {
		return Receivable{}, receivableWriteErrors.Resolve(err)
	}
	return receivable, nil
}

func payPayable(ctx context.Context, repository Repository, id uuid.UUID, payment int64, today time.Time) (Payable, error) {
	payable, err := repository.LockPayable(ctx, id)
	if err != nil {
		return Payable{}, payableReadErrors.Resolve(err)
	}
	paid, err := addPayment(payable.Amount, payable.PaidAmount, payment)
	if err != nil {
		return Payable{}, err
	}

	payable.PaidAmount = paid
	payable.Status = storedStatus(payable.DueDate, payable.Amount, paid, today)
	if err := repository.SavePayable(ctx, &payable); err != nil {
		return Payable{}, payableWriteErrors.Resolve(err)
	}
	return payable, nil
}

func addPayment(amount, paid, payment int64) (int64, error) {
	if paid+payment > amount {
		return 0, errPaymentOverdrawn
	}
	return paid + payment, nil
}

func settledAmount(amount, paid int64) bool {
	return amount > 0 && paid >= amount
}

func storedStatus(dueDate time.Time, amount, paid int64, today time.Time) string {
	return duedate.Status(dueDate, settledAmount(amount, paid), today)
}

func dueFilterOf(status string, today time.Time) DueFilter {
	window := duedate.RangeOf(status, today)
	return DueFilter{Settled: window.Settled, DueFrom: window.From, DueBefore: window.Before}
}

func applyInvoiceRequest(invoice *Invoice, request InvoiceRequest, today time.Time) {
	invoice.Number = strings.TrimSpace(request.Number)
	invoice.Note = strings.TrimSpace(request.Note)
	invoice.PartyType = request.PartyType
	invoice.PartyID = request.PartyID
	invoice.ProjectID = request.ProjectID
	invoice.DueDate = request.DueDate
	invoice.Amount = request.Amount
	invoice.Status = storedStatus(request.DueDate, request.Amount, invoice.PaidAmount, today)
}

func applyReceivableRequest(receivable *Receivable, request ReceivableRequest, today time.Time) {
	receivable.CustomerID = request.CustomerID
	receivable.Reference = strings.TrimSpace(request.Reference)
	receivable.DueDate = request.DueDate
	receivable.Amount = request.Amount
	receivable.Status = storedStatus(request.DueDate, request.Amount, receivable.PaidAmount, today)
}

func applyPayableRequest(payable *Payable, request PayableRequest, today time.Time) {
	payable.VendorID = request.VendorID
	payable.Reference = strings.TrimSpace(request.Reference)
	payable.DueDate = request.DueDate
	payable.Amount = request.Amount
	payable.Status = storedStatus(request.DueDate, request.Amount, payable.PaidAmount, today)
}
