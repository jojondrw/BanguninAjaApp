import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useBillingPayments, usePayBilling } from '../../controllers/useBilling'
import {
  PAYMENT_METHODS,
  PAYMENT_METHOD_LABEL,
  PAYMENT_NOTE_MAX_LENGTH,
  PAYMENT_REFERENCE_MAX_LENGTH,
  paidPercent,
  paymentFormOf,
  type BillingKind,
  type BillingPayment,
  type BillingRecord,
  type PaymentFormValues,
  type PaymentMethod,
} from '../../models/billing'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, shortDate } from '../../shared/format'
import { Bar, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { Pager, SelectField } from '../components/RecordControls'

const HISTORY_PAGE_SIZE = 5

// Faktur dan piutang yang tertaut berbagi baris pembayaran. Rincian faktur
// menyebut piutangnya, rincian piutang menyebut fakturnya.
function alsoRecordedIn(kind: BillingKind, payment: BillingPayment): string | null {
  if (kind === 'invoices' && payment.receivableReference) {
    return `Juga tercatat di piutang ${payment.receivableReference}`
  }
  if (kind === 'receivables' && payment.invoiceNumber) {
    return `Juga tercatat di faktur ${payment.invoiceNumber}`
  }
  return null
}

function PaymentItem({ kind, payment }: { kind: BillingKind; payment: BillingPayment }) {
  const linked = alsoRecordedIn(kind, payment)
  const details = [PAYMENT_METHOD_LABEL[payment.method] ?? payment.method, payment.reference].filter(Boolean).join(' · ')

  return (
    <li className="flex flex-wrap items-start justify-between gap-x-4 gap-y-1 py-3">
      <div className="min-w-0">
        <p className="text-sm text-slate-900">
          <span className="font-medium">{shortDate(payment.paidAt)}</span>
          <span className="text-slate-500"> · {details}</span>
        </p>
        {payment.note ? <p className="mt-0.5 text-xs break-words text-slate-600">{payment.note}</p> : null}
        {linked ? <p className="mt-0.5 text-xs text-navy-700">{linked}</p> : null}
        {payment.createdByName ? (
          <p className="mt-0.5 text-xs text-slate-500">Dicatat oleh {payment.createdByName}</p>
        ) : null}
      </div>
      <span className="text-sm font-medium text-slate-900 tabular-nums">{rupiah(payment.amount)}</span>
    </li>
  )
}

export function PaymentHistory({ kind, record }: { kind: BillingKind; record: BillingRecord }) {
  const [page, setPage] = useState(1)
  const payments = useBillingPayments(kind, record.id, page, HISTORY_PAGE_SIZE)

  return (
    <section className="mt-6 border-t border-slate-100 pt-5" aria-labelledby="billing-history-title">
      <h3 id="billing-history-title" className="text-sm font-semibold text-slate-900">
        Riwayat pembayaran
      </h3>
      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-slate-700 tabular-nums">
          {rupiah(record.paidAmount)} dari {rupiah(record.amount)} sudah dibayar
          {record.lastPaidAt ? `, terakhir ${shortDate(record.lastPaidAt)}` : ''}
        </p>
        <Bar percent={paidPercent(record)} />
      </div>

      {payments.isPending ? <Loading label="Mengambil riwayat pembayaran..." /> : null}
      {payments.isError ? <LoadFailed onRetry={() => payments.refetch()} /> : null}
      {payments.data ? (
        payments.data.items.length === 0 ? (
          <p className="mt-3 text-sm text-slate-500">Belum ada pembayaran yang dicatat.</p>
        ) : (
          <>
            <ul className="mt-2 divide-y divide-slate-100">
              {payments.data.items.map((payment) => (
                <PaymentItem key={payment.id} kind={kind} payment={payment} />
              ))}
            </ul>
            <Pager
              page={payments.data.page}
              totalPages={payments.data.totalPages}
              totalItems={payments.data.totalItems}
              unit="pembayaran"
              onChange={setPage}
            />
          </>
        )
      ) : null}
    </section>
  )
}

type Change = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

export function PaymentForm({ record, noun, today, linkedTo, pay }: {
  record: BillingRecord
  noun: string
  today: string
  linkedTo: string | null
  pay: ReturnType<typeof usePayBilling>
}) {
  const [values, setValues] = useState<PaymentFormValues>(() => paymentFormOf(record, today))
  const [isConfirming, setIsConfirming] = useState(false)
  const isSettled = record.outstanding === 0
  const amount = Number(values.amount)

  const change = (key: keyof PaymentFormValues): Change => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  // Tombol pertama hanya memeriksa isian lewat validasi peramban, pencatatan
  // baru terjadi setelah dikonfirmasi di tempat yang sama.
  const submit = (event: FormEvent) => {
    event.preventDefault()
    setIsConfirming(true)
  }

  const confirm = () => pay.mutate({ id: record.id, values }, { onSettled: () => setIsConfirming(false) })

  return (
    <form onSubmit={submit}>
      <fieldset disabled={isSettled} className="space-y-3">
        <legend className="mb-3 text-sm font-semibold text-slate-900">Catat pembayaran</legend>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Field
            id="billing-payment-amount"
            label="Nominal pembayaran"
            type="number"
            inputMode="numeric"
            min={1}
            max={record.outstanding}
            step={1}
            required
            disabled={isConfirming}
            hint={
              isSettled
                ? `Sudah lunas, tidak ada sisa ${noun} yang perlu dibayar`
                : `Terisi sisa ${rupiah(record.outstanding)}. Boleh dibayar sebagian.`
            }
            value={values.amount}
            onChange={change('amount')}
          />
          <Field
            id="billing-payment-date"
            label="Tanggal bayar"
            type="date"
            max={today}
            required
            disabled={isConfirming}
            hint="Tidak boleh setelah hari ini"
            value={values.paidAt}
            onChange={change('paidAt')}
          />
          <SelectField
            id="billing-payment-method"
            label="Metode"
            disabled={isConfirming}
            value={values.method}
            onChange={(event) => setValues((current) => ({ ...current, method: event.target.value as PaymentMethod }))}
          >
            {PAYMENT_METHODS.map((method) => (
              <option key={method} value={method}>
                {PAYMENT_METHOD_LABEL[method]}
              </option>
            ))}
          </SelectField>
          <Field
            id="billing-payment-reference"
            label="Referensi (opsional)"
            placeholder="No. bukti transfer atau kuitansi"
            autoComplete="off"
            maxLength={PAYMENT_REFERENCE_MAX_LENGTH}
            disabled={isConfirming}
            value={values.reference}
            onChange={change('reference')}
          />
        </div>
        <Field
          id="billing-payment-note"
          label="Catatan (opsional)"
          placeholder="Pelunasan termin 2"
          autoComplete="off"
          maxLength={PAYMENT_NOTE_MAX_LENGTH}
          hint={`Maksimal ${PAYMENT_NOTE_MAX_LENGTH} karakter`}
          disabled={isConfirming}
          value={values.note}
          onChange={change('note')}
        />
        {linkedTo ? <p className="text-xs text-slate-600">Pembayaran ini juga tercatat di {linkedTo}.</p> : null}

        {isConfirming ? (
          <div className="flex flex-wrap items-center gap-2 rounded-xl bg-amber-50 px-3.5 py-2.5">
            <p className="mr-auto text-[13px] text-amber-900">
              Catat pembayaran {rupiah(amount)} lewat {PAYMENT_METHOD_LABEL[values.method].toLowerCase()} tanggal{' '}
              {shortDate(values.paidAt)}? Pembayaran yang sudah dicatat tidak bisa dibatalkan.
            </p>
            <Button isPending={pay.isPending} pendingLabel="Mencatat pembayaran" onClick={confirm}>
              Ya, catat pembayaran
            </Button>
            {pay.isPending ? null : (
              <Button variant="subtle" onClick={() => setIsConfirming(false)}>
                Batal
              </Button>
            )}
          </div>
        ) : (
          <Button type="submit">Catat pembayaran</Button>
        )}
      </fieldset>

      <div className="mt-3">
        {pay.isError ? <ErrorNote message={errorMessage(pay.error)} /> : null}
        {pay.isSuccess ? (
          <SuccessNote
            message={`Pembayaran ${rupiah(Number(pay.variables.values.amount))} tercatat. ${
              pay.data.status === 'paid' ? `Sekarang ${noun} ini lunas.` : `Sisa ${rupiah(pay.data.outstanding)}.`
            }`}
          />
        ) : null}
      </div>
    </form>
  )
}
