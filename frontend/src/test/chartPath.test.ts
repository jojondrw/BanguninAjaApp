import { describe, expect, it } from 'vitest'

import { areaPath, barPath, smoothPath, type Point } from '../views/components/chartPath'

function controlYs(path: string): number[] {
  return [...path.matchAll(/C([\d.-]+),([\d.-]+) ([\d.-]+),([\d.-]+) /g)].flatMap((match) => [
    Number(match[2]),
    Number(match[4]),
  ])
}

describe('kurva halus untuk grafik', () => {
  it('melewati setiap titik data', () => {
    const points: Point[] = [[0, 100], [10, 80], [20, 30], [30, 20]]
    const path = smoothPath(points)
    expect(path.startsWith('M0,100')).toBe(true)
    for (const [x, y] of points.slice(1)) {
      expect(path).toContain(` ${x},${y}`)
    }
  })

  it('tidak melampaui titik data pada kurva yang terus naik', () => {
    const points: Point[] = [[0, 200], [10, 200], [20, 120], [30, 20], [40, 0], [50, 0]]
    const ys = controlYs(smoothPath(points))
    expect(Math.min(...ys)).toBeGreaterThanOrEqual(0)
    expect(Math.max(...ys)).toBeLessThanOrEqual(200)
  })

  it('menangani satu titik dan daftar kosong', () => {
    expect(smoothPath([])).toBe('')
    expect(smoothPath([[5, 6]])).toBe('M5,6')
  })

  it('menutup area sampai garis dasar', () => {
    expect(areaPath([[0, 10], [20, 5]], 50).endsWith('L20,50 L0,50 Z')).toBe(true)
    expect(areaPath([[0, 10]], 50)).toBe('')
  })

  it('membulatkan sedikit hanya sudut ujung luar batang', () => {
    expect(barPath(0, 10, 16, 40, 'up')).toBe('M0,50 L0,13 A3,3 0 0 1 3,10 L13,10 A3,3 0 0 1 16,13 L16,50 Z')
    expect(barPath(0, 0, 16, 2, 'down')).toBe('M0,0 L0,0 A2,2 0 0 0 2,2 L14,2 A2,2 0 0 0 16,0 L16,0 Z')
    expect(barPath(0, 0, 16, 0, 'up')).toBe('')
  })
})
