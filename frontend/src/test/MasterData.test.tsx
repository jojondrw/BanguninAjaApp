import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { AccessLevel, User } from '../models/auth'
import type { Account } from '../models/master'
import { AccessScope } from '../views/access/AccessScope'
import { AccountsSection } from '../views/master/AccountsSection'
import { renderWithProviders } from './render'

const master = vi.hoisted(() => ({
  useAccounts: vi.fn(),
  useCreateAccount: vi.fn(),
  useUpdateAccount: vi.fn(),
  useDeleteAccount: vi.fn(),
}))

vi.mock('../controllers/useMaster', () => master)

function account(code: string, name: string, extra: Partial<Account> = {}): Account {
  return {
    id: `id-${code}`,
    code,
    name,
    type: 'asset',
    parentId: null,
    parentCode: null,
    parentName: null,
    postable: true,
    createdAt: '2026-10-01T00:00:00Z',
    updatedAt: '2026-10-01T00:00:00Z',
    ...extra,
  }
}

const ACCOUNTS = [
  account('1100', 'Kas dan Setara Kas', { postable: false }),
  account('1110', 'Kas', { parentId: 'id-1100', parentCode: '1100', parentName: 'Kas dan Setara Kas' }),
  account('1120', 'Bank', { parentId: 'id-1100', parentCode: '1100', parentName: 'Kas dan Setara Kas' }),
]

function signIn(level: AccessLevel) {
  const user: User = {
    id: 'u-1', name: 'Uji', email: 'uji@contoh.id', active: true,
    role: { code: 'admin', name: 'Admin' }, permissions: { master: level }, createdAt: '2026-01-01T00:00:00Z',
  }
  useSessionStore.setState({ accessToken: null, user })
}

function mutation() {
  return { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: false, variables: undefined }
}

function setup() {
  renderWithProviders(
    <AccessScope module="master">
      <AccountsSection />
    </AccessScope>,
  )
  return userEvent.setup()
}

describe('Bagan akun di Data Master', () => {
  beforeEach(() => {
    master.useAccounts.mockReturnValue({
      isPending: false,
      isError: false,
      data: { items: ACCOUNTS, page: 1, pageSize: 20, totalItems: ACCOUNTS.length, totalPages: 1 },
    })
    master.useCreateAccount.mockReturnValue(mutation())
    master.useUpdateAccount.mockReturnValue(mutation())
    master.useDeleteAccount.mockReturnValue(mutation())
  })

  it('menampilkan induk dan status bisa diposting', () => {
    signIn('read')
    setup()

    const bank = screen.getByRole('row', { name: /1120/ })
    expect(within(bank).getByText('1100 Kas dan Setara Kas')).toBeInTheDocument()
    expect(within(bank).getByText('Ya')).toBeInTheDocument()

    const parent = screen.getByRole('row', { name: /^1100/ })
    expect(within(parent).getByText('Akun utama')).toBeInTheDocument()
    expect(within(parent).getByText('Tidak, akun induk')).toBeInTheDocument()
  })

  it('menyembunyikan tambah, ubah, dan hapus untuk peran yang hanya boleh melihat', () => {
    signIn('read')
    setup()

    expect(screen.queryByRole('button', { name: 'Tambah akun' })).not.toBeInTheDocument()
    expect(screen.queryByRole('columnheader', { name: 'Aksi' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Ubah|Hapus/ })).not.toBeInTheDocument()
  })

  it('meminta konfirmasi di baris sebelum menghapus, dan tidak menawarkan hapus akun Kas', async () => {
    signIn('write')
    const deleteAccount = mutation()
    master.useDeleteAccount.mockReturnValue(deleteAccount)
    const user = setup()

    expect(screen.getByRole('button', { name: 'Tambah akun' })).toBeInTheDocument()
    expect(within(screen.getByRole('row', { name: /1110/ })).queryByRole('button', { name: /Hapus/ })).not.toBeInTheDocument()

    const bank = screen.getByRole('row', { name: /1120/ })
    await user.click(within(bank).getByRole('button', { name: 'Hapus: 1120 Bank' }))
    expect(within(bank).getByText('Hapus akun 1120?')).toBeInTheDocument()

    await user.click(within(bank).getByRole('button', { name: 'Ya, hapus' }))
    expect(deleteAccount.mutate).toHaveBeenCalledWith(ACCOUNTS[2], expect.anything())
  })
})
