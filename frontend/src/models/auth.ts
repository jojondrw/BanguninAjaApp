// Kunci modul sama dengan matriks backend di internal/shared/access/matrix.go.
export type ModuleKey =
  | 'projects'
  | 'locations'
  | 'finance'
  | 'billing'
  | 'sales'
  | 'procurement'
  | 'inventory'
  | 'hr'
  | 'master'
  | 'reports'
  | 'users'

export type AccessLevel = 'read' | 'write'

export type Permissions = Partial<Record<ModuleKey, AccessLevel>>

export interface UserRole {
  code: string
  name: string
}

export interface User {
  id: string
  name: string
  email: string
  active: boolean
  role: UserRole | null
  permissions: Permissions
  createdAt: string
}

export interface Session {
  accessToken: string
  accessTokenExpiresAt: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

// Batas sama dengan binding backend (internal/auth/dto.go): nama 2–100 karakter
// setelah dipangkas, kata sandi 8–72 karakter karena bcrypt memotong setelah 72 byte.
export const NAME_MIN_LENGTH = 2
export const NAME_MAX_LENGTH = 100
export const PASSWORD_MIN_LENGTH = 8
export const PASSWORD_MAX_LENGTH = 72

export interface UpdateProfileRequest {
  name: string
}

export interface ChangePasswordRequest {
  currentPassword: string
  newPassword: string
}

export interface ApiErrorBody {
  code: string
  message: string
}
