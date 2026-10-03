import { keepPreviousData, useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  comparisonLocationIds,
  type ComparisonFilter,
  type ComparisonRequest,
} from '../models/location'
import { projectApi } from '../models/erpApi'
import { comparisonApi, dimensionApi, regionApi, savedLocationApi } from '../models/locationApi'
import { DATA_QUERY, PROJECTS_KEY, SAVED_LOCATIONS_KEY } from './useErp'

export const COMPARISONS_KEY = 'location-comparisons'
const DIMENSIONS_KEY = 'scoring-dimensions'
const REGIONS_KEY = 'regions'
const TEN_MINUTES = 600_000

// Rincian disimpan di bawah kunci lokasi tersimpan, jadi evaluasi baru yang
// membatalkan kunci itu ikut menyegarkan rinciannya.
export function useSavedLocationDetails(ids: string[]) {
  return useQueries({
    queries: ids.map((id) => ({
      queryKey: [SAVED_LOCATIONS_KEY, 'detail', id],
      queryFn: () => savedLocationApi.get(id),
      ...DATA_QUERY,
    })),
    combine: (results) => ({
      locations: results.flatMap((result) => (result.data ? [result.data] : [])),
      isPending: results.some((result) => result.isPending),
      error: results.find((result) => result.isError)?.error ?? null,
      failedIds: results.flatMap((result, index) => (result.isError ? [ids[index]] : [])),
      retry: () => {
        for (const result of results) {
          if (result.isError) {
            void result.refetch()
          }
        }
      },
    }),
  })
}

// Dimensi skor jarang berubah, jadi disimpan lebih lama dari data biasa.
export function useDimensions() {
  return useQuery({
    queryKey: [DIMENSIONS_KEY],
    queryFn: dimensionApi.list,
    ...DATA_QUERY,
    staleTime: TEN_MINUTES,
  })
}

export function useRegionNames(ids: string[]) {
  return useQueries({
    queries: [...new Set(ids)].map((id) => ({
      queryKey: [REGIONS_KEY, id],
      queryFn: () => regionApi.get(id),
      ...DATA_QUERY,
      staleTime: TEN_MINUTES,
    })),
    combine: (results) =>
      new Map(results.flatMap((result) => (result.data ? [[result.data.id, result.data.name] as const] : []))),
  })
}

// Nama proyek per id. Kuncinya sama dengan kepala proyek di halaman proyek,
// jadi proyek yang sudah pernah dibuka tidak diambil ulang.
export function useProjectNames(ids: string[]) {
  return useQueries({
    queries: [...new Set(ids)].map((id) => ({
      queryKey: [PROJECTS_KEY, 'detail', id],
      queryFn: () => projectApi.get(id),
      ...DATA_QUERY,
    })),
    combine: (results) =>
      new Map(results.flatMap((result) => (result.data ? [[result.data.id, result.data.name] as const] : []))),
  })
}

export function useComparisons(filter: ComparisonFilter = {}) {
  return useQuery({
    queryKey: [COMPARISONS_KEY, 'list', filter],
    queryFn: () => comparisonApi.list(filter),
    ...DATA_QUERY,
    placeholderData: keepPreviousData,
  })
}

// Membuka perbandingan tersimpan dipicu klik, jadi dibungkus mutation supaya
// tombolnya punya keadaan sedang memuat dan pesan gagal sendiri.
export function useOpenComparison() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) =>
      queryClient.fetchQuery({
        queryKey: [COMPARISONS_KEY, 'detail', id],
        queryFn: () => comparisonApi.get(id),
        ...DATA_QUERY,
      }),
  })
}

function useComparisonMutation<TVariables, TData>(mutationFn: (variables: TVariables) => Promise<TData>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [COMPARISONS_KEY] }),
  })
}

export function useCreateComparison() {
  return useComparisonMutation((body: ComparisonRequest) => comparisonApi.create(body))
}

export function useUpdateComparison() {
  return useComparisonMutation(({ id, body }: { id: string; body: ComparisonRequest }) =>
    comparisonApi.update(id, body),
  )
}

// PUT backend selalu meminta nama dan daftar lokasi sekaligus, sedangkan daftar
// perbandingan tidak memuat lokasinya. Rinciannya diambil dulu supaya lokasi
// yang tersimpan tetap utuh saat hanya namanya yang diganti.
export function useRenameComparison() {
  return useComparisonMutation(async ({ id, name }: { id: string; name: string }) => {
    const detail = await comparisonApi.get(id)
    return comparisonApi.update(id, { name, savedLocationIds: comparisonLocationIds(detail) })
  })
}

export function useDeleteComparison() {
  return useComparisonMutation(({ id }: { id: string; name: string }) => comparisonApi.remove(id))
}
