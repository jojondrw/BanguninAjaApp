import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  BudgetItem,
  BudgetItemFilter,
  BudgetItemPage,
  BudgetItemRequest,
  Permit,
  PermitRequest,
  PhaseRequest,
  Project,
  ProjectPhase,
  ProjectRequest,
  ProjectStatus,
  Region,
} from './project'

const REGION_RESULT_LIMIT = 30

// Endpoint ubah dan hapus untuk ruang kerja proyek. Endpoint baca dan tambah
// yang lebih dulu ada tetap di projectApi (erpApi.ts).
export const projectWorkspaceApi = {
  update: (id: string, body: ProjectRequest) =>
    request<Project>(`/projects/${id}`, { method: 'PUT', body }),

  updateStatus: (id: string, status: ProjectStatus) =>
    request<Project>(`/projects/${id}/status`, { method: 'PATCH', body: { status } }),

  remove: (id: string) => request<void>(`/projects/${id}`, { method: 'DELETE' }),

  updatePhase: (projectId: string, phaseId: string, body: PhaseRequest) =>
    request<ProjectPhase>(`/projects/${projectId}/phases/${phaseId}`, { method: 'PUT', body }),

  deletePhase: (projectId: string, phaseId: string) =>
    request<void>(`/projects/${projectId}/phases/${phaseId}`, { method: 'DELETE' }),

  updatePermit: (projectId: string, permitId: string, body: PermitRequest) =>
    request<Permit>(`/projects/${projectId}/permits/${permitId}`, { method: 'PUT', body }),

  deletePermit: (projectId: string, permitId: string) =>
    request<void>(`/projects/${projectId}/permits/${permitId}`, { method: 'DELETE' }),

  budgetItems: (projectId: string, filter: BudgetItemFilter = {}) =>
    request<BudgetItemPage>(`/projects/${projectId}/budget-items${toQueryString({ ...filter })}`),

  createBudgetItem: (projectId: string, body: BudgetItemRequest) =>
    request<BudgetItem>(`/projects/${projectId}/budget-items`, { method: 'POST', body }),

  updateBudgetItem: (projectId: string, itemId: string, body: BudgetItemRequest) =>
    request<BudgetItem>(`/projects/${projectId}/budget-items/${itemId}`, { method: 'PUT', body }),

  deleteBudgetItem: (projectId: string, itemId: string) =>
    request<void>(`/projects/${projectId}/budget-items/${itemId}`, { method: 'DELETE' }),
}

export const regionApi = {
  search: (search: string) =>
    request<Page<Region>>(`/master/regions${toQueryString({ search, pageSize: REGION_RESULT_LIMIT })}`),

  get: (id: string) => request<Region>(`/master/regions/${id}`),
}
