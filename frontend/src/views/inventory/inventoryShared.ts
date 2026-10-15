import { number } from '../../shared/format'

export const PAGE_SIZE = 20

// Tabel stok, mutasi, dan stok menipis membawa kode satuan dari backend.
// Formulir yang baru memilih material memakai useUnitCode.
// Kode satuan dari data master ditulis polos (m2, m3); tampilkan pangkatnya.
export function unitText(unit: string): string {
  return unit.replace(/^(c?m|km)([23])$/i, (_, base: string, power: string) => `${base}${power === '2' ? '²' : '³'}`)
}

export function quantityText(quantity: number, unit: string): string {
  return unit === '' ? number(quantity) : `${number(quantity)} ${unitText(unit)}`
}

export function warehouseText(id: string | null, name: string | null | undefined): string {
  if (id === null) {
    return '-'
  }
  return name ?? id.slice(0, 8)
}
