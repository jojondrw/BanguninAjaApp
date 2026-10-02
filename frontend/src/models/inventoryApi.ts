import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  Asset,
  AssetFilter,
  AssetPage,
  AssetRequest,
  Equipment,
  EquipmentFilter,
  EquipmentRequest,
  LowStockMaterial,
  Material,
  MaterialFilter,
  MaterialRequest,
  Stock,
  StockFilter,
  StockMovement,
  StockMovementFilter,
  StockMovementRequest,
  UnitOfMeasure,
  Warehouse,
  WarehouseFilter,
  WarehouseRequest,
} from './inventory'

interface PageQuery {
  page?: number
  pageSize?: number
}

export const inventoryApi = {
  materials: (filter: MaterialFilter = {}) =>
    request<Page<Material>>(`/inventory/materials${toQueryString({ ...filter })}`),

  lowStockMaterials: (query: PageQuery = {}) =>
    request<Page<LowStockMaterial>>(`/inventory/materials/low-stock${toQueryString({ ...query })}`),

  createMaterial: (body: MaterialRequest) =>
    request<Material>('/inventory/materials', { method: 'POST', body }),

  updateMaterial: (id: string, body: MaterialRequest) =>
    request<Material>(`/inventory/materials/${id}`, { method: 'PUT', body }),

  deleteMaterial: (id: string) => request<void>(`/inventory/materials/${id}`, { method: 'DELETE' }),

  warehouses: (filter: WarehouseFilter = {}) =>
    request<Page<Warehouse>>(`/inventory/warehouses${toQueryString({ ...filter })}`),

  createWarehouse: (body: WarehouseRequest) =>
    request<Warehouse>('/inventory/warehouses', { method: 'POST', body }),

  updateWarehouse: (id: string, body: WarehouseRequest) =>
    request<Warehouse>(`/inventory/warehouses/${id}`, { method: 'PUT', body }),

  deleteWarehouse: (id: string) => request<void>(`/inventory/warehouses/${id}`, { method: 'DELETE' }),

  stocks: (filter: StockFilter = {}) =>
    request<Page<Stock>>(`/inventory/stocks${toQueryString({ ...filter })}`),

  stockMovements: (filter: StockMovementFilter = {}) =>
    request<Page<StockMovement>>(`/inventory/stock-movements${toQueryString({ ...filter })}`),

  recordStockMovement: (body: StockMovementRequest) =>
    request<StockMovement>('/inventory/stock-movements', { method: 'POST', body }),
}

export const unitOfMeasureApi = {
  list: (query: PageQuery = {}) =>
    request<Page<UnitOfMeasure>>(`/master/units-of-measure${toQueryString({ ...query })}`),
}

export const assetApi = {
  assets: (filter: AssetFilter = {}) => request<AssetPage>(`/assets${toQueryString({ ...filter })}`),

  createAsset: (body: AssetRequest) => request<Asset>('/assets', { method: 'POST', body }),

  updateAsset: (id: string, body: AssetRequest) => request<Asset>(`/assets/${id}`, { method: 'PUT', body }),

  deleteAsset: (id: string) => request<void>(`/assets/${id}`, { method: 'DELETE' }),

  equipment: (filter: EquipmentFilter = {}) =>
    request<Page<Equipment>>(`/equipment${toQueryString({ ...filter })}`),

  createEquipment: (body: EquipmentRequest) => request<Equipment>('/equipment', { method: 'POST', body }),

  updateEquipment: (id: string, body: EquipmentRequest) =>
    request<Equipment>(`/equipment/${id}`, { method: 'PUT', body }),

  deleteEquipment: (id: string) => request<void>(`/equipment/${id}`, { method: 'DELETE' }),
}
