import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { DailyReport, DailyReportPage, DailyReportRequest } from './dailyReport'
import type { OwnerContractRequest, OwnerSchedule } from './ownerBilling'
import type {
  ContractChangeRequest,
  ContractChangeStatus,
  ContractRequest,
  PaymentSchedule,
  PaymentTermRequest,
} from './paymentTerm'
import type {
  BudgetItem,
  BudgetItemFilter,
  BudgetItemPage,
  BudgetItemRequest,
  Permit,
  PermitRequest,
  PhaseRequest,
  ProgressCurve,
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

  progressCurve: (projectId: string) => request<ProgressCurve>(`/projects/${projectId}/progress-curve`),

  dailyReports: (projectId: string, page: number, pageSize: number) =>
    request<DailyReportPage>(`/projects/${projectId}/daily-reports${toQueryString({ page, pageSize })}`),

  paymentSchedule: (projectId: string) => request<PaymentSchedule>(`/projects/${projectId}/payment-schedule`),

  saveContract: (projectId: string, body: ContractRequest) =>
    request<PaymentSchedule>(`/projects/${projectId}/construction-contract`, { method: 'PUT', body }),

  createPaymentTerm: (projectId: string, body: PaymentTermRequest) =>
    request<PaymentSchedule>(`/projects/${projectId}/payment-terms`, { method: 'POST', body }),

  deletePaymentTerm: (projectId: string, termId: string) =>
    request<PaymentSchedule>(`/projects/${projectId}/payment-terms/${termId}`, { method: 'DELETE' }),

  claimPaymentTerm: (projectId: string, termId: string) =>
    request<PaymentSchedule>(`/projects/${projectId}/payment-terms/${termId}/claim`, { method: 'POST' }),

  releaseRetention: (projectId: string) =>
    request<PaymentSchedule>(`/projects/${projectId}/retention-release`, { method: 'POST' }),

  ownerSchedule: (projectId: string) => request<OwnerSchedule>(`/projects/${projectId}/owner-billing`),

  saveOwnerContract: (projectId: string, body: OwnerContractRequest) =>
    request<OwnerSchedule>(`/projects/${projectId}/owner-billing/contract`, { method: 'PUT', body }),

  createOwnerTerm: (projectId: string, body: PaymentTermRequest) =>
    request<OwnerSchedule>(`/projects/${projectId}/owner-billing/terms`, { method: 'POST', body }),

  deleteOwnerTerm: (projectId: string, termId: string) =>
    request<OwnerSchedule>(`/projects/${projectId}/owner-billing/terms/${termId}`, { method: 'DELETE' }),

  claimOwnerTerm: (projectId: string, termId: string) =>
    request<OwnerSchedule>(`/projects/${projectId}/owner-billing/terms/${termId}/claim`, { method: 'POST' }),

  releaseOwnerRetention: (projectId: string) =>
    request<OwnerSchedule>(`/projects/${projectId}/owner-billing/retention-release`, { method: 'POST' }),

  proposeContractChange: (projectId: string, body: ContractChangeRequest) =>
    request<PaymentSchedule>(`/projects/${projectId}/contract-changes`, { method: 'POST', body }),

  decideContractChange: (projectId: string, changeId: string, status: Exclude<ContractChangeStatus, 'proposed'>) =>
    request<PaymentSchedule>(`/projects/${projectId}/contract-changes/${changeId}`, { method: 'PATCH', body: { status } }),

  deleteContractChange: (projectId: string, changeId: string) =>
    request<PaymentSchedule>(`/projects/${projectId}/contract-changes/${changeId}`, { method: 'DELETE' }),

  createDailyReport: (projectId: string, body: DailyReportRequest) =>
    request<DailyReport>(`/projects/${projectId}/daily-reports`, { method: 'POST', body }),

  updateDailyReport: (projectId: string, reportId: string, body: DailyReportRequest) =>
    request<DailyReport>(`/projects/${projectId}/daily-reports/${reportId}`, { method: 'PUT', body }),

  deleteDailyReport: (projectId: string, reportId: string) =>
    request<void>(`/projects/${projectId}/daily-reports/${reportId}`, { method: 'DELETE' }),

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
