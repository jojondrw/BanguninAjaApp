import type { ReactNode } from 'react'

import { Button, ErrorNote } from './Form'

export function Card({ title, description, children }: {
  title: string
  description?: string
  children: ReactNode
}) {
  return (
    <section className="rounded-xl border border-slate-200 bg-white">
      <div className="border-b border-slate-200 px-5 py-4">
        <h2 className="text-sm font-semibold text-slate-900">{title}</h2>
        {description ? <p className="mt-0.5 text-xs text-slate-500">{description}</p> : null}
      </div>
      <div className="p-5">{children}</div>
    </section>
  )
}

export function Kpi({ label, value, note }: { label: string; value: string; note?: string }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <p className="text-xs text-slate-500">{label}</p>
      <p className="mt-2 text-2xl font-semibold tracking-tight text-slate-900 tabular-nums">{value}</p>
      {note ? <p className="mt-1 text-xs text-slate-500">{note}</p> : null}
    </div>
  )
}

export function KpiRow({ children }: { children: ReactNode }) {
  return <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">{children}</div>
}

export function Loading({ label = 'Mengambil data...' }: { label?: string }) {
  return <p className="py-6 text-sm text-slate-500">{label}</p>
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
    <div className="rounded-lg border border-dashed border-slate-300 px-4 py-10 text-center">
      <p className="text-sm text-slate-500">{message}</p>
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
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-200">
            {columns.map((column) => (
              <th
                key={column.header}
                scope="col"
                className={`px-3 py-2 text-xs font-medium tracking-wide text-slate-500 uppercase ${
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
            <tr key={index} className="border-b border-slate-100 last:border-0">
              {columns.map((column) => (
                <td
                  key={column.header}
                  className={`px-3 py-2.5 text-slate-700 ${
                    column.align === 'right' ? 'text-right tabular-nums' : 'text-left'
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
      <span className="h-1.5 w-20 overflow-hidden rounded-full bg-slate-200">
        <span className="block h-full rounded-full bg-navy-600" style={{ width: `${safe}%` }} />
      </span>
      <span className="w-9 text-right text-xs tabular-nums text-slate-600">{safe}%</span>
    </span>
  )
}
