import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { Budget, CashFlow, CashFlowFilter, CashTransaction } from './finance'
import type { Project, ProjectFilter } from './project'

export const projectApi = {
  list: (filter: ProjectFilter = {}) =>
    request<Page<Project>>(`/projects${toQueryString({ ...filter })}`),

  get: (id: string) => request<Project>(`/projects/${id}`),
}

export const financeApi = {
  budgets: (projectId?: string) =>
    request<Page<Budget>>(`/finance/budgets${toQueryString({ projectId, pageSize: 100 })}`),

  cashFlow: (filter: CashFlowFilter = {}) =>
    request<CashFlow>(`/finance/cash-flow${toQueryString({ ...filter })}`),

  cashTransactions: (pageSize = 10) =>
    request<Page<CashTransaction>>(`/finance/cash-transactions${toQueryString({ pageSize })}`),
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
