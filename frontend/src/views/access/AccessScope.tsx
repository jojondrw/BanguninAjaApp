import type { ReactNode } from 'react'

import { AccessScopeContext } from '../../controllers/useAccess'
import type { ModuleKey } from '../../models/auth'

// Bagian halaman yang menulis ke modul lain dari modul halamannya, misalnya tab
// Keuangan di ruang kerja proyek. Tombol bersama di dalamnya mengikuti hak modul ini.
export function AccessScope({ module, children }: { module: ModuleKey; children: ReactNode }) {
  return <AccessScopeContext.Provider value={module}>{children}</AccessScopeContext.Provider>
}
