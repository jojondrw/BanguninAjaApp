import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { GatewayConfig, GatewayPublicPayment, GatewayRequest } from './gateway'
import type { OnchainBillingKind } from './onchain'
import { publicCall } from './onchainApi'

// Endpoint staf di bawah /billing, memakai token login seperti tagihan lain.
export const gatewayApi = {
  config: () => request<GatewayConfig>('/billing/gateway/config'),

  list: (kind: OnchainBillingKind, id: string, page: number, pageSize: number) =>
    request<Page<GatewayRequest>>(`/billing/${kind}/${id}/gateway-requests${toQueryString({ page, pageSize })}`),

  create: (kind: OnchainBillingKind, id: string, amount: number) =>
    request<GatewayRequest>(`/billing/${kind}/${id}/gateway-requests`, { method: 'POST', body: { amount } }),

  cancel: (id: string) => request<GatewayRequest>(`/billing/gateway/requests/${id}/cancel`, { method: 'POST' }),
}

// Halaman bayar QRIS dibuka pembeli tanpa akun. refresh meminta server
// menanyakan status ke gerbang pembayaran, jadi tetap jalan walau webhook
// belum sampai.
export const qrisPayApi = {
  get: (token: string) => publicCall<GatewayPublicPayment>(`/pay/qris/${encodeURIComponent(token)}`),
  refresh: (token: string) => publicCall<GatewayPublicPayment>(`/pay/qris/${encodeURIComponent(token)}/refresh`, {}),
  simulate: (token: string) => publicCall<GatewayPublicPayment>(`/pay/qris/${encodeURIComponent(token)}/simulate`, {}),
}
