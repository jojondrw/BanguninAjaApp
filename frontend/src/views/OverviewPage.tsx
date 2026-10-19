import { ChevronRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useCan } from '../controllers/useAccess'
import { useCashFlow, useProjects } from '../controllers/useErp'
import { shortDate } from '../shared/format'
import { monthsAgo, todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { CashFlowChart } from './components/CashFlowChart'
import { Card, Empty, LoadFailed, Loading } from './components/Data'
import { AttentionList } from './overview/AttentionList'
import { BusinessFlow } from './overview/BusinessFlow'
import { BusinessHealth } from './overview/BusinessHealth'
import { InvestorMetrics } from './overview/InvestorMetrics'
import { ProjectCards } from './overview/ProjectCards'
import { SalesFunnel } from './overview/SalesFunnel'
import { WelcomeBanner } from './overview/WelcomeBanner'

const CHART_MONTHS = 12
const RECENT_PROJECTS = 8

export function OverviewPage() {
  // Data modul yang tidak boleh dibaca peran ini tidak diminta sama sekali,
  // supaya tidak ada permintaan 403 di latar belakang.
  const canReadProjects = useCan('projects', 'read')
  const canReadFinance = useCan('finance', 'read')
  const canReadSales = useCan('sales', 'read')
  const cashFrom = monthsAgo(CHART_MONTHS - 1)
  const cashTo = todayDate()
  const projects = useProjects({ pageSize: RECENT_PROJECTS }, { enabled: canReadProjects })
  const yearFlow = useCashFlow({ dateFrom: cashFrom, dateTo: cashTo }, { enabled: canReadFinance })

  return (
    <AppShell title="Ringkasan" description={`Kondisi bisnis per ${shortDate(cashTo)}`}>
      <WelcomeBanner />

      <div className="mt-8">
        <InvestorMetrics cashFrom={cashFrom} cashTo={cashTo} />
      </div>

      <div className="mt-8 grid gap-8 lg:grid-cols-2">
        <Card title="Kesehatan bisnis" description="Empat rasio yang paling cepat menunjukkan bisnis berjalan baik">
          <BusinessHealth />
        </Card>
        <Card
          title="Funnel penjualan unit"
          description="Dari unit yang dipasarkan sampai kontrak lunas"
          action={canReadSales ? <CardLink to="/penjualan" label="Buka penjualan" /> : undefined}
        >
          {canReadSales ? <SalesFunnel /> : <Empty message="Funnel hanya tampil untuk peran yang punya akses Penjualan." />}
        </Card>
      </div>

      <section aria-labelledby="business-flow" className="mt-8 space-y-3">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <h2 id="business-flow" className="text-[15px] font-semibold text-slate-900">
            Alur bisnis
          </h2>
          <p className="text-xs text-slate-500">Dari lahan yang dinilai sampai unit yang terjual</p>
        </div>
        <BusinessFlow />
      </section>

      <div className="mt-8 grid gap-8 xl:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]">
        <Card
          title="Arus kas 12 bulan"
          description="Kas masuk di atas garis nol, kas keluar di bawahnya"
          action={canReadFinance ? <CardLink to="/keuangan" label="Buka keuangan" /> : undefined}
        >
          {canReadFinance ? null : <Empty message="Arus kas hanya tampil untuk peran yang punya akses Keuangan." />}
          {canReadFinance && yearFlow.isPending ? <Loading /> : null}
          {canReadFinance && yearFlow.isError ? <LoadFailed onRetry={() => yearFlow.refetch()} /> : null}
          {canReadFinance && yearFlow.data ? (
            yearFlow.data.periods.every((period) => period.cashIn === 0 && period.cashOut === 0) ? (
              <Empty message="Belum ada transaksi kas dalam 12 bulan terakhir." />
            ) : (
              <CashFlowChart periods={yearFlow.data.periods} />
            )
          ) : null}
        </Card>

        <Card title="Perlu perhatian" description="Hal yang menunggu tindakan di semua modul">
          <AttentionList />
        </Card>
      </div>

      <div className="mt-8">
        <Card
          title="Proyek"
          description="Progres, status, dan nilai kontrak tiap proyek"
          action={canReadProjects ? <CardLink to="/proyek" label="Semua proyek" /> : undefined}
        >
          {canReadProjects ? null : <Empty message="Daftar proyek hanya tampil untuk peran yang punya akses Proyek." />}
          {canReadProjects && projects.isPending ? <Loading /> : null}
          {projects.isError ? <LoadFailed onRetry={() => projects.refetch()} /> : null}
          {projects.data ? (
            projects.data.items.length === 0 ? (
              <Empty message="Belum ada proyek. Buat proyek pertama dari halaman Proyek." />
            ) : (
              <ProjectCards projects={projects.data.items} />
            )
          ) : null}
        </Card>
      </div>
    </AppShell>
  )
}

function CardLink({ to, label }: { to: string; label: string }) {
  return (
    <Link
      to={to}
      className="inline-flex items-center gap-0.5 rounded-md px-1.5 py-1 text-[13px] font-medium text-navy-600 transition-colors
                 hover:bg-navy-50"
    >
      {label}
      <ChevronRight aria-hidden="true" className="size-3.5" strokeWidth={2} />
    </Link>
  )
}
