import type { ReactNode } from 'react'
import { useLocation } from 'react-router-dom'

import { canAccess, usePermissions } from '../../controllers/useAccess'
import { useCurrentUser } from '../../controllers/useAuth'
import { moduleForPath } from '../components/navigation'
import { NoAccessPage } from './NoAccessPage'
import { PendingApprovalPage } from './PendingApprovalPage'

// Penjaga rute: alamat yang diketik langsung tetap diperiksa walau menunya
// sudah disembunyikan. Modul rute dibaca dari navigation.ts. Akun yang belum
// diberi peran tidak melihat aplikasi sama sekali sampai admin memberi peran.
export function RouteAccess({ children }: { children: ReactNode }) {
  const { pathname } = useLocation()
  const user = useCurrentUser()
  const permissions = usePermissions()
  const module = moduleForPath(pathname)

  if (user !== null && !user.role) {
    return <PendingApprovalPage />
  }

  if (module !== null && !canAccess(permissions, module, 'read')) {
    return <NoAccessPage module={module} />
  }

  return children
}
