import { create } from 'zustand'

import type { User } from '../models/auth'

interface SessionState {
  accessToken: string | null
  user: User | null
  isRestored: boolean
  setSession: (accessToken: string, user: User) => void
  clearSession: () => void
  markRestored: () => void
}

export const useSessionStore = create<SessionState>((set) => ({
  accessToken: null,
  user: null,
  isRestored: false,
  setSession: (accessToken, user) => set({ accessToken, user }),
  clearSession: () => set({ accessToken: null, user: null }),
  markRestored: () => set({ isRestored: true }),
}))

export const getAccessToken = () => useSessionStore.getState().accessToken

export const setSession = useSessionStore.getState().setSession

export const clearSession = useSessionStore.getState().clearSession
