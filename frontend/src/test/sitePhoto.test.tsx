import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Project } from '../models/project'
import { SITE_PHOTO_MAX_BYTES, sitePhotoProblem } from '../models/sitePhoto'
import { ProjectCard } from '../views/project/ProjectCard'
import { ProjectRow } from '../views/project/ProjectRow'
import { renderWithProviders } from './render'

const PROJECT: Project = {
  id: 'p-1',
  code: 'DEMO-RSIA',
  name: 'Rumah Sakit Ibu & Anak Cirebon',
  regionId: null,
  type: 'Rumah Sakit',
  status: 'ongoing',
  progress: 46,
  startDate: null,
  targetEndDate: '2027-06-01T00:00:00Z',
  contractValue: 67_000_000_000,
  coverPhotoId: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('foto lapangan', () => {
  it('menolak berkas yang bukan foto atau terlalu besar sebelum diunggah', () => {
    expect(sitePhotoProblem(null)).toMatch(/Pilih foto/)
    expect(sitePhotoProblem(new File(['%PDF'], 'a.pdf', { type: 'application/pdf' }))).toMatch(/JPG, PNG, atau WebP/)
    const big = new File([new Uint8Array(SITE_PHOTO_MAX_BYTES + 1)], 'besar.jpg', { type: 'image/jpeg' })
    expect(sitePhotoProblem(big)).toMatch(/3 MB/)
    expect(sitePhotoProblem(new File(['x'], 'ok.webp', { type: 'image/webp' }))).toBeNull()
  })

  it('kartu proyek tanpa foto menampilkan jenis proyek dan tidak meminta berkas', () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    renderWithProviders(<ProjectCard project={PROJECT} />)

    expect(screen.getByText('Rumah Sakit')).toBeInTheDocument()
    expect(screen.getByText('Belum ada foto lapangan')).toBeInTheDocument()
    expect(screen.getByRole('progressbar', { name: /Progres Rumah Sakit/ })).toHaveAttribute('aria-valuenow', '46')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('kartu proyek dengan foto memuat berkas sampulnya', async () => {
    URL.createObjectURL = vi.fn(() => 'blob:sampul')
    const fetchMock = vi.fn(async (_url: string) => new Response('x', { status: 200, headers: { 'Content-Type': 'image/jpeg' } }))
    vi.stubGlobal('fetch', fetchMock)
    renderWithProviders(<ProjectCard project={{ ...PROJECT, coverPhotoId: 'f-1' }} />)

    const image = await screen.findByRole('img', { name: /Kondisi lapangan/ })
    expect(image).toHaveAttribute('src', 'blob:sampul')
    expect(String(fetchMock.mock.calls[0][0])).toContain('/projects/p-1/site-photos/f-1/file')
  })

  it('baris tampilan daftar memuat foto di kiri dan keterangan proyek di kanan', async () => {
    URL.createObjectURL = vi.fn(() => 'blob:sampul')
    vi.stubGlobal('fetch', vi.fn(async () => new Response('x', { status: 200, headers: { 'Content-Type': 'image/jpeg' } })))
    renderWithProviders(<ProjectRow project={{ ...PROJECT, coverPhotoId: 'f-1' }} />)

    const link = screen.getByRole('link')
    const image = await screen.findByRole('img', { name: /Kondisi lapangan/ })
    expect(link).toHaveAttribute('href', '/proyek/p-1')
    expect(link.firstElementChild).toContainElement(image)
    expect(screen.getByText('Rumah Sakit Ibu & Anak Cirebon')).toBeInTheDocument()
    expect(screen.getByRole('progressbar', { name: /Progres Rumah Sakit/ })).toHaveAttribute('aria-valuenow', '46')
  })
})
