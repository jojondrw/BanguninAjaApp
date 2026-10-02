import { toApiDate } from '../shared/localDate'
import type { Page } from './common'

export type MovementType = 'in' | 'out' | 'transfer' | 'adjustment'
export type AdjustmentDirection = 'increase' | 'decrease'
export type AssetStatus = 'in_use' | 'maintenance' | 'for_sale' | 'disposed'
export type EquipmentStatus = 'operating' | 'idle' | 'maintenance' | 'broken'

export interface UnitOfMeasure {
  id: string
  code: string
  name: string
  createdAt: string
  updatedAt: string
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

export interface MaterialFilter {
  search?: string
  category?: string
  page?: number
  pageSize?: number
}

export interface MaterialRequest {
  code: string
  name: string
  category: string
  unitOfMeasureId: string
  minimumStock: number
  lastPrice: number
}

export interface LowStockMaterial {
  materialId: string
  code: string
  name: string
  unitOfMeasureId: string
  minimumStock: number
  totalQuantity: number
}

export interface Warehouse {
  id: string
  code: string
  name: string
  projectId: string | null
  createdAt: string
  updatedAt: string
}

export interface WarehouseFilter {
  search?: string
  projectId?: string
  page?: number
  pageSize?: number
}

export interface WarehouseRequest {
  code: string
  name: string
  projectId?: string
}

export interface Stock {
  id: string
  materialId: string
  materialCode: string
  materialName: string
  warehouseId: string
  warehouseCode: string
  warehouseName: string
  quantity: number
  updatedAt: string
}

export interface StockFilter {
  materialId?: string
  warehouseId?: string
  page?: number
  pageSize?: number
}

export interface StockMovement {
  id: string
  date: string
  type: MovementType
  materialId: string
  quantity: number
  sourceWarehouseId: string | null
  targetWarehouseId: string | null
  reference: string
  createdAt: string
}

// dateFrom dan dateTo dibaca backend dengan format "YYYY-MM-DD".
export interface StockMovementFilter {
  materialId?: string
  warehouseId?: string
  type?: MovementType
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

export interface StockMovementRequest {
  date: string
  type: MovementType
  materialId: string
  quantity: number
  sourceWarehouseId?: string
  targetWarehouseId?: string
  reference: string
}

export interface Asset {
  id: string
  code: string
  name: string
  category: string
  acquisitionDate: string | null
  acquisitionValue: number
  accumulatedDepreciation: number
  bookValue: number
  usefulLifeMonths: number
  monthlyDepreciation: number
  status: AssetStatus
  createdAt: string
  updatedAt: string
}

export interface AssetPage extends Page<Asset> {
  totalAcquisitionValue: number
  totalAccumulatedDepreciation: number
  totalBookValue: number
}

export interface Equipment {
  id: string
  code: string
  name: string
  assetId: string | null
  projectId: string | null
  operatingHours: number
  nextServiceDate: string | null
  status: EquipmentStatus
  createdAt: string
  updatedAt: string
}

// Batas panjang mengikuti tag binding di backend.
export const MATERIAL_CODE_MAX_LENGTH = 20
export const MATERIAL_NAME_MAX_LENGTH = 160
export const MATERIAL_CATEGORY_MAX_LENGTH = 60
export const WAREHOUSE_CODE_MAX_LENGTH = 20
export const WAREHOUSE_NAME_MAX_LENGTH = 120
export const MOVEMENT_REFERENCE_MAX_LENGTH = 60

export const MATERIAL_CATEGORY_SUGGESTIONS = ['Semen', 'Besi', 'Pasir', 'Batu', 'Kayu', 'Cat', 'Keramik', 'Pipa', 'Listrik']

export const MOVEMENT_TYPES: MovementType[] = ['in', 'out', 'transfer', 'adjustment']

export const MOVEMENT_TYPE_LABEL: Record<MovementType, string> = {
  in: 'Masuk',
  out: 'Keluar',
  transfer: 'Pindah gudang',
  adjustment: 'Penyesuaian',
}

export const MOVEMENT_TYPE_TONE: Record<MovementType, string> = {
  in: 'bg-green-100 text-green-800',
  out: 'bg-amber-100 text-amber-800',
  transfer: 'bg-blue-100 text-blue-800',
  adjustment: 'bg-slate-100 text-slate-700',
}

export const ASSET_STATUS_LABEL: Record<AssetStatus, string> = {
  in_use: 'Dipakai',
  maintenance: 'Perawatan',
  for_sale: 'Dijual',
  disposed: 'Dilepas',
}

export const EQUIPMENT_STATUS_LABEL: Record<EquipmentStatus, string> = {
  operating: 'Beroperasi',
  idle: 'Siaga',
  maintenance: 'Perawatan',
  broken: 'Rusak',
}

export const EQUIPMENT_STATUS_TONE: Record<EquipmentStatus, string> = {
  operating: 'bg-green-100 text-green-800',
  idle: 'bg-slate-100 text-slate-700',
  maintenance: 'bg-amber-100 text-amber-800',
  broken: 'bg-red-100 text-red-800',
}

export interface MaterialFormValues {
  code: string
  name: string
  category: string
  unitOfMeasureId: string
  minimumStock: string
  lastPrice: string
}

export const EMPTY_MATERIAL_FORM: MaterialFormValues = {
  code: '',
  name: '',
  category: '',
  unitOfMeasureId: '',
  minimumStock: '',
  lastPrice: '',
}

export interface WarehouseFormValues {
  code: string
  name: string
  projectId: string
}

export const EMPTY_WAREHOUSE_FORM: WarehouseFormValues = {
  code: '',
  name: '',
  projectId: '',
}

export interface MovementFormValues {
  type: MovementType
  direction: AdjustmentDirection
  materialId: string
  quantity: string
  sourceWarehouseId: string
  targetWarehouseId: string
  date: string
  reference: string
}

export function emptyMovementForm(date: string): MovementFormValues {
  return {
    type: 'in',
    direction: 'increase',
    materialId: '',
    quantity: '',
    sourceWarehouseId: '',
    targetWarehouseId: '',
    date,
    reference: '',
  }
}

// Gudang mana yang wajib diisi bergantung pada jenis mutasi. Aturan yang sama
// dijaga backend lewat directionMatchesType.
export function movementNeedsSource(values: MovementFormValues): boolean {
  if (values.type === 'adjustment') {
    return values.direction === 'decrease'
  }
  return values.type === 'out' || values.type === 'transfer'
}

export function movementNeedsTarget(values: MovementFormValues): boolean {
  if (values.type === 'adjustment') {
    return values.direction === 'increase'
  }
  return values.type === 'in' || values.type === 'transfer'
}

function amount(value: string): number {
  return value === '' ? 0 : Number(value)
}

export function toMaterialRequest(values: MaterialFormValues): MaterialRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    category: values.category.trim(),
    unitOfMeasureId: values.unitOfMeasureId,
    minimumStock: amount(values.minimumStock),
    lastPrice: amount(values.lastPrice),
  }
}

export function toWarehouseRequest(values: WarehouseFormValues): WarehouseRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    projectId: values.projectId === '' ? undefined : values.projectId,
  }
}

export function toStockMovementRequest(values: MovementFormValues): StockMovementRequest {
  return {
    date: toApiDate(values.date),
    type: values.type,
    materialId: values.materialId,
    quantity: amount(values.quantity),
    sourceWarehouseId: movementNeedsSource(values) ? values.sourceWarehouseId : undefined,
    targetWarehouseId: movementNeedsTarget(values) ? values.targetWarehouseId : undefined,
    reference: values.reference.trim(),
  }
}
