import { useMutation } from '@tanstack/react-query'

import { locationReportApi, type LocationReportFile } from '../models/locationReportApi'
import type { FileSession } from '../models/reportApi'
import { clearSession, getAccessToken } from './sessionStore'

const URL_RELEASE_DELAY_MS = 1_000

const FILE_SESSION: FileSession = { accessToken: getAccessToken, onExpired: clearSession }

// Unduhan dipicu lewat tautan sementara ke object URL karena berkasnya datang
// dari fetch yang membawa header Authorization. Nama berkas dikembalikan supaya
// tampilan bisa memberi tahu berkas mana yang sudah diunduh.
function saveFile(file: LocationReportFile): string {
  const url = URL.createObjectURL(file.blob)
  const link = document.createElement('a')
  link.href = url
  link.download = file.fileName
  document.body.append(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), URL_RELEASE_DELAY_MS)
  return file.fileName
}

export function useDownloadSiteReport() {
  return useMutation({
    mutationFn: async (savedLocationId: string) => saveFile(await locationReportApi.site(savedLocationId, FILE_SESSION)),
  })
}

export function useDownloadComparisonReport() {
  return useMutation({
    mutationFn: async (comparisonId: string) => saveFile(await locationReportApi.comparison(comparisonId, FILE_SESSION)),
  })
}
