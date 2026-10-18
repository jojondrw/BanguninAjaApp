import { useId } from 'react'

import { areaPath, smoothPath, type Point } from './chartPath'

const WIDTH = 120
const HEIGHT = 36
const PAD = 3

// Garis kecil bergaya Apple Stocks: kurva halus dengan area yang memudar ke
// transparan dan titik di nilai terakhir. Skala mengikuti nilai min sampai max.
export function Sparkline({ values, tone = 'navy' }: { values: number[]; tone?: 'navy' | 'red' }) {
  const gradientId = useId()
  if (values.length < 2) {
    return null
  }
  const low = Math.min(...values)
  const high = Math.max(...values)
  const span = high - low || 1
  const points: Point[] = values.map((value, index) => [
    PAD + (index / (values.length - 1)) * (WIDTH - PAD * 2),
    PAD + (1 - (value - low) / span) * (HEIGHT - PAD * 2),
  ])
  const last = points[points.length - 1]
  const stroke = tone === 'red' ? 'var(--color-red-500)' : 'var(--color-navy-600)'

  return (
    <svg width={WIDTH} height={HEIGHT} viewBox={`0 0 ${WIDTH} ${HEIGHT}`} aria-hidden="true" className="block overflow-visible">
      <defs>
        <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={stroke} stopOpacity={0.22} />
          <stop offset="100%" stopColor={stroke} stopOpacity={0} />
        </linearGradient>
      </defs>
      <path d={areaPath(points, HEIGHT)} fill={`url(#${gradientId})`} />
      <path d={smoothPath(points)} fill="none" stroke={stroke} strokeWidth={1.75} strokeLinecap="round" strokeLinejoin="round" />
      <circle cx={last[0]} cy={last[1]} r={2.75} fill={stroke} stroke="white" strokeWidth={1.5} />
    </svg>
  )
}
