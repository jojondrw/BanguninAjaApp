import type { StockMovement, StockMovementRequest } from './inventory'
import { toApiDate, todayInput } from './sales'

export type VendorRating = 'new' | 'good' | 'fair' | 'poor'
export type PurchaseRequestStatus = 'draft' | 'submitted' | 'approved' | 'rejected' | 'completed'
export type PurchaseOrderStatus = 'draft' | 'sent' | 'partially_received' | 'completed' | 'cancelled'
export type GoodsReceiptCondition = 'good' | 'partially_damaged' | 'broken'

// Status yang boleh dikirim lewat PATCH .../status. partially_received dan
// completed pada pesanan hanya diatur oleh penerimaan barang.
export type PurchaseRequestStatusChange = Exclude<PurchaseRequestStatus, 'draft'>
export type PurchaseOrderStatusChange = 'sent' | 'cancelled'

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
  // Hanya ada di respons daftar.
  projectName?: string
  createdAt: string
  updatedAt: string
}

export interface PurchaseRequestItem {
  id: string
  materialId: string
  quantity: number
  unitOfMeasureId: string
}

export interface PurchaseRequestDetail extends PurchaseRequest {
  items: PurchaseRequestItem[]
}

export interface PurchaseRequestFilter {
  search?: string
  projectId?: string
  status?: PurchaseRequestStatus
  page?: number
  pageSize?: number
}

export interface PurchaseRequestItemRequest {
  materialId: string
  quantity: number
  unitOfMeasureId: string
}

export interface PurchaseRequestRequest {
  number: string
  projectId: string
  date: string
  note: string
  items: PurchaseRequestItemRequest[]
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
  // Hanya ada di respons daftar.
  vendorName?: string
  projectName?: string
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
  purchaseRequestId?: string
  date: string
  dueDate?: string
  items: PurchaseOrderItemRequest[]
}

export interface GoodsReceipt {
  id: string
  number: string
  purchaseOrderId: string
  warehouseId: string
  date: string
  condition: GoodsReceiptCondition
  note: string
  itemCount: number
  // Hanya ada di respons daftar.
  warehouseName?: string
  purchaseOrderNumber?: string
  createdAt: string
}

export interface GoodsReceiptItem {
  id: string
  purchaseOrderItemId: string
  acceptedQuantity: number
  rejectedQuantity: number
}

export interface GoodsReceiptDetail extends GoodsReceipt {
  items: GoodsReceiptItem[]
}

// dateFrom dan dateTo dibaca backend dengan format "YYYY-MM-DD".
export interface GoodsReceiptFilter {
  search?: string
  purchaseOrderId?: string
  warehouseId?: string
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

export interface GoodsReceiptItemRequest {
  purchaseOrderItemId: string
  acceptedQuantity: number
  rejectedQuantity: number
}

export interface GoodsReceiptRequest {
  number: string
  purchaseOrderId: string
  warehouseId: string
  date: string
  condition: GoodsReceiptCondition
  note: string
  items: GoodsReceiptItemRequest[]
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
export const REQUEST_NUMBER_MAX_LENGTH = 40
export const REQUEST_NOTE_MAX_LENGTH = 2000
export const RECEIPT_NUMBER_MAX_LENGTH = 40
export const RECEIPT_NOTE_MAX_LENGTH = 2000

export const VENDOR_CATEGORY_SUGGESTIONS = ['Material', 'Alat berat', 'Subkontraktor', 'Jasa', 'Logistik']

export const OPEN_ORDER_STATUSES: PurchaseOrderStatus[] = ['draft', 'sent', 'partially_received']

// Sama dengan pengecekan receiveGoods di backend.
export const RECEIVABLE_ORDER_STATUSES: PurchaseOrderStatus[] = ['sent', 'partially_received']

export const RECEIPT_CONDITIONS: GoodsReceiptCondition[] = ['good', 'partially_damaged', 'broken']

function toAmount(value: string): number {
  return value === '' ? 0 : Number(value)
}

// Backend mengirim tanggal sebagai "2026-10-02T00:00:00Z", isian tanggal butuh
// "2026-10-02".
export function toInputDate(iso: string | null): string {
  return iso ? iso.slice(0, 10) : ''
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

export function vendorFormFrom(vendor: Vendor): VendorFormValues {
  return {
    code: vendor.code,
    name: vendor.name,
    category: vendor.category,
    contact: vendor.contact,
    taxNumber: vendor.taxNumber,
    paymentTermDays: String(vendor.paymentTermDays),
    rating: vendor.rating,
  }
}

// Vendor baru selalu aktif. Saat mengubah vendor, status aktifnya dibawa dari
// data yang ada karena PUT mengganti semua kolom, termasuk active.
export function toVendorRequest(values: VendorFormValues, active = true): VendorRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    category: values.category.trim(),
    contact: values.contact.trim(),
    taxNumber: values.taxNumber.trim(),
    paymentTermDays: toAmount(values.paymentTermDays),
    rating: values.rating,
    active,
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
  purchaseRequestId: string
  date: string
  dueDate: string
  items: OrderLineValues[]
}

// Baris permintaan memakai bentuk yang sama dengan baris pesanan supaya bisa
// diteruskan ke formulir pesanan. Harga satuan diabaikan saat dikirim.
export interface PurchaseRequestFormValues {
  number: string
  projectId: string
  date: string
  note: string
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
    purchaseRequestId: '',
    date: todayInput(),
    dueDate: '',
    items: [emptyOrderLine()],
  }
}

function lineFrom(item: PurchaseRequestItemRequest, unitPrice: string): OrderLineValues {
  return {
    ...emptyOrderLine(),
    materialId: item.materialId,
    quantity: String(item.quantity),
    unitOfMeasureId: item.unitOfMeasureId,
    unitPrice,
  }
}

export function purchaseOrderFormFrom(order: PurchaseOrderDetail): PurchaseOrderFormValues {
  return {
    number: order.number,
    vendorId: order.vendorId,
    projectId: order.projectId,
    purchaseRequestId: order.purchaseRequestId ?? '',
    date: toInputDate(order.date),
    dueDate: toInputDate(order.dueDate),
    items: order.items.map((item) => lineFrom(item, String(item.unitPrice))),
  }
}

// Pesanan dari permintaan: proyek, material, jumlah, dan satuan diambil dari
// permintaan, harga satuan dari harga terakhir material. Vendor dan nomor
// dipilih sendiri.
export function purchaseOrderFormFromRequest(request: PurchaseRequestDetail, materials: Material[]): PurchaseOrderFormValues {
  return {
    ...emptyPurchaseOrderForm(),
    projectId: request.projectId,
    purchaseRequestId: request.id,
    items: request.items.map((item) => {
      const material = materials.find((candidate) => candidate.id === item.materialId)
      return lineFrom(item, material ? String(material.lastPrice) : '')
    }),
  }
}

export function emptyPurchaseRequestForm(): PurchaseRequestFormValues {
  return {
    number: '',
    projectId: '',
    date: todayInput(),
    note: '',
    items: [emptyOrderLine()],
  }
}

export function purchaseRequestFormFrom(request: PurchaseRequestDetail): PurchaseRequestFormValues {
  return {
    number: request.number,
    projectId: request.projectId,
    date: toInputDate(request.date),
    note: request.note,
    items: request.items.map((item) => lineFrom(item, '')),
  }
}

export function toPurchaseRequestRequest(values: PurchaseRequestFormValues): PurchaseRequestRequest {
  return {
    number: values.number.trim(),
    projectId: values.projectId,
    date: toApiDate(values.date),
    note: values.note.trim(),
    items: values.items.map((line) => ({
      materialId: line.materialId,
      quantity: toAmount(line.quantity),
      unitOfMeasureId: line.unitOfMeasureId,
    })),
  }
}

// Sama dengan lineTotal di backend: jumlah dikali harga, dibulatkan ke rupiah.
export function orderLineTotal(line: OrderLineValues): number {
  const total = Number(line.quantity) * Number(line.unitPrice)
  return Number.isFinite(total) ? Math.round(total) : 0
}

// purchaseRequestId ikut dikirim saat mengubah pesanan, karena PUT mengganti
// semua kolom dan tautan ke permintaan akan hilang kalau dikosongkan.
export function toPurchaseOrderRequest(values: PurchaseOrderFormValues): PurchaseOrderRequest {
  return {
    number: values.number.trim(),
    vendorId: values.vendorId,
    projectId: values.projectId,
    purchaseRequestId: values.purchaseRequestId === '' ? undefined : values.purchaseRequestId,
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

// Cermin requestTransitions di backend/internal/procurement/service.go.
export const REQUEST_TRANSITIONS: Record<PurchaseRequestStatus, PurchaseRequestStatusChange[]> = {
  draft: ['submitted'],
  submitted: ['approved', 'rejected'],
  approved: ['completed'],
  rejected: [],
  completed: [],
}

export const RECEIPT_CONDITION_LABEL: Record<GoodsReceiptCondition, string> = {
  good: 'Baik',
  partially_damaged: 'Rusak sebagian',
  broken: 'Rusak',
}

export const RECEIPT_CONDITION_TONE: Record<GoodsReceiptCondition, string> = {
  good: 'bg-green-100 text-green-800',
  partially_damaged: 'bg-amber-100 text-amber-800',
  broken: 'bg-red-100 text-red-800',
}

export interface GoodsReceiptFormValues {
  number: string
  warehouseId: string
  date: string
  condition: GoodsReceiptCondition
  note: string
}

export function emptyGoodsReceiptForm(): GoodsReceiptFormValues {
  return { number: '', warehouseId: '', date: todayInput(), condition: 'good', note: '' }
}

export interface ReceiptLineValues {
  accepted: string
  rejected: string
}

// Jumlah diterima diisi sisa pesanan, jadi kiriman yang datang lengkap cukup
// disimpan tanpa mengetik ulang.
export function defaultReceiptLine(item: PurchaseOrderItem): ReceiptLineValues {
  return { accepted: item.remainingQuantity > 0 ? String(item.remainingQuantity) : '', rejected: '' }
}

export function receiptLineQuantities(line: ReceiptLineValues): { accepted: number; rejected: number } {
  return { accepted: toAmount(line.accepted), rejected: toAmount(line.rejected) }
}

// Baris tanpa jumlah diterima maupun ditolak tidak dikirim, karena backend
// menolak baris kosong (goods_receipt_quantity_invalid).
export function toGoodsReceiptRequest(
  purchaseOrderId: string,
  values: GoodsReceiptFormValues,
  lines: { purchaseOrderItemId: string; values: ReceiptLineValues }[],
): GoodsReceiptRequest {
  return {
    number: values.number.trim(),
    purchaseOrderId,
    warehouseId: values.warehouseId,
    date: toApiDate(values.date),
    condition: values.condition,
    note: values.note.trim(),
    items: lines
      .map((line) => {
        const quantities = receiptLineQuantities(line.values)
        return {
          purchaseOrderItemId: line.purchaseOrderItemId,
          acceptedQuantity: quantities.accepted,
          rejectedQuantity: quantities.rejected,
        }
      })
      .filter((item) => item.acceptedQuantity + item.rejectedQuantity > 0),
  }
}

export type ReceiptStockState = 'recorded' | 'pending' | 'unit_mismatch' | 'none'

export interface ReceiptStockLine {
  receiptItemId: string
  materialId: string
  unitOfMeasureId: string
  acceptedQuantity: number
  rejectedQuantity: number
  state: ReceiptStockState
}

// Penerimaan belum menambah stok sendiri (lihat slice-bisnis.md). Stok masuk
// dicatat sebagai mutasi "in" ke gudang penerimaan dengan referensi nomor
// penerimaan, jadi baris yang sudah punya mutasi itu dianggap sudah masuk stok.
// Stok dihitung dalam satuan material, jadi baris yang satuannya berbeda harus
// dikonversi manual di Persediaan.
export function receiptStockLines(
  receipt: GoodsReceiptDetail,
  orderItems: PurchaseOrderItem[],
  materials: Material[],
  movements: StockMovement[],
): ReceiptStockLine[] {
  const recorded = movements.filter((movement) => movement.reference === receipt.number)

  return receipt.items.map((item) => {
    const orderItem = orderItems.find((candidate) => candidate.id === item.purchaseOrderItemId)
    const materialId = orderItem?.materialId ?? ''
    const unitOfMeasureId = orderItem?.unitOfMeasureId ?? ''
    const material = materials.find((candidate) => candidate.id === materialId)

    let state: ReceiptStockState = 'pending'
    if (item.acceptedQuantity <= 0) {
      state = 'none'
    } else if (recorded.some((movement) => movement.materialId === materialId)) {
      state = 'recorded'
    } else if (material && material.unitOfMeasureId !== unitOfMeasureId) {
      state = 'unit_mismatch'
    }

    return {
      receiptItemId: item.id,
      materialId,
      unitOfMeasureId,
      acceptedQuantity: item.acceptedQuantity,
      rejectedQuantity: item.rejectedQuantity,
      state,
    }
  })
}

export function receiptStockRequests(receipt: GoodsReceipt, lines: ReceiptStockLine[]): StockMovementRequest[] {
  return lines
    .filter((line) => line.state === 'pending' && line.materialId !== '')
    .map((line) => ({
      date: receipt.date,
      type: 'in',
      materialId: line.materialId,
      quantity: line.acceptedQuantity,
      targetWarehouseId: receipt.warehouseId,
      reference: receipt.number,
    }))
}
