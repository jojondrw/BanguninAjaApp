import type { SavedLocation } from './erpApi'

export interface SiteSummary {
  count: number
  best: SavedLocation | null
  latest: SavedLocation | null
  averageScore: number
}

export function rankSites(sites: SavedLocation[]): SavedLocation[] {
  return [...sites].sort((a, b) => b.score - a.score || a.name.localeCompare(b.name))
}

export function summarizeSites(sites: SavedLocation[]): SiteSummary {
  if (sites.length === 0) {
    return { count: 0, best: null, latest: null, averageScore: 0 }
  }

  const total = sites.reduce((sum, site) => sum + site.score, 0)
  const latest = sites.reduce((newest, site) => (site.savedAt > newest.savedAt ? site : newest))

  return {
    count: sites.length,
    best: rankSites(sites)[0],
    latest,
    averageScore: Math.round(total / sites.length),
  }
}

// Bentuk DTO mengikuti backend/internal/location/dto.go.
export interface DimensionScoreValue {
  dimensionId: string
  value: number
}

export interface SavedLocationDetail extends SavedLocation {
  dimensionScores: DimensionScoreValue[]
}

export interface Dimension {
  id: string
  code: string
  name: string
  description: string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface Region {
  id: string
  code: string
  name: string
  type: string
  parentId: string | null
  createdAt: string
  updatedAt: string
}

export interface Comparison {
  id: string
  name: string
  locationCount: number
  createdAt: string
  updatedAt: string
}

export interface ComparisonItem {
  savedLocationId: string
  sortOrder: number
  name: string
  score: number
  areaSqm: number
  landPricePerSqm: number
  floodIndex: number
  earthquakeIndex: number
  dimensionScores: DimensionScoreValue[]
}

export interface ComparisonDetail {
  id: string
  name: string
  items: ComparisonItem[]
  createdAt: string
  updatedAt: string
}

export interface ComparisonRequest {
  name: string
  savedLocationIds: string[]
}

export interface ComparisonFilter {
  page?: number
  pageSize?: number
}

// Backend menerima 2 sampai 5 lokasi. Tampilan dibatasi 4 supaya kolomnya
// masih terbaca berdampingan di layar laptop.
export const MIN_COMPARED = 2
export const MAX_COMPARED = 4
export const COMPARISON_NAME_MAX = 160

export function searchSites(sites: SavedLocation[], keyword: string): SavedLocation[] {
  const needle = keyword.trim().toLowerCase()
  return needle === '' ? sites : sites.filter((site) => site.name.toLowerCase().includes(needle))
}

// Pilihan baru ditambahkan di ujung supaya urutan kolom mengikuti urutan klik.
export function toggleSelection(selected: string[], id: string): string[] {
  if (selected.includes(id)) {
    return selected.filter((current) => current !== id)
  }
  return selected.length >= MAX_COMPARED ? selected : [...selected, id]
}

export function comparisonLocationIds(detail: ComparisonDetail): string[] {
  return [...detail.items]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((item) => item.savedLocationId)
}

export function isSameSelection(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((id, index) => id === b[index])
}

export type RiskLevel = 'high' | 'medium'

export interface RiskFlag {
  key: string
  label: string
  level: RiskLevel
}

export const RISK_LEVEL_LABEL: Record<RiskLevel, string> = {
  high: 'tinggi',
  medium: 'sedang',
}

// Ambang indeks bahaya sama dengan layanan scoring: 0,4 ke atas sedang,
// 0,66 ke atas tinggi.
const HAZARD_MEDIUM = 0.4
const HAZARD_HIGH = 0.66

// Skor Fisik & Lingkungan adalah kebalikan persentil indeks multi-bahaya
// InaRISK dan kemiringan lahan, jadi skor rendah berarti lahan relatif rawan.
// Di bawah 34 berarti termasuk sepertiga paling rawan.
export const PHYSICAL_RISK_BELOW: Record<RiskLevel, number> = { high: 34, medium: 50 }
const PHYSICAL_DIMENSION_CODE = 'fisik_lingkungan'

function hazardLevel(index: number): RiskLevel | null {
  if (index >= HAZARD_HIGH) {
    return 'high'
  }
  return index >= HAZARD_MEDIUM ? 'medium' : null
}

function physicalLevel(score: number | undefined): RiskLevel | null {
  if (score === undefined) {
    return null
  }
  if (score < PHYSICAL_RISK_BELOW.high) {
    return 'high'
  }
  return score < PHYSICAL_RISK_BELOW.medium ? 'medium' : null
}

function riskFlags(location: SavedLocationDetail, physicalDimensionId: string | undefined): RiskFlag[] {
  const physicalScore = location.dimensionScores.find((score) => score.dimensionId === physicalDimensionId)?.value
  const candidates: { key: string; label: string; level: RiskLevel | null }[] = [
    { key: 'banjir', label: 'banjir', level: hazardLevel(location.floodIndex) },
    { key: 'gempa', label: 'gempa bumi', level: hazardLevel(location.earthquakeIndex) },
    { key: 'fisik', label: 'fisik & lingkungan', level: physicalLevel(physicalScore) },
  ]

  return candidates
    .flatMap((flag) => (flag.level ? [{ ...flag, level: flag.level }] : []))
    .sort((a, b) => (a.level === b.level ? 0 : a.level === 'high' ? -1 : 1))
}

type Better = 'higher' | 'lower'

// Indeks kolom yang nilainya paling unggul. Kosong kalau kurang dari dua
// kolom punya nilai atau semua nilainya sama, karena tidak ada yang unggul.
function bestIndexes(values: (number | null)[], better: Better): Set<number> {
  const present = values.filter((value): value is number => value !== null)
  if (present.length < 2 || present.every((value) => value === present[0])) {
    return new Set()
  }

  const target = better === 'higher' ? Math.max(...present) : Math.min(...present)
  return new Set(values.flatMap((value, index) => (value === target ? [index] : [])))
}

function positiveOrNull(value: number): number | null {
  return value > 0 ? value : null
}

export interface DimensionRow {
  dimension: Dimension
  values: (number | null)[]
  best: Set<number>
}

export interface ComparisonTable {
  labels: string[]
  scoreBest: Set<number>
  dimensionRows: DimensionRow[]
  risks: RiskFlag[][]
  floodBest: Set<number>
  earthquakeBest: Set<number>
  landPriceBest: Set<number>
  hasRegion: boolean
  hasArea: boolean
  hasLandPrice: boolean
  hasHazardIndex: boolean
  hasNote: boolean
  summary: string
}

// Nama kembar diberi nomor urut supaya kolom dan kalimat ringkasan tidak
// tertukar, misalnya dua evaluasi di titik yang sama.
function labelLocations(locations: SavedLocationDetail[]): string[] {
  const totals = new Map<string, number>()
  for (const location of locations) {
    totals.set(location.name, (totals.get(location.name) ?? 0) + 1)
  }

  const seen = new Map<string, number>()
  return locations.map((location) => {
    if ((totals.get(location.name) ?? 0) < 2) {
      return location.name
    }
    const order = (seen.get(location.name) ?? 0) + 1
    seen.set(location.name, order)
    return `${location.name} (${order})`
  })
}

function joinWords(words: string[]): string {
  if (words.length <= 1) {
    return words.join('')
  }
  return `${words.slice(0, -1).join(', ')} dan ${words[words.length - 1]}`
}

function pick(labels: string[], indexes: Iterable<number>): string[] {
  return [...indexes].map((index) => labels[index])
}

function scoreClause(labels: string[], scores: number[], best: Set<number>): string {
  if (best.size === 0) {
    return `semua lokasi punya skor keseluruhan sama (${scores[0]})`
  }
  const top = Math.max(...scores)
  if (best.size === 1) {
    return `${pick(labels, best)[0]} punya skor keseluruhan tertinggi (${top})`
  }
  return `${joinWords(pick(labels, best))} sama-sama punya skor keseluruhan tertinggi (${top})`
}

function dimensionLeaders(rows: DimensionRow[], count: number): { leaders: Set<number>; wins: number } {
  const wins = Array.from({ length: count }, (_, index) => rows.filter((row) => row.best.has(index)).length)
  const most = Math.max(0, ...wins)
  if (most === 0) {
    return { leaders: new Set(), wins: 0 }
  }
  return { leaders: new Set(wins.flatMap((value, index) => (value === most ? [index] : []))), wins: most }
}

// Kalau pemimpin skor keseluruhan juga paling banyak unggul di dimensi,
// keduanya digabung jadi satu klausa.
function leadClauses(labels: string[], scores: number[], scoreBest: Set<number>, rows: DimensionRow[]): string[] {
  const overall = scoreClause(labels, scores, scoreBest)
  const { leaders, wins } = dimensionLeaders(rows, labels.length)
  if (leaders.size === 0) {
    return [overall]
  }

  const each = leaders.size > 1 ? 'masing-masing ' : ''
  const tally = `${each}unggul di ${wins} dari ${rows.length} dimensi`
  const isSameLeader = leaders.size === scoreBest.size && [...leaders].every((index) => scoreBest.has(index))
  if (isSameLeader) {
    return [`${overall} dan ${tally}`]
  }
  return [overall, `${joinWords(pick(labels, leaders))} ${tally}`]
}

function subject(labels: string[], indexes: number[]): string {
  return labels.length > 1 && indexes.length === labels.length ? 'semua lokasi' : joinWords(pick(labels, indexes))
}

// Hanya tingkat risiko paling berat yang disebut supaya kalimatnya tetap
// pendek; rincian lengkapnya ada di baris penanda risiko. Lokasi dengan risiko
// yang sama digabung dalam satu klausa.
function riskClause(labels: string[], risks: RiskFlag[][]): string {
  for (const level of ['high', 'medium'] as const) {
    const groups = new Map<string, number[]>()
    risks.forEach((flags, index) => {
      const matching = flags.filter((flag) => flag.level === level).map((flag) => flag.label)
      if (matching.length > 0) {
        const phrase = `risiko ${joinWords(matching)} ${RISK_LEVEL_LABEL[level]}`
        groups.set(phrase, [...(groups.get(phrase) ?? []), index])
      }
    })
    if (groups.size > 0) {
      return [...groups].map(([phrase, indexes]) => `${subject(labels, indexes)} punya ${phrase}`).join('; ')
    }
  }
  return 'tidak ada penanda risiko tinggi atau sedang'
}

function capitalize(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

function summarize(labels: string[], scores: number[], scoreBest: Set<number>, rows: DimensionRow[], risks: RiskFlag[][]): string {
  const clauses = [...leadClauses(labels, scores, scoreBest, rows), riskClause(labels, risks)]
  return `${capitalize(clauses.join('; '))}.`
}

export function compareLocations(locations: SavedLocationDetail[], dimensions: Dimension[]): ComparisonTable {
  const labels = labelLocations(locations)
  const scores = locations.map((location) => location.score)
  const scoreBest = bestIndexes(scores, 'higher')

  const dimensionRows = [...dimensions]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((dimension) => {
      const values = locations.map(
        (location) => location.dimensionScores.find((score) => score.dimensionId === dimension.id)?.value ?? null,
      )
      return { dimension, values, best: bestIndexes(values, 'higher') }
    })
    .filter((row) => row.values.some((value) => value !== null))

  const physicalId = dimensions.find((dimension) => dimension.code === PHYSICAL_DIMENSION_CODE)?.id
  const risks = locations.map((location) => riskFlags(location, physicalId))

  return {
    labels,
    scoreBest,
    dimensionRows,
    risks,
    floodBest: bestIndexes(locations.map((location) => location.floodIndex), 'lower'),
    earthquakeBest: bestIndexes(locations.map((location) => location.earthquakeIndex), 'lower'),
    landPriceBest: bestIndexes(locations.map((location) => positiveOrNull(location.landPricePerSqm)), 'lower'),
    hasRegion: locations.some((location) => location.regionId !== null),
    hasArea: locations.some((location) => location.areaSqm > 0),
    hasLandPrice: locations.some((location) => location.landPricePerSqm > 0),
    hasHazardIndex: locations.some((location) => location.floodIndex > 0 || location.earthquakeIndex > 0),
    hasNote: locations.some((location) => location.note.trim() !== ''),
    summary: summarize(labels, scores, scoreBest, dimensionRows, risks),
  }
}
