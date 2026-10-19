import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { defaultQrisAmount, qrisAmountError, qrisPayLink, type GatewayPublicPayment } from '../models/gateway'
import { QrisPayPage } from '../views/pay/QrisPayPage'
import { renderWithProviders } from './render'

const TOKEN = 'AbCdEfGhIjKlMnOpQrStUvWxYz012345'

const PENDING: GatewayPublicPayment = {
  targetKind: 'receivable',
  targetNumber: 'DEMO-PIU-009',
  projectName: 'Cluster Dago',
  amountIdr: 5_000_000,
  provider: 'xendit',
  channel: 'QRIS',
  qrString: '00020101021226QRIS-UJI',
  status: 'pending',
  failureReason: '',
  testMode: true,
  expiresAt: '2099-01-01T00:30:00Z',
  paidAt: null,
  createdAt: '2099-01-01T00:00:00Z',
}

const PAID: GatewayPublicPayment = { ...PENDING, status: 'paid', qrString: '', testMode: false, paidAt: '2099-01-01T00:05:00Z' }

function json(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function mockServer(refresh: () => GatewayPublicPayment, simulate?: () => GatewayPublicPayment) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, _init?: RequestInit) => {
    const url = String(input)
    if (url.endsWith(`/pay/qris/${TOKEN}/refresh`)) {
      return json(200, { data: refresh() })
    }
    if (url.endsWith(`/pay/qris/${TOKEN}/simulate`) && simulate) {
      return json(200, { data: simulate() })
    }
    return json(404, { error: { code: 'gateway_request_not_found', message: 'Permintaan bayar QRIS tidak ditemukan' } })
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

function renderPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/bayar/qris/:token" element={<QrisPayPage />} />
    </Routes>,
    { route: `/bayar/qris/${TOKEN}` },
  )
}

describe('QrisPayPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('menampilkan QRIS dan nominal, lalu berubah jadi lunas setelah simulasi di mode uji', async () => {
    const fetchMock = mockServer(() => PENDING, () => PAID)
    renderPage()

    expect(await screen.findByRole('img', { name: 'Kode QRIS' })).toBeInTheDocument()
    expect(screen.getByText(/DEMO-PIU-009/)).toBeInTheDocument()
    expect(screen.getByText(/QRIS lewat Xendit/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Simulasikan pembayaran' }))
    expect(await screen.findByText('Pembayaran diterima')).toBeInTheDocument()
    expect(screen.queryByRole('img', { name: 'Kode QRIS' })).not.toBeInTheDocument()
    expect(fetchMock.mock.calls.every(([, init]) => (init as RequestInit | undefined)?.credentials === 'omit')).toBe(true)
  })

  it('tidak menawarkan simulasi di luar mode uji dan menjelaskan QRIS yang kedaluwarsa', async () => {
    mockServer(() => ({ ...PENDING, status: 'expired', qrString: '', testMode: false }))
    renderPage()

    expect(await screen.findByText(/sudah kedaluwarsa/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Simulasikan pembayaran' })).not.toBeInTheDocument()
  })

  it('menampilkan galat untuk token yang tidak dikenal', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => json(404, { error: { code: 'gateway_request_not_found', message: 'Permintaan bayar QRIS tidak ditemukan' } })))
    renderPage()

    await waitFor(() => expect(screen.getByText('Halaman bayar tidak bisa dibuka')).toBeInTheDocument())
    expect(screen.queryByRole('button', { name: 'Coba lagi' })).not.toBeInTheDocument()
  })
})

describe('aturan nominal QRIS', () => {
  it('mengisi sisa tagihan tetapi tidak lebih dari batas per transaksi', () => {
    expect(defaultQrisAmount(25_000_000, 10_000_000)).toBe(10_000_000)
    expect(defaultQrisAmount(4_000_000, 10_000_000)).toBe(4_000_000)
  })

  it('menolak nominal kosong, melebihi sisa, atau melebihi batas', () => {
    expect(qrisAmountError(0, 5_000_000, 10_000_000)).not.toBeNull()
    expect(qrisAmountError(1.5, 5_000_000, 10_000_000)).not.toBeNull()
    expect(qrisAmountError(6_000_000, 5_000_000, 10_000_000)).toMatch(/sisa tagihan/)
    expect(qrisAmountError(12_000_000, 20_000_000, 10_000_000)).toMatch(/Rp 10 juta/)
    expect(qrisAmountError(5_000_000, 5_000_000, 10_000_000)).toBeNull()
  })

  it('membuat tautan halaman bayar QRIS', () => {
    expect(qrisPayLink('https://banguninaja.vercel.app', TOKEN)).toBe(`https://banguninaja.vercel.app/bayar/qris/${TOKEN}`)
  })
})
