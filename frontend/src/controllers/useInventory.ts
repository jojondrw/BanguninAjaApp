import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  toAssetRequest,
  toEquipmentRequest,
  toMaterialRequest,
  toStockMovementRequest,
  toWarehouseRequest,
  type Asset,
  type AssetFilter,
  type AssetFormValues,
  type Equipment,
  type EquipmentFilter,
  type EquipmentFormValues,
  type Material,
  type MaterialFilter,
  type MaterialFormValues,
  type MovementFormValues,
  type StockFilter,
  type StockMovementFilter,
  type Warehouse,
  type WarehouseFilter,
  type WarehouseFormValues,
} from '../models/inventory'
import { assetApi, inventoryApi, unitOfMeasureApi } from '../models/inventoryApi'

const ONE_MINUTE = 60_000
const TEN_MINUTES = 600_000
const MAX_PAGE_SIZE = 100

// Kunci ini juga dipakai controllers/useProcurement.ts supaya penerimaan barang
// menyegarkan stok di halaman Inventaris. Jangan diganti sepihak.
const INVENTORY_KEY = 'inventory'
const MATERIALS_KEY = [INVENTORY_KEY, 'materials'] as const
const LOW_STOCK_KEY = [INVENTORY_KEY, 'low-stock'] as const
const WAREHOUSES_KEY = [INVENTORY_KEY, 'warehouses'] as const
const STOCKS_KEY = [INVENTORY_KEY, 'stocks'] as const
const MOVEMENTS_KEY = [INVENTORY_KEY, 'stock-movements'] as const
const ASSETS_KEY = ['assets'] as const
const EQUIPMENT_KEY = ['equipment'] as const

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

export function useLowStockMaterials(page: number) {
  return useQuery({
    queryKey: [...LOW_STOCK_KEY, page],
    queryFn: () => inventoryApi.lowStockMaterials({ page, pageSize: 10 }),
    ...PAGED_QUERY,
  })
}

// Nama material dan gudang ikut tampil di tabel stok, mutasi, dan stok
// menipis, jadi semua data inventaris disegarkan setelah katalog berubah.
function useInvalidateInventory() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: [INVENTORY_KEY] })
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

export interface MaterialUpdate {
  material: Material
  values: MaterialFormValues
}

export function useUpdateMaterial() {
  const invalidateInventory = useInvalidateInventory()

  return useMutation({
    mutationFn: ({ material, values }: MaterialUpdate) =>
      inventoryApi.updateMaterial(material.id, toMaterialRequest(values)),
    onSuccess: invalidateInventory,
  })
}

export function useDeleteMaterial() {
  const invalidateInventory = useInvalidateInventory()

  return useMutation({
    mutationFn: (material: Material) => inventoryApi.deleteMaterial(material.id),
    onSuccess: invalidateInventory,
  })
}

export function useWarehouses(filter: WarehouseFilter) {
  return useQuery({
    queryKey: [...WAREHOUSES_KEY, filter],
    queryFn: () => inventoryApi.warehouses(filter),
    ...PAGED_QUERY,
  })
}

export function useCreateWarehouse() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: WarehouseFormValues) => inventoryApi.createWarehouse(toWarehouseRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: WAREHOUSES_KEY }),
  })
}

export interface WarehouseUpdate {
  warehouse: Warehouse
  values: WarehouseFormValues
}

export function useUpdateWarehouse() {
  const invalidateInventory = useInvalidateInventory()

  return useMutation({
    mutationFn: ({ warehouse, values }: WarehouseUpdate) =>
      inventoryApi.updateWarehouse(warehouse.id, toWarehouseRequest(values)),
    onSuccess: invalidateInventory,
  })
}

export function useDeleteWarehouse() {
  const invalidateInventory = useInvalidateInventory()

  return useMutation({
    mutationFn: (warehouse: Warehouse) => inventoryApi.deleteWarehouse(warehouse.id),
    onSuccess: invalidateInventory,
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

export function useAssets(filter: AssetFilter) {
  return useQuery({
    queryKey: [...ASSETS_KEY, filter],
    queryFn: () => assetApi.assets(filter),
    ...PAGED_QUERY,
  })
}

// Menghapus aset melepas tautan alat ke aset itu, jadi daftar alat ikut
// disegarkan.
function useInvalidateAssets() {
  const queryClient = useQueryClient()

  return () => {
    void queryClient.invalidateQueries({ queryKey: ASSETS_KEY })
    void queryClient.invalidateQueries({ queryKey: EQUIPMENT_KEY })
  }
}

export function useCreateAsset() {
  const invalidateAssets = useInvalidateAssets()

  return useMutation({
    mutationFn: (values: AssetFormValues) => assetApi.createAsset(toAssetRequest(values)),
    onSuccess: invalidateAssets,
  })
}

export interface AssetUpdate {
  asset: Asset
  values: AssetFormValues
}

export function useUpdateAsset() {
  const invalidateAssets = useInvalidateAssets()

  return useMutation({
    mutationFn: ({ asset, values }: AssetUpdate) => assetApi.updateAsset(asset.id, toAssetRequest(values)),
    onSuccess: invalidateAssets,
  })
}

export function useDeleteAsset() {
  const invalidateAssets = useInvalidateAssets()

  return useMutation({
    mutationFn: (asset: Asset) => assetApi.deleteAsset(asset.id),
    onSuccess: invalidateAssets,
  })
}

export function useEquipment(filter: EquipmentFilter) {
  return useQuery({
    queryKey: [...EQUIPMENT_KEY, filter],
    queryFn: () => assetApi.equipment(filter),
    ...PAGED_QUERY,
  })
}

function useInvalidateEquipment() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: EQUIPMENT_KEY })
}

export function useCreateEquipment() {
  const invalidateEquipment = useInvalidateEquipment()

  return useMutation({
    mutationFn: (values: EquipmentFormValues) => assetApi.createEquipment(toEquipmentRequest(values)),
    onSuccess: invalidateEquipment,
  })
}

export interface EquipmentUpdate {
  equipment: Equipment
  values: EquipmentFormValues
}

export function useUpdateEquipment() {
  const invalidateEquipment = useInvalidateEquipment()

  return useMutation({
    mutationFn: ({ equipment, values }: EquipmentUpdate) =>
      assetApi.updateEquipment(equipment.id, toEquipmentRequest(values)),
    onSuccess: invalidateEquipment,
  })
}

export function useDeleteEquipment() {
  const invalidateEquipment = useInvalidateEquipment()

  return useMutation({
    mutationFn: (equipment: Equipment) => assetApi.deleteEquipment(equipment.id),
    onSuccess: invalidateEquipment,
  })
}
