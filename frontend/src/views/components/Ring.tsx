import { useId } from 'react'

const STROKE = 9

// Cincin progres bergaya Apple Activity: lintasan tipis abu, isian gradasi
// dengan ujung membulat, angka di tengah.
export function Ring({ value, size = 96, label, caption }: {
  value: number
  size?: number
  label: string
  caption?: string
}) {
  const gradientId = useId()
  const safe = Math.max(0, Math.min(100, value))
  const radius = (size - STROKE) / 2
  const circumference = 2 * Math.PI * radius

  return (
    <div className="relative grid shrink-0 place-items-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="-rotate-90" role="img" aria-label={`${label}: ${safe}`}>
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stopColor="var(--color-navy-500)" />
            <stop offset="100%" stopColor="var(--color-navy-700)" />
          </linearGradient>
        </defs>
        <circle cx={size / 2} cy={size / 2} r={radius} fill="none" className="stroke-slate-100" strokeWidth={STROKE} />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={`url(#${gradientId})`}
          strokeWidth={STROKE}
          strokeLinecap="round"
          strokeDasharray={`${(safe / 100) * circumference} ${circumference}`}
          className="motion-safe:transition-[stroke-dasharray] motion-safe:duration-700"
        />
      </svg>
      <div className="absolute inset-0 grid place-items-center text-center">
        <div>
          <p className="text-[22px] leading-none font-semibold tracking-[-0.03em] text-slate-900 tabular-nums">{label}</p>
          {caption ? <p className="mt-1 text-[11px] text-slate-500">{caption}</p> : null}
        </div>
      </div>
    </div>
  )
}
