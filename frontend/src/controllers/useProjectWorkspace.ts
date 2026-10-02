import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { financeApi, masterApi, projectApi } from '../models/erpApi'
import {
  lastMonthsRange,
  toCashTransactionRequest,
  type CashTransactionFormValues,
} from '../models/finance'
import {
  REGION_SEARCH_MIN_LENGTH,
  toBudgetItemRequest,
  toPermitRequest,
  toPhaseEditRequest,
  toPhaseRequest,
  toProjectUpdateRequest,
  type BudgetItemFilter,
  type BudgetItemFormValues,
  type PermitFormValues,
  type PhaseEditValues,
  type PhaseFormValues,
  type Project,
  type ProjectFormValues,
  type ProjectStatus,
} from '../models/project'
import { projectWorkspaceApi, regionApi } from '../models/projectWorkspaceApi'
import {
  BUDGETS_KEY,
  CASH_FLOW_KEY,
  CASH_TRANSACTIONS_KEY,
  DATA_QUERY,
  PROJECTS_KEY,
} from './useErp'
import { JOURNAL_KEY, LEDGER_KEY } from './useFinance'

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
// ruang kerja proyek maupun di halaman Keuangan dan Ringkasan. Backend juga
// membuat jurnal otomatisnya, jadi jurnal dan buku besar ikut diambil ulang.
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
        queryClient.invalidateQueries({ queryKey: [JOURNAL_KEY] }),
        queryClient.invalidateQueries({ queryKey: [LEDGER_KEY] }),
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

// Kepala proyek langsung diganti dengan jawaban server supaya judul dan status
// berubah tanpa menunggu, lalu semua query proyek (daftar /proyek, pilihan
// proyek di /lokasi, ringkasan) ikut diambil ulang.
function useProjectSaved(projectId: string) {
  const queryClient = useQueryClient()

  return (project: Project) => {
    queryClient.setQueryData(detailKey(projectId), project)
    return queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] })
  }
}

export function useUpdateProject(projectId: string) {
  const onSaved = useProjectSaved(projectId)

  return useMutation({
    mutationFn: ({ values, regionId }: { values: ProjectFormValues; regionId: string }) =>
      projectWorkspaceApi.update(projectId, toProjectUpdateRequest(values, regionId)),
    onSuccess: onSaved,
  })
}

export function useUpdateProjectStatus(projectId: string) {
  const onSaved = useProjectSaved(projectId)

  return useMutation({
    mutationFn: (status: ProjectStatus) => projectWorkspaceApi.updateStatus(projectId, status),
    onSuccess: onSaved,
  })
}

// Menghapus proyek ikut menghapus tahap, izin, RAB, dan anggaran, serta
// melepas transaksi kas dan lokasi dari proyek. Semua query cukup ditandai
// basi tanpa diambil ulang sekarang: halaman ini akan ditinggalkan, dan
// halaman tujuan mengambil data segar saat dibuka.
export function useDeleteProject(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => projectWorkspaceApi.remove(projectId),
    onSuccess: () => queryClient.invalidateQueries({ refetchType: 'none' }),
  })
}

// Mengubah atau menghapus tahap menghitung ulang progres proyek di server,
// jadi seluruh query proyek diambil ulang, bukan hanya daftar tahap.
export function useUpdatePhase(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ phaseId, values }: { phaseId: string; values: PhaseEditValues }) =>
      projectWorkspaceApi.updatePhase(projectId, phaseId, toPhaseEditRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] }),
  })
}

export function useDeletePhase(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (phaseId: string) => projectWorkspaceApi.deletePhase(projectId, phaseId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] }),
  })
}

export function useUpdatePermit(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ permitId, values }: { permitId: string; values: PermitFormValues }) =>
      projectWorkspaceApi.updatePermit(projectId, permitId, toPermitRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: detailKey(projectId, 'permits') }),
  })
}

export function useDeletePermit(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (permitId: string) => projectWorkspaceApi.deletePermit(projectId, permitId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: detailKey(projectId, 'permits') }),
  })
}

// Kuncinya di bawah detailKey(projectId, 'budget-items'), jadi satu
// pembatalan menyegarkan tabel RAB sekaligus KPI Nilai RAB.
export function useBudgetItemPage(projectId: string, filter: BudgetItemFilter) {
  return useQuery({
    queryKey: [...detailKey(projectId, 'budget-items'), filter],
    queryFn: () => projectWorkspaceApi.budgetItems(projectId, filter),
    placeholderData: keepPreviousData,
    ...DATA_QUERY,
  })
}

export function useCreateBudgetItem(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: BudgetItemFormValues) =>
      projectWorkspaceApi.createBudgetItem(projectId, toBudgetItemRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: detailKey(projectId, 'budget-items') }),
  })
}

export function useUpdateBudgetItem(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ itemId, values }: { itemId: string; values: BudgetItemFormValues }) =>
      projectWorkspaceApi.updateBudgetItem(projectId, itemId, toBudgetItemRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: detailKey(projectId, 'budget-items') }),
  })
}

export function useDeleteBudgetItem(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (itemId: string) => projectWorkspaceApi.deleteBudgetItem(projectId, itemId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: detailKey(projectId, 'budget-items') }),
  })
}

export function useRegion(regionId: string) {
  return useQuery({
    queryKey: [MASTER_KEY, 'region', regionId],
    queryFn: () => regionApi.get(regionId),
    enabled: regionId !== '',
    ...DATA_QUERY,
  })
}

export function useRegionSearch(term: string) {
  return useQuery({
    queryKey: [MASTER_KEY, 'region-search', term],
    queryFn: () => regionApi.search(term),
    enabled: term.length >= REGION_SEARCH_MIN_LENGTH,
    placeholderData: keepPreviousData,
    ...DATA_QUERY,
  })
}
