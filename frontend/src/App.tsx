import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

import { useSession, useSessionRestore } from './controllers/useAuth'
import { FinancialPage } from './views/FinancialPage'
import { LocationPage } from './views/LocationPage'
import { LoginPage } from './views/LoginPage'
import { OverviewPage } from './views/OverviewPage'
import { ProjectsPage } from './views/ProjectsPage'
import { RegisterPage } from './views/RegisterPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, refetchOnWindowFocus: false },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <SessionGate>
          <Routes>
            <Route path="/masuk" element={<GuestOnly><LoginPage /></GuestOnly>} />
            <Route path="/daftar" element={<GuestOnly><RegisterPage /></GuestOnly>} />
            <Route path="/" element={<RequireSession><OverviewPage /></RequireSession>} />
            <Route path="/proyek" element={<RequireSession><ProjectsPage /></RequireSession>} />
            <Route path="/keuangan" element={<RequireSession><FinancialPage /></RequireSession>} />
            <Route path="/lokasi" element={<RequireSession><LocationPage /></RequireSession>} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </SessionGate>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

function SessionGate({ children }: { children: ReactNode }) {
  const isRestored = useSessionRestore()

  if (!isRestored) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-slate-500">Memeriksa sesi...</p>
      </div>
    )
  }

  return children
}

function RequireSession({ children }: { children: ReactNode }) {
  const { accessToken } = useSession()

  if (accessToken === null) {
    return <Navigate to="/masuk" replace />
  }

  return children
}

function GuestOnly({ children }: { children: ReactNode }) {
  const { accessToken } = useSession()

  if (accessToken !== null) {
    return <Navigate to="/" replace />
  }

  return children
}
