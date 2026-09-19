import { create } from 'zustand'

import type { Pengguna } from '../models/auth'

interface SesiState {
  accessToken: string | null
  pengguna: Pengguna | null
  sudahDipulihkan: boolean
  simpanSesi: (accessToken: string, pengguna: Pengguna) => void
  hapusSesi: () => void
  tandaiSudahDipulihkan: () => void
}

export const useSesiStore = create<SesiState>((set) => ({
  accessToken: null,
  pengguna: null,
  sudahDipulihkan: false,
  simpanSesi: (accessToken, pengguna) => set({ accessToken, pengguna }),
  hapusSesi: () => set({ accessToken: null, pengguna: null }),
  tandaiSudahDipulihkan: () => set({ sudahDipulihkan: true }),
}))

export const ambilAccessToken = () => useSesiStore.getState().accessToken

export const simpanSesiLangsung = useSesiStore.getState().simpanSesi

export const hapusSesiLangsung = useSesiStore.getState().hapusSesi
