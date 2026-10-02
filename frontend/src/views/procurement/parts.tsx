import type { ReactNode, TextareaHTMLAttributes } from 'react'

import { BUTTON_BASE, BUTTON_PRIMARY, BUTTON_SUBTLE, CONTROL_CLASS, LABEL_CLASS } from '../components/Form'

type Tone = 'neutral' | 'danger'
type Size = 'sm' | 'md'

const SMALL_BASE =
  'inline-flex h-7 items-center justify-center gap-1.5 rounded-lg px-2.5 text-xs font-medium whitespace-nowrap ' +
  'transition-[background-color,color,transform] motion-safe:active:scale-[0.97] disabled:cursor-not-allowed ' +
  'disabled:active:scale-100'

const ASK_CLASS: Record<Size, Record<Tone, string>> = {
  sm: {
    neutral: `${SMALL_BASE} bg-white text-slate-700 shadow-hairline hover:bg-slate-50 disabled:text-slate-400`,
    danger: `${SMALL_BASE} bg-white text-red-700 shadow-hairline hover:bg-red-50 disabled:text-red-300`,
  },
  md: {
    neutral: `h-9 ${BUTTON_BASE} ${BUTTON_SUBTLE}`,
    danger:
      `h-9 ${BUTTON_BASE} border border-red-200 bg-white text-red-700 shadow-control hover:border-red-300 ` +
      'hover:bg-red-50',
  },
}

const CONFIRM_TONE: Record<Tone, string> = {
  neutral: BUTTON_PRIMARY,
  danger: 'bg-red-600 text-white shadow-button hover:bg-red-700 disabled:bg-red-600/55',
}

const CONFIRM_SIZE: Record<Size, string> = {
  sm: `${SMALL_BASE} px-3`,
  md: `h-9 ${BUTTON_BASE}`,
}

const DISMISS_CLASS: Record<Size, string> = {
  sm: 'rounded-lg px-2 py-1 text-xs text-slate-600 hover:bg-slate-100',
  md: `h-9 ${BUTTON_BASE} ${BUTTON_SUBTLE}`,
}

function Spinner() {
  return (
    <span
      aria-hidden="true"
      className="size-3 rounded-full border-2 border-current border-r-transparent motion-safe:animate-spin"
    />
  )
}

// Tombol kecil untuk aksi di baris tabel yang tidak perlu konfirmasi.
export function SmallButton({ children, onClick, isPending = false, pendingLabel = 'Memproses', tone = 'neutral' }: {
  children: ReactNode
  onClick: () => void
  isPending?: boolean
  pendingLabel?: string
  tone?: Tone
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isPending}
      aria-busy={isPending}
      className={ASK_CLASS.sm[tone]}
    >
      {isPending ? <Spinner /> : null}
      {isPending ? pendingLabel : children}
    </button>
  )
}

// Konfirmasi dua langkah di tempat, pengganti confirm(). Langkah kedua
// menyebut akibatnya ("Ya, hapus"), dan tombol Batal hilang selama permintaan
// berjalan supaya tidak terkesan bisa dibatalkan di tengah jalan.
export function ConfirmAction({
  label,
  prompt,
  confirmLabel,
  pendingLabel,
  isAsking,
  isPending,
  onAsk,
  onCancel,
  onConfirm,
  tone = 'neutral',
  size = 'sm',
}: {
  label: string
  prompt?: string
  confirmLabel: string
  pendingLabel: string
  isAsking: boolean
  isPending: boolean
  onAsk: () => void
  onCancel: () => void
  onConfirm: () => void
  tone?: Tone
  size?: Size
}) {
  if (!isAsking && !isPending) {
    return (
      <button type="button" onClick={onAsk} className={ASK_CLASS[size][tone]}>
        {label}
      </button>
    )
  }

  return (
    <span className="inline-flex flex-wrap items-center gap-2">
      {prompt ? <span className={size === 'sm' ? 'text-xs text-slate-600' : 'text-[13px] text-slate-700'}>{prompt}</span> : null}
      <button
        type="button"
        onClick={onConfirm}
        disabled={isPending}
        aria-busy={isPending}
        className={`${CONFIRM_SIZE[size]} ${CONFIRM_TONE[tone]}`}
      >
        {isPending ? <Spinner /> : null}
        {isPending ? pendingLabel : confirmLabel}
      </button>
      {isPending ? null : (
        <button type="button" onClick={onCancel} className={DISMISS_CLASS[size]}>
          Batal
        </button>
      )}
    </span>
  )
}

interface TextAreaFieldProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  id: string
  label: string
  hint?: string
}

// Pasangan Field untuk catatan panjang, dengan label dan keterangan yang sama.
export function TextAreaField({ id, label, hint, ...rest }: TextAreaFieldProps) {
  const hintId = hint ? `${id}-hint` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={LABEL_CLASS}>
        {label}
      </label>
      <textarea id={id} aria-describedby={hintId} rows={2} className={`${CONTROL_CLASS} px-3 py-2`} {...rest} />
      {hint ? (
        <p id={hintId} className="text-xs text-slate-500">
          {hint}
        </p>
      ) : null}
    </div>
  )
}

// Ringkasan kecil "label: nilai" di bawah judul rincian.
export function Facts({ items }: { items: { label: string; value: ReactNode }[] }) {
  return (
    <dl className="mb-4 grid gap-x-6 gap-y-2 text-[13px] sm:grid-cols-2 lg:grid-cols-4">
      {items.map((item) => (
        <div key={item.label} className="min-w-0">
          <dt className="text-xs text-slate-500">{item.label}</dt>
          <dd className="mt-0.5 text-slate-900">{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}
