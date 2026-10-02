import type { ReactNode } from 'react'

import { CHIP_CLASS } from '../components/ListTools'

const FULL_PERCENT = 100

export function Chip({ tone, children }: { tone: string; children: ReactNode }) {
  return (
    <span className={`${CHIP_CLASS} ${tone}`}>
      {children}
    </span>
  )
}

// Batang 0 sampai 100. suffix dikosongkan untuk skor kelayakan, yang memang
// berskala 0 sampai 100 tetapi bukan persentase.
export function ProgressMeter({ percent, label, suffix = '%' }: {
  percent: number
  label: string
  suffix?: string
}) {
  const safe = Math.max(0, Math.min(FULL_PERCENT, Math.round(percent)))

  return (
    <div className="flex items-center gap-3">
      <div
        role="progressbar"
        aria-label={label}
        aria-valuenow={safe}
        aria-valuemin={0}
        aria-valuemax={FULL_PERCENT}
        className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-100"
      >
        <div className="h-full rounded-full bg-navy-600" style={{ width: `${safe}%` }} />
      </div>
      <span className="w-10 text-right text-xs tabular-nums text-slate-600">
        {safe}
        {suffix}
      </span>
    </div>
  )
}

// Wadah formulir yang dibuka di bawah isi kartu, dipisah garis supaya jelas
// bagian mana yang data dan bagian mana yang isian.
export function FormSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="mt-5 border-t border-slate-100 pt-5">
      <h3 className="mb-4 text-[15px] font-semibold text-slate-900">{title}</h3>
      {children}
    </div>
  )
}
