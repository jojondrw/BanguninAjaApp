import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import type { SitePhotoUpload } from '../models/sitePhoto'
import { sitePhotoApi } from '../models/sitePhotoApi'
import { DATA_QUERY, PROJECTS_KEY } from './useErp'

const SITE_PHOTOS = 'site-photos'
const PHOTO_FILE = 'site-photo-file'

export function useSitePhotos(projectId: string) {
  return useQuery({
    queryKey: [PROJECTS_KEY, 'detail', projectId, SITE_PHOTOS],
    queryFn: () => sitePhotoApi.list(projectId),
    ...DATA_QUERY,
  })
}

// Isi foto tersimpan tidak pernah berubah, jadi URL objek dari blob-nya
// disimpan di cache query sepanjang sesi dan dipakai bersama semua kartu.
export function useSitePhotoUrl(projectId: string, photoId: string | null | undefined): string | null {
  const file = useQuery({
    queryKey: [PHOTO_FILE, photoId],
    queryFn: async () => URL.createObjectURL(await sitePhotoApi.file(projectId, photoId as string)),
    enabled: Boolean(photoId),
    staleTime: Infinity,
    gcTime: Infinity,
  })
  return photoId ? (file.data ?? null) : null
}

export function useUploadSitePhoto(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (upload: SitePhotoUpload) => sitePhotoApi.upload(projectId, upload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] }),
  })
}

export function useDeleteSitePhoto(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (photoId: string) => sitePhotoApi.remove(projectId, photoId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY] }),
  })
}
