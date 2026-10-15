import { screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '../controllers/sessionStore'
import type { Permissions, User } from '../models/auth'
import { OverviewPage } from '../views/OverviewPage'
import { AttentionList } from '../views/overview/AttentionList'
import { renderWithProviders } from './render'

const fetchMock = vi.fn((_url: string) => new Promise<Response>(() => {}))

function signIn(permissions: Permissions) {
  const user: User = {
    id: 'u-1', name: 'Uji', email: 'uji@contoh.id', active: true,
    role: { code: 'uji', name: 'Uji' }, permissions, createdAt: '2026-01-01T00:00:00Z',
  }
  useSessionStore.setState({ accessToken: null, user })
}

// Segmen pertama setelah /api, sama dengan cara backend memetakan modul.
function requestedSegments(): string[] {
  const segments = fetchMock.mock.calls.map(([url]) => new URL(url).pathname.replace(/^\/api\//, '').split('/')[0])
  return [...new Set(segments)].sort()
}

beforeEach(() => {
  fetchMock.mockClear()
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  vi.unstubAllGlobals()
  useSessionStore.setState({ accessToken: null, user: null })
})

describe('AttentionList', () => {
  it('only asks the server about modules the role can read', async () => {
    signIn({ sales: 'write', projects: 'read', billing: 'read', master: 'read', reports: 'read' })

    renderWithProviders(<AttentionList />)

    await waitFor(() => expect(fetchMock).toHaveBeenCalled())
    expect(requestedSegments()).toEqual(['sales'])
    expect(screen.getByText('Cicilan terlambat')).toBeInTheDocument()
    expect(screen.queryByText('Gaji belum dibayar')).not.toBeInTheDocument()
  })

  it('still asks every module for a role that reads them all', async () => {
    signIn({ sales: 'read', procurement: 'read', inventory: 'read', hr: 'read' })

    renderWithProviders(<AttentionList />)

    await waitFor(() => expect(requestedSegments()).toEqual(['hr', 'inventory', 'procurement', 'sales']))
  })
})

describe('OverviewPage', () => {
  it('skips finance requests for a role without finance access', async () => {
    signIn({ hr: 'write', projects: 'read', master: 'read', reports: 'read' })

    renderWithProviders(<OverviewPage />)

    await waitFor(() => expect(requestedSegments()).toEqual(['hr', 'projects']))
    expect(screen.getAllByText('butuh akses Keuangan')).toHaveLength(2)
  })

  it('skips project requests for a role without project access', async () => {
    signIn({ finance: 'read' })

    renderWithProviders(<OverviewPage />)

    await waitFor(() => expect(requestedSegments()).toEqual(['finance']))
    expect(screen.getAllByText('butuh akses Proyek')).toHaveLength(2)
  })
})
