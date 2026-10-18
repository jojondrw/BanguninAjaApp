import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { LandingPage } from '../views/LandingPage'
import { renderWithProviders } from './render'

describe('halaman depan', () => {
  it('mengarahkan pengunjung ke daftar dan masuk', () => {
    renderWithProviders(<LandingPage />)

    expect(screen.getByRole('heading', { level: 1, name: 'Dari memilih lahan sampai menyerahkan kunci.' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Mulai sekarang/ })).toHaveAttribute('href', '/daftar')
    expect(screen.getByRole('link', { name: 'Masuk ke akun' })).toHaveAttribute('href', '/masuk')
  })

  it('menampilkan enam tahap alur bisnis dan sumber data lokasi', () => {
    renderWithProviders(<LandingPage />)

    const stages = screen.getAllByRole('heading', { level: 3 }).map((heading) => heading.textContent)
    expect(stages).toEqual(['Lokasi', 'Proyek', 'Pengadaan', 'Inventaris', 'SDM', 'Penjualan'])
    expect(screen.getByRole('list', { name: 'Sumber data' })).toHaveTextContent('BNPB InaRISK')
    expect(document.title).toContain('BanguninAja')
  })
})
