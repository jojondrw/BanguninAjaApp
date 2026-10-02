import { request } from '../shared/apiClient'
import type {
  BudgetFilter,
  BudgetRequest,
  CashListFilter,
  JournalEntry,
  JournalEntryDetail,
  JournalEntryRequest,
  JournalFilter,
  LedgerFilter,
  LedgerPage,
} from './accounting'
import { toQueryString, type Page } from './common'
import type { Budget, CashTransaction, CashTransactionRequest } from './finance'

export const accountingApi = {
  budgets: (filter: BudgetFilter = {}) =>
    request<Page<Budget>>(`/finance/budgets${toQueryString({ ...filter })}`),

  createBudget: (body: BudgetRequest) => request<Budget>('/finance/budgets', { method: 'POST', body }),

  updateBudget: (id: string, body: BudgetRequest) =>
    request<Budget>(`/finance/budgets/${id}`, { method: 'PUT', body }),

  deleteBudget: (id: string) => request<void>(`/finance/budgets/${id}`, { method: 'DELETE' }),

  cashTransactions: (filter: CashListFilter = {}) =>
    request<Page<CashTransaction>>(`/finance/cash-transactions${toQueryString({ ...filter })}`),

  cashTransaction: (id: string) => request<CashTransaction>(`/finance/cash-transactions/${id}`),

  createCashTransaction: (body: CashTransactionRequest) =>
    request<CashTransaction>('/finance/cash-transactions', { method: 'POST', body }),

  updateCashTransaction: (id: string, body: CashTransactionRequest) =>
    request<CashTransaction>(`/finance/cash-transactions/${id}`, { method: 'PUT', body }),

  deleteCashTransaction: (id: string) =>
    request<void>(`/finance/cash-transactions/${id}`, { method: 'DELETE' }),

  journalEntries: (filter: JournalFilter = {}) =>
    request<Page<JournalEntry>>(`/finance/journal-entries${toQueryString({ ...filter })}`),

  journalEntry: (id: string) => request<JournalEntryDetail>(`/finance/journal-entries/${id}`),

  createJournalEntry: (body: JournalEntryRequest) =>
    request<JournalEntryDetail>('/finance/journal-entries', { method: 'POST', body }),

  ledger: (filter: LedgerFilter) => request<LedgerPage>(`/finance/ledger${toQueryString({ ...filter })}`),
}
