import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { financeApi, masterApi, projectApi } from '../models/erpApi'
import {
  lastMonthsRange,
  toCashTransactionRequest,
  type CashTransactionFormValues,
} from '../models/finance'
import {
  toPermitRequest,
  toPhaseRequest,
  type PermitFormValues,
  type PhaseFormValues,
} from '../models/project'
import {
  BUDGETS_KEY,
  CASH_FLOW_KEY,
  CASH_TRANSACTIONS_KEY,
  DATA_QUERY,
  PROJECTS_KEY,
} from './useErp'

const MASTER_KEY = 'master'
const PROJECT_TRANSACTION_LIMIT = 50
export const CASH_FLOW_MONTHS = 12

// Semua kunci di bawah [PROJECTS_KEY, 'detail', id], jadi membatalkan
// PROJECTS_KEY ikut menyegarkan kepala proyek, tahap, izin, dan RAB.
function detailKey(projectId: string, part?: string) {
  return part ? [PROJECTS_KEY, 'detail', projectId, part] : [PROJECTS_KEY, 'detail', projectId]
}

export function useProject(projectId: string) {
  return useQuery({
    queryKey: detailKey(projectId),
    queryFn: () => projectApi.get(projectId),
    enabled: projectId !== '',
    ...DATA_QUERY,
  })
}

export function usePhases(projectId: string) {
  return useQuery({
    queryKey: detailKey(projectId, 'phases'),
    queryFn: () => projectApi.phases(projectId),
    ...DATA_QUERY,
  })
}

// Tahap baru mengubah progres proyek di server, jadi daftar proyek dan kepala
// proyek ikut diambil ulang, bukan hanya daftar tahap.
export function useCreatePhase(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ values, sortOrder }: { values: PhaseFormValues; sortOrder: number }) =>
      projectApi.createPhase(projectId, toPhaseRequest(values, sortOrder)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] }),
  })
}

export function usePermits(projectId: string) {
  return useQuery({
    queryKey: detailKey(projectId, 'permits'),
    queryFn: () => projectApi.permits(projectId),
    ...DATA_QUERY,
  })
}

export function useCreatePermit(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: PermitFormValues) => projectApi.createPermit(projectId, toPermitRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: detailKey(projectId, 'permits') }),
  })
}

export function useBudgetItems(projectId: string) {
  return useQuery({
    queryKey: detailKey(projectId, 'budget-items'),
    queryFn: () => projectApi.budgetItems(projectId),
    ...DATA_QUERY,
  })
}

// Rentang tanggal dihitung saat mengambil data, bukan saat render, supaya
// kunci query tetap sama sepanjang hari.
export function useProjectCashFlow(projectId: string) {
  return useQuery({
    queryKey: [CASH_FLOW_KEY, { projectId, months: CASH_FLOW_MONTHS }],
    queryFn: () =>
      financeApi.cashFlow({ projectId, ...lastMonthsRange(new Date(), CASH_FLOW_MONTHS) }),
    ...DATA_QUERY,
  })
}

export function useProjectCashTransactions(projectId: string) {
  return useQuery({
    queryKey: [CASH_TRANSACTIONS_KEY, 'project', projectId],
    queryFn: () =>
      financeApi.cashTransactions({ projectId, pageSize: PROJECT_TRANSACTION_LIMIT }),
    ...DATA_QUERY,
  })
}

// Transaksi kas mengubah realisasi anggaran, arus kas, dan saldo, baik di
// ruang kerja proyek maupun di halaman Keuangan dan Ringkasan.
export function useCreateCashTransaction(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: CashTransactionFormValues) =>
      financeApi.createCashTransaction(toCashTransactionRequest(values, projectId)),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: [CASH_TRANSACTIONS_KEY] }),
        queryClient.invalidateQueries({ queryKey: [CASH_FLOW_KEY] }),
        queryClient.invalidateQueries({ queryKey: [BUDGETS_KEY] }),
      ]),
  })
}

export function useAccounts() {
  return useQuery({
    queryKey: [MASTER_KEY, 'accounts'],
    queryFn: masterApi.accounts,
    ...DATA_QUERY,
  })
}

export function useUnitsOfMeasure() {
  return useQuery({
    queryKey: [MASTER_KEY, 'units-of-measure'],
    queryFn: masterApi.unitsOfMeasure,
    ...DATA_QUERY,
  })
}
