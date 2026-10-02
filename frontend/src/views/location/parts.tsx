import type { ReactNode } from 'react'

import { RISK_LEVEL_LABEL, type RiskFlag, type RiskLevel } from '../../models/location'
import { Chip } from '../components/ListTools'

const FULL_SCORE = 100

const RISK_TONE: Record<RiskLevel, string> = {
  high: 'bg-red-100 text-red-800',
  medium: 'bg-amber-100 text-amber-800',
}

// Penanda nilai paling unggul di satu baris. Berupa teks, bukan warna saja,
// supaya tetap terbaca oleh pembaca layar dan saat dicetak hitam putih.
export function BestMark() {
  return (
    <span className="rounded-full bg-navy-50 px-1.5 py-px text-[11px] font-medium text-navy-700">terbaik</span>
  )
}

export function ValueWithMark({ isBest, children }: { isBest: boolean; children: ReactNode }) {
  return (
    <span className="inline-flex flex-wrap items-center gap-1.5">
      <span className={`tabular-nums ${isBest ? 'font-semibold text-navy-700' : 'text-slate-700'}`}>{children}</span>
      {isBest ? <BestMark /> : null}
    </span>
  )
}

// Angka di atas, batang di bawah, supaya batang di satu kolom tetap sejajar
// walaupun sebagian baris punya penanda terbaik.
export function ScoreCell({ value, isBest }: { value: number | null; isBest: boolean }) {
  if (value === null) {
    return <span className="text-slate-400">Tidak ada data</span>
  }

  const safe = Math.max(0, Math.min(FULL_SCORE, Math.round(value)))

  return (
    <div className="min-w-28">
      <ValueWithMark isBest={isBest}>{safe}</ValueWithMark>
      <span aria-hidden="true" className="mt-1.5 block h-1.5 overflow-hidden rounded-full bg-slate-100">
        <span
          className={`block h-full rounded-full ${isBest ? 'bg-navy-700' : 'bg-navy-500/60'}`}
          style={{ width: `${safe}%` }}
        />
      </span>
    </div>
  )
}

export function RiskChips({ flags }: { flags: RiskFlag[] }) {
  if (flags.length === 0) {
    return <span className="text-slate-500">Tidak ada</span>
  }

  return (
    <span className="flex flex-col items-start gap-1">
      {flags.map((flag) => (
        <Chip key={flag.key} label={`Risiko ${flag.label} ${RISK_LEVEL_LABEL[flag.level]}`} tone={RISK_TONE[flag.level]} />
      ))}
    </span>
  )
}
