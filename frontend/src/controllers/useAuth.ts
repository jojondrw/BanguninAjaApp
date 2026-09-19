import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

import { authApi } from '../models/authApi'
import type { LoginRequest, RegisterRequest } from '../models/auth'
import { refreshSession } from '../shared/apiClient'
import { useSessionStore } from './sessionStore'

export function useSession() {
  const accessToken = useSessionStore((state) => state.accessToken)
  const user = useSessionStore((state) => state.user)
  const isRestored = useSessionStore((state) => state.isRestored)

  return { accessToken, user, isRestored }
}

export function useSessionRestore() {
  const markRestored = useSessionStore((state) => state.markRestored)
  const isRestored = useSessionStore((state) => state.isRestored)

  useEffect(() => {
    if (isRestored) {
      return
    }

    refreshSession().finally(markRestored)
  }, [isRestored, markRestored])

  return isRestored
}

export function useLogin() {
  const setSession = useSessionStore((state) => state.setSession)

  return useMutation({
    mutationFn: (payload: LoginRequest) => authApi.login(payload),
    onSuccess: (session) => setSession(session.accessToken, session.user),
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: (payload: RegisterRequest) => authApi.register(payload),
  })
}

export function useLogout() {
  const clearSession = useSessionStore((state) => state.clearSession)
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => authApi.logout(),
    onSettled: () => {
      clearSession()
      queryClient.clear()
    },
  })
}

export function useProfile() {
  const accessToken = useSessionStore((state) => state.accessToken)

  return useQuery({
    queryKey: ['profile'],
    queryFn: () => authApi.profile(),
    enabled: accessToken !== null,
    staleTime: 60_000,
  })
}
