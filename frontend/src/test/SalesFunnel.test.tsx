import { screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { User } from '../models/auth'
import { SalesFunnel } from '../views/overview/SalesFunnel'
import { renderWithProviders } from './render'

const UNITS = [
  { status: 'available', total: 10 },
  { status: 'on_hold', total: 4 },
  { status: 'reserved', total: 3 },
  { status: 'sold', total: 4 },
]

const CONTRACTS = { count: 8, totalValue: 3_475_000_000, byStatus: { draft: 1, active: 2, paid: 4, cancelled: 1 } }

function mockApi(units: unknown) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)
      const body = url.includes('/sales/units/summary') ? { data: units } : { data: CONTRACTS }
      return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }),
  )
}

beforeEach(() => {
  const user: User = {
    id: 'u-1', name: 'Uji', email: 'uji@contoh.id', active: true,
    role: { code: 'uji', name: 'Uji' }, permissions: { sales: 'read' }, createdAt: '2026-01-01T00:00:00Z',
  }
  useSessionStore.setState({ accessToken: null, user })
})

afterEach(() => {
  vi.unstubAllGlobals()
  useSessionStore.setState({ accessToken: null, user: null })
})

describe('SalesFunnel', () => {
  it('menampilkan empat tahap dengan konversi antartahap', async () => {
    mockApi(UNITS)
    renderWithProviders(<SalesFunnel />)

    expect(await screen.findByText('Unit dipasarkan')).toBeInTheDocument()
    expect(screen.getByText('17')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
    expect(screen.getByText('konversi 41%')).toBeInTheDocument()
    expect(screen.getByText('konversi 57%')).toBeInTheDocument()
    expect(screen.getByText('24%')).toBeInTheDocument()
  })

  it('memberi pesan kosong kalau belum ada unit dipasarkan', async () => {
    mockApi([{ status: 'on_hold', total: 2 }])
    renderWithProviders(<SalesFunnel />)

    expect(await screen.findByText('Belum ada unit yang dipasarkan.')).toBeInTheDocument()
  })
})
