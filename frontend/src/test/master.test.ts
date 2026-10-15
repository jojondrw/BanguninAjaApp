import { describe, expect, it } from 'vitest'

import { isLookupOf } from '../controllers/useMaster'
import { accountOptions, postableAccountOptions, SEARCH_LIMIT, unitOfMeasureOptions } from '../models/lookupApi'
import {
  accountParentText,
  accountTypeLocked,
  EMPTY_ACCOUNT_FORM,
  EMPTY_REGION_FORM,
  regionParentHint,
  regionParentText,
  toAccountRequest,
  toRegionRequest,
  toUnitRequest,
  withAccountType,
  withRegionType,
  type Region,
} from '../models/master'
import { mergeRegionOptions, regionOptions } from '../models/masterApi'

function region(id: string, name: string, type: Region['type'], code = id): Region {
  return {
    id,
    code,
    name,
    type,
    parentId: null,
    parentCode: null,
    parentName: null,
    createdAt: '2026-10-01T00:00:00Z',
    updatedAt: '2026-10-01T00:00:00Z',
  }
}

describe('formulir akun', () => {
  it('memangkas spasi dan tidak mengirim induk yang kosong', () => {
    expect(toAccountRequest({ code: ' 1130 ', name: ' Kas Kecil ', type: 'asset', parentId: '' })).toEqual({
      code: '1130',
      name: 'Kas Kecil',
      type: 'asset',
      parentId: undefined,
    })
    expect(toAccountRequest({ ...EMPTY_ACCOUNT_FORM, code: '1130', name: 'Kas Kecil', parentId: 'induk' }).parentId).toBe(
      'induk',
    )
  })

  it('melepas induk saat jenis akun berganti, karena induk harus sejenis', () => {
    const values = { code: '5110', name: 'Semen', type: 'expense' as const, parentId: 'beban-konstruksi' }

    expect(withAccountType(values, 'expense')).toBe(values)
    expect(withAccountType(values, 'asset')).toEqual({ ...values, type: 'asset', parentId: '' })
  })

  it('menyebut induk dengan kode dan nama, atau akun utama bila tanpa induk', () => {
    expect(accountParentText({ parentId: null, parentCode: null, parentName: null })).toBe('Akun utama')
    expect(accountParentText({ parentId: 'p', parentCode: '1100', parentName: 'Kas dan Setara Kas' })).toBe(
      '1100 Kas dan Setara Kas',
    )
    expect(accountParentText({ parentId: 'p', parentCode: null, parentName: null })).toBe('-')
  })

  it('mengunci jenis akun yang punya akun turunan', () => {
    expect(accountTypeLocked(null)).toBe(false)
    expect(accountTypeLocked({ postable: true })).toBe(false)
    expect(accountTypeLocked({ postable: false })).toBe(true)
  })
})

describe('formulir wilayah', () => {
  it('provinsi tidak pernah mengirim induk', () => {
    expect(toRegionRequest({ code: ' 32 ', name: ' Jawa Barat ', type: 'province', parentId: 'sisa' })).toEqual({
      code: '32',
      name: 'Jawa Barat',
      type: 'province',
      parentId: undefined,
    })
    expect(toRegionRequest({ ...EMPTY_REGION_FORM, type: 'regency', parentId: 'jabar' }).parentId).toBe('jabar')
  })

  it('induk tetap selama jenis induk yang diterima sama', () => {
    const city = { code: '3271', name: 'Kota Bogor', type: 'city' as const, parentId: 'jabar' }

    expect(withRegionType(city, 'regency').parentId).toBe('jabar')
    expect(withRegionType(city, 'district')).toEqual({ ...city, type: 'district', parentId: '' })
    expect(withRegionType({ ...city, type: 'district', parentId: 'bogor' }, 'province').parentId).toBe('')
  })

  it('menjelaskan induk yang wajib dipilih', () => {
    expect(regionParentHint('province')).toBe('Provinsi adalah tingkat teratas, tanpa induk')
    expect(regionParentHint('city')).toBe('Wajib, pilih provinsi')
    expect(regionParentHint('district')).toBe('Wajib, pilih kota atau kabupaten')
  })

  it('menampilkan nama induk atau tanda strip', () => {
    expect(regionParentText({ parentId: null, parentName: null })).toBe('-')
    expect(regionParentText({ parentId: 'jabar', parentName: 'Jawa Barat' })).toBe('Jawa Barat')
  })

  it('menggabungkan hasil kota dan kabupaten lalu mengurutkan menurut nama', () => {
    const options = mergeRegionOptions([
      [region('c1', 'Kota Bogor', 'city', '3271')],
      [region('r2', 'Bekasi', 'regency', '3216'), region('r1', 'Bogor', 'regency', '3201')],
    ])

    expect(options.map((option) => option.label)).toEqual(['Bekasi', 'Bogor', 'Kota Bogor'])
    expect(options[1]).toEqual({ value: 'r1', label: 'Bogor', hint: 'Kabupaten 3201' })
  })

  it('membatasi gabungan ke satu halaman pencarian', () => {
    const many = Array.from({ length: SEARCH_LIMIT }, (_, index) => region(`r${index}`, `Wilayah ${index}`, 'regency'))

    expect(mergeRegionOptions([many, many])).toHaveLength(SEARCH_LIMIT)
  })
})

describe('formulir satuan', () => {
  it('memangkas spasi', () => {
    expect(toUnitRequest({ code: ' m3 ', name: ' Meter kubik ' })).toEqual({ code: 'm3', name: 'Meter kubik' })
  })
})

describe('penyegaran pilihan di halaman lain', () => {
  it('mengenali cache SearchSelect untuk akun, wilayah, dan satuan', () => {
    const selected = (cacheKey: string) => ['search-select', cacheKey, 'selected', 'id-1']

    expect(isLookupOf(['search-select', accountOptions().cacheKey, ''], '/master/accounts')).toBe(true)
    expect(isLookupOf(selected(postableAccountOptions.cacheKey), '/master/accounts')).toBe(true)
    expect(isLookupOf(selected(accountOptions('expense').cacheKey), '/master/accounts')).toBe(true)
    expect(isLookupOf(selected(regionOptions(['city', 'regency']).cacheKey), '/master/regions')).toBe(true)
    expect(isLookupOf(selected(unitOfMeasureOptions.cacheKey), '/master/units-of-measure')).toBe(true)
  })

  it('tidak menyentuh cache lain', () => {
    expect(isLookupOf(['search-select', '/projects', ''], '/master/accounts')).toBe(false)
    expect(isLookupOf(['master', 'accounts'], '/master/accounts')).toBe(false)
    expect(isLookupOf(['search-select', '/master/regions', ''], '/master/accounts')).toBe(false)
  })
})
