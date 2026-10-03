import type { ReactNode } from 'react'
import { useLocation } from 'react-router-dom'

import { canAccess, usePermissions } from '../../controllers/useAccess'
import { moduleForPath } from '../components/navigation'
import { NoAccessPage } from './NoAccessPage'

// Penjaga rute: alamat yang diketik langsung tetap diperiksa walau menunya
// sudah disembunyikan. Modul rute dibaca dari navigation.ts.
export function RouteAccess({ children }: { children: ReactNode }) {
  const { pathname } = useLocation()
  const permissions = usePermissions()
  const module = moduleForPath(pathname)

  if (module !== null && !canAccess(permissions, module, 'read')) {
    return <NoAccessPage module={module} />
  }

  return children
}
