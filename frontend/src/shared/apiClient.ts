import type { ApiErrorBody, Session } from '../models/auth'
import { clearSession, getAccessToken, setSession } from '../controllers/sessionStore'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api'

interface Envelope<T> {
  data?: T
  error?: ApiErrorBody
}

export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

export interface RequestOptions {
  method?: string
  body?: unknown
  skipTokenRefresh?: boolean
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await send(path, options)

  if (response.status === 401 && !options.skipTokenRefresh) {
    const refreshed = await refreshSession()
    if (refreshed) {
      return parseBody<T>(await send(path, options))
    }
    clearSession()
  }

  return parseBody<T>(response)
}

async function send(path: string, options: RequestOptions): Promise<Response> {
  const accessToken = getAccessToken()

  return fetch(`${BASE_URL}${path}`, {
    method: options.method ?? 'GET',
    credentials: 'include',
    headers: {
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
  })
}

async function parseBody<T>(response: Response): Promise<T> {
  if (response.status === 204) {
    return undefined as T
  }

  const envelope = (await response.json().catch(() => ({}))) as Envelope<T>

  if (!response.ok || envelope.error) {
    throw new ApiError(
      envelope.error?.code ?? 'network_error',
      envelope.error?.message ?? 'Tidak bisa menghubungi server',
      response.status,
    )
  }

  return envelope.data as T
}

export async function refreshSession(): Promise<boolean> {
  const response = await send('/auth/refresh', { method: 'POST', skipTokenRefresh: true })
  if (!response.ok) {
    return false
  }

  const envelope = (await response.json()) as Envelope<Session>
  if (!envelope.data) {
    return false
  }

  setSession(envelope.data.accessToken, envelope.data.user)
  return true
}
