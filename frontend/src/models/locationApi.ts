import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  Comparison,
  ComparisonDetail,
  ComparisonFilter,
  ComparisonRequest,
  Dimension,
  Region,
  SavedLocationDetail,
} from './location'

// Daftar lokasi tersimpan sudah ada di erpApi.locationApi.saved. Yang di sini
// hanya rincian per lokasi (beserta skor dimensi) dan sumber daya perbandingan.
export const savedLocationApi = {
  get: (id: string) => request<SavedLocationDetail>(`/locations/saved/${id}`),
}

export const dimensionApi = {
  list: () => request<Dimension[]>('/scoring/dimensions'),
}

export const regionApi = {
  get: (id: string) => request<Region>(`/master/regions/${id}`),
}

export const comparisonApi = {
  list: (filter: ComparisonFilter = {}) =>
    request<Page<Comparison>>(`/locations/comparisons${toQueryString({ ...filter })}`),

  get: (id: string) => request<ComparisonDetail>(`/locations/comparisons/${id}`),

  create: (body: ComparisonRequest) =>
    request<ComparisonDetail>('/locations/comparisons', { method: 'POST', body }),

  update: (id: string, body: ComparisonRequest) =>
    request<ComparisonDetail>(`/locations/comparisons/${id}`, { method: 'PUT', body }),

  remove: (id: string) => request<void>(`/locations/comparisons/${id}`, { method: 'DELETE' }),
}
