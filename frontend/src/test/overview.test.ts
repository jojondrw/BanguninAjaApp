import { describe, expect, it } from 'vitest'

import type { CashFlowPeriod } from '../models/finance'
import { monthEndBalances, paidContractShare, percentOf, runwayMonths, unitFunnel } from '../models/overview'
import type { ContractSummary } from '../models/sales'

function period(net: number): CashFlowPeriod {
  return { period: '2026-01', cashIn: Math.max(net, 0), cashOut: Math.max(-net, 0), net }
}

const contracts: ContractSummary = {
  count: 8,
  totalValue: 3_475_000_000,
  byStatus: { draft: 1, active: 2, paid: 4, cancelled: 1 },
}

describe('dashboard Ringkasan', () => {
  it('menghitung saldo akhir bulan mundur dari saldo sekarang', () => {
    expect(monthEndBalances([period(-10), period(5), period(-2)], 100)).toEqual([97, 102, 100])
  })

  it('menghitung runway hanya saat kas terbakar', () => {
    expect(runwayMonths([period(-10), period(-20)], 100)).toBe(6)
    expect(runwayMonths([period(10), period(-5)], 100)).toBeNull()
    expect(runwayMonths([], 100)).toBeNull()
  })

  it('menyusun funnel unit tanpa unit yang ditahan', () => {
    const funnel = unitFunnel(
      [
        { status: 'available', total: 10 },
        { status: 'on_hold', total: 4 },
        { status: 'reserved', total: 3 },
        { status: 'sold', total: 4 },
      ],
      contracts,
    )
    expect(funnel.map((stage) => stage.value)).toEqual([17, 7, 4, 4])
  })

  it('menghitung persen dengan aman', () => {
    expect(percentOf(1, 3)).toBe(33)
    expect(percentOf(5, 0)).toBe(0)
    expect(paidContractShare(contracts)).toBe(67)
  })
})
