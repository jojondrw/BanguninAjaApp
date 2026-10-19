import { useLayoutEffect, useRef, useState } from 'react'

import { useLeadFlows } from '../../controllers/useSales'
import { LEAD_STAGE_LABEL, type LeadStage } from '../../models/sales'
import { number } from '../../shared/format'
import { Empty, LoadFailed, Loading } from '../components/Data'
import { layoutSankey, ribbonPath, type SankeyLink } from './sankeyLayout'

const HEIGHT = 300
const GAP = 8
const NODE_WIDTH = 10
const NARROW = 480

const STAGE_COLOR: Record<LeadStage, string> = {
  new: 'var(--color-slate-400)',
  interested: 'var(--color-navy-500)',
  negotiating: 'var(--color-amber-400)',
  won: 'var(--color-green-500)',
  cancelled: 'var(--color-red-400)',
}

function percent(part: number, whole: number): number {
  return whole > 0 ? Math.round((part / whole) * 100) : 0
}

// Sankey dari sumber prospek ke tahapnya sekarang. Aplikasi tidak menyimpan
// riwayat perpindahan tahap, jadi pita menunjukkan posisi prospek hari ini,
// bukan jalur yang ditempuhnya.
export function LeadSankey({ onOpenStage }: { onOpenStage: (stage: LeadStage) => void }) {
  const flows = useLeadFlows()
  const frameRef = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(0)
  const [active, setActive] = useState<SankeyLink | null>(null)

  useLayoutEffect(() => {
    const frame = frameRef.current
    if (!frame) {
      return
    }
    const measure = () => setWidth(frame.clientWidth)
    measure()
    const observer = new ResizeObserver(measure)
    observer.observe(frame)
    return () => observer.disconnect()
  }, [flows.isSuccess])

  if (flows.isPending) {
    return <Loading />
  }
  if (flows.isError) {
    return <LoadFailed onRetry={() => flows.refetch()} />
  }
  const layout = layoutSankey(flows.data, HEIGHT, GAP)
  if (layout.total === 0) {
    return <Empty message="Belum ada prospek." />
  }

  const isNarrow = width < NARROW
  const leftLabel = isNarrow ? 96 : 176
  const rightLabel = isNarrow ? 86 : 132
  const sourceX = leftLabel
  const stageX = Math.max(sourceX + NODE_WIDTH + 40, width - rightLabel - NODE_WIDTH)
  const ribbonLeft = sourceX + NODE_WIDTH
  const sourceTotal = (source: string) => layout.sources.find((node) => node.key === source)?.total ?? 0
  const dims = (link: SankeyLink) =>
    active !== null && (active.source !== link.source || active.stage !== link.stage)

  return (
    <div>
      <div ref={frameRef} className="relative" onPointerLeave={() => setActive(null)}>
        {width > 0 ? (
          <svg width={width} height={HEIGHT} role="img" aria-label="Sankey sumber prospek ke tahap sekarang" className="block">
            {layout.links.map((link) => (
              <path
                key={`${link.source}|${link.stage}`}
                d={ribbonPath(link, ribbonLeft, stageX)}
                fill={STAGE_COLOR[link.stage]}
                fillOpacity={dims(link) ? 0.08 : active ? 0.55 : 0.3}
                className="cursor-pointer transition-[fill-opacity] duration-200"
                onPointerEnter={() => setActive(link)}
                onClick={() => onOpenStage(link.stage)}
              />
            ))}

            {layout.sources.map((node) => (
              <g key={node.key}>
                <rect x={sourceX} y={node.top} width={NODE_WIDTH} height={Math.max(node.height, 1)} rx={2} className="fill-slate-700" />
                <text
                  x={sourceX - 8}
                  y={node.top + node.height / 2}
                  textAnchor="end"
                  dominantBaseline="middle"
                  className="fill-slate-700 text-[12px]"
                >
                  {truncate(node.key, isNarrow ? 11 : 22)}
                  <tspan className="fill-slate-400 tabular-nums"> {number(node.total)}</tspan>
                </text>
              </g>
            ))}

            {layout.stages.map((node) => {
              const stage = node.key as LeadStage
              return (
                <g key={node.key} className="cursor-pointer" onClick={() => onOpenStage(stage)}>
                  <rect
                    x={stageX}
                    y={node.top}
                    width={NODE_WIDTH}
                    height={Math.max(node.height, 1)}
                    rx={2}
                    fill={STAGE_COLOR[stage]}
                  />
                  <text x={stageX + NODE_WIDTH + 8} y={node.top + node.height / 2} dominantBaseline="middle" className="text-[12px]">
                    <tspan className="fill-slate-900 font-medium">{LEAD_STAGE_LABEL[stage]}</tspan>
                    <tspan className="fill-slate-400 tabular-nums"> {number(node.total)}</tspan>
                  </text>
                </g>
              )
            })}
          </svg>
        ) : (
          <div style={{ height: HEIGHT }} />
        )}

        {active ? (
          <div className="pointer-events-none absolute top-1 left-1/2 z-10 w-60 -translate-x-1/2 rounded-xl bg-white/80 p-3 text-xs shadow-panel ring-1 ring-black/5 backdrop-blur-xl">
            <p className="font-semibold text-slate-900">
              {active.source} ke {LEAD_STAGE_LABEL[active.stage]}
            </p>
            <p className="mt-1 text-slate-600 tabular-nums">
              {number(active.total)} prospek, {percent(active.total, sourceTotal(active.source))}% dari sumber ini
            </p>
          </div>
        ) : null}
      </div>
      <p className="mt-3 text-xs text-slate-500">
        Posisi {number(layout.total)} prospek hari ini menurut sumbernya. Riwayat perpindahan tahap tidak dicatat, jadi pita
        tidak menunjukkan di tahap mana prospek yang batal berhenti.
      </p>

      <table className="sr-only">
        <caption>Jumlah prospek per sumber dan tahap</caption>
        <thead>
          <tr>
            <th scope="col">Sumber</th>
            <th scope="col">Tahap</th>
            <th scope="col">Prospek</th>
          </tr>
        </thead>
        <tbody>
          {layout.links.map((link) => (
            <tr key={`${link.source}|${link.stage}`}>
              <td>{link.source}</td>
              <td>{LEAD_STAGE_LABEL[link.stage]}</td>
              <td>{link.total}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function truncate(value: string, length: number): string {
  return value.length > length ? `${value.slice(0, length - 1)}…` : value
}
