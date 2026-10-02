import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'

export const PAGE_SIZE = 10
export const OPTION_LIMIT = 100

// Pesan penolakan backend diberi konteks baris yang ditindak, misalnya
// "Unit A-01 tidak bisa dihapus. Unit masih dipakai oleh kontrak."
export function refusalText(context: string, error: unknown): string {
  const message = errorMessage(error)
  return `${context} ${/[.!?]$/.test(message) ? message : `${message}.`}`
}

// Setelah baris terakhir di halaman terakhir dihapus, pindah satu halaman ke
// belakang supaya tabel tidak tampil kosong.
export function pageAfterRemoval(page: number, rowsOnPage: number): number {
  return rowsOnPage <= 1 && page > 1 ? page - 1 : page
}

export function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return fallback
  }
  return rupiah(amount)
}

export function projectLabel(projects: Project[], id: string | null): string {
  if (id === null) {
    return '-'
  }
  const project = projects.find((item) => item.id === id)
  return project ? project.name : '-'
}
