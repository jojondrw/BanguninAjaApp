import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { Employee } from './hr'
import type { Asset, Material, Warehouse } from './inventory'
import type { OptionSource, SearchOption } from './lookup'
import { ACCOUNT_TYPE_LABEL, type Account, type AccountType, type UnitOfMeasure } from './master'
import { ORDER_STATUS_LABEL, type PurchaseOrder, type PurchaseOrderStatus, type Vendor } from './procurement'
import type { Project } from './project'
import type { Customer, PropertyUnit, UnitStatus } from './sales'

// Satu halaman hasil per ketikan. Cukup untuk dipilih, sisanya lewat kata kunci.
export const SEARCH_LIMIT = 20

type Params = Record<string, string | number | undefined>

function source<T, O extends SearchOption = SearchOption>(
  path: string,
  toOption: (item: T) => O,
  params: Params = {},
): OptionSource<O> {
  return {
    cacheKey: `${path}${toQueryString(params)}`,
    fetchOptions: (search) =>
      request<Page<T>>(`${path}${toQueryString({ ...params, search: search.trim(), pageSize: SEARCH_LIMIT })}`).then(
        (page) => page.items.map(toOption),
      ),
    fetchSelected: (value) => request<T>(`${path}/${encodeURIComponent(value)}`).then(toOption),
  }
}

const coded = (item: { id: string; code: string; name: string }): SearchOption => ({
  value: item.id,
  label: item.name,
  hint: item.code,
})

export const projectOptions = source<Project>('/projects', coded)

// Material membawa satuan bawaan dan harga terakhir untuk mengisi baris pesanan.
export interface MaterialOption extends SearchOption {
  unitOfMeasureId: string
  lastPrice: number
}

export const materialOptions = source<Material, MaterialOption>('/inventory/materials', (material) => ({
  ...coded(material),
  unitOfMeasureId: material.unitOfMeasureId,
  lastPrice: material.lastPrice,
}))

export function warehouseOptions(projectId?: string): OptionSource {
  return source<Warehouse>('/inventory/warehouses', coded, { projectId })
}

export function vendorOptions(active?: boolean): OptionSource {
  return source<Vendor>('/procurement/vendors', coded, { active: active === undefined ? undefined : String(active) })
}

export const activeVendorOptions = vendorOptions(true)

export const customerOptions = source<Customer>('/sales/customers', (customer) => ({
  value: customer.id,
  label: customer.name,
  hint: customer.identityNumber || customer.contact,
}))

export function employeeOptions(active?: boolean): OptionSource {
  return source<Employee>(
    '/hr/employees',
    (employee) => ({ value: employee.id, label: employee.name, hint: employee.position }),
    { active: active === undefined ? undefined : String(active) },
  )
}

export const activeEmployeeOptions = employeeOptions(true)

export const assetOptions = source<Asset>('/assets', coded)

// Aset induk untuk alat. Aset yang sudah dilepas ditolak backend
// (equipment_asset_disposed), jadi disaring dari hasil pencarian.
export const parentAssetOptions: OptionSource = {
  cacheKey: '/assets?parent',
  fetchOptions: (search) =>
    request<Page<Asset>>(`/assets${toQueryString({ search: search.trim(), pageSize: SEARCH_LIMIT })}`).then((page) =>
      page.items.filter((asset) => asset.status !== 'disposed').map(coded),
    ),
  fetchSelected: assetOptions.fetchSelected,
}

export function accountOptions(type?: AccountType): OptionSource {
  return source<Account>(
    '/master/accounts',
    (account) => ({ value: account.id, label: `${account.code} ${account.name}`, hint: ACCOUNT_TYPE_LABEL[account.type] }),
    { type },
  )
}

export const unitOfMeasureOptions = source<UnitOfMeasure>('/master/units-of-measure', (unit) => ({
  value: unit.id,
  label: unit.code,
  hint: unit.name,
}))

const orderOption = (order: PurchaseOrder): SearchOption => ({
  value: order.id,
  label: order.number,
  hint: [order.vendorName, ORDER_STATUS_LABEL[order.status]].filter(Boolean).join(', '),
})

export function purchaseOrderOptions(status?: PurchaseOrderStatus): OptionSource {
  return source<PurchaseOrder>('/procurement/purchase-orders', orderOption, { status })
}

// Pesanan yang masih menunggu barang: sudah dikirim atau baru diterima sebagian.
// Backend menyaring satu status per panggilan, jadi dua hasil digabung.
const SENT_ORDERS = purchaseOrderOptions('sent')
const PARTIAL_ORDERS = purchaseOrderOptions('partially_received')

export const receivableOrderOptions: OptionSource = {
  cacheKey: '/procurement/purchase-orders?receivable',
  fetchOptions: (search) =>
    Promise.all([SENT_ORDERS.fetchOptions(search), PARTIAL_ORDERS.fetchOptions(search)]).then(([sent, partial]) =>
      [...sent, ...partial].slice(0, SEARCH_LIMIT),
    ),
  fetchSelected: SENT_ORDERS.fetchSelected,
}

// Unit membawa harga untuk mengisi nilai kontrak.
export interface PropertyUnitOption extends SearchOption {
  price: number
}

export function propertyUnitOptions(
  filter: { status?: UnitStatus; projectId?: string } = {},
): OptionSource<PropertyUnitOption> {
  return source<PropertyUnit, PropertyUnitOption>(
    '/sales/units',
    (unit) => ({ value: unit.id, label: unit.code, hint: unit.unitType, price: unit.price }),
    filter,
  )
}

// Akun daun saja (tanpa akun anak), karena kas dan jurnal dicatat di akun
// rincian. Backend yang menyaring lewat postable=true, jadi semua akun daun
// bisa dicari, bukan hanya yang masuk 100 data pertama.
export const postableAccountOptions = source<Account>(
  '/master/accounts',
  (account) => ({ value: account.id, label: `${account.code} ${account.name}`, hint: ACCOUNT_TYPE_LABEL[account.type] }),
  { postable: 'true' },
)
