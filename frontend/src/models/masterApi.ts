import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { OptionSource, SearchOption } from './lookup'
import { SEARCH_LIMIT } from './lookupApi'
import {
  REGION_TYPE_LABEL,
  type Account,
  type AccountFilter,
  type AccountRequest,
  type Region,
  type RegionFilter,
  type RegionRequest,
  type RegionType,
  type UnitOfMeasure,
  type UnitOfMeasureFilter,
  type UnitOfMeasureRequest,
} from './master'

export const masterApi = {
  accounts: (filter: AccountFilter = {}) =>
    request<Page<Account>>(`/master/accounts${toQueryString({ ...filter })}`),

  account: (id: string) => request<Account>(`/master/accounts/${encodeURIComponent(id)}`),

  createAccount: (body: AccountRequest) => request<Account>('/master/accounts', { method: 'POST', body }),

  updateAccount: (id: string, body: AccountRequest) =>
    request<Account>(`/master/accounts/${encodeURIComponent(id)}`, { method: 'PUT', body }),

  deleteAccount: (id: string) => request<void>(`/master/accounts/${encodeURIComponent(id)}`, { method: 'DELETE' }),

  regions: (filter: RegionFilter = {}) => request<Page<Region>>(`/master/regions${toQueryString({ ...filter })}`),

  region: (id: string) => request<Region>(`/master/regions/${encodeURIComponent(id)}`),

  createRegion: (body: RegionRequest) => request<Region>('/master/regions', { method: 'POST', body }),

  updateRegion: (id: string, body: RegionRequest) =>
    request<Region>(`/master/regions/${encodeURIComponent(id)}`, { method: 'PUT', body }),

  deleteRegion: (id: string) => request<void>(`/master/regions/${encodeURIComponent(id)}`, { method: 'DELETE' }),

  unitsOfMeasure: (filter: UnitOfMeasureFilter = {}) =>
    request<Page<UnitOfMeasure>>(`/master/units-of-measure${toQueryString({ ...filter })}`),

  createUnitOfMeasure: (body: UnitOfMeasureRequest) =>
    request<UnitOfMeasure>('/master/units-of-measure', { method: 'POST', body }),

  updateUnitOfMeasure: (id: string, body: UnitOfMeasureRequest) =>
    request<UnitOfMeasure>(`/master/units-of-measure/${encodeURIComponent(id)}`, { method: 'PUT', body }),

  deleteUnitOfMeasure: (id: string) =>
    request<void>(`/master/units-of-measure/${encodeURIComponent(id)}`, { method: 'DELETE' }),
}

export function regionOption(region: Region): SearchOption {
  return { value: region.id, label: region.name, hint: `${REGION_TYPE_LABEL[region.type] ?? region.type} ${region.code}` }
}

// Wilayah untuk SearchSelect, dibatasi ke beberapa tingkat. Backend hanya
// menyaring satu tingkat per panggilan, jadi hasil tiap tingkat digabung dan
// diurutkan menurut nama.
export function regionOptions(types: RegionType[] = []): OptionSource {
  const perType: (RegionType | undefined)[] = types.length === 0 ? [undefined] : types

  return {
    cacheKey: `/master/regions${toQueryString({ type: types.join(',') })}`,
    fetchOptions: (search) =>
      Promise.all(
        perType.map((type) => masterApi.regions({ search: search.trim(), type, pageSize: SEARCH_LIMIT })),
      ).then((pages) => mergeRegionOptions(pages.map((page) => page.items))),
    fetchSelected: (value) => masterApi.region(value).then(regionOption),
  }
}

export function mergeRegionOptions(groups: Region[][]): SearchOption[] {
  return groups
    .flat()
    .sort((first, second) => first.name.localeCompare(second.name, 'id') || first.code.localeCompare(second.code))
    .slice(0, SEARCH_LIMIT)
    .map(regionOption)
}
