import { API_BASE_URL, ApiError, refreshSession, request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import {
  fallbackFileName,
  type Report,
  type ReportCriteria,
  type ReportFile,
  type ReportFilter,
  type ReportPreview,
  type ReportRequest,
} from './report'

const FILE_NAME_PATTERN = /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i

// Sesi dipegang controller, jadi token dan tindakan saat sesi habis dioper
// dari sana supaya model tidak bergantung ke controller.
export interface FileSession {
  accessToken: () => string | null
  onExpired: () => void
}

function fetchFile(path: string, session: FileSession): Promise<Response> {
  const accessToken = session.accessToken()
  return fetch(`${API_BASE_URL}${path}`, {
    credentials: 'include',
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
  })
}

function fileNameFrom(disposition: string | null): string | null {
  const match = disposition?.match(FILE_NAME_PATTERN)
  return match ? decodeURIComponent(match[1]) : null
}

async function failure(response: Response): Promise<ApiError> {
  const envelope = (await response.json().catch(() => ({}))) as { error?: { code: string; message: string } }
  return new ApiError(
    envelope.error?.code ?? 'network_error',
    envelope.error?.message ?? 'Tidak bisa menghubungi server',
    response.status,
  )
}

// Berkas laporan butuh header Authorization, jadi tidak bisa lewat tautan biasa.
// Diambil sebagai blob, dengan percobaan ulang sekali setelah token disegarkan.
async function download(report: Report, session: FileSession): Promise<ReportFile> {
  const path = `/reports/${report.id}/download`
  let response = await fetchFile(path, session)

  if (response.status === 401) {
    if (await refreshSession()) {
      response = await fetchFile(path, session)
    } else {
      session.onExpired()
    }
  }

  if (!response.ok) {
    throw await failure(response)
  }

  return {
    blob: await response.blob(),
    fileName: fileNameFrom(response.headers.get('Content-Disposition')) ?? fallbackFileName(report),
  }
}

export const reportApi = {
  reports: (filter: ReportFilter = {}) => request<Page<Report>>(`/reports${toQueryString({ ...filter })}`),

  createReport: (body: ReportRequest) => request<Report>('/reports', { method: 'POST', body }),

  deleteReport: (id: string) => request<void>(`/reports/${id}`, { method: 'DELETE' }),

  preview: (criteria: ReportCriteria) =>
    request<ReportPreview>(
      `/reports/preview${toQueryString({
        type: criteria.type,
        projectId: criteria.projectId,
        periodStart: criteria.periodStart,
        periodEnd: criteria.periodEnd,
      })}`,
    ),

  download,
}
