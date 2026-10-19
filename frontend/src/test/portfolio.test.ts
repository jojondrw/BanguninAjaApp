import { describe, expect, it } from 'vitest'

import { scheduleGaps, type PortfolioProject } from '../models/project'

function project(id: string, progress: number, planned: number | null): PortfolioProject {
  return { id, name: id, type: 'Perumahan', status: planned === null ? 'completed' : 'ongoing', progress, planned, contractValue: 0 }
}

describe('jadwal portofolio', () => {
  it('mengurutkan proyek paling tertinggal dulu dan melewati yang tanpa rencana', () => {
    const gaps = scheduleGaps([project('a', 50, 55), project('b', 30, 60), project('c', 100, null), project('d', 70, 60)])

    expect(gaps.map((item) => item.id)).toEqual(['b', 'a', 'd'])
    expect(gaps[0].gap).toBe(-30)
    expect(gaps[2].gap).toBe(10)
  })
})
