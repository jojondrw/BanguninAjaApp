import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { PublicPayment } from '../models/onchain'
import { PayPage } from '../views/pay/PayPage'
import { renderWithProviders } from './render'

const TOKEN = 'AbCdEfGhIjKlMnOpQrStUvWxYz012345'
const HASH = '0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f'
const RECEIVER = '0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed'

const PENDING: PublicPayment = {
  targetKind: 'invoice',
  targetNumber: 'UJI-INV-7',
  projectName: 'Griya Asri Cibubur',
  amountIdr: 50_000_000,
  amountWei: '50000000000000',
  amountEth: '0.00005',
  idrPerEth: 1e12,
  receiver: RECEIVER,
  chainId: 11155111,
  paymentUri: `ethereum:${RECEIVER}@11155111?value=50000000000000`,
  explorerUrl: 'https://sepolia.etherscan.io',
  status: 'pending',
  txHash: null,
  txUrl: null,
  payerAddress: null,
  failureReason: '',
  expiresAt: '2099-01-01T00:30:00Z',
  paidAt: null,
  createdAt: '2099-01-01T00:00:00Z',
}

const PAID: PublicPayment = {
  ...PENDING,
  status: 'paid',
  txHash: HASH,
  txUrl: `https://sepolia.etherscan.io/tx/${HASH}`,
  payerAddress: '0x690c39adabdea83322bf8e90626cd40eeb456a95',
  paidAt: '2099-01-01T00:05:00Z',
}

interface Reply {
  status: number
  body: unknown
}

function json({ status, body }: Reply): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function mockServer(routes: { page: () => Reply; confirm?: (body: { txHash: string }) => Reply }) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)
    if (url.endsWith(`/pay/${TOKEN}`)) {
      return json(routes.page())
    }
    if (url.endsWith(`/pay/${TOKEN}/confirm`) && routes.confirm) {
      return json(routes.confirm(JSON.parse(String(init?.body))))
    }
    return json({ status: 404, body: { error: { code: 'not_found', message: 'Tidak ditemukan' } } })
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

function renderPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/bayar/:token" element={<PayPage />} />
    </Routes>,
    { route: `/bayar/${TOKEN}` },
  )
}

describe('PayPage', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    delete (window as unknown as { ethereum?: unknown }).ethereum
  })

  it('menampilkan tagihan, nominal dengan kurs simulasi, QR, dan penjelasan jaringan uji tanpa login', async () => {
    const fetchMock = mockServer({ page: () => ({ status: 200, body: { data: PENDING } }) })
    renderPage()

    expect(await screen.findByRole('heading', { name: 'Bayar faktur UJI-INV-7' })).toBeInTheDocument()
    expect(screen.getByText('Proyek Griya Asri Cibubur')).toBeInTheDocument()
    expect(screen.getByText(/0,00005 SepoliaETH/)).toBeInTheDocument()
    expect(screen.getByText(/Kurs simulasi: 1 SepoliaETH/)).toBeInTheDocument()
    expect(screen.getByRole('img', { name: 'Kode QR pembayaran untuk dompet di HP' })).toBeInTheDocument()
    expect(screen.getByText(/bukan uang sungguhan/)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Buka di aplikasi MetaMask' })).toBeInTheDocument()

    const [, init] = fetchMock.mock.calls[0]
    expect(init?.credentials).toBe('omit')
    expect(new Headers(init?.headers).has('Authorization')).toBe(false)
  })

  it('memeriksa hash yang ditempel dan menampilkan lunas dengan tautan Etherscan', async () => {
    const confirm = vi.fn(() => ({ status: 200, body: { data: { ...PAID, verification: 'paid', message: 'ok', confirmations: 3 } } }))
    mockServer({ page: () => ({ status: 200, body: { data: PENDING } }), confirm })
    const user = userEvent.setup()
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Saya sudah bayar, tempel hash transaksi' }))
    await user.type(screen.getByLabelText('Hash transaksi'), '0x1234')
    await user.click(screen.getByRole('button', { name: 'Periksa transaksi' }))
    expect(screen.getByText(/harus 0x diikuti 64 karakter/)).toBeInTheDocument()
    expect(confirm).not.toHaveBeenCalled()

    await user.clear(screen.getByLabelText('Hash transaksi'))
    await user.paste(` ${HASH.toUpperCase().replace('0X', '0x')} `)
    await user.click(screen.getByRole('button', { name: 'Periksa transaksi' }))

    expect(await screen.findByText(/Pembayaran diterima dan sudah tercatat/)).toBeInTheDocument()
    expect(confirm).toHaveBeenCalledWith({ txHash: HASH })
    expect(screen.getByRole('link', { name: /0xc9be6b…a62b0f/ })).toHaveAttribute('href', `https://sepolia.etherscan.io/tx/${HASH}`)
  })

  it('membayar lewat MetaMask lalu menunggu konfirmasi jaringan', async () => {
    const calls: string[] = []
    const wallet = {
      request: vi.fn(async ({ method, params }: { method: string; params?: unknown[] }) => {
        calls.push(method)
        if (method === 'eth_requestAccounts') {
          return ['0x690c39adabdea83322bf8e90626cd40eeb456a95']
        }
        if (method === 'eth_sendTransaction') {
          expect(params).toEqual([{ from: '0x690c39adabdea83322bf8e90626cd40eeb456a95', to: RECEIVER, value: '0x2d79883d2000' }])
          return HASH
        }
        return null
      }),
    }
    ;(window as unknown as { ethereum: typeof wallet }).ethereum = wallet
    mockServer({
      page: () => ({ status: 200, body: { data: PENDING } }),
      confirm: () => ({ status: 202, body: { data: { ...PENDING, verification: 'waiting', message: 'Transaksi sudah terkirim dan sedang menunggu masuk blok', confirmations: 0 } } }),
    })
    const user = userEvent.setup()
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Bayar dengan MetaMask' }))

    expect(await screen.findByText('Menunggu konfirmasi jaringan')).toBeInTheDocument()
    expect(await screen.findByText(/sedang menunggu masuk blok/)).toBeInTheDocument()
    expect(calls).toEqual(['eth_requestAccounts', 'wallet_switchEthereumChain', 'eth_sendTransaction'])
    expect(window.sessionStorage.getItem(`bayar:${TOKEN}`)).toBe(HASH)
  })

  it('menampilkan alasan penolakan dari server lalu mengizinkan mencoba lagi', async () => {
    mockServer({
      page: () => ({ status: 200, body: { data: PENDING } }),
      confirm: () => ({
        status: 422,
        body: { error: { code: 'onchain_wrong_receiver', message: 'Transaksi ini dikirim ke alamat lain, bukan ke alamat penerima perusahaan' } },
      }),
    })
    const user = userEvent.setup()
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Saya sudah bayar, tempel hash transaksi' }))
    await user.click(screen.getByLabelText('Hash transaksi'))
    await user.paste(HASH)
    await user.click(screen.getByRole('button', { name: 'Periksa transaksi' }))

    expect(await screen.findByText(/dikirim ke alamat lain/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Saya sudah bayar, tempel hash transaksi' })).toBeInTheDocument()
    await waitFor(() => expect(window.sessionStorage.getItem(`bayar:${TOKEN}`)).toBeNull())
  })

  it('menjelaskan tautan yang kedaluwarsa atau tidak ditemukan', async () => {
    mockServer({ page: () => ({ status: 200, body: { data: { ...PENDING, status: 'expired' } } }) })
    const { unmount } = renderPage()
    expect(await screen.findByText(/sudah kedaluwarsa/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Bayar dengan MetaMask' })).not.toBeInTheDocument()
    unmount()

    mockServer({
      page: () => ({ status: 404, body: { error: { code: 'onchain_request_not_found', message: 'Permintaan bayar tidak ditemukan. Periksa lagi tautannya' } } }),
    })
    renderPage()
    expect(await screen.findByText('Halaman bayar tidak bisa dibuka')).toBeInTheDocument()
    expect(screen.getByText(/Periksa lagi tautannya/)).toBeInTheDocument()
  })
})
