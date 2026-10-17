import { useId, useLayoutEffect, useRef, useState } from 'react'

import type { CurvePoint } from '../../models/project'
import { shortDate } from '../../shared/format'
import { areaPath, smoothPath, type Point } from '../components/chartPath'

const HEIGHT = 220
const PAD_TOP = 12
const PAD_BOTTOM = 28
const PAD_LEFT = 44
const PAD_RIGHT = 8
const TICKS = [0, 25, 50, 75, 100]
const MIN_MONTH_SLOT = 52
const TOOLTIP_WIDTH = 176

const PERCENT = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 1 })

function percent(value: number): string {
  return `${PERCENT.format(value)}%`
}

// Tanggal YYYY-MM-DD dibaca sebagai UTC supaya tidak bergeser sehari di WIB.
function dayValue(date: string): number {
  const [year, month, day] = date.split('-').map(Number)
  return Date.UTC(year, month - 1, day)
}

function monthStarts(start: number, end: number): number[] {
  const first = new Date(start)
  const months: number[] = []
  let cursor = Date.UTC(first.getUTCFullYear(), first.getUTCMonth() + (first.getUTCDate() === 1 ? 0 : 1), 1)
  while (cursor <= end) {
    months.push(cursor)
    const next = new Date(cursor)
    cursor = Date.UTC(next.getUTCFullYear(), next.getUTCMonth() + 1, 1)
  }
  return months
}

function monthTick(value: number): string {
  const date = new Date(value)
  const month = date.toLocaleDateString('id-ID', { month: 'short', timeZone: 'UTC' })
  return date.getUTCMonth() === 0 ? `${month} ${String(date.getUTCFullYear()).slice(2)}` : month
}

// Kurva S: garis putus-putus adalah rencana dari jadwal tahap, garis tegas
// adalah progres nyata yang tercatat. Jarak keduanya hari ini adalah deviasi.
export function ProgressCurveChart({ points, today }: { points: CurvePoint[]; today: string }) {
  const gradientId = useId()
  const frameRef = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(0)
  const [active, setActive] = useState<number | null>(null)

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
  }, [])

  const start = dayValue(points[0].date)
  const end = dayValue(points[points.length - 1].date)
  const span = Math.max(1, end - start)
  const plotWidth = Math.max(0, width - PAD_LEFT - PAD_RIGHT)
  const plotHeight = HEIGHT - PAD_TOP - PAD_BOTTOM

  const toX = (value: number) => PAD_LEFT + ((value - start) / span) * plotWidth
  const toY = (value: number) => PAD_TOP + plotHeight - (Math.min(100, Math.max(0, value)) / 100) * plotHeight

  const baseline = toY(0)
  const plannedPoints: Point[] = points.map((point) => [toX(dayValue(point.date)), toY(point.planned)])
  const recorded = points.filter((point) => point.actual !== null)
  const actualPoints: Point[] = recorded.map((point) => [toX(dayValue(point.date)), toY(point.actual ?? 0)])
  const lastRecorded = recorded.at(-1)

  const todayValue = dayValue(today)
  const showToday = todayValue >= start && todayValue <= end
  const months = monthStarts(start, end)
  const monthSlot = months.length > 0 ? plotWidth / Math.max(1, (end - start) / (30 * 86_400_000)) : 0
  const labelEvery = monthSlot >= MIN_MONTH_SLOT ? 1 : Math.ceil(MIN_MONTH_SLOT / Math.max(monthSlot, 1))
  const selected = active === null ? null : points[active]

  const pick = (clientX: number, left: number) => {
    const x = clientX - left
    let nearest = 0
    let distance = Number.POSITIVE_INFINITY
    points.forEach((point, index) => {
      const gap = Math.abs(toX(dayValue(point.date)) - x)
      if (gap < distance) {
        distance = gap
        nearest = index
      }
    })
    setActive(nearest)
  }

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500">
        <Legend label="Rencana" dashed />
        <Legend label="Realisasi" />
      </div>

      <div ref={frameRef} className="relative" onPointerLeave={() => setActive(null)}>
        <svg
          width={width}
          height={HEIGHT}
          role="img"
          aria-label="Kurva S progres rencana dan realisasi"
          className="block overflow-visible"
        >
          {TICKS.map((tick) => (
            <g key={tick}>
              <line
                x1={PAD_LEFT}
                x2={width - PAD_RIGHT}
                y1={toY(tick)}
                y2={toY(tick)}
                className="stroke-slate-200"
                strokeWidth={1}
                strokeDasharray={tick === 0 ? undefined : '2 4'}
              />
              <text
                x={PAD_LEFT - 10}
                y={toY(tick)}
                textAnchor="end"
                dominantBaseline="middle"
                className="fill-slate-400 text-[11px] tabular-nums"
              >
                {tick}%
              </text>
            </g>
          ))}

          {months.map((month, index) =>
            index % labelEvery === 0 ? (
              <text
                key={month}
                x={toX(month)}
                y={HEIGHT - 8}
                textAnchor="middle"
                className="fill-slate-400 text-[11px]"
              >
                {monthTick(month)}
              </text>
            ) : null,
          )}

          {showToday ? (
            <g>
              <line
                x1={toX(todayValue)}
                x2={toX(todayValue)}
                y1={PAD_TOP}
                y2={PAD_TOP + plotHeight}
                className="stroke-amber-merek"
                strokeWidth={1.5}
                strokeDasharray="3 3"
              />
              <text x={toX(todayValue) + 4} y={PAD_TOP + 10} className="fill-slate-500 text-[11px]">
                Hari ini
              </text>
            </g>
          ) : null}

          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--color-navy-500)" stopOpacity={0.28} />
              <stop offset="100%" stopColor="var(--color-navy-500)" stopOpacity={0} />
            </linearGradient>
          </defs>

          <path
            d={smoothPath(plannedPoints)}
            fill="none"
            className="stroke-slate-300"
            strokeWidth={1.75}
            strokeDasharray="4 5"
            strokeLinecap="round"
          />
          {recorded.length > 1 ? (
            <>
              <path d={areaPath(actualPoints, baseline)} fill={`url(#${gradientId})`} className="motion-safe:animate-fade" />
              <path
                d={smoothPath(actualPoints)}
                pathLength={1}
                fill="none"
                className="stroke-navy-600 [stroke-dasharray:1] motion-safe:animate-draw"
                strokeWidth={2.5}
                strokeLinejoin="round"
                strokeLinecap="round"
              />
            </>
          ) : null}
          {lastRecorded ? (
            <g>
              <circle cx={toX(dayValue(lastRecorded.date))} cy={toY(lastRecorded.actual ?? 0)} r={10} className="fill-navy-500/15" />
              <circle
                cx={toX(dayValue(lastRecorded.date))}
                cy={toY(lastRecorded.actual ?? 0)}
                r={4.5}
                className="fill-white stroke-navy-600"
                strokeWidth={2.5}
              />
            </g>
          ) : null}

          {selected ? (
            <g className="pointer-events-none">
              <line
                x1={toX(dayValue(selected.date))}
                x2={toX(dayValue(selected.date))}
                y1={PAD_TOP}
                y2={PAD_TOP + plotHeight}
                className="stroke-slate-300"
                strokeWidth={1}
                strokeDasharray="2 3"
              />
              <circle
                cx={toX(dayValue(selected.date))}
                cy={toY(selected.planned)}
                r={3.5}
                className="fill-white stroke-slate-400"
                strokeWidth={2}
              />
              {selected.actual !== null ? (
                <>
                  <circle cx={toX(dayValue(selected.date))} cy={toY(selected.actual)} r={9} className="fill-navy-500/15" />
                  <circle
                    cx={toX(dayValue(selected.date))}
                    cy={toY(selected.actual)}
                    r={4.5}
                    className="fill-white stroke-navy-600"
                    strokeWidth={2.5}
                  />
                </>
              ) : null}
            </g>
          ) : null}

          <rect
            x={PAD_LEFT}
            y={PAD_TOP}
            width={plotWidth}
            height={plotHeight}
            className="fill-transparent"
            onPointerMove={(event) => pick(event.clientX, event.currentTarget.getBoundingClientRect().left - PAD_LEFT)}
          />
        </svg>

        {selected && active !== null ? (
          <div
            className="pointer-events-none absolute top-1 z-10 rounded-xl bg-white/75 p-3 text-xs shadow-panel ring-1 ring-black/5 backdrop-blur-xl backdrop-saturate-150"
            style={{
              width: TOOLTIP_WIDTH,
              left: Math.min(
                Math.max(0, toX(dayValue(selected.date)) - TOOLTIP_WIDTH / 2),
                Math.max(0, width - TOOLTIP_WIDTH),
              ),
            }}
          >
            <p className="font-semibold text-slate-900">{shortDate(selected.date)}</p>
            <TooltipRow label="Rencana" value={percent(selected.planned)} />
            <TooltipRow label="Realisasi" value={selected.actual === null ? 'belum tercatat' : percent(selected.actual)} />
          </div>
        ) : null}
      </div>

      <table className="sr-only">
        <caption>Progres rencana dan realisasi per tanggal</caption>
        <thead>
          <tr>
            <th scope="col">Tanggal</th>
            <th scope="col">Rencana</th>
            <th scope="col">Realisasi</th>
          </tr>
        </thead>
        <tbody>
          {points.map((point) => (
            <tr key={point.date}>
              <th scope="row">{shortDate(point.date)}</th>
              <td>{percent(point.planned)}</td>
              <td>{point.actual === null ? 'belum tercatat' : percent(point.actual)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function Legend({ label, dashed = false }: { label: string; dashed?: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <svg aria-hidden="true" width={18} height={6}>
        <line
          x1={1}
          x2={17}
          y1={3}
          y2={3}
          className={dashed ? 'stroke-slate-300' : 'stroke-navy-600'}
          strokeWidth={2.5}
          strokeDasharray={dashed ? '4 3' : undefined}
          strokeLinecap="round"
        />
      </svg>
      {label}
    </span>
  )
}

function TooltipRow({ label, value }: { label: string; value: string }) {
  return (
    <p className="mt-1 flex justify-between gap-3">
      <span className="text-slate-500">{label}</span>
      <span className="font-medium text-slate-900 tabular-nums">{value}</span>
    </p>
  )
}
