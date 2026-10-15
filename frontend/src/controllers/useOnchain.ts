import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

import { isOpenRequest, ONCHAIN_LIST_PAGE_SIZE, type OnchainBillingKind } from '../models/onchain'
import { onchainApi, payApi } from '../models/onchainApi'
import { ApiError } from '../shared/apiClient'

// Disimpan di bawah kunci "billing" supaya pembayaran manual ikut menyegarkan
// daftar permintaan on-chain, dan sebaliknya.
const BILLING_KEY = 'billing'
const ONCHAIN_KEY = 'onchain'
const PAY_KEY = 'pay'

const FIVE_MINUTES = 5 * 60_000
const STAFF_POLL_MS = 5_000
const PUBLIC_POLL_MS = 15_000
export const CONFIRM_POLL_MS = 5_000
const MAX_TRANSIENT_RETRIES = 6

export function useOnchainConfig() {
  return useQuery({
    queryKey: [BILLING_KEY, ONCHAIN_KEY, 'config'],
    queryFn: onchainApi.config,
    staleTime: FIVE_MINUTES,
    retry: 0,
  })
}

// Daftar disegarkan tiap 5 detik selama masih ada permintaan yang menunggu,
// supaya status berubah sendiri ketika pembeli selesai membayar. Begitu ada
// permintaan yang baru lunas, rincian, riwayat, dan KPI tagihan ikut diambil
// ulang karena saldonya sudah bergeser.
export function useOnchainRequests(kind: OnchainBillingKind, id: string, page: number, enabled: boolean) {
  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: [BILLING_KEY, ONCHAIN_KEY, 'requests', kind, id, page],
    queryFn: () => onchainApi.list(kind, id, page, ONCHAIN_LIST_PAGE_SIZE),
    enabled,
    refetchInterval: (current) => (current.state.data?.items.some(isOpenRequest) ? STAFF_POLL_MS : false),
    placeholderData: keepPreviousData,
    retry: 0,
  })

  const paidIds = (query.data?.items ?? [])
    .filter((request) => request.status === 'paid')
    .map((request) => request.id)
    .join(',')
  const seenPaidIds = useRef<string | null>(null)

  useEffect(() => {
    if (seenPaidIds.current !== null && seenPaidIds.current !== paidIds) {
      void queryClient.invalidateQueries({
        queryKey: [BILLING_KEY],
        predicate: (candidate) => candidate.queryKey[1] !== ONCHAIN_KEY,
      })
    }
    seenPaidIds.current = paidIds
  }, [paidIds, queryClient])

  return query
}

export function useCreateOnchainRequest(kind: OnchainBillingKind, id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (amount: number) => onchainApi.create(kind, id, { amount }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BILLING_KEY, ONCHAIN_KEY] }),
  })
}

export function useCancelOnchainRequest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (requestId: string) => onchainApi.cancel(requestId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BILLING_KEY, ONCHAIN_KEY] }),
  })
}

function isTransient(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 0 || error.status === 429 || error.status >= 500)
}

// Halaman bayar publik ikut menyegarkan statusnya, misalnya ketika pembeli
// membayar dari HP lalu kembali ke tab ini.
export function usePublicPayment(token: string) {
  return useQuery({
    queryKey: [PAY_KEY, token],
    queryFn: () => payApi.get(token),
    refetchInterval: (current) => (current.state.data?.status === 'pending' ? PUBLIC_POLL_MS : false),
    refetchOnWindowFocus: true,
    retry: (count, error) => isTransient(error) && count < 2,
  })
}

// Konfirmasi dijalankan sebagai query yang diulang selama backend menjawab
// "waiting". Galat sementara (node RPC atau jaringan putus) dicoba lagi;
// penolakan seperti salah alamat langsung ditampilkan.
export function usePaymentConfirmation(token: string, txHash: string | null) {
  return useQuery({
    queryKey: [PAY_KEY, token, 'confirm', txHash],
    queryFn: () => payApi.confirm(token, txHash ?? ''),
    enabled: txHash !== null,
    refetchInterval: (current) => (current.state.data?.verification === 'waiting' ? CONFIRM_POLL_MS : false),
    retry: (count, error) => isTransient(error) && count < MAX_TRANSIENT_RETRIES,
    retryDelay: CONFIRM_POLL_MS,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  })
}
