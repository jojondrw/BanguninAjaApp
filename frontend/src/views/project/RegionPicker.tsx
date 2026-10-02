import { type KeyboardEvent, useEffect, useState } from 'react'

import { useRegion, useRegionSearch } from '../../controllers/useProjectWorkspace'
import { REGION_SEARCH_MIN_LENGTH, regionLabel, type Region } from '../../models/project'
import { Field, SelectField } from '../components/Form'

const SEARCH_DELAY_MS = 300

// Menunggu pengguna berhenti mengetik sebelum kata kunci dipakai, supaya tidak
// ada satu permintaan per huruf.
function useSettledValue(value: string, delay: number): string {
  const [settled, setSettled] = useState(value)

  useEffect(() => {
    const timer = window.setTimeout(() => setSettled(value), delay)
    return () => window.clearTimeout(timer)
  }, [value, delay])

  return settled
}

function searchHint(term: string, isFetching: boolean, isError: boolean, found: number | undefined): string {
  if (term.length < REGION_SEARCH_MIN_LENGTH) {
    return `Ketik minimal ${REGION_SEARCH_MIN_LENGTH} huruf, misalnya Bandung`
  }
  if (isError) {
    return 'Gagal mencari wilayah. Coba kata lain atau ulangi sebentar lagi.'
  }
  if (isFetching || found === undefined) {
    return 'Mencari wilayah...'
  }
  if (found === 0) {
    return `Tidak ada wilayah yang cocok dengan "${term}"`
  }
  return `${found} wilayah cocok. Pilih di kolom Wilayah.`
}

// Dua isian berdampingan: kolom cari mengisi pilihan di kolom Wilayah. Nilai
// yang sedang dipakai selalu ada di pilihan walaupun tidak muncul di hasil cari.
export function RegionPicker({ idPrefix, value, onChange }: {
  idPrefix: string
  value: string
  onChange: (regionId: string) => void
}) {
  const [search, setSearch] = useState('')
  const term = useSettledValue(search.trim(), SEARCH_DELAY_MS)
  const current = useRegion(value)
  const results = useRegionSearch(term)

  const found = term.length >= REGION_SEARCH_MIN_LENGTH ? results.data?.items : undefined
  const selected = current.data?.id === value ? current.data : found?.find((region) => region.id === value)
  const options: Region[] = [
    ...(selected ? [selected] : []),
    ...(found ?? []).filter((region) => region.id !== value),
  ]
  const isCurrentLoading = value !== '' && !selected

  // Enter di kolom cari tidak boleh mengirim formulir proyek.
  const keepFormOpen = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      event.preventDefault()
    }
  }

  return (
    <>
      <Field
        id={`${idPrefix}-region-search`}
        label="Cari wilayah"
        type="search"
        autoComplete="off"
        placeholder="Nama kota atau kabupaten"
        hint={searchHint(term, results.isFetching, results.isError, found?.length)}
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        onKeyDown={keepFormOpen}
      />
      <SelectField
        id={`${idPrefix}-region`}
        label="Wilayah (opsional)"
        hint={current.isError ? 'Wilayah tersimpan tidak bisa dimuat' : 'Dipakai untuk analisis lokasi dan laporan per wilayah'}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        <option value="">Tanpa wilayah</option>
        {isCurrentLoading ? <option value={value}>Memuat wilayah...</option> : null}
        {options.map((region) => (
          <option key={region.id} value={region.id}>
            {regionLabel(region)}
          </option>
        ))}
      </SelectField>
    </>
  )
}
