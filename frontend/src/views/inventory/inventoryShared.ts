import type { UnitOfMeasure } from '../../models/inventory'
import { number } from '../../shared/format'

export const PAGE_SIZE = 20

// Daftar satuan untuk label jumlah. Material, gudang, dan proyek dipilih lewat
// SearchSelect, dan tabel stok serta mutasi sudah membawa namanya dari backend.
export interface Lookups {
  units: UnitOfMeasure[]
}

export function quantityText(quantity: number, unit: string): string {
  return unit === '' ? number(quantity) : `${number(quantity)} ${unit}`
}

export function unitCode(lookups: Lookups, unitId: string | undefined): string {
  return lookups.units.find((unit) => unit.id === unitId)?.code ?? ''
}

export function warehouseText(id: string | null, name: string | null | undefined): string {
  if (id === null) {
    return '-'
  }
  return name ?? id.slice(0, 8)
}
