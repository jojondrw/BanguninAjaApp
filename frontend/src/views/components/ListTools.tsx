import type { ReactNode, SelectHTMLAttributes } from 'react'

import { number } from '../../shared/format'

interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  hint?: string
  children: ReactNode
}

// Pasangan Field untuk pilihan: label terhubung lewat htmlFor, keterangan lewat
// aria-describedby, gaya sama dengan isian teks.
export function SelectField({ label, hint, id, children, ...rest }: SelectFieldProps) {
  const hintId = hint ? `${id}-hint` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-slate-700">
        {label}
      </label>
      <select
        id={id}
        aria-describedby={hintId}
        className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 transition
                   focus:border-navy-600 focus:ring-2 focus:ring-navy-100 disabled:bg-slate-100"
        {...rest}
      >
        {children}
      </select>
      {hint ? (
        <p id={hintId} className="text-xs text-slate-500">
          {hint}
        </p>
      ) : null}
    </div>
  )
}

export function Chip({ label, tone }: { label: string; tone: string }) {
  return <span className={`rounded-md px-2 py-0.5 text-xs font-medium whitespace-nowrap ${tone}`}>{label}</span>
}

const SMALL_BUTTON =
  'rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 transition hover:bg-slate-50 ' +
  'disabled:cursor-not-allowed disabled:text-slate-400 disabled:hover:bg-white'

export function Pager({ page, totalPages, totalItems, onChange }: {
  page: number
  totalPages: number
  totalItems: number
  onChange: (page: number) => void
}) {
  if (totalPages <= 1) {
    return null
  }

  return (
    <nav className="mt-4 flex flex-wrap items-center justify-between gap-2" aria-label="Halaman tabel">
      <p className="text-xs text-slate-500">
        Halaman {page} dari {totalPages}, {number(totalItems)} data
      </p>
      <div className="flex gap-2">
        <button type="button" className={SMALL_BUTTON} disabled={page <= 1} onClick={() => onChange(page - 1)}>
          Sebelumnya
        </button>
        <button
          type="button"
          className={SMALL_BUTTON}
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
        >
          Berikutnya
        </button>
      </div>
    </nav>
  )
}

export function RowAction({ label, onClick, isActive = false }: {
  label: string
  onClick: () => void
  isActive?: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={isActive}
      className={`rounded-md px-2 py-1 text-xs font-medium transition ${
        isActive ? 'bg-navy-700 text-white' : 'text-navy-700 hover:bg-navy-50'
      }`}
    >
      {label}
    </button>
  )
}

interface Choice<T extends string> {
  value: T
  label: string
}

// Pindah bagian halaman. Gayanya garis bawah supaya tidak tertukar dengan
// tombol saring di dalam tabel.
export function SectionTabs<T extends string>({ tabs, active, onChange, label }: {
  tabs: Choice<T>[]
  active: T
  onChange: (value: T) => void
  label: string
}) {
  return (
    <div className="flex gap-1 overflow-x-auto border-b border-slate-200" role="group" aria-label={label}>
      {tabs.map((tab) => (
        <button
          key={tab.value}
          type="button"
          onClick={() => onChange(tab.value)}
          aria-pressed={active === tab.value}
          className={`-mb-px border-b-2 px-3 py-2 text-sm whitespace-nowrap transition ${
            active === tab.value
              ? 'border-navy-700 font-medium text-navy-700'
              : 'border-transparent text-slate-600 hover:text-slate-900'
          }`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}

export function FilterChips<T extends string>({ filters, active, onChange, label }: {
  filters: Choice<T>[]
  active: T
  onChange: (value: T) => void
  label: string
}) {
  return (
    <div className="flex flex-wrap gap-1" role="group" aria-label={label}>
      {filters.map((filter) => (
        <button
          key={filter.label}
          type="button"
          onClick={() => onChange(filter.value)}
          aria-pressed={active === filter.value}
          className={`rounded-lg px-3 py-1.5 text-sm transition ${
            active === filter.value
              ? 'bg-navy-700 text-white'
              : 'border border-slate-300 text-slate-600 hover:bg-slate-50'
          }`}
        >
          {filter.label}
        </button>
      ))}
    </div>
  )
}
