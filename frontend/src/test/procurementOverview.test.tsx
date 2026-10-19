import { fireEvent, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ProcurementOverview as Overview } from '../models/procurement'
import { ProcurementOverview } from '../views/procurement/ProcurementOverview'
import { renderWithProviders } from './render'

const OVERVIEW: Overview = {
  requests: { submitted: 2, approved: 6, draft: 1 },
  orders: { count: 10, openCount: 4, openValue: 900_000_000, totalValue: 2_000_000_000, byStatus: { draft: 1, sent: 2, partially_received: 1, completed: 6, cancelled: 0 } },
  lateOrders: 3,
  lateValue: 400_000_000,
  recentReceipts: 8,
  damagedRecent: 2,
  receiptDays: 90,
  vendorRatings: { good: 4, poor: 1 },
  largestVendors: [{ name: 'PT Beton Readymix', rating: 'fair', orders: 5, value: 700_000_000 }],
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ringkasan pengadaan', () => {
  it('menampilkan alur, hal mendesak, dan vendor lalu membuka tab tersaring', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ data: OVERVIEW }), { headers: { 'Content-Type': 'application/json' } })))
    const drill = { openRequests: vi.fn(), openOrders: vi.fn(), openReceipts: vi.fn(), openVendors: vi.fn() }

    renderWithProviders(<ProcurementOverview drill={drill} />)

    expect(await screen.findByText('PT Beton Readymix')).toBeInTheDocument()
    expect(screen.getByText('25% dari 8 penerimaan 90 hari terakhir')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Disetujui/ }))
    expect(drill.openRequests).toHaveBeenCalledWith('approved')
    fireEvent.click(screen.getByRole('button', { name: /Selesai/ }))
    expect(drill.openOrders).toHaveBeenCalledWith('completed')
  })
})
