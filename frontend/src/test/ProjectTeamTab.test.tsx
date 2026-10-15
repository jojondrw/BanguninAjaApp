import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { Permissions, User } from '../models/auth'
import type { Project } from '../models/project'
import { toMemberCreateRequest, type ProjectMember } from '../models/projectMember'
import { AccessScope } from '../views/access/AccessScope'
import { ProjectTeamTab } from '../views/project/ProjectTeamTab'
import { renderWithProviders } from './render'

const memberHooks = vi.hoisted(() => ({
  useProjectMembers: vi.fn(),
  useAddProjectMember: vi.fn(),
  useUpdateProjectMember: vi.fn(),
  useRemoveProjectMember: vi.fn(),
}))

vi.mock('../controllers/useProjectMembers', () => memberHooks)

const PROJECT: Project = {
  id: 'p-1', code: 'UJI-1', name: 'Griya Uji', regionId: null, type: 'Perumahan', status: 'ongoing', progress: 40,
  startDate: null, targetEndDate: null, contractValue: 0, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z',
}

const MEMBERS: ProjectMember[] = [
  {
    id: 'm-1', projectId: 'p-1', userId: 'u-1', name: 'Rina Wulandari', email: 'rina@contoh.id', active: true,
    position: 'Site manager', role: { code: 'manajer_proyek', name: 'Manajer Proyek' },
    joinedAt: '2026-10-01T08:00:00Z', updatedAt: '2026-10-01T08:00:00Z',
  },
  {
    id: 'm-2', projectId: 'p-1', userId: 'u-2', name: 'Budi Santoso', email: 'budi@contoh.id', active: false,
    position: null, role: { code: 'keuangan', name: 'Keuangan' },
    joinedAt: '2026-10-02T08:00:00Z', updatedAt: '2026-10-02T08:00:00Z',
  },
]

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

function renderTab() {
  renderWithProviders(
    <AccessScope module="projects">
      <ProjectTeamTab project={PROJECT} />
    </AccessScope>,
  )
  return userEvent.setup()
}

describe('ProjectTeamTab', () => {
  let update: ReturnType<typeof mutation>
  let remove: ReturnType<typeof mutation>

  beforeEach(() => {
    update = mutation()
    remove = mutation()
    memberHooks.useProjectMembers.mockReturnValue({ isPending: false, isError: false, data: MEMBERS, refetch: vi.fn() })
    memberHooks.useAddProjectMember.mockReturnValue(mutation())
    memberHooks.useUpdateProjectMember.mockReturnValue(update)
    memberHooks.useRemoveProjectMember.mockReturnValue(remove)
  })

  afterEach(() => {
    useSessionStore.setState({ accessToken: null, user: null })
  })

  it('menampilkan nama, email, posisi, peran, dan penanda akun nonaktif', () => {
    signIn({ projects: 'write' })
    renderTab()

    const rina = screen.getByRole('row', { name: /Rina Wulandari/ })
    expect(within(rina).getByText('(Anda)')).toBeInTheDocument()
    expect(within(rina).getByText('rina@contoh.id')).toBeInTheDocument()
    expect(within(rina).getByText('Site manager')).toBeInTheDocument()
    expect(within(rina).getByText('Manajer Proyek')).toBeInTheDocument()

    const budi = screen.getByRole('row', { name: /Budi Santoso/ })
    expect(within(budi).getByText('Akun nonaktif')).toBeInTheDocument()
    expect(within(budi).getByText('-')).toBeInTheDocument()
  })

  it('menyembunyikan tambah, ubah, dan keluarkan untuk peran yang hanya boleh melihat proyek', () => {
    signIn({ projects: 'read' })
    renderTab()

    expect(screen.getByText('Rina Wulandari')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Tambah anggota/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Ubah' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Keluarkan' })).not.toBeInTheDocument()
    expect(screen.queryByRole('columnheader', { name: 'Aksi' })).not.toBeInTheDocument()
  })

  it('mengeluarkan anggota hanya setelah konfirmasi kedua', async () => {
    signIn({ projects: 'write' })
    const user = renderTab()
    const budi = screen.getByRole('row', { name: /Budi Santoso/ })

    await user.click(within(budi).getByRole('button', { name: 'Keluarkan' }))
    expect(remove.mutate).not.toHaveBeenCalled()

    await user.click(within(budi).getByRole('button', { name: 'Ya, keluarkan dari tim' }))
    expect(remove.mutate).toHaveBeenCalledWith('m-2', expect.anything())
  })

  it('mengubah posisi dan peran lewat formulir ubah', async () => {
    signIn({ projects: 'write' })
    const user = renderTab()
    const rina = screen.getByRole('row', { name: /Rina Wulandari/ })

    await user.click(within(rina).getByRole('button', { name: 'Ubah' }))
    const position = screen.getByLabelText('Posisi (opsional)')
    expect(position).toHaveValue('Site manager')

    await user.clear(position)
    await user.type(position, 'Estimator')
    await user.selectOptions(screen.getByLabelText('Peran proyek'), 'pengamat')
    await user.click(screen.getByRole('button', { name: 'Simpan perubahan anggota' }))

    expect(update.mutate).toHaveBeenCalledWith(
      { memberId: 'm-1', values: { position: 'Estimator', role: 'pengamat' } },
      expect.anything(),
    )
  })
})

describe('toMemberCreateRequest', () => {
  it('tidak mengirim posisi kosong dan peran yang mengikuti akun', () => {
    expect(toMemberCreateRequest({ userId: 'u-9', position: '   ', role: '' })).toEqual({ userId: 'u-9' })
  })

  it('mengirim posisi yang sudah dirapikan dan peran yang dipilih', () => {
    expect(toMemberCreateRequest({ userId: 'u-9', position: ' Estimator ', role: 'keuangan' })).toEqual({
      userId: 'u-9',
      position: 'Estimator',
      role: 'keuangan',
    })
  })
})
