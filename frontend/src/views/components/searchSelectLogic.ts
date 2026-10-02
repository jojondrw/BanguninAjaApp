import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'

import type { OptionSource, SearchOption } from '../../models/lookup'

export const SEARCH_DELAY_MS = 250

// Nilai yang baru ikut berubah setelah pengguna berhenti mengetik sejenak, supaya
// server tidak dipanggil di setiap huruf.
export function useDebounced<T>(value: T, delay = SEARCH_DELAY_MS): T {
  const [settled, setSettled] = useState(value)

  useEffect(() => {
    const timer = window.setTimeout(() => setSettled(value), delay)
    return () => window.clearTimeout(timer)
  }, [value, delay])

  return settled
}

export function optionDomId(listId: string, index: number): string {
  return `${listId}-option-${index}`
}

// Pindah sorotan dengan panah, berputar di ujung daftar.
export function stepIndex(current: number, count: number, step: 1 | -1): number {
  if (count === 0) return -1
  if (current < 0 || current >= count) return step === 1 ? 0 : count - 1
  return (current + step + count) % count
}

// Label pilihan saat ini dicari dari yang baru dipilih, isian awal, hasil
// pencarian, lalu hasil ambil-per-id. Kosong kalau belum ada yang tahu.
export function resolveLabel(value: string, candidates: (SearchOption | null | undefined)[], results: SearchOption[]) {
  if (value === '') return ''
  for (const candidate of candidates) {
    if (candidate && candidate.value === value) return candidate.label
  }
  return results.find((option) => option.value === value)?.label ?? ''
}

// Label satu data lewat id, memakai cache yang sama dengan SearchSelect. Untuk
// judul rincian dan baris tabel yang hanya membawa id.
export function useLookupLabel(source: OptionSource, value: string | null | undefined): string | undefined {
  const query = useQuery({
    queryKey: ['search-select', source.cacheKey, 'selected', value ?? ''],
    queryFn: () => source.fetchSelected(value ?? ''),
    enabled: Boolean(value),
    staleTime: 5 * 60_000,
    retry: false,
  })
  if (!value) return undefined
  return query.data?.label ?? (query.isError ? '-' : '…')
}
