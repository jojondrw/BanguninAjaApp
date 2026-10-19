export interface Budget {
  id: string
  projectId: string
  // Diisi backend dari tabel proyek, jadi tabel tidak perlu memuat daftar proyek.
  projectName: string
  year: number
  value: number
  note: string
  realized: number
  remaining: number
  absorption: number
  createdAt: string
  updatedAt: string
}

export interface CashTransaction {
  id: string
  date: string
  type: 'in' | 'out'
  accountId: string
  // Kode dan nama akun serta nama proyek diisi backend pada daftar dan rincian.
  accountCode: string
  accountName: string
  projectId: string | null
  // Tidak dikirim untuk transaksi kantor pusat (tanpa proyek).
  projectName?: string
  amount: number
  note: string
  // Jurnal otomatis milik transaksi ini. null hanya untuk transaksi lama yang
  // belum diisi ulang oleh cmd/migrate.
  journalEntryId: string | null
  // Terisi kalau kas ini dicatat otomatis dari pembayaran di modul Tagihan.
  billingPaymentId: string | null
  createdAt: string
  updatedAt: string
}

export interface CashFlowPeriod {
  period: string
  cashIn: number
  cashOut: number
  net: number
}

export interface CashFlow {
  dateFrom: string
  dateTo: string
  periods: CashFlowPeriod[]
  totalIn: number
  totalOut: number
  net: number
  balance: number
}

// Jumlah kas per akun dalam satu rentang, dari server.
export interface CashAccountTotal {
  accountCode: string
  accountName: string
  type: 'in' | 'out'
  total: number
}

export interface CashFlowFilter {
  projectId?: string
  dateFrom?: string
  dateTo?: string
}

export type CashType = CashTransaction['type']

export interface CashTransactionFilter {
  projectId?: string
  type?: CashType
  page?: number
  pageSize?: number
}

export interface CashTransactionRequest {
  date: string
  type: CashType
  accountId: string
  projectId?: string
  amount: number
  note: string
}

export interface CashTransactionFormValues {
  date: string
  type: CashType
  accountId: string
  amount: string
  note: string
}

// Batas panjang mengikuti tag binding CashTransactionRequest di backend.
export const CASH_NOTE_MAX_LENGTH = 200

export const CASH_TYPE_LABEL: Record<CashType, string> = {
  in: 'Masuk',
  out: 'Keluar',
}

export const CASH_TYPE_TONE: Record<CashType, string> = {
  in: 'bg-green-100 text-green-800',
  out: 'bg-amber-100 text-amber-800',
}

export const BILLING_PATH = '/tagihan'

// Kas dari pembayaran tagihan diubah atau dihapus lewat Tagihan, supaya sisa
// tagihan dan kasnya tidak pernah berbeda.
export function isFromBilling(transaction: Pick<CashTransaction, 'billingPaymentId'>): boolean {
  return transaction.billingPaymentId !== null
}

// Tanggal lokal "YYYY-MM-DD" untuk isian bertipe date. toISOString tidak
// dipakai karena memakai UTC, jadi lewat tengah malam WIB tanggalnya mundur.
export function toInputDate(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

export function emptyCashTransactionForm(today: Date): CashTransactionFormValues {
  return { date: toInputDate(today), type: 'out', accountId: '', amount: '', note: '' }
}

export function toCashTransactionRequest(
  values: CashTransactionFormValues,
  projectId: string,
): CashTransactionRequest {
  return {
    date: `${values.date}T00:00:00Z`,
    type: values.type,
    accountId: values.accountId,
    projectId,
    amount: Number(values.amount),
    note: values.note.trim(),
  }
}

// Rentang arus kas: awal bulan ke-(months - 1) sebelum bulan ini sampai hari ini.
export function lastMonthsRange(today: Date, months: number): { dateFrom: string; dateTo: string } {
  const start = new Date(today.getFullYear(), today.getMonth() - (months - 1), 1)
  return { dateFrom: toInputDate(start), dateTo: toInputDate(today) }
}

export interface BudgetSummary {
  years: number
  value: number
  realized: number
  remaining: number
  absorption: number
}

export function summarizeBudgets(budgets: Budget[]): BudgetSummary {
  const value = budgets.reduce((total, budget) => total + budget.value, 0)
  const realized = budgets.reduce((total, budget) => total + budget.realized, 0)

  return {
    years: budgets.length,
    value,
    realized,
    remaining: value - realized,
    absorption: value > 0 ? Math.round((realized / value) * 100) : 0,
  }
}

// Akun terbesar ditampilkan satu per satu, sisanya digabung jadi "Lainnya"
// supaya daftar tetap pendek tetapi totalnya tetap utuh.
export function topAccounts(accounts: CashAccountTotal[], type: CashAccountTotal['type'], limit: number) {
  const rows = accounts.filter((account) => account.type === type).sort((a, b) => b.total - a.total)
  const total = rows.reduce((sum, row) => sum + row.total, 0)
  const shown = rows.slice(0, limit).map((row) => ({ label: row.accountName, code: row.accountCode, total: row.total }))
  const rest = rows.slice(limit).reduce((sum, row) => sum + row.total, 0)
  return { total, items: rest > 0 ? [...shown, { label: 'Lainnya', code: '', total: rest }] : shown }
}

// Serapan tertinggi dulu: proyek yang hampir atau sudah melewati anggarannya
// yang paling perlu dilihat.
export function busiestBudgets(budgets: Budget[], limit: number): Budget[] {
  return [...budgets].sort((a, b) => b.absorption - a.absorption).slice(0, limit)
}
