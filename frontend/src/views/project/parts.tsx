import { type ReactNode, useState } from 'react'

import { BUTTON_BASE, BUTTON_PRIMARY, BUTTON_SUBTLE } from '../components/Form'
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

const DANGER_ROW_ACTION =
  'rounded-md px-2 py-1 text-xs font-medium text-red-700 transition-[background-color,transform] hover:bg-red-50 ' +
  'motion-safe:active:scale-95'

const DANGER_CONFIRM =
  'rounded-lg bg-red-600 px-2.5 py-1 text-xs font-medium whitespace-nowrap text-white hover:bg-red-700 ' +
  'disabled:cursor-not-allowed disabled:bg-red-600/60'

// Hapus dua langkah di baris yang sama, bukan lewat confirm() yang memblokir
// halaman. Fokus pindah ke Batal supaya Enter yang tertekan dua kali tidak
// langsung menghapus.
export function InlineConfirm({ label, confirmLabel, pendingLabel, isPending, onConfirm }: {
  label: string
  confirmLabel: string
  pendingLabel: string
  isPending: boolean
  onConfirm: () => void
}) {
  const [isAsking, setIsAsking] = useState(false)

  if (!isAsking && !isPending) {
    return (
      <button type="button" onClick={() => setIsAsking(true)} className={DANGER_ROW_ACTION}>
        {label}
      </button>
    )
  }

  return (
    <span className="inline-flex items-center gap-1.5">
      <button type="button" disabled={isPending} aria-busy={isPending} onClick={onConfirm} className={DANGER_CONFIRM}>
        {isPending ? pendingLabel : confirmLabel}
      </button>
      {isPending ? null : (
        <button
          type="button"
          autoFocus
          onClick={() => setIsAsking(false)}
          className="rounded-lg px-2 py-1 text-xs text-slate-600 hover:bg-slate-100"
        >
          Batal
        </button>
      )}
    </span>
  )
}

// Tombol pembuka panel di kepala halaman. Gayanya berganti saat panel terbuka,
// sama seperti FormToggle.
export function PanelToggle({ label, panelId, isOpen, onToggle }: {
  label: string
  panelId: string
  isOpen: boolean
  onToggle: () => void
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-expanded={isOpen}
      aria-controls={isOpen ? panelId : undefined}
      className={`h-9 ${BUTTON_BASE} ${isOpen ? BUTTON_PRIMARY : BUTTON_SUBTLE}`}
    >
      {label}
    </button>
  )
}

// Tombol kirim yang bisa dikunci dengan alasan, misalnya status Selesai saat
// masih ada tahap di bawah 100%. Button bawaan hanya terkunci selama proses.
export function GuardedButton({ children, isPending, pendingLabel, disabled, describedBy, onClick }: {
  children: ReactNode
  isPending: boolean
  pendingLabel: string
  disabled?: boolean
  describedBy?: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isPending || disabled}
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
