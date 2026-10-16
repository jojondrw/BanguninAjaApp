import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ScoreCell } from '../views/location/parts'

describe('skor cadangan di tabel perbandingan', () => {
  it('menandai skor simulasi supaya tidak terbaca sebagai hasil data GIS', () => {
    render(<ScoreCell value={72} isBest={false} isSimulated />)

    expect(screen.getByText('72')).toBeInTheDocument()
    expect(screen.getByText('Skor simulasi')).toBeInTheDocument()
  })

  it('tidak memberi tanda pada skor dari model penilaian', () => {
    render(<ScoreCell value={72} isBest={false} />)

    expect(screen.queryByText('Skor simulasi')).not.toBeInTheDocument()
  })
})
