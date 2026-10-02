import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { financeApi, locationApi, projectApi, scoringApi, siteApi } from '../models/erpApi'
import type { CashFlowFilter } from '../models/finance'
import { toProjectRequest, type ProjectFilter, type ProjectFormValues } from '../models/project'

const ONE_MINUTE = 60_000
const MAX_PAGE_SIZE = 100
export const PROJECTS_KEY = 'projects'
export const SAVED_LOCATIONS_KEY = 'saved-locations'
export const BUDGETS_KEY = 'budgets'
export const CASH_FLOW_KEY = 'cash-flow'
export const CASH_TRANSACTIONS_KEY = 'cash-transactions'

export const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

export function useProjects(filter: ProjectFilter = {}) {
  return useQuery({
    queryKey: [PROJECTS_KEY, filter],
    queryFn: () => projectApi.list(filter),
    ...DATA_QUERY,
  })
}

// Membatalkan semua query proyek, jadi daftar di /proyek, pilihan proyek di
// /lokasi, dan ringkasan ikut mengambil ulang data terbaru.
export function useCreateProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: ProjectFormValues) => projectApi.create(toProjectRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] }),
  })
}

export function useBudgets(projectId?: string) {
  return useQuery({
    queryKey: [BUDGETS_KEY, projectId ?? 'all'],
    queryFn: () => financeApi.budgets(projectId),
    ...DATA_QUERY,
  })
}

export function useCashFlow(filter: CashFlowFilter = {}) {
  return useQuery({
    queryKey: [CASH_FLOW_KEY, filter],
    queryFn: () => financeApi.cashFlow(filter),
    ...DATA_QUERY,
  })
}

export function useCashTransactions(pageSize = 10) {
  return useQuery({
    queryKey: [CASH_TRANSACTIONS_KEY, pageSize],
    queryFn: () => financeApi.cashTransactions({ pageSize }),
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
