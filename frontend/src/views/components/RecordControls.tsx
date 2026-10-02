import { Plus, Search, X } from 'lucide-react'
import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'

import { number } from '../../shared/format'
import { BUTTON_BASE, BUTTON_PRIMARY, BUTTON_SUBTLE, CONTROL_CLASS, LABEL_CLASS } from './Form'
import { CHIP_CLASS, PageButtons } from './ListTools'

interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  id: string
  label: string
  hint?: string
  children: ReactNode
}

// Pasangan Field untuk pilihan: label terhubung lewat htmlFor, keterangan lewat
// aria-describedby, dengan gaya yang sama.
export function SelectField({ id, label, hint, children, ...rest }: SelectFieldProps) {
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

interface FilterSelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  id: string
  label: string
  children: ReactNode
}

// Versi ringkas untuk baris penyaring di atas tabel. Labelnya tetap ada untuk
// pembaca layar.
export function FilterSelect({ id, label, children, ...rest }: FilterSelectProps) {
  return (
    <div>
      <label htmlFor={id} className="sr-only">
        {label}
      </label>
      <select id={id} className={`${CONTROL_CLASS} h-8 px-2.5 text-[13px]`} {...rest}>
        {children}
      </select>
    </div>
  )
}

interface ToolbarInputProps extends InputHTMLAttributes<HTMLInputElement> {
  id: string
  label: string
  showLabel?: boolean
}

// Isian ringkas untuk baris penyaring: pencarian, tanggal, atau periode.
export function ToolbarInput({ id, label, showLabel = false, className = '', ...rest }: ToolbarInputProps) {
  const isSearch = rest.type === 'search'

  return (
    <div className="flex items-center gap-2">
      <label htmlFor={id} className={showLabel ? 'text-[13px] text-slate-600' : 'sr-only'}>
        {label}
      </label>
      <div className="relative">
        {isSearch ? (
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-slate-400"
            strokeWidth={2}
          />
        ) : null}
        <input
          id={id}
          autoComplete="off"
          {...rest}
          className={`${CONTROL_CLASS} h-8 text-[13px] ${isSearch ? 'pr-3 pl-8' : 'px-2.5'} ${className}`}
        />
      </div>
    </div>
  )
}

export function Toolbar({ children }: { children: ReactNode }) {
  return <div className="mb-4 flex flex-wrap items-center gap-2">{children}</div>
}

export function Chip({ tone, children }: { tone: string; children: ReactNode }) {
  return <span className={`${CHIP_CLASS} ${tone}`}>{children}</span>
}

interface PagerProps {
  page: number
  totalPages: number
  totalItems: number
  unit: string
  onChange: (page: number) => void
}

export function Pager({ page, totalPages, totalItems, unit, onChange }: PagerProps) {
  if (totalItems === 0) {
    return null
  }

  return (
    <nav aria-label="Halaman tabel" className="mt-4 flex flex-wrap items-center justify-between gap-2">
      <p className="text-xs text-slate-500 tabular-nums">
        {number(totalItems)} {unit}
        {totalPages > 1 ? ` · halaman ${page} dari ${totalPages}` : ''}
      </p>
      {totalPages > 1 ? <PageButtons page={page} totalPages={totalPages} onChange={onChange} /> : null}
    </nav>
  )
}

// Tombol pembuka formulir di dalam kartu. Teksnya berganti supaya jelas apa
// akibat menekannya.
export function FormToggle({ isOpen, openLabel, onToggle }: {
  isOpen: boolean
  openLabel: string
  onToggle: () => void
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-expanded={isOpen}
      className={`ml-auto h-8 ${BUTTON_BASE} ${isOpen ? BUTTON_SUBTLE : BUTTON_PRIMARY}`}
    >
      {isOpen ? (
        <X aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
      ) : (
        <Plus aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
      )}
      {isOpen ? 'Tutup formulir' : openLabel}
    </button>
  )
}

export function FormPanel({ children }: { children: ReactNode }) {
  return <div className="mb-5 rounded-xl bg-slate-50 p-4 shadow-[inset_0_0_0_1px_rgb(0_0_0/0.04)] motion-safe:animate-enter">{children}</div>
}
