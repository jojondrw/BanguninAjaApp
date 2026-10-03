import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  GoodsReceipt,
  GoodsReceiptDetail,
  GoodsReceiptFilter,
  GoodsReceiptRequest,
  PurchaseOrder,
  PurchaseOrderDetail,
  PurchaseOrderFilter,
  PurchaseOrderRequest,
  PurchaseOrderStatusChange,
  PurchaseOrderSummary,
  PurchaseRequest,
  PurchaseRequestDetail,
  PurchaseRequestFilter,
  PurchaseRequestRequest,
  PurchaseRequestStatusChange,
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

  updateVendor: (id: string, body: VendorRequest) =>
    request<Vendor>(`/procurement/vendors/${id}`, { method: 'PUT', body }),

  deleteVendor: (id: string) =>
    request<void>(`/procurement/vendors/${id}`, { method: 'DELETE' }),

  purchaseRequests: (filter: PurchaseRequestFilter = {}) =>
    request<Page<PurchaseRequest>>(`/procurement/purchase-requests${toQueryString({ ...filter })}`),

  purchaseRequest: (id: string) =>
    request<PurchaseRequestDetail>(`/procurement/purchase-requests/${id}`),

  createPurchaseRequest: (body: PurchaseRequestRequest) =>
    request<PurchaseRequestDetail>('/procurement/purchase-requests', { method: 'POST', body }),

  updatePurchaseRequest: (id: string, body: PurchaseRequestRequest) =>
    request<PurchaseRequestDetail>(`/procurement/purchase-requests/${id}`, { method: 'PUT', body }),

  updatePurchaseRequestStatus: (id: string, status: PurchaseRequestStatusChange) =>
    request<PurchaseRequestDetail>(`/procurement/purchase-requests/${id}/status`, {
      method: 'PATCH',
      body: { status },
    }),

  deletePurchaseRequest: (id: string) =>
    request<void>(`/procurement/purchase-requests/${id}`, { method: 'DELETE' }),

  purchaseOrders: (filter: PurchaseOrderFilter = {}) =>
    request<Page<PurchaseOrder>>(`/procurement/purchase-orders${toQueryString({ ...filter })}`),

  purchaseOrderSummary: (projectId?: string) =>
    request<PurchaseOrderSummary>(`/procurement/purchase-orders/summary${toQueryString({ projectId })}`),

  purchaseOrder: (id: string) =>
    request<PurchaseOrderDetail>(`/procurement/purchase-orders/${id}`),

  createPurchaseOrder: (body: PurchaseOrderRequest) =>
    request<PurchaseOrderDetail>('/procurement/purchase-orders', { method: 'POST', body }),

  updatePurchaseOrder: (id: string, body: PurchaseOrderRequest) =>
    request<PurchaseOrderDetail>(`/procurement/purchase-orders/${id}`, { method: 'PUT', body }),

  updatePurchaseOrderStatus: (id: string, status: PurchaseOrderStatusChange) =>
    request<PurchaseOrderDetail>(`/procurement/purchase-orders/${id}/status`, {
      method: 'PATCH',
      body: { status },
    }),

  deletePurchaseOrder: (id: string) =>
    request<void>(`/procurement/purchase-orders/${id}`, { method: 'DELETE' }),

  goodsReceipts: (filter: GoodsReceiptFilter = {}) =>
    request<Page<GoodsReceipt>>(`/procurement/goods-receipts${toQueryString({ ...filter })}`),

  goodsReceipt: (id: string) =>
    request<GoodsReceiptDetail>(`/procurement/goods-receipts/${id}`),

  recordGoodsReceipt: (body: GoodsReceiptRequest) =>
    request<GoodsReceiptDetail>('/procurement/goods-receipts', { method: 'POST', body }),
}
