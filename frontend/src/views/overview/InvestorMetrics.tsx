import type { ReactNode } from 'react'

import { useCan } from '../../controllers/useAccess'
import { useBillingTotals } from '../../controllers/useBilling'
import { useCashFlow, useProjectSummary } from '../../controllers/useErp'
import { monthEndBalances, percentOf, runwayMonths } from '../../models/overview'
import { number, rupiahShort } from '../../shared/format'
import { Money, MoneyText } from '../components/Money'
import { Sparkline } from '../components/Sparkline'

const DASH = '—'

// Metrik keuangan utama: posisi kas dan runway, kas masuk,
// belanja dan margin kas, modal kerja (piutang dan hutang), dan nilai kontrak.
// Setiap kelompok hanya diambil kalau peran ini boleh membaca modulnya.
export function InvestorMetrics({ cashFrom, cashTo }: { cashFrom: string; cashTo: string }) {
  const canReadFinance = useCan('finance', 'read')
  const canReadBilling = useCan('billing', 'read')
  const canReadProjects = useCan('projects', 'read')

  return (
    <section aria-labelledby="investor-metrics" className="space-y-3">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="investor-metrics" className="text-[15px] font-semibold text-slate-900">
          Metrik keuangan
        </h2>
        <p className="text-xs text-slate-500">12 bulan terakhir</p>
      </div>
      <div className="grid gap-px overflow-hidden rounded-2xl bg-slate-200/70 shadow-hairline sm:grid-cols-2 xl:grid-cols-3">
        {canReadFinance ? <CashTiles from={cashFrom} to={cashTo} /> : <LockedTiles count={3} module="Keuangan" />}
        {canReadBilling ? <WorkingCapitalTiles /> : <LockedTiles count={2} module="Tagihan" />}
        {canReadProjects ? <OrderBookTile /> : <LockedTiles count={1} module="Proyek" />}
      </div>
    </section>
  )
}

function CashTiles({ from, to }: { from: string; to: string }) {
  const flow = useCashFlow({ dateFrom: from, dateTo: to })
  const data = flow.data
  if (!data) {
    return (
      <>
        <Tile label="Saldo kas" value={flow.isError ? DASH : '...'} />
        <Tile label="Kas masuk" value={flow.isError ? DASH : '...'} />
        <Tile label="Belanja (kas keluar)" value={flow.isError ? DASH : '...'} />
      </>
    )
  }

  const months = Math.max(1, data.periods.length)
  const runway = runwayMonths(data.periods, data.balance)
  const averageNet = Math.round(data.net / months)

  return (
    <>
      <Tile
        label="Saldo kas"
        value={<Money>{rupiahShort(data.balance)}</Money>}
        note={runway === null ? 'kas tidak terbakar dalam 12 bulan ini' : `cukup untuk ±${number(runway)} bulan pada laju sekarang`}
        chart={<Sparkline values={monthEndBalances(data.periods, data.balance)} />}
      />
      <Tile
        label="Kas masuk"
        value={<Money>{rupiahShort(data.totalIn)}</Money>}
        note={`rata-rata ${rupiahShort(Math.round(data.totalIn / months))} per bulan`}
        chart={<Sparkline values={data.periods.map((period) => period.cashIn)} />}
      />
      <Tile
        label="Belanja (kas keluar)"
        value={<Money>{rupiahShort(data.totalOut)}</Money>}
        note={`margin kas bersih ${percentOf(data.net, data.totalIn)}%, ${rupiahShort(averageNet)} per bulan`}
        tone={data.net < 0 ? 'negative' : 'neutral'}
        chart={<Sparkline values={data.periods.map((period) => period.cashOut)} tone={data.net < 0 ? 'red' : 'navy'} />}
      />
    </>
  )
}

function WorkingCapitalTiles() {
  const receivables = useBillingTotals('receivables')
  const payables = useBillingTotals('payables')
  const owed = receivables.data
  const owing = payables.data

  return (
    <>
      <Tile
        label="Piutang beredar"
        value={owed ? <Money>{rupiahShort(owed.outstanding)}</Money> : receivables.isError ? DASH : '...'}
        note={owed ? `${rupiahShort(owed.overdueOutstanding)} lewat jatuh tempo` : undefined}
        chart={owed ? <Meter percent={percentOf(owed.overdueOutstanding, owed.outstanding)} label="lewat jatuh tempo" /> : null}
      />
      <Tile
        label="Hutang beredar"
        value={owing ? <Money>{rupiahShort(owing.outstanding)}</Money> : payables.isError ? DASH : '...'}
        note={owing ? `${number(owing.dueSoonCount)} jatuh tempo segera, ${number(owing.overdueCount)} terlambat` : undefined}
        chart={owing ? <Meter percent={percentOf(owing.overdueOutstanding, owing.outstanding)} label="lewat jatuh tempo" /> : null}
      />
    </>
  )
}

function OrderBookTile() {
  const summary = useProjectSummary()
  const data = summary.data

  return (
    <Tile
      label="Nilai kontrak proyek"
      value={data ? <Money>{rupiahShort(data.contractValue)}</Money> : summary.isError ? DASH : '...'}
      note={data ? `${number(data.count)} proyek, rata-rata progres ${data.averageProgress}%` : undefined}
      chart={data ? <Meter percent={data.averageProgress} label="dikerjakan" /> : null}
    />
  )
}

function LockedTiles({ count, module }: { count: number; module: string }) {
  return Array.from({ length: count }, (_, index) => (
    <Tile key={index} label={`Butuh akses ${module}`} value={DASH} note="Metrik ini disembunyikan untuk peran Anda" />
  ))
}

function Tile({ label, value, note, chart, tone = 'neutral' }: {
  label: string
  value: ReactNode
  note?: string
  chart?: ReactNode
  tone?: 'neutral' | 'negative'
}) {
  return (
    <div className="flex min-w-0 flex-col justify-between gap-3 bg-white px-5 py-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[13px] font-medium text-slate-500">{label}</p>
          <p
            className={`mt-1.5 text-[24px] leading-8 font-semibold tracking-[-0.03em] tabular-nums ${
              tone === 'negative' ? 'text-red-600' : 'text-slate-900'
            }`}
          >
            {value}
          </p>
        </div>
        {chart ? <div className="shrink-0 pt-1">{chart}</div> : null}
      </div>
      {note ? <p className="text-xs text-slate-500"><MoneyText>{note}</MoneyText></p> : null}
    </div>
  )
}

function Meter({ percent, label }: { percent: number; label: string }) {
  const safe = Math.max(0, Math.min(100, percent))
  return (
    <div className="w-[120px]">
      <span aria-hidden="true" className="block h-1.5 overflow-hidden rounded-[2px] bg-slate-100">
        <span className="bar-fill block h-full rounded-[2px]" style={{ width: `${safe}%` }} />
      </span>
      <p className="mt-1 text-right text-[11px] text-slate-500 tabular-nums">
        {safe}% {label}
      </p>
    </div>
  )
}
