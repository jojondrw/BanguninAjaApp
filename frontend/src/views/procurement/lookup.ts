import { useQueries } from '@tanstack/react-query'

import { materialOptions, unitOfMeasureOptions, type MaterialOption } from '../../models/lookupApi'
import type { PurchaseOrderFormValues } from '../../models/procurement'
import { useLookupLabel } from '../components/searchSelectLogic'

export const PAGE_SIZE = 10
export const OPTION_LIMIT = 100

// Isian awal formulir pesanan yang dibuka lewat "Buat PO dari permintaan".
export interface OrderPrefill {
  values: PurchaseOrderFormValues
  requestNumber: string
}

// Kode satuan dari id, diambil satu per satu lewat sumber pilihan satuan.
// Kosong selama satuan belum dipilih, masih dimuat, atau tidak ditemukan.
export function useUnitCode(unitOfMeasureId: string | null | undefined): string {
  const label = useLookupLabel(unitOfMeasureOptions, unitOfMeasureId)
  return label === undefined || label === '…' || label === '-' ? '' : label
}

// Material per id beserta satuan stok dan harga terakhirnya. Kuncinya sama
// dengan LookupName untuk materialOptions, jadi nama yang sudah tampil di
// tabel tidak diambil dua kali.
export function useMaterialLookups(materialIds: string[]) {
  const results = useQueries({
    queries: materialIds.map((id) => ({
      queryKey: ['search-select', materialOptions.cacheKey, 'selected', id],
      queryFn: () => materialOptions.fetchSelected(id),
      staleTime: 5 * 60_000,
      retry: false,
    })),
  })

  const byId = new Map<string, MaterialOption>()
  results.forEach((result) => {
    if (result.data) {
      byId.set(result.data.value, result.data)
    }
  })

  return {
    byId,
    isPending: results.some((result) => result.isPending),
  }
}
