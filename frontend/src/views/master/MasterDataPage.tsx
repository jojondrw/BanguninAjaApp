import { useState } from 'react'

import { AccessScope } from '../access/AccessScope'
import { AppShell } from '../components/AppShell'
import { SectionTabs } from '../components/ListTools'
import { AccountsSection } from './AccountsSection'
import { RegionsSection } from './RegionsSection'
import { UnitsSection } from './UnitsSection'

type MasterTab = 'accounts' | 'regions' | 'units'

const TABS: { value: MasterTab; label: string }[] = [
  { value: 'accounts', label: 'Bagan akun' },
  { value: 'regions', label: 'Wilayah' },
  { value: 'units', label: 'Satuan' },
]

// Halaman ini selalu memakai hak modul master, jadi tombol tambah, ubah, dan
// hapus tetap tersembunyi untuk peran yang hanya boleh melihat, termasuk
// sebelum rutenya terdaftar di menu.
export function MasterDataPage() {
  const [tab, setTab] = useState<MasterTab>('accounts')

  return (
    <AppShell
      title="Data Master"
      description="Bagan akun, wilayah, dan satuan yang dipakai keuangan, proyek, inventaris, dan pengadaan"
    >
      <AccessScope module="master">
        <div className="mb-4">
          <SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian data master" />
        </div>

        {tab === 'accounts' ? <AccountsSection /> : null}
        {tab === 'regions' ? <RegionsSection /> : null}
        {tab === 'units' ? <UnitsSection /> : null}
      </AccessScope>
    </AppShell>
  )
}
