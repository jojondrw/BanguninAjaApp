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

export interface ApiErrorBody {
  code: string
  message: string
}
