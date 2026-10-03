import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

import { authApi } from '../models/authApi'
import type { LoginRequest, RegisterRequest, User } from '../models/auth'
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

export const PROFILE_KEY = 'profile'

export function useProfile() {
  const accessToken = useSessionStore((state) => state.accessToken)
  const userId = useSessionStore((state) => state.user?.id ?? null)

  // Id user ikut di kunci supaya profil akun sebelumnya tidak terbawa
  // ketika akun lain masuk tanpa keluar lebih dulu.
  return useQuery({
    queryKey: [PROFILE_KEY, userId],
    queryFn: () => authApi.profile(),
    enabled: accessToken !== null,
    staleTime: 60_000,
  })
}

// Profil terbaru dari /auth/me kalau sudah dimuat, supaya perubahan peran oleh
// admin terasa tanpa harus keluar-masuk. Sebelum itu pakai user dari sesi.
export function useCurrentUser(): User | null {
  const sessionUser = useSessionStore((state) => state.user)
  const profile = useProfile()
  return profile.data ?? sessionUser
}
