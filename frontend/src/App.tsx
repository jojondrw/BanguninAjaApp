import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { lazy, type ReactNode, Suspense } from 'react'
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom'

import { useSession, useSessionRestore } from './controllers/useAuth'
import { RouteAccess } from './views/access/RouteAccess'
import { BaniScreen } from './views/components/BaniLoader'
import { ErrorBoundary } from './views/components/ErrorBoundary'
import { LandingPage } from './views/LandingPage'
import { LoginPage } from './views/LoginPage'
import { RegisterPage } from './views/RegisterPage'

// Halaman dimuat saat dibuka supaya muatan awal kecil. Masuk dan Daftar
// tetap ikut paket utama karena itu layar pertama yang dilihat.
const AccountPage = lazy(() => import('./views/account/AccountPage').then((module) => ({ default: module.AccountPage })))
const BillingPage = lazy(() => import('./views/BillingPage').then((module) => ({ default: module.BillingPage })))
const FinancialPage = lazy(() => import('./views/FinancialPage').then((module) => ({ default: module.FinancialPage })))
const HrPage = lazy(() => import('./views/HrPage').then((module) => ({ default: module.HrPage })))
const InventoryPage = lazy(() => import('./views/InventoryPage').then((module) => ({ default: module.InventoryPage })))
const MasterDataPage = lazy(() => import('./views/master/MasterDataPage').then((module) => ({ default: module.MasterDataPage })))
const LocationPage = lazy(() => import('./views/LocationPage').then((module) => ({ default: module.LocationPage })))
const OverviewPage = lazy(() => import('./views/OverviewPage').then((module) => ({ default: module.OverviewPage })))
const PayPage = lazy(() => import('./views/pay/PayPage').then((module) => ({ default: module.PayPage })))
const VerifyWagePage = lazy(() => import('./views/verify/WageProofPages').then((module) => ({ default: module.VerifyWagePage })))
const WageReceiptPage = lazy(() => import('./views/verify/WageProofPages').then((module) => ({ default: module.WageReceiptPage })))
const VerifyClaimPage = lazy(() => import('./views/verify/VerifyClaimPage').then((module) => ({ default: module.VerifyClaimPage })))
const ProcurementPage = lazy(() => import('./views/ProcurementPage').then((module) => ({ default: module.ProcurementPage })))
const ProjectsPage = lazy(() => import('./views/ProjectsPage').then((module) => ({ default: module.ProjectsPage })))
const ProjectWorkspacePage = lazy(() => import('./views/ProjectWorkspacePage').then((module) => ({ default: module.ProjectWorkspacePage })))
const ReportsPage = lazy(() => import('./views/ReportsPage').then((module) => ({ default: module.ReportsPage })))
const SalesPage = lazy(() => import('./views/SalesPage').then((module) => ({ default: module.SalesPage })))
const UsersPage = lazy(() => import('./views/users/UsersPage').then((module) => ({ default: module.UsersPage })))

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
          <RouteBoundary>
            <Suspense fallback={<PageFallback />}>
              <Routes>
                <Route path="/masuk" element={<GuestOnly><LoginPage /></GuestOnly>} />
                <Route path="/bayar/:token" element={<PayPage />} />
                <Route path="/verifikasi/:token" element={<VerifyClaimPage />} />
                <Route path="/verifikasi-upah/:token" element={<VerifyWagePage />} />
                <Route path="/kuitansi-upah/:token" element={<WageReceiptPage />} />
                <Route path="/daftar" element={<GuestOnly><RegisterPage /></GuestOnly>} />
                <Route path="/" element={<HomeRoute />} />
                <Route path="/proyek" element={<RequireSession><ProjectsPage /></RequireSession>} />
                <Route path="/proyek/:id" element={<RequireSession><ProjectWorkspacePage /></RequireSession>} />
                <Route path="/keuangan" element={<RequireSession><FinancialPage /></RequireSession>} />
              <Route path="/tagihan" element={<RequireSession><BillingPage /></RequireSession>} />
                <Route path="/lokasi" element={<RequireSession><LocationPage /></RequireSession>} />
                <Route path="/penjualan" element={<RequireSession><SalesPage /></RequireSession>} />
                <Route path="/pengadaan" element={<RequireSession><ProcurementPage /></RequireSession>} />
                <Route path="/sdm" element={<RequireSession><HrPage /></RequireSession>} />
                <Route path="/inventaris" element={<RequireSession><InventoryPage /></RequireSession>} />
                <Route path="/data-master" element={<RequireSession><MasterDataPage /></RequireSession>} />
                <Route path="/akun" element={<RequireSession><AccountPage /></RequireSession>} />
                <Route path="/pengguna" element={<RequireSession><UsersPage /></RequireSession>} />
                <Route path="/laporan" element={<RequireSession><ReportsPage /></RequireSession>} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </Suspense>
          </RouteBoundary>
        </SessionGate>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

function RouteBoundary({ children }: { children: ReactNode }) {
  const { pathname } = useLocation()
  return <ErrorBoundary resetKey={pathname}>{children}</ErrorBoundary>
}

function PageFallback() {
  return (
    <BaniScreen label="Memuat halaman..." />
  )
}

function SessionGate({ children }: { children: ReactNode }) {
  const isRestored = useSessionRestore()

  if (!isRestored) {
    return (
      <BaniScreen label="Memeriksa sesi..." />
    )
  }

  return children
}

function RequireSession({ children }: { children: ReactNode }) {
  const { accessToken } = useSession()

  if (accessToken === null) {
    return <Navigate to="/masuk" replace />
  }

  return <RouteAccess>{children}</RouteAccess>
}

// Pengunjung yang belum masuk melihat halaman depan; yang sudah masuk langsung
// melihat Ringkasan di alamat yang sama.
function HomeRoute() {
  const { accessToken } = useSession()

  if (accessToken === null) {
    return <LandingPage />
  }

  return <RouteAccess><OverviewPage /></RouteAccess>
}

function GuestOnly({ children }: { children: ReactNode }) {
  const { accessToken } = useSession()

  if (accessToken !== null) {
    return <Navigate to="/" replace />
  }

  return children
}
