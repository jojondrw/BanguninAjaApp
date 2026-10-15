import { useReadableReportTypes } from '../controllers/useReports'
import { AppShell } from './components/AppShell'
import { Card, Empty } from './components/Data'
import { ReportBuilder } from './reports/ReportBuilder'
import { ReportHistory } from './reports/ReportHistory'

export function ReportsPage() {
  const types = useReadableReportTypes()

  return (
    <AppShell
      title="Laporan"
      description="Unduh laporan sebagai PDF, Excel, atau CSV. Jenis laporan mengikuti modul yang boleh dibuka peran Anda."
    >
      {types.length === 0 ? (
        <Card title="Belum ada laporan untuk peran Anda">
          <Empty message="Setiap laporan memakai data satu modul, misalnya Keuangan, Tagihan, atau SDM. Peran Anda belum bisa membuka modul mana pun yang dipakai laporan. Minta admin mengubah peran Anda di halaman Pengguna." />
        </Card>
      ) : (
        <div className="grid gap-6">
          {/* Dipasang ulang kalau hak akses berubah, supaya pilihan jenis tidak tertinggal. */}
          <ReportBuilder key={types.map((type) => type.value).join()} types={types} />
          <ReportHistory types={types} />
        </div>
      )}
    </AppShell>
  )
}
