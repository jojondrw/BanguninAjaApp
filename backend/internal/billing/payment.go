package billing

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	defaultPaymentMethod = "transfer"
	jakartaOffsetSeconds = 7 * 60 * 60
	thousandsGroup       = 3
)

var jakarta = time.FixedZone("Asia/Jakarta", jakartaOffsetSeconds)

var (
	errPaidAtInFuture   = apperror.Unprocessable("payment_date_in_future", "Tanggal bayar tidak boleh setelah hari ini")
	errLinkChanged      = apperror.Conflict("billing_link_changed", "Tautan faktur dan piutang ini baru saja berubah. Muat ulang lalu coba lagi")
	errLinkedHasPayment = apperror.Unprocessable("receivable_link_has_payment", "Piutang yang sudah dibayar tidak bisa dilepas dari fakturnya atau dipindah ke faktur lain")
	errPaymentParent    = apperror.Unprocessable("payment_parent_not_found", "Faktur, piutang, atau utang yang dibayar sudah tidak ada")

	paymentWriteErrors = database.ErrorMap{Referenced: errPaymentParent, Invalid: errPaymentOverdrawn}
)

type alignment int

const (
	alreadyAligned alignment = iota
	copiedToReceivable
	copiedToInvoice
)

type paymentParents struct {
	invoice    *Invoice
	receivable *Receivable
	payable    *Payable
}

func linkedOverpaid(noun, label string, outstanding int64) *apperror.Error {
	return apperror.Unprocessable("payment_exceeds_outstanding", fmt.Sprintf(
		"Pembayaran melebihi sisa %s %s yang tertaut, sisanya tinggal %s", noun, label, rupiah(outstanding),
	))
}

func linkPaidMismatch(invoice Invoice, receivable Receivable) *apperror.Error {
	return apperror.Unprocessable("link_paid_mismatch", fmt.Sprintf(
		"Faktur %s sudah dibayar %s, sedangkan piutang ini %s. Keduanya hanya bisa ditautkan kalau total bayarnya sama atau salah satunya belum dibayar",
		invoice.Number, rupiah(invoice.PaidAmount), rupiah(receivable.PaidAmount),
	))
}

func receivableBelowInvoicePaid(invoice Invoice) *apperror.Error {
	return apperror.Unprocessable("receivable_below_invoice_paid", fmt.Sprintf(
		"Faktur %s sudah dibayar %s, jadi nilai piutangnya minimal sebesar itu", invoice.Number, rupiah(invoice.PaidAmount),
	))
}

func rupiah(value int64) string {
	digits := strconv.FormatInt(value, 10)
	var grouped strings.Builder
	for index, digit := range digits {
		if index > 0 && (len(digits)-index)%thousandsGroup == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(digit)
	}
	return "Rp " + grouped.String()
}

func jakartaToday(now time.Time) time.Time {
	return duedate.Today(now.In(jakarta))
}

func paymentDate(requested *time.Time, today time.Time) (time.Time, error) {
	if requested == nil {
		return today, nil
	}
	date := duedate.Today(*requested)
	if date.After(today) {
		return time.Time{}, errPaidAtInFuture
	}
	return date, nil
}

func paymentMethod(requested string) string {
	if requested == "" {
		return defaultPaymentMethod
	}
	return requested
}

func sameID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func lockInvoicePayment(ctx context.Context, repository Repository, id uuid.UUID) (paymentParents, error) {
	invoice, err := repository.LockInvoice(ctx, id)
	if err != nil {
		return paymentParents{}, invoiceReadErrors.Resolve(err)
	}
	parents := paymentParents{invoice: &invoice}
	linked, err := receivableOfInvoice(ctx, repository, id)
	if err != nil || linked == nil {
		return parents, err
	}
	receivable, err := repository.LockReceivable(ctx, linked.ID)
	if err != nil {
		return paymentParents{}, receivableReadErrors.Resolve(err)
	}
	if !sameID(receivable.InvoiceID, &invoice.ID) {
		return paymentParents{}, errLinkChanged
	}
	parents.receivable = &receivable
	return parents, nil
}

func lockReceivablePayment(ctx context.Context, repository Repository, id uuid.UUID) (paymentParents, error) {
	current, err := repository.FindReceivable(ctx, id)
	if err != nil {
		return paymentParents{}, receivableReadErrors.Resolve(err)
	}
	var parents paymentParents
	if current.InvoiceID != nil {
		invoice, err := repository.LockInvoice(ctx, *current.InvoiceID)
		if err != nil {
			return paymentParents{}, invoiceReadErrors.Resolve(err)
		}
		parents.invoice = &invoice
	}
	receivable, err := repository.LockReceivable(ctx, id)
	if err != nil {
		return paymentParents{}, receivableReadErrors.Resolve(err)
	}
	if !sameID(receivable.InvoiceID, current.InvoiceID) {
		return paymentParents{}, errLinkChanged
	}
	parents.receivable = &receivable
	return parents, nil
}

func lockPayablePayment(ctx context.Context, repository Repository, id uuid.UUID) (paymentParents, error) {
	payable, err := repository.LockPayable(ctx, id)
	if err != nil {
		return paymentParents{}, payableReadErrors.Resolve(err)
	}
	return paymentParents{payable: &payable}, nil
}

func settlePayment(ctx context.Context, repository Repository, parents paymentParents, request PaymentRequest, actor *uuid.UUID, now time.Time) error {
	paidAt, err := paymentDate(request.PaidAt, jakartaToday(now))
	if err != nil {
		return err
	}
	if err := parents.credit(request.Amount, duedate.Today(now)); err != nil {
		return err
	}
	payment := parents.newPayment(request, actor, paidAt)
	if err := repository.CreatePayment(ctx, &payment); err != nil {
		return paymentWriteErrors.Resolve(err)
	}
	return parents.save(ctx, repository)
}

func (p paymentParents) isLinked() bool {
	return p.invoice != nil && p.receivable != nil
}

func (p paymentParents) credit(amount int64, today time.Time) error {
	if p.invoice != nil {
		if err := creditInvoice(p.invoice, amount, today, p.isLinked()); err != nil {
			return err
		}
	}
	if p.receivable != nil {
		if err := creditReceivable(p.receivable, amount, today, p.isLinked()); err != nil {
			return err
		}
	}
	if p.payable != nil {
		paid, err := addPayment(p.payable.Amount, p.payable.PaidAmount, amount)
		if err != nil {
			return err
		}
		p.payable.PaidAmount = paid
		p.payable.Status = storedStatus(p.payable.DueDate, p.payable.Amount, paid, today)
	}
	return nil
}

func creditInvoice(invoice *Invoice, amount int64, today time.Time, linked bool) error {
	paid, err := addPayment(invoice.Amount, invoice.PaidAmount, amount)
	if err != nil && linked {
		return linkedOverpaid("faktur", invoice.Number, invoice.Amount-invoice.PaidAmount)
	}
	if err != nil {
		return err
	}
	invoice.PaidAmount = paid
	invoice.Status = storedStatus(invoice.DueDate, invoice.Amount, paid, today)
	return nil
}

func creditReceivable(receivable *Receivable, amount int64, today time.Time, linked bool) error {
	paid, err := addPayment(receivable.Amount, receivable.PaidAmount, amount)
	if err != nil && linked {
		return linkedOverpaid("piutang", receivable.Reference, receivable.Amount-receivable.PaidAmount)
	}
	if err != nil {
		return err
	}
	receivable.PaidAmount = paid
	receivable.Status = storedStatus(receivable.DueDate, receivable.Amount, paid, today)
	return nil
}

func (p paymentParents) newPayment(request PaymentRequest, actor *uuid.UUID, paidAt time.Time) Payment {
	payment := Payment{
		Amount:    request.Amount,
		PaidAt:    paidAt,
		Method:    paymentMethod(request.Method),
		Reference: strings.TrimSpace(request.Reference),
		Note:      strings.TrimSpace(request.Note),
		CreatedBy: actor,
	}
	if p.invoice != nil {
		payment.InvoiceID = &p.invoice.ID
	}
	if p.receivable != nil {
		payment.ReceivableID = &p.receivable.ID
	}
	if p.payable != nil {
		payment.PayableID = &p.payable.ID
	}
	return payment
}

func (p paymentParents) save(ctx context.Context, repository Repository) error {
	if p.invoice != nil {
		if err := repository.SaveInvoice(ctx, p.invoice); err != nil {
			return invoiceWriteErrors.Resolve(err)
		}
	}
	if p.receivable != nil {
		if err := repository.SaveReceivable(ctx, p.receivable); err != nil {
			return receivableWriteErrors.Resolve(err)
		}
	}
	if p.payable != nil {
		return payableWriteErrors.Resolve(repository.SavePayable(ctx, p.payable))
	}
	return nil
}

func lockRequestedInvoice(ctx context.Context, repository Repository, invoiceID *uuid.UUID) (*Invoice, error) {
	if invoiceID == nil {
		return nil, nil
	}
	invoice, err := repository.LockInvoice(ctx, *invoiceID)
	if err != nil {
		return nil, database.ErrorMap{NotFound: errLinkedInvoiceNotFound}.Resolve(err)
	}
	return &invoice, nil
}

func alignLinkedPaid(invoice *Invoice, receivable *Receivable, today time.Time) (alignment, error) {
	switch {
	case invoice.PaidAmount == receivable.PaidAmount:
		return alreadyAligned, nil
	case invoice.PaidAmount > 0 && receivable.PaidAmount > 0:
		return alreadyAligned, linkPaidMismatch(*invoice, *receivable)
	case invoice.PaidAmount > 0:
		if invoice.PaidAmount > receivable.Amount {
			return alreadyAligned, receivableBelowInvoicePaid(*invoice)
		}
		receivable.PaidAmount = invoice.PaidAmount
		receivable.Status = storedStatus(receivable.DueDate, receivable.Amount, receivable.PaidAmount, today)
		return copiedToReceivable, nil
	default:
		if receivable.PaidAmount > invoice.Amount {
			return alreadyAligned, errReceivableExceeds
		}
		invoice.PaidAmount = receivable.PaidAmount
		invoice.Status = storedStatus(invoice.DueDate, invoice.Amount, invoice.PaidAmount, today)
		return copiedToInvoice, nil
	}
}

func applyAlignment(ctx context.Context, repository Repository, invoice *Invoice, receivable *Receivable, aligned alignment) error {
	if aligned == alreadyAligned {
		return nil
	}
	if aligned == copiedToInvoice {
		if err := repository.SaveInvoice(ctx, invoice); err != nil {
			return invoiceWriteErrors.Resolve(err)
		}
	}
	if err := repository.LinkPayments(ctx, invoice.ID, receivable.ID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func listPayments(ctx context.Context, repository Repository, column string, parentID uuid.UUID, query PaymentQuery) (pagination.Page[PaymentResponse], error) {
	rows, total, err := repository.ListPayments(ctx, PaymentFilter{
		Column:   column,
		ParentID: parentID,
		Offset:   query.Offset(),
		Limit:    query.Size(),
	})
	if err != nil {
		return pagination.Page[PaymentResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newPaymentResponse), query.Query, total), nil
}
