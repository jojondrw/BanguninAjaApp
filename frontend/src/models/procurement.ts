import { toApiDate, todayInput } from './sales'

export type VendorRating = 'new' | 'good' | 'fair' | 'poor'
export type PurchaseRequestStatus = 'draft' | 'submitted' | 'approved' | 'rejected' | 'completed'
export type PurchaseOrderStatus = 'draft' | 'sent' | 'partially_received' | 'completed' | 'cancelled'

export interface Vendor {
  id: string
  code: string
  name: string
  category: string
  contact: string
  taxNumber: string
  paymentTermDays: number
  rating: VendorRating
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface VendorFilter {
  search?: string
  category?: string
  rating?: VendorRating
  active?: boolean
  page?: number
  pageSize?: number
}

export interface VendorRequest {
  code: string
  name: string
  category: string
  contact: string
  taxNumber: string
  paymentTermDays: number
  rating: VendorRating
  active: boolean
}

export interface PurchaseRequest {
  id: string
  number: string
  projectId: string
  requesterId: string
  date: string
  status: PurchaseRequestStatus
  note: string
  itemCount: number
  createdAt: string
  updatedAt: string
}

export interface PurchaseRequestFilter {
  search?: string
  projectId?: string
  status?: PurchaseRequestStatus
  page?: number
  pageSize?: number
}

export interface PurchaseOrder {
  id: string
  number: string
  vendorId: string
  projectId: string
  purchaseRequestId: string | null
  date: string
  dueDate: string | null
  value: number
  status: PurchaseOrderStatus
  createdAt: string
  updatedAt: string
}

export interface PurchaseOrderItem {
  id: string
  materialId: string
  quantity: number
  unitOfMeasureId: string
  unitPrice: number
  total: number
  receivedQuantity: number
  remainingQuantity: number
}

export interface PurchaseOrderDetail extends PurchaseOrder {
  items: PurchaseOrderItem[]
}

export interface PurchaseOrderFilter {
  search?: string
  vendorId?: string
  projectId?: string
  status?: PurchaseOrderStatus
  page?: number
  pageSize?: number
}

export interface PurchaseOrderItemRequest {
  materialId: string
  quantity: number
  unitOfMeasureId: string
  unitPrice: number
}

export interface PurchaseOrderRequest {
  number: string
  vendorId: string
  projectId: string
  date: string
  dueDate?: string
  items: PurchaseOrderItemRequest[]
}

export interface Material {
  id: string
  code: string
  name: string
  category: string
  unitOfMeasureId: string
  minimumStock: number
  lastPrice: number
  createdAt: string
  updatedAt: string
}

export interface UnitOfMeasure {
  id: string
  code: string
  name: string
  createdAt: string
  updatedAt: string
}

// Batas panjang mengikuti tag binding di backend/internal/procurement/dto.go.
export const VENDOR_CODE_MAX_LENGTH = 20
export const VENDOR_NAME_MAX_LENGTH = 160
export const VENDOR_CATEGORY_MAX_LENGTH = 60
export const VENDOR_CONTACT_MAX_LENGTH = 60
export const VENDOR_TAX_NUMBER_MAX_LENGTH = 25
export const ORDER_NUMBER_MAX_LENGTH = 40

export const VENDOR_CATEGORY_SUGGESTIONS = ['Material', 'Alat berat', 'Subkontraktor', 'Jasa', 'Logistik']

export const OPEN_ORDER_STATUSES: PurchaseOrderStatus[] = ['draft', 'sent', 'partially_received']

function toAmount(value: string): number {
  return value === '' ? 0 : Number(value)
}

export interface VendorFormValues {
  code: string
  name: string
  category: string
  contact: string
  taxNumber: string
  paymentTermDays: string
  rating: VendorRating
}

export const EMPTY_VENDOR_FORM: VendorFormValues = {
  code: '',
  name: '',
  category: '',
  contact: '',
  taxNumber: '',
  paymentTermDays: '30',
  rating: 'new',
}

// Vendor baru selalu aktif. Backend saat ini mengabaikan active=false saat
// membuat vendor (kolomnya punya default true di GORM), jadi pilihan nonaktif
// baru masuk akal lewat ubah vendor.
export function toVendorRequest(values: VendorFormValues): VendorRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    category: values.category.trim(),
    contact: values.contact.trim(),
    taxNumber: values.taxNumber.trim(),
    paymentTermDays: toAmount(values.paymentTermDays),
    rating: values.rating,
    active: true,
  }
}

export interface OrderLineValues {
  key: string
  materialId: string
  quantity: string
  unitOfMeasureId: string
  unitPrice: string
}

export interface PurchaseOrderFormValues {
  number: string
  vendorId: string
  projectId: string
  date: string
  dueDate: string
  items: OrderLineValues[]
}

// Kunci baris hanya untuk React. Penghitung biasa dipakai karena
// crypto.randomUUID tidak tersedia di origin http selain localhost.
let lineSequence = 0

export function emptyOrderLine(): OrderLineValues {
  lineSequence += 1
  return { key: `line-${lineSequence}`, materialId: '', quantity: '', unitOfMeasureId: '', unitPrice: '' }
}

export function emptyPurchaseOrderForm(): PurchaseOrderFormValues {
  return {
    number: '',
    vendorId: '',
    projectId: '',
    date: todayInput(),
    dueDate: '',
    items: [emptyOrderLine()],
  }
}

// Sama dengan lineTotal di backend: jumlah dikali harga, dibulatkan ke rupiah.
export function orderLineTotal(line: OrderLineValues): number {
  const total = Number(line.quantity) * Number(line.unitPrice)
  return Number.isFinite(total) ? Math.round(total) : 0
}

export function toPurchaseOrderRequest(values: PurchaseOrderFormValues): PurchaseOrderRequest {
  return {
    number: values.number.trim(),
    vendorId: values.vendorId,
    projectId: values.projectId,
    date: toApiDate(values.date),
    dueDate: values.dueDate === '' ? undefined : toApiDate(values.dueDate),
    items: values.items.map((line) => ({
      materialId: line.materialId,
      quantity: toAmount(line.quantity),
      unitOfMeasureId: line.unitOfMeasureId,
      unitPrice: toAmount(line.unitPrice),
    })),
  }
}

export const VENDOR_RATING_LABEL: Record<VendorRating, string> = {
  new: 'Baru',
  good: 'Baik',
  fair: 'Cukup',
  poor: 'Buruk',
}

export const VENDOR_RATING_TONE: Record<VendorRating, string> = {
  new: 'bg-slate-100 text-slate-700',
  good: 'bg-green-100 text-green-800',
  fair: 'bg-amber-100 text-amber-800',
  poor: 'bg-red-100 text-red-800',
}

export const REQUEST_STATUS_LABEL: Record<PurchaseRequestStatus, string> = {
  draft: 'Draf',
  submitted: 'Diajukan',
  approved: 'Disetujui',
  rejected: 'Ditolak',
  completed: 'Selesai',
}

export const REQUEST_STATUS_TONE: Record<PurchaseRequestStatus, string> = {
  draft: 'bg-slate-100 text-slate-700',
  submitted: 'bg-amber-100 text-amber-800',
  approved: 'bg-blue-100 text-blue-800',
  rejected: 'bg-red-100 text-red-800',
  completed: 'bg-green-100 text-green-800',
}

export const ORDER_STATUS_LABEL: Record<PurchaseOrderStatus, string> = {
  draft: 'Draf',
  sent: 'Dikirim ke vendor',
  partially_received: 'Diterima sebagian',
  completed: 'Selesai',
  cancelled: 'Batal',
}

export const ORDER_STATUS_TONE: Record<PurchaseOrderStatus, string> = {
  draft: 'bg-slate-100 text-slate-700',
  sent: 'bg-blue-100 text-blue-800',
  partially_received: 'bg-amber-100 text-amber-800',
  completed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
}
