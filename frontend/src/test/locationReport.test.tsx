import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { User } from '../models/auth'
import { SiteReportPanel } from '../views/location/ReportDownload'
import { renderWithProviders } from './render'

const LOCATION_ID = 'loc-1'
const FILE_NAME = 'analisis-lokasi-uji-lahan-cibubur-2026-10-15.pdf'

const USER: User = {
  id: 'u-1', name: 'Uji', email: 'uji@contoh.id', active: true,
  role: { code: 'admin', name: 'Admin' }, permissions: {}, createdAt: '2026-01-01T00:00:00Z',
} as User

function pdfResponse(): Response {
  return new Response('%PDF-1.4', {
    status: 200,
    headers: { 'Content-Type': 'application/pdf', 'Content-Disposition': `attachment; filename=${FILE_NAME}` },
  })
}

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function authorizationOf(call: unknown[]): string | undefined {
  const init = call[1] as RequestInit | undefined
  return (init?.headers as Record<string, string> | undefined)?.Authorization
}

// Unduhan PDF memakai fetch dengan header Authorization lalu disimpan lewat
// object URL, sama dengan halaman Laporan.
describe('SiteReportPanel', () => {
  const fetchMock = vi.fn<typeof fetch>()

  beforeEach(() => {
    useSessionStore.setState({ accessToken: 'token-lama', user: USER })
    vi.stubGlobal('fetch', fetchMock)
    URL.createObjectURL = vi.fn(() => 'blob:laporan')
    URL.revokeObjectURL = vi.fn()
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
  })

  afterEach(() => {
    fetchMock.mockReset()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('mengunduh PDF lokasi dan menyebut nama berkasnya', async () => {
    fetchMock.mockResolvedValueOnce(pdfResponse())
    renderWithProviders(<SiteReportPanel savedLocationId={LOCATION_ID} />)

    await userEvent.setup().click(screen.getByRole('button', { name: /Unduh laporan \(PDF\)/ }))

    expect(await screen.findByText(`Berkas ${FILE_NAME} sudah diunduh.`)).toBeInTheDocument()
    expect(String(fetchMock.mock.calls[0][0])).toMatch(/\/locations\/saved\/loc-1\/report$/)
    expect(authorizationOf(fetchMock.mock.calls[0])).toBe('Bearer token-lama')
    expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
    expect(HTMLAnchorElement.prototype.click).toHaveBeenCalledTimes(1)
  })

  it('mencoba sekali lagi dengan token baru setelah sesi disegarkan', async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse(401, { error: { code: 'token_expired', message: 'Sesi habis' } }))
      .mockResolvedValueOnce(jsonResponse(200, { data: { accessToken: 'token-baru', user: USER } }))
      .mockResolvedValueOnce(pdfResponse())
    renderWithProviders(<SiteReportPanel savedLocationId={LOCATION_ID} />)

    await userEvent.setup().click(screen.getByRole('button', { name: /Unduh laporan \(PDF\)/ }))

    expect(await screen.findByText(`Berkas ${FILE_NAME} sudah diunduh.`)).toBeInTheDocument()
    expect(String(fetchMock.mock.calls[1][0])).toMatch(/\/auth\/refresh$/)
    expect(authorizationOf(fetchMock.mock.calls[2])).toBe('Bearer token-baru')
  })

  it('menampilkan pesan backend saat lokasi tidak ditemukan', async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(404, { error: { code: 'saved_location_not_found', message: 'Lokasi tersimpan tidak ditemukan' } }),
    )
    renderWithProviders(<SiteReportPanel savedLocationId={LOCATION_ID} />)

    await userEvent.setup().click(screen.getByRole('button', { name: /Unduh laporan \(PDF\)/ }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Lokasi tersimpan tidak ditemukan')
    expect(URL.createObjectURL).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: /Unduh laporan \(PDF\)/ })).toBeEnabled()
  })
})
