export interface User {
  id: string
  name: string
  email: string
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
