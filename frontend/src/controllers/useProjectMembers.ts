import { useMutation, useQuery, useQueryClient, type QueryClient, type QueryKey } from '@tanstack/react-query'

import type { ProjectFilter } from '../models/project'
import {
  toMemberCreateRequest,
  toMemberUpdateRequest,
  type MemberEditValues,
  type MemberFormValues,
} from '../models/projectMember'
import { memberCandidatesPath, projectMemberApi } from '../models/projectMemberApi'
import { DATA_QUERY, PROJECTS_KEY } from './useErp'

// Di bawah [PROJECTS_KEY, 'detail', id] seperti tahap dan izin, jadi
// membatalkan PROJECTS_KEY ikut menyegarkan tim.
function membersKey(projectId: string): QueryKey {
  return [PROJECTS_KEY, 'detail', projectId, 'members']
}

function isMyProjectList(key: QueryKey): boolean {
  const filter = key[1] as ProjectFilter | undefined
  return key[0] === PROJECTS_KEY && typeof filter === 'object' && filter !== null && filter.mine === 'true'
}

// Setelah tim berubah: daftar anggota, kandidat di kotak cari (kunci milik
// SearchSelect, ['search-select', cacheKey, kata]), dan daftar "Proyek saya"
// diambil ulang. Daftar proyek lain tidak terpengaruh keanggotaan.
function refreshTeam(queryClient: QueryClient, projectId: string) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: membersKey(projectId) }),
    queryClient.invalidateQueries({ queryKey: ['search-select', memberCandidatesPath(projectId)] }),
    queryClient.invalidateQueries({ queryKey: [PROJECTS_KEY], predicate: (query) => isMyProjectList(query.queryKey) }),
  ])
}

export function useProjectMembers(projectId: string) {
  return useQuery({
    queryKey: membersKey(projectId),
    queryFn: () => projectMemberApi.list(projectId),
    ...DATA_QUERY,
  })
}

export function useAddProjectMember(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: MemberFormValues) => projectMemberApi.add(projectId, toMemberCreateRequest(values)),
    onSuccess: () => refreshTeam(queryClient, projectId),
  })
}

// Mengubah posisi atau peran tidak mengubah siapa yang ada di tim, jadi cukup
// daftar anggota yang diambil ulang.
export function useUpdateProjectMember(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ memberId, values }: { memberId: string; values: MemberEditValues }) =>
      projectMemberApi.update(projectId, memberId, toMemberUpdateRequest(values)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: membersKey(projectId) }),
  })
}

export function useRemoveProjectMember(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (memberId: string) => projectMemberApi.remove(projectId, memberId),
    onSuccess: () => refreshTeam(queryClient, projectId),
  })
}
