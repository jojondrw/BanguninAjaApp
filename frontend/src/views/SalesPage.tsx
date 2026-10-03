import { useState } from 'react'

import { useContractSummary, useInstallmentCount, useUnitSummary } from '../controllers/useSales'
import type { UnitStatus } from '../models/sales'
import { number, rupiahShort } from '../shared/format'
import { kpiText } from '../shared/kpiText'
import { AppShell } from './components/AppShell'
import { Kpi, KpiRow } from './components/Data'
import { SectionTabs } from './components/ListTools'
import { ContractsSection } from './sales/ContractsSection'
import { CustomersSection } from './sales/CustomersSection'
import { LeadsSection } from './sales/LeadsSection'
import { UnitsSection } from './sales/UnitsSection'

type SalesTab = 'units' | 'leads' | 'customers' | 'contracts'

// Urutannya mengikuti alur jual: stok unit, calon pembeli, pelanggan, kontrak.
const TABS: { value: SalesTab; label: string }[] = [
  { value: 'units', label: 'Unit' },
  { value: 'leads', label: 'Prospek' },
  { value: 'customers', label: 'Pelanggan' },
  { value: 'contracts', label: 'Kontrak' },
]

function SalesKpis() {
  const summary = useUnitSummary()
  const contracts = useContractSummary()
  const overdue = useInstallmentCount('overdue')
  const due = useInstallmentCount('due')

  const countOf = (status: UnitStatus) =>
    summary.data?.find((item) => item.status === status)?.total ?? 0
  const totalUnits = summary.data?.reduce((total, item) => total + item.total, 0) ?? 0

  return (
    <KpiRow>
      <Kpi
        label="Unit tersedia"
        value={kpiText(summary, () => number(countOf('available')))}
        note={summary.data ? `dari ${number(totalUnits)} unit, ${number(countOf('on_hold'))} ditahan` : undefined}
      />
      <Kpi
        label="Unit terjual"
        value={kpiText(summary, () => number(countOf('sold')))}
        note={summary.data ? `${number(countOf('reserved'))} unit sedang dipesan` : undefined}
      />
      <Kpi
        label="Nilai kontrak"
        value={kpiText(contracts, (summary) => rupiahShort(summary.totalValue))}
        note={
          contracts.data
            ? `${number(contracts.data.count - contracts.data.byStatus.cancelled)} kontrak di luar yang batal`
            : undefined
        }
      />
      <Kpi
        label="Cicilan terlambat"
        value={kpiText(overdue, number)}
        note={due.data === undefined ? undefined : `${number(due.data)} jatuh tempo dalam 7 hari`}
      />
    </KpiRow>
  )
}

export function SalesPage() {
  const [tab, setTab] = useState<SalesTab>('units')

  return (
    <AppShell title="Penjualan" description="Unit properti, prospek, pelanggan, kontrak, dan jadwal cicilannya">
      <SalesKpis />

      <div className="mt-6 mb-4">
        <SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian penjualan" />
      </div>

      {tab === 'units' ? <UnitsSection /> : null}
      {tab === 'leads' ? <LeadsSection /> : null}
      {tab === 'customers' ? <CustomersSection /> : null}
      {tab === 'contracts' ? <ContractsSection /> : null}
    </AppShell>
  )
}
