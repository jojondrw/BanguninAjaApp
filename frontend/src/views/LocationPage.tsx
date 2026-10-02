import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'

import { AppShell } from './components/AppShell'
import { SiteEvaluation } from './components/SiteEvaluation'
import type { MapPoint } from './components/SiteMap'
import { Tabs, type TabItem } from './components/Tabs'
import { LocationCompare } from './location/LocationCompare'

type LocationTab = 'evaluasi' | 'bandingkan'

const TABS: readonly TabItem<LocationTab>[] = [
  { id: 'evaluasi', label: 'Evaluasi' },
  { id: 'bandingkan', label: 'Bandingkan' },
]

const DEFAULT_TAB: LocationTab = 'evaluasi'
const TAB_PARAM = 'tab'

function toTab(value: string | null): LocationTab {
  return TABS.find((tab) => tab.id === value)?.id ?? DEFAULT_TAB
}

export function LocationPage() {
  const [focus, setFocus] = useState<MapPoint | null>(null)
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = toTab(searchParams.get(TAB_PARAM))

  // Sama dengan ruang kerja proyek: tab disimpan di URL supaya muat ulang dan
  // tautan yang dibagikan membuka tab yang sama.
  function changeTab(next: LocationTab) {
    setSearchParams(next === DEFAULT_TAB ? {} : { [TAB_PARAM]: next }, { replace: true })
  }

  return (
    <AppShell
      title="Analisis Lokasi"
      description="Evaluasi kelayakan lokasi, lalu bandingkan kandidat sebelum memilih"
    >
      <Tabs idPrefix="lokasi" label="Bagian analisis lokasi" tabs={TABS} active={tab} onChange={changeTab}>
        {tab === 'evaluasi' ? <SiteEvaluation focus={focus} onFocus={setFocus} /> : null}
        {tab === 'bandingkan' ? <LocationCompare onEvaluate={() => changeTab('evaluasi')} /> : null}
      </Tabs>
    </AppShell>
  )
}
