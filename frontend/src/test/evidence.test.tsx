import { screen } from '@testing-library/react'
import { Route, Routes } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PublicClaim } from '../models/evidence'
import { sendCalldata, signTypedData, type Eip1193Provider } from '../views/pay/wallet'
import { VerifyClaimPage } from '../views/verify/VerifyClaimPage'
import { renderWithProviders } from './render'

const ACCOUNT = '0x05788638f4e23932884765f666D85654547AecB4'

function fakeWallet(responses: Record<string, unknown>) {
  const calls: { method: string; params?: unknown[] }[] = []
  const wallet: Eip1193Provider = {
    request: async (args) => {
      calls.push(args)
      return responses[args.method] ?? null
    },
  }
  return { wallet, calls }
}

const CLAIM: PublicClaim = {
  projectCode: 'DEMO-RSIA',
  projectName: 'Rumah Sakit Ibu & Anak Cirebon',
  contractNumber: 'DEMO-KTR-RSIA-01',
  ownerName: 'Yayasan Kesehatan Permata Bunda Cirebon',
  label: 'Termin 2',
  progress: 46,
  grossAmount: 13_400_000_000,
  retentionAmount: 670_000_000,
  netAmount: 12_730_000_000,
  claimedOn: '2026-10-01T00:00:00Z',
  items: [
    { kind: 'daily_report', subjectId: 'r-1', label: 'Pengecoran kolom', date: '2026-10-10T00:00:00Z', status: 'anchored', anchoredAt: '2026-10-11T00:00:00Z', txHash: '0x' + 'ab'.repeat(32), txUrl: 'https://sepolia.etherscan.io/tx/0xab' },
    { kind: 'site_photo', subjectId: 'p-1', label: 'Struktur lantai 4', date: '2026-10-14T00:00:00Z', status: 'changed', anchoredAt: '2026-10-15T00:00:00Z', txHash: '0x' + 'ab'.repeat(32), txUrl: 'https://sepolia.etherscan.io/tx/0xab' },
  ],
  summary: { anchored: 1, pending: 0, changed: 1, deleted: 0 },
  typedData: { types: {}, primaryType: 'OwnerClaimApproval', domain: { name: 'BanguninAja', version: '1', chainId: 11155111 }, message: {} },
  approverAddress: ACCOUNT,
  approval: { termId: 't-1', token: 'tok', signer: '', approvedAt: null, stillSame: false },
  explorerUrl: 'https://sepolia.etherscan.io',
}

function mockClaim(claim: PublicClaim) {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ data: claim }), { status: 200, headers: { 'Content-Type': 'application/json' } })))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('jejak bukti blockchain', () => {
  it('mencatat sidik jari dengan transaksi 0 ETH ke alamat sendiri yang membawa data', async () => {
    const { wallet, calls } = fakeWallet({ eth_requestAccounts: [ACCOUNT], eth_sendTransaction: '0xhash' })

    await expect(sendCalldata(wallet, 11155111, '0x424e474eabcd')).resolves.toBe('0xhash')

    const send = calls.find((call) => call.method === 'eth_sendTransaction')
    expect(send?.params).toEqual([{ from: ACCOUNT, to: ACCOUNT, value: '0x0', data: '0x424e474eabcd' }])
    expect(calls.some((call) => call.method === 'wallet_switchEthereumChain')).toBe(true)
  })

  it('memanggil kontrak EvidenceRegistry bila alamatnya diberikan', async () => {
    const registry = '0x2c7536E3605D9C16a7a3D7b1898e529396a65c23'
    const { wallet, calls } = fakeWallet({ eth_requestAccounts: [ACCOUNT], eth_sendTransaction: '0xhash' })

    await sendCalldata(wallet, 11155111, '0xabcdef01', registry)

    const send = calls.find((call) => call.method === 'eth_sendTransaction')
    expect(send?.params).toEqual([{ from: ACCOUNT, to: registry, value: '0x0', data: '0xabcdef01' }])
  })

  it('meminta tanda tangan EIP-712 dari akun yang terhubung', async () => {
    const { wallet, calls } = fakeWallet({ eth_requestAccounts: [ACCOUNT], eth_signTypedData_v4: '0xsig' })

    await expect(signTypedData(wallet, CLAIM.typedData)).resolves.toBe('0xsig')

    const sign = calls.find((call) => call.method === 'eth_signTypedData_v4')
    expect(sign?.params?.[0]).toBe(ACCOUNT)
    expect(JSON.parse(String(sign?.params?.[1]))).toEqual(CLAIM.typedData)
  })

  it('halaman verifikasi menandai bukti yang berubah dan memperingatkan sebelum menyetujui', async () => {
    mockClaim(CLAIM)
    renderWithProviders(<Routes><Route path="/verifikasi/:token" element={<VerifyClaimPage />} /></Routes>, { route: '/verifikasi/tok' })

    expect(await screen.findByRole('heading', { level: 1, name: /Termin 2/ })).toBeInTheDocument()
    expect(screen.getByText('Berubah setelah dicatat')).toBeInTheDocument()
    expect(screen.getByText(/Ada bukti yang berubah atau dihapus/)).toBeInTheDocument()
  })

  it('halaman verifikasi menampilkan persetujuan yang sudah ditandatangani', async () => {
    mockClaim({ ...CLAIM, summary: { anchored: 2, pending: 0, changed: 0, deleted: 0 }, approval: { ...CLAIM.approval, signer: ACCOUNT, approvedAt: '2026-10-18T00:00:00Z', stillSame: true } })
    renderWithProviders(<Routes><Route path="/verifikasi/:token" element={<VerifyClaimPage />} /></Routes>, { route: '/verifikasi/tok' })

    expect(await screen.findByText(/Bukti proyek masih sama seperti saat disetujui/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Setujui dengan dompet/ })).not.toBeInTheDocument()
  })
})
