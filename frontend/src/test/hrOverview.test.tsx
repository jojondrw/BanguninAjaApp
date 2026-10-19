import { fireEvent, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { HrOverview as Overview } from '../models/hr'
import { HrOverview } from '../views/hr/HrOverview'
import { renderWithProviders } from './render'

const OVERVIEW: Overview = {
  headcount: 12,
  byType: [{ employmentType: 'permanent', total: 7 }, { employmentType: 'daily', total: 5 }],
  byProject: [{ projectId: null, total: 4 }],
  days: [{ date: '2026-10-19', present: 8, sick: 1, permitted: 0, absent: 1 }],
  payrolls: [{ period: '2026-10', netPay: 100_000_000, unpaid: 100_000_000, slips: 12 }],
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ringkasan SDM', () => {
  it('menghitung kehadiran dan gaji belum dibayar lalu membuka tab rincian', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ data: OVERVIEW }), { headers: { 'Content-Type': 'application/json' } })))
    const drill = { openEmployees: vi.fn(), openAttendance: vi.fn(), openPayroll: vi.fn() }

    renderWithProviders(<HrOverview drill={drill} />)

    expect(await screen.findByText('Kehadiran 80% dari absensi yang tercatat')).toBeInTheDocument()
    expect(screen.getByText('Kantor pusat')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Penggajian/ }))
    expect(drill.openPayroll).toHaveBeenCalled()
  })
})
