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
