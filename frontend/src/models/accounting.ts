import { toApiDate } from '../shared/localDate'
import type { Page } from './common'
import type { Budget, CashTransaction, CashTransactionRequest, CashType } from './finance'
import { ACCOUNT_TYPE_LABEL, type Account, type AccountGroup, type AccountType } from './master'

// Batas di bawah mengikuti tag binding di backend/internal/finance/dto.go.
export const BUDGET_YEAR_MIN = 2000
export const BUDGET_YEAR_MAX = 2100
export const BUDGET_NOTE_MAX_LENGTH = 200
export const JOURNAL_NUMBER_MAX_LENGTH = 40
export const JOURNAL_NOTE_MAX_LENGTH = 200
export const JOURNAL_SOURCE_MAX_LENGTH = 40
export const JOURNAL_MIN_LINES = 2

export interface BudgetFilter {
  projectId?: string
  year?: number
  page?: number
  pageSize?: number
}

export interface BudgetRequest {
  projectId: string
  year: number
  value: number
  note: string
}

export interface BudgetFormValues {
  projectId: string
  year: string
  value: string
  note: string
}

export function emptyBudgetForm(year: number): BudgetFormValues {
  return { projectId: '', year: String(year), value: '', note: '' }
}

export function budgetFormFrom(budget: Budget): BudgetFormValues {
  return {
    projectId: budget.projectId,
    year: String(budget.year),
    value: String(budget.value),
    note: budget.note,
  }
}

export function toBudgetRequest(values: BudgetFormValues): BudgetRequest {
  return {
    projectId: values.projectId,
    year: Number(values.year),
    value: values.value === '' ? 0 : Number(values.value),
    note: values.note.trim(),
  }
}

export function isBudgetYear(value: string): boolean {
  const year = Number(value)
  return /^\d{4}$/.test(value) && year >= BUDGET_YEAR_MIN && year <= BUDGET_YEAR_MAX
}

export interface CashListFilter {
  projectId?: string
  type?: CashType
  accountId?: string
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

export interface CashFormValues {
  date: string
  type: CashType
  accountId: string
  projectId: string
  amount: string
  note: string
}

export function emptyCashForm(today: string): CashFormValues {
  return { date: today, type: 'out', accountId: '', projectId: '', amount: '', note: '' }
}

export function cashFormFrom(transaction: CashTransaction): CashFormValues {
  return {
    date: transaction.date.slice(0, 10),
    type: transaction.type,
    accountId: transaction.accountId,
    projectId: transaction.projectId ?? '',
    amount: String(transaction.amount),
    note: transaction.note,
  }
}

// projectId yang kosong tidak dikirim, jadi backend menyimpannya sebagai null
// (transaksi kantor pusat). Saat mengubah, ini juga melepas proyek lamanya.
export function toCashRequest(values: CashFormValues): CashTransactionRequest {
  return {
    date: toApiDate(values.date),
    type: values.type,
    accountId: values.accountId,
    projectId: values.projectId === '' ? undefined : values.projectId,
    amount: Number(values.amount),
    note: values.note.trim(),
  }
}

export function signedAmount(transaction: Pick<CashTransaction, 'type' | 'amount'>): number {
  return transaction.type === 'in' ? transaction.amount : -transaction.amount
}

export interface JournalEntry {
  id: string
  number: string
  date: string
  note: string
  source: string
  // Terisi untuk jurnal otomatis dari transaksi kas, null untuk jurnal manual.
  cashTransactionId: string | null
  total: number
  createdAt: string
}

// Mengikuti backend/internal/finance/cashjournal.go. Akun Kas dicari lewat
// kodenya, sedangkan sumber dan awalan nomor di bawah khusus untuk jurnal
// otomatis, jadi jurnal manual yang memakainya ditolak backend.
export const CASH_ACCOUNT_CODE = '1110'
export const CASH_JOURNAL_SOURCE = 'kas'
export const CASH_JOURNAL_NUMBER_PREFIX = 'KAS-'
export const CASH_JOURNAL_LABEL = 'Otomatis dari kas'

export function isCashJournal(entry: Pick<JournalEntry, 'cashTransactionId'>): boolean {
  return entry.cashTransactionId !== null
}

// Pesan yang sama dengan journal_source_reserved dan journal_number_reserved,
// supaya formulir sudah menolak sebelum dikirim.
export function reservedJournalReason(values: { number: string; source: string }): string | null {
  if (values.source.trim().toLowerCase() === CASH_JOURNAL_SOURCE) {
    return 'Sumber "kas" khusus untuk jurnal otomatis dari transaksi kas. Pakai sumber lain.'
  }
  if (values.number.trim().toUpperCase().startsWith(CASH_JOURNAL_NUMBER_PREFIX)) {
    return `Nomor berawalan ${CASH_JOURNAL_NUMBER_PREFIX} khusus untuk jurnal otomatis dari transaksi kas. Pakai nomor lain.`
  }
  return null
}

export interface JournalLine {
  id: string
  accountId: string
  debit: number
  credit: number
}

export interface JournalEntryDetail extends JournalEntry {
  lines: JournalLine[]
}

export interface JournalLineRequest {
  accountId: string
  debit: number
  credit: number
}

export interface JournalEntryRequest {
  number: string
  date: string
  note: string
  source: string
  lines: JournalLineRequest[]
}

export interface JournalFilter {
  search?: string
  accountId?: string
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

export type JournalSide = 'debit' | 'credit'

export const JOURNAL_SIDE_LABEL: Record<JournalSide, string> = {
  debit: 'Debit',
  credit: 'Kredit',
}

export interface JournalLineValues {
  key: string
  accountId: string
  side: JournalSide
  amount: string
}

export interface JournalFormValues {
  number: string
  date: string
  note: string
  source: string
  lines: JournalLineValues[]
}

// Kunci baris hanya untuk React. Penghitung biasa dipakai karena
// crypto.randomUUID tidak tersedia di origin http selain localhost.
let lineSequence = 0

export function emptyJournalLine(side: JournalSide): JournalLineValues {
  lineSequence += 1
  return { key: `journal-line-${lineSequence}`, accountId: '', side, amount: '' }
}

export function emptyJournalForm(today: string): JournalFormValues {
  return {
    number: '',
    date: today,
    note: '',
    source: '',
    lines: [emptyJournalLine('debit'), emptyJournalLine('credit')],
  }
}

function lineAmount(line: JournalLineValues): number {
  const amount = Number(line.amount)
  return line.amount !== '' && Number.isFinite(amount) && amount > 0 ? amount : 0
}

export interface JournalBalance {
  debit: number
  credit: number
  difference: number
  isBalanced: boolean
}

// Aturan yang sama dengan validateJournalLines di backend: total debit harus
// sama dengan total kredit. Jurnal bernilai nol juga ditolak di sini, karena
// backend mewajibkan setiap baris bernilai lebih dari nol.
export function journalBalance(lines: JournalLineValues[]): JournalBalance {
  const debit = lines.filter((line) => line.side === 'debit').reduce((total, line) => total + lineAmount(line), 0)
  const credit = lines.filter((line) => line.side === 'credit').reduce((total, line) => total + lineAmount(line), 0)
  return { debit, credit, difference: debit - credit, isBalanced: debit > 0 && debit === credit }
}

// Satu baris hanya berisi debit atau kredit, sesuai constraint single_sided di
// tabel journal_line.
export function toJournalRequest(values: JournalFormValues): JournalEntryRequest {
  return {
    number: values.number.trim(),
    date: toApiDate(values.date),
    note: values.note.trim(),
    source: values.source.trim(),
    lines: values.lines.map((line) => ({
      accountId: line.accountId,
      debit: line.side === 'debit' ? Number(line.amount) : 0,
      credit: line.side === 'credit' ? Number(line.amount) : 0,
    })),
  }
}

// Backend menyimpan baris dalam satu insert, jadi urutannya tidak bermakna.
// Ditampilkan seperti jurnal umum: baris debit dulu, lalu kredit.
export function orderedLines(lines: JournalLine[]): JournalLine[] {
  return [...lines.filter((line) => line.debit > 0), ...lines.filter((line) => line.debit === 0)]
}

export type JournalDetailRow =
  | { kind: 'line'; line: JournalLine }
  | { kind: 'total'; debit: number; credit: number }

export function journalTotals(lines: JournalLine[]): { debit: number; credit: number } {
  return {
    debit: lines.reduce((total, line) => total + line.debit, 0),
    credit: lines.reduce((total, line) => total + line.credit, 0),
  }
}

export function journalDetailRows(lines: JournalLine[]): JournalDetailRow[] {
  const rows: JournalDetailRow[] = orderedLines(lines).map((line) => ({ kind: 'line', line }))
  return [...rows, { kind: 'total', ...journalTotals(lines) }]
}

export interface LedgerLine {
  journalEntryId: string
  number: string
  date: string
  note: string
  cashTransactionId: string | null
  debit: number
  credit: number
  balance: number
}

export interface LedgerPage extends Page<LedgerLine> {
  openingBalance: number
  totalDebit: number
  totalCredit: number
  closingBalance: number
}

export type LedgerRow =
  | { kind: 'line'; line: LedgerLine }
  | { kind: 'opening'; balance: number }
  | { kind: 'closing'; balance: number; debit: number; credit: number }

// Baris saldo awal hanya di halaman pertama dan saldo akhir hanya di halaman
// terakhir, supaya tabel terbaca seperti buku besar cetak.
export function ledgerRows(ledger: LedgerPage): LedgerRow[] {
  if (ledger.items.length === 0) {
    return []
  }
  const rows: LedgerRow[] = ledger.items.map((line) => ({ kind: 'line', line }))
  if (ledger.page === 1) {
    rows.unshift({ kind: 'opening', balance: ledger.openingBalance })
  }
  if (ledger.page >= ledger.totalPages) {
    rows.push({ kind: 'closing', balance: ledger.closingBalance, debit: ledger.totalDebit, credit: ledger.totalCredit })
  }
  return rows
}

export interface LedgerFilter {
  accountId: string
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

// Saldo buku besar dihitung backend sebagai debit dikurangi kredit. Positif
// berarti bersaldo debit, negatif berarti bersaldo kredit.
export type BalanceSide = 'D' | 'K' | ''

export function balanceSide(balance: number): BalanceSide {
  if (balance > 0) {
    return 'D'
  }
  return balance < 0 ? 'K' : ''
}

export const NORMAL_BALANCE: Record<AccountType, Exclude<BalanceSide, ''>> = {
  asset: 'D',
  expense: 'D',
  liability: 'K',
  equity: 'K',
  revenue: 'K',
}

const ACCOUNT_TYPE_ORDER: AccountType[] = ['asset', 'liability', 'equity', 'revenue', 'expense']

function groupAccounts(accounts: Account[]): AccountGroup[] {
  const sorted = [...accounts].sort((a, b) => a.code.localeCompare(b.code))
  return ACCOUNT_TYPE_ORDER.map((type) => ({
    type,
    label: ACCOUNT_TYPE_LABEL[type],
    accounts: sorted.filter((account) => account.type === type),
  })).filter((group) => group.accounts.length > 0)
}

// Akun induk (yang punya anak) hanya untuk pengelompokan, jadi jurnal dicatat
// di akun rinciannya.
export function postableAccountGroups(accounts: Account[]): AccountGroup[] {
  const parents = new Set(accounts.map((account) => account.parentId).filter((id) => id !== null))
  return groupAccounts(accounts.filter((account) => !parents.has(account.id)))
}

// Buku besar menawarkan semua akun, termasuk induk, supaya baris yang terlanjur
// dicatat langsung ke akun induk lewat API tetap bisa diperiksa.
export function allAccountGroups(accounts: Account[]): AccountGroup[] {
  return groupAccounts(accounts)
}

export function accountLabel(accounts: Account[], id: string): string {
  const account = accounts.find((item) => item.id === id)
  return account ? `${account.code} ${account.name}` : '-'
}

// Tanggal "YYYY-MM-DD" bisa dibandingkan sebagai teks.
export function isDateRangeValid(dateFrom: string, dateTo: string): boolean {
  return dateFrom === '' || dateTo === '' || dateFrom <= dateTo
}
