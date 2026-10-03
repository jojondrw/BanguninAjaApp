import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  billingTotals,
  toInvoiceRequest,
  toPayableRequest,
  toPaymentRequest,
  toReceivableRequest,
  type BillingFilterOf,
  type BillingKind,
  type InvoiceFormValues,
  type PayableFormValues,
  type PaymentFormValues,
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
const PAYMENTS = 'payments'

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

export function useBillingPayments(kind: BillingKind, id: string, page: number, pageSize: number) {
  return useQuery({
    queryKey: [BILLING_KEY, kind, PAYMENTS, id, page, pageSize],
    queryFn: () => billingApi.payments(kind, id, page, pageSize),
    ...PAGED_QUERY,
  })
}

// Pilihan faktur di formulir piutang: faktur pelanggan yang belum dicatat
// sebagai piutang, dipersempit ke satu pelanggan kalau sudah dipilih.
export function useUnrecordedInvoices(customerId: string) {
  return useBillingList('invoices', {
    partyType: 'customer',
    partyId: customerId === '' ? undefined : customerId,
    recorded: false,
    pageSize: MAX_PAGE_SIZE,
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

// Pembayaran faktur yang tertaut piutang ikut menggeser piutangnya (dan
// sebaliknya), jadi semua query tagihan dibatalkan, bukan hanya jenis ini.
export function usePayBilling(kind: BillingKind) {
  return useBillingMutation(({ id, values }: { id: string; values: PaymentFormValues }) =>
    billingApi.pay(kind, id, toPaymentRequest(values)),
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
        predicate: (query) => query.queryKey[2] !== DETAIL && query.queryKey[2] !== PAYMENTS,
      }),
  })
}
