import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  defaultReportName,
  type Report,
  type ReportCriteria,
  type ReportFilter,
  type ReportFormValues,
  type ReportRequest,
} from '../models/report'
import { reportApi, type FileSession } from '../models/reportApi'
import { toApiDate } from '../shared/localDate'
import { clearSession, getAccessToken } from './sessionStore'

const REPORTS_KEY = 'reports'
const LIST_KEY = [REPORTS_KEY, 'list'] as const
const PREVIEW_KEY = [REPORTS_KEY, 'preview'] as const
const ONE_MINUTE = 60_000
const URL_RELEASE_DELAY_MS = 1_000

const FILE_SESSION: FileSession = { accessToken: getAccessToken, onExpired: clearSession }

export function useReports(filter: ReportFilter) {
  return useQuery({
    queryKey: [...LIST_KEY, filter],
    queryFn: () => reportApi.reports(filter),
    staleTime: ONE_MINUTE,
    retry: 0,
    placeholderData: keepPreviousData,
  })
}

// Pratinjau baru diambil setelah tombol "Lihat pratinjau" ditekan, jadi
// kriterianya null sampai saat itu.
export function useReportPreview(criteria: ReportCriteria | null) {
  return useQuery({
    queryKey: [...PREVIEW_KEY, criteria],
    queryFn: () => reportApi.preview(criteria as ReportCriteria),
    enabled: criteria !== null,
    staleTime: 0,
    retry: 0,
  })
}

function toReportRequest(values: ReportFormValues, scope: string): ReportRequest {
  const name = values.name.trim() || defaultReportName(values.type, values.periodStart, values.periodEnd)
  return {
    name,
    type: values.type,
    scope,
    projectId: values.projectId || null,
    periodStart: values.periodStart ? toApiDate(values.periodStart) : null,
    periodEnd: values.periodEnd ? toApiDate(values.periodEnd) : null,
    format: values.format,
  }
}

// Unduhan dipicu lewat tautan sementara ke object URL karena berkasnya datang
// dari fetch yang membawa header Authorization.
function saveFile(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.append(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), URL_RELEASE_DELAY_MS)
}

async function downloadAndSave(report: Report): Promise<string> {
  const file = await reportApi.download(report, FILE_SESSION)
  saveFile(file.blob, file.fileName)
  return file.fileName
}

export function useCreateReport() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ values, scope }: { values: ReportFormValues; scope: string }) => {
      const report = await reportApi.createReport(toReportRequest(values, scope))
      void queryClient.invalidateQueries({ queryKey: LIST_KEY })
      return downloadAndSave(report)
    },
  })
}

export function useDownloadReport() {
  return useMutation({ mutationFn: downloadAndSave })
}

export function useDeleteReport() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => reportApi.deleteReport(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: LIST_KEY })
    },
  })
}
