import { API_BASE_URL, ApiError, refreshSession } from '../shared/apiClient'
import type { FileSession } from './reportApi'

export interface LocationReportFile {
  blob: Blob
  fileName: string
}

const FILE_NAME_PATTERN = /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i
const SITE_FALLBACK_NAME = 'analisis-lokasi.pdf'
const COMPARISON_FALLBACK_NAME = 'perbandingan-lokasi.pdf'

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

// Sama dengan unduhan di halaman Laporan: PDF butuh header Authorization, jadi
// diambil lewat fetch sebagai blob, dengan percobaan ulang sekali setelah token
// disegarkan. Pembantu di reportApi khusus riwayat laporan, jadi tidak dipakai.
async function downloadPdf(path: string, fallbackName: string, session: FileSession): Promise<LocationReportFile> {
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
    fileName: fileNameFrom(response.headers.get('Content-Disposition')) ?? fallbackName,
  }
}

export const locationReportApi = {
  site: (savedLocationId: string, session: FileSession) =>
    downloadPdf(`/locations/saved/${savedLocationId}/report`, SITE_FALLBACK_NAME, session),

  comparison: (comparisonId: string, session: FileSession) =>
    downloadPdf(`/locations/comparisons/${comparisonId}/report`, COMPARISON_FALLBACK_NAME, session),
}
