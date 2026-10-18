import { getAccessToken } from '../controllers/sessionStore'
import { API_BASE_URL, ApiError, refreshSession, request } from '../shared/apiClient'
import type { SitePhoto, SitePhotoUpload } from './sitePhoto'

function photosPath(projectId: string): string {
  return `/projects/${projectId}/site-photos`
}

// Unggahan dan berkas foto tidak berupa JSON, jadi dikirim lewat fetch langsung
// dengan header Authorization dan satu kali coba ulang setelah token disegarkan.
async function authorizedFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const send = () => {
    const accessToken = getAccessToken()
    return fetch(`${API_BASE_URL}${path}`, {
      ...init,
      credentials: 'include',
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
    })
  }
  const response = await send()
  if (response.status === 401 && (await refreshSession())) {
    return send()
  }
  return response
}

async function failure(response: Response): Promise<ApiError> {
  const envelope = (await response.json().catch(() => ({}))) as { error?: { code: string; message: string } }
  return new ApiError(
    envelope.error?.code ?? 'network_error',
    envelope.error?.message ?? 'Tidak bisa menghubungi server',
    response.status,
  )
}

export const sitePhotoApi = {
  list: (projectId: string) => request<SitePhoto[]>(photosPath(projectId)),

  upload: async (projectId: string, upload: SitePhotoUpload): Promise<SitePhoto> => {
    const form = new FormData()
    form.append('file', upload.file)
    form.append('caption', upload.caption.trim())
    if (upload.takenOn) {
      form.append('takenOn', upload.takenOn)
    }
    const response = await authorizedFetch(photosPath(projectId), { method: 'POST', body: form })
    if (!response.ok) {
      throw await failure(response)
    }
    return ((await response.json()) as { data: SitePhoto }).data
  },

  file: async (projectId: string, photoId: string): Promise<Blob> => {
    const response = await authorizedFetch(`${photosPath(projectId)}/${photoId}/file`)
    if (!response.ok) {
      throw await failure(response)
    }
    return response.blob()
  },

  remove: (projectId: string, photoId: string) =>
    request<void>(`${photosPath(projectId)}/${photoId}`, { method: 'DELETE' }),
}
