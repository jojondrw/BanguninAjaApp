import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { ManagedUser, RoleCode, UserAuditEntry, UserFilter } from './user'

export const userApi = {
  list: (filter: UserFilter = {}) => request<Page<ManagedUser>>(`/users${toQueryString({ ...filter })}`),

  changeRole: (id: string, role: RoleCode) =>
    request<ManagedUser>(`/users/${id}/role`, { method: 'PATCH', body: { role } }),

  changeActive: (id: string, active: boolean) =>
    request<ManagedUser>(`/users/${id}/active`, { method: 'PATCH', body: { active } }),

  audit: () => request<UserAuditEntry[]>('/users/audit'),
}
