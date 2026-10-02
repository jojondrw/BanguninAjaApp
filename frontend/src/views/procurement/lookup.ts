import type { PurchaseOrderFormValues } from '../../models/procurement'

export const PAGE_SIZE = 10
export const OPTION_LIMIT = 100

// Isian awal formulir pesanan yang dibuka lewat "Buat PO dari permintaan".
export interface OrderPrefill {
  values: PurchaseOrderFormValues
  requestNumber: string
}

// Respons backend hanya membawa id, jadi nama dicari dari daftar pilihan yang
// sudah dimuat. Strip kalau datanya di luar 100 pilihan pertama.
export function nameOf(items: { id: string; name: string }[], id: string): string {
  return items.find((item) => item.id === id)?.name ?? '-'
}

export function codeOf(items: { id: string; code: string }[], id: string): string {
  return items.find((item) => item.id === id)?.code ?? '-'
}

export function numberOf(items: { id: string; number: string }[], id: string): string {
  return items.find((item) => item.id === id)?.number ?? '-'
}
