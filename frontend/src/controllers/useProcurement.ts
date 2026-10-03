import { keepPreviousData, type QueryClient, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import type { StockMovementFilter, StockMovementRequest } from '../models/inventory'
import { inventoryApi } from '../models/inventoryApi'
import {
  receiptNeedsMovementCheck,
  toGoodsReceiptRequest,
  toInputDate,
  toPurchaseOrderRequest,
  toPurchaseRequestRequest,
  toVendorRequest,
  type GoodsReceiptDetail,
  type GoodsReceiptFilter,
  type GoodsReceiptFormValues,
  type PurchaseOrderFilter,
  type PurchaseOrderFormValues,
  type PurchaseOrderStatusChange,
  type PurchaseRequestFilter,
  type PurchaseRequestFormValues,
  type PurchaseRequestStatus,
  type PurchaseRequestStatusChange,
  type ReceiptLineValues,
  type VendorFilter,
  type VendorFormValues,
} from '../models/procurement'
import { procurementApi } from '../models/procurementApi'

const ONE_MINUTE = 60_000
const MAX_PAGE_SIZE = 100
const PROCUREMENT_KEY = 'procurement'

// Kunci yang sama dengan controllers/useInventory.ts, supaya stok masuk dari
// penerimaan barang langsung terlihat di halaman Persediaan.
const INVENTORY_KEY = 'inventory'
const WAREHOUSES_KEY = [INVENTORY_KEY, 'warehouses'] as const
const STOCKS_KEY = [INVENTORY_KEY, 'stocks'] as const
const MOVEMENTS_KEY = [INVENTORY_KEY, 'stock-movements'] as const
const LOW_STOCK_KEY = [INVENTORY_KEY, 'low-stock'] as const

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

export function usePurchaseRequest(id: string | null) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'purchase-request', id],
    queryFn: () => procurementApi.purchaseRequest(id ?? ''),
    enabled: id !== null,
    ...DATA_QUERY,
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

// Dipakai untuk nomor pesanan di tabel penerimaan dan pilihan pesanan di
// penyaring, karena respons penerimaan hanya membawa purchaseOrderId.
export function usePurchaseOrderOptions() {
  return usePurchaseOrders({ pageSize: MAX_PAGE_SIZE })
}

export function usePurchaseOrder(id: string | null) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'purchase-order', id],
    queryFn: () => procurementApi.purchaseOrder(id ?? ''),
    enabled: id !== null,
    ...DATA_QUERY,
  })
}

export function useGoodsReceipts(filter: GoodsReceiptFilter = {}) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'goods-receipts', filter],
    queryFn: () => procurementApi.goodsReceipts(filter),
    ...PAGED_QUERY,
  })
}

export function useGoodsReceipt(id: string | null) {
  return useQuery({
    queryKey: [PROCUREMENT_KEY, 'goods-receipt', id],
    queryFn: () => procurementApi.goodsReceipt(id ?? ''),
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

// Kunci dan filternya sama dengan useWarehouseOptions di useInventory, jadi
// datanya berbagi cache dengan halaman Persediaan.
export function useWarehouseOptions() {
  const filter = { pageSize: MAX_PAGE_SIZE }

  return useQuery({
    queryKey: [...WAREHOUSES_KEY, filter],
    queryFn: () => inventoryApi.warehouses(filter),
    ...DATA_QUERY,
  })
}

// Mutasi masuk di gudang dan tanggal penerimaan. Yang referensinya nomor
// penerimaan dipilah di receiptStockLines. Penerimaan yang semua barisnya
// sudah dicatat otomatis tidak perlu mengambil mutasi sama sekali.
export function useReceiptStockMovements(receipt: GoodsReceiptDetail | undefined) {
  const day = toInputDate(receipt?.date ?? null)
  const filter: StockMovementFilter = {
    type: 'in',
    warehouseId: receipt?.warehouseId,
    dateFrom: day,
    dateTo: day,
    pageSize: MAX_PAGE_SIZE,
  }

  return useQuery({
    queryKey: [...MOVEMENTS_KEY, filter],
    queryFn: () => inventoryApi.stockMovements(filter),
    enabled: receipt !== undefined && receiptNeedsMovementCheck(receipt),
    ...DATA_QUERY,
  })
}

function invalidateStock(queryClient: QueryClient) {
  return Promise.all(
    [STOCKS_KEY, MOVEMENTS_KEY, LOW_STOCK_KEY].map((queryKey) => queryClient.invalidateQueries({ queryKey })),
  )
}

function useProcurementMutation<TVariables, TData>(mutationFn: (variables: TVariables) => Promise<TData>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROCUREMENT_KEY] }),
  })
}

// Rincian dokumen yang baru dihapus tidak diambil ulang, karena jawabannya
// pasti 404 dan sempat tampil sebagai gagal memuat sebelum rinciannya ditutup.
function useProcurementDelete(mutationFn: (id: string) => Promise<void>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: (_data, id) =>
      queryClient.invalidateQueries({
        queryKey: [PROCUREMENT_KEY],
        predicate: (query) => query.queryKey[2] !== id,
      }),
  })
}

export function useCreateVendor() {
  return useProcurementMutation((values: VendorFormValues) =>
    procurementApi.createVendor(toVendorRequest(values)),
  )
}

export function useUpdateVendor() {
  return useProcurementMutation(({ id, values, active }: { id: string; values: VendorFormValues; active: boolean }) =>
    procurementApi.updateVendor(id, toVendorRequest(values, active)),
  )
}

export function useDeleteVendor() {
  return useProcurementDelete(procurementApi.deleteVendor)
}

export function useCreatePurchaseRequest() {
  return useProcurementMutation((values: PurchaseRequestFormValues) =>
    procurementApi.createPurchaseRequest(toPurchaseRequestRequest(values)),
  )
}

export function useUpdatePurchaseRequest() {
  return useProcurementMutation(({ id, values }: { id: string; values: PurchaseRequestFormValues }) =>
    procurementApi.updatePurchaseRequest(id, toPurchaseRequestRequest(values)),
  )
}

export function useUpdatePurchaseRequestStatus() {
  return useProcurementMutation(({ id, status }: { id: string; status: PurchaseRequestStatusChange }) =>
    procurementApi.updatePurchaseRequestStatus(id, status),
  )
}

export function useDeletePurchaseRequest() {
  return useProcurementDelete(procurementApi.deletePurchaseRequest)
}

export function useCreatePurchaseOrder() {
  return useProcurementMutation((values: PurchaseOrderFormValues) =>
    procurementApi.createPurchaseOrder(toPurchaseOrderRequest(values)),
  )
}

export function useUpdatePurchaseOrder() {
  return useProcurementMutation(({ id, values }: { id: string; values: PurchaseOrderFormValues }) =>
    procurementApi.updatePurchaseOrder(id, toPurchaseOrderRequest(values)),
  )
}

export function useUpdatePurchaseOrderStatus() {
  return useProcurementMutation(({ id, status }: { id: string; status: PurchaseOrderStatusChange }) =>
    procurementApi.updatePurchaseOrderStatus(id, status),
  )
}

export function useDeletePurchaseOrder() {
  return useProcurementDelete(procurementApi.deletePurchaseOrder)
}

interface GoodsReceiptVariables {
  purchaseOrderId: string
  values: GoodsReceiptFormValues
  lines: { purchaseOrderItemId: string; values: ReceiptLineValues }[]
}

// Pesanan diambil ulang sesudah penerimaan tersimpan supaya status barunya
// (diterima sebagian atau selesai) datang dari server. Kalau pengambilan itu
// gagal, penerimaannya tetap dianggap berhasil.
//
// Backend menambah stok gudang dalam transaksi yang sama dengan penerimaan,
// jadi data Persediaan (stok, mutasi, stok menipis) ikut diambil ulang.
//
// Pengambilan ulang daftar tidak ditunggu. Pesanan yang jadi Selesai menutup
// formulir penerimaan, dan callback onSuccess milik mutate tidak dipanggil
// kalau komponennya sudah hilang lebih dulu.
export function useRecordGoodsReceipt() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ purchaseOrderId, values, lines }: GoodsReceiptVariables) => {
      const receipt = await procurementApi.recordGoodsReceipt(toGoodsReceiptRequest(purchaseOrderId, values, lines))
      const order = await procurementApi.purchaseOrder(purchaseOrderId).catch(() => null)
      return { receipt, order }
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [PROCUREMENT_KEY] })
      void invalidateStock(queryClient)
    },
  })
}

// Untuk baris yang tidak dicatat otomatis (penerimaan lama atau satuan beda).
// Dicatat satu per satu dan berhenti di kegagalan pertama. Daftar mutasi tetap
// diambil ulang walaupun gagal di tengah, jadi baris yang sudah masuk stok
// tidak dicatat dua kali saat diulang.
export function useRecordReceiptStock() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (requests: StockMovementRequest[]) => {
      for (const body of requests) {
        await inventoryApi.recordStockMovement(body)
      }
      return requests.length
    },
    onSettled: () => invalidateStock(queryClient),
  })
}
