import { request } from '../shared/apiClient'
import type {
  BillingFilterOf,
  BillingKind,
  BillingRecordOf,
  BillingRequestOf,
  PaymentRequest,
} from './billing'
import { toQueryString, type Page } from './common'

// Faktur, piutang, dan utang memakai bentuk endpoint yang sama di bawah
// /billing/<kind>, jadi satu set pemanggil cukup untuk ketiganya.
export const billingApi = {
  list: <K extends BillingKind>(kind: K, filter: BillingFilterOf[K] = {}) =>
    request<Page<BillingRecordOf[K]>>(`/billing/${kind}${toQueryString({ ...filter })}`),

  get: <K extends BillingKind>(kind: K, id: string) => request<BillingRecordOf[K]>(`/billing/${kind}/${id}`),

  create: <K extends BillingKind>(kind: K, body: BillingRequestOf[K]) =>
    request<BillingRecordOf[K]>(`/billing/${kind}`, { method: 'POST', body }),

  update: <K extends BillingKind>(kind: K, id: string, body: BillingRequestOf[K]) =>
    request<BillingRecordOf[K]>(`/billing/${kind}/${id}`, { method: 'PUT', body }),

  remove: (kind: BillingKind, id: string) => request<void>(`/billing/${kind}/${id}`, { method: 'DELETE' }),

  pay: <K extends BillingKind>(kind: K, id: string, body: PaymentRequest) =>
    request<BillingRecordOf[K]>(`/billing/${kind}/${id}/payments`, { method: 'POST', body }),
}
