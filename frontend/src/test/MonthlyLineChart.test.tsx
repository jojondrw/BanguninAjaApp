import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { MonthlyLineChart } from '../views/components/MonthlyLineChart'

const points = [
  { period: '2026-08', value: 1_000_000_000 },
  { period: '2026-09', value: 3_000_000_000 },
  { period: '2026-10', value: 2_000_000_000 },
]

describe('grafik garis per bulan', () => {
  it('memberi pembaca layar tabel nilai tiap bulan', () => {
    render(<MonthlyLineChart label="Kas keluar proyek 3 bulan terakhir" points={points} />)

    expect(screen.getByRole('img', { name: 'Kas keluar proyek 3 bulan terakhir' })).toBeInTheDocument()
    expect(screen.getAllByRole('row')).toHaveLength(points.length)
  })

  it('menggambar satu garis tren dan garis rata-rata', () => {
    const { container } = render(<MonthlyLineChart label="Kas keluar" points={points} />)

    expect(container.querySelector('path.stroke-navy-600')).not.toBeNull()
    expect(screen.getByText(/rata-rata/)).toHaveTextContent('rata-rata Rp 2 M')
  })
})
