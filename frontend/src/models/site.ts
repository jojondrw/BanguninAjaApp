// Site Intelligence models.

export interface LatLng {
  latitude: number
  longitude: number
}

// Default map view: centered on Indonesia so the whole archipelago is visible.
export const INDONESIA_CENTER: LatLng = { latitude: -2.5, longitude: 118.0 }
export const INDONESIA_ZOOM = 4.2

// Building profile as returned by GET /api/scoring/building-profiles (camelCase envelope).
export interface BuildingProfile {
  id: string
  code: string
  name: string
}

// Request body for POST /api/site/evaluate (T1 contract §2.1, snake_case wire format).
export interface EvaluateRequest {
  project_id?: string
  latitude: number
  longitude: number
  building_profile_id: string
  name: string
}

// Response `data` object from POST /api/site/evaluate (T1 contract §2.2).
export interface EvaluateResponse {
  saved_location_id: string
  predictive: Predictive
  descriptive: Descriptive | null
}

export interface Predictive {
  overall_score: number
  dimension_scores: DimensionScore[]
  risk_flags: RiskFlag[]
}

export interface DimensionScore {
  dimension_code: string
  value: number
  explanation: string
}

export type RiskSeverity = 'info' | 'warning' | 'critical'

export interface RiskFlag {
  code: string
  severity: RiskSeverity
  message: string
}

// Reserved shape for the descriptive block (contract §2.3, owned by T18). Always null
// from the current backend, but typed so the panel can render it once T18 lands.
export interface Descriptive {
  regulasi: {
    kdb: number
    klb: number
    zona: string
    is_simulated: boolean
  } | null
  news: { title: string; url: string; source: string; published_at: string }[]
}

// Human-readable Indonesian labels for the five scored dimensions (contract §1.4).
export const DIMENSION_LABEL: Record<string, string> = {
  fisik_lingkungan: 'Fisik & Lingkungan',
  infrastruktur: 'Infrastruktur & Aksesibilitas',
  demografi_sosial: 'Demografi & Sosial',
  pasar_kompetisi: 'Pasar & Kompetisi',
  finansial_proyek: 'Finansial Proyek',
}

export const RISK_SEVERITY_TONE: Record<RiskSeverity, string> = {
  info: 'bg-slate-100 text-slate-700',
  warning: 'bg-amber-100 text-amber-800',
  critical: 'bg-red-100 text-red-800',
}
