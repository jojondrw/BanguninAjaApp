import type { FormEvent, ReactNode } from 'react'

import { ErrorNote, SuccessNote } from '../components/Form'

// Tombol aksi baris untuk SDM dan Inventaris. Bentuknya sama dengan RowAction
// dan "Tandai dibayar", hanya ditambah keadaan sedang proses.

const TEXT_ACTION_TONE = {
  accent: 'text-navy-600 hover:bg-navy-50',
  muted: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
}

const CONFIRM_TONE = {
  danger: 'bg-red-600 hover:bg-red-700 disabled:bg-red-600/60',
  primary: 'bg-navy-700 hover:bg-navy-900 disabled:bg-navy-700/60',
}

export function TextAction({ label, srLabel, onClick, tone = 'accent', isPending = false, pendingLabel }: {
  label: string
  srLabel?: string
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
      {srLabel ? <span className="sr-only">: {srLabel}</span> : null}
    </button>
  )
}

export function ActionGroup({ children }: { children: ReactNode }) {
  return <div className="inline-flex flex-wrap items-center justify-end gap-1">{children}</div>
}

interface ConfirmActionProps {
  label: string
  srLabel?: string
  question?: string
  confirmLabel: string
  pendingLabel: string
  tone?: keyof typeof CONFIRM_TONE
  isAsking: boolean
  isPending: boolean
  onAsk: () => void
  onCancel: () => void
  onConfirm: () => void
  children?: ReactNode
}

// Konfirmasi dua langkah di baris yang sama, bukan confirm() yang memblokir
// halaman. Isian tambahan (misalnya tanggal keluar) masuk lewat children dan
// ikut terkirim saat Enter ditekan. Tanpa isian, fokus pindah ke Batal supaya
// Enter yang tertekan dua kali tidak langsung menjalankan tindakannya.
export function ConfirmAction({
  label,
  srLabel,
  question,
  confirmLabel,
  pendingLabel,
  tone = 'danger',
  isAsking,
  isPending,
  onAsk,
  onCancel,
  onConfirm,
  children,
}: ConfirmActionProps) {
  if (!isAsking && !isPending) {
    return <TextAction label={label} srLabel={srLabel} tone={tone === 'danger' ? 'muted' : 'accent'} onClick={onAsk} />
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
          Batal
        </button>
      )}
    </form>
  )
}

// Catatan hasil tindakan baris, ditaruh di atas tabel supaya terlihat tanpa
// menggulir. Pemanggil menjaga hanya satu catatan yang aktif.
export function RowNotice({ success, error }: { success?: string | null; error?: string | null }) {
  if (!success && !error) {
    return null
  }

  return <div className="mb-4">{error ? <ErrorNote message={error} /> : <SuccessNote message={success ?? ''} />}</div>
}
