import { useId, useLayoutEffect, useRef, useState } from 'react'

import { monthLabel, rupiahShort } from '../../shared/format'
import { areaPath, smoothPath, type Point } from './chartPath'
import { Money } from './Money'

export interface MonthlyValue {
  period: string
  value: number
}

const HEIGHT = 220
const PAD_TOP = 14
const PAD_BOTTOM = 28
const PAD_LEFT = 64
const PAD_RIGHT = 12
const TICKS = 4
const MIN_LABEL_SLOT = 32

// Garis per bulan dengan isian gradasi dan garis rata-rata putus-putus, jadi
// bulan yang melonjak langsung terlihat dibanding kebiasaan proyeknya.
export function MonthlyLineChart({ points, label }: { points: MonthlyValue[]; label: string }) {
  const gradient = `${useId()}-area`
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

  const scale = niceCeiling(Math.max(1, ...points.map((point) => point.value)))
  const average = points.length > 0 ? points.reduce((sum, point) => sum + point.value, 0) / points.length : 0
  const plotHeight = HEIGHT - PAD_TOP - PAD_BOTTOM
  const baseline = PAD_TOP + plotHeight
  const plotWidth = Math.max(0, width - PAD_LEFT - PAD_RIGHT)
  const step = points.length > 1 ? plotWidth / (points.length - 1) : 0
  const labelEvery = step >= MIN_LABEL_SLOT ? 1 : Math.ceil(MIN_LABEL_SLOT / Math.max(step, 1))
  const ticks = Array.from({ length: TICKS + 1 }, (_, index) => scale - (index * scale) / TICKS)

  const toX = (index: number) => PAD_LEFT + step * index
  const toY = (value: number) => baseline - (value / scale) * plotHeight
  const line: Point[] = points.map((point, index) => [toX(index), toY(point.value)])
  const selected = active === null ? null : points[active]

  return (
    <div>
      <div ref={frameRef} className="relative" onPointerLeave={() => setActive(null)}>
        <svg width={width} height={HEIGHT} role="img" aria-label={label} className="block overflow-visible">
          <defs>
            <linearGradient id={gradient} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--color-navy-500)" stopOpacity={0.22} />
              <stop offset="100%" stopColor="var(--color-navy-500)" stopOpacity={0} />
            </linearGradient>
          </defs>

          {ticks.map((tick) => (
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
                className="fill-slate-400 font-uang text-[11px] tabular-nums"
              >
                {tick === 0 ? '0' : rupiahShort(tick)}
              </text>
            </g>
          ))}

          {line.length > 1 ? (
            <>
              <path d={areaPath(line, baseline)} fill={`url(#${gradient})`} className="motion-safe:animate-fade" />
              <path
                d={smoothPath(line)}
                pathLength={1}
                fill="none"
                className="stroke-navy-600 [stroke-dasharray:1] motion-safe:animate-draw"
                strokeWidth={2.25}
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </>
          ) : null}

          {average > 0 ? (
            <g className="pointer-events-none">
              <line
                x1={PAD_LEFT}
                x2={width - PAD_RIGHT}
                y1={toY(average)}
                y2={toY(average)}
                className="stroke-amber-merek"
                strokeWidth={1.25}
                strokeDasharray="4 4"
              />
              <text
                x={width - PAD_RIGHT}
                y={toY(average) - 6}
                textAnchor="end"
                paintOrder="stroke"
                strokeWidth={4}
                className="fill-slate-500 stroke-white text-[11px]"
              >
                rata-rata {rupiahShort(average)}
              </text>
            </g>
          ) : null}

          {points.map((point, index) => {
            const isActive = active === index
            const showLabel = isActive || (points.length - 1 - index) % labelEvery === 0
            const slotWidth = Math.max(step, 1)
            return (
              <g key={point.period}>
                <rect
                  x={toX(index) - slotWidth / 2}
                  y={PAD_TOP}
                  width={slotWidth}
                  height={plotHeight}
                  className="fill-transparent"
                  onPointerEnter={() => setActive(index)}
                />
                {showLabel ? (
                  <text
                    x={toX(index)}
                    y={HEIGHT - 8}
                    textAnchor="middle"
                    className={`text-[11px] ${isActive ? 'fill-slate-900 font-medium' : 'fill-slate-400'}`}
                  >
                    {shortMonth(point.period)}
                  </text>
                ) : null}
              </g>
            )
          })}

          {selected && active !== null ? (
            <g className="pointer-events-none">
              <line x1={toX(active)} x2={toX(active)} y1={PAD_TOP} y2={baseline} className="stroke-slate-300" strokeWidth={1} />
              <circle cx={toX(active)} cy={toY(selected.value)} r={8} className="fill-navy-500/15" />
              <circle cx={toX(active)} cy={toY(selected.value)} r={4} className="fill-white stroke-navy-600" strokeWidth={2} />
            </g>
          ) : null}
        </svg>

        {selected && active !== null ? (
          <div
            className="pointer-events-none absolute top-1 z-10 w-44 rounded-xl bg-white/75 p-3 text-xs shadow-panel ring-1 ring-black/5 backdrop-blur-xl backdrop-saturate-150"
            style={{ left: tooltipLeft(toX(active), width) }}
          >
            <p className="font-semibold text-slate-900">{monthLabel(selected.period)}</p>
            <p className="mt-1 flex justify-between gap-3">
              <span className="text-slate-500">Kas keluar</span>
              <span className="font-uang font-medium tabular-nums text-slate-900">{rupiahShort(selected.value)}</span>
            </p>
          </div>
        ) : null}
      </div>

      <table className="sr-only">
        <caption>{label}</caption>
        <tbody>
          {points.map((point) => (
            <tr key={point.period}>
              <th scope="row">{monthLabel(point.period)}</th>
              <td>
                <Money>{rupiahShort(point.value)}</Money>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function shortMonth(period: string): string {
  const [year, month] = period.split('-')
  return new Date(Number(year), Number(month) - 1).toLocaleDateString('id-ID', { month: 'short' })
}

function tooltipLeft(center: number, width: number): number {
  const tooltipWidth = 176
  return Math.min(Math.max(0, center - tooltipWidth / 2), Math.max(0, width - tooltipWidth))
}

function niceCeiling(value: number): number {
  const magnitude = 10 ** Math.floor(Math.log10(value))
  const steps = [1, 2, 2.5, 5, 10]
  const step = steps.find((candidate) => candidate * magnitude >= value) ?? 10
  return step * magnitude
}
