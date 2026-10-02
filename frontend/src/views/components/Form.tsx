import { CircleAlert, CircleCheck } from 'lucide-react'
import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'

export const CONTROL_CLASS =
  'rounded-lg border border-slate-200 bg-white text-sm text-slate-900 shadow-control transition-[border-color,box-shadow] ' +
  'placeholder:text-slate-400 hover:border-slate-300 focus:border-navy-500 focus:ring-4 focus:ring-navy-500/15 ' +
  'focus:outline-none disabled:bg-slate-50 disabled:text-slate-500'

export const LABEL_CLASS = 'text-[13px] font-medium text-slate-700'

export const BUTTON_BASE =
  'inline-flex items-center justify-center gap-2 rounded-lg px-3.5 text-[13px] font-medium whitespace-nowrap ' +
  'transition-[background-color,border-color,color,box-shadow,transform] duration-150 select-none ' +
  'motion-safe:active:scale-[0.97] disabled:cursor-not-allowed disabled:active:scale-100'

export const BUTTON_PRIMARY = 'bg-navy-700 text-white shadow-button hover:bg-navy-800 disabled:bg-navy-700/55'

export const BUTTON_SUBTLE =
  'border border-slate-200 bg-white text-slate-700 shadow-control hover:border-slate-300 hover:bg-slate-50 ' +
  'disabled:text-slate-400 disabled:hover:bg-white'

interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  hint?: string
}

export function Field({ label, hint, id, ...rest }: FieldProps) {
  const hintId = hint ? `${id}-hint` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={LABEL_CLASS}>
        {label}
      </label>
      <input id={id} aria-describedby={hintId} className={`${CONTROL_CLASS} h-9 px-3`} {...rest} />
      {hint ? (
        <p id={hintId} className="text-xs text-slate-500">
          {hint}
        </p>
      ) : null}
    </div>
  )
}

interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  hint?: string
  children: ReactNode
}

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
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={isPending}
      aria-busy={isPending}
      className={`h-9 ${BUTTON_BASE} ${variant === 'primary' ? BUTTON_PRIMARY : BUTTON_SUBTLE}`}
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

export function ErrorNote({ message }: { message: string }) {
  return (
    <p role="alert" className="flex items-start gap-2 rounded-xl bg-red-50 px-3.5 py-2.5 text-[13px] text-red-700">
      <CircleAlert aria-hidden="true" className="mt-px size-4 shrink-0" strokeWidth={1.9} />
      {message}
    </p>
  )
}

export function SuccessNote({ message }: { message: string }) {
  return (
    <p role="status" className="flex items-start gap-2 rounded-xl bg-green-50 px-3.5 py-2.5 text-[13px] text-green-700">
      <CircleCheck aria-hidden="true" className="mt-px size-4 shrink-0" strokeWidth={1.9} />
      {message}
    </p>
  )
}
