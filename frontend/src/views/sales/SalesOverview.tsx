import { ChevronRight } from 'lucide-react'

import { useContractSummary, useLeadStageCounts, useProjectUnitSummary } from '../../controllers/useSales'
import { projectOptions } from '../../models/lookupApi'
import {
  CONTRACT_STATUS_LABEL,
  LEAD_STAGE_LABEL,
  LEAD_STAGES,
  type ContractStatus,
  type LeadStage,
} from '../../models/sales'
import { number, rupiahShort } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { LookupName } from '../components/LookupName'
import { Money } from '../components/Money'
import { SalesFunnel } from '../overview/SalesFunnel'
import { ProjectUnitBar, UnitBarLegend } from './ProjectUnitBar'
import { unitBreakdown } from './unitSegments'

const TOP_PROJECTS = 6
const CONTRACT_ORDER: ContractStatus[] = ['draft', 'active', 'paid', 'cancelled']
const STAGE_COLOR: Record<LeadStage, string> = {
  new: 'bg-slate-400',
  interested: 'bg-navy-500',
  negotiating: 'bg-amber-400',
  won: 'bg-green-500',
  cancelled: 'bg-red-300',
}

export interface SalesDrill {
  openUnits: (projectId: string) => void
  openLeads: (stage: LeadStage | '') => void
  openContracts: (status: ContractStatus | '') => void
}

function LinkButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex items-center gap-0.5 text-[13px] font-medium text-navy-600 hover:text-navy-800"
    >
      {label}
      <ChevronRight aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
    </button>
  )
}

function LeadPipeline({ onOpen }: { onOpen: SalesDrill['openLeads'] }) {
  const counts = useLeadStageCounts()

  if (counts.isPending) {
    return <Loading />
  }
  if (counts.isError) {
    return <LoadFailed onRetry={() => counts.refetch()} />
  }
  const highest = Math.max(1, ...LEAD_STAGES.map((stage) => counts.data[stage]))
  return (
    <ul className="space-y-1">
      {LEAD_STAGES.map((stage) => (
        <li key={stage}>
          <button
            type="button"
            onClick={() => onOpen(stage)}
            className="grid w-full grid-cols-[6rem_1fr_2.5rem] items-center gap-3 rounded-lg px-2 py-2 text-left hover:bg-slate-50"
          >
            <span className="text-[13px] text-slate-700">{LEAD_STAGE_LABEL[stage]}</span>
            <span aria-hidden="true" className="h-2 overflow-hidden rounded-full bg-slate-100">
              <span
                className={`block h-full rounded-full ${STAGE_COLOR[stage]}`}
                style={{ width: `${(counts.data[stage] / highest) * 100}%` }}
              />
            </span>
            <span className="text-right text-[13px] font-medium text-slate-900 tabular-nums">
              {number(counts.data[stage])}
            </span>
          </button>
        </li>
      ))}
    </ul>
  )
}

function ContractTiles({ onOpen }: { onOpen: SalesDrill['openContracts'] }) {
  const summary = useContractSummary()

  if (summary.isPending) {
    return <Loading />
  }
  if (summary.isError) {
    return <LoadFailed onRetry={() => summary.refetch()} />
  }
  return (
    <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
      {CONTRACT_ORDER.map((status) => (
        <button
          key={status}
          type="button"
          onClick={() => onOpen(status)}
          className="rounded-xl bg-slate-50 px-3.5 py-3 text-left transition-colors hover:bg-slate-100"
        >
          <span className="block text-xs text-slate-500">{CONTRACT_STATUS_LABEL[status]}</span>
          <span className="mt-0.5 block text-xl font-semibold tracking-[-0.02em] text-slate-900 tabular-nums">
            {number(summary.data.byStatus[status])}
          </span>
        </button>
      ))}
    </div>
  )
}

function ProjectStock({ onOpen }: { onOpen: SalesDrill['openUnits'] }) {
  const projects = useProjectUnitSummary()

  if (projects.isPending) {
    return <Loading />
  }
  if (projects.isError) {
    return <LoadFailed onRetry={() => projects.refetch()} />
  }
  if (projects.data.length === 0) {
    return <Empty message="Belum ada unit yang dipasarkan." />
  }
  const top = [...projects.data].sort((a, b) => b.total - a.total).slice(0, TOP_PROJECTS)
  return (
    <div className="space-y-3">
      <UnitBarLegend />
      <ul className="divide-y divide-slate-100">
        {top.map((summary) => (
          <li key={summary.projectId}>
            <button
              type="button"
              onClick={() => onOpen(summary.projectId)}
              className="grid w-full gap-1.5 py-3 text-left sm:grid-cols-[minmax(0,14rem)_1fr_7rem] sm:items-center sm:gap-4"
            >
              <span className="truncate text-[13px] font-medium text-slate-900">
                <LookupName source={projectOptions} value={summary.projectId} />
              </span>
              <span className="min-w-0 space-y-1">
                <ProjectUnitBar summary={summary} />
                <span className="block text-xs text-slate-500">
                  {number(summary.total)} unit: {unitBreakdown(summary)}
                </span>
              </span>
              <span className="text-xs text-slate-500 sm:text-right">
                stok <Money>{rupiahShort(summary.stockValue)}</Money>
              </span>
            </button>
          </li>
        ))}
      </ul>
      {projects.data.length > TOP_PROJECTS ? (
        <LinkButton label={`Semua ${number(projects.data.length)} proyek di tab Unit`} onClick={() => onOpen('')} />
      ) : null}
    </div>
  )
}

// Tab pertama Penjualan: corong, pipeline, kontrak, dan stok per proyek dalam
// satu layar. Setiap baris membuka tab rinciannya yang sudah tersaring.
export function SalesOverview({ drill }: { drill: SalesDrill }) {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 xl:grid-cols-2">
        <Card title="Corong penjualan" description="Dari unit dipasarkan sampai lunas">
          <SalesFunnel />
        </Card>
        <Card
          title="Pipeline prospek"
          description="Pilih tahap untuk melihat prospeknya"
          action={<LinkButton label="Semua prospek" onClick={() => drill.openLeads('')} />}
        >
          <LeadPipeline onOpen={drill.openLeads} />
        </Card>
      </div>
      <Card
        title="Kontrak"
        description="Pilih status untuk melihat daftar kontraknya"
        action={<LinkButton label="Semua kontrak" onClick={() => drill.openContracts('')} />}
      >
        <ContractTiles onOpen={drill.openContracts} />
      </Card>
      <Card title="Stok unit per proyek" description="Proyek dengan unit terbanyak. Pilih proyek untuk membuka unitnya">
        <ProjectStock onOpen={drill.openUnits} />
      </Card>
    </div>
  )
}
