import { describe, expect, it } from 'vitest'

import { moduleForPath, visibleMenu } from '../views/components/navigation'
import { canAccess } from '../controllers/useAccess'

describe('canAccess', () => {
  it('treats write as including read', () => {
    expect(canAccess({ finance: 'write' }, 'finance', 'read')).toBe(true)
    expect(canAccess({ finance: 'write' }, 'finance', 'write')).toBe(true)
  })

  it('keeps read-only modules read-only and missing modules closed', () => {
    expect(canAccess({ sales: 'read' }, 'sales', 'write')).toBe(false)
    expect(canAccess({ sales: 'read' }, 'hr', 'read')).toBe(false)
    expect(canAccess(undefined, 'projects', 'read')).toBe(false)
  })
})

describe('navigation access', () => {
  it('maps routes and sub-routes to their module', () => {
    expect(moduleForPath('/proyek')).toBe('projects')
    expect(moduleForPath('/proyek/abc')).toBe('projects')
    expect(moduleForPath('/pengguna')).toBe('users')
    expect(moduleForPath('/')).toBeNull()
    expect(moduleForPath('/proyekx')).toBeNull()
  })

  it('hides menu items and empty groups the role cannot read', () => {
    const permissions = { hr: 'write', projects: 'read' } as const
    const menu = visibleMenu((module) => canAccess(permissions, module, 'read'))
    const labels = menu.flatMap((group) => group.items.map((item) => item.label))

    expect(labels).toEqual(['Ringkasan', 'Proyek', 'SDM'])
    expect(menu.map((group) => group.label)).not.toContain('Pengaturan')
  })
})
