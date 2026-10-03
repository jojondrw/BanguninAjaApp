import { useState } from 'react'

import type { BillingKind, Invoice } from '../models/billing'
import { todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { SectionTabs } from './components/ListTools'
import { BillingKpis } from './billing/BillingKpis'
import { InvoicesSection } from './billing/InvoicesSection'
import { PayablesSection } from './billing/PayablesSection'
import { ReceivablesSection } from './billing/ReceivablesSection'

const TABS: { value: BillingKind; label: string }[] = [
  { value: 'invoices', label: 'Faktur' },
  { value: 'receivables', label: 'Piutang' },
  { value: 'payables', label: 'Utang' },
]

export function BillingPage() {
  const [tab, setTab] = useState<BillingKind>('invoices')
  const [draftInvoice, setDraftInvoice] = useState<Invoice | null>(null)
  const [today] = useState(todayDate)

  const changeTab = (next: BillingKind) => {
    setDraftInvoice(null)
    setTab(next)
  }

  // "Catat sebagai piutang" di rincian faktur pindah ke tab Piutang dengan
  // formulir yang sudah terisi dari faktur itu.
  const recordReceivable = (invoice: Invoice) => {
    setDraftInvoice(invoice)
    setTab('receivables')
  }

  return (
    <AppShell title="Tagihan" description="Faktur, piutang pelanggan, dan utang ke vendor beserta pembayarannya">
      <BillingKpis />

      <div className="mt-6 mb-4">
        <SectionTabs tabs={TABS} active={tab} onChange={changeTab} label="Bagian tagihan" />
      </div>

      {tab === 'invoices' ? <InvoicesSection today={today} onRecordReceivable={recordReceivable} /> : null}
      {tab === 'receivables' ? <ReceivablesSection today={today} draftInvoice={draftInvoice} /> : null}
      {tab === 'payables' ? <PayablesSection today={today} /> : null}
    </AppShell>
  )
}
