import { useState } from 'react'

import { rupiahShort } from '../../shared/format'
import { Money } from './Money'

export interface DonutSlice {
  label: string
  value: number
}

const SIZE = 156
const STROKE = 16
const RADIUS = (SIZE - STROKE) / 2
const CIRCUMFERENCE = 2 * Math.PI * RADIUS
const GAP = 2

// Lingkaran donat dari stroke-dasharray: setiap irisan satu lingkaran yang
// digeser sejauh irisan sebelumnya. Celah kecil memisahkan irisan tanpa garis.
export function DonutChart({ slices, colors, caption }: { slices: DonutSlice[]; colors: string[]; caption: string }) {
  const [active, setActive] = useState<number | null>(null)
  const total = slices.reduce((sum, slice) => sum + slice.value, 0)
  if (total === 0) {
    return <p className="text-[13px] text-slate-500">Belum ada transaksi.</p>
  }

  const lengths = slices.map((slice) => (slice.value / total) * CIRCUMFERENCE)
  const arcs = lengths.map((length, index) => ({
    index,
    length: Math.max(0, length - (slices.length > 1 ? GAP : 0)),
    offset: lengths.slice(0, index).reduce((sum, previous) => sum + previous, 0),
  }))
  const focus = active === null ? null : slices[active]

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="relative" onPointerLeave={() => setActive(null)}>
        <svg width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} role="img" aria-label={caption} className="-rotate-90">
          <circle cx={SIZE / 2} cy={SIZE / 2} r={RADIUS} fill="none" strokeWidth={STROKE} className="stroke-slate-100" />
          {arcs.map((arc) => (
            <circle
              key={arc.index}
              cx={SIZE / 2}
              cy={SIZE / 2}
              r={RADIUS}
              fill="none"
              stroke={colors[arc.index % colors.length]}
              strokeWidth={active === arc.index ? STROKE + 4 : STROKE}
              strokeDasharray={`${arc.length} ${CIRCUMFERENCE - arc.length}`}
              strokeDashoffset={-arc.offset}
              opacity={active === null || active === arc.index ? 1 : 0.35}
              className="cursor-default transition-[opacity,stroke-width] duration-200"
              onPointerEnter={() => setActive(arc.index)}
            />
          ))}
        </svg>
        <div className="pointer-events-none absolute inset-0 grid place-items-center text-center">
          <div>
            <p className="text-[13px] font-semibold text-slate-900 tabular-nums">
              <Money>{rupiahShort(focus ? focus.value : total)}</Money>
            </p>
            <p className="text-[11px] text-slate-500">
              {focus ? `${Math.round((focus.value / total) * 100)}%` : 'total'}
            </p>
          </div>
        </div>
      </div>

      <ul className="w-full space-y-1.5">
        {slices.map((slice, index) => (
          <li
            key={slice.label}
            onPointerEnter={() => setActive(index)}
            onPointerLeave={() => setActive(null)}
            className={`flex items-start gap-2 rounded-md px-1.5 py-1 text-[13px] ${active === index ? 'bg-slate-50' : ''}`}
          >
            <span aria-hidden="true" className="mt-1.5 size-2 shrink-0 rounded-full" style={{ backgroundColor: colors[index % colors.length] }} />
            <span className="min-w-0 flex-1 text-slate-700">{slice.label}</span>
            <span className="shrink-0 text-slate-900 tabular-nums">
              <Money>{rupiahShort(slice.value)}</Money>
              <span className="ml-1.5 inline-block w-8 text-right text-xs text-slate-400">
                {Math.round((slice.value / total) * 100)}%
              </span>
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
