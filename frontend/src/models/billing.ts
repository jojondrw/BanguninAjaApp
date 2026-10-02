import { toApiDate } from '../shared/localDate'

// Tiga jenis catatan di backend/internal/billing. Nilainya sekaligus potongan
// path endpoint, jadi /billing/<kind>.
export type BillingKind = 'invoices' | 'receivables' | 'payables'
export type BillingStatus = 'not_due' | 'due' | 'paid' | 'overdue'
export type PartyType = 'customer' | 'vendor'

// Bagian yang sama di ketiga respons: nominal, pembayaran, dan tenggat.
export interface BillingBalance {
  dueDate: string
  amount: number
  paidAmount: number
  outstanding: number
  status: BillingStatus
  daysOverdue: number
}

export interface BillingRecord extends BillingBalance {
  id: string
  createdAt: string
  updatedAt: string
}

export interface Invoice extends BillingRecord {
  number: string
  note: string
  partyType: PartyType
  partyId: string
  projectId: string | null
}

export interface Receivable extends BillingRecord {
  customerId: string
  reference: string
}

export interface Payable extends BillingRecord {
  vendorId: string
  reference: string
}

interface PageQuery {
  search?: string
  status?: BillingStatus
  page?: number
  pageSize?: number
}

export interface InvoiceFilter extends PageQuery {
  partyType?: PartyType
  partyId?: string
  projectId?: string
}

export interface ReceivableFilter extends PageQuery {
  customerId?: string
}

export interface PayableFilter extends PageQuery {
  vendorId?: string
}

export interface InvoiceRequest {
  number: string
  note: string
  partyType: PartyType
  partyId: string
  projectId?: string
  dueDate: string
  amount: number
}

export interface ReceivableRequest {
  customerId: string
  reference: string
  dueDate: string
  amount: number
}

export interface PayableRequest {
  vendorId: string
  reference: string
  dueDate: string
  amount: number
}

// Backend hanya menerima nominal. Belum ada tanggal, metode, atau referensi
// per pembayaran, dan belum ada riwayatnya.
export interface PaymentRequest {
  amount: number
}

export interface BillingRecordOf {
  invoices: Invoice
  receivables: Receivable
  payables: Payable
}

export interface BillingFilterOf {
  invoices: InvoiceFilter
  receivables: ReceivableFilter
  payables: PayableFilter
}

export interface BillingRequestOf {
  invoices: InvoiceRequest
  receivables: ReceivableRequest
  payables: PayableRequest
}

// Batas mengikuti tag binding di backend/internal/billing/dto.go.
export const INVOICE_NUMBER_MAX_LENGTH = 40
export const INVOICE_NOTE_MAX_LENGTH = 200
export const REFERENCE_MAX_LENGTH = 60
export const INVOICE_SEARCH_MAX_LENGTH = 200
export const REFERENCE_SEARCH_MAX_LENGTH = 60

export const BILLING_STATUSES: BillingStatus[] = ['overdue', 'due', 'not_due', 'paid']
export const PARTY_TYPES: PartyType[] = ['customer', 'vendor']

export const BILLING_STATUS_LABEL: Record<BillingStatus, string> = {
  not_due: 'Belum jatuh tempo',
  due: 'Jatuh tempo 7 hari',
  paid: 'Lunas',
  overdue: 'Terlambat',
}

export const BILLING_STATUS_TONE: Record<BillingStatus, string> = {
  not_due: 'bg-slate-100 text-slate-700',
  due: 'bg-amber-100 text-amber-800',
  paid: 'bg-green-100 text-green-800',
  overdue: 'bg-red-100 text-red-800',
}

export const PARTY_TYPE_LABEL: Record<PartyType, string> = {
  customer: 'Pelanggan',
  vendor: 'Vendor',
}

export const BILLING_KIND_NOUN: Record<BillingKind, string> = {
  invoices: 'faktur',
  receivables: 'piutang',
  payables: 'utang',
}

const DAY_MS = 86_400_000

function calendarDay(value: string): number {
  const [year, month, day] = value.slice(0, 10).split('-').map(Number)
  return Date.UTC(year, month - 1, day)
}

// Selisih hari kalender antara jatuh tempo dan hari ini di jam peramban.
// Negatif berarti sudah lewat.
export function daysUntil(dueDate: string, today: string): number {
  return Math.round((calendarDay(dueDate) - calendarDay(today)) / DAY_MS)
}

// Keterangan tenggat dalam kalimat, supaya status tidak hanya dibedakan warna.
export function dueNote(record: BillingBalance, today: string): string | null {
  if (record.status === 'paid') {
    return null
  }
  if (record.status === 'overdue') {
    return `Terlambat ${record.daysOverdue} hari`
  }
  const days = daysUntil(record.dueDate, today)
  if (days <= 0) {
    return 'Jatuh tempo hari ini'
  }
  if (days === 1) {
    return 'Jatuh tempo besok'
  }
  return record.status === 'due' ? `${days} hari lagi` : null
}

export function paidPercent(record: BillingBalance): number {
  if (record.amount <= 0) {
    return 0
  }
  return Math.floor((record.paidAmount / record.amount) * 100)
}

function toAmount(value: string): number {
  return value === '' ? 0 : Number(value)
}

function toDateInput(iso: string): string {
  return iso.slice(0, 10)
}

export interface InvoiceFormValues {
  number: string
  note: string
  partyType: PartyType
  partyId: string
  projectId: string
  dueDate: string
  amount: string
}

export const EMPTY_INVOICE_FORM: InvoiceFormValues = {
  number: '',
  note: '',
  partyType: 'customer',
  partyId: '',
  projectId: '',
  dueDate: '',
  amount: '',
}

export function invoiceFormOf(invoice: Invoice): InvoiceFormValues {
  return {
    number: invoice.number,
    note: invoice.note,
    partyType: invoice.partyType,
    partyId: invoice.partyId,
    projectId: invoice.projectId ?? '',
    dueDate: toDateInput(invoice.dueDate),
    amount: String(invoice.amount),
  }
}

export function toInvoiceRequest(values: InvoiceFormValues): InvoiceRequest {
  return {
    number: values.number.trim(),
    note: values.note.trim(),
    partyType: values.partyType,
    partyId: values.partyId,
    projectId: values.projectId === '' ? undefined : values.projectId,
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export interface ReceivableFormValues {
  customerId: string
  reference: string
  dueDate: string
  amount: string
}

export const EMPTY_RECEIVABLE_FORM: ReceivableFormValues = {
  customerId: '',
  reference: '',
  dueDate: '',
  amount: '',
}

export function receivableFormOf(receivable: Receivable): ReceivableFormValues {
  return {
    customerId: receivable.customerId,
    reference: receivable.reference,
    dueDate: toDateInput(receivable.dueDate),
    amount: String(receivable.amount),
  }
}

export function toReceivableRequest(values: ReceivableFormValues): ReceivableRequest {
  return {
    customerId: values.customerId,
    reference: values.reference.trim(),
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export interface PayableFormValues {
  vendorId: string
  reference: string
  dueDate: string
  amount: string
}

export const EMPTY_PAYABLE_FORM: PayableFormValues = {
  vendorId: '',
  reference: '',
  dueDate: '',
  amount: '',
}

export function payableFormOf(payable: Payable): PayableFormValues {
  return {
    vendorId: payable.vendorId,
    reference: payable.reference,
    dueDate: toDateInput(payable.dueDate),
    amount: String(payable.amount),
  }
}

export function toPayableRequest(values: PayableFormValues): PayableRequest {
  return {
    vendorId: values.vendorId,
    reference: values.reference.trim(),
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export interface BillingTotals {
  count: number
  openCount: number
  outstanding: number
  paidAmount: number
  overdueCount: number
  overdueOutstanding: number
  dueSoonCount: number
  isPartial: boolean
}

// Ringkasan dihitung dari satu halaman daftar karena backend belum punya
// endpoint ringkasan. isPartial menandai daftar yang lebih panjang dari
// halaman yang diambil.
export function billingTotals(items: BillingBalance[], totalItems: number): BillingTotals {
  const open = items.filter((item) => item.status !== 'paid')
  const overdue = open.filter((item) => item.status === 'overdue')

  return {
    count: items.length,
    openCount: open.length,
    outstanding: open.reduce((sum, item) => sum + item.outstanding, 0),
    paidAmount: items.reduce((sum, item) => sum + item.paidAmount, 0),
    overdueCount: overdue.length,
    overdueOutstanding: overdue.reduce((sum, item) => sum + item.outstanding, 0),
    dueSoonCount: open.filter((item) => item.status === 'due').length,
    isPartial: totalItems > items.length,
  }
}
