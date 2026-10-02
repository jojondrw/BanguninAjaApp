import { useState } from 'react'

import { useSavedLocations } from '../../controllers/useErp'
import {
  comparisonLocationIds,
  isSameSelection,
  MIN_COMPARED,
  toggleSelection,
  type ComparisonDetail,
} from '../../models/location'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Button } from '../components/Form'
import { COMPARISON_RESULT_ID, ComparisonResult, type OpenedComparison } from './ComparisonResult'
import { LocationPicker } from './LocationPicker'
import { SavedComparisons } from './SavedComparisons'

function toOpened(detail: ComparisonDetail): OpenedComparison {
  return { id: detail.id, name: detail.name, locationIds: comparisonLocationIds(detail) }
}

// Tab "Bandingkan" di /lokasi: pilih 2 sampai 4 lokasi tersimpan, lihat
// berdampingan, lalu simpan sebagai perbandingan bernama.
export function LocationCompare({ onEvaluate }: { onEvaluate: () => void }) {
  const allSites = useSavedLocations()
  const [projectId, setProjectId] = useState('')
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [opened, setOpened] = useState<OpenedComparison | null>(null)

  const isChanged = opened !== null && !isSameSelection(opened.locationIds, selectedIds)

  function handleOpened(detail: ComparisonDetail) {
    const next = toOpened(detail)
    setOpened(next)
    setSelectedIds(next.locationIds)
    // Lokasi perbandingan bisa berasal dari beberapa proyek, jadi saringan
    // proyek dilepas supaya semua centangnya terlihat.
    setProjectId('')
    document.getElementById(COMPARISON_RESULT_ID)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  function handleRenamed(detail: ComparisonDetail) {
    if (opened?.id === detail.id) {
      setOpened({ ...opened, name: detail.name })
    }
  }

  function handleDeleted(id: string) {
    if (opened?.id === id) {
      setOpened(null)
    }
  }

  if (allSites.isPending) {
    return (
      <Card title="Bandingkan lokasi">
        <Loading label="Mengambil lokasi tersimpan..." />
      </Card>
    )
  }

  if (allSites.isError) {
    return (
      <Card title="Bandingkan lokasi">
        <LoadFailed onRetry={() => allSites.refetch()} />
      </Card>
    )
  }

  if (allSites.data.totalItems < MIN_COMPARED) {
    return <NotEnoughSites count={allSites.data.totalItems} onEvaluate={onEvaluate} />
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-6 xl:grid-cols-3">
        <div className="min-w-0 xl:col-span-2">
          <LocationPicker
            projectId={projectId}
            onProjectChange={setProjectId}
            selectedIds={selectedIds}
            onToggle={(id) => setSelectedIds((current) => toggleSelection(current, id))}
            onClear={() => setSelectedIds([])}
          />
        </div>
        <div className="min-w-0">
          <SavedComparisons
            openedId={opened?.id ?? null}
            onOpened={handleOpened}
            onRenamed={handleRenamed}
            onDeleted={handleDeleted}
          />
        </div>
      </div>

      <ComparisonResult
        selectedIds={selectedIds}
        opened={opened}
        isChanged={isChanged}
        onRemove={(id) => setSelectedIds((current) => current.filter((selected) => selected !== id))}
        onDropMany={(ids) => setSelectedIds((current) => current.filter((selected) => !ids.includes(selected)))}
        onSaved={(detail) => setOpened(toOpened(detail))}
      />
    </div>
  )
}

function NotEnoughSites({ count, onEvaluate }: { count: number; onEvaluate: () => void }) {
  const status = count === 0 ? 'Belum ada lokasi tersimpan.' : 'Baru ada 1 lokasi tersimpan.'

  return (
    <Card title="Bandingkan lokasi" description={`Perbandingan butuh minimal ${MIN_COMPARED} lokasi tersimpan.`}>
      <Empty
        message={`${status} Buka tab Evaluasi, klik titik di peta, pilih profil bangunan, lalu jalankan evaluasi. Setiap hasil evaluasi otomatis tersimpan sebagai lokasi dan bisa dibandingkan di sini.`}
      />
      <div className="mt-4 flex justify-center">
        <Button onClick={onEvaluate}>Buka tab Evaluasi</Button>
      </div>
    </Card>
  )
}
