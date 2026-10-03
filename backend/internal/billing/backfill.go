package billing

import (
	"context"
	"fmt"
)

const (
	backfillMethod = "lainnya"
	backfillNote   = "Saldo terbayar sebelum riwayat pembayaran dicatat"

	linkPaymentsStatement = `UPDATE billing_payment SET invoice_id = ?, receivable_id = ?, updated_at = now()
WHERE (invoice_id = ? AND receivable_id IS NULL) OR (receivable_id = ? AND invoice_id IS NULL)`

	backfillLinkedPairs = `INSERT INTO billing_payment (id, created_at, updated_at, invoice_id, receivable_id, amount, paid_at, method, reference, note)
SELECT gen_random_uuid(), now(), now(), invoice.id, receivable.id, invoice.paid_amount,
       GREATEST(invoice.updated_at, receivable.updated_at)::date, @method, '', @note
FROM receivable
JOIN invoice ON invoice.id = receivable.invoice_id
WHERE invoice.paid_amount > 0
  AND invoice.paid_amount = receivable.paid_amount
  AND NOT EXISTS (
    SELECT 1 FROM billing_payment
    WHERE billing_payment.invoice_id = invoice.id OR billing_payment.receivable_id = receivable.id
  )`

	backfillInvoices = `INSERT INTO billing_payment (id, created_at, updated_at, invoice_id, amount, paid_at, method, reference, note)
SELECT gen_random_uuid(), now(), now(), invoice.id, invoice.paid_amount, invoice.updated_at::date, @method, '', @note
FROM invoice
WHERE invoice.paid_amount > 0
  AND NOT EXISTS (SELECT 1 FROM billing_payment WHERE billing_payment.invoice_id = invoice.id)`

	backfillReceivables = `INSERT INTO billing_payment (id, created_at, updated_at, receivable_id, amount, paid_at, method, reference, note)
SELECT gen_random_uuid(), now(), now(), receivable.id, receivable.paid_amount, receivable.updated_at::date, @method, '', @note
FROM receivable
WHERE receivable.paid_amount > 0
  AND NOT EXISTS (SELECT 1 FROM billing_payment WHERE billing_payment.receivable_id = receivable.id)`

	backfillPayables = `INSERT INTO billing_payment (id, created_at, updated_at, payable_id, amount, paid_at, method, reference, note)
SELECT gen_random_uuid(), now(), now(), payable.id, payable.paid_amount, payable.updated_at::date, @method, '', @note
FROM payable
WHERE payable.paid_amount > 0
  AND NOT EXISTS (SELECT 1 FROM billing_payment WHERE billing_payment.payable_id = payable.id)`
)

var backfillStatements = []string{backfillLinkedPairs, backfillInvoices, backfillReceivables, backfillPayables}

func BackfillPayments(ctx context.Context, repository Repository) (int64, error) {
	var inserted int64
	err := repository.Transaction(ctx, func(repository Repository) error {
		var err error
		inserted, err = repository.BackfillPayments(ctx)
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("backfill billing payments: %w", err)
	}
	return inserted, nil
}
