import { describe, expect, it } from 'vitest'

import {
  MAX_COMPARED,
  compareLocations,
  hasStoredRisk,
  searchSites,
  toggleSelection,
  type Dimension,
  type SavedLocationDetail,
} from '../models/location'

function dimension(id: string, code: string, sortOrder: number): Dimension {
  return { id, code, name: code, description: '', sortOrder, createdAt: '', updatedAt: '' }
}

const DIMENSIONS = [
  dimension('d-akses', 'aksesibilitas', 2),
  dimension('d-fisik', 'fisik_lingkungan', 1),
]

function site(overrides: Partial<SavedLocationDetail>): SavedLocationDetail {
  return {
    id: overrides.name ?? 'x',
    name: 'Lokasi',
    regionId: null,
    buildingProfileId: null,
    projectId: null,
    latitude: -6.2,
    longitude: 106.8,
    areaSqm: 0,
    landPricePerSqm: 0,
    score: 50,
    floodIndex: 0,
    earthquakeIndex: 0,
    note: '',
    riskFlags: null,
    savedAt: '2026-10-01T00:00:00Z',
    createdAt: '2026-10-01T00:00:00Z',
    updatedAt: '2026-10-01T00:00:00Z',
    dimensionScores: [],
    ...overrides,
  }
}

describe('compareLocations', () => {
  const cibubur = site({
    name: 'Cibubur',
    score: 82,
    landPricePerSqm: 4_000_000,
    floodIndex: 0.1,
    dimensionScores: [
      { dimensionId: 'd-akses', value: 90 },
      { dimensionId: 'd-fisik', value: 70 },
    ],
  })
  const bekasi = site({
    name: 'Bekasi',
    score: 64,
    landPricePerSqm: 3_000_000,
    floodIndex: 0.7,
    dimensionScores: [
      { dimensionId: 'd-akses', value: 60 },
      { dimensionId: 'd-fisik', value: 40 },
    ],
  })

  it('menandai kolom yang unggul per baris', () => {
    const table = compareLocations([cibubur, bekasi], DIMENSIONS)

    expect([...table.scoreBest]).toEqual([0])
    expect([...table.floodBest]).toEqual([0])
    expect([...table.landPriceBest]).toEqual([1])
    expect(table.dimensionRows.map((row) => row.dimension.code)).toEqual(['fisik_lingkungan', 'aksesibilitas'])
    expect(table.dimensionRows.every((row) => row.best.has(0))).toBe(true)
  })

  it('tidak ada yang unggul kalau semua nilainya sama', () => {
    const table = compareLocations([site({ name: 'A', score: 70 }), site({ name: 'B', score: 70 })], DIMENSIONS)

    expect(table.scoreBest.size).toBe(0)
    expect(table.summary).toContain('Semua lokasi punya skor keseluruhan sama (70)')
  })

  it('merangkum pemimpin skor, dimensi, dan risiko dalam satu kalimat', () => {
    const table = compareLocations([cibubur, bekasi], DIMENSIONS)

    expect(table.summary).toBe(
      'Cibubur punya skor keseluruhan tertinggi (82) dan unggul di 2 dari 2 dimensi; ' +
        'Bekasi punya risiko banjir tinggi.',
    )
  })

  it('memberi nomor pada nama lokasi yang kembar', () => {
    const table = compareLocations([site({ name: 'Titik' }), site({ name: 'Titik', score: 60 })], DIMENSIONS)

    expect(table.labels).toEqual(['Titik (1)', 'Titik (2)'])
  })

  it('memakai penanda risiko tersimpan bila ada, dan menaksir dari indeks bila tidak', () => {
    const stored = site({
      name: 'Tersimpan',
      floodIndex: 0.9,
      riskFlags: [{ code: 'rawan_longsor', severity: 'medium', message: '' }],
    })
    const estimated = site({
      name: 'Taksiran',
      earthquakeIndex: 0.5,
      dimensionScores: [{ dimensionId: 'd-fisik', value: 20 }],
    })
    const table = compareLocations([stored, estimated], DIMENSIONS)

    expect(hasStoredRisk(stored)).toBe(true)
    expect(hasStoredRisk(estimated)).toBe(false)
    expect(table.risks[0]).toEqual([{ key: 'rawan_longsor', label: 'tanah longsor', level: 'medium' }])
    expect(table.risks[1].map((flag) => [flag.key, flag.level])).toEqual([
      ['fisik', 'high'],
      ['gempa', 'medium'],
    ])
    expect(table.hasEstimatedRisk).toBe(true)
  })
})

describe('pilihan lokasi', () => {
  it('toggleSelection menambah di ujung, melepas, dan berhenti di batas', () => {
    expect(toggleSelection(['a'], 'b')).toEqual(['a', 'b'])
    expect(toggleSelection(['a', 'b'], 'a')).toEqual(['b'])
    const full = Array.from({ length: MAX_COMPARED }, (_, index) => `l${index}`)
    expect(toggleSelection(full, 'baru')).toEqual(full)
  })

  it('searchSites mencocokkan nama tanpa peduli huruf besar', () => {
    const sites = [site({ name: 'Griya Cibubur' }), site({ name: 'Bekasi Timur' })]
    expect(searchSites(sites, '  cibubur ').map((item) => item.name)).toEqual(['Griya Cibubur'])
    expect(searchSites(sites, '')).toHaveLength(2)
  })
})
