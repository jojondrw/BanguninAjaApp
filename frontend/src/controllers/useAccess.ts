import { createContext, useContext } from 'react'

import type { AccessLevel, ModuleKey, Permissions } from '../models/auth'
import { useCurrentUser } from './useAuth'

// Modul halaman yang sedang dibuka. AppShell mengisinya dari rute, dan bagian
// halaman yang menulis ke modul lain bisa menimpanya lewat AccessScope.
// Tanpa modul (null) berarti tidak dibatasi, misalnya di luar AppShell.
export const AccessScopeContext = createContext<ModuleKey | null>(null)

export function canAccess(permissions: Permissions | undefined, module: ModuleKey, level: AccessLevel): boolean {
  const granted = permissions?.[module]
  if (level === 'read') {
    return granted === 'read' || granted === 'write'
  }
  return granted === 'write'
}

export function usePermissions(): Permissions {
  return useCurrentUser()?.permissions ?? {}
}

export function useCan(module: ModuleKey, level: AccessLevel): boolean {
  return canAccess(usePermissions(), module, level)
}

// Dipakai komponen bersama seperti FormToggle dan RowAction: cek hak pada modul
// halaman tempat komponen itu dipasang.
export function useScopedCan(level: AccessLevel): boolean {
  const module = useContext(AccessScopeContext)
  const permissions = usePermissions()
  return module === null || canAccess(permissions, module, level)
}
