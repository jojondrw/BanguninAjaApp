import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'

export const PAGE_SIZE = 20

export function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return fallback
  }
  return rupiah(amount)
}

// Alasan penolakan datang dari backend. Kalimat pembukanya menyebut data mana
// yang ditolak, supaya jelas baris mana yang gagal.
export function refusalText(opening: string, error: unknown): string {
  const reason = errorMessage(error)
  return `${opening} ${reason.endsWith('.') ? reason : `${reason}.`}`
}

// Halaman yang isinya baru saja habis dihapus diganti ke halaman sebelumnya,
// supaya tidak tampil tabel kosong.
export function pageAfterRemoval(page: number, rowsOnPage: number): number {
  return rowsOnPage <= 1 && page > 1 ? page - 1 : page
}
