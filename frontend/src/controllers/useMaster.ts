import { keepPreviousData, useMutation, useQuery, useQueryClient, type QueryKey } from '@tanstack/react-query'

import {
  toAccountRequest,
  toRegionRequest,
  toUnitRequest,
  type Account,
  type AccountFilter,
  type AccountFormValues,
  type Region,
  type RegionFilter,
  type RegionFormValues,
  type UnitOfMeasure,
  type UnitOfMeasureFilter,
  type UnitOfMeasureFormValues,
} from '../models/master'
import { masterApi } from '../models/masterApi'
import { CASH_TRANSACTIONS_KEY, DATA_QUERY } from './useErp'
import { JOURNAL_KEY, LEDGER_KEY } from './useFinance'

// Sama dengan MASTER_KEY di useProjectWorkspace (wilayah proyek), jadi
// perubahan wilayah di sini ikut menyegarkan ruang kerja proyek.
const MASTER_KEY = 'master'
const ACCOUNTS_KEY = [MASTER_KEY, 'accounts'] as const
const REGIONS_KEY = [MASTER_KEY, 'regions'] as const
const UNITS_KEY = [MASTER_KEY, 'units-of-measure'] as const

// Kunci milik controller lain yang belum diekspor: nama wilayah di
// perbandingan lokasi (useLocations), kode satuan yang dibawa baris stok
// (useInventory) dan baris pesanan pembelian (useProcurement), serta judul
// buku besar (useFinance).
const LOCATION_REGIONS_KEY = 'regions'
const INVENTORY_KEY = 'inventory'
const PROCUREMENT_KEY = 'procurement'
const LEDGER_ACCOUNT_KEY = 'ledger-account'

const ACCOUNTS_PATH = '/master/accounts'
const REGIONS_PATH = '/master/regions'
const UNITS_PATH = '/master/units-of-measure'

const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

// Cache SearchSelect dan LookupName berkunci ['search-select', cacheKey, ...]
// dengan cacheKey berawalan alamat endpoint (models/lookupApi.ts), jadi semua
// pilihan akun, wilayah, atau satuan di halaman lain bisa dibuang sekaligus.
export function isLookupOf(queryKey: QueryKey, path: string): boolean {
  const [scope, cacheKey] = queryKey
  return scope === 'search-select' && typeof cacheKey === 'string' && cacheKey.startsWith(path)
}

function useInvalidate() {
  const queryClient = useQueryClient()

  return (keys: QueryKey[], lookupPath: string) =>
    Promise.all([
      ...keys.map((queryKey) => queryClient.invalidateQueries({ queryKey })),
      queryClient.invalidateQueries({ predicate: (query) => isLookupOf(query.queryKey, lookupPath) }),
    ])
}

export function useAccounts(filter: AccountFilter) {
  return useQuery({
    queryKey: [...ACCOUNTS_KEY, filter],
    queryFn: () => masterApi.accounts(filter),
    ...PAGED_QUERY,
  })
}

// Nama dan kode akun ikut tampil di transaksi kas, jurnal, dan buku besar
// (judulnya diambil lewat useLedgerAccount).
const ACCOUNT_USERS: QueryKey[] = [ACCOUNTS_KEY, [CASH_TRANSACTIONS_KEY], [JOURNAL_KEY], [LEDGER_KEY], [LEDGER_ACCOUNT_KEY]]

export function useCreateAccount() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: (values: AccountFormValues) => masterApi.createAccount(toAccountRequest(values)),
    onSuccess: () => invalidate([ACCOUNTS_KEY], ACCOUNTS_PATH),
  })
}

export interface AccountUpdate {
  account: Account
  values: AccountFormValues
}

export function useUpdateAccount() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: ({ account, values }: AccountUpdate) => masterApi.updateAccount(account.id, toAccountRequest(values)),
    onSuccess: () => invalidate(ACCOUNT_USERS, ACCOUNTS_PATH),
  })
}

export function useDeleteAccount() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: (account: Account) => masterApi.deleteAccount(account.id),
    onSuccess: () => invalidate([ACCOUNTS_KEY], ACCOUNTS_PATH),
  })
}

export function useRegions(filter: RegionFilter) {
  return useQuery({
    queryKey: [...REGIONS_KEY, filter],
    queryFn: () => masterApi.regions(filter),
    ...PAGED_QUERY,
  })
}

// Seluruh kunci master ikut dibuang karena ruang kerja proyek menyimpan
// wilayah di bawah [master, region, id].
const REGION_USERS: QueryKey[] = [[MASTER_KEY], [LOCATION_REGIONS_KEY]]

export function useCreateRegion() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: (values: RegionFormValues) => masterApi.createRegion(toRegionRequest(values)),
    onSuccess: () => invalidate([REGIONS_KEY], REGIONS_PATH),
  })
}

export interface RegionUpdate {
  region: Region
  values: RegionFormValues
}

export function useUpdateRegion() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: ({ region, values }: RegionUpdate) => masterApi.updateRegion(region.id, toRegionRequest(values)),
    onSuccess: () => invalidate(REGION_USERS, REGIONS_PATH),
  })
}

export function useDeleteRegion() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: (region: Region) => masterApi.deleteRegion(region.id),
    onSuccess: () => invalidate(REGION_USERS, REGIONS_PATH),
  })
}

export function useUnitsOfMeasure(filter: UnitOfMeasureFilter) {
  return useQuery({
    queryKey: [...UNITS_KEY, filter],
    queryFn: () => masterApi.unitsOfMeasure(filter),
    ...PAGED_QUERY,
  })
}

const UNIT_USERS: QueryKey[] = [UNITS_KEY, [INVENTORY_KEY], [PROCUREMENT_KEY]]

export function useCreateUnitOfMeasure() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: (values: UnitOfMeasureFormValues) => masterApi.createUnitOfMeasure(toUnitRequest(values)),
    onSuccess: () => invalidate([UNITS_KEY], UNITS_PATH),
  })
}

export interface UnitOfMeasureUpdate {
  unit: UnitOfMeasure
  values: UnitOfMeasureFormValues
}

export function useUpdateUnitOfMeasure() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: ({ unit, values }: UnitOfMeasureUpdate) => masterApi.updateUnitOfMeasure(unit.id, toUnitRequest(values)),
    onSuccess: () => invalidate(UNIT_USERS, UNITS_PATH),
  })
}

export function useDeleteUnitOfMeasure() {
  const invalidate = useInvalidate()

  return useMutation({
    mutationFn: (unit: UnitOfMeasure) => masterApi.deleteUnitOfMeasure(unit.id),
    onSuccess: () => invalidate([UNITS_KEY], UNITS_PATH),
  })
}
