import type { Material, UnitOfMeasure, Warehouse } from '../../models/inventory'
import type { Project } from '../../models/project'
import { number } from '../../shared/format'

export const PAGE_SIZE = 20

// Daftar pilihan untuk formulir dan penyaring, masing masing paling banyak 100
// data (ukuran halaman terbesar backend). Tabel stok dan mutasi tidak memakai
// ini untuk nama, karena backend sudah mengirim nama material dan gudang.
export interface Lookups {
  materials: Material[]
  warehouses: Warehouse[]
  units: UnitOfMeasure[]
  projects: Project[]
}

export function quantityText(quantity: number, unit: string): string {
  return unit === '' ? number(quantity) : `${number(quantity)} ${unit}`
}

export function unitCode(lookups: Lookups, unitId: string | undefined): string {
  return lookups.units.find((unit) => unit.id === unitId)?.code ?? ''
}

export function projectLabel(lookups: Lookups, id: string | null, empty: string): string {
  if (id === null) {
    return empty
  }
  return lookups.projects.find((project) => project.id === id)?.name ?? id.slice(0, 8)
}

export function warehouseText(id: string | null, name: string | null | undefined): string {
  if (id === null) {
    return '-'
  }
  return name ?? id.slice(0, 8)
}
