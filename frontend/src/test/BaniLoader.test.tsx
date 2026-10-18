import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { BaniLoop, BaniScreen } from '../views/components/BaniLoader'
import { Loading } from '../views/components/Data'

describe('animasi loading Bani', () => {
  it('layar penuh memakai animasi membangun rumah dengan gambar diam untuk pengguna yang mengurangi gerak', () => {
    const { container } = render(<BaniScreen label="Memeriksa sesi..." />)

    expect(screen.getByRole('status')).toHaveTextContent('Memeriksa sesi...')
    expect(container.querySelector('img')).toHaveAttribute('src', '/images/bani/build-house.webp')
    expect(container.querySelector('source')).toHaveAttribute('media', '(prefers-reduced-motion: reduce)')
    expect(container.querySelector('source')).toHaveAttribute('srcset', '/images/bani/build-house-still.webp')
  })

  it('kartu yang sedang memuat menampilkan salah satu dari empat animasi kecil', () => {
    const { container } = render(<Loading />)

    expect(screen.getByRole('status')).toHaveTextContent('Mengambil data...')
    expect(container.querySelector('img')?.getAttribute('src')).toMatch(/\/images\/bani\/(bricks|wheelbarrow|trowel|blueprint)\.webp$/)
  })

  it('animasi yang sama dipakai selama komponen tetap terpasang', () => {
    const { container, rerender } = render(<BaniLoop label="Memuat" />)
    const first = container.querySelector('img')?.getAttribute('src')

    rerender(<BaniLoop label="Memuat lagi" />)

    expect(container.querySelector('img')?.getAttribute('src')).toBe(first)
  })
})
