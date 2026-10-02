import { useState } from 'react'

import { useProjects, useSavedLocations } from '../../controllers/useErp'
import type { SavedLocation } from '../../models/erpApi'
import { MAX_COMPARED, MIN_COMPARED, searchSites } from '../../models/location'
import { number, shortDate } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { FilterSelect, Toolbar, ToolbarInput } from '../components/RecordControls'
import { ROW_BUTTON } from './buttons'

const PROJECT_OPTION_LIMIT = 100
const NO_SITES: SavedLocation[] = []

interface LocationPickerProps {
  projectId: string
  onProjectChange: (projectId: string) => void
  selectedIds: string[]
  onToggle: (id: string) => void
  onClear: () => void
}

function emptyMessage(total: number, isProjectChosen: boolean, isSearching: boolean): string {
  if (isSearching) {
    return 'Tidak ada lokasi yang namanya cocok dengan pencarian.'
  }
  if (isProjectChosen) {
    return total === 0
      ? 'Proyek ini belum punya lokasi tersimpan. Pilih "Semua proyek" untuk melihat lokasi lain.'
      : 'Proyek ini baru punya 1 lokasi tersimpan. Pilih "Semua proyek" untuk membandingkannya dengan lokasi lain.'
  }
  return total === 0
    ? 'Belum ada lokasi tersimpan.'
    : 'Baru ada 1 lokasi tersimpan. Evaluasi satu lokasi lagi supaya bisa dibandingkan.'
}

export function LocationPicker({ projectId, onProjectChange, selectedIds, onToggle, onClear }: LocationPickerProps) {
  const [keyword, setKeyword] = useState('')
  const sites = useSavedLocations(projectId || undefined)
  const projects = useProjects({ pageSize: PROJECT_OPTION_LIMIT })

  const items = sites.data?.items ?? NO_SITES
  const shown = searchSites(items, keyword)
  const isFull = selectedIds.length >= MAX_COMPARED
  const isSearching = keyword.trim() !== ''
  const tooFew = !isSearching && items.length < MIN_COMPARED

  const projectName = (id: string | null) =>
    id ? (projects.data?.items.find((project) => project.id === id)?.name ?? 'Proyek lain') : 'Tanpa proyek'

  return (
    <Card
      title="Pilih lokasi"
      description={`Centang ${MIN_COMPARED} sampai ${MAX_COMPARED} lokasi tersimpan. Urutan kolom mengikuti urutan centang.`}
    >
      <Toolbar>
        <FilterSelect
          id="compare-project"
          label="Saring lokasi menurut proyek"
          value={projectId}
          onChange={(event) => onProjectChange(event.target.value)}
          disabled={projects.isPending}
        >
          <option value="">{projects.isPending ? 'Memuat proyek...' : 'Semua proyek'}</option>
          {projects.data?.items.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name} ({project.code})
            </option>
          ))}
        </FilterSelect>
        <ToolbarInput
          id="compare-search"
          label="Cari nama lokasi"
          type="search"
          placeholder="Cari nama lokasi"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
        />
        <p role="status" className="ml-auto text-xs text-slate-500 tabular-nums">
          {selectedIds.length} dipilih · maksimal {MAX_COMPARED}
        </p>
        {selectedIds.length > 0 ? (
          <button type="button" onClick={onClear} className={ROW_BUTTON}>
            Kosongkan pilihan
          </button>
        ) : null}
      </Toolbar>

      {projects.isError ? (
        <p role="alert" className="mb-3 text-xs text-red-600">
          Gagal mengambil daftar proyek, jadi nama proyek tidak bisa ditampilkan.
        </p>
      ) : null}

      {isFull ? (
        <p className="mb-3 text-xs text-slate-500">
          Sudah {MAX_COMPARED} lokasi dipilih. Lepas salah satu untuk memilih lokasi lain.
        </p>
      ) : null}

      {sites.isPending ? <Loading label="Mengambil lokasi tersimpan..." /> : null}
      {sites.isError ? <LoadFailed onRetry={() => sites.refetch()} /> : null}
      {sites.data && (shown.length === 0 || tooFew) ? (
        <div className={shown.length > 0 ? 'mb-3' : ''}>
          <Empty message={emptyMessage(items.length, projectId !== '', isSearching)} />
        </div>
      ) : null}
      {sites.data && shown.length > 0 ? (
        <ul aria-label="Lokasi tersimpan" className="grid max-h-[26rem] gap-2 overflow-y-auto p-0.5 sm:grid-cols-2 2xl:grid-cols-3">
          {shown.map((site) => {
            const isChecked = selectedIds.includes(site.id)
            const isDisabled = !isChecked && isFull
            const order = selectedIds.indexOf(site.id) + 1

            return (
              <li key={site.id}>
                <label
                  htmlFor={`compare-pick-${site.id}`}
                  className={`flex items-start gap-3 rounded-xl p-3 transition-[background-color,box-shadow] ${
                    isChecked
                      ? 'bg-navy-50 shadow-[inset_0_0_0_1px_var(--color-navy-200)]'
                      : 'bg-white shadow-hairline hover:bg-slate-50'
                  } ${isDisabled ? 'cursor-not-allowed opacity-55' : 'cursor-pointer'}`}
                >
                  <input
                    id={`compare-pick-${site.id}`}
                    type="checkbox"
                    checked={isChecked}
                    disabled={isDisabled}
                    onChange={() => onToggle(site.id)}
                    className="mt-0.5 size-4 shrink-0 accent-navy-700"
                  />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-[13px] font-medium text-slate-900">{site.name}</span>
                    <span className="mt-0.5 block truncate text-xs text-slate-500">
                      {projectName(site.projectId)} · {shortDate(site.savedAt)}
                    </span>
                    {isChecked ? (
                      <span className="mt-1 block text-[11px] font-medium text-navy-700">Kolom {order}</span>
                    ) : null}
                  </span>
                  <span className="shrink-0 text-right">
                    <span className="block text-sm font-semibold text-slate-900 tabular-nums">{site.score}</span>
                    <span className="block text-[11px] text-slate-500">skor</span>
                  </span>
                </label>
              </li>
            )
          })}
        </ul>
      ) : null}

      {sites.data && sites.data.totalItems > items.length ? (
        <p className="mt-3 text-xs text-slate-500">
          Menampilkan {number(items.length)} lokasi dengan skor tertinggi dari {number(sites.data.totalItems)} lokasi
          tersimpan.
        </p>
      ) : null}
    </Card>
  )
}
