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

export interface UserAuditEntry {
  id: string
  actorName: string | null
  userName: string | null
  userEmail: string | null
  action: 'role' | 'active'
  oldValue: string
  newValue: string
  createdAt: string
}

// Satu kalimat per perubahan. Pelaku kosong berarti perubahan dari sistem,
// misalnya seed data demo.
export function auditSentence(entry: UserAuditEntry): string {
  const actor = entry.actorName ?? 'Sistem'
  const target = entry.userName ?? 'pengguna yang sudah dihapus'
  if (entry.action === 'active') {
    return entry.newValue === 'true' ? `${actor} mengaktifkan lagi akun ${target}` : `${actor} menonaktifkan akun ${target}`
  }
  return `${actor} mengubah peran ${target} dari ${roleLabel(entry.oldValue || undefined)} menjadi ${roleLabel(entry.newValue || undefined)}`
}

