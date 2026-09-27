import { useState } from 'react'

import type { LatLng } from '../models/site'
import { AppShell } from './components/AppShell'
import { Card } from './components/Data'
import { SiteMap } from './components/SiteMap'

export function LocationPage() {
  const [picked, setPicked] = useState<LatLng | null>(null)

  return (
    <AppShell
      title="Analisis Lokasi"
      description="Klik di peta untuk memilih titik lokasi calon proyek."
    >
      <div className="grid gap-6 xl:grid-cols-[1.6fr_1fr]">
        <Card title="Peta lokasi" description="Klik peta untuk menempatkan penanda">
          <SiteMap onPick={setPicked} selected={picked} />
        </Card>

        <Card title="Titik terpilih" description="Koordinat titik yang Anda klik di peta">
          {picked ? (
            <dl className="grid grid-cols-2 gap-4">
              <div>
                <dt className="text-xs text-slate-500">Latitude</dt>
                <dd className="mt-1 text-lg font-semibold tabular-nums text-slate-900">
                  {picked.latitude.toFixed(6)}
                </dd>
              </div>
              <div>
                <dt className="text-xs text-slate-500">Longitude</dt>
                <dd className="mt-1 text-lg font-semibold tabular-nums text-slate-900">
                  {picked.longitude.toFixed(6)}
                </dd>
              </div>
            </dl>
          ) : (
            <p className="py-6 text-sm text-slate-500">
              Belum ada titik yang dipilih. Klik di mana saja pada peta untuk memilih lokasi.
            </p>
          )}
          <p className="mt-4 text-xs text-slate-400">
            Penilaian kelayakan lokasi akan tampil di sini setelah peta terhubung ke layanan
            penilaian (dikerjakan pada tugas berikutnya).
          </p>
        </Card>
      </div>
    </AppShell>
  )
}
