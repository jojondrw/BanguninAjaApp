import { number } from '../../shared/format'

export const PAGE_SIZE = 20

// Tabel stok, mutasi, dan stok menipis membawa kode satuan dari backend.
// Formulir yang baru memilih material memakai useUnitCode.
export function quantityText(quantity: number, unit: string): string {
  return unit === '' ? number(quantity) : `${number(quantity)} ${unit}`
}

export function warehouseText(id: string | null, name: string | null | undefined): string {
  if (id === null) {
    return '-'
  }
  return name ?? id.slice(0, 8)
}
