import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { AccessScopeContext } from '../controllers/useAccess'
import { useSessionStore } from '../controllers/sessionStore'
import type { Permissions, User } from '../models/auth'
import type { BillingRecord } from '../models/billing'
import type { OnchainConfig, OnchainRequest } from '../models/onchain'
import { OnchainSection } from '../views/billing/OnchainPanel'
import { renderWithProviders } from './render'

const INVOICE: BillingRecord = {
  id: 'inv-1',
  dueDate: '2099-01-31T00:00:00Z',
  amount: 80_000_000,
  paidAmount: 30_000_000,
  outstanding: 50_000_000,
  status: 'not_due',
  daysOverdue: 0,
  lastPaidAt: null,
  createdAt: '2099-01-01T00:00:00Z',
  updatedAt: '2099-01-01T00:00:00Z',
}

const CONFIG: OnchainConfig = {
  enabled: true,
  chainId: 11155111,
  receiver: '0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed',
  idrPerEth: 1e12,
  explorerUrl: 'https://sepolia.etherscan.io',
  minConfirmations: 1,
  requestTtlMinutes: 30,
}

const ACTIVE: OnchainRequest = {
  id: 'req-1',
  targetKind: 'invoice',
  targetId: 'inv-1',
  token: 'AbCdEfGhIjKlMnOpQrStUvWxYz012345',
  amountIdr: 50_000_000,
  amountWei: '50000000000000',
  amountEth: '0.00005',
  idrPerEth: 1e12,
  receiver: CONFIG.receiver,
  chainId: 11155111,
  paymentUri: `ethereum:${CONFIG.receiver}@11155111?value=50000000000000`,
  status: 'pending',
  txHash: null,
  txUrl: null,
  payerAddress: null,
  failureReason: '',
  createdBy: 'u-1',
  createdByName: 'Uji Keuangan',
  expiresAt: '2099-01-01T00:30:00Z',
  paidAt: null,
  createdAt: '2099-01-01T00:00:00Z',
}

function page(items: OnchainRequest[]) {
  return { items, page: 1, pageSize: 5, totalItems: items.length, totalPages: items.length > 0 ? 1 : 0 }
}

function signIn(permissions: Permissions) {
  const user: User = {
    id: 'u-1', name: 'Uji', email: 'uji@contoh.id', active: true,
    role: { code: 'keuangan', name: 'Keuangan' }, permissions, createdAt: '2026-01-01T00:00:00Z',
  }
  useSessionStore.setState({ accessToken: null, user })
}

function mockApi(config: OnchainConfig, items: OnchainRequest[]) {
  const created: unknown[] = []
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)
    let body: unknown = { error: { code: 'not_found', message: 'Tidak ditemukan' } }
    let status = 404
    if (url.endsWith('/billing/onchain/config')) {
      ;[status, body] = [200, { data: config }]
    } else if (url.includes('/billing/invoices/inv-1/onchain-requests') && init?.method === 'POST') {
      created.push(JSON.parse(String(init.body)))
      ;[status, body] = [201, { data: ACTIVE }]
    } else if (url.includes('/billing/invoices/inv-1/onchain-requests')) {
      ;[status, body] = [200, { data: page(items) }]
    }
    return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
  })
  vi.stubGlobal('fetch', fetchMock)
  return { fetchMock, created }
}

describe('OnchainSection', () => {
  beforeEach(() => signIn({ billing: 'write' }))

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('tidak tampil kalau fitur dimatikan di server', async () => {
    const { fetchMock } = mockApi({ ...CONFIG, enabled: false }, [])
    renderWithProviders(<OnchainSection kind="invoices" record={INVOICE} />)

    await waitFor(() => expect(fetchMock).toHaveBeenCalled())
    expect(screen.queryByText('Pembayaran on-chain')).not.toBeInTheDocument()
    expect(fetchMock.mock.calls.every(([url]) => !String(url).includes('onchain-requests'))).toBe(true)
  })

  it('membuat permintaan bayar sebesar sisa tagihan dengan pratinjau kurs simulasi', async () => {
    const { created } = mockApi(CONFIG, [])
    const user = userEvent.setup()
    renderWithProviders(<OnchainSection kind="invoices" record={INVOICE} />)

    await user.click(await screen.findByRole('button', { name: 'Bayar on-chain' }))
    expect(screen.getByLabelText('Nominal yang ditagih')).toHaveValue(50_000_000)
    expect(screen.getByText('0,00005 SepoliaETH')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Buat permintaan bayar' }))

    expect(await screen.findByText(/Permintaan bayar dibuat/)).toBeInTheDocument()
    expect(created).toEqual([{ amount: 50_000_000 }])
  })

  it('menampilkan QR, tautan bayar, dan pembatalan dua langkah untuk permintaan aktif', async () => {
    mockApi(CONFIG, [ACTIVE])
    const user = userEvent.setup()
    renderWithProviders(<OnchainSection kind="invoices" record={INVOICE} />)

    expect(await screen.findByRole('img', { name: 'Kode QR pembayaran on-chain' })).toBeInTheDocument()
    expect(screen.getByLabelText('Tautan halaman bayar untuk pembeli')).toHaveValue(`${window.location.origin}/bayar/${ACTIVE.token}`)
    expect(screen.queryByRole('button', { name: 'Bayar on-chain' })).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Batalkan permintaan' }))
    expect(screen.getByText(/Tautannya tidak bisa dipakai lagi/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Ya, batalkan' })).toBeInTheDocument()
  })

  it('menyembunyikan tombol bayar dan pembatalan untuk peran yang hanya melihat', async () => {
    signIn({ billing: 'read' })
    mockApi(CONFIG, [ACTIVE])
    renderWithProviders(
      <AccessScopeContext.Provider value="billing">
        <OnchainSection kind="invoices" record={INVOICE} />
      </AccessScopeContext.Provider>,
    )

    expect(await screen.findByRole('img', { name: 'Kode QR pembayaran on-chain' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Bayar on-chain' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Batalkan permintaan' })).not.toBeInTheDocument()
  })
})
