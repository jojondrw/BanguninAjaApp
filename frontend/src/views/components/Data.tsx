import { Inbox } from 'lucide-react'
import type { ReactNode } from 'react'

import { Button, ErrorNote } from './Form'

export function Card({ title, description, action, children }: {
  title: string
  description?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="rounded-2xl bg-white shadow-hairline">
      <div className="flex items-start justify-between gap-3 px-5 pt-4 pb-3">
        <div className="min-w-0">
          <h2 className="text-[15px] font-semibold text-slate-900">{title}</h2>
          {description ? <p className="mt-0.5 text-[13px] text-slate-500">{description}</p> : null}
        </div>
        {action ? <div className="-mr-1.5 shrink-0">{action}</div> : null}
      </div>
      <div className="px-5 pb-5">{children}</div>
    </section>
  )
}

export function Kpi({ label, value, note }: { label: string; value: string; note?: string }) {
  return (
    <div className="min-w-0 bg-white px-4 py-3.5 sm:px-5 sm:py-4">
      <p className="text-[13px] font-medium text-slate-500">{label}</p>
      <p className="mt-1.5 text-[22px] leading-7 font-semibold tracking-[-0.03em] text-slate-900 tabular-nums sm:text-[26px] sm:leading-8">
        {value}
      </p>
      {note ? <p className="mt-1 text-xs text-slate-500">{note}</p> : null}
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

export function Empty({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl bg-slate-50 px-4 py-10 text-center">
      <Inbox aria-hidden="true" className="size-5 text-slate-400" strokeWidth={1.6} />
      <p className="max-w-sm text-[13px] text-slate-500">{message}</p>
    </div>
  )
}

export interface Column<T> {
  header: string
  cell: (row: T) => ReactNode
  align?: 'left' | 'right'
}

export function Table<T>({ rows, columns, emptyMessage }: {
  rows: T[]
  columns: Column<T>[]
  emptyMessage: string
}) {
  if (rows.length === 0) {
    return <Empty message={emptyMessage} />
  }

  return (
    <div className="-mx-5 overflow-x-auto">
      <table className="w-full text-[13px]">
        <thead>
          <tr className="border-y border-slate-100 bg-slate-50/70">
            {columns.map((column) => (
              <th
                key={column.header}
                scope="col"
                className={`px-3 py-2 text-xs font-medium whitespace-nowrap text-slate-500 first:pl-5 last:pr-5 ${
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
                  className={`px-3 py-3 text-slate-700 first:pl-5 last:pr-5 max-md:whitespace-nowrap ${
                    column.align === 'right' ? 'text-right whitespace-nowrap tabular-nums' : 'text-left'
                  }`}
                >
                  {column.cell(row)}
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
