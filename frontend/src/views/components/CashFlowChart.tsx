import { useId, useLayoutEffect, useRef, useState } from 'react'

import { monthLabel, rupiahShort } from '../../shared/format'
import { barPath, smoothPath, type Point } from './chartPath'
import { Money } from './Money'

export interface CashFlowBar {
  period: string
  cashIn: number
  cashOut: number
  net: number
}

const HEIGHT = 220
const PAD_TOP = 12
const PAD_BOTTOM = 28
const PAD_LEFT = 64
const TICKS = 4
const MIN_LABEL_SLOT = 32
const MIN_BAR_HEIGHT = 2

// Masuk digambar ke atas dan keluar ke bawah dari garis nol, jadi bulan yang
// tekor langsung terlihat tanpa harus membandingkan dua batang berdampingan.
export function CashFlowChart({ periods }: { periods: CashFlowBar[] }) {
  const idPrefix = useId()
  const inGradient = `${idPrefix}-masuk`
  const outGradient = `${idPrefix}-keluar`
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

  const peak = Math.max(1, ...periods.map((period) => Math.max(period.cashIn, period.cashOut)))
  const scale = niceCeiling(peak)
  const plotHeight = HEIGHT - PAD_TOP - PAD_BOTTOM
  const zeroY = PAD_TOP + plotHeight / 2
  const half = plotHeight / 2
  const plotWidth = Math.max(0, width - PAD_LEFT)
  const slot = periods.length > 0 ? plotWidth / periods.length : 0
  const barWidth = Math.min(16, slot * 0.4)
  const labelEvery = slot >= MIN_LABEL_SLOT ? 1 : Math.ceil(MIN_LABEL_SLOT / Math.max(slot, 1))
  const ticks = Array.from({ length: TICKS + 1 }, (_, index) => scale - (index * 2 * scale) / TICKS)
  const selected = active === null ? null : periods[active]

  const toY = (value: number) => zeroY - (value / scale) * half
  const centerOf = (index: number) => PAD_LEFT + slot * index + slot / 2
  const barHeight = (amount: number) => (amount > 0 ? Math.max(MIN_BAR_HEIGHT, (amount / scale) * half) : 0)
  const netPoints: Point[] = periods.map((period, index) => [centerOf(index), toY(period.net)])

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500">
        <Legend className="bar-fill" label="Kas masuk" />
        <Legend className="bg-slate-300" label="Kas keluar" />
        <Legend className="bg-amber-merek" label="Selisih" isLine />
      </div>

      <div ref={frameRef} className="relative" onPointerLeave={() => setActive(null)}>
        <svg
          width={width}
          height={HEIGHT}
          role="img"
          aria-label={`Arus kas ${periods.length} bulan terakhir`}
          className="block overflow-visible"
        >
          <defs>
            <linearGradient id={inGradient} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--color-navy-500)" />
              <stop offset="100%" stopColor="var(--color-navy-700)" />
            </linearGradient>
            <linearGradient id={outGradient} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--color-slate-300)" />
              <stop offset="100%" stopColor="var(--color-slate-200)" />
            </linearGradient>
          </defs>

          {ticks.map((tick) => (
            <g key={tick}>
              <line
                x1={PAD_LEFT}
                x2={width}
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
                className="fill-slate-400 font-uang text-[11px] tabular-nums"
              >
                {tick === 0 ? '0' : rupiahShort(tick)}
              </text>
            </g>
          ))}

          {periods.map((period, index) => {
            const center = centerOf(index)
            const isActive = active === index
            const dimmed = active !== null && !isActive
            const showLabel = isActive || (periods.length - 1 - index) % labelEvery === 0
            return (
              <g key={period.period} className="transition-opacity duration-200" opacity={dimmed ? 0.4 : 1}>
                <rect
                  x={PAD_LEFT + slot * index + 2}
                  y={PAD_TOP}
                  width={Math.max(0, slot - 4)}
                  height={plotHeight}
                  rx={8}
                  className={isActive ? 'fill-slate-100/70' : 'fill-transparent'}
                  onPointerEnter={() => setActive(index)}
                />
                <path
                  d={barPath(center - barWidth / 2, zeroY - barHeight(period.cashIn), barWidth, barHeight(period.cashIn), 'up')}
                  fill={`url(#${inGradient})`}
                  className="pointer-events-none origin-bottom [transform-box:fill-box] motion-safe:animate-grow"
                  style={{ animationDelay: `${index * 30}ms` }}
                />
                <path
                  d={barPath(center - barWidth / 2, zeroY, barWidth, barHeight(period.cashOut), 'down')}
                  fill={`url(#${outGradient})`}
                  className="pointer-events-none origin-top [transform-box:fill-box] motion-safe:animate-grow"
                  style={{ animationDelay: `${index * 30}ms` }}
                />
                {showLabel ? (
                  <text
                    x={center}
                    y={HEIGHT - 8}
                    textAnchor="middle"
                    className={`text-[11px] ${isActive ? 'fill-slate-900 font-medium' : 'fill-slate-400'}`}
                  >
                    {shortMonth(period.period)}
                  </text>
                ) : null}
              </g>
            )
          })}

          {netPoints.length > 1 ? (
            <path
              d={smoothPath(netPoints)}
              pathLength={1}
              fill="none"
              className="pointer-events-none stroke-amber-merek [stroke-dasharray:1] motion-safe:animate-draw"
              strokeWidth={2}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          ) : null}
          {selected && active !== null ? (
            <g className="pointer-events-none">
              <circle cx={centerOf(active)} cy={toY(selected.net)} r={8} className="fill-amber-merek/20" />
              <circle cx={centerOf(active)} cy={toY(selected.net)} r={4} className="fill-white stroke-amber-merek" strokeWidth={2} />
            </g>
          ) : null}
        </svg>

        {selected && active !== null ? (
          <div
            className="pointer-events-none absolute top-1 z-10 w-48 rounded-xl bg-white/75 p-3 text-xs shadow-panel ring-1 ring-black/5 backdrop-blur-xl backdrop-saturate-150"
            style={{ left: tooltipLeft(PAD_LEFT + slot * active + slot / 2, width) }}
          >
            <p className="font-semibold text-slate-900">{monthLabel(selected.period)}</p>
            <TooltipRow label="Masuk" value={rupiahShort(selected.cashIn)} />
            <TooltipRow label="Keluar" value={rupiahShort(selected.cashOut)} />
            <TooltipRow
              label="Selisih"
              value={rupiahShort(selected.net)}
              tone={selected.net < 0 ? 'text-red-600' : 'text-green-700'}
            />
          </div>
        ) : null}
      </div>

      <table className="sr-only">
        <caption>Arus kas per bulan</caption>
        <thead>
          <tr>
            <th scope="col">Periode</th>
            <th scope="col">Masuk</th>
            <th scope="col">Keluar</th>
            <th scope="col">Selisih</th>
          </tr>
        </thead>
        <tbody>
          {periods.map((period) => (
            <tr key={period.period}>
              <th scope="row">{monthLabel(period.period)}</th>
              <td><Money>{rupiahShort(period.cashIn)}</Money></td>
              <td><Money>{rupiahShort(period.cashOut)}</Money></td>
              <td><Money>{rupiahShort(period.net)}</Money></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function Legend({ className, label, isLine = false }: { className: string; label: string; isLine?: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span aria-hidden="true" className={`${isLine ? 'h-0.5 w-3.5' : 'h-2.5 w-2'} rounded-full ${className}`} />
      {label}
    </span>
  )
}

function TooltipRow({ label, value, tone = 'text-slate-900' }: { label: string; value: string; tone?: string }) {
  return (
    <p className="mt-1 flex justify-between gap-3">
      <span className="text-slate-500">{label}</span>
      <span className={`font-uang font-medium tabular-nums ${tone}`}>{value}</span>
    </p>
  )
}

function shortMonth(period: string): string {
  const [year, month] = period.split('-')
  return new Date(Number(year), Number(month) - 1).toLocaleDateString('id-ID', { month: 'short' })
}

function tooltipLeft(center: number, width: number): number {
  const tooltipWidth = 192
  return Math.min(Math.max(0, center - tooltipWidth / 2), Math.max(0, width - tooltipWidth))
}

function niceCeiling(value: number): number {
  const magnitude = 10 ** Math.floor(Math.log10(value))
  const steps = [1, 2, 2.5, 5, 10]
  const step = steps.find((candidate) => candidate * magnitude >= value) ?? 10
  return step * magnitude
}
