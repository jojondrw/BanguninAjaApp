import { toApiDate } from '../shared/localDate'

// Tiga jenis catatan di backend/internal/billing. Nilainya sekaligus potongan
// path endpoint, jadi /billing/<kind>.
export type BillingKind = 'invoices' | 'receivables' | 'payables'
export type BillingStatus = 'not_due' | 'due' | 'paid' | 'overdue'
export type PartyType = 'customer' | 'vendor'
// Mengikuti oneof di PaymentRequest backend. Daftarnya sengaja terbuka untuk
// metode baru nanti.
export type PaymentMethod = 'transfer' | 'tunai' | 'cek' | 'lainnya'

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
  // Tanggal pembayaran terakhir dari riwayat, null kalau belum pernah dibayar.
  lastPaidAt: string | null
  createdAt: string
  updatedAt: string
}

// receivableId terisi kalau faktur ini sudah dicatat sebagai piutang. Satu
// faktur hanya boleh menjadi satu piutang. partyName digabung backend dari
// pelanggan atau vendor sesuai partyType.
export interface Invoice extends BillingRecord {
  number: string
  note: string
  partyType: PartyType
  partyId: string
  partyName: string | null
  projectId: string | null
  projectName: string | null
  receivableId: string | null
  receivableReference: string | null
}

// Nama dan nomor tautan sudah digabung backend, jadi tabel tidak perlu
// mencari pelanggan, proyek, kontrak, atau faktur sendiri.
export interface Receivable extends BillingRecord {
  customerId: string
  customerName: string
  projectId: string | null
  projectName: string | null
  contractId: string | null
  contractNumber: string | null
  invoiceId: string | null
  invoiceNumber: string | null
  reference: string
}

export interface Payable extends BillingRecord {
  vendorId: string
  vendorName: string
  projectId: string | null
  projectName: string | null
  purchaseOrderId: string | null
  purchaseOrderNumber: string | null
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
  // false hanya menampilkan faktur yang belum dicatat sebagai piutang.
  recorded?: boolean
}

export interface ReceivableFilter extends PageQuery {
  customerId?: string
  projectId?: string
  contractId?: string
}

export interface PayableFilter extends PageQuery {
  vendorId?: string
  projectId?: string
  purchaseOrderId?: string
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
  projectId?: string
  contractId?: string
  invoiceId?: string
  reference: string
  dueDate: string
  amount: number
}

export interface PayableRequest {
  vendorId: string
  projectId?: string
  purchaseOrderId?: string
  reference: string
  dueDate: string
  amount: number
}

// Tanggal kosong berarti hari ini (WIB) dan metode kosong berarti transfer.
export interface PaymentRequest {
  amount: number
  paidAt?: string
  method?: PaymentMethod
  reference?: string
  note?: string
}

// Satu baris riwayat. Pembayaran faktur yang tertaut piutang tercatat sekali
// dengan invoiceId dan receivableId sekaligus.
export interface BillingPayment {
  id: string
  invoiceId: string | null
  invoiceNumber: string | null
  receivableId: string | null
  receivableReference: string | null
  payableId: string | null
  amount: number
  paidAt: string
  method: PaymentMethod
  reference: string
  note: string
  createdBy: string | null
  createdByName: string | null
  createdAt: string
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
export const PAYMENT_REFERENCE_MAX_LENGTH = 60
export const PAYMENT_NOTE_MAX_LENGTH = 200

export const PAYMENT_METHODS: PaymentMethod[] = ['transfer', 'tunai', 'cek', 'lainnya']

export const PAYMENT_METHOD_LABEL: Record<PaymentMethod, string> = {
  transfer: 'Transfer bank',
  tunai: 'Tunai',
  cek: 'Cek atau giro',
  lainnya: 'Lainnya',
}

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

function optionalId(value: string): string | undefined {
  return value === '' ? undefined : value
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
    projectId: optionalId(values.projectId),
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export interface ReceivableFormValues {
  customerId: string
  projectId: string
  contractId: string
  invoiceId: string
  reference: string
  dueDate: string
  amount: string
}

export const EMPTY_RECEIVABLE_FORM: ReceivableFormValues = {
  customerId: '',
  projectId: '',
  contractId: '',
  invoiceId: '',
  reference: '',
  dueDate: '',
  amount: '',
}

export function receivableFormOf(receivable: Receivable): ReceivableFormValues {
  return {
    customerId: receivable.customerId,
    projectId: receivable.projectId ?? '',
    contractId: receivable.contractId ?? '',
    invoiceId: receivable.invoiceId ?? '',
    reference: receivable.reference,
    dueDate: toDateInput(receivable.dueDate),
    amount: String(receivable.amount),
  }
}

// Isian awal "Catat sebagai piutang" dari faktur pelanggan. Nominal memakai
// nilai penuh faktur, karena backend menyalin total yang sudah dibayar ke
// piutang supaya sisa keduanya sama.
export function receivableFormFromInvoice(invoice: Invoice): ReceivableFormValues {
  return {
    customerId: invoice.partyId,
    projectId: invoice.projectId ?? '',
    contractId: '',
    invoiceId: invoice.id,
    reference: invoice.number,
    dueDate: toDateInput(invoice.dueDate),
    amount: String(invoice.amount),
  }
}

export function toReceivableRequest(values: ReceivableFormValues): ReceivableRequest {
  return {
    customerId: values.customerId,
    projectId: optionalId(values.projectId),
    contractId: optionalId(values.contractId),
    invoiceId: optionalId(values.invoiceId),
    reference: values.reference.trim(),
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export interface PayableFormValues {
  vendorId: string
  projectId: string
  purchaseOrderId: string
  reference: string
  dueDate: string
  amount: string
}

export const EMPTY_PAYABLE_FORM: PayableFormValues = {
  vendorId: '',
  projectId: '',
  purchaseOrderId: '',
  reference: '',
  dueDate: '',
  amount: '',
}

export function payableFormOf(payable: Payable): PayableFormValues {
  return {
    vendorId: payable.vendorId,
    projectId: payable.projectId ?? '',
    purchaseOrderId: payable.purchaseOrderId ?? '',
    reference: payable.reference,
    dueDate: toDateInput(payable.dueDate),
    amount: String(payable.amount),
  }
}

export function toPayableRequest(values: PayableFormValues): PayableRequest {
  return {
    vendorId: values.vendorId,
    projectId: optionalId(values.projectId),
    purchaseOrderId: optionalId(values.purchaseOrderId),
    reference: values.reference.trim(),
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export interface PaymentFormValues {
  amount: string
  paidAt: string
  method: PaymentMethod
  reference: string
  note: string
}

export function paymentFormOf(record: BillingBalance, today: string): PaymentFormValues {
  return { amount: String(record.outstanding), paidAt: today, method: 'transfer', reference: '', note: '' }
}

export function toPaymentRequest(values: PaymentFormValues): PaymentRequest {
  return {
    amount: toAmount(values.amount),
    paidAt: toApiDate(values.paidAt),
    method: values.method,
    reference: values.reference.trim(),
    note: values.note.trim(),
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
