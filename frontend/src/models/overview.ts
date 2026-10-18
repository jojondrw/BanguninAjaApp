import type { CashFlowPeriod } from './finance'
import type { ContractSummary, UnitStatusCount } from './sales'

// Hitungan dashboard Ringkasan. Semuanya dari data ringkasan yang sudah ada,
// jadi tidak ada endpoint khusus dashboard.

export function percentOf(part: number, whole: number): number {
  return whole > 0 ? Math.round((part / whole) * 100) : 0
}

// Saldo kas di akhir setiap bulan, dihitung mundur dari saldo sekarang.
export function monthEndBalances(periods: CashFlowPeriod[], balance: number): number[] {
  const balances = new Array<number>(periods.length)
  let running = balance
  for (let index = periods.length - 1; index >= 0; index--) {
    balances[index] = running
    running -= periods[index].net
  }
  return balances
}

// Berapa bulan saldo kas cukup kalau arus kas bersih rata-rata tetap seperti
// periode ini. null kalau arus kas bersihnya tidak negatif (tidak membakar kas).
export function runwayMonths(periods: CashFlowPeriod[], balance: number): number | null {
  if (periods.length === 0 || balance <= 0) {
    return null
  }
  const averageNet = periods.reduce((sum, period) => sum + period.net, 0) / periods.length
  if (averageNet >= 0) {
    return null
  }
  return Math.floor(balance / -averageNet)
}

export interface FunnelStage {
  label: string
  value: number
}

// Unit yang ditahan (on_hold) belum dipasarkan, jadi tidak masuk funnel. Unit
// baru berstatus terjual saat kontraknya lunas, jadi tahap terakhir memakai
// status unit; tahap ketiga memakai kontrak yang sudah ditandatangani.
export function unitFunnel(units: UnitStatusCount[], contracts: ContractSummary | undefined): FunnelStage[] {
  const count = (status: UnitStatusCount['status']) => units.find((unit) => unit.status === status)?.total ?? 0
  const sold = count('sold')
  const reserved = count('reserved')
  const signed = contracts ? contracts.byStatus.active + contracts.byStatus.paid : 0
  return [
    { label: 'Unit dipasarkan', value: count('available') + reserved + sold },
    { label: 'Dipesan', value: reserved + sold },
    { label: 'Kontrak ditandatangani', value: signed },
    { label: 'Lunas dan terjual', value: sold },
  ]
}

// Kontrak yang sudah berjalan atau lunas; draf dan batal belum atau tidak lagi
// menghasilkan uang.
export function paidContractShare(contracts: ContractSummary): number {
  const live = contracts.byStatus.active + contracts.byStatus.paid
  return percentOf(contracts.byStatus.paid, live)
}
