import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { BaniScreen } from '../views/components/BaniLoader'
import { Loading } from '../views/components/Data'

describe('animasi loading Bani', () => {
  it('layar penuh menampilkan Bani dan rumah yang dibangun bata demi bata', () => {
    const { container } = render(<BaniScreen label="Memeriksa sesi..." />)

    expect(screen.getByRole('status')).toHaveTextContent('Memeriksa sesi...')
    expect(container.querySelector('img')).toHaveAttribute('src', '/images/bani/bani-stand.webp')
    const parts = [...container.querySelectorAll('svg [style*="animation-delay"]')]
    expect(parts).toHaveLength(15)
    const delays = parts.map((part) => Number.parseInt((part as SVGElement).style.animationDelay, 10))
    expect(delays).toEqual([...delays].sort((first, second) => first - second))
  })

  it('kartu yang sedang memuat memakai adegan yang sama dalam ukuran kecil', () => {
    const { container } = render(<Loading />)

    expect(screen.getByRole('status')).toHaveTextContent('Mengambil data...')
    expect(container.querySelector('img')).toHaveStyle({ height: '48px' })
    expect(container.querySelector('svg')).toHaveAttribute('width', '40')
  })
})
