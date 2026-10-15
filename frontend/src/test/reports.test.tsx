import { screen, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { Permissions, User } from '../models/auth'
import { readableReportTypes, REPORT_TYPES } from '../models/report'
import { ReportsPage } from '../views/ReportsPage'
import { renderWithProviders } from './render'

const SALES_ROLE: Permissions = { sales: 'write', projects: 'read', billing: 'read', master: 'read', reports: 'read' }

function signIn(permissions: Permissions) {
  const user: User = {
    id: 'u-1', name: 'Uji', email: 'uji@contoh.id', active: true,
    role: { code: 'penjualan', name: 'Penjualan' }, permissions, createdAt: '2026-01-01T00:00:00Z',
  }
  useSessionStore.setState({ accessToken: null, user })
}

describe('readableReportTypes', () => {
  it('maps every report kind to the module whose data it reads', () => {
    const modules = Object.fromEntries(REPORT_TYPES.map((type) => [type.value, type.module]))
    expect(modules).toEqual({
      arus_kas: 'finance',
      anggaran: 'finance',
      proyek: 'projects',
      piutang_utang: 'billing',
      stok: 'inventory',
      penggajian: 'hr',
      penjualan: 'sales',
    })
  })

  it('keeps only kinds whose module the role can read, in catalog order', () => {
    const readable = readableReportTypes((module) => SALES_ROLE[module] !== undefined)
    expect(readable.map((type) => type.value)).toEqual(['proyek', 'piutang_utang', 'penjualan'])
  })
})

describe('ReportsPage', () => {
  beforeEach(() => {
    // Permintaan ke server dibiarkan menggantung: yang diuji hanya pilihan di layar.
    vi.stubGlobal('fetch', vi.fn(() => new Promise(() => {})))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    useSessionStore.setState({ accessToken: null, user: null })
  })

  it('offers only readable kinds and hides creating for a read-only role', () => {
    signIn(SALES_ROLE)
    renderWithProviders(<ReportsPage />, { route: '/laporan' })

    const kinds = within(screen.getByLabelText('Jenis laporan')).getAllByRole('option')
    expect(kinds.map((option) => option.textContent)).toEqual(['Proyek', 'Umur piutang dan utang', 'Penjualan'])
    const historyKinds = within(screen.getByLabelText('Saring jenis laporan')).getAllByRole('option')
    expect(historyKinds.map((option) => option.textContent)).toEqual([
      'Semua jenis', 'Proyek', 'Umur piutang dan utang', 'Penjualan',
    ])
    expect(screen.queryByRole('button', { name: 'Unduh laporan' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Lihat pratinjau' })).toBeInTheDocument()
  })

  it('lets a role that writes reports download', () => {
    signIn({ ...SALES_ROLE, finance: 'write', reports: 'write' })
    renderWithProviders(<ReportsPage />, { route: '/laporan' })

    expect(screen.getByLabelText('Jenis laporan')).toHaveValue('arus_kas')
    expect(screen.getByRole('button', { name: 'Unduh laporan' })).toBeInTheDocument()
  })

  it('shows an empty state when no report kind is readable', () => {
    signIn({ master: 'read', reports: 'read' })
    renderWithProviders(<ReportsPage />, { route: '/laporan' })

    expect(screen.getByRole('heading', { name: 'Belum ada laporan untuk peran Anda' })).toBeInTheDocument()
    expect(screen.queryByLabelText('Jenis laporan')).not.toBeInTheDocument()
    expect(fetch).not.toHaveBeenCalled()
  })
})
