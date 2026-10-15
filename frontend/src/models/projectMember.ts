import type { UserRole } from './auth'
import type { RoleCode } from './user'

// Anggota tim proyek, dari GET /projects/:id/members. role adalah peran
// aplikasi di proyek ini, bawaannya sama dengan peran akun saat ditambahkan.
export interface ProjectMember {
  id: string
  projectId: string
  userId: string
  name: string
  email: string
  active: boolean
  position: string | null
  role: UserRole
  joinedAt: string
  updatedAt: string
}

// Pengguna aktif yang belum masuk tim. Sengaja hanya id, nama, dan email.
export interface MemberCandidate {
  id: string
  name: string
  email: string
}

// Sama dengan MemberCreateRequest di backend. role kosong berarti ikut peran akun.
export interface MemberCreateRequest {
  userId: string
  position?: string
  role?: RoleCode
}

// PATCH: field yang tidak dikirim tidak berubah, posisi kosong menghapus posisi.
export interface MemberUpdateRequest {
  position: string
  role: RoleCode
}

// Batas panjang mengikuti tag binding position (max=80) di backend.
export const MEMBER_POSITION_MAX_LENGTH = 80

export const MEMBER_POSITION_SUGGESTIONS = [
  'Manajer proyek',
  'Site manager',
  'Pengawas lapangan',
  'Estimator',
  'Quantity surveyor',
  'Drafter',
  'Admin proyek',
  'Petugas K3',
]

export interface MemberFormValues {
  userId: string
  position: string
  role: RoleCode | ''
}

export const EMPTY_MEMBER_FORM: MemberFormValues = { userId: '', position: '', role: '' }

export interface MemberEditValues {
  position: string
  role: RoleCode
}

export function toMemberCreateRequest(values: MemberFormValues): MemberCreateRequest {
  const position = values.position.trim()
  return {
    userId: values.userId,
    ...(position === '' ? {} : { position }),
    ...(values.role === '' ? {} : { role: values.role }),
  }
}

export function memberEditValues(member: ProjectMember): MemberEditValues {
  return { position: member.position ?? '', role: member.role.code as RoleCode }
}

export function toMemberUpdateRequest(values: MemberEditValues): MemberUpdateRequest {
  return { position: values.position.trim(), role: values.role }
}

// Kalimat konfirmasi setelah anggota ditambahkan atau diubah.
export function memberSummary(member: ProjectMember): string {
  return member.position ? `${member.name} sebagai ${member.position}` : member.name
}
