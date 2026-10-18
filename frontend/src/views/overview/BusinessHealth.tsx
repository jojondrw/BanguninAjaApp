import { useCan } from '../../controllers/useAccess'
import { useBudgetSummary, useProjectSummary } from '../../controllers/useErp'
import { useContractSummary, useUnitSummary } from '../../controllers/useSales'
import { paidContractShare, percentOf, unitFunnel } from '../../models/overview'
import { number, rupiahShort } from '../../shared/format'
import { MoneyText } from '../components/Money'
import { Ring } from '../components/Ring'

const RING_SIZE = 92

// Empat cincin yang menjawab "bisnisnya sehat atau tidak": proyek maju,
// anggaran terserap wajar, unit laku, dan kontrak benar-benar dibayar lunas.
export function BusinessHealth() {
  const canReadProjects = useCan('projects', 'read')
  const canReadFinance = useCan('finance', 'read')
  const canReadSales = useCan('sales', 'read')

  return (
    <div className="grid grid-cols-2 gap-x-4 gap-y-7">
      {canReadProjects ? <ProgressRing /> : <LockedRing label="Progres proyek" module="Proyek" />}
      {canReadFinance ? <BudgetRing /> : <LockedRing label="Serapan anggaran" module="Keuangan" />}
      {canReadSales ? <UnitRing /> : <LockedRing label="Unit laku" module="Penjualan" />}
      {canReadSales ? <ContractRing /> : <LockedRing label="Kontrak lunas" module="Penjualan" />}
    </div>
  )
}

function ProgressRing() {
  const summary = useProjectSummary()
  const data = summary.data
  return (
    <HealthRing
      title="Progres proyek"
      value={data?.averageProgress}
      failed={summary.isError}
      note={data ? `rata-rata ${number(data.count)} proyek` : undefined}
    />
  )
}

function BudgetRing() {
  const budgets = useBudgetSummary()
  const data = budgets.data
  return (
    <HealthRing
      title="Serapan anggaran"
      value={data ? percentOf(data.realized, data.budget) : undefined}
      failed={budgets.isError}
      note={data ? `${rupiahShort(data.realized)} dari ${rupiahShort(data.budget)}` : undefined}
    />
  )
}

function UnitRing() {
  const units = useUnitSummary()
  const stages = units.data ? unitFunnel(units.data, undefined) : undefined
  return (
    <HealthRing
      title="Unit laku"
      value={stages ? percentOf(stages[1].value, stages[0].value) : undefined}
      failed={units.isError}
      note={stages ? `${number(stages[1].value)} dari ${number(stages[0].value)} unit dipasarkan` : undefined}
    />
  )
}

function ContractRing() {
  const contracts = useContractSummary()
  const data = contracts.data
  return (
    <HealthRing
      title="Kontrak lunas"
      value={data ? paidContractShare(data) : undefined}
      failed={contracts.isError}
      note={data ? `${number(data.byStatus.paid)} dari ${number(data.byStatus.active + data.byStatus.paid)} kontrak aktif` : undefined}
    />
  )
}

function LockedRing({ label, module }: { label: string; module: string }) {
  return <HealthRing title={label} failed={false} note={`butuh akses ${module}`} />
}

function HealthRing({ title, value, failed, note }: { title: string; value?: number; failed: boolean; note?: string }) {
  const known = value !== undefined
  return (
    <div className="flex min-w-0 flex-col items-center text-center">
      <Ring value={value ?? 0} size={RING_SIZE} label={known ? `${value}%` : failed ? '—' : '...'} />
      <p className="mt-3 text-[13px] font-medium text-slate-900">{title}</p>
      {note ? (
        <p className="mt-0.5 text-xs text-slate-500">
          <MoneyText>{note}</MoneyText>
        </p>
      ) : null}
    </div>
  )
}
