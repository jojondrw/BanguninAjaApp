import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

import { gatewayApi, qrisPayApi } from '../models/gatewayApi'
import { ONCHAIN_LIST_PAGE_SIZE, type OnchainBillingKind } from '../models/onchain'
import { ApiError } from '../shared/apiClient'

// Di bawah kunci "billing" supaya pembayaran manual dan QRIS saling
// menyegarkan.
const BILLING_KEY = 'billing'
const GATEWAY_KEY = 'gateway'
const QRIS_PAY_KEY = 'pay-qris'

const FIVE_MINUTES = 5 * 60_000
const STAFF_POLL_MS = 4_000
export const PUBLIC_POLL_MS = 3_000

export function useGatewayConfig() {
  return useQuery({
    queryKey: [BILLING_KEY, GATEWAY_KEY, 'config'],
    queryFn: gatewayApi.config,
    staleTime: FIVE_MINUTES,
    retry: 0,
  })
}

// Selama ada QRIS yang menunggu, daftar disegarkan tiap beberapa detik. Begitu
// ada yang lunas, rincian, riwayat pembayaran, dan ringkasan ikut diambil ulang.
export function useGatewayRequests(kind: OnchainBillingKind, id: string, page: number, enabled: boolean) {
  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: [BILLING_KEY, GATEWAY_KEY, 'requests', kind, id, page],
    queryFn: () => gatewayApi.list(kind, id, page, ONCHAIN_LIST_PAGE_SIZE),
    enabled,
    refetchInterval: (current) => (current.state.data?.items.some((item) => item.status === 'pending') ? STAFF_POLL_MS : false),
    placeholderData: keepPreviousData,
    retry: 0,
  })

  const paidIds = (query.data?.items ?? [])
    .filter((item) => item.status === 'paid')
    .map((item) => item.id)
    .join(',')
  const seenPaidIds = useRef<string | null>(null)

  useEffect(() => {
    if (seenPaidIds.current !== null && seenPaidIds.current !== paidIds) {
      void queryClient.invalidateQueries({
        queryKey: [BILLING_KEY],
        predicate: (candidate) => candidate.queryKey[1] !== GATEWAY_KEY,
      })
    }
    seenPaidIds.current = paidIds
  }, [paidIds, queryClient])

  return query
}

export function useCreateGatewayRequest(kind: OnchainBillingKind, id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (amount: number) => gatewayApi.create(kind, id, amount),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BILLING_KEY, GATEWAY_KEY] }),
  })
}

export function useCancelGatewayRequest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (requestId: string) => gatewayApi.cancel(requestId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BILLING_KEY, GATEWAY_KEY] }),
  })
}

function isTransient(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 0 || error.status === 429 || error.status >= 500)
}

// Halaman bayar menanyakan status lewat refresh tiap tiga detik selama masih
// menunggu, sehingga layar berubah jadi lunas tanpa pembeli menekan apa pun.
export function useQrisPayment(token: string) {
  return useQuery({
    queryKey: [QRIS_PAY_KEY, token],
    queryFn: () => qrisPayApi.refresh(token),
    refetchInterval: (current) => (current.state.data?.status === 'pending' ? PUBLIC_POLL_MS : false),
    refetchOnWindowFocus: true,
    retry: (count, error) => isTransient(error) && count < 2,
  })
}

export function useSimulateQrisPayment(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => qrisPayApi.simulate(token),
    onSuccess: (payment) => queryClient.setQueryData([QRIS_PAY_KEY, token], payment),
  })
}
