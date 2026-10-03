import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ErrorBoundary } from '../views/components/ErrorBoundary'

function Broken({ message }: { message: string }): never {
  throw new Error(message)
}

describe('ErrorBoundary', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('menampilkan anaknya kalau tidak ada galat', () => {
    render(
      <ErrorBoundary resetKey="/">
        <p>Halaman aman</p>
      </ErrorBoundary>,
    )
    expect(screen.getByText('Halaman aman')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('menampilkan tampilan pemulihan saat anaknya galat', () => {
    render(
      <ErrorBoundary resetKey="/">
        <Broken message="undefined is not a function" />
      </ErrorBoundary>,
    )

    const alert = screen.getByRole('alert')
    expect(alert).toHaveTextContent('Halaman ini gagal ditampilkan')
    expect(screen.getByRole('button', { name: 'Muat ulang' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ke Ringkasan' })).toHaveAttribute('href', '/')
    expect(console.error).toHaveBeenCalled()
  })

  it('menyarankan muat ulang saat berkas aplikasi sudah berganti versi', () => {
    render(
      <ErrorBoundary resetKey="/">
        <Broken message="Failed to fetch dynamically imported module: /assets/Sales.js" />
      </ErrorBoundary>,
    )
    expect(screen.getByRole('alert')).toHaveTextContent('Versi aplikasi baru tersedia')
  })

  it('pulih sendiri saat alamat berganti', () => {
    const { rerender } = render(
      <ErrorBoundary resetKey="/penjualan">
        <Broken message="rusak" />
      </ErrorBoundary>,
    )
    expect(screen.getByRole('alert')).toBeInTheDocument()

    rerender(
      <ErrorBoundary resetKey="/sdm">
        <p>Halaman SDM</p>
      </ErrorBoundary>,
    )
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByText('Halaman SDM')).toBeInTheDocument()
  })
})
