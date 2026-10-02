import {
  BILLING_STATUS_LABEL,
  BILLING_STATUS_TONE,
  dueNote,
  paidPercent,
  type BillingBalance,
  type BillingStatus,
} from '../../models/billing'
import { rupiah, shortDate } from '../../shared/format'
import { Chip } from '../components/RecordControls'

export function StatusChip({ status }: { status: BillingStatus }) {
  return <Chip tone={BILLING_STATUS_TONE[status]}>{BILLING_STATUS_LABEL[status]}</Chip>
}

// Tanggal jatuh tempo dengan keterangan tertulis. Merah hanya penegas, isi
// kalimatnya yang membawa informasi.
export function DueDateCell({ record, today }: { record: BillingBalance; today: string }) {
  const note = dueNote(record, today)
  const tone = record.status === 'overdue' ? 'font-medium text-red-700' : 'text-amber-700'

  return (
    <span className="flex flex-col">
      <span className="text-slate-900">{shortDate(record.dueDate)}</span>
      {note ? <span className={`text-xs ${tone}`}>{note}</span> : null}
    </span>
  )
}

export function BalanceCell({ record }: { record: BillingBalance }) {
  if (record.outstanding === 0) {
    return <span className="text-slate-500">{rupiah(0)}</span>
  }

  return (
    <span className="flex flex-col items-end">
      <span className="font-medium text-slate-900">{rupiah(record.outstanding)}</span>
      {record.paidAmount > 0 ? (
        <span className="text-xs text-slate-500">terbayar {paidPercent(record)}%</span>
      ) : null}
    </span>
  )
}
