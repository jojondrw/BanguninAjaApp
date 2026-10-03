import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link, type To } from 'react-router-dom'

import type { CashType } from '../../models/finance'
import { rupiah } from '../../shared/format'
import { BUTTON_BASE, BUTTON_PRIMARY } from '../components/Form'
import { ToolbarInput } from '../components/RecordControls'

// Hapus dua langkah di baris yang sama, bukan confirm() yang memblokir
// halaman. Fokus pindah ke Batal supaya Enter tidak langsung menghapus.
export function ConfirmDelete({ subject, isConfirming, isPending, onAsk, onCancel, onConfirm }: {
  subject: string
  isConfirming: boolean
  isPending: boolean
  onAsk: () => void
  onCancel: () => void
  onConfirm: () => void
}) {
  if (!isConfirming && !isPending) {
    return (
      <button
        type="button"
        onClick={onAsk}
        aria-label={`Hapus ${subject}`}
        className="rounded-md px-2 py-1 text-xs font-medium text-red-700 transition-[background-color,transform]
                   hover:bg-red-50 motion-safe:active:scale-95"
      >
        Hapus
      </button>
    )
  }

  return (
    <span className="inline-flex items-center gap-1.5">
      <button
        type="button"
        disabled={isPending}
        aria-busy={isPending}
        aria-label={isPending ? undefined : `Ya, hapus ${subject}`}
        onClick={onConfirm}
        className="rounded-lg bg-red-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-red-700
                   disabled:cursor-not-allowed disabled:bg-red-600/60"
      >
        {isPending ? 'Menghapus' : 'Ya, hapus'}
      </button>
      {isPending ? null : (
        <button
          type="button"
          autoFocus
          onClick={onCancel}
          className="rounded-lg px-2 py-1 text-xs text-slate-600 hover:bg-slate-100"
        >
          Batal
        </button>
      )}
    </span>
  )
}

export function RowActions({ children }: { children: ReactNode }) {
  return <span className="inline-flex items-center justify-end gap-1">{children}</span>
}

// Tautan antar tab dengan bentuk yang sama seperti RowAction. Ikonnya
// menandai bahwa tautan ini pindah ke tab lain, bukan aksi di baris itu.
export function RowLink({ to, label, icon: Icon, description }: {
  to: To
  label: string
  icon: LucideIcon
  description?: string
}) {
  return (
    <Link
      to={to}
      aria-label={description}
      title={description}
      className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium whitespace-nowrap text-navy-600
                 transition-[background-color,color,transform] hover:bg-navy-50 motion-safe:active:scale-95"
    >
      <Icon aria-hidden="true" strokeWidth={1.8} className="size-4" />
      {label}
    </Link>
  )
}

// Tanda dan warna dipakai bersamaan, supaya arah kas tetap terbaca tanpa warna.
export function SignedAmount({ type, amount }: { type: CashType; amount: number }) {
  return (
    <span className={`font-medium ${type === 'in' ? 'text-green-700' : 'text-red-700'}`}>
      {type === 'in' ? '+' : '−'}
      {rupiah(amount)}
    </span>
  )
}

export function DateRangeFilter({ idPrefix, dateFrom, dateTo, onChange }: {
  idPrefix: string
  dateFrom: string
  dateTo: string
  onChange: (range: { dateFrom: string; dateTo: string }) => void
}) {
  return (
    <>
      <ToolbarInput
        id={`${idPrefix}-date-from`}
        label="Dari"
        showLabel
        type="date"
        max={dateTo || undefined}
        value={dateFrom}
        onChange={(event) => onChange({ dateFrom: event.target.value, dateTo })}
      />
      <ToolbarInput
        id={`${idPrefix}-date-to`}
        label="Sampai"
        showLabel
        type="date"
        min={dateFrom || undefined}
        value={dateTo}
        onChange={(event) => onChange({ dateFrom, dateTo: event.target.value })}
      />
    </>
  )
}

// Button bersama tidak punya prop disabled, padahal formulir jurnal harus
// terkunci sampai seimbang. Gayanya tetap memakai konstanta yang sama.
export function SubmitButton({ children, isPending, pendingLabel, isBlocked, describedBy }: {
  children: ReactNode
  isPending: boolean
  pendingLabel: string
  isBlocked: boolean
  describedBy?: string
}) {
  return (
    <button
      type="submit"
      disabled={isPending || isBlocked}
      aria-busy={isPending}
      aria-describedby={describedBy}
      className={`h-9 ${BUTTON_BASE} ${BUTTON_PRIMARY}`}
    >
      {isPending ? (
        <span
          aria-hidden="true"
          className="size-3.5 rounded-full border-2 border-current border-r-transparent motion-safe:animate-spin"
        />
      ) : null}
      {isPending ? pendingLabel : children}
    </button>
  )
}

export function Notice({ children }: { children: ReactNode }) {
  return <div className="mb-4">{children}</div>
}
