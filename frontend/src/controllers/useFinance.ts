import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  toBudgetRequest,
  toCashRequest,
  toJournalRequest,
  type BudgetFilter,
  type BudgetFormValues,
  type CashFormValues,
  type CashListFilter,
  type JournalFilter,
  type JournalFormValues,
  type LedgerFilter,
} from '../models/accounting'
import { accountingApi } from '../models/accountingApi'
import { BUDGETS_KEY, CASH_FLOW_KEY, CASH_TRANSACTIONS_KEY, DATA_QUERY } from './useErp'

const JOURNAL_KEY = 'journal-entries'
const LEDGER_KEY = 'ledger'

// Daftar yang dipaginasi tetap menampilkan halaman lama selama halaman baru
// dimuat, supaya tabel tidak berkedip kosong setiap pindah halaman.
const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

// Kunci di bawah BUDGETS_KEY dan CASH_TRANSACTIONS_KEY dipakai bersama halaman
// Ringkasan dan ruang kerja proyek, jadi perubahan di sini ikut menyegarkannya.
export function useBudgetList(filter: BudgetFilter) {
  return useQuery({
    queryKey: [BUDGETS_KEY, 'list', filter],
    queryFn: () => accountingApi.budgets(filter),
    ...PAGED_QUERY,
  })
}

export function useCreateBudget() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: BudgetFormValues) => accountingApi.createBudget(toBudgetRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BUDGETS_KEY] }),
  })
}

export function useUpdateBudget() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, values }: { id: string; values: BudgetFormValues }) =>
      accountingApi.updateBudget(id, toBudgetRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BUDGETS_KEY] }),
  })
}

export function useDeleteBudget() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => accountingApi.deleteBudget(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BUDGETS_KEY] }),
  })
}

export function useCashTransactionList(filter: CashListFilter, enabled = true) {
  return useQuery({
    queryKey: [CASH_TRANSACTIONS_KEY, 'list', filter],
    queryFn: () => accountingApi.cashTransactions(filter),
    enabled,
    ...PAGED_QUERY,
  })
}

// Transaksi kas mengubah arus kas, saldo, dan realisasi anggaran (kas keluar
// proyek di tahun anggaran), jadi ketiganya diambil ulang.
function useRefreshCash() {
  const queryClient = useQueryClient()

  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: [CASH_TRANSACTIONS_KEY] }),
      queryClient.invalidateQueries({ queryKey: [CASH_FLOW_KEY] }),
      queryClient.invalidateQueries({ queryKey: [BUDGETS_KEY] }),
    ])
}

export function useRecordCashTransaction() {
  const refresh = useRefreshCash()

  return useMutation({
    mutationFn: (values: CashFormValues) => accountingApi.createCashTransaction(toCashRequest(values)),
    onSuccess: refresh,
  })
}

export function useUpdateCashTransaction() {
  const refresh = useRefreshCash()

  return useMutation({
    mutationFn: ({ id, values }: { id: string; values: CashFormValues }) =>
      accountingApi.updateCashTransaction(id, toCashRequest(values)),
    onSuccess: refresh,
  })
}

export function useDeleteCashTransaction() {
  const refresh = useRefreshCash()

  return useMutation({
    mutationFn: (id: string) => accountingApi.deleteCashTransaction(id),
    onSuccess: refresh,
  })
}

export function useJournalEntries(filter: JournalFilter, enabled = true) {
  return useQuery({
    queryKey: [JOURNAL_KEY, 'list', filter],
    queryFn: () => accountingApi.journalEntries(filter),
    enabled,
    ...PAGED_QUERY,
  })
}

export function useJournalEntry(id: string | null) {
  return useQuery({
    queryKey: [JOURNAL_KEY, 'detail', id],
    queryFn: () => accountingApi.journalEntry(id ?? ''),
    enabled: id !== null,
    ...DATA_QUERY,
  })
}

// Jurnal baru menambah baris di buku besar akun yang dipakainya.
export function useCreateJournalEntry() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: JournalFormValues) => accountingApi.createJournalEntry(toJournalRequest(values)),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: [JOURNAL_KEY] }),
        queryClient.invalidateQueries({ queryKey: [LEDGER_KEY] }),
      ]),
  })
}

export function useLedger(filter: LedgerFilter, enabled: boolean) {
  return useQuery({
    queryKey: [LEDGER_KEY, filter],
    queryFn: () => accountingApi.ledger(filter),
    enabled: enabled && filter.accountId !== '',
    ...PAGED_QUERY,
  })
}
