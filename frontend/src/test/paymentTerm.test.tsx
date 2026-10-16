import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { Permissions, User } from '../models/auth'
import {
  changeTrail,
  nextTermForm,
  remainingPercent,
  termNote,
  toContractChangeRequest,
  toContractRequest,
  type ContractChange,
  type PaymentSchedule,
  type PaymentTerm,
} from '../models/paymentTerm'
import type { Project } from '../models/project'
import { ContractChangesCard } from '../views/project/ContractChangesCard'
import { PaymentTermsCard } from '../views/project/PaymentTermsCard'
import { renderWithProviders } from './render'

const termHooks = vi.hoisted(() => ({
  usePaymentSchedule: vi.fn(),
  useSaveContract: vi.fn(),
  useCreatePaymentTerm: vi.fn(),
  useDeletePaymentTerm: vi.fn(),
  useClaimPaymentTerm: vi.fn(),
  useReleaseRetention: vi.fn(),
  useProposeContractChange: vi.fn(),
  useDecideContractChange: vi.fn(),
  useDeleteContractChange: vi.fn(),
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
    number: 'SPK-UJI-01', value: 23_500_000_000, currentValue: 23_500_000_000, approvedChange: 0, extensionDays: 0,
    retentionPercent: 5,
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
  changes: [],
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

function signIn(permissions: Permissions, role = { code: 'manajer_proyek', name: 'Manajer Proyek' }) {
  const user: User = {
    id: 'u-1', name: 'Rina Wulandari', email: 'rina@contoh.id', active: true,
    role, permissions, createdAt: '2026-01-01T00:00:00Z',
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
    termHooks.useProposeContractChange.mockReturnValue(mutation())
    termHooks.useDecideContractChange.mockReturnValue(mutation())
    termHooks.useDeleteContractChange.mockReturnValue(mutation())
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

const PROPOSED: ContractChange = {
  id: 'c-2', label: 'Adendum 2', description: 'Pengurangan pekerjaan lansekap', valueChange: -350_000_000, daysChange: 0,
  status: 'proposed', proposerName: 'Rina Wulandari', deciderName: null, decidedOn: null, createdAt: '2026-10-14T03:00:00Z',
}

describe('ContractChangesCard', () => {
  let decide: ReturnType<typeof mutation>

  beforeEach(() => {
    decide = mutation()
    showSchedule({ ...SCHEDULE, changes: [PROPOSED] })
    termHooks.useProposeContractChange.mockReturnValue(mutation())
    termHooks.useDecideContractChange.mockReturnValue(decide)
    termHooks.useDeleteContractChange.mockReturnValue(mutation())
  })

  afterEach(() => {
    useSessionStore.setState({ accessToken: null, user: null })
  })

  it('membiarkan Admin menyetujui usulan lewat konfirmasi', async () => {
    signIn({ projects: 'write' }, { code: 'admin', name: 'Admin' })
    renderWithProviders(<ContractChangesCard project={PROJECT} />)
    const user = userEvent.setup()

    await user.click(screen.getByRole('button', { name: 'Setujui' }))
    await user.click(screen.getByRole('button', { name: 'Ya, setujui adendum 2' }))

    expect(decide.mutate).toHaveBeenCalledWith({ changeId: 'c-2', status: 'approved' })
  })

  it('tidak memberi Manajer Proyek tombol persetujuan', () => {
    signIn({ projects: 'write' })
    renderWithProviders(<ContractChangesCard project={PROJECT} />)

    expect(screen.queryByRole('button', { name: 'Setujui' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Hapus' })).toBeInTheDocument()
    expect(screen.getByText('Usulan adendum menunggu persetujuan Admin.')).toBeInTheDocument()
  })

  it('tidak tampil sebelum kontrak kontraktor diatur', () => {
    signIn({ projects: 'write' })
    showSchedule({ ...SCHEDULE, contract: null })
    renderWithProviders(<ContractChangesCard project={PROJECT} />)

    expect(screen.queryByText('Pekerjaan tambah-kurang')).not.toBeInTheDocument()
  })
})

describe('model adendum', () => {
  it('mengirim pekerjaan kurang sebagai nilai negatif dari angka positif', () => {
    expect(toContractChangeRequest({ kind: 'deduct', amount: '350000000', days: '', description: ' Lansekap ' })).toEqual({
      description: 'Lansekap', valueChange: -350_000_000, daysChange: 0,
    })
    expect(toContractChangeRequest({ kind: 'add', amount: '', days: '30', description: 'Waktu saja' })).toEqual({
      description: 'Waktu saja', valueChange: 0, daysChange: 30,
    })
  })

  it('menulis jejak siapa mengusulkan dan siapa memutuskan', () => {
    const format = (iso: string) => iso.slice(0, 10)
    expect(changeTrail(PROPOSED, format)).toBe('diusulkan Rina Wulandari, 2026-10-14')
    expect(changeTrail({ ...PROPOSED, status: 'approved', deciderName: 'Budi', decidedOn: '2026-10-15' }, format)).toBe(
      'diusulkan Rina Wulandari, 2026-10-14; disetujui Budi, 2026-10-15',
    )
  })
})
