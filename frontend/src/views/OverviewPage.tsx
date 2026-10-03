import { ChevronRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useBudgetSummary, useCashFlow, useProjects, useProjectSummary } from '../controllers/useErp'
import { number, rupiahShort, shortDate } from '../shared/format'
import { kpiText } from '../shared/kpiText'
import { todayDate } from '../shared/localDate'
import { PROJECT_STATUS_LABEL, PROJECT_STATUS_TONE, type Project } from '../models/project'
import { AppShell } from './components/AppShell'
import { CHIP_CLASS } from './components/ListTools'
import { CashFlowChart } from './components/CashFlowChart'
import { Bar, Card, Empty, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'
import { AttentionList } from './overview/AttentionList'

const CHART_MONTHS = 12
const RECENT_PROJECTS = 8

export function OverviewPage() {
  const summary = useProjectSummary()
  const projects = useProjects({ pageSize: RECENT_PROJECTS })
  const budgets = useBudgetSummary()
  const cashFlow = useCashFlow()
  const yearFlow = useCashFlow({ dateFrom: monthsAgo(CHART_MONTHS - 1), dateTo: todayDate() })

  const items = projects.data?.items ?? []

  const budgetTotal = budgets.data?.budget ?? 0
  const realizedTotal = budgets.data?.realized ?? 0
  const absorption = budgetTotal > 0 ? Math.round((realizedTotal / budgetTotal) * 100) : 0

  return (
    <AppShell title="Ringkasan" description={`Kondisi seluruh proyek per ${shortDate(todayDate())}`}>
      <KpiRow>
        <Kpi
          label="Proyek berjalan"
          value={kpiText(summary, (data) => number(data.byStatus.ongoing))}
          note={summary.data ? `dari ${number(summary.data.count)} proyek terdaftar` : undefined}
        />
        <Kpi
          label="Nilai kontrak"
          value={kpiText(summary, (data) => rupiahShort(data.contractValue))}
          note={summary.data ? `seluruh proyek, rata-rata progres ${summary.data.averageProgress}%` : undefined}
        />
        <Kpi
          label="Saldo kas"
          value={cashFlow.isPending ? '...' : rupiahShort(cashFlow.data?.balance ?? 0)}
          note={cashFlow.data ? `masuk ${rupiahShort(cashFlow.data.totalIn)}, keluar ${rupiahShort(cashFlow.data.totalOut)}` : undefined}
        />
        <Kpi
          label="Serapan anggaran"
          value={serapan(budgets.isPending, budgets.isError, absorption)}
          note={serapanNote(budgets.isError, budgetTotal, realizedTotal)}
        />
      </KpiRow>

      <div className="mt-6 grid gap-6 xl:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]">
        <Card
          title="Arus kas 12 bulan"
          description="Kas masuk di atas garis nol, kas keluar di bawahnya"
          action={<CardLink to="/keuangan" label="Buka keuangan" />}
        >
          {yearFlow.isPending ? <Loading /> : null}
          {yearFlow.isError ? <LoadFailed onRetry={() => yearFlow.refetch()} /> : null}
          {yearFlow.data ? (
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

      <div className="mt-6">
        <Card
          title="Proyek"
          description="Status, target, dan progres tiap proyek"
          action={<CardLink to="/proyek" label="Semua proyek" />}
        >
          {projects.isPending ? <Loading /> : null}
          {projects.isError ? <LoadFailed onRetry={() => projects.refetch()} /> : null}
          {projects.data ? (
            <Table
              rows={items}
              emptyMessage="Belum ada proyek. Buat proyek pertama dari halaman Proyek."
              columns={[
                { header: 'Kode', cell: (row: Project) => row.code },
                {
                  header: 'Nama',
                  cell: (row: Project) => (
                    <Link
                      to={`/proyek/${row.id}`}
                      className="font-medium text-slate-900 underline-offset-4 hover:text-navy-600 hover:underline"
                    >
                      {row.name}
                    </Link>
                  ),
                },
                { header: 'Status', cell: (row: Project) => <StatusChip project={row} /> },
                { header: 'Target selesai', cell: (row: Project) => shortDate(row.targetEndDate) },
                {
                  header: 'Nilai',
                  align: 'right',
                  cell: (row: Project) => rupiahShort(row.contractValue),
                },
                { header: 'Progres', align: 'right', cell: (row: Project) => <Bar percent={row.progress} /> },
              ]}
            />
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

function monthsAgo(count: number): string {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth() - count, 1)
  return `${start.getFullYear()}-${String(start.getMonth() + 1).padStart(2, '0')}-01`
}

function serapan(isPending: boolean, isError: boolean, absorption: number): string {
  if (isError) {
    return '—'
  }
  return isPending ? '...' : `${absorption}%`
}

function serapanNote(isError: boolean, budgetTotal: number, realizedTotal: number): string {
  if (isError) {
    return 'Data anggaran gagal dimuat'
  }
  if (budgetTotal === 0) {
    return 'anggaran belum diisi'
  }
  return `${rupiahShort(realizedTotal)} dari ${rupiahShort(budgetTotal)}`
}

export function StatusChip({ project }: { project: Project }) {
  return (
    <span className={`${CHIP_CLASS} ${PROJECT_STATUS_TONE[project.status]}`}>
      {PROJECT_STATUS_LABEL[project.status]}
    </span>
  )
}
