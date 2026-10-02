import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { financeApi, locationApi, projectApi, scoringApi, siteApi } from '../models/erpApi'
import type { CashFlowFilter } from '../models/finance'
import type { ProjectFilter } from '../models/project'

const ONE_MINUTE = 60_000
const MAX_PAGE_SIZE = 100
const SAVED_LOCATIONS_KEY = 'saved-locations'

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

export function useSiteEvaluate() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: siteApi.evaluate,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [SAVED_LOCATIONS_KEY] }),
  })
}

export function useSavedLocations(projectId?: string) {
  return useQuery({
    queryKey: [SAVED_LOCATIONS_KEY, projectId ?? 'all'],
    queryFn: () => locationApi.saved({ projectId, pageSize: MAX_PAGE_SIZE }),
    ...DATA_QUERY,
  })
}

export function useBuildingProfiles() {
  return useQuery({
    queryKey: ['building-profiles'],
    queryFn: scoringApi.buildingProfiles,
    ...DATA_QUERY,
  })
}
