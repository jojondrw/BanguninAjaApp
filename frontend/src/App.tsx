import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

import { usePemulihanSesi, useSesi } from './controllers/useAuth'
import { DaftarPage } from './views/DaftarPage'
import { DasborPage } from './views/DasborPage'
import { MasukPage } from './views/MasukPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, refetchOnWindowFocus: false },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <PemulihSesi>
          <Routes>
            <Route path="/masuk" element={<HanyaTamu><MasukPage /></HanyaTamu>} />
            <Route path="/daftar" element={<HanyaTamu><DaftarPage /></HanyaTamu>} />
            <Route path="/" element={<ButuhSesi><DasborPage /></ButuhSesi>} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </PemulihSesi>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

function PemulihSesi({ children }: { children: ReactNode }) {
  const sudahDipulihkan = usePemulihanSesi()

  if (!sudahDipulihkan) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-slate-500">Memeriksa sesi...</p>
      </div>
    )
  }

  return children
}

function ButuhSesi({ children }: { children: ReactNode }) {
  const { accessToken } = useSesi()

  if (accessToken === null) {
    return <Navigate to="/masuk" replace />
  }

  return children
}

function HanyaTamu({ children }: { children: ReactNode }) {
  const { accessToken } = useSesi()

  if (accessToken !== null) {
    return <Navigate to="/" replace />
  }

  return children
}
