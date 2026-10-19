import { fireEvent, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AssetOverview, InventoryOverview as Stock } from '../models/inventory'
import { InventoryOverview } from '../views/inventory/InventoryOverview'
import { renderWithProviders } from './render'

const STOCK: Stock = {
  stockValue: 1_000_000_000,
  lowStock: 4,
  warehouses: [{ name: 'Gudang Pusat Cikarang', value: 700_000_000, items: 9 }, { name: 'Gudang Proyek Sky', value: 300_000_000, items: 3 }],
  weeks: [{ week: '2026-10-12', inValue: 50_000_000, outValue: 20_000_000 }],
}

const ASSETS: AssetOverview = {
  bookValue: 5_000_000_000,
  categories: [{ category: 'Alat berat', count: 4, bookValue: 4_000_000_000 }],
  equipmentStatus: { operating: 10, broken: 1 },
  serviceOverdue: 2,
  serviceSoon: 3,
  serviceWindow: 14,
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ringkasan inventaris', () => {
  it('menggabungkan stok dan aset lalu membuka tab rinciannya', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url: string) => {
      const body = String(url).includes('/assets/overview') ? ASSETS : STOCK
      return new Response(JSON.stringify({ data: body }), { headers: { 'Content-Type': 'application/json' } })
    }))
    const drill = { openStock: vi.fn(), openAssets: vi.fn() }

    renderWithProviders(<InventoryOverview drill={drill} />)

    expect(await screen.findByText('Gudang Pusat Cikarang')).toBeInTheDocument()
    expect(screen.getByText('dalam 14 hari')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Stok menipis/ }))
    expect(drill.openStock).toHaveBeenCalled()
    fireEvent.click(screen.getByRole('button', { name: /Servis terlambat/ }))
    expect(drill.openAssets).toHaveBeenCalled()
  })
})
