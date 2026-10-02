package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	customerParty = "customer"
	vendorParty   = "vendor"

	invoiceRecordedCode = "invoice_already_recorded"
)

var (
	errInvoiceNotFound      = apperror.NotFound("invoice_not_found", "Tagihan tidak ditemukan")
	errInvoiceNumberUsed    = apperror.Conflict("invoice_number_used", "Nomor tagihan sudah dipakai")
	errInvoiceHasPayment    = apperror.Conflict("invoice_has_payment", "Tagihan yang sudah dibayar sebagian atau penuh tidak bisa dihapus")
	errInvoicePartyMissing  = apperror.Unprocessable("invoice_party_not_found", "Pelanggan atau vendor yang ditagih tidak ditemukan")
	errProjectNotFound      = apperror.Unprocessable("project_not_found", "Proyek tidak ditemukan")
	errInvoiceRecorded      = apperror.Conflict(invoiceRecordedCode, "Faktur ini sudah dicatat sebagai piutang lain")
	errInvoiceHasReceivable = apperror.Conflict("invoice_has_receivable", "Faktur ini sudah dicatat sebagai piutang. Hapus piutangnya dulu")

	errReceivableNotFound    = apperror.NotFound("receivable_not_found", "Piutang tidak ditemukan")
	errReceivableHasPayment  = apperror.Conflict("receivable_has_payment", "Piutang yang sudah dibayar sebagian atau penuh tidak bisa dihapus")
	errReceivableCustomer    = apperror.Unprocessable("customer_not_found", "Pelanggan tidak ditemukan")
	errReceivableLinkMissing = apperror.Unprocessable("receivable_link_not_found", "Pelanggan, proyek, kontrak, atau faktur yang dipilih sudah tidak ada")
	errContractNotFound      = apperror.Unprocessable("contract_not_found", "Kontrak penjualan tidak ditemukan")
	errContractCustomer      = apperror.Unprocessable("contract_customer_mismatch", "Kontrak penjualan ini milik pelanggan lain, bukan pelanggan piutang ini")
	errLinkedInvoiceNotFound = apperror.Unprocessable("linked_invoice_not_found", "Faktur yang dihubungkan tidak ditemukan")
	errInvoiceCustomer       = apperror.Unprocessable("invoice_customer_mismatch", "Faktur ini tidak ditagihkan ke pelanggan piutang ini")
	errReceivableExceeds     = apperror.Unprocessable("receivable_exceeds_invoice", "Nilai piutang tidak boleh melebihi nilai faktur")

	errPayableNotFound      = apperror.NotFound("payable_not_found", "Hutang tidak ditemukan")
	errPayableHasPayment    = apperror.Conflict("payable_has_payment", "Hutang yang sudah dibayar sebagian atau penuh tidak bisa dihapus")
	errPayableVendor        = apperror.Unprocessable("vendor_not_found", "Vendor tidak ditemukan")
	errPayableLinkMissing   = apperror.Unprocessable("payable_link_not_found", "Vendor, proyek, atau pesanan pembelian yang dipilih sudah tidak ada")
	errPurchaseOrderMissing = apperror.Unprocessable("purchase_order_not_found", "Pesanan pembelian tidak ditemukan")
	errPurchaseOrderVendor  = apperror.Unprocessable("purchase_order_vendor_mismatch", "Pesanan pembelian ini milik vendor lain, bukan vendor utang ini")

	errAmountBelowPaid  = apperror.Unprocessable("amount_below_paid", "Jumlah tagihan tidak boleh lebih kecil dari yang sudah dibayar")
	errPaymentOverdrawn = apperror.Unprocessable("payment_exceeds_outstanding", "Pembayaran melebihi sisa yang belum dibayar")
)

var (
	invoiceReadErrors   = database.ErrorMap{NotFound: errInvoiceNotFound}
	invoiceWriteErrors  = database.ErrorMap{NotFound: errInvoiceNotFound, Duplicate: errInvoiceNumberUsed, Referenced: errProjectNotFound, Invalid: errPaymentOverdrawn}
	invoiceDeleteErrors = database.ErrorMap{NotFound: errInvoiceNotFound, Referenced: errInvoiceHasReceivable}

	receivableReadErrors  = database.ErrorMap{NotFound: errReceivableNotFound}
	receivableWriteErrors = database.ErrorMap{NotFound: errReceivableNotFound, Duplicate: errInvoiceRecorded, Referenced: errReceivableLinkMissing, Invalid: errPaymentOverdrawn}

	payableReadErrors  = database.ErrorMap{NotFound: errPayableNotFound}
	payableWriteErrors = database.ErrorMap{NotFound: errPayableNotFound, Referenced: errPayableLinkMissing, Invalid: errPaymentOverdrawn}
)

func invoiceRecordedAs(reference string) *apperror.Error {
	return apperror.Conflict(invoiceRecordedCode, fmt.Sprintf("Faktur ini sudah dicatat sebagai piutang %s", reference))
}

func invoiceHasReceivable(reference string) *apperror.Error {
	return apperror.Conflict("invoice_has_receivable", fmt.Sprintf("Faktur ini sudah dicatat sebagai piutang %s. Hapus piutangnya dulu", reference))
}

func invoiceLockedByReceivable(reference string) *apperror.Error {
	return apperror.Unprocessable("invoice_receivable_mismatch", fmt.Sprintf(
		"Faktur ini sudah dicatat sebagai piutang %s, jadi pelanggannya harus tetap sama dan nilainya tidak boleh di bawah nilai piutang", reference,
	))
}

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
	rows, total, err := s.repository.ListInvoices(ctx, InvoiceFilter{
		DueFilter: dueFilterOf(query.Status, today),
		Search:    query.Search,
		PartyType: query.PartyType,
		PartyID:   query.PartyID,
		ProjectID: query.ProjectID,
		Recorded:  query.Recorded,
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[InvoiceResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(rows, func(row InvoiceRow) InvoiceResponse {
		return newInvoiceResponse(row, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) GetInvoice(ctx context.Context, id uuid.UUID) (InvoiceResponse, error) {
	return invoiceResponse(ctx, s.repository, id, s.today())
}

func (s *service) CreateInvoice(ctx context.Context, request InvoiceRequest) (InvoiceResponse, error) {
	if err := ensureParty(ctx, s.repository, request.PartyType, request.PartyID, errInvoicePartyMissing); err != nil {
		return InvoiceResponse{}, err
	}

	today := s.today()
	var invoice Invoice
	applyInvoiceRequest(&invoice, request, today)
	if err := s.repository.CreateInvoice(ctx, &invoice); err != nil {
		return InvoiceResponse{}, invoiceWriteErrors.Resolve(err)
	}
	return invoiceResponse(ctx, s.repository, invoice.ID, today)
}

func (s *service) UpdateInvoice(ctx context.Context, id uuid.UUID, request InvoiceRequest) (InvoiceResponse, error) {
	today := s.today()
	var response InvoiceResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		invoice, err := repository.LockInvoice(ctx, id)
		if err != nil {
			return invoiceReadErrors.Resolve(err)
		}
		if request.Amount < invoice.PaidAmount {
			return errAmountBelowPaid
		}
		if err := ensureParty(ctx, repository, request.PartyType, request.PartyID, errInvoicePartyMissing); err != nil {
			return err
		}
		if err := ensureInvoiceStillCoversReceivable(ctx, repository, id, request); err != nil {
			return err
		}
		applyInvoiceRequest(&invoice, request, today)
		if err := repository.SaveInvoice(ctx, &invoice); err != nil {
			return invoiceWriteErrors.Resolve(err)
		}
		response, err = invoiceResponse(ctx, repository, id, today)
		return err
	})
	if err != nil {
		return InvoiceResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		invoice, err := repository.LockInvoice(ctx, id)
		if err != nil {
			return invoiceReadErrors.Resolve(err)
		}
		if invoice.PaidAmount > 0 {
			return errInvoiceHasPayment
		}
		receivable, err := receivableOfInvoice(ctx, repository, id)
		if err != nil {
			return err
		}
		if receivable != nil {
			return invoiceHasReceivable(receivable.Reference)
		}
		return invoiceDeleteErrors.Resolve(repository.DeleteInvoice(ctx, id))
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) PayInvoice(ctx context.Context, id uuid.UUID, request PaymentRequest) (InvoiceResponse, error) {
	today := s.today()
	var response InvoiceResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		if err := payInvoice(ctx, repository, id, request.Amount, today); err != nil {
			return err
		}
		var err error
		response, err = invoiceResponse(ctx, repository, id, today)
		return err
	})
	if err != nil {
		return InvoiceResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) ListReceivables(ctx context.Context, query ReceivableQuery) (pagination.Page[ReceivableResponse], error) {
	today := s.today()
	rows, total, err := s.repository.ListReceivables(ctx, ReceivableFilter{
		DueFilter:  dueFilterOf(query.Status, today),
		Search:     query.Search,
		CustomerID: query.CustomerID,
		ProjectID:  query.ProjectID,
		ContractID: query.ContractID,
		Offset:     query.Offset(),
		Limit:      query.Size(),
	})
	if err != nil {
		return pagination.Page[ReceivableResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(rows, func(row ReceivableRow) ReceivableResponse {
		return newReceivableResponse(row, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) GetReceivable(ctx context.Context, id uuid.UUID) (ReceivableResponse, error) {
	return receivableResponse(ctx, s.repository, id, s.today())
}

func (s *service) CreateReceivable(ctx context.Context, request ReceivableRequest) (ReceivableResponse, error) {
	if err := ensureReceivableLinks(ctx, s.repository, request, uuid.Nil); err != nil {
		return ReceivableResponse{}, err
	}

	today := s.today()
	var receivable Receivable
	applyReceivableRequest(&receivable, request, today)
	if err := s.repository.CreateReceivable(ctx, &receivable); err != nil {
		return ReceivableResponse{}, receivableWriteErrors.Resolve(err)
	}
	return receivableResponse(ctx, s.repository, receivable.ID, today)
}

func (s *service) UpdateReceivable(ctx context.Context, id uuid.UUID, request ReceivableRequest) (ReceivableResponse, error) {
	today := s.today()
	var response ReceivableResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		receivable, err := repository.LockReceivable(ctx, id)
		if err != nil {
			return receivableReadErrors.Resolve(err)
		}
		if request.Amount < receivable.PaidAmount {
			return errAmountBelowPaid
		}
		if err := ensureReceivableLinks(ctx, repository, request, id); err != nil {
			return err
		}
		applyReceivableRequest(&receivable, request, today)
		if err := repository.SaveReceivable(ctx, &receivable); err != nil {
			return receivableWriteErrors.Resolve(err)
		}
		response, err = receivableResponse(ctx, repository, id, today)
		return err
	})
	if err != nil {
		return ReceivableResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) DeleteReceivable(ctx context.Context, id uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		receivable, err := repository.LockReceivable(ctx, id)
		if err != nil {
			return receivableReadErrors.Resolve(err)
		}
		if receivable.PaidAmount > 0 {
			return errReceivableHasPayment
		}
		return receivableReadErrors.Resolve(repository.DeleteReceivable(ctx, id))
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) PayReceivable(ctx context.Context, id uuid.UUID, request PaymentRequest) (ReceivableResponse, error) {
	today := s.today()
	var response ReceivableResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		if err := payReceivable(ctx, repository, id, request.Amount, today); err != nil {
			return err
		}
		var err error
		response, err = receivableResponse(ctx, repository, id, today)
		return err
	})
	if err != nil {
		return ReceivableResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) ListPayables(ctx context.Context, query PayableQuery) (pagination.Page[PayableResponse], error) {
	today := s.today()
	rows, total, err := s.repository.ListPayables(ctx, PayableFilter{
		DueFilter:       dueFilterOf(query.Status, today),
		Search:          query.Search,
		VendorID:        query.VendorID,
		ProjectID:       query.ProjectID,
		PurchaseOrderID: query.PurchaseOrderID,
		Offset:          query.Offset(),
		Limit:           query.Size(),
	})
	if err != nil {
		return pagination.Page[PayableResponse]{}, apperror.Internal(err)
	}
	responses := pagination.Map(rows, func(row PayableRow) PayableResponse {
		return newPayableResponse(row, today)
	})
	return pagination.New(responses, query.Query, total), nil
}

func (s *service) GetPayable(ctx context.Context, id uuid.UUID) (PayableResponse, error) {
	return payableResponse(ctx, s.repository, id, s.today())
}

func (s *service) CreatePayable(ctx context.Context, request PayableRequest) (PayableResponse, error) {
	if err := ensurePayableLinks(ctx, s.repository, request); err != nil {
		return PayableResponse{}, err
	}

	today := s.today()
	var payable Payable
	applyPayableRequest(&payable, request, today)
	if err := s.repository.CreatePayable(ctx, &payable); err != nil {
		return PayableResponse{}, payableWriteErrors.Resolve(err)
	}
	return payableResponse(ctx, s.repository, payable.ID, today)
}

func (s *service) UpdatePayable(ctx context.Context, id uuid.UUID, request PayableRequest) (PayableResponse, error) {
	today := s.today()
	var response PayableResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		payable, err := repository.LockPayable(ctx, id)
		if err != nil {
			return payableReadErrors.Resolve(err)
		}
		if request.Amount < payable.PaidAmount {
			return errAmountBelowPaid
		}
		if err := ensurePayableLinks(ctx, repository, request); err != nil {
			return err
		}
		applyPayableRequest(&payable, request, today)
		if err := repository.SavePayable(ctx, &payable); err != nil {
			return payableWriteErrors.Resolve(err)
		}
		response, err = payableResponse(ctx, repository, id, today)
		return err
	})
	if err != nil {
		return PayableResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) DeletePayable(ctx context.Context, id uuid.UUID) error {
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		payable, err := repository.LockPayable(ctx, id)
		if err != nil {
			return payableReadErrors.Resolve(err)
		}
		if payable.PaidAmount > 0 {
			return errPayableHasPayment
		}
		return payableReadErrors.Resolve(repository.DeletePayable(ctx, id))
	})
	if err != nil {
		return apperror.From(err)
	}
	return nil
}

func (s *service) PayPayable(ctx context.Context, id uuid.UUID, request PaymentRequest) (PayableResponse, error) {
	today := s.today()
	var response PayableResponse
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		if err := payPayable(ctx, repository, id, request.Amount, today); err != nil {
			return err
		}
		var err error
		response, err = payableResponse(ctx, repository, id, today)
		return err
	})
	if err != nil {
		return PayableResponse{}, apperror.From(err)
	}
	return response, nil
}

func (s *service) today() time.Time {
	return duedate.Today(s.now())
}

func invoiceResponse(ctx context.Context, repository Repository, id uuid.UUID, today time.Time) (InvoiceResponse, error) {
	row, err := repository.FindInvoiceRow(ctx, id)
	if err != nil {
		return InvoiceResponse{}, invoiceReadErrors.Resolve(err)
	}
	return newInvoiceResponse(row, today), nil
}

func receivableResponse(ctx context.Context, repository Repository, id uuid.UUID, today time.Time) (ReceivableResponse, error) {
	row, err := repository.FindReceivableRow(ctx, id)
	if err != nil {
		return ReceivableResponse{}, receivableReadErrors.Resolve(err)
	}
	return newReceivableResponse(row, today), nil
}

func payableResponse(ctx context.Context, repository Repository, id uuid.UUID, today time.Time) (PayableResponse, error) {
	row, err := repository.FindPayableRow(ctx, id)
	if err != nil {
		return PayableResponse{}, payableReadErrors.Resolve(err)
	}
	return newPayableResponse(row, today), nil
}

func ensureParty(ctx context.Context, repository Repository, partyType string, partyID uuid.UUID, missing *apperror.Error) error {
	exists, err := repository.PartyExists(ctx, partyType, partyID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !exists {
		return missing
	}
	return nil
}

func ensureProject(ctx context.Context, repository Repository, projectID *uuid.UUID) error {
	if projectID == nil {
		return nil
	}
	exists, err := repository.ProjectExists(ctx, *projectID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !exists {
		return errProjectNotFound
	}
	return nil
}

func ensureReceivableLinks(ctx context.Context, repository Repository, request ReceivableRequest, selfID uuid.UUID) error {
	if err := ensureParty(ctx, repository, customerParty, request.CustomerID, errReceivableCustomer); err != nil {
		return err
	}
	if err := ensureProject(ctx, repository, request.ProjectID); err != nil {
		return err
	}
	if request.ContractID != nil {
		contract, err := repository.FindContract(ctx, *request.ContractID)
		if err != nil {
			return database.ErrorMap{NotFound: errContractNotFound}.Resolve(err)
		}
		if contract.CustomerID != request.CustomerID {
			return errContractCustomer
		}
	}
	if request.InvoiceID == nil {
		return nil
	}

	invoice, err := repository.FindInvoice(ctx, *request.InvoiceID)
	if err != nil {
		return database.ErrorMap{NotFound: errLinkedInvoiceNotFound}.Resolve(err)
	}
	if invoice.PartyType != customerParty || invoice.PartyID != request.CustomerID {
		return errInvoiceCustomer
	}
	if request.Amount > invoice.Amount {
		return errReceivableExceeds
	}
	recorded, err := receivableOfInvoice(ctx, repository, invoice.ID)
	if err != nil {
		return err
	}
	if recorded != nil && (selfID == uuid.Nil || recorded.ID != selfID) {
		return invoiceRecordedAs(recorded.Reference)
	}
	return nil
}

func ensurePayableLinks(ctx context.Context, repository Repository, request PayableRequest) error {
	if err := ensureParty(ctx, repository, vendorParty, request.VendorID, errPayableVendor); err != nil {
		return err
	}
	if err := ensureProject(ctx, repository, request.ProjectID); err != nil {
		return err
	}
	if request.PurchaseOrderID == nil {
		return nil
	}
	order, err := repository.FindPurchaseOrder(ctx, *request.PurchaseOrderID)
	if err != nil {
		return database.ErrorMap{NotFound: errPurchaseOrderMissing}.Resolve(err)
	}
	if order.VendorID != request.VendorID {
		return errPurchaseOrderVendor
	}
	return nil
}

func ensureInvoiceStillCoversReceivable(ctx context.Context, repository Repository, invoiceID uuid.UUID, request InvoiceRequest) error {
	receivable, err := receivableOfInvoice(ctx, repository, invoiceID)
	if err != nil || receivable == nil {
		return err
	}
	if request.PartyType != customerParty || request.PartyID != receivable.CustomerID || request.Amount < receivable.Amount {
		return invoiceLockedByReceivable(receivable.Reference)
	}
	return nil
}

func receivableOfInvoice(ctx context.Context, repository Repository, invoiceID uuid.UUID) (*Receivable, error) {
	receivable, err := repository.FindReceivableByInvoice(ctx, invoiceID)
	if errors.Is(err, database.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &receivable, nil
}

func payInvoice(ctx context.Context, repository Repository, id uuid.UUID, payment int64, today time.Time) error {
	invoice, err := repository.LockInvoice(ctx, id)
	if err != nil {
		return invoiceReadErrors.Resolve(err)
	}
	paid, err := addPayment(invoice.Amount, invoice.PaidAmount, payment)
	if err != nil {
		return err
	}

	invoice.PaidAmount = paid
	invoice.Status = storedStatus(invoice.DueDate, invoice.Amount, paid, today)
	return invoiceWriteErrors.Resolve(repository.SaveInvoice(ctx, &invoice))
}

func payReceivable(ctx context.Context, repository Repository, id uuid.UUID, payment int64, today time.Time) error {
	receivable, err := repository.LockReceivable(ctx, id)
	if err != nil {
		return receivableReadErrors.Resolve(err)
	}
	paid, err := addPayment(receivable.Amount, receivable.PaidAmount, payment)
	if err != nil {
		return err
	}

	receivable.PaidAmount = paid
	receivable.Status = storedStatus(receivable.DueDate, receivable.Amount, paid, today)
	return receivableWriteErrors.Resolve(repository.SaveReceivable(ctx, &receivable))
}

func payPayable(ctx context.Context, repository Repository, id uuid.UUID, payment int64, today time.Time) error {
	payable, err := repository.LockPayable(ctx, id)
	if err != nil {
		return payableReadErrors.Resolve(err)
	}
	paid, err := addPayment(payable.Amount, payable.PaidAmount, payment)
	if err != nil {
		return err
	}

	payable.PaidAmount = paid
	payable.Status = storedStatus(payable.DueDate, payable.Amount, paid, today)
	return payableWriteErrors.Resolve(repository.SavePayable(ctx, &payable))
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
	receivable.ProjectID = request.ProjectID
	receivable.ContractID = request.ContractID
	receivable.InvoiceID = request.InvoiceID
	receivable.Reference = strings.TrimSpace(request.Reference)
	receivable.DueDate = request.DueDate
	receivable.Amount = request.Amount
	receivable.Status = storedStatus(request.DueDate, request.Amount, receivable.PaidAmount, today)
}

func applyPayableRequest(payable *Payable, request PayableRequest, today time.Time) {
	payable.VendorID = request.VendorID
	payable.ProjectID = request.ProjectID
	payable.PurchaseOrderID = request.PurchaseOrderID
	payable.Reference = strings.TrimSpace(request.Reference)
	payable.DueDate = request.DueDate
	payable.Amount = request.Amount
	payable.Status = storedStatus(request.DueDate, request.Amount, payable.PaidAmount, today)
}
