import { ApiError } from './apiClient'

export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message
  }
  return 'Tidak bisa menghubungi server. Coba lagi sebentar.'
}
