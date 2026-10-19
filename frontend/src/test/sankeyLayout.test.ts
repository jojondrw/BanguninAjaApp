import { describe, expect, it } from 'vitest'

import type { LeadFlow } from '../models/sales'
import { layoutSankey } from '../views/sales/sankeyLayout'

const FLOWS: LeadFlow[] = [
  { source: 'Instagram', stage: 'new', total: 3 },
  { source: 'Instagram', stage: 'won', total: 1 },
  { source: 'Pameran', stage: 'negotiating', total: 4 },
  { source: 'Pameran', stage: 'cancelled', total: 2 },
]

describe('tata letak sankey prospek', () => {
  it('memakai skala yang sama di kedua kolom dan urutan tahap tetap', () => {
    const layout = layoutSankey(FLOWS, 200, 10)

    expect(layout.total).toBe(10)
    expect(layout.sources.map((node) => node.key)).toEqual(['Pameran', 'Instagram'])
    expect(layout.stages.map((node) => node.key)).toEqual(['new', 'negotiating', 'won', 'cancelled'])
    const leftHeight = layout.sources.reduce((sum, node) => sum + node.height, 0)
    const rightHeight = layout.stages.reduce((sum, node) => sum + node.height, 0)
    expect(leftHeight).toBeCloseTo(rightHeight)
  })

  it('menumpuk pita di dalam simpulnya tanpa saling tumpang tindih', () => {
    const layout = layoutSankey(FLOWS, 200, 10)
    const pameran = layout.sources.find((node) => node.key === 'Pameran')!
    const fromPameran = layout.links.filter((link) => link.source === 'Pameran').sort((a, b) => a.sourceTop - b.sourceTop)

    expect(fromPameran[0].sourceTop).toBeCloseTo(pameran.top)
    expect(fromPameran[1].sourceTop).toBeCloseTo(fromPameran[0].sourceTop + fromPameran[0].height)
    expect(fromPameran[1].sourceTop + fromPameran[1].height).toBeCloseTo(pameran.top + pameran.height)
  })

  it('tanpa prospek tidak menggambar apa pun', () => {
    expect(layoutSankey([], 200, 10).total).toBe(0)
  })
})
