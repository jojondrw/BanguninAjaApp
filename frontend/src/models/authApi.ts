import { request } from '../shared/apiClient'
import type { ChangePasswordRequest, LoginRequest, RegisterRequest, Session, UpdateProfileRequest, User } from './auth'

export const authApi = {
  register: (payload: RegisterRequest) =>
    request<User>('/auth/register', { method: 'POST', body: payload }),

  login: (payload: LoginRequest) =>
    request<Session>('/auth/login', { method: 'POST', body: payload, skipTokenRefresh: true }),

  logout: () => request<void>('/auth/logout', { method: 'POST', skipTokenRefresh: true }),

  profile: () => request<User>('/auth/me'),

  updateProfile: (payload: UpdateProfileRequest) => request<User>('/auth/me', { method: 'PATCH', body: payload }),

  // Sesi yang sedang dipakai tetap berlaku; refresh token di perangkat lain dicabut.
  changePassword: (payload: ChangePasswordRequest) =>
    request<void>('/auth/password', { method: 'PATCH', body: payload }),
}
