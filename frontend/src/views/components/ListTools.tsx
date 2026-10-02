import { ChevronLeft, ChevronRight } from 'lucide-react'
import { type ReactNode, type SelectHTMLAttributes, useRef } from 'react'

import { number } from '../../shared/format'
import { CONTROL_CLASS, LABEL_CLASS } from './Form'
import { SegmentIndicator } from './Segmented'
import { SEGMENT_LIST, segmentButton, useSegmentIndicator } from './segmentIndicator'

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
      <label htmlFor={id} className={LABEL_CLASS}>
        {label}
      </label>
      <select id={id} aria-describedby={hintId} className={`${CONTROL_CLASS} h-9 px-2.5`} {...rest}>
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

export const CHIP_CLASS =
  'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium whitespace-nowrap ' +
  'before:size-1.5 before:rounded-full before:bg-current before:opacity-75'

export function Chip({ label, tone }: { label: string; tone: string }) {
  return <span className={`${CHIP_CLASS} ${tone}`}>{label}</span>
}

export const PAGE_BUTTON =
  'inline-flex h-8 items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 text-[13px] font-medium ' +
  'text-slate-700 shadow-control transition-[background-color,transform] hover:bg-slate-50 motion-safe:active:scale-[0.97] ' +
  'disabled:cursor-not-allowed disabled:text-slate-300 disabled:shadow-none disabled:hover:bg-white disabled:active:scale-100'

export function PageButtons({ page, totalPages, onChange }: {
  page: number
  totalPages: number
  onChange: (page: number) => void
}) {
  return (
    <div className="flex gap-1.5">
      <button type="button" className={PAGE_BUTTON} disabled={page <= 1} onClick={() => onChange(page - 1)}>
        <ChevronLeft aria-hidden="true" className="size-4" strokeWidth={1.8} />
        Sebelumnya
      </button>
      <button type="button" className={PAGE_BUTTON} disabled={page >= totalPages} onClick={() => onChange(page + 1)}>
        Berikutnya
        <ChevronRight aria-hidden="true" className="size-4" strokeWidth={1.8} />
      </button>
    </div>
  )
}

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
      <p className="text-xs text-slate-500 tabular-nums">
        Halaman {page} dari {totalPages} · {number(totalItems)} data
      </p>
      <PageButtons page={page} totalPages={totalPages} onChange={onChange} />
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
      className={`rounded-md px-2 py-1 text-xs font-medium transition-[background-color,color,transform]
                  motion-safe:active:scale-95 ${
                    isActive ? 'bg-navy-700 text-white' : 'text-navy-600 hover:bg-navy-50'
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

// Pindah bagian halaman. Bentuknya segmented control supaya tidak tertukar
// dengan tombol saring di dalam tabel.
export function SectionTabs<T extends string>({ tabs, active, onChange, label }: {
  tabs: Choice<T>[]
  active: T
  onChange: (value: T) => void
  label: string
}) {
  const listRef = useRef<HTMLDivElement>(null)
  const indicator = useSegmentIndicator(listRef, active)

  return (
    <div ref={listRef} className={SEGMENT_LIST} role="group" aria-label={label}>
      <SegmentIndicator indicator={indicator} />
      {tabs.map((tab) => (
        <button
          key={tab.value}
          data-segment={tab.value}
          type="button"
          onClick={() => onChange(tab.value)}
          aria-pressed={active === tab.value}
          className={segmentButton(active === tab.value, indicator !== null)}
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
    <div className="flex flex-wrap gap-1.5" role="group" aria-label={label}>
      {filters.map((filter) => (
        <button
          key={filter.label}
          type="button"
          onClick={() => onChange(filter.value)}
          aria-pressed={active === filter.value}
          className={`h-8 rounded-full px-3 text-[13px] font-medium transition-[background-color,color,transform]
                      motion-safe:active:scale-[0.97] ${
                        active === filter.value
                          ? 'bg-slate-900 text-white'
                          : 'bg-white text-slate-600 shadow-hairline hover:bg-slate-50 hover:text-slate-900'
                      }`}
        >
          {filter.label}
        </button>
      ))}
    </div>
  )
}
