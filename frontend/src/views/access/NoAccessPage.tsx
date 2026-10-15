import { ShieldOff } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useCurrentUser } from '../../controllers/useAuth'
import type { ModuleKey } from '../../models/auth'
import { AppShell } from '../components/AppShell'
import { Card } from '../components/Data'
import { BUTTON_BASE, BUTTON_PRIMARY } from '../components/Form'
import { MENU_ITEMS } from '../components/navigation'

export function NoAccessPage({ module }: { module: ModuleKey }) {
  const user = useCurrentUser()
  const page = MENU_ITEMS.find((item) => item.module === module)?.label ?? 'ini'
  const role = user?.role?.name

  return (
    <AppShell title="Tidak punya akses" description={`Halaman ${page} tidak terbuka untuk peran Anda`}>
      <Card title={`Anda belum bisa membuka ${page}`}>
        <div className="flex flex-col items-start gap-4 sm:flex-row sm:items-center">
          <span className="grid size-10 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-500">
            <ShieldOff aria-hidden="true" className="size-4" strokeWidth={1.8} />
          </span>
          <p className="max-w-xl flex-1 text-[13px] text-slate-600">
            {role ? `Peran Anda saat ini ${role}.` : 'Akun Anda belum diberi peran.'} Minta admin mengubah peran Anda
            di halaman Pengguna kalau Anda perlu membuka {page}.
          </p>
          <Link to="/" className={`h-9 ${BUTTON_BASE} ${BUTTON_PRIMARY}`}>
            Kembali ke Ringkasan
          </Link>
        </div>
      </Card>
    </AppShell>
  )
}
