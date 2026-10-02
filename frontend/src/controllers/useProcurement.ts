import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  toPurchaseOrderRequest,
  toVendorRequest,
  type PurchaseOrderFilter,
  type PurchaseOrderFormValues,
  type PurchaseRequestFilter,
  type PurchaseRequestStatus,
  type VendorFilter,
  type VendorFormValues,
} from '../models/procurement'
import { procurementApi } from '../models/procurementApi'

const ONE_MINUTE = 60_000
const MAX_PAGE_SIZE = 100
const PROCUREMENT_KEY = 'procurement'

const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

export function useVendors(filter: VendorFilter = {}) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'vendors', filter],
    queryFn: () => procurementApi.vendors(filter),
    ...PAGED_QUERY,
  })
}

// Dipakai untuk pilihan vendor di formulir dan untuk menampilkan nama vendor
// di tabel pesanan, karena respons pesanan hanya membawa vendorId.
export function useVendorOptions() {
  return useVendors({ pageSize: MAX_PAGE_SIZE })
}

export function useActiveVendorCount() {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'vendor-count', 'active'],
    queryFn: () => procurementApi.vendors({ active: true, pageSize: 1 }),
    select: (page) => page.totalItems,
    ...DATA_QUERY,
  })
}

export function usePurchaseRequests(filter: PurchaseRequestFilter = {}) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'purchase-requests', filter],
    queryFn: () => procurementApi.purchaseRequests(filter),
    ...PAGED_QUERY,
  })
}

export function usePurchaseRequestCount(status: PurchaseRequestStatus) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'purchase-request-count', status],
    queryFn: () => procurementApi.purchaseRequests({ status, pageSize: 1 }),
    select: (page) => page.totalItems,
    ...DATA_QUERY,
  })
}

export function usePurchaseOrders(filter: PurchaseOrderFilter = {}) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'purchase-orders', filter],
    queryFn: () => procurementApi.purchaseOrders(filter),
    ...PAGED_QUERY,
  })
}

export function usePurchaseOrder(id: string | null) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'purchase-order', id],
    queryFn: () => procurementApi.purchaseOrder(id ?? ''),
    enabled: id !== null,
    ...DATA_QUERY,
  })
}

export function useMaterials() {
  return useQuery({
    queryKey: ['inventory', 'materials', 'options'],
    queryFn: () => procurementApi.materials(MAX_PAGE_SIZE),
    ...DATA_QUERY,
  })
}

export function useUnitsOfMeasure() {
  return useQuery({
    queryKey: ['master', 'units-of-measure', 'options'],
    queryFn: () => procurementApi.unitsOfMeasure(MAX_PAGE_SIZE),
    staleTime: Infinity,
    retry: 0,
  })
}

function useProcurementMutation<TVariables, TData>(mutationFn: (variables: TVariables) => Promise<TData>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROCUREMENT_KEY] }),
  })
}

export function useCreateVendor() {
  return useProcurementMutation((values: VendorFormValues) =>
    procurementApi.createVendor(toVendorRequest(values)),
  )
}

export function useCreatePurchaseOrder() {
  return useProcurementMutation((values: PurchaseOrderFormValues) =>
    procurementApi.createPurchaseOrder(toPurchaseOrderRequest(values)),
  )
}

export function useUpdatePurchaseOrderStatus() {
  return useProcurementMutation(({ id, status }: { id: string; status: 'sent' | 'cancelled' }) =>
    procurementApi.updatePurchaseOrderStatus(id, status),
  )
}
