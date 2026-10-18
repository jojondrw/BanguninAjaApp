import { screen } from '@testing-library/react'
import { Route, Routes } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { periodLabel, recentPeriods, type PublicReceipt, type PublicWageProof } from '../models/wageProof'
import { VerifyWagePage, WageReceiptPage } from '../views/verify/WageProofPages'
import { renderWithProviders } from './render'

const ANCHOR = { status: 'pending' as const, anchoredAt: null, txHash: '', txUrl: '' }

function respond(data: unknown) {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ data }), { status: 200, headers: { 'Content-Type': 'application/json' } })))
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('bukti upah minimum', () => {
  it('menyusun periode penggajian terbaru lebih dulu', () => {
    expect(recentPeriods(new Date(2026, 0, 15), 3)).toEqual(['2026-01', '2025-12', '2025-11'])
    expect(periodLabel('2026-09')).toBe('September 2026')
  })

  it('pemberi kerja melihat hasil bukti tanpa satu pun gaji', async () => {
    const proof: PublicWageProof = {
      projectCode: 'DEMO-RSIA', projectName: 'RSIA Cirebon', period: '2026-09', regionName: 'Kota Cirebon',
      minimumWage: 2_878_646, dailyMinimum: 115_146, wageSource: 'Kepgub Jabar 561/Kep.862-Kesra/2025', workerCount: 3,
      root: '0x' + '12'.repeat(32), verified: true, anchor: ANCHOR, proofHex: '0xabcd', verifyingKeyHex: '0xef01',
    }
    respond(proof)
    renderWithProviders(<Routes><Route path="/verifikasi-upah/:token" element={<VerifyWagePage />} /></Routes>, { route: '/verifikasi-upah/tok' })

    expect(await screen.findByText(/Terbukti: semua 3 pekerja dibayar paling sedikit UMK Kota Cirebon/)).toBeInTheDocument()
    expect(document.body.textContent).not.toMatch(/12\.000\.000/)
  })

  it('bukti yang tidak valid ditandai jelas', async () => {
    respond({ projectCode: 'X', projectName: 'Y', period: '2026-09', regionName: 'Kota Cirebon', minimumWage: 1, dailyMinimum: 1,
      wageSource: '', workerCount: 1, root: '0x00', verified: false, anchor: ANCHOR, proofHex: '0x', verifyingKeyHex: '0x' })
    renderWithProviders(<Routes><Route path="/verifikasi-upah/:token" element={<VerifyWagePage />} /></Routes>, { route: '/verifikasi-upah/tok' })

    expect(await screen.findByText('Bukti tidak valid. Hubungi pengembang.')).toBeInTheDocument()
  })

  it('pekerja melihat upahnya sendiri dan apakah ikut dihitung', async () => {
    const receipt: PublicReceipt = {
      name: 'Bambang Sutrisno', projectCode: 'DEMO-RSIA', projectName: 'RSIA Cirebon', period: '2026-09', pay: 12_000_000,
      days: 25, leafIndex: 0, salt: '0x01', path: [], root: '0x02', included: true, proofToken: 'p', anchor: ANCHOR,
    }
    respond(receipt)
    renderWithProviders(<Routes><Route path="/kuitansi-upah/:token" element={<WageReceiptPage />} /></Routes>, { route: '/kuitansi-upah/tok' })

    expect(await screen.findByText('Upah Anda ikut dihitung dalam bukti upah minimum proyek ini.')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Kuitansi upah Bambang Sutrisno' })).toBeInTheDocument()
  })
})
