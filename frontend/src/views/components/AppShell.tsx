import type { ReactNode } from 'react'
import { NavLink, useNavigate } from 'react-router-dom'

import { useLogout, useSession } from '../../controllers/useAuth'
import { Button } from './Form'

interface MenuItem {
  to: string
  label: string
  note?: string
}

const MENU: MenuItem[] = [
  { to: '/', label: 'Ringkasan' },
  { to: '/proyek', label: 'Proyek' },
  { to: '/keuangan', label: 'Keuangan' },
  { to: '/lokasi', label: 'Analisis Lokasi', note: 'Menyusul' },
]

interface AppShellProps {
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}

export function AppShell({ title, description, actions, children }: AppShellProps) {
  const { user } = useSession()
  const logout = useLogout()
  const navigate = useNavigate()

  const endSession = () => {
    logout.mutate(undefined, { onSettled: () => navigate('/masuk', { replace: true }) })
  }

  return (
    <div className="min-h-screen bg-slate-50 lg:grid lg:grid-cols-[240px_1fr]">
      <aside className="hidden border-r border-slate-200 bg-white lg:flex lg:flex-col">
        <div className="flex h-16 items-center border-b border-slate-200 px-6">
          <span className="text-sm font-semibold tracking-tight text-slate-900">BanguninAja</span>
        </div>

        <nav className="flex flex-1 flex-col gap-1 p-3" aria-label="Menu utama">
          {MENU.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                `flex items-center justify-between rounded-lg px-3 py-2 text-sm transition ${
                  isActive
                    ? 'bg-navy-50 font-medium text-navy-700'
                    : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
                }`
              }
            >
              {item.label}
              {item.note ? (
                <span className="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] text-slate-500">
                  {item.note}
                </span>
              ) : null}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-slate-200 p-3">
          <p className="truncate px-3 text-xs text-slate-500">{user?.email ?? ''}</p>
          <div className="mt-2 px-1">
            <Button variant="subtle" isPending={logout.isPending} pendingLabel="Keluar" onClick={endSession}>
              Keluar
            </Button>
          </div>
        </div>
      </aside>

      <div className="flex min-h-screen flex-col">
        <header className="border-b border-slate-200 bg-white">
          <div className="flex flex-wrap items-center justify-between gap-3 px-6 py-4">
            <div>
              <h1 className="text-lg font-semibold tracking-tight text-slate-900">{title}</h1>
              {description ? <p className="mt-1 text-sm text-slate-600">{description}</p> : null}
            </div>
            {actions}
          </div>

          <nav className="flex gap-1 overflow-x-auto px-4 pb-2 lg:hidden" aria-label="Menu ponsel">
            {MENU.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) =>
                  `rounded-lg px-3 py-1.5 text-sm whitespace-nowrap ${
                    isActive ? 'bg-navy-50 font-medium text-navy-700' : 'text-slate-600'
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </header>

        <main className="flex-1 px-6 py-6">{children}</main>
      </div>
    </div>
  )
}
