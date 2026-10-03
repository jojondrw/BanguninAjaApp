import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import type { ManagedUser, RoleCode, UserFilter } from '../models/user'
import { userApi } from '../models/userApi'
import { PROFILE_KEY } from './useAuth'

const USERS_KEY = 'users'

export function useUsers(filter: UserFilter) {
  return useQuery({
    queryKey: [USERS_KEY, filter],
    queryFn: () => userApi.list(filter),
    staleTime: 30_000,
    retry: 0,
    placeholderData: keepPreviousData,
  })
}

// Profil ikut disegarkan karena admin bisa mengubah peran atau status akunnya
// sendiri, dan menu harus langsung menyesuaikan.
function useUserMutation<TVariables>(mutationFn: (variables: TVariables) => Promise<ManagedUser>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [USERS_KEY] })
      void queryClient.invalidateQueries({ queryKey: [PROFILE_KEY] })
    },
  })
}

export function useChangeUserRole() {
  return useUserMutation(({ user, role }: { user: ManagedUser; role: RoleCode }) => userApi.changeRole(user.id, role))
}

export function useChangeUserActive() {
  return useUserMutation(({ user, active }: { user: ManagedUser; active: boolean }) =>
    userApi.changeActive(user.id, active),
  )
}
