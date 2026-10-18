import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BaniScreen } from '../views/components/BaniLoader'
import { Loading } from '../views/components/Data'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('animasi loading Bani', () => {
  it('layar penuh memutar video membangun rumah di atas latar halaman', () => {
    const { container } = render(<BaniScreen label="Memeriksa sesi..." />)
    const video = container.querySelector('video')

    expect(screen.getByRole('status')).toHaveTextContent('Memeriksa sesi...')
    expect(video).toHaveAttribute('src', '/images/bani/build-house.mp4')
    expect(video).toHaveAttribute('poster', '/images/bani/build-house-still.webp')
    expect(video).toHaveProperty('muted', true)
    expect(video).toHaveAttribute('loop')
  })

  it('kartu yang sedang memuat memakai versi berlatar putih dalam ukuran kecil', () => {
    const { container } = render(<Loading />)

    expect(screen.getByRole('status')).toHaveTextContent('Mengambil data...')
    expect(container.querySelector('video')).toHaveAttribute('src', '/images/bani/build-house-card.mp4')
    expect(container.querySelector('video')).toHaveStyle({ height: '56px' })
  })

  it('pengguna yang mengurangi gerak melihat gambar diam, bukan video', () => {
    vi.stubGlobal('matchMedia', (query: string) => ({ matches: query.includes('reduce'), media: query }))
    const { container } = render(<BaniScreen label="Memuat" />)

    expect(container.querySelector('video')).toBeNull()
    expect(container.querySelector('img')).toHaveAttribute('src', '/images/bani/build-house-still.webp')
  })
})
