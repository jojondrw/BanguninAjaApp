import { API_BASE_URL, ApiError, request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  ConfirmOutcome,
  OnchainBillingKind,
  OnchainConfig,
  OnchainRequest,
  OnchainRequestCreate,
  PublicPayment,
} from './onchain'

// Endpoint staf di bawah /billing, memakai token login seperti tagihan lain.
export const onchainApi = {
  config: () => request<OnchainConfig>('/billing/onchain/config'),

  list: (kind: OnchainBillingKind, id: string, page: number, pageSize: number) =>
    request<Page<OnchainRequest>>(`/billing/${kind}/${id}/onchain-requests${toQueryString({ page, pageSize })}`),

  create: (kind: OnchainBillingKind, id: string, body: OnchainRequestCreate) =>
    request<OnchainRequest>(`/billing/${kind}/${id}/onchain-requests`, { method: 'POST', body }),

  cancel: (id: string) => request<OnchainRequest>(`/billing/onchain/requests/${id}/cancel`, { method: 'POST' }),
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

// Halaman bayar dibuka pembeli tanpa akun, jadi dipanggil dengan fetch biasa:
// tanpa token, tanpa cookie, dan tanpa percobaan refresh sesi.
export async function publicCall<T>(path: string, body?: unknown): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: body === undefined ? 'GET' : 'POST',
      credentials: 'omit',
      headers: body === undefined ? {} : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError('network_error', 'Tidak bisa menghubungi server. Periksa koneksi lalu coba lagi.', 0)
  }

  const envelope = (await response.json().catch(() => ({}))) as Envelope<T>
  if (!response.ok || envelope.error || envelope.data === undefined) {
    throw new ApiError(
      envelope.error?.code ?? 'network_error',
      envelope.error?.message ?? 'Tidak bisa menghubungi server. Coba lagi sebentar.',
      response.status,
    )
  }
  return envelope.data
}

export const payApi = {
  get: (token: string) => publicCall<PublicPayment>(`/pay/${encodeURIComponent(token)}`),

  // 202 berarti transaksi masih menunggu blok atau konfirmasi; isinya tetap
  // ConfirmOutcome dengan verification "waiting".
  confirm: (token: string, txHash: string) =>
    publicCall<ConfirmOutcome>(`/pay/${encodeURIComponent(token)}/confirm`, { txHash }),
}
