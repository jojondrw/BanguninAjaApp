import { useBudgets, useCashFlow, useProjects } from '../controllers/useErp'
import { monthLabel, rupiahShort, shortDate } from '../shared/format'
import { PROJECT_STATUS_LABEL, PROJECT_STATUS_TONE, type Project } from '../models/project'
import { AppShell } from './components/AppShell'
import { Bar, Card, Empty, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'

export function OverviewPage() {
  const projects = useProjects({ pageSize: 100 })
  const budgets = useBudgets()
  const cashFlow = useCashFlow()

  const items = projects.data?.items ?? []
  const ongoing = items.filter((project) => project.status === 'ongoing')
  const contractValue = items.reduce((total, project) => total + project.contractValue, 0)

  const budgetItems = budgets.data?.items ?? []
  const budgetTotal = budgetItems.reduce((total, budget) => total + budget.value, 0)
  const realizedTotal = budgetItems.reduce((total, budget) => total + budget.realized, 0)
  const absorption = budgetTotal > 0 ? Math.round((realizedTotal / budgetTotal) * 100) : 0

  return (
    <AppShell
      title="Ringkasan"
      description="Angka di halaman ini dihitung langsung dari database, bukan contoh."
    >
      <KpiRow>
        <Kpi
          label="Proyek berjalan"
          value={projects.isPending ? '...' : String(ongoing.length)}
          note={`dari ${items.length} proyek terdaftar`}
        />
        <Kpi
          label="Nilai kontrak"
          value={projects.isPending ? '...' : rupiahShort(contractValue)}
          note="seluruh proyek"
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

      <div className="mt-6 grid gap-6 xl:grid-cols-[1.4fr_1fr]">
        <Card title="Proyek" description="Diurutkan sesuai urutan dari server">
          {projects.isPending ? <Loading /> : null}
          {projects.isError ? <LoadFailed onRetry={() => projects.refetch()} /> : null}
          {projects.data ? (
            <Table
              rows={items.slice(0, 8)}
              emptyMessage="Belum ada proyek. Tambahkan lewat API POST /api/projects."
              columns={[
                { header: 'Kode', cell: (row: Project) => row.code },
                { header: 'Nama', cell: (row: Project) => row.name },
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

        <Card title="Arus kas" description="Dikelompokkan per bulan oleh server">
          {cashFlow.isPending ? <Loading /> : null}
          {cashFlow.isError ? <LoadFailed onRetry={() => cashFlow.refetch()} /> : null}
          {cashFlow.data ? (
            cashFlow.data.periods.length === 0 ? (
              <Empty message="Belum ada transaksi kas yang tercatat." />
            ) : (
              <Table
                rows={cashFlow.data.periods.slice(-6)}
                emptyMessage="Belum ada transaksi kas."
                columns={[
                  { header: 'Periode', cell: (row) => monthLabel(row.period) },
                  { header: 'Masuk', align: 'right', cell: (row) => rupiahShort(row.cashIn) },
                  { header: 'Keluar', align: 'right', cell: (row) => rupiahShort(row.cashOut) },
                  {
                    header: 'Selisih',
                    align: 'right',
                    cell: (row) => (
                      <span className={row.net < 0 ? 'text-red-600' : 'text-green-700'}>
                        {rupiahShort(row.net)}
                      </span>
                    ),
                  },
                ]}
              />
            )
          ) : null}
        </Card>
      </div>
    </AppShell>
  )
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
    <span className={`rounded px-2 py-0.5 text-xs font-medium ${PROJECT_STATUS_TONE[project.status]}`}>
      {PROJECT_STATUS_LABEL[project.status]}
    </span>
  )
}
