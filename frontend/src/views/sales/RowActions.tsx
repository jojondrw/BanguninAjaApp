import type { FormEvent, ReactNode } from 'react'

import { BUTTON_BASE, BUTTON_PRIMARY, BUTTON_SUBTLE } from '../components/Form'

const TEXT_ACTION_TONE = {
  accent: 'text-navy-600 hover:bg-navy-50',
  muted: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
}

const CONFIRM_TONE = {
  danger: 'bg-red-600 hover:bg-red-700 disabled:bg-red-600/60',
  primary: 'bg-navy-700 hover:bg-navy-900 disabled:bg-navy-700/60',
}

// Tombol teks kecil untuk kolom Aksi. Sama bentuknya dengan RowAction, tapi
// bukan tombol tekan-tahan, jadi tanpa aria-pressed.
export function TextAction({ label, onClick, tone = 'accent', isPending = false, pendingLabel }: {
  label: string
  onClick: () => void
  tone?: keyof typeof TEXT_ACTION_TONE
  isPending?: boolean
  pendingLabel?: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isPending}
      aria-busy={isPending}
      className={`rounded-md px-2 py-1 text-xs font-medium whitespace-nowrap transition-[background-color,color,transform]
                  motion-safe:active:scale-95 disabled:cursor-not-allowed disabled:opacity-60 ${TEXT_ACTION_TONE[tone]}`}
    >
      {isPending && pendingLabel ? pendingLabel : label}
    </button>
  )
}

export function ActionGroup({ children }: { children: ReactNode }) {
  return <div className="inline-flex flex-wrap items-center justify-end gap-1">{children}</div>
}

interface ConfirmActionProps {
  label: string
  question?: string
  confirmLabel: string
  pendingLabel: string
  cancelLabel?: string
  tone?: keyof typeof CONFIRM_TONE
  isAsking: boolean
  isPending: boolean
  onAsk: () => void
  onCancel: () => void
  onConfirm: () => void
  children?: ReactNode
}

// Konfirmasi dua langkah di tempat, mengikuti pola "Tandai dibayar" di halaman
// SDM. Isian tambahan (misalnya tanggal bayar) masuk lewat children dan ikut
// terkirim saat Enter ditekan.
export function ConfirmAction({
  label,
  question,
  confirmLabel,
  pendingLabel,
  cancelLabel = 'Batal',
  tone = 'danger',
  isAsking,
  isPending,
  onAsk,
  onCancel,
  onConfirm,
  children,
}: ConfirmActionProps) {
  if (!isAsking && !isPending) {
    return <TextAction label={label} tone={tone === 'danger' ? 'muted' : 'accent'} onClick={onAsk} />
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    onConfirm()
  }

  return (
    <form onSubmit={submit} className="inline-flex flex-wrap items-center justify-end gap-2">
      {question ? <span className="text-xs text-slate-700">{question}</span> : null}
      {children}
      <button
        type="submit"
        disabled={isPending}
        aria-busy={isPending}
        className={`rounded-lg px-3 py-1 text-xs font-medium whitespace-nowrap text-white disabled:cursor-not-allowed ${CONFIRM_TONE[tone]}`}
      >
        {isPending ? pendingLabel : confirmLabel}
      </button>
      {isPending ? null : (
        <button
          type="button"
          autoFocus={children === undefined}
          onClick={onCancel}
          className="rounded-lg px-2 py-1 text-xs whitespace-nowrap text-slate-600 hover:bg-slate-100"
        >
          {cancelLabel}
        </button>
      )}
    </form>
  )
}

const PANEL_CONFIRM_TONE = {
  danger: 'bg-red-600 text-white shadow-button hover:bg-red-700 disabled:bg-red-600/55',
  primary: BUTTON_PRIMARY,
}

// Versi kartu dari ConfirmAction: kalimat akibatnya ditulis lengkap, tombolnya
// berukuran normal. Fokus pindah ke tombol batal supaya Enter tidak langsung
// menjalankan tindakan yang tidak bisa diulang.
export function ConfirmPanel({
  question,
  confirmLabel,
  pendingLabel,
  cancelLabel = 'Tidak jadi',
  tone,
  isPending,
  onConfirm,
  onCancel,
}: {
  question: string
  confirmLabel: string
  pendingLabel: string
  cancelLabel?: string
  tone: keyof typeof PANEL_CONFIRM_TONE
  isPending: boolean
  onConfirm: () => void
  onCancel: () => void
}) {
  return (
    <div className="flex flex-wrap items-center gap-3 rounded-xl bg-slate-50 px-3.5 py-3 shadow-[inset_0_0_0_1px_rgb(0_0_0/0.04)]">
      <p className="min-w-0 flex-1 basis-64 text-[13px] text-slate-700">{question}</p>
      <div className="flex gap-2">
        <button
          type="button"
          onClick={onConfirm}
          disabled={isPending}
          aria-busy={isPending}
          className={`h-9 ${BUTTON_BASE} ${PANEL_CONFIRM_TONE[tone]}`}
        >
          {isPending ? (
            <span
              aria-hidden="true"
              className="size-3.5 rounded-full border-2 border-current border-r-transparent motion-safe:animate-spin"
            />
          ) : null}
          {isPending ? pendingLabel : confirmLabel}
        </button>
        {isPending ? null : (
          <button type="button" autoFocus onClick={onCancel} className={`h-9 ${BUTTON_BASE} ${BUTTON_SUBTLE}`}>
            {cancelLabel}
          </button>
        )}
      </div>
    </div>
  )
}
