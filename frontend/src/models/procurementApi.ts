import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  Material,
  PurchaseOrder,
  PurchaseOrderDetail,
  PurchaseOrderFilter,
  PurchaseOrderRequest,
  PurchaseRequest,
  PurchaseRequestFilter,
  UnitOfMeasure,
  Vendor,
  VendorFilter,
  VendorRequest,
} from './procurement'

export const procurementApi = {
  vendors: ({ active, ...filter }: VendorFilter = {}) =>
    request<Page<Vendor>>(
      `/procurement/vendors${toQueryString({ ...filter, active: active === undefined ? undefined : String(active) })}`,
    ),

  createVendor: (body: VendorRequest) =>
    request<Vendor>('/procurement/vendors', { method: 'POST', body }),

  purchaseRequests: (filter: PurchaseRequestFilter = {}) =>
    request<Page<PurchaseRequest>>(`/procurement/purchase-requests${toQueryString({ ...filter })}`),

  purchaseOrders: (filter: PurchaseOrderFilter = {}) =>
    request<Page<PurchaseOrder>>(`/procurement/purchase-orders${toQueryString({ ...filter })}`),

  purchaseOrder: (id: string) =>
    request<PurchaseOrderDetail>(`/procurement/purchase-orders/${id}`),

  createPurchaseOrder: (body: PurchaseOrderRequest) =>
    request<PurchaseOrderDetail>('/procurement/purchase-orders', { method: 'POST', body }),

  updatePurchaseOrderStatus: (id: string, status: 'sent' | 'cancelled') =>
    request<PurchaseOrderDetail>(`/procurement/purchase-orders/${id}/status`, {
      method: 'PATCH',
      body: { status },
    }),

  materials: (pageSize: number) =>
    request<Page<Material>>(`/inventory/materials${toQueryString({ pageSize })}`),

  unitsOfMeasure: (pageSize: number) =>
    request<Page<UnitOfMeasure>>(`/master/units-of-measure${toQueryString({ pageSize })}`),
}
