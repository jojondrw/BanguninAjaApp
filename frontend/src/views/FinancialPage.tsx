import { useBudgets, useCashFlow, useCashTransactions, useProjects } from '../controllers/useErp'
import { monthLabel, rupiah, rupiahShort, shortDate } from '../shared/format'
import { AppShell } from './components/AppShell'
import { Bar, Card, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'

export function FinancialPage() {
  const cashFlow = useCashFlow()
  const budgets = useBudgets()
  const transactions = useCashTransactions(8)
  const projects = useProjects({ pageSize: 100 })

  const projectName = (id: string) =>
    projects.data?.items.find((project) => project.id === id)?.name ?? id.slice(0, 8)

  return (
    <AppShell title="Keuangan" description="Anggaran, arus kas, dan transaksi terbaru">
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

      <div className="mt-6 grid gap-6">
        <Card title="Anggaran per proyek" description="Realisasi dihitung server dari transaksi kas">
          {budgets.isPending ? <Loading /> : null}
          {budgets.isError ? <LoadFailed onRetry={() => budgets.refetch()} /> : null}
          {budgets.data ? (
            <Table
              rows={budgets.data.items}
              emptyMessage="Belum ada anggaran. Tambahkan lewat API POST /api/finance/budgets."
              columns={[
                { header: 'Proyek', cell: (row) => projectName(row.projectId) },
                { header: 'Tahun', cell: (row) => String(row.year) },
                { header: 'Anggaran', align: 'right', cell: (row) => rupiahShort(row.value) },
                { header: 'Realisasi', align: 'right', cell: (row) => rupiahShort(row.realized) },
                { header: 'Sisa', align: 'right', cell: (row) => rupiahShort(row.remaining) },
                { header: 'Serapan', align: 'right', cell: (row) => <Bar percent={row.absorption} /> },
              ]}
            />
          ) : null}
        </Card>

        <div className="grid gap-6 xl:grid-cols-2">
          <Card title="Arus kas per bulan">
            {cashFlow.isPending ? <Loading /> : null}
            {cashFlow.isError ? <LoadFailed onRetry={() => cashFlow.refetch()} /> : null}
            {cashFlow.data ? (
              <Table
                rows={cashFlow.data.periods}
                emptyMessage="Belum ada transaksi kas yang tercatat."
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
            ) : null}
          </Card>

          <Card title="Transaksi terbaru">
            {transactions.isPending ? <Loading /> : null}
            {transactions.isError ? <LoadFailed onRetry={() => transactions.refetch()} /> : null}
            {transactions.data ? (
              <Table
                rows={transactions.data.items}
                emptyMessage="Belum ada transaksi kas."
                columns={[
                  { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                  { header: 'Keterangan', cell: (row) => row.note || '-' },
                  {
                    header: 'Jenis',
                    cell: (row) => (
                      <span
                        className={`rounded px-2 py-0.5 text-xs font-medium ${
                          row.type === 'in' ? 'bg-green-100 text-green-800' : 'bg-amber-100 text-amber-800'
                        }`}
                      >
                        {row.type === 'in' ? 'Masuk' : 'Keluar'}
                      </span>
                    ),
                  },
                  { header: 'Jumlah', align: 'right', cell: (row) => rupiah(row.amount) },
                ]}
              />
            ) : null}
          </Card>
        </div>
      </div>
    </AppShell>
  )
}
