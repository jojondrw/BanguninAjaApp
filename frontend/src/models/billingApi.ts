import { request } from '../shared/apiClient'
import type {
  BillingFilterOf,
  BillingKind,
  BillingPayment,
  BillingRecordOf,
  BillingRequestOf,
  PaymentRequest,
} from './billing'
import { toQueryString, type Page } from './common'

// Filter boolean seperti recorded dikirim sebagai teks "true" atau "false".
function billingQuery(filter: Record<string, string | number | boolean | undefined>): string {
  const params: Record<string, string | number | undefined> = {}
  for (const [key, value] of Object.entries(filter)) {
    params[key] = typeof value === 'boolean' ? String(value) : value
  }
  return toQueryString(params)
}

// Faktur, piutang, dan utang memakai bentuk endpoint yang sama di bawah
// /billing/<kind>, jadi satu set pemanggil cukup untuk ketiganya.
export const billingApi = {
  list: <K extends BillingKind>(kind: K, filter: BillingFilterOf[K] = {}) =>
    request<Page<BillingRecordOf[K]>>(`/billing/${kind}${billingQuery({ ...filter })}`),

  get: <K extends BillingKind>(kind: K, id: string) => request<BillingRecordOf[K]>(`/billing/${kind}/${id}`),

  create: <K extends BillingKind>(kind: K, body: BillingRequestOf[K]) =>
    request<BillingRecordOf[K]>(`/billing/${kind}`, { method: 'POST', body }),

  update: <K extends BillingKind>(kind: K, id: string, body: BillingRequestOf[K]) =>
    request<BillingRecordOf[K]>(`/billing/${kind}/${id}`, { method: 'PUT', body }),

  remove: (kind: BillingKind, id: string) => request<void>(`/billing/${kind}/${id}`, { method: 'DELETE' }),

  pay: <K extends BillingKind>(kind: K, id: string, body: PaymentRequest) =>
    request<BillingRecordOf[K]>(`/billing/${kind}/${id}/payments`, { method: 'POST', body }),

  // Riwayat pembayaran, terbaru dulu.
  payments: (kind: BillingKind, id: string, page: number, pageSize: number) =>
    request<Page<BillingPayment>>(`/billing/${kind}/${id}/payments${toQueryString({ page, pageSize })}`),
}
