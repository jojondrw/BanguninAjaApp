import type { KesalahanApi, Sesi } from '../models/auth'
import { ambilAccessToken, hapusSesiLangsung, simpanSesiLangsung } from '../controllers/sesiStore'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api'

interface Amplop<T> {
  data?: T
  error?: KesalahanApi
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

interface OpsiPermintaan {
  method?: string
  body?: unknown
  tanpaPembaruanToken?: boolean
}

export async function panggilApi<T>(path: string, opsi: OpsiPermintaan = {}): Promise<T> {
  const respons = await kirim(path, opsi)

  if (respons.status === 401 && !opsi.tanpaPembaruanToken) {
    const berhasil = await perbaruiSesi()
    if (berhasil) {
      return bacaBody<T>(await kirim(path, opsi))
    }
    hapusSesiLangsung()
  }

  return bacaBody<T>(respons)
}

async function kirim(path: string, opsi: OpsiPermintaan): Promise<Response> {
  const accessToken = ambilAccessToken()

  return fetch(`${BASE_URL}${path}`, {
    method: opsi.method ?? 'GET',
    credentials: 'include',
    headers: {
      ...(opsi.body ? { 'Content-Type': 'application/json' } : {}),
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    body: opsi.body ? JSON.stringify(opsi.body) : undefined,
  })
}

async function bacaBody<T>(respons: Response): Promise<T> {
  if (respons.status === 204) {
    return undefined as T
  }

  const amplop = (await respons.json().catch(() => ({}))) as Amplop<T>

  if (!respons.ok || amplop.error) {
    throw new ApiError(
      amplop.error?.code ?? 'kesalahan_jaringan',
      amplop.error?.message ?? 'Tidak bisa menghubungi server',
      respons.status,
    )
  }

  return amplop.data as T
}

export async function perbaruiSesi(): Promise<boolean> {
  const respons = await kirim('/auth/refresh', { method: 'POST', tanpaPembaruanToken: true })
  if (!respons.ok) {
    return false
  }

  const amplop = (await respons.json()) as Amplop<Sesi>
  if (!amplop.data) {
    return false
  }

  simpanSesiLangsung(amplop.data.accessToken, amplop.data.user)
  return true
}
