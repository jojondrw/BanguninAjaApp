import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, type RenderOptions } from '@testing-library/react'
import type { ReactElement, ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'

// QueryClient baru per tes supaya cache tidak bocor antartes. Retry dimatikan
// supaya galat langsung terlihat.
export function testQueryClient(): QueryClient {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
}

export function renderWithProviders(ui: ReactElement, options: RenderOptions & { route?: string } = {}) {
  const client = testQueryClient()
  const { route = '/', ...rest } = options

  function Providers({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }

  return { client, ...render(ui, { wrapper: Providers, ...rest }) }
}
