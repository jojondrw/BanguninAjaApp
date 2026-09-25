import { AppShell } from './components/AppShell'
import { Card, Empty } from './components/Data'

export function LocationPage() {
  return (
    <AppShell
      title="Analisis Lokasi"
      description="Peta dan skor kelayakan lokasi, dikerjakan di vertikal terpisah"
    >
      <Card title="Peta lokasi" description="Menunggu endpoint penilaian siap">
        <Empty message="Halaman ini disiapkan oleh pemilik vertikal Site Intelligence. Rangka navigasi dan sesi login sudah siap menampungnya." />
      </Card>
    </AppShell>
  )
}
