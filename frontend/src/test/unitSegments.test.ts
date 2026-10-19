import { describe, expect, it } from 'vitest'

import { unitBreakdown } from '../views/sales/unitSegments'

describe('rincian unit per proyek', () => {
  it('menyebut hanya status yang punya unit, terjual dulu', () => {
    expect(
      unitBreakdown({
        projectId: 'p-1',
        total: 10,
        available: 4,
        reserved: 2,
        sold: 3,
        onHold: 1,
        stockValue: 0,
        soldValue: 0,
      }),
    ).toBe('3 terjual, 2 dipesan, 4 tersedia, 1 ditahan')
    expect(
      unitBreakdown({ projectId: 'p-2', total: 5, available: 0, reserved: 0, sold: 0, onHold: 5, stockValue: 0, soldValue: 0 }),
    ).toBe('5 ditahan')
  })
})
