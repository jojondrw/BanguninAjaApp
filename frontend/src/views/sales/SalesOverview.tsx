import { ChevronRight } from 'lucide-react'

import { useContractSummary, useProjectUnitSummary } from '../../controllers/useSales'
import { projectOptions } from '../../models/lookupApi'
import { CONTRACT_STATUS_LABEL, type ContractStatus, type LeadStage } from '../../models/sales'
import { number, rupiahShort } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { LookupName } from '../components/LookupName'
import { Money } from '../components/Money'
import { SalesFunnel } from '../overview/SalesFunnel'
import { LeadSankey } from './LeadSankey'
import { ProjectUnitBar, UnitBarLegend } from './ProjectUnitBar'
import { unitBreakdown } from './unitSegments'

const TOP_PROJECTS = 6
const CONTRACT_ORDER: ContractStatus[] = ['draft', 'active', 'paid', 'cancelled']

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
      <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
        <Card title="Corong penjualan" description="Dari unit dipasarkan sampai lunas">
          <SalesFunnel />
        </Card>
        <Card
          title="Sumber prospek"
          description="Dari mana prospek datang dan di tahap mana mereka sekarang. Pilih pita atau tahap untuk melihat prospeknya"
          action={<LinkButton label="Semua prospek" onClick={() => drill.openLeads('')} />}
        >
          <LeadSankey onOpenStage={drill.openLeads} />
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
