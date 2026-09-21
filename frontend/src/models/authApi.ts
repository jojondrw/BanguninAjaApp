import { request } from '../shared/apiClient'
import type { LoginRequest, RegisterRequest, Session, User } from './auth'

export const authApi = {
  register: (payload: RegisterRequest) =>
    request<User>('/auth/register', { method: 'POST', body: payload }),

  login: (payload: LoginRequest) =>
    request<Session>('/auth/login', { method: 'POST', body: payload, skipTokenRefresh: true }),

  logout: () => request<void>('/auth/logout', { method: 'POST', skipTokenRefresh: true }),

  profile: () => request<User>('/auth/me'),
}
