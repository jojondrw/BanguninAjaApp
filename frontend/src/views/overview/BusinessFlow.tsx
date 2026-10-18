import { Boxes, ChevronRight, FolderKanban, Handshake, Lock, MapPinned, ShoppingCart, Users, type LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import { useCan } from '../../controllers/useAccess'
import { useProjectSummary, useSavedLocationSummary } from '../../controllers/useErp'
import { useEmployeeHeadcount } from '../../controllers/useHr'
import { useLowStockMaterials, useMaterials } from '../../controllers/useInventory'
import { usePurchaseOrderSummary } from '../../controllers/useProcurement'
import { useContractSummary, useUnitSummary } from '../../controllers/useSales'
import type { ModuleKey } from '../../models/auth'
import { EMPLOYMENT_TYPE_LABEL } from '../../models/hr'
import { percentOf, unitFunnel } from '../../models/overview'
import { PROJECT_STATUS_LABEL, type ProjectStatus } from '../../models/project'
import { number, rupiahShort } from '../../shared/format'
import { MoneyText } from '../components/Money'

interface Stage {
  module: ModuleKey
  moduleLabel: string
  to: string
  title: string
  icon: LucideIcon
  body: () => ReactNode
}

// Satu baris yang membaca bisnis dari hulu ke hilir: lahan dinilai, jadi
// proyek, material dibeli dan disimpan, tenaga kerja dibayar, unit dijual.
const STAGES: Stage[] = [
  { module: 'locations', moduleLabel: 'Analisis Lokasi', to: '/lokasi', title: 'Lokasi', icon: MapPinned, body: () => <LocationBody /> },
  { module: 'projects', moduleLabel: 'Proyek', to: '/proyek', title: 'Proyek', icon: FolderKanban, body: () => <ProjectBody /> },
  { module: 'procurement', moduleLabel: 'Pengadaan', to: '/pengadaan', title: 'Pengadaan', icon: ShoppingCart, body: () => <ProcurementBody /> },
  { module: 'inventory', moduleLabel: 'Inventaris', to: '/inventaris', title: 'Inventaris', icon: Boxes, body: () => <InventoryBody /> },
  { module: 'hr', moduleLabel: 'SDM', to: '/sdm', title: 'SDM', icon: Users, body: () => <HrBody /> },
  { module: 'sales', moduleLabel: 'Penjualan', to: '/penjualan', title: 'Penjualan', icon: Handshake, body: () => <SalesBody /> },
]

const SEGMENT_TONES = ['bg-navy-700', 'bg-navy-500', 'bg-navy-200', 'bg-slate-300', 'bg-slate-200']

export function BusinessFlow() {
  return (
    <ol className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-6">
      {STAGES.map((stage, index) => (
        <StageCard key={stage.module} stage={stage} step={index + 1} last={index === STAGES.length - 1} />
      ))}
    </ol>
  )
}

function StageCard({ stage, step, last }: { stage: Stage; step: number; last: boolean }) {
  const allowed = useCan(stage.module, 'read')
  const Icon = stage.icon
  const header = (
    <div className="flex items-center gap-2.5">
      <span className={`grid size-8 shrink-0 place-items-center rounded-lg ${allowed ? 'bg-navy-50 text-navy-600' : 'bg-slate-100 text-slate-400'}`}>
        <Icon aria-hidden="true" className="size-4" strokeWidth={1.75} />
      </span>
      <div className="min-w-0">
        <p className="text-[11px] font-medium text-slate-500 tabular-nums">Tahap {step}</p>
        <p className="truncate text-[14px] font-semibold text-slate-900">{stage.title}</p>
      </div>
    </div>
  )

  return (
    <li className="relative min-w-0">
      {allowed ? (
        <Link
          to={stage.to}
          className="group flex h-full flex-col gap-4 rounded-2xl bg-white p-4 shadow-hairline transition-shadow
                     hover:shadow-[0_0_0_1px_rgb(0_0_0/0.06),0_6px_20px_rgb(15_23_42/0.08)]"
        >
          {header}
          {stage.body()}
          <span className="mt-auto inline-flex items-center gap-0.5 text-xs font-medium text-navy-600">
            Buka {stage.moduleLabel}
            <ChevronRight aria-hidden="true" className="size-3.5 transition-transform group-hover:translate-x-0.5" strokeWidth={2} />
          </span>
        </Link>
      ) : (
        <div className="flex h-full flex-col gap-4 rounded-2xl bg-slate-50 p-4 shadow-hairline">
          {header}
          <p className="flex items-center gap-1.5 text-xs text-slate-500">
            <Lock aria-hidden="true" className="size-3.5" strokeWidth={1.75} />
            Butuh akses {stage.moduleLabel}
          </p>
        </div>
      )}
      {last ? null : (
        <span
          aria-hidden="true"
          className="absolute top-1/2 -right-[13px] z-10 hidden size-[22px] -translate-y-1/2 place-items-center rounded-full bg-white
                     text-slate-400 shadow-hairline 2xl:grid"
        >
          <ChevronRight className="size-3.5" strokeWidth={2} />
        </span>
      )}
    </li>
  )
}

function LocationBody() {
  const summary = useSavedLocationSummary()
  const data = summary.data
  return (
    <StageFigure
      value={data ? number(data.count) : placeholder(summary.isError)}
      unit="lokasi dievaluasi"
      visual={data ? <Meter percent={data.averageScore} /> : null}
      detail={data ? `skor rata-rata ${data.averageScore} dari 100` : undefined}
    />
  )
}

function ProjectBody() {
  const summary = useProjectSummary()
  const data = summary.data
  const order: ProjectStatus[] = ['ongoing', 'planning', 'completed', 'on_hold', 'cancelled']
  const parts = data ? order.map((status) => ({ label: PROJECT_STATUS_LABEL[status], value: data.byStatus[status] ?? 0 })) : []
  return (
    <StageFigure
      value={data ? number(data.count) : placeholder(summary.isError)}
      unit="proyek"
      visual={data ? <Segments parts={parts} /> : null}
      detail={data ? describe(parts) : undefined}
    />
  )
}

function ProcurementBody() {
  const summary = usePurchaseOrderSummary()
  const data = summary.data
  return (
    <StageFigure
      value={data ? number(data.openCount) : placeholder(summary.isError)}
      unit="PO terbuka"
      visual={data ? <Meter percent={percentOf(data.openValue, data.totalValue)} /> : null}
      detail={data ? `${rupiahShort(data.openValue)} belum diterima` : undefined}
    />
  )
}

function InventoryBody() {
  const materials = useMaterials({ pageSize: 1 })
  const lowStock = useLowStockMaterials(1)
  const total = materials.data?.totalItems
  const low = lowStock.data?.totalItems
  const ready = total !== undefined && low !== undefined
  return (
    <StageFigure
      value={total !== undefined ? number(total) : placeholder(materials.isError)}
      unit="jenis material"
      visual={ready ? <Meter percent={percentOf(total - low, total)} /> : null}
      detail={ready ? `${number(low)} stoknya menipis` : undefined}
    />
  )
}

function HrBody() {
  const headcount = useEmployeeHeadcount()
  const data = headcount.data
  const parts = data
    ? data.byType
        .map((row) => ({ label: EMPLOYMENT_TYPE_LABEL[row.employmentType], value: row.total }))
        .sort((left, right) => right.value - left.value)
    : []
  return (
    <StageFigure
      value={data ? number(data.total) : placeholder(headcount.isError)}
      unit="karyawan"
      visual={data ? <Segments parts={parts} /> : null}
      detail={data ? describe(parts) : undefined}
    />
  )
}

function SalesBody() {
  const units = useUnitSummary()
  const contracts = useContractSummary()
  const stages = units.data ? unitFunnel(units.data, contracts.data) : undefined
  return (
    <StageFigure
      value={stages ? number(stages[3].value) : placeholder(units.isError)}
      unit="unit terjual"
      visual={stages ? <Meter percent={percentOf(stages[3].value, stages[0].value)} /> : null}
      detail={contracts.data ? `nilai kontrak ${rupiahShort(contracts.data.totalValue)}` : undefined}
    />
  )
}

function StageFigure({ value, unit, visual, detail }: { value: string; unit: string; visual: ReactNode; detail?: string }) {
  return (
    <div className="space-y-2.5">
      <p className="flex items-baseline gap-1.5">
        <span className="text-[26px] leading-none font-semibold tracking-[-0.03em] text-slate-900 tabular-nums">{value}</span>
        <span className="text-xs text-slate-500">{unit}</span>
      </p>
      {visual}
      {detail ? (
        <p className="text-xs text-slate-500">
          <MoneyText>{detail}</MoneyText>
        </p>
      ) : null}
    </div>
  )
}

function Meter({ percent }: { percent: number }) {
  const safe = Math.max(0, Math.min(100, percent))
  return (
    <span aria-hidden="true" className="block h-1.5 overflow-hidden rounded-[2px] bg-slate-100">
      <span className="bar-fill block h-full rounded-[2px]" style={{ width: `${safe}%` }} />
    </span>
  )
}

function Segments({ parts }: { parts: { label: string; value: number }[] }) {
  const total = parts.reduce((sum, part) => sum + part.value, 0)
  if (total === 0) {
    return <span aria-hidden="true" className="block h-1.5 rounded-[2px] bg-slate-100" />
  }
  return (
    <span aria-hidden="true" className="flex h-1.5 gap-[2px] overflow-hidden rounded-[2px]">
      {parts.map((part, index) =>
        part.value > 0 ? (
          <span
            key={part.label}
            className={`block h-full ${SEGMENT_TONES[index % SEGMENT_TONES.length]}`}
            style={{ width: `${(part.value / total) * 100}%` }}
          />
        ) : null,
      )}
    </span>
  )
}

function describe(parts: { label: string; value: number }[]): string {
  return parts
    .filter((part) => part.value > 0)
    .map((part) => `${number(part.value)} ${part.label.toLowerCase()}`)
    .join(', ')
}

function placeholder(failed: boolean): string {
  return failed ? '—' : '...'
}
