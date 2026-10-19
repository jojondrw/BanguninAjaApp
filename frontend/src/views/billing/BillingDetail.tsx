import { type ReactNode, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useBillingRecord, useDeleteBilling, usePayBilling } from '../../controllers/useBilling'
import {
  BILLING_KIND_NOUN,
  dueNote,
  type BillingKind,
  type BillingRecord,
  type BillingRecordOf,
} from '../../models/billing'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiah, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { BUTTON_BASE, Button, ErrorNote, SuccessNote } from '../components/Form'
import { FormPanel } from '../components/RecordControls'
import { Money, MoneyText } from '../components/Money'
import type { OnchainBillingKind } from '../../models/onchain'
import { StatusChip } from './BillingCells'
import { OnchainSection } from './OnchainPanel'
import { QrisSection } from './QrisPanel'
import { PaymentForm, PaymentHistory } from './BillingPayments'

export interface DetailField {
  label: string
  value: ReactNode
}

interface EditControls {
  onSaved: () => void
  onCancel: () => void
}

const DANGER_BUTTON = 'bg-red-600 text-white shadow-button hover:bg-red-700 disabled:bg-red-600/55'

// Pembayaran on-chain hanya untuk sisi pelanggan: faktur ke pelanggan dan
// piutang. Utang dan faktur vendor tidak ditawari.
function isOnchainKind(kind: BillingKind): kind is OnchainBillingKind {
  return kind !== 'payables'
}

function isCustomerSide(record: BillingRecord): boolean {
  return !('partyType' in record) || record.partyType === 'customer'
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
          <dd className="mt-0.5 text-sm break-words text-slate-900"><MoneyText>{field.value}</MoneyText></dd>
        </div>
      ))}
    </dl>
  )
}

// Rincian dibaca ulang dari GET /billing/<kind>/:id supaya status, sisa, dan
// keterlambatan selalu hasil hitungan server terbaru. linkedTo menyebut
// catatan pasangan (faktur atau piutang) yang ikut terbayar.
interface BillingDetailProps<K extends BillingKind> {
  kind: K
  id: string
  today: string
  heading: (record: BillingRecordOf[K]) => string
  fields: (record: BillingRecordOf[K]) => DetailField[]
  linkedTo?: (record: BillingRecordOf[K]) => string | null
  renderEdit: (record: BillingRecordOf[K], controls: EditControls) => ReactNode
  onDeleted: (message: string) => void
}

export function BillingDetail<K extends BillingKind>({ kind, id, today, heading, fields, linkedTo, renderEdit, onDeleted }: BillingDetailProps<K>) {
  const record = useBillingRecord(kind, id)
  const pay = usePayBilling(kind)
  const canWrite = useScopedCan('write')
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
        {canWrite ? (
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
        ) : null}
      </div>

      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}

      {isEditing && canWrite ? (
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
          { label: 'Sisa', value: <span className="font-medium"><Money>{rupiah(data.outstanding)}</Money></span> },
          {
            label: 'Keterlambatan',
            value: data.daysOverdue > 0 ? `${number(data.daysOverdue)} hari lewat jatuh tempo` : 'Tidak terlambat',
          },
        ]}
      />

      <PaymentHistory kind={kind} record={data} />

      {isOnchainKind(kind) && isCustomerSide(data) ? <QrisSection kind={kind} record={data} /> : null}

      {isOnchainKind(kind) && isCustomerSide(data) ? <OnchainSection kind={kind} record={data} /> : null}

      {canWrite ? (
        <section className="mt-6 border-t border-slate-100 pt-5">
          <PaymentForm
            key={data.outstanding}
            record={data}
            noun={noun}
            today={today}
            linkedTo={linkedTo?.(data) ?? null}
            pay={pay}
          />
        </section>
      ) : null}
    </Card>
  )
}
