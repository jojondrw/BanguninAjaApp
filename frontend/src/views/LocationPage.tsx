import { useState } from 'react'

import { AppShell } from './components/AppShell'
import { SiteEvaluation } from './components/SiteEvaluation'
import type { MapPoint } from './components/SiteMap'

export function LocationPage() {
  const [focus, setFocus] = useState<MapPoint | null>(null)

  return (
    <AppShell
      title="Analisis Lokasi"
      description="Evaluasi kelayakan lokasi dan konteks regulasi serta berita sekitar"
    >
      <SiteEvaluation focus={focus} onFocus={setFocus} />
    </AppShell>
  )
}
