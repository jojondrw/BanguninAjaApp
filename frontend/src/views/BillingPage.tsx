import { useState } from 'react'

import { useProjects } from '../controllers/useErp'
import { useVendorOptions } from '../controllers/useProcurement'
import { useCustomers } from '../controllers/useSales'
import type { BillingKind } from '../models/billing'
import { todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { ErrorNote } from './components/Form'
import { SectionTabs } from './components/ListTools'
import { BillingKpis } from './billing/BillingKpis'
import { InvoicesSection } from './billing/InvoicesSection'
import { PayablesSection } from './billing/PayablesSection'
import { ReceivablesSection } from './billing/ReceivablesSection'
import type { Directory } from './billing/directory'

const OPTION_LIMIT = 100

const TABS: { value: BillingKind; label: string }[] = [
  { value: 'invoices', label: 'Faktur' },
  { value: 'receivables', label: 'Piutang' },
  { value: 'payables', label: 'Utang' },
]

export function BillingPage() {
  const [tab, setTab] = useState<BillingKind>('invoices')
  const [today] = useState(todayDate)
  const customers = useCustomers({ pageSize: OPTION_LIMIT })
  const vendors = useVendorOptions()
  const projects = useProjects({ pageSize: OPTION_LIMIT })

  const directory: Directory = {
    customers: customers.data?.items ?? [],
    vendors: vendors.data?.items ?? [],
    projects: projects.data?.items ?? [],
    isLoading: customers.isPending || vendors.isPending || projects.isPending,
    isError: customers.isError || vendors.isError || projects.isError,
  }

  return (
    <AppShell title="Tagihan" description="Faktur, piutang pelanggan, dan utang ke vendor beserta pembayarannya">
      <BillingKpis />

      <div className="mt-6 mb-4">
        <SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian tagihan" />
      </div>

      {directory.isError ? (
        <div className="mb-4">
          <ErrorNote message="Daftar pelanggan, vendor, atau proyek gagal dimuat, jadi sebagian nama tampil sebagai kode singkat." />
        </div>
      ) : null}

      {tab === 'invoices' ? <InvoicesSection directory={directory} today={today} /> : null}
      {tab === 'receivables' ? <ReceivablesSection directory={directory} today={today} /> : null}
      {tab === 'payables' ? <PayablesSection directory={directory} today={today} /> : null}
    </AppShell>
  )
}
