import type { UserRole } from './auth'

export interface ManagedUser {
  id: string
  name: string
  email: string
  active: boolean
  role: UserRole | null
  createdAt: string
}

// Nilai saringan peran untuk akun yang belum diberi peran (pendaftar baru).
// Sama dengan pendingRoleFilter di backend (internal/auth/dto.go).
export const PENDING_ROLE_FILTER = 'none'

export interface UserFilter {
  search?: string
  role?: string
  active?: 'true' | 'false'
  page?: number
  pageSize?: number
}

// Daftar tetap, sama dengan peran yang di-seed backend (internal/shared/access).
export const ROLE_OPTIONS = [
  { code: 'admin', label: 'Admin', hint: 'Semua modul dan kelola pengguna' },
  { code: 'manajer_proyek', label: 'Manajer Proyek', hint: 'Proyek, lokasi, inventaris; lihat pengadaan dan keuangan' },
  { code: 'keuangan', label: 'Keuangan', hint: 'Keuangan dan tagihan; lihat proyek, penjualan, pengadaan' },
  { code: 'pengadaan', label: 'Pengadaan', hint: 'Pengadaan dan inventaris' },
  { code: 'penjualan', label: 'Penjualan', hint: 'Penjualan; lihat proyek dan tagihan' },
  { code: 'sdm', label: 'SDM', hint: 'Karyawan, absensi, penggajian' },
  { code: 'pengamat', label: 'Pengamat', hint: 'Lihat semua modul tanpa mengubah' },
] as const

export type RoleCode = (typeof ROLE_OPTIONS)[number]['code']

export function roleLabel(code: string | undefined): string {
  return ROLE_OPTIONS.find((option) => option.code === code)?.label ?? 'Tanpa peran'
}
