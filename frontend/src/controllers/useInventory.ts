import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  toMaterialRequest,
  toStockMovementRequest,
  toWarehouseRequest,
  type MaterialFilter,
  type MaterialFormValues,
  type MovementFormValues,
  type StockFilter,
  type StockMovementFilter,
  type WarehouseFilter,
  type WarehouseFormValues,
} from '../models/inventory'
import { assetApi, inventoryApi, unitOfMeasureApi } from '../models/inventoryApi'

const ONE_MINUTE = 60_000
const TEN_MINUTES = 600_000
const MAX_PAGE_SIZE = 100

const INVENTORY_KEY = 'inventory'
const MATERIALS_KEY = [INVENTORY_KEY, 'materials'] as const
const LOW_STOCK_KEY = [INVENTORY_KEY, 'low-stock'] as const
const WAREHOUSES_KEY = [INVENTORY_KEY, 'warehouses'] as const
const STOCKS_KEY = [INVENTORY_KEY, 'stocks'] as const
const MOVEMENTS_KEY = [INVENTORY_KEY, 'stock-movements'] as const

const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

export function useMaterials(filter: MaterialFilter) {
  return useQuery({
    queryKey: [...MATERIALS_KEY, filter],
    queryFn: () => inventoryApi.materials(filter),
    ...PAGED_QUERY,
  })
}

// Dipakai untuk pilihan di formulir mutasi dan untuk menerjemahkan materialId
// di tabel stok serta mutasi. Batasnya 100, ukuran halaman terbesar backend.
export function useMaterialOptions() {
  return useMaterials({ pageSize: MAX_PAGE_SIZE })
}

export function useLowStockMaterials(page: number) {
  return useQuery({
    queryKey: [...LOW_STOCK_KEY, page],
    queryFn: () => inventoryApi.lowStockMaterials({ page, pageSize: 10 }),
    ...PAGED_QUERY,
  })
}

export function useCreateMaterial() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: MaterialFormValues) => inventoryApi.createMaterial(toMaterialRequest(values)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: MATERIALS_KEY })
      void queryClient.invalidateQueries({ queryKey: LOW_STOCK_KEY })
    },
  })
}

export function useWarehouses(filter: WarehouseFilter) {
  return useQuery({
    queryKey: [...WAREHOUSES_KEY, filter],
    queryFn: () => inventoryApi.warehouses(filter),
    ...PAGED_QUERY,
  })
}

export function useWarehouseOptions() {
  return useWarehouses({ pageSize: MAX_PAGE_SIZE })
}

export function useCreateWarehouse() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: WarehouseFormValues) => inventoryApi.createWarehouse(toWarehouseRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: WAREHOUSES_KEY }),
  })
}

export function useStocks(filter: StockFilter) {
  return useQuery({
    queryKey: [...STOCKS_KEY, filter],
    queryFn: () => inventoryApi.stocks(filter),
    ...PAGED_QUERY,
  })
}

export function useStockMovements(filter: StockMovementFilter) {
  return useQuery({
    queryKey: [...MOVEMENTS_KEY, filter],
    queryFn: () => inventoryApi.stockMovements(filter),
    ...PAGED_QUERY,
  })
}

// Hitungan dari totalItems, jadi server yang menghitung, bukan peramban.
export function useStockMovementCount(filter: Omit<StockMovementFilter, 'page' | 'pageSize'>) {
  return useQuery({
    queryKey: [...MOVEMENTS_KEY, 'count', filter],
    queryFn: () => inventoryApi.stockMovements({ ...filter, pageSize: 1 }),
    select: (page) => page.totalItems,
    ...DATA_QUERY,
  })
}

// Mutasi mengubah stok, daftar stok menipis, dan riwayat mutasi sekaligus.
export function useRecordStockMovement() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: MovementFormValues) =>
      inventoryApi.recordStockMovement(toStockMovementRequest(values)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: STOCKS_KEY })
      void queryClient.invalidateQueries({ queryKey: MOVEMENTS_KEY })
      void queryClient.invalidateQueries({ queryKey: LOW_STOCK_KEY })
    },
  })
}

// Satuan adalah data master yang diisi seed dan jarang berubah.
export function useUnitsOfMeasure() {
  return useQuery({
    queryKey: ['master', 'units-of-measure'],
    queryFn: () => unitOfMeasureApi.list({ pageSize: MAX_PAGE_SIZE }),
    staleTime: TEN_MINUTES,
    retry: 0,
  })
}

export function useAssets(page: number) {
  return useQuery({
    queryKey: ['assets', page],
    queryFn: () => assetApi.assets({ page, pageSize: 10 }),
    ...PAGED_QUERY,
  })
}

export function useEquipment(page: number) {
  return useQuery({
    queryKey: ['equipment', page],
    queryFn: () => assetApi.equipment({ page, pageSize: 10 }),
    ...PAGED_QUERY,
  })
}
