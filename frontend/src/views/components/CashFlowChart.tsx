import { useLayoutEffect, useRef, useState } from 'react'

import { monthLabel, rupiahShort } from '../../shared/format'

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

// Masuk digambar ke atas dan keluar ke bawah dari garis nol, jadi bulan yang
// tekor langsung terlihat tanpa harus membandingkan dua batang berdampingan.
export function CashFlowChart({ periods }: { periods: CashFlowBar[] }) {
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
  const barWidth = Math.min(28, slot * 0.5)
  const ticks = Array.from({ length: TICKS + 1 }, (_, index) => scale - (index * 2 * scale) / TICKS)
  const selected = active === null ? null : periods[active]

  const toY = (value: number) => zeroY - (value / scale) * half

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500">
        <Legend className="bg-navy-600" label="Kas masuk" />
        <Legend className="bg-slate-300" label="Kas keluar" />
        <Legend className="bg-amber-merek" label="Selisih" isDot />
      </div>

      <div ref={frameRef} className="relative" onPointerLeave={() => setActive(null)}>
        <svg
          width={width}
          height={HEIGHT}
          role="img"
          aria-label={`Arus kas ${periods.length} bulan terakhir`}
          className="block overflow-visible"
        >
          {ticks.map((tick) => (
            <g key={tick}>
              <line
                x1={PAD_LEFT}
                x2={width}
                y1={toY(tick)}
                y2={toY(tick)}
                className={tick === 0 ? 'stroke-slate-300' : 'stroke-slate-100'}
                strokeWidth={1}
              />
              <text
                x={PAD_LEFT - 10}
                y={toY(tick)}
                textAnchor="end"
                dominantBaseline="middle"
                className="fill-slate-500 text-[11px] tabular-nums"
              >
                {tick === 0 ? '0' : rupiahShort(tick)}
              </text>
            </g>
          ))}

          {periods.map((period, index) => {
            const center = PAD_LEFT + slot * index + slot / 2
            const isActive = active === index
            const dimmed = active !== null && !isActive
            return (
              <g key={period.period} className="transition-opacity duration-150" opacity={dimmed ? 0.45 : 1}>
                <rect
                  x={PAD_LEFT + slot * index}
                  y={PAD_TOP}
                  width={slot}
                  height={plotHeight}
                  className={isActive ? 'fill-slate-50' : 'fill-transparent'}
                  onPointerEnter={() => setActive(index)}
                />
                <rect
                  x={center - barWidth / 2}
                  y={toY(period.cashIn)}
                  width={barWidth}
                  height={Math.max(0, zeroY - toY(period.cashIn))}
                  rx={3}
                  className="pointer-events-none fill-navy-600"
                />
                <rect
                  x={center - barWidth / 2}
                  y={zeroY}
                  width={barWidth}
                  height={Math.max(0, toY(-period.cashOut) - zeroY)}
                  rx={3}
                  className="pointer-events-none fill-slate-300"
                />
                {period.cashIn !== 0 || period.cashOut !== 0 ? (
                  <circle
                    cx={center}
                    cy={toY(period.net)}
                    r={3.5}
                    className="pointer-events-none fill-amber-merek stroke-white"
                    strokeWidth={1.5}
                  />
                ) : null}
                <text
                  x={center}
                  y={HEIGHT - 8}
                  textAnchor="middle"
                  className={`text-[11px] ${isActive ? 'fill-slate-900 font-medium' : 'fill-slate-500'}`}
                >
                  {shortMonth(period.period)}
                </text>
              </g>
            )
          })}
        </svg>

        {selected && active !== null ? (
          <div
            className="pointer-events-none absolute top-1 z-10 w-48 rounded-xl bg-white/95 p-3 text-xs shadow-panel backdrop-blur"
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
              <td>{rupiahShort(period.cashIn)}</td>
              <td>{rupiahShort(period.cashOut)}</td>
              <td>{rupiahShort(period.net)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function Legend({ className, label, isDot = false }: { className: string; label: string; isDot?: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span aria-hidden="true" className={`${isDot ? 'size-2 rounded-full' : 'h-2 w-3 rounded-sm'} ${className}`} />
      {label}
    </span>
  )
}

function TooltipRow({ label, value, tone = 'text-slate-900' }: { label: string; value: string; tone?: string }) {
  return (
    <p className="mt-1 flex justify-between gap-3">
      <span className="text-slate-500">{label}</span>
      <span className={`font-medium tabular-nums ${tone}`}>{value}</span>
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
