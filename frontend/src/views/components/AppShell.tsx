import { CircleUserRound, LogOut, Search } from 'lucide-react'
import { type ReactNode, useEffect, useRef, useState } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'

import { AccessScopeContext, canAccess, usePermissions } from '../../controllers/useAccess'
import { useCurrentUser, useLogout } from '../../controllers/useAuth'
import { BrandName } from './Brand'
import { CommandPalette } from './CommandPalette'
import { type MenuItem, moduleForPath, visibleMenu } from './navigation'

const SCROLL_EDGE_PX = 4
const PALETTE_SHORTCUT = /Mac|iPhone|iPad/.test(navigator.userAgent) ? '⌘K' : 'Ctrl K'

interface AppShellProps {
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}

// Modul halaman diturunkan dari rute, supaya tombol bersama (FormToggle,
// RowAction) tahu hak akses yang berlaku tanpa diatur satu per satu.
export function AppShell(props: AppShellProps) {
  const { pathname } = useLocation()

  return (
    <AccessScopeContext.Provider value={moduleForPath(pathname)}>
      <ShellFrame {...props} />
    </AccessScopeContext.Provider>
  )
}

function ShellFrame({ title, description, actions, children }: AppShellProps) {
  const user = useCurrentUser()
  const permissions = usePermissions()
  const logout = useLogout()
  const navigate = useNavigate()
  const menu = visibleMenu((module) => canAccess(permissions, module, 'read'))
  const menuItems = menu.flatMap((group) => group.items)
  const panelRef = useRef<HTMLDivElement>(null)
  const mobileNavRef = useRef<HTMLElement>(null)
  const [isScrolled, setIsScrolled] = useState(false)
  const [isPaletteOpen, setIsPaletteOpen] = useState(false)

  useEffect(() => {
    const panel = panelRef.current
    const update = () => setIsScrolled(Math.max(panel?.scrollTop ?? 0, window.scrollY) > SCROLL_EDGE_PX)
    panel?.addEventListener('scroll', update, { passive: true })
    window.addEventListener('scroll', update, { passive: true })
    return () => {
      panel?.removeEventListener('scroll', update)
      window.removeEventListener('scroll', update)
    }
  }, [])

  useEffect(() => {
    mobileNavRef.current
      ?.querySelector('[aria-current="page"]')
      ?.scrollIntoView({ block: 'nearest', inline: 'center' })
  }, [])

  useEffect(() => {
    document.title = `${title} · BanguninAja`
  }, [title])

  useEffect(() => {
    const openPalette = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        setIsPaletteOpen((open) => !open)
      }
    }
    window.addEventListener('keydown', openPalette)
    return () => window.removeEventListener('keydown', openPalette)
  }, [])

  const endSession = () => {
    logout.mutate(undefined, { onSettled: () => navigate('/masuk', { replace: true }) })
  }

  return (
    <div className="min-h-screen bg-slate-100 lg:grid lg:h-screen lg:grid-cols-[236px_minmax(0,1fr)] lg:overflow-hidden">
      <a
        href="#konten"
        className="sr-only z-50 rounded-lg bg-white px-3 py-2 text-[13px] font-medium text-slate-900 shadow-panel
                   focus:not-sr-only focus:fixed focus:top-3 focus:left-3"
      >
        Lewati ke konten
      </a>

      <aside className="hidden lg:flex lg:min-h-0 lg:flex-col lg:px-3 lg:pt-4 lg:pb-3">
        <BrandName className="h-9 px-2" />

        <button
          type="button"
          onClick={() => setIsPaletteOpen(true)}
          className="mt-4 flex h-8 items-center gap-2 rounded-lg bg-white/70 px-2.5 text-[13px] text-slate-500 shadow-hairline
                     transition-[background-color,transform] hover:bg-white motion-safe:active:scale-[0.98]"
        >
          <Search aria-hidden="true" className="size-3.5" strokeWidth={2} />
          <span className="flex-1 text-left">Cari…</span>
          <kbd className="rounded bg-slate-100 px-1 font-sans text-[10px] font-medium text-slate-500">{PALETTE_SHORTCUT}</kbd>
        </button>

        <nav className="mt-4 flex flex-1 flex-col gap-5 overflow-y-auto" aria-label="Menu utama">
          {menu.map((group) => (
            <div key={group.label}>
              <p className="px-2.5 pb-1.5 text-[11px] font-semibold text-slate-500">{group.label}</p>
              <ul className="flex flex-col gap-0.5">
                {group.items.map((item) => (
                  <li key={item.to}>
                    <SideLink item={item} />
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </nav>

        <div className="mt-3 flex items-center gap-1 rounded-xl px-1 py-1">
          <NavLink
            to="/akun"
            aria-label="Akun saya"
            className={({ isActive }) =>
              `flex min-w-0 flex-1 items-center gap-2.5 rounded-lg px-1 py-1 transition-[background-color,transform]
               motion-safe:active:scale-[0.98] ${isActive ? 'bg-white shadow-hairline' : 'hover:bg-black/[0.045]'}`
            }
          >
            <Avatar name={user?.name ?? user?.email ?? ''} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-[13px] font-medium text-slate-900">{user?.name ?? 'Pengguna'}</p>
              <p className="truncate text-xs text-slate-500" title={user?.email ?? ''}>
                {user?.role?.name ?? 'Tanpa peran'}
              </p>
            </div>
          </NavLink>
          <SignOutButton isPending={logout.isPending} onClick={endSession} />
        </div>
      </aside>

      <div className="flex min-h-screen min-w-0 flex-col lg:min-h-0 lg:py-2 lg:pr-2">
        <div
          ref={panelRef}
          className="flex flex-1 flex-col bg-white lg:min-h-0 lg:overflow-y-auto lg:rounded-2xl lg:shadow-panel"
        >
          <header className="sticky top-0 z-20 bg-white/80 backdrop-blur-xl backdrop-saturate-150">
            <div
              aria-hidden="true"
              className={`pointer-events-none absolute inset-x-0 top-full h-5 bg-gradient-to-b from-slate-900/[0.07] to-transparent
                          transition-opacity duration-300 ${isScrolled ? 'opacity-100' : 'opacity-0'}`}
            />
            <div className="flex items-center justify-between gap-3 px-5 pt-3 lg:hidden">
              <BrandName />
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setIsPaletteOpen(true)}
                  aria-label="Cari halaman atau proyek"
                  className="grid size-8 place-items-center rounded-lg text-slate-500 hover:bg-black/[0.05] hover:text-slate-900"
                >
                  <Search aria-hidden="true" className="size-4" strokeWidth={1.8} />
                </button>
                <NavLink
                  to="/akun"
                  aria-label="Akun saya"
                  title="Akun saya"
                  className="grid size-8 place-items-center rounded-lg text-slate-500 hover:bg-black/[0.05] hover:text-slate-900"
                >
                  <CircleUserRound aria-hidden="true" className="size-4" strokeWidth={1.8} />
                </NavLink>
                <SignOutButton isPending={logout.isPending} onClick={endSession} />
              </div>
            </div>

            <nav
              ref={mobileNavRef}
              className="relative flex gap-1 overflow-x-auto px-4 pt-3 [scrollbar-width:none] lg:hidden"
              aria-label="Menu ponsel"
            >
              {menuItems.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.to === '/'}
                  className={({ isActive }) =>
                    `rounded-full px-3 py-1.5 text-[13px] font-medium whitespace-nowrap transition-colors ${
                      isActive ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-100'
                    }`
                  }
                >
                  {item.label}
                </NavLink>
              ))}
            </nav>

            <div className="px-5 py-4 lg:px-10 lg:py-6">
              <div className="mx-auto flex max-w-[1280px] flex-wrap items-center justify-between gap-x-4 gap-y-3">
                <div className="min-w-0">
                  <h1 className="truncate text-[22px] leading-7 font-semibold tracking-[-0.022em] text-slate-900">
                    {title}
                  </h1>
                  {description ? <p className="mt-1 text-[13px] text-slate-500">{description}</p> : null}
                </div>
                {actions ? <div className="flex items-center gap-2">{actions}</div> : null}
              </div>
            </div>
          </header>

          <main id="konten" tabIndex={-1} className="flex-1 px-5 pt-2 pb-16 focus:outline-none lg:px-10">
            <div className="mx-auto max-w-[1280px] motion-safe:animate-enter">{children}</div>
          </main>
        </div>
      </div>

      <CommandPalette isOpen={isPaletteOpen} onClose={() => setIsPaletteOpen(false)} />
    </div>
  )
}

function SideLink({ item }: { item: MenuItem }) {
  const Icon = item.icon

  return (
    <NavLink
      to={item.to}
      end={item.to === '/'}
      className={({ isActive }) =>
        `group flex h-8 items-center gap-2.5 rounded-lg px-2.5 text-[13px] font-medium transition-[background-color,color,transform]
         duration-150 motion-safe:active:scale-[0.98] ${
           isActive ? 'bg-white text-slate-900 shadow-hairline' : 'text-slate-600 hover:bg-black/[0.045] hover:text-slate-900'
         }`
      }
    >
      {({ isActive }) => (
        <>
          <Icon
            aria-hidden="true"
            strokeWidth={1.8}
            className={`size-4 shrink-0 transition-colors ${
              isActive ? 'text-navy-600' : 'text-slate-400 group-hover:text-slate-600'
            }`}
          />
          {item.label}
        </>
      )}
    </NavLink>
  )
}

function Avatar({ name }: { name: string }) {
  const initials = name
    .split(/[\s@._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('')

  return (
    <span
      aria-hidden="true"
      className="grid size-8 shrink-0 place-items-center rounded-full bg-navy-700 text-[11px] font-semibold text-white"
    >
      {initials || '?'}
    </span>
  )
}

function SignOutButton({ isPending, onClick }: { isPending: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isPending}
      aria-label="Keluar"
      title="Keluar"
      className="grid size-8 shrink-0 place-items-center rounded-lg text-slate-500 transition-[background-color,color,transform]
                 hover:bg-black/[0.05] hover:text-slate-900 disabled:opacity-50 motion-safe:active:scale-95"
    >
      <LogOut aria-hidden="true" className="size-4" strokeWidth={1.8} />
    </button>
  )
}
