import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  billingTotals,
  toInvoiceRequest,
  toPayableRequest,
  toReceivableRequest,
  type BillingFilterOf,
  type BillingKind,
  type InvoiceFormValues,
  type PayableFormValues,
  type ReceivableFormValues,
} from '../models/billing'
import { billingApi } from '../models/billingApi'

const ONE_MINUTE = 60_000
const MAX_PAGE_SIZE = 100
const BILLING_KEY = 'billing'

const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

const DETAIL = 'detail'

export function useBillingList<K extends BillingKind>(kind: K, filter: BillingFilterOf[K]) {
  return useQuery({
    queryKey: [BILLING_KEY, kind, 'list', filter],
    queryFn: () => billingApi.list(kind, filter),
    ...PAGED_QUERY,
  })
}

export function useBillingRecord<K extends BillingKind>(kind: K, id: string | null) {
  return useQuery({
    queryKey: [BILLING_KEY, kind, DETAIL, id],
    queryFn: () => billingApi.get(kind, id ?? ''),
    enabled: id !== null,
    ...DATA_QUERY,
  })
}

// Backend belum punya endpoint ringkasan, jadi KPI dihitung dari 100 catatan
// pertama tiap jenis. Urutan backend menurut jatuh tempo terlama dulu, jadi
// yang paling mendesak selalu ikut terhitung.
export function useBillingTotals(kind: BillingKind) {
  return useQuery({
    queryKey: [BILLING_KEY, kind, 'totals'],
    queryFn: () => billingApi.list(kind, { pageSize: MAX_PAGE_SIZE }),
    select: (page) => billingTotals(page.items, page.totalItems),
    ...DATA_QUERY,
  })
}

// Pembayaran dan perubahan nominal menggeser status, sisa, dan KPI sekaligus,
// jadi semua query tagihan dibatalkan setelah mutasi apa pun.
function useBillingMutation<TVariables, TData>(mutationFn: (variables: TVariables) => Promise<TData>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [BILLING_KEY] }),
  })
}

export function useCreateInvoice() {
  return useBillingMutation((values: InvoiceFormValues) =>
    billingApi.create('invoices', toInvoiceRequest(values)),
  )
}

export function useUpdateInvoice() {
  return useBillingMutation(({ id, values }: { id: string; values: InvoiceFormValues }) =>
    billingApi.update('invoices', id, toInvoiceRequest(values)),
  )
}

export function useCreateReceivable() {
  return useBillingMutation((values: ReceivableFormValues) =>
    billingApi.create('receivables', toReceivableRequest(values)),
  )
}

export function useUpdateReceivable() {
  return useBillingMutation(({ id, values }: { id: string; values: ReceivableFormValues }) =>
    billingApi.update('receivables', id, toReceivableRequest(values)),
  )
}

export function useCreatePayable() {
  return useBillingMutation((values: PayableFormValues) =>
    billingApi.create('payables', toPayableRequest(values)),
  )
}

export function useUpdatePayable() {
  return useBillingMutation(({ id, values }: { id: string; values: PayableFormValues }) =>
    billingApi.update('payables', id, toPayableRequest(values)),
  )
}

export function usePayBilling(kind: BillingKind) {
  return useBillingMutation(({ id, amount }: { id: string; amount: number }) =>
    billingApi.pay(kind, id, { amount }),
  )
}

// Rincian tidak ikut diambil ulang setelah hapus, karena catatannya sudah
// tidak ada dan hanya akan berakhir 404 sebelum panelnya tertutup.
export function useDeleteBilling(kind: BillingKind) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => billingApi.remove(kind, id),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: [BILLING_KEY],
        predicate: (query) => query.queryKey[2] !== DETAIL,
      }),
  })
}
