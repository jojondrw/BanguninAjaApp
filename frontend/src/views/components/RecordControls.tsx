import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'

import { number } from '../../shared/format'

const CONTROL_STYLE =
  'rounded-lg border border-slate-300 bg-white text-sm text-slate-900 transition ' +
  'focus:border-navy-600 focus:ring-2 focus:ring-navy-100 disabled:bg-slate-100 disabled:text-slate-500'

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
      <label htmlFor={id} className="text-sm font-medium text-slate-700">
        {label}
      </label>
      <select id={id} aria-describedby={hintId} className={`${CONTROL_STYLE} px-3 py-2`} {...rest}>
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
      <select id={id} className={`${CONTROL_STYLE} px-3 py-1.5`} {...rest}>
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
  return (
    <div className="flex items-center gap-2">
      <label htmlFor={id} className={showLabel ? 'text-sm text-slate-600' : 'sr-only'}>
        {label}
      </label>
      <input
        id={id}
        autoComplete="off"
        {...rest}
        className={`${CONTROL_STYLE} px-3 py-1.5 placeholder:text-slate-400 ${className}`}
      />
    </div>
  )
}

export function Toolbar({ children }: { children: ReactNode }) {
  return <div className="mb-4 flex flex-wrap items-center gap-2">{children}</div>
}

export function Chip({ tone, children }: { tone: string; children: ReactNode }) {
  return <span className={`rounded px-2 py-0.5 text-xs font-medium whitespace-nowrap ${tone}`}>{children}</span>
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

  const pageButton =
    'rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 transition ' +
    'hover:bg-slate-50 disabled:cursor-not-allowed disabled:text-slate-400 disabled:hover:bg-white'

  return (
    <nav aria-label="Halaman tabel" className="mt-4 flex flex-wrap items-center justify-between gap-2">
      <p className="text-xs text-slate-500">
        {number(totalItems)} {unit}
        {totalPages > 1 ? ` · halaman ${page} dari ${totalPages}` : ''}
      </p>
      {totalPages > 1 ? (
        <div className="flex gap-2">
          <button type="button" className={pageButton} disabled={page <= 1} onClick={() => onChange(page - 1)}>
            Sebelumnya
          </button>
          <button
            type="button"
            className={pageButton}
            disabled={page >= totalPages}
            onClick={() => onChange(page + 1)}
          >
            Berikutnya
          </button>
        </div>
      ) : null}
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
      className={`ml-auto rounded-lg px-3 py-1.5 text-sm font-medium transition ${
        isOpen
          ? 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50'
          : 'bg-navy-700 text-white hover:bg-navy-900'
      }`}
    >
      {isOpen ? 'Tutup formulir' : openLabel}
    </button>
  )
}

export function FormPanel({ children }: { children: ReactNode }) {
  return <div className="mb-5 rounded-lg border border-slate-200 bg-slate-50 p-4">{children}</div>
}
