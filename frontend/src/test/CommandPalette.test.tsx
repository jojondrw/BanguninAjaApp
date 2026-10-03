import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useLocation } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { CommandPalette } from '../views/components/CommandPalette'
import { renderWithProviders } from './render'

const projectsHook = vi.hoisted(() => ({
  useProjects: vi.fn(),
}))

vi.mock('../controllers/useErp', () => projectsHook)

const PROJECTS = [
  { id: 'p-1', name: 'Griya Asri Cibubur', code: 'GAC' },
  { id: 'p-2', name: 'Bekasi Residence', code: 'BKR' },
]

function CurrentPath() {
  return <output aria-label="Alamat">{useLocation().pathname}</output>
}

function setup() {
  const onClose = vi.fn()
  renderWithProviders(
    <>
      <CommandPalette isOpen onClose={onClose} />
      <CurrentPath />
    </>,
  )
  return { onClose, user: userEvent.setup() }
}

describe('CommandPalette', () => {
  beforeEach(() => {
    projectsHook.useProjects.mockImplementation(({ search }: { search?: string }) => ({
      isPending: false,
      data: {
        items: PROJECTS.filter((project) => !search || project.name.toLowerCase().includes(search.toLowerCase())),
      },
    }))
  })

  it('tidak merender apa pun saat tertutup', () => {
    renderWithProviders(<CommandPalette isOpen={false} onClose={() => {}} />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('memfokuskan kotak cari dan menampilkan halaman serta proyek terbaru', () => {
    setup()
    expect(screen.getByRole('combobox', { name: 'Cari halaman atau proyek' })).toHaveFocus()
    expect(screen.getByRole('option', { name: /Ringkasan/ })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: /Griya Asri Cibubur/ })).toBeInTheDocument()
  })

  it('kata kunci "gaji" menemukan halaman SDM saja', async () => {
    const { user } = setup()
    await user.keyboard('gaji')

    await waitFor(() => {
      const pages = screen.getAllByRole('option').map((option) => option.textContent)
      expect(pages).toEqual(['SDM'])
    })
  })

  it('Enter membuka hasil yang disorot lalu menutup palet', async () => {
    const { onClose, user } = setup()
    await user.keyboard('gaji')
    await waitFor(() => expect(screen.getAllByRole('option')).toHaveLength(1))

    await user.keyboard('{Enter}')

    expect(onClose).toHaveBeenCalled()
    expect(screen.getByRole('status', { name: 'Alamat' })).toHaveTextContent('/sdm')
  })

  it('panah atas berputar ke proyek terakhir dan membuka halaman proyeknya', async () => {
    const { user } = setup()
    expect(screen.getByRole('option', { name: /Ringkasan/ })).toHaveAttribute('aria-selected', 'true')

    await user.keyboard('{ArrowUp}')
    expect(screen.getByRole('option', { name: /Bekasi Residence/ })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('option', { name: /Ringkasan/ })).toHaveAttribute('aria-selected', 'false')

    await user.keyboard('{Enter}')
    expect(screen.getByRole('status', { name: 'Alamat' })).toHaveTextContent('/proyek/p-2')
  })

  it('pencarian proyek ikut dikirim ke hook proyek', async () => {
    const { user } = setup()
    await user.keyboard('bekasi')

    await waitFor(() => expect(screen.getAllByRole('option')).toHaveLength(1))
    expect(screen.getByRole('option')).toHaveTextContent('Bekasi Residence')
    expect(projectsHook.useProjects).toHaveBeenLastCalledWith({ search: 'bekasi', pageSize: 5 })
  })

  it('Escape menutup palet', async () => {
    const { onClose, user } = setup()
    await user.keyboard('{Escape}')
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
