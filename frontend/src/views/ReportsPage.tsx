import { AppShell } from './components/AppShell'
import { ReportBuilder } from './reports/ReportBuilder'
import { ReportHistory } from './reports/ReportHistory'

export function ReportsPage() {
  return (
    <AppShell
      title="Laporan"
      description="Unduh laporan keuangan, proyek, piutang, stok, penggajian, dan penjualan sebagai PDF, Excel, atau CSV"
    >
      <div className="grid gap-6">
        <ReportBuilder />
        <ReportHistory />
      </div>
    </AppShell>
  )
}
