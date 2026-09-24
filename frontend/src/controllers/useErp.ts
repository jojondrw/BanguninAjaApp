import { useQuery } from '@tanstack/react-query'

import { financeApi, projectApi } from '../models/erpApi'
import type { CashFlowFilter } from '../models/finance'
import type { ProjectFilter } from '../models/project'

const ONE_MINUTE = 60_000

const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

export function useProjects(filter: ProjectFilter = {}) {
  return useQuery({
    queryKey: ['projects', filter],
    queryFn: () => projectApi.list(filter),
    ...DATA_QUERY,
  })
}

export function useBudgets(projectId?: string) {
  return useQuery({
    queryKey: ['budgets', projectId ?? 'all'],
    queryFn: () => financeApi.budgets(projectId),
    ...DATA_QUERY,
  })
}

export function useCashFlow(filter: CashFlowFilter = {}) {
  return useQuery({
    queryKey: ['cash-flow', filter],
    queryFn: () => financeApi.cashFlow(filter),
    ...DATA_QUERY,
  })
}

export function useCashTransactions(pageSize = 10) {
  return useQuery({
    queryKey: ['cash-transactions', pageSize],
    queryFn: () => financeApi.cashTransactions(pageSize),
    ...DATA_QUERY,
  })
}
