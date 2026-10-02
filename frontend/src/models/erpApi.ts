import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  Budget,
  CashFlow,
  CashFlowFilter,
  CashTransaction,
  CashTransactionFilter,
  CashTransactionRequest,
} from './finance'
import type { Account, UnitOfMeasure } from './master'
import type {
  BudgetItemPage,
  Permit,
  PermitRequest,
  PhaseRequest,
  Project,
  ProjectFilter,
  ProjectPhase,
  ProjectRequest,
} from './project'

const MAX_PAGE_SIZE = 100

export const projectApi = {
  list: (filter: ProjectFilter = {}) =>
    request<Page<Project>>(`/projects${toQueryString({ ...filter })}`),

  get: (id: string) => request<Project>(`/projects/${id}`),

  create: (body: ProjectRequest) => request<Project>('/projects', { method: 'POST', body }),

  phases: (id: string) => request<ProjectPhase[]>(`/projects/${id}/phases`),

  createPhase: (id: string, body: PhaseRequest) =>
    request<ProjectPhase>(`/projects/${id}/phases`, { method: 'POST', body }),

  permits: (id: string) => request<Permit[]>(`/projects/${id}/permits`),

  createPermit: (id: string, body: PermitRequest) =>
    request<Permit>(`/projects/${id}/permits`, { method: 'POST', body }),

  budgetItems: (id: string) =>
    request<BudgetItemPage>(`/projects/${id}/budget-items${toQueryString({ pageSize: MAX_PAGE_SIZE })}`),
}

export const financeApi = {
  budgets: (projectId?: string) =>
    request<Page<Budget>>(`/finance/budgets${toQueryString({ projectId, pageSize: 100 })}`),

  cashFlow: (filter: CashFlowFilter = {}) =>
    request<CashFlow>(`/finance/cash-flow${toQueryString({ ...filter })}`),

  cashTransactions: (filter: CashTransactionFilter = {}) =>
    request<Page<CashTransaction>>(`/finance/cash-transactions${toQueryString({ ...filter })}`),

  createCashTransaction: (body: CashTransactionRequest) =>
    request<CashTransaction>('/finance/cash-transactions', { method: 'POST', body }),
}

export const masterApi = {
  accounts: () =>
    request<Page<Account>>(`/master/accounts${toQueryString({ pageSize: MAX_PAGE_SIZE })}`),

  unitsOfMeasure: () =>
    request<Page<UnitOfMeasure>>(`/master/units-of-measure${toQueryString({ pageSize: MAX_PAGE_SIZE })}`),
}

export interface SiteEvaluateRequest {
  project_id?: string
  latitude: number
  longitude: number
  building_profile_id: string
  name: string
}

export interface SiteDimensionScore {
  dimension_code: string
  value: number
  explanation: string
}

export interface SiteRiskFlag {
  code: string
  severity: string
  message: string
}

export interface SiteRegulation {
  kdb: number
  klb: number
  zona: string
  is_simulated: boolean
}

export interface SiteNews {
  title: string
  url: string
  source: string
  published_at: string
}

export interface SiteEvaluateResponse {
  saved_location_id: string
  predictive: {
    overall_score: number
    dimension_scores: SiteDimensionScore[]
    risk_flags: SiteRiskFlag[]
  }
  descriptive: {
    regulasi: SiteRegulation | null
    news: SiteNews[]
  } | null
}

export interface BuildingProfile {
  id: string
  code: string
  name: string
  createdAt: string
  updatedAt: string
}

export const scoringApi = {
  buildingProfiles: () =>
    request<BuildingProfile[]>('/scoring/building-profiles'),
}

export const siteApi = {
  evaluate: (body: SiteEvaluateRequest) =>
    request<SiteEvaluateResponse>('/site/evaluate', {
      method: 'POST',
      body,
    }),
}

export interface StoredRiskFlag {
  code: string
  severity: 'high' | 'medium'
  message: string
}

// riskFlags bernilai null untuk lokasi yang dievaluasi sebelum penanda risiko
// ikut disimpan, atau yang dinilai skor simulasi.
export interface SavedLocation {
  id: string
  name: string
  regionId: string | null
  buildingProfileId: string | null
  projectId: string | null
  latitude: number
  longitude: number
  areaSqm: number
  landPricePerSqm: number
  score: number
  floodIndex: number
  earthquakeIndex: number
  note: string
  riskFlags: StoredRiskFlag[] | null
  savedAt: string
  createdAt: string
  updatedAt: string
}

export interface SavedLocationFilter {
  projectId?: string
  page?: number
  pageSize?: number
}

export const locationApi = {
  saved: (filter: SavedLocationFilter = {}) =>
    request<Page<SavedLocation>>(`/locations/saved${toQueryString({ ...filter })}`),
}
