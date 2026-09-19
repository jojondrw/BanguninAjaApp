import { ApiError } from './apiClient'

export function pesanKesalahan(kesalahan: unknown): string {
  if (kesalahan instanceof ApiError) {
    return kesalahan.message
  }
  return 'Tidak bisa menghubungi server. Coba lagi sebentar.'
}
