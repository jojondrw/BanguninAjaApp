import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { User } from '../models/auth'
import { RouteAccess } from '../views/access/RouteAccess'
import { UsersPage } from '../views/users/UsersPage'
import { renderWithProviders } from './render'

const PENDING_USER: User = {
  id: 'u-pending', name: 'Uji Tamu', email: 'uji-tamu@contoh.id', active: true,
  role: null, permissions: {}, createdAt: '2026-10-01T00:00:00Z',
}

const SALES_USER: User = {
  ...PENDING_USER,
  role: { code: 'penjualan', name: 'Penjualan' },
  permissions: { sales: 'write', projects: 'read' },
}

const ADMIN: User = {
  id: 'u-admin', name: 'Uji Admin', email: 'uji-admin@contoh.id', active: true,
  role: { code: 'admin', name: 'Admin' }, permissions: { users: 'write', projects: 'write' }, createdAt: '2026-01-01T00:00:00Z',
}

function reply(data: unknown) {
  return Promise.resolve({ ok: true, status: 200, json: async () => ({ data }) } as Response)
}

function page<T>(items: T[], totalItems = items.length) {
  return { items, page: 1, pageSize: 20, totalItems, totalPages: totalItems > 0 ? 1 : 0 }
}

afterEach(() => {
  vi.unstubAllGlobals()
  useSessionStore.setState({ accessToken: null, user: null })
})

describe('RouteAccess for an account without role', () => {
  it('shows the waiting screen instead of the app', () => {
    vi.stubGlobal('fetch', vi.fn(() => new Promise(() => {})))
    useSessionStore.setState({ accessToken: null, user: PENDING_USER })

    renderWithProviders(<RouteAccess><p>Isi aplikasi</p></RouteAccess>, { route: '/proyek' })

    expect(screen.getByRole('heading', { name: 'Akun menunggu persetujuan' })).toBeInTheDocument()
    expect(screen.getByText('uji-tamu@contoh.id')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Keluar' })).toBeInTheDocument()
    expect(screen.queryByText('Isi aplikasi')).not.toBeInTheDocument()
  })

  it('opens the app once "Periksa lagi" finds a role', async () => {
    const profiles = [PENDING_USER, SALES_USER]
    const fetchMock = vi.fn((url: string) => (url.endsWith('/auth/me') ? reply(profiles.shift() ?? SALES_USER) : new Promise(() => {})))
    vi.stubGlobal('fetch', fetchMock)
    useSessionStore.setState({ accessToken: 'token-uji', user: PENDING_USER })
    const user = userEvent.setup()

    renderWithProviders(<RouteAccess><p>Isi aplikasi</p></RouteAccess>, { route: '/' })
    await waitFor(() => expect(screen.getByRole('button', { name: 'Periksa lagi' })).toBeEnabled())
    await user.click(screen.getByRole('button', { name: 'Periksa lagi' }))

    expect(await screen.findByText('Isi aplikasi')).toBeInTheDocument()
    expect(fetchMock.mock.calls.filter(([url]) => url.endsWith('/auth/me'))).toHaveLength(2)
  })

  it('leaves users with a role alone', () => {
    useSessionStore.setState({ accessToken: null, user: SALES_USER })

    renderWithProviders(<RouteAccess><p>Isi aplikasi</p></RouteAccess>, { route: '/penjualan' })

    expect(screen.getByText('Isi aplikasi')).toBeInTheDocument()
  })
})

describe('UsersPage pending registrations', () => {
  const pendingRow = { id: 'u-pending', name: 'Uji Tamu', email: 'uji-tamu@contoh.id', active: true, role: null, createdAt: '2026-10-01T00:00:00Z' }

  it('flags pending users and lets the admin give one a role', async () => {
    const fetchMock = vi.fn((url: string, init?: RequestInit) => {
      if (init?.method === 'PATCH') {
        return reply({ ...pendingRow, role: { code: 'penjualan', name: 'Penjualan' } })
      }
      if (url.includes('role=none')) {
        return reply(page([pendingRow], 2))
      }
      return reply(page([pendingRow]))
    })
    vi.stubGlobal('fetch', fetchMock)
    useSessionStore.setState({ accessToken: null, user: ADMIN })
    const user = userEvent.setup()

    renderWithProviders(<UsersPage />, { route: '/pengguna' })

    expect(await screen.findByText(/2 pendaftar menunggu peran/)).toBeInTheDocument()
    const roleSelect = await screen.findByLabelText('Peran Uji Tamu')
    expect(within(roleSelect).getByRole('option', { name: 'Menunggu peran' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Menunggu peran' }))
    expect(screen.getByRole('button', { name: 'Menunggu peran' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByLabelText('Saring peran')).toHaveValue('none')

    await user.selectOptions(screen.getByLabelText('Peran Uji Tamu'), 'penjualan')
    await user.click(screen.getByRole('button', { name: 'Simpan peran' }))

    await waitFor(() => {
      const patch = fetchMock.mock.calls.find(([, init]) => init?.method === 'PATCH')
      expect(patch?.[0]).toMatch(/\/users\/u-pending\/role$/)
      expect(JSON.parse(String(patch?.[1]?.body))).toEqual({ role: 'penjualan' })
    })
  })
})
