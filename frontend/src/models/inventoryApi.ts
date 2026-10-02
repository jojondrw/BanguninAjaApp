import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  AssetPage,
  Equipment,
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

  warehouses: (filter: WarehouseFilter = {}) =>
    request<Page<Warehouse>>(`/inventory/warehouses${toQueryString({ ...filter })}`),

  createWarehouse: (body: WarehouseRequest) =>
    request<Warehouse>('/inventory/warehouses', { method: 'POST', body }),

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
  assets: (query: PageQuery = {}) => request<AssetPage>(`/assets${toQueryString({ ...query })}`),

  equipment: (query: PageQuery = {}) =>
    request<Page<Equipment>>(`/equipment${toQueryString({ ...query })}`),
}
