import { REGION_TYPE_LABEL, type RegionType } from './project'

export { REGION_TYPE_LABEL, type RegionType }

// Bentuk DTO mengikuti backend/internal/master/dto.go.
export type AccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense'

export interface Account {
  id: string
  code: string
  name: string
  type: AccountType
  parentId: string | null
  parentCode: string | null
  parentName: string | null
  // Akun tanpa akun turunan. Hanya akun ini yang dipakai transaksi kas dan jurnal.
  postable: boolean
  createdAt: string
  updatedAt: string
}

export interface Region {
  id: string
  code: string
  name: string
  type: RegionType
  parentId: string | null
  parentCode: string | null
  parentName: string | null
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

export const ACCOUNT_TYPE_LABEL: Record<AccountType, string> = {
  asset: 'Aset',
  liability: 'Liabilitas',
  equity: 'Ekuitas',
  revenue: 'Pendapatan',
  expense: 'Beban',
}

export const ACCOUNT_TYPES = Object.keys(ACCOUNT_TYPE_LABEL) as AccountType[]
export const REGION_TYPES = Object.keys(REGION_TYPE_LABEL) as RegionType[]

// Batas di bawah mengikuti tag binding di backend/internal/master/dto.go.
export const ACCOUNT_CODE_MAX_LENGTH = 20
export const ACCOUNT_NAME_MAX_LENGTH = 120
export const REGION_CODE_MAX_LENGTH = 20
export const REGION_NAME_MAX_LENGTH = 120
export const UNIT_CODE_MAX_LENGTH = 20
export const UNIT_NAME_MAX_LENGTH = 60
export const ACCOUNT_SEARCH_MAX_LENGTH = 120
export const REGION_SEARCH_MAX_LENGTH = 120
export const UNIT_SEARCH_MAX_LENGTH = 60

// Jurnal kas otomatis mencari akun Kas lewat kode ini
// (backend/internal/finance/cashjournal.go), jadi backend menolak kodenya
// diganti dan akunnya dihapus.
export const CASH_ACCOUNT_CODE = '1110'

// Jenis wilayah induk yang diterima backend untuk tiap tingkat wilayah.
export const REGION_PARENT_TYPES: Record<RegionType, RegionType[]> = {
  province: [],
  city: ['province'],
  regency: ['province'],
  district: ['city', 'regency'],
}

export type PostableFilter = '' | 'true' | 'false'

export interface AccountFilter {
  search?: string
  type?: AccountType
  postable?: Exclude<PostableFilter, ''>
  page?: number
  pageSize?: number
}

export interface RegionFilter {
  search?: string
  type?: RegionType
  parentId?: string
  page?: number
  pageSize?: number
}

export interface UnitOfMeasureFilter {
  search?: string
  page?: number
  pageSize?: number
}

export interface AccountRequest {
  code: string
  name: string
  type: AccountType
  parentId?: string
}

export interface RegionRequest {
  code: string
  name: string
  type: RegionType
  parentId?: string
}

export interface UnitOfMeasureRequest {
  code: string
  name: string
}

export interface AccountFormValues {
  code: string
  name: string
  type: AccountType
  parentId: string
}

export interface RegionFormValues {
  code: string
  name: string
  type: RegionType
  parentId: string
}

export interface UnitOfMeasureFormValues {
  code: string
  name: string
}

export const EMPTY_ACCOUNT_FORM: AccountFormValues = { code: '', name: '', type: 'asset', parentId: '' }
export const EMPTY_REGION_FORM: RegionFormValues = { code: '', name: '', type: 'province', parentId: '' }
export const EMPTY_UNIT_FORM: UnitOfMeasureFormValues = { code: '', name: '' }

export function accountFormValues(account: Account): AccountFormValues {
  return { code: account.code, name: account.name, type: account.type, parentId: account.parentId ?? '' }
}

export function regionFormValues(region: Region): RegionFormValues {
  return { code: region.code, name: region.name, type: region.type, parentId: region.parentId ?? '' }
}

export function unitFormValues(unit: UnitOfMeasure): UnitOfMeasureFormValues {
  return { code: unit.code, name: unit.name }
}

export function toAccountRequest(values: AccountFormValues): AccountRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    type: values.type,
    parentId: values.parentId === '' ? undefined : values.parentId,
  }
}

export function toRegionRequest(values: RegionFormValues): RegionRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    type: values.type,
    parentId: regionNeedsParent(values.type) && values.parentId !== '' ? values.parentId : undefined,
  }
}

export function toUnitRequest(values: UnitOfMeasureFormValues): UnitOfMeasureRequest {
  return { code: values.code.trim(), name: values.name.trim() }
}

// Akun induk harus sejenis, jadi induk yang sudah dipilih dilepas begitu
// jenisnya berganti.
export function withAccountType(values: AccountFormValues, type: AccountType): AccountFormValues {
  return type === values.type ? values : { ...values, type, parentId: '' }
}

export function regionNeedsParent(type: RegionType): boolean {
  return REGION_PARENT_TYPES[type].length > 0
}

// Induk tetap dipakai selama jenis induk yang diterima tidak berubah, misalnya
// kota ke kabupaten (keduanya di bawah provinsi).
export function withRegionType(values: RegionFormValues, type: RegionType): RegionFormValues {
  const sameParents = REGION_PARENT_TYPES[type].join() === REGION_PARENT_TYPES[values.type].join()
  return { ...values, type, parentId: sameParents ? values.parentId : '' }
}

export function regionParentHint(type: RegionType): string {
  const parents = REGION_PARENT_TYPES[type]
  if (parents.length === 0) {
    return 'Provinsi adalah tingkat teratas, tanpa induk'
  }
  return `Wajib, pilih ${parents.map((parent) => REGION_TYPE_LABEL[parent].toLowerCase()).join(' atau ')}`
}

export function accountParentText(account: Pick<Account, 'parentId' | 'parentCode' | 'parentName'>): string {
  if (account.parentId === null) {
    return 'Akun utama'
  }
  return [account.parentCode, account.parentName].filter(Boolean).join(' ') || '-'
}

export function regionParentText(region: Pick<Region, 'parentId' | 'parentName'>): string {
  if (region.parentId === null) {
    return '-'
  }
  return region.parentName ?? '-'
}

export function isCashAccount(account: Pick<Account, 'code'>): boolean {
  return account.code === CASH_ACCOUNT_CODE
}

// Akun yang punya akun turunan tidak bisa berganti jenis (account_type_locked).
export function accountTypeLocked(account: Pick<Account, 'postable'> | null): boolean {
  return account !== null && !account.postable
}
