import { useBillingSummary } from '../../controllers/useBilling'
import { AGING_LABEL, type AgingBucket, type BalanceSummary, type BillingStatus } from '../../models/billing'
import { number, rupiahShort } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { LinkButton } from '../components/LinkButton'
import { Money } from '../components/Money'

const BUCKET_COLOR: Record<AgingBucket, string> = {
  current: 'bg-green-500',
  days_30: 'bg-amber-300',
  days_60: 'bg-amber-500',
  days_90: 'bg-orange-600',
  days_over_90: 'bg-red-600',
}

export interface BillingDrill {
  openReceivables: (status: BillingStatus | '') => void
  openPayables: (status: BillingStatus | '') => void
}

function AgingBar({ summary }: { summary: BalanceSummary }) {
  if (summary.outstanding === 0) {
    return null
  }
  return (
    <span aria-hidden="true" className="flex h-3 w-full overflow-hidden rounded-full bg-slate-100">
      {summary.buckets.map((bucket) =>
        bucket.amount > 0 ? (
          <span
            key={bucket.bucket}
            className={`h-full ${BUCKET_COLOR[bucket.bucket]}`}
            style={{ width: `${(bucket.amount / summary.outstanding) * 100}%` }}
          />
        ) : null,
      )}
    </span>
  )
}

function BalanceCard({ title, partyLabel, summary, onOpen }: {
  title: string
  partyLabel: string
  summary: BalanceSummary
  onOpen: (status: BillingStatus | '') => void
}) {
  const overdueShare = summary.outstanding > 0 ? Math.round((summary.overdue / summary.outstanding) * 100) : 0
  return (
    <Card title={title} action={<LinkButton label={`Semua ${title.toLowerCase()}`} onClick={() => onOpen('')} />}>
      <div className="grid grid-cols-2 gap-3">
        <button type="button" onClick={() => onOpen('')} className="rounded-xl bg-slate-50 px-3.5 py-3 text-left hover:bg-slate-100">
          <span className="block text-xs text-slate-500">Sisa belum lunas</span>
          <span className="mt-0.5 block text-xl font-semibold tracking-[-0.02em] text-slate-900">
            <Money>{rupiahShort(summary.outstanding)}</Money>
          </span>
          <span className="text-xs text-slate-500">{number(summary.openCount)} tagihan terbuka</span>
        </button>
        <button type="button" onClick={() => onOpen('overdue')} className="rounded-xl bg-red-50 px-3.5 py-3 text-left hover:bg-red-100/70">
          <span className="block text-xs text-red-700">Terlambat</span>
          <span className="mt-0.5 block text-xl font-semibold tracking-[-0.02em] text-red-700">
            <Money>{rupiahShort(summary.overdue)}</Money>
          </span>
          <span className="text-xs text-red-700/80">{overdueShare}% dari sisa</span>
        </button>
      </div>

      <div className="mt-5 space-y-2">
        <AgingBar summary={summary} />
        <ul className="space-y-1">
          {summary.buckets.map((bucket) => (
            <li key={bucket.bucket}>
              <button
                type="button"
                onClick={() => onOpen(bucket.bucket === 'current' ? '' : 'overdue')}
                className="flex w-full items-center gap-2 rounded-md px-1.5 py-1 text-left text-[13px] hover:bg-slate-50"
              >
                <span aria-hidden="true" className={`size-2 shrink-0 rounded-full ${BUCKET_COLOR[bucket.bucket]}`} />
                <span className="min-w-0 flex-1 text-slate-700">{AGING_LABEL[bucket.bucket]}</span>
                <span className="text-xs text-slate-400 tabular-nums">{number(bucket.count)}</span>
                <span className="w-24 text-right text-slate-900 tabular-nums">
                  <Money>{rupiahShort(bucket.amount)}</Money>
                </span>
              </button>
            </li>
          ))}
        </ul>
      </div>

      {summary.largest.length > 0 ? (
        <div className="mt-5">
          <p className="mb-2 text-xs font-medium text-slate-500">{partyLabel} dengan sisa terbesar</p>
          <ul className="divide-y divide-slate-100">
            {summary.largest.map((party) => (
              <li key={party.name} className="flex items-center justify-between gap-3 py-2 text-[13px]">
                <span className="min-w-0 truncate text-slate-900">{party.name}</span>
                <span className="shrink-0 text-slate-900 tabular-nums">
                  <Money>{rupiahShort(party.amount)}</Money>
                  <span className="ml-1.5 text-xs text-slate-400">{number(party.count)} tagihan</span>
                </span>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </Card>
  )
}

// Ringkasan tagihan: umur piutang dan utang beserta pihak terbesar. Klik angka
// terlambat atau umur membuka tabnya dengan saringan Terlambat.
export function BillingOverview({ drill }: { drill: BillingDrill }) {
  const summary = useBillingSummary()

  if (summary.isPending) {
    return <Loading />
  }
  if (summary.isError) {
    return <LoadFailed onRetry={() => summary.refetch()} />
  }
  return (
    <div className="grid gap-6 xl:grid-cols-2">
      <BalanceCard title="Piutang" partyLabel="Pelanggan" summary={summary.data.receivables} onOpen={drill.openReceivables} />
      <BalanceCard title="Utang" partyLabel="Vendor" summary={summary.data.payables} onOpen={drill.openPayables} />
    </div>
  )
}
