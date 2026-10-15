import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { User } from '../models/auth'
import { AccountPage } from '../views/account/AccountPage'
import { renderWithProviders } from './render'

const USER: User = {
  id: 'u-1', name: 'Uji Lama', email: 'uji@contoh.id', active: true,
  role: { code: 'keuangan', name: 'Keuangan' }, permissions: { finance: 'write' }, createdAt: '2026-01-01T00:00:00Z',
}

type Reply = { status: number; body?: unknown }

let replies: Reply[] = []
const fetchMock = vi.fn((_url: string, _init?: RequestInit) => {
  const next = replies.shift() ?? { status: 204 }
  return Promise.resolve({
    ok: next.status < 400,
    status: next.status,
    json: async () => next.body ?? {},
  } as Response)
})

function patches(path: string) {
  return fetchMock.mock.calls.filter(([url, init]) => url.endsWith(path) && init?.method === 'PATCH')
}

beforeEach(() => {
  replies = []
  fetchMock.mockClear()
  vi.stubGlobal('fetch', fetchMock)
  useSessionStore.setState({ accessToken: null, user: USER })
})

afterEach(() => {
  vi.unstubAllGlobals()
  useSessionStore.setState({ accessToken: null, user: null })
})

function setup() {
  renderWithProviders(<AccountPage />, { route: '/akun' })
  return userEvent.setup()
}

describe('AccountPage profile', () => {
  it('shows email and role read-only and saves a trimmed name into the session', async () => {
    const user = setup()
    expect(screen.getByText('uji@contoh.id')).toBeInTheDocument()
    expect(screen.getAllByText('Keuangan').length).toBeGreaterThan(0)

    replies.push({ status: 200, body: { data: { ...USER, name: 'Uji Baru' } } })
    await user.clear(screen.getByLabelText('Nama'))
    await user.type(screen.getByLabelText('Nama'), '  Uji Baru  ')
    await user.click(screen.getByRole('button', { name: 'Simpan nama' }))

    expect(await screen.findByText('Nama tersimpan.')).toBeInTheDocument()
    expect(JSON.parse(String(patches('/auth/me')[0][1]?.body))).toEqual({ name: 'Uji Baru' })
    expect(useSessionStore.getState().user?.name).toBe('Uji Baru')
  })

  it('checks the name length before calling the server', async () => {
    const user = setup()
    await user.clear(screen.getByLabelText('Nama'))
    await user.type(screen.getByLabelText('Nama'), ' a ')
    await user.click(screen.getByRole('button', { name: 'Simpan nama' }))

    expect(screen.getByRole('alert')).toHaveTextContent('Nama harus 2 sampai 100 karakter.')
    expect(patches('/auth/me')).toHaveLength(0)
  })
})

describe('AccountPage password', () => {
  async function fillPasswords(user: ReturnType<typeof userEvent.setup>, current: string, next: string, confirm: string) {
    await user.type(screen.getByLabelText('Kata sandi saat ini'), current)
    await user.type(screen.getByLabelText('Kata sandi baru'), next)
    await user.type(screen.getByLabelText('Ulangi kata sandi baru'), confirm)
    await user.click(screen.getByRole('button', { name: 'Ganti kata sandi' }))
  }

  it('refuses a confirmation that does not match', async () => {
    const user = setup()
    await fillPasswords(user, 'rahasia123', 'sandi-baru-456', 'sandi-baru-457')

    expect(screen.getByText('Belum sama dengan kata sandi baru')).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent('Ulangan kata sandi baru belum sama.')
    expect(patches('/auth/password')).toHaveLength(0)
  })

  it('sends current and new password, then clears the form', async () => {
    const user = setup()
    replies.push({ status: 204 })
    await fillPasswords(user, 'rahasia123', 'sandi-baru-456', 'sandi-baru-456')

    expect(await screen.findByText(/Kata sandi diganti/)).toBeInTheDocument()
    expect(JSON.parse(String(patches('/auth/password')[0][1]?.body))).toEqual({
      currentPassword: 'rahasia123',
      newPassword: 'sandi-baru-456',
    })
    expect(screen.getByLabelText('Kata sandi saat ini')).toHaveValue('')
    expect(screen.getByLabelText('Kata sandi baru')).toHaveValue('')
  })

  it('shows the server message when the current password is wrong', async () => {
    const user = setup()
    replies.push({ status: 422, body: { error: { code: 'wrong_current_password', message: 'Kata sandi saat ini salah' } } })
    await fillPasswords(user, 'salah-12345', 'sandi-baru-456', 'sandi-baru-456')

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Kata sandi saat ini salah'))
    expect(screen.getByLabelText('Kata sandi saat ini')).toHaveValue('salah-12345')
  })
})
