import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type { SearchOption } from './lookup'
import { SEARCH_LIMIT } from './lookupApi'
import type { MemberCandidate, MemberCreateRequest, MemberUpdateRequest, ProjectMember } from './projectMember'

function membersPath(projectId: string): string {
  return `/projects/${projectId}/members`
}

export function memberCandidatesPath(projectId: string): string {
  return `/projects/${projectId}/member-candidates`
}

export const projectMemberApi = {
  list: (projectId: string) => request<ProjectMember[]>(membersPath(projectId)),

  candidates: (projectId: string, search: string) =>
    request<Page<MemberCandidate>>(
      `${memberCandidatesPath(projectId)}${toQueryString({ search: search.trim(), pageSize: SEARCH_LIMIT })}`,
    ),

  add: (projectId: string, body: MemberCreateRequest) =>
    request<ProjectMember>(membersPath(projectId), { method: 'POST', body }),

  update: (projectId: string, memberId: string, body: MemberUpdateRequest) =>
    request<ProjectMember>(`${membersPath(projectId)}/${memberId}`, { method: 'PATCH', body }),

  remove: (projectId: string, memberId: string) =>
    request<void>(`${membersPath(projectId)}/${memberId}`, { method: 'DELETE' }),
}

// Sumber pilihan SearchSelect untuk formulir tambah anggota. /users khusus
// admin, jadi pencarian lewat endpoint kandidat milik modul proyek. Kandidat
// tidak punya endpoint satuan, jadi tanpa fetchSelected: label pilihan diambil
// dari hasil pencarian yang dipilih.
export function memberCandidateOptions(projectId: string) {
  return {
    cacheKey: memberCandidatesPath(projectId),
    fetchOptions: (search: string): Promise<SearchOption[]> =>
      projectMemberApi
        .candidates(projectId, search)
        .then((page) => page.items.map((candidate) => ({ value: candidate.id, label: candidate.name, hint: candidate.email }))),
  }
}
