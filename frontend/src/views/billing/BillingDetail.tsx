import { type FormEvent, type ReactNode, useState } from 'react'

import { useBillingRecord, useDeleteBilling, usePayBilling } from '../../controllers/useBilling'
import {
  BILLING_KIND_NOUN,
  dueNote,
  paidPercent,
  type BillingKind,
  type BillingRecord,
  type BillingRecordOf,
} from '../../models/billing'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiah, shortDate } from '../../shared/format'
import { Bar, Card, LoadFailed, Loading } from '../components/Data'
import { BUTTON_BASE, Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { FormPanel } from '../components/RecordControls'
import { StatusChip } from './BillingCells'

export interface DetailField {
  label: string
  value: ReactNode
}

interface EditControls {
  onSaved: () => void
  onCancel: () => void
}

const DANGER_BUTTON = 'bg-red-600 text-white shadow-button hover:bg-red-700 disabled:bg-red-600/55'

function PaymentForm({ record, noun, pay }: {
  record: BillingRecord
  noun: string
  pay: ReturnType<typeof usePayBilling>
}) {
  const [amount, setAmount] = useState(String(record.outstanding))
  const [isConfirming, setIsConfirming] = useState(false)
  const isSettled = record.outstanding === 0
  const value = Number(amount)

  // Tombol pertama hanya memeriksa isian lewat validasi peramban, pencatatan
  // baru terjadi setelah dikonfirmasi di tempat yang sama.
  const submit = (event: FormEvent) => {
    event.preventDefault()
    setIsConfirming(true)
  }

  const confirm = () =>
    pay.mutate({ id: record.id, amount: value }, { onSettled: () => setIsConfirming(false) })

  return (
    <form onSubmit={submit}>
      <fieldset disabled={isSettled} className="space-y-3">
        <legend className="mb-3 text-sm font-semibold text-slate-900">Catat pembayaran</legend>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
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
            value={amount}
            onChange={(event) => setAmount(event.target.value)}
          />
        </div>

        {isConfirming ? (
          <div className="flex flex-wrap items-center gap-2 rounded-xl bg-amber-50 px-3.5 py-2.5">
            <p className="mr-auto text-[13px] text-amber-900">
              Catat pembayaran {rupiah(value)}? Pembayaran yang sudah dicatat tidak bisa dibatalkan.
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
            message={`Pembayaran ${rupiah(pay.variables.amount)} tercatat. ${
              pay.data.status === 'paid' ? `Sekarang ${noun} ini lunas.` : `Sisa ${rupiah(pay.data.outstanding)}.`
            }`}
          />
        ) : null}
      </div>
    </form>
  )
}

function DeleteAction({ kind, record, label, onDeleted }: {
  kind: BillingKind
  record: BillingRecord
  label: string
  onDeleted: (message: string) => void
}) {
  const remove = useDeleteBilling(kind)
  const [isConfirming, setIsConfirming] = useState(false)
  const noun = BILLING_KIND_NOUN[kind]

  if (record.paidAmount > 0) {
    return <p className="text-xs text-slate-500">Sudah ada pembayaran, jadi {noun} ini tidak bisa dihapus.</p>
  }

  if (!isConfirming) {
    return (
      <Button variant="subtle" onClick={() => setIsConfirming(true)}>
        Hapus {noun}
      </Button>
    )
  }

  return (
    <span className="flex flex-col items-end gap-2">
      <span className="flex flex-wrap items-center justify-end gap-2">
        <span className="text-[13px] text-slate-700">Hapus {label} untuk selamanya?</span>
        <button
          type="button"
          disabled={remove.isPending}
          aria-busy={remove.isPending}
          onClick={() =>
            remove.mutate(record.id, {
              onSuccess: () => onDeleted(`Data ${label} sudah dihapus.`),
            })
          }
          className={`h-9 ${BUTTON_BASE} ${DANGER_BUTTON}`}
        >
          {remove.isPending ? 'Menghapus' : `Ya, hapus ${noun}`}
        </button>
        {remove.isPending ? null : (
          <Button variant="subtle" onClick={() => setIsConfirming(false)}>
            Batal
          </Button>
        )}
      </span>
      {remove.isError ? <ErrorNote message={errorMessage(remove.error)} /> : null}
    </span>
  )
}

function DetailList({ fields }: { fields: DetailField[] }) {
  return (
    <dl className="grid gap-x-6 gap-y-4 sm:grid-cols-2 lg:grid-cols-3">
      {fields.map((field) => (
        <div key={field.label} className="min-w-0">
          <dt className="text-xs font-medium text-slate-500">{field.label}</dt>
          <dd className="mt-0.5 text-sm break-words text-slate-900">{field.value}</dd>
        </div>
      ))}
    </dl>
  )
}

// Rincian dibaca ulang dari GET /billing/<kind>/:id supaya status, sisa, dan
// keterlambatan selalu hasil hitungan server terbaru.
export function BillingDetail<K extends BillingKind>({ kind, id, today, heading, fields, renderEdit, onDeleted }: {
  kind: K
  id: string
  today: string
  heading: (record: BillingRecordOf[K]) => string
  fields: (record: BillingRecordOf[K]) => DetailField[]
  renderEdit: (record: BillingRecordOf[K], controls: EditControls) => ReactNode
  onDeleted: (message: string) => void
}) {
  const record = useBillingRecord(kind, id)
  const pay = usePayBilling(kind)
  const [isEditing, setIsEditing] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const noun = BILLING_KIND_NOUN[kind]

  if (record.isPending) {
    return (
      <Card title={`Rincian ${noun}`}>
        <Loading />
      </Card>
    )
  }

  if (record.isError) {
    return (
      <Card title={`Rincian ${noun}`}>
        <LoadFailed onRetry={() => record.refetch()} />
      </Card>
    )
  }

  const data = record.data
  const title = heading(data)
  const note = dueNote(data, today)

  return (
    <Card title={`Rincian ${title}`} description={`Dibuat ${shortDate(data.createdAt)}, terakhir diubah ${shortDate(data.updatedAt)}`}>
      <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
        <span className="flex flex-wrap items-center gap-3">
          <StatusChip status={data.status} />
          {note ? (
            <span className={`text-sm ${data.status === 'overdue' ? 'font-medium text-red-700' : 'text-amber-700'}`}>
              {note}
            </span>
          ) : null}
        </span>
        <span className="flex flex-wrap items-start justify-end gap-2">
          <Button
            variant="subtle"
            onClick={() => {
              setIsEditing((open) => !open)
              setNotice(null)
            }}
          >
            {isEditing ? 'Tutup formulir ubah' : 'Ubah data'}
          </Button>
          <DeleteAction kind={kind} record={data} label={title} onDeleted={onDeleted} />
        </span>
      </div>

      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}

      {isEditing ? (
        <FormPanel>
          {renderEdit(data, {
            onSaved: () => {
              setIsEditing(false)
              setNotice(`Perubahan ${title} disimpan.`)
            },
            onCancel: () => setIsEditing(false),
          })}
        </FormPanel>
      ) : null}

      <DetailList
        fields={[
          ...fields(data),
          { label: 'Jatuh tempo', value: shortDate(data.dueDate) },
          { label: 'Nilai', value: rupiah(data.amount) },
          { label: 'Sudah dibayar', value: rupiah(data.paidAmount) },
          { label: 'Sisa', value: <span className="font-medium">{rupiah(data.outstanding)}</span> },
          {
            label: 'Keterlambatan',
            value: data.daysOverdue > 0 ? `${number(data.daysOverdue)} hari lewat jatuh tempo` : 'Tidak terlambat',
          },
        ]}
      />

      <section className="mt-6 border-t border-slate-100 pt-5" aria-labelledby="billing-history-title">
        <h3 id="billing-history-title" className="text-sm font-semibold text-slate-900">
          Riwayat pembayaran
        </h3>
        <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-slate-700 tabular-nums">
            {rupiah(data.paidAmount)} dari {rupiah(data.amount)} sudah dibayar
          </p>
          <Bar percent={paidPercent(data)} />
        </div>
        <p className="mt-2 text-xs text-slate-500">
          Server baru menyimpan total yang sudah dibayar, belum rincian tanggal dan nominal tiap pembayaran.
        </p>
      </section>

      <section className="mt-6 border-t border-slate-100 pt-5">
        <PaymentForm key={data.outstanding} record={data} noun={noun} pay={pay} />
      </section>
    </Card>
  )
}
