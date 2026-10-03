import { type FormEvent, useState } from 'react'

import { useBuildingProfiles } from '../../controllers/useErp'
import {
  useCreateComparison,
  useDimensions,
  useProjectNames,
  useRegionNames,
  useSavedLocationDetails,
  useUpdateComparison,
} from '../../controllers/useLocations'
import {
  COMPARISON_NAME_MAX,
  compareLocations,
  MIN_COMPARED,
  PHYSICAL_RISK_BELOW,
  type ComparisonDetail,
} from '../../models/location'
import { errorMessage } from '../../shared/errorMessage'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { ROW_BUTTON } from './buttons'
import { ComparisonGrid, type LocationNames } from './ComparisonGrid'

export const COMPARISON_RESULT_ID = 'hasil-perbandingan'

export interface OpenedComparison {
  id: string
  name: string
  locationIds: string[]
}

interface ComparisonResultProps {
  selectedIds: string[]
  opened: OpenedComparison | null
  isChanged: boolean
  onRemove: (id: string) => void
  onDropMany: (ids: string[]) => void
  onSaved: (detail: ComparisonDetail) => void
}

function pickMoreMessage(count: number): string {
  if (count === 0) {
    return `Centang minimal ${MIN_COMPARED} lokasi di atas untuk mulai membandingkan.`
  }
  return `Sudah 1 lokasi dipilih. Centang ${MIN_COMPARED - count} lokasi lagi untuk mulai membandingkan.`
}

export function ComparisonResult({ selectedIds, opened, isChanged, onRemove, onDropMany, onSaved }: ComparisonResultProps) {
  const details = useSavedLocationDetails(selectedIds)
  const dimensions = useDimensions()
  const projectNames = useProjectNames(details.locations.flatMap((location) => (location.projectId ? [location.projectId] : [])))
  const profiles = useBuildingProfiles()
  const regionNames = useRegionNames(details.locations.flatMap((location) => (location.regionId ? [location.regionId] : [])))

  const hasEnough = selectedIds.length >= MIN_COMPARED
  const table =
    hasEnough && details.locations.length >= MIN_COMPARED && dimensions.data
      ? compareLocations(details.locations, dimensions.data)
      : null
  const isLoading = hasEnough && table === null && (details.isPending || dimensions.isPending)

  const names: LocationNames = {
    project: (id) =>
      id ? (projectNames.get(id) ?? 'Memuat proyek...') : 'Tanpa proyek',
    profile: (id) => profiles.data?.find((profile) => profile.id === id)?.name ?? '-',
    region: (id) => (id ? (regionNames.get(id) ?? 'Memuat wilayah...') : '-'),
  }

  return (
    <div id={COMPARISON_RESULT_ID} className="scroll-mt-6">
      <Card
        title="Hasil perbandingan"
        description="Skor berskala 0 sampai 100: makin tinggi makin baik. Nilai paling unggul di tiap baris ditandai terbaik."
      >
        {opened ? (
          <div className="mb-4 flex flex-wrap items-center gap-x-3 gap-y-1 rounded-xl bg-navy-50 px-4 py-2.5 text-[13px] text-navy-900">
            <span>
              Dibuka dari perbandingan tersimpan <span className="font-semibold">{opened.name}</span>.
            </span>
            {isChanged ? <span className="text-navy-700">Pilihan lokasi sudah diubah dan belum disimpan.</span> : null}
          </div>
        ) : null}

        {hasEnough ? null : <Empty message={pickMoreMessage(selectedIds.length)} />}

        {details.error ? (
          <div className="mb-4 space-y-3">
            <ErrorNote message={`Sebagian lokasi gagal dimuat: ${errorMessage(details.error)}`} />
            <div className="flex flex-wrap gap-2">
              <button type="button" onClick={details.retry} className={ROW_BUTTON}>
                Coba lagi
              </button>
              <button type="button" onClick={() => onDropMany(details.failedIds)} className={ROW_BUTTON}>
                Keluarkan lokasi yang gagal dimuat
              </button>
            </div>
          </div>
        ) : null}

        {dimensions.isError ? <LoadFailed onRetry={() => dimensions.refetch()} /> : null}
        {isLoading ? <Loading label="Mengambil rincian lokasi..." /> : null}

        {table ? (
          <>
            <p className="mb-4 rounded-xl bg-slate-50 px-4 py-3 text-sm leading-6 text-slate-800">{table.summary}</p>

            {details.isPending ? (
              <p role="status" className="mb-3 text-xs text-slate-500">
                Mengambil rincian lokasi tambahan...
              </p>
            ) : null}

            <ComparisonGrid locations={details.locations} table={table} names={names} onRemove={onRemove} />

            {table.hasEstimatedRisk ? (
              <p className="mt-3 text-xs leading-5 text-slate-500">
                Sebagian lokasi dievaluasi sebelum penanda risiko ikut disimpan. Untuk lokasi itu penanda risiko
                diperkirakan dari skor Fisik &amp; Lingkungan: di bawah {PHYSICAL_RISK_BELOW.high} berarti risiko tinggi,
                di bawah {PHYSICAL_RISK_BELOW.medium} berarti sedang. Evaluasi ulang lokasinya untuk penanda yang lengkap.
              </p>
            ) : null}

            <SaveComparisonForm selectedIds={selectedIds} opened={opened} isChanged={isChanged} onSaved={onSaved} />
          </>
        ) : null}
      </Card>
    </div>
  )
}

function SaveComparisonForm({ selectedIds, opened, isChanged, onSaved }: {
  selectedIds: string[]
  opened: OpenedComparison | null
  isChanged: boolean
  onSaved: (detail: ComparisonDetail) => void
}) {
  const [name, setName] = useState('')
  const [isBlank, setIsBlank] = useState(false)
  const create = useCreateComparison()
  const update = useUpdateComparison()

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = name.trim()
    setIsBlank(trimmed === '')
    if (trimmed === '') {
      return
    }
    update.reset()
    create.mutate(
      { name: trimmed, savedLocationIds: selectedIds },
      {
        onSuccess: (detail) => {
          setName('')
          onSaved(detail)
        },
      },
    )
  }

  function handleUpdate(target: OpenedComparison) {
    create.reset()
    update.mutate({ id: target.id, body: { name: target.name, savedLocationIds: selectedIds } }, { onSuccess: onSaved })
  }

  // Catatan berhasil hanya relevan selama perbandingan yang baru disimpan
  // masih yang sedang dibuka dan pilihannya belum diubah lagi.
  const isCurrent = (detail: ComparisonDetail) => detail.id === opened?.id && !isChanged

  return (
    <form onSubmit={handleSubmit} className="mt-6 border-t border-slate-100 pt-5">
      <h3 className="text-[15px] font-semibold text-slate-900">Simpan perbandingan</h3>
      <p className="mt-0.5 text-[13px] text-slate-500">
        Beri nama supaya pilihan {selectedIds.length} lokasi ini bisa dibuka lagi nanti.
      </p>

      <div className="mt-4 flex flex-wrap items-end gap-3">
        <div className="w-full sm:w-80">
          <Field
            id="comparison-name"
            label="Nama perbandingan"
            placeholder="Contoh: Kandidat RS Makassar"
            value={name}
            onChange={(event) => setName(event.target.value)}
            required
            maxLength={COMPARISON_NAME_MAX}
          />
        </div>
        <Button type="submit" isPending={create.isPending} pendingLabel="Menyimpan...">
          {opened ? 'Simpan sebagai perbandingan baru' : 'Simpan perbandingan'}
        </Button>
        {opened && isChanged ? (
          <Button variant="subtle" isPending={update.isPending} pendingLabel="Memperbarui..." onClick={() => handleUpdate(opened)}>
            Perbarui perbandingan yang dibuka
          </Button>
        ) : null}
      </div>

      <div className="mt-3 space-y-3">
        {isBlank ? <ErrorNote message="Isi nama perbandingan dulu." /> : null}
        {create.isError ? <ErrorNote message={errorMessage(create.error)} /> : null}
        {update.isError ? <ErrorNote message={errorMessage(update.error)} /> : null}
        {create.isSuccess && isCurrent(create.data) ? (
          <SuccessNote message={`Perbandingan "${create.data.name}" disimpan.`} />
        ) : null}
        {update.isSuccess && isCurrent(update.data) ? (
          <SuccessNote message={`Perbandingan "${update.data.name}" diperbarui dengan ${update.data.items.length} lokasi.`} />
        ) : null}
      </div>
    </form>
  )
}
