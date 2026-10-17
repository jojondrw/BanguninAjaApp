import { Inbox } from 'lucide-react'
import type { ReactNode } from 'react'

import type { StockPhoto } from '../../shared/photos'
import { Button, ErrorNote } from './Form'
import { MoneyText } from './Money'

export function Card({ title, description, action, children }: {
  title: string
  description?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="min-w-0 rounded-2xl bg-white shadow-hairline">
      <div className="flex flex-wrap items-start justify-between gap-x-3 gap-y-2 px-5 pt-4 pb-3 sm:px-6 sm:pt-5 sm:pb-4">
        <div className="min-w-0 flex-1 basis-44">
          <h2 className="text-[15px] font-semibold text-slate-900">{title}</h2>
          {description ? <p className="mt-0.5 text-[13px] text-slate-500"><MoneyText>{description}</MoneyText></p> : null}
        </div>
        {action ? <div className="-mr-1.5 shrink-0">{action}</div> : null}
      </div>
      <div className="px-5 pb-5 sm:px-6 sm:pb-6">{children}</div>
    </section>
  )
}

export function Kpi({ label, value, note }: { label: string; value: string; note?: string }) {
  return (
    <div className="min-w-0 bg-white px-4 py-3.5 sm:px-6 sm:py-5">
      <p className="text-[13px] font-medium text-slate-500">{label}</p>
      <p className="mt-1.5 text-[22px] leading-7 font-semibold tracking-[-0.03em] text-slate-900 tabular-nums sm:text-[26px] sm:leading-8">
        <MoneyText>{value}</MoneyText>
      </p>
      {note ? <p className="mt-1 text-xs text-slate-500"><MoneyText>{note}</MoneyText></p> : null}
    </div>
  )
}

// Celah 1px di atas latar abu menjadi garis pemisah, jadi pemisahnya tetap rapi
// di semua jumlah kolom.
export function KpiRow({ children }: { children: ReactNode }) {
  return (
    <div className="grid gap-px overflow-hidden rounded-2xl bg-slate-200/70 shadow-hairline grid-cols-2 xl:grid-cols-4">
      {children}
    </div>
  )
}

export function Loading({ label = 'Mengambil data...' }: { label?: string }) {
  return (
    <div role="status" className="flex flex-col gap-2.5 py-4">
      <span className="sr-only">{label}</span>
      {[92, 76, 84].map((width) => (
        <span
          key={width}
          aria-hidden="true"
          className="h-3 rounded-full bg-slate-100 motion-safe:animate-pulse"
          style={{ width: `${width}%` }}
        />
      ))}
    </div>
  )
}

export function LoadFailed({ onRetry }: { onRetry: () => void }) {
  return (
    <div>
      <ErrorNote message="Gagal mengambil data dari server." />
      <div className="mt-3">
        <Button variant="subtle" onClick={onRetry}>
          Coba lagi
        </Button>
      </div>
    </div>
  )
}

export function Empty({ message, photo }: { message: string; photo?: StockPhoto }) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl bg-slate-50 px-4 py-10 text-center">
      {photo ? (
        <img
          src={photo.src}
          alt=""
          aria-hidden="true"
          loading="lazy"
          decoding="async"
          className="mb-2 h-28 w-44 rounded-xl object-cover shadow-hairline"
        />
      ) : (
        <Inbox aria-hidden="true" className="size-5 text-slate-400" strokeWidth={1.6} />
      )}
      <p className="max-w-sm text-[13px] text-slate-500">{message}</p>
    </div>
  )
}

export interface Column<T> {
  header: string
  cell: (row: T) => ReactNode
  align?: 'left' | 'right'
}

export function Table<T>({ rows, columns, emptyMessage, emptyPhoto }: {
  rows: T[]
  columns: Column<T>[]
  emptyMessage: string
  emptyPhoto?: StockPhoto
}) {
  if (rows.length === 0) {
    return <Empty message={emptyMessage} photo={emptyPhoto} />
  }

  return (
    <div className="relative -mx-5 overflow-x-auto sm:-mx-6">
      <table className="w-full text-[13px]">
        <thead>
          <tr className="border-y border-slate-100 bg-slate-50/70">
            {columns.map((column) => (
              <th
                key={column.header}
                scope="col"
                className={`px-3 py-2 text-xs font-medium whitespace-nowrap text-slate-500 first:pl-5 last:pr-5 sm:first:pl-6 sm:last:pr-6 ${
                  column.align === 'right' ? 'text-right' : 'text-left'
                }`}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, index) => (
            <tr key={index} className="border-b border-slate-100 transition-colors last:border-0 hover:bg-slate-50/80">
              {columns.map((column) => (
                <td
                  key={column.header}
                  className={`px-3 py-3.5 text-slate-700 first:pl-5 last:pr-5 max-md:whitespace-nowrap sm:first:pl-6 sm:last:pr-6 ${
                    column.align === 'right' ? 'text-right whitespace-nowrap tabular-nums' : 'text-left'
                  }`}
                >
                  <MoneyText>{column.cell(row)}</MoneyText>
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Bar({ percent }: { percent: number }) {
  const safe = Math.max(0, Math.min(100, percent))

  return (
    <span className="flex items-center justify-end gap-2">
      <span className="h-1.5 w-20 overflow-hidden rounded-full bg-slate-100">
        <span className="block h-full rounded-full bg-navy-600" style={{ width: `${safe}%` }} />
      </span>
      <span className="w-9 text-right text-xs tabular-nums text-slate-600">{safe}%</span>
    </span>
  )
}
