import type { InputHTMLAttributes, ReactNode } from 'react'

interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  hint?: string
}

export function Field({ label, hint, id, ...rest }: FieldProps) {
  const hintId = hint ? `${id}-hint` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-slate-700">
        {label}
      </label>
      <input
        id={id}
        aria-describedby={hintId}
        className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 transition
                   placeholder:text-slate-400 focus:border-navy-600 focus:ring-2 focus:ring-navy-100
                   disabled:bg-slate-100"
        {...rest}
      />
      {hint ? (
        <p id={hintId} className="text-xs text-slate-500">
          {hint}
        </p>
      ) : null}
    </div>
  )
}

interface ButtonProps {
  children: ReactNode
  type?: 'button' | 'submit'
  isPending?: boolean
  pendingLabel?: string
  onClick?: () => void
  variant?: 'primary' | 'subtle'
}

export function Button({
  children,
  type = 'button',
  isPending = false,
  pendingLabel = 'Memproses',
  onClick,
  variant = 'primary',
}: ButtonProps) {
  const style =
    variant === 'primary'
      ? 'bg-navy-700 text-white hover:bg-navy-900 disabled:bg-navy-700/60'
      : 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 disabled:text-slate-400'

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={isPending}
      aria-busy={isPending}
      className={`inline-flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium
                  transition disabled:cursor-not-allowed ${style}`}
    >
      {isPending ? pendingLabel : children}
    </button>
  )
}

export function ErrorNote({ message }: { message: string }) {
  return (
    <p role="alert" className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
      {message}
    </p>
  )
}

export function SuccessNote({ message }: { message: string }) {
  return (
    <p role="status" className="rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-700">
      {message}
    </p>
  )
}
