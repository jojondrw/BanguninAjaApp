import { fireEvent, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { BalanceSummary } from '../models/billing'
import { BillingOverview } from '../views/billing/BillingOverview'
import { renderWithProviders } from './render'

function side(outstanding: number, overdue: number, name: string): BalanceSummary {
  return {
    outstanding,
    overdue,
    openCount: 4,
    buckets: [
      { bucket: 'current', count: 2, amount: outstanding - overdue },
      { bucket: 'days_30', count: 1, amount: overdue },
      { bucket: 'days_60', count: 0, amount: 0 },
      { bucket: 'days_90', count: 0, amount: 0 },
      { bucket: 'days_over_90', count: 1, amount: 0 },
    ],
    largest: [{ name, count: 2, amount: outstanding }],
  }
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ringkasan tagihan', () => {
  it('menampilkan sisa, terlambat, dan pihak terbesar lalu membuka tab tersaring', async () => {
    const summary = { receivables: side(1_000_000_000, 400_000_000, 'PT Hotelindo Makassar'), payables: side(500_000_000, 0, 'CV Sumber') }
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ data: summary }), { headers: { 'Content-Type': 'application/json' } })))
    const drill = { openReceivables: vi.fn(), openPayables: vi.fn() }

    renderWithProviders(<BillingOverview drill={drill} />)

    expect(await screen.findByText('PT Hotelindo Makassar')).toBeInTheDocument()
    expect(screen.getByText('40% dari sisa')).toBeInTheDocument()
    fireEvent.click(screen.getAllByRole('button', { name: /Terlambat/ })[0])
    expect(drill.openReceivables).toHaveBeenCalledWith('overdue')
  })
})
