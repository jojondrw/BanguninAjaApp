import { request } from '../shared/apiClient'
import type { BuildingProfile, EvaluateRequest, EvaluateResponse } from './site'

export const siteApi = {
  // GET /api/scoring/building-profiles returns a plain array in the data envelope.
  buildingProfiles: () => request<BuildingProfile[]>('/scoring/building-profiles'),

  // POST /api/site/evaluate (auth required) — the T4 endpoint.
  evaluate: (payload: EvaluateRequest) =>
    request<EvaluateResponse>('/site/evaluate', { method: 'POST', body: payload }),
}
