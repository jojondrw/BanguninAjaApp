import type { InputHTMLAttributes, ReactNode } from 'react'

interface KolomProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  bantuan?: string
}

export function Kolom({ label, bantuan, id, ...sisa }: KolomProps) {
  const bantuanId = bantuan ? `${id}-bantuan` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-slate-700">
        {label}
      </label>
      <input
        id={id}
        aria-describedby={bantuanId}
        className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 transition
                   placeholder:text-slate-400 focus:border-navy-600 focus:ring-2 focus:ring-navy-100
                   disabled:bg-slate-100"
        {...sisa}
      />
      {bantuan ? (
        <p id={bantuanId} className="text-xs text-slate-500">
          {bantuan}
        </p>
      ) : null}
    </div>
  )
}

interface TombolProps {
  children: ReactNode
  type?: 'button' | 'submit'
  sedangProses?: boolean
  labelProses?: string
  onClick?: () => void
  varian?: 'utama' | 'halus'
}

export function Tombol({
  children,
  type = 'button',
  sedangProses = false,
  labelProses = 'Memproses',
  onClick,
  varian = 'utama',
}: TombolProps) {
  const gaya =
    varian === 'utama'
      ? 'bg-navy-700 text-white hover:bg-navy-900 disabled:bg-navy-700/60'
      : 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 disabled:text-slate-400'

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={sedangProses}
      aria-busy={sedangProses}
      className={`inline-flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium
                  transition disabled:cursor-not-allowed ${gaya}`}
    >
      {sedangProses ? labelProses : children}
    </button>
  )
}

export function Peringatan({ pesan }: { pesan: string }) {
  return (
    <p role="alert" className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
      {pesan}
    </p>
  )
}

export function Berhasil({ pesan }: { pesan: string }) {
  return (
    <p role="status" className="rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-700">
      {pesan}
    </p>
  )
}
