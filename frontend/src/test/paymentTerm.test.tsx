import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { Permissions, User } from '../models/auth'
import {
  nextTermForm,
  remainingPercent,
  termNote,
  toContractRequest,
  type PaymentSchedule,
  type PaymentTerm,
} from '../models/paymentTerm'
import type { Project } from '../models/project'
import { PaymentTermsCard } from '../views/project/PaymentTermsCard'
import { renderWithProviders } from './render'

const termHooks = vi.hoisted(() => ({
  usePaymentSchedule: vi.fn(),
  useSaveContract: vi.fn(),
  useCreatePaymentTerm: vi.fn(),
  useDeletePaymentTerm: vi.fn(),
  useClaimPaymentTerm: vi.fn(),
  useReleaseRetention: vi.fn(),
}))

vi.mock('../controllers/useProjectWorkspace', () => termHooks)

const PROJECT: Project = {
  id: 'p-1', code: 'UJI-RSIA', name: 'RSIA Uji', regionId: null, type: 'Rumah Sakit', status: 'ongoing', progress: 46,
  startDate: null, targetEndDate: null, contractValue: 0, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z',
}

function term(overrides: Partial<PaymentTerm>): PaymentTerm {
  return {
    id: 't-1', label: 'Termin 1', kind: 'progress', state: 'planned', requiredProgress: 0, percent: 10,
    grossAmount: 2_350_000_000, retentionAmount: 117_500_000, netAmount: 2_232_500_000, paidAmount: 0,
    payableId: null, claimedOn: null, claimedProgress: null, ...overrides,
  }
}

const SCHEDULE: PaymentSchedule = {
  contract: {
    vendorId: 'v-1', vendorName: 'PT Graha Struktur Nusantara', paymentTermDays: 30,
    number: 'SPK-UJI-01', value: 23_500_000_000, retentionPercent: 5,
  },
  progress: 46,
  terms: [
    term({ id: 't-1', state: 'paid', payableId: 'h-1', claimedOn: '2026-01-04', claimedProgress: 0, paidAmount: 2_232_500_000 }),
    term({
      id: 't-2', label: 'Termin 2', state: 'ready', requiredProgress: 40, percent: 20,
      grossAmount: 4_700_000_000, retentionAmount: 235_000_000, netAmount: 4_465_000_000,
    }),
    term({ id: 't-3', label: 'Termin 3', requiredProgress: 70, percent: 25 }),
  ],
  plannedPercent: 55,
  claimedAmount: 2_350_000_000,
  paidAmount: 2_232_500_000,
  retentionHeld: 117_500_000,
  retentionReleased: false,
  contractLocked: true,
  canReleaseRetention: false,
}

function mutation() {
  return { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: false, variables: undefined }
}

function signIn(permissions: Permissions) {
  const user: User = {
    id: 'u-1', name: 'Rina Wulandari', email: 'rina@contoh.id', active: true,
    role: { code: 'manajer_proyek', name: 'Manajer Proyek' }, permissions, createdAt: '2026-01-01T00:00:00Z',
  }
  useSessionStore.setState({ accessToken: null, user })
}

function showSchedule(schedule: PaymentSchedule) {
  termHooks.usePaymentSchedule.mockReturnValue({ isPending: false, isError: false, data: schedule, refetch: vi.fn() })
}

describe('PaymentTermsCard', () => {
  let claim: ReturnType<typeof mutation>

  beforeEach(() => {
    claim = mutation()
    showSchedule(SCHEDULE)
    termHooks.useSaveContract.mockReturnValue(mutation())
    termHooks.useCreatePaymentTerm.mockReturnValue(mutation())
    termHooks.useDeletePaymentTerm.mockReturnValue(mutation())
    termHooks.useClaimPaymentTerm.mockReturnValue(claim)
    termHooks.useReleaseRetention.mockReturnValue(mutation())
  })

  afterEach(() => {
    useSessionStore.setState({ accessToken: null, user: null })
  })

  it('mengajukan termin yang siap lewat konfirmasi dua langkah', async () => {
    signIn({ projects: 'write', finance: 'read' })
    renderWithProviders(<PaymentTermsCard project={PROJECT} />)
    const user = userEvent.setup()

    await user.click(screen.getByRole('button', { name: 'Ajukan' }))
    expect(claim.mutate).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: /Ya, ajukan/ }))
    expect(claim.mutate).toHaveBeenCalledWith('t-2')
  })

  it('hanya menampilkan tautan Tagihan untuk peran Keuangan yang tidak boleh mengajukan', () => {
    signIn({ projects: 'read', finance: 'write', billing: 'write' })
    renderWithProviders(<PaymentTermsCard project={PROJECT} />)

    expect(screen.queryByRole('button', { name: 'Ajukan' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Tambah termin' })).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Lihat di Tagihan' })).toHaveAttribute('href', '/tagihan')
  })

  it('menyebut kapan retensi bisa dilepas selama proyek belum selesai', () => {
    signIn({ projects: 'write' })
    renderWithProviders(<PaymentTermsCard project={PROJECT} />)

    expect(screen.getByText(/dilepas setelah proyek selesai dan diserahterimakan/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Lepas retensi' })).not.toBeInTheDocument()
  })

  it('mengajak mengatur kontrak kalau proyek belum punya kontrak kontraktor', () => {
    signIn({ projects: 'write' })
    showSchedule({ ...SCHEDULE, contract: null, terms: [], plannedPercent: 0, retentionHeld: 0, contractLocked: false })
    renderWithProviders(<PaymentTermsCard project={PROJECT} />)

    expect(screen.getByText(/Belum ada kontrak kontraktor/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Atur kontrak' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Tambah termin' })).not.toBeInTheDocument()
  })
})

describe('model termin', () => {
  it('mengisi termin berikutnya dengan sisa persen dan syarat progres di atas termin terakhir', () => {
    expect(nextTermForm(SCHEDULE)).toEqual({ requiredProgress: '95', percent: '45' })
    expect(nextTermForm({ plannedPercent: 0, terms: [] })).toEqual({ requiredProgress: '0', percent: '100' })
  })

  it('menghitung sisa persen tanpa galat pembulatan', () => {
    expect(remainingPercent({ plannedPercent: 33.33 + 33.33 + 33.33 })).toBe(0.01)
    expect(remainingPercent({ plannedPercent: 100 })).toBe(0)
  })

  it('mengubah isian kontrak menjadi request dengan nomor yang dirapikan', () => {
    expect(toContractRequest({ vendorId: 'v-1', number: '  SPK-01 ', value: '1500000', retentionPercent: '' })).toEqual({
      vendorId: 'v-1', number: 'SPK-01', value: 1_500_000, retentionPercent: 0,
    })
  })

  it('menjelaskan status termin dengan progres saat diajukan, bukan progres sekarang', () => {
    const format = (iso: string) => iso
    expect(termNote(SCHEDULE.terms[0], 46, format)).toBe('diajukan 2026-01-04 pada progres 0%')
    expect(termNote(SCHEDULE.terms[2], 46, format)).toBe('progres sekarang 46%')
    expect(termNote(term({ requiredProgress: 40 }), 46, format)).toBe('setelah termin sebelumnya')
    expect(termNote(SCHEDULE.terms[1], 46, format)).toBe('')
  })
})
