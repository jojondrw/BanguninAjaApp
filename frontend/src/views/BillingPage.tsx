import { useMemo, useState } from 'react'

import type { BillingKind, BillingStatus, Invoice } from '../models/billing'
import { todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { SectionTabs } from './components/ListTools'
import { BillingKpis } from './billing/BillingKpis'
import { BillingOverview, type BillingDrill } from './billing/BillingOverview'
import { InvoicesSection } from './billing/InvoicesSection'
import { PayablesSection } from './billing/PayablesSection'
import { ReceivablesSection } from './billing/ReceivablesSection'
import { ReconciliationBanner } from './billing/ReconciliationBanner'

type BillingTab = 'overview' | BillingKind

const TABS: { value: BillingTab; label: string }[] = [
  { value: 'overview', label: 'Ringkasan' },
  { value: 'invoices', label: 'Faktur' },
  { value: 'receivables', label: 'Piutang' },
  { value: 'payables', label: 'Utang' },
]

export function BillingPage() {
  const [tab, setTab] = useState<BillingTab>('overview')
  const [draftInvoice, setDraftInvoice] = useState<Invoice | null>(null)
  const [status, setStatus] = useState<BillingStatus | ''>('')
  const [visit, setVisit] = useState(0)
  const [today] = useState(todayDate)

  const changeTab = (next: BillingTab) => {
    setDraftInvoice(null)
    setStatus('')
    setTab(next)
  }

  const drill = useMemo<BillingDrill>(() => {
    const open = (next: BillingKind) => (filter: BillingStatus | '') => {
      setDraftInvoice(null)
      setStatus(filter)
      setVisit((current) => current + 1)
      setTab(next)
    }
    return { openReceivables: open('receivables'), openPayables: open('payables') }
  }, [])

  // "Catat sebagai piutang" di rincian faktur pindah ke tab Piutang dengan
  // formulir yang sudah terisi dari faktur itu.
  const recordReceivable = (invoice: Invoice) => {
    setDraftInvoice(invoice)
    setTab('receivables')
  }

  return (
    <AppShell
      title="Tagihan"
      description="Faktur, piutang pelanggan, dan utang ke vendor beserta pembayarannya"
      tabs={<SectionTabs tabs={TABS} active={tab} onChange={changeTab} label="Bagian tagihan" />}
    >
      <BillingKpis />
      <ReconciliationBanner />

      <div className="mb-6" />

      {tab === 'overview' ? <BillingOverview drill={drill} /> : null}
      {tab === 'invoices' ? <InvoicesSection today={today} onRecordReceivable={recordReceivable} /> : null}
      {tab === 'receivables' ? (
        <ReceivablesSection key={`piutang-${visit}`} today={today} draftInvoice={draftInvoice} initialStatus={status} />
      ) : null}
      {tab === 'payables' ? <PayablesSection key={`utang-${visit}`} today={today} initialStatus={status} /> : null}
    </AppShell>
  )
}
