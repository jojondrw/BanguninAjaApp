export type MapLayerId = 'bahaya' | 'banjir' | 'gempabumi' | 'longsor' | 'penduduk' | 'harga_tanah'

export interface MapLayerDefinition {
  id: MapLayerId
  label: string
  description: string
  minZoom: number
  gradient: string
  ticks: [string, string, string]
  attribution: string
}

const HAZARD_GRADIENT = 'linear-gradient(90deg, rgb(46 157 85 / 0.35), #f5c518, #f08a24, #d7263d)'
const HAZARD_TICKS: [string, string, string] = ['Rendah', 'Sedang', 'Tinggi']
const INARISK = 'Bahaya: <a href="https://inarisk.bnpb.go.id">InaRISK BNPB</a>'

// Skala warna di sini harus sama dengan Ramp di scoring/banguninaja_scoring/tiles.py.
export const MAP_LAYERS: MapLayerDefinition[] = [
  {
    id: 'bahaya',
    label: 'Bahaya bencana gabungan',
    description: 'Indeks multi-bahaya InaRISK (banjir, gempa, longsor, tsunami, dan lainnya)',
    minZoom: 7,
    gradient: HAZARD_GRADIENT,
    ticks: HAZARD_TICKS,
    attribution: INARISK,
  },
  {
    id: 'banjir',
    label: 'Bahaya banjir',
    description: 'Indeks bahaya banjir InaRISK',
    minZoom: 7,
    gradient: HAZARD_GRADIENT,
    ticks: HAZARD_TICKS,
    attribution: INARISK,
  },
  {
    id: 'gempabumi',
    label: 'Bahaya gempa bumi',
    description: 'Indeks bahaya gempa bumi InaRISK',
    minZoom: 7,
    gradient: HAZARD_GRADIENT,
    ticks: HAZARD_TICKS,
    attribution: INARISK,
  },
  {
    id: 'longsor',
    label: 'Bahaya tanah longsor',
    description: 'Indeks bahaya tanah longsor InaRISK',
    minZoom: 7,
    gradient: HAZARD_GRADIENT,
    ticks: HAZARD_TICKS,
    attribution: INARISK,
  },
  {
    id: 'penduduk',
    label: 'Kepadatan penduduk',
    description: 'Jiwa per km² dari WorldPop 2020',
    minZoom: 7,
    gradient: 'linear-gradient(90deg, rgb(196 181 253 / 0.45), #8b5cf6, #5b21b6, #2e1065)',
    ticks: ['100', '1.000', '20.000+ jiwa/km²'],
    attribution: 'Penduduk: <a href="https://www.worldpop.org">WorldPop</a>',
  },
  {
    id: 'harga_tanah',
    label: 'Harga tanah (ZNT)',
    description: 'Zona nilai tanah ATR/BPN, rupiah per m²',
    minZoom: 10,
    gradient: 'linear-gradient(90deg, #fde68a, #f59e0b, #b45309, #7c2d12)',
    ticks: ['Rp100 rb', 'Rp5 jt', 'Rp60 jt+/m²'],
    attribution: 'Harga tanah: ATR/BPN',
  },
]

export function findMapLayer(id: MapLayerId | null): MapLayerDefinition | null {
  return MAP_LAYERS.find((layer) => layer.id === id) ?? null
}
