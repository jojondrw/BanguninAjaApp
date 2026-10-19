import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { busiestBudgets, topAccounts, type Budget, type CashAccountTotal } from '../models/finance'
import { DonutChart } from '../views/components/DonutChart'

const ACCOUNTS: CashAccountTotal[] = [
  { accountCode: '5100', accountName: 'Beban Konstruksi', type: 'out', total: 600 },
  { accountCode: '5500', accountName: 'Beban Gaji', type: 'out', total: 100 },
  { accountCode: '5600', accountName: 'Beban Kantor', type: 'out', total: 50 },
  { accountCode: '4100', accountName: 'Penjualan Unit', type: 'in', total: 900 },
]

function budget(id: string, absorption: number): Budget {
  return {
    id, projectId: id, projectName: id, year: 2026, value: 100, note: '', realized: absorption,
    remaining: 100 - absorption, absorption, createdAt: '', updatedAt: '',
  }
}

describe('ringkasan keuangan', () => {
  it('menampilkan akun terbesar dan menggabungkan sisanya jadi Lainnya', () => {
    const { total, items } = topAccounts(ACCOUNTS, 'out', 2)
    expect(total).toBe(750)
    expect(items.map((item) => item.label)).toEqual(['Beban Konstruksi', 'Beban Gaji', 'Lainnya'])
    expect(items[2].total).toBe(50)
  })

  it('mengurutkan anggaran dari serapan tertinggi', () => {
    expect(busiestBudgets([budget('a', 40), budget('b', 120), budget('c', 80)], 2).map((item) => item.id)).toEqual(['b', 'c'])
  })

  it('donat menulis total dan persentase tiap irisan', () => {
    render(<DonutChart caption="Kas keluar" colors={['#000', '#111']} slices={[{ label: 'A', value: 75 }, { label: 'B', value: 25 }]} />)
    expect(screen.getByRole('img', { name: 'Kas keluar' })).toBeInTheDocument()
    expect(screen.getByText('75%')).toBeInTheDocument()
    expect(screen.getByText('25%')).toBeInTheDocument()
  })
})
