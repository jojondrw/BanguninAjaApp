import { useSearchParams } from 'react-router-dom'

import { useCashFlow } from '../controllers/useErp'
import { rupiahShort } from '../shared/format'
import { AppShell } from './components/AppShell'
import { Kpi, KpiRow } from './components/Data'
import { SectionTabs } from './components/ListTools'
import { BudgetsTab } from './finance/BudgetsTab'
import { CashTab } from './finance/CashTab'
import { FinanceSummaryTab } from './finance/FinanceSummaryTab'
import {
  DEFAULT_FINANCE_TAB,
  FINANCE_TABS,
  TAB_PARAM,
  toFinanceTab,
  type FinanceTab,
} from './finance/financeTabs'
import { JournalTab } from './finance/JournalTab'
import { LedgerTab } from './finance/LedgerTab'

export function FinancialPage() {
  const cashFlow = useCashFlow()
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = toFinanceTab(searchParams.get(TAB_PARAM))

  // Tab disimpan di URL supaya muat ulang dan tautan yang dibagikan membuka tab
  // yang sama. replace dipakai agar tombol kembali tidak menelusuri tiap tab.
  function changeTab(next: FinanceTab) {
    setSearchParams(next === DEFAULT_FINANCE_TAB ? {} : { [TAB_PARAM]: next }, { replace: true })
  }

  return (
    <AppShell title="Keuangan" description="Anggaran, kas, jurnal umum, dan buku besar">
      <KpiRow>
        <Kpi label="Kas masuk" value={cashFlow.isPending ? '...' : rupiahShort(cashFlow.data?.totalIn ?? 0)} />
        <Kpi label="Kas keluar" value={cashFlow.isPending ? '...' : rupiahShort(cashFlow.data?.totalOut ?? 0)} />
        <Kpi
          label="Selisih"
          value={cashFlow.isPending ? '...' : rupiahShort(cashFlow.data?.net ?? 0)}
          note="masuk dikurangi keluar"
        />
        <Kpi label="Saldo" value={cashFlow.isPending ? '...' : rupiahShort(cashFlow.data?.balance ?? 0)} />
      </KpiRow>

      <div className="mt-6 mb-4">
        <SectionTabs tabs={FINANCE_TABS} active={tab} onChange={changeTab} label="Bagian keuangan" />
      </div>

      {tab === 'ringkasan' ? <FinanceSummaryTab /> : null}
      {tab === 'anggaran' ? <BudgetsTab /> : null}
      {tab === 'kas' ? <CashTab /> : null}
      {tab === 'jurnal' ? <JournalTab /> : null}
      {tab === 'buku-besar' ? <LedgerTab /> : null}
    </AppShell>
  )
}
