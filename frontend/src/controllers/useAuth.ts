import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

import { authApi } from '../models/authApi'
import type { PermintaanDaftar, PermintaanMasuk } from '../models/auth'
import { perbaruiSesi } from '../shared/apiClient'
import { useSesiStore } from './sesiStore'

export function useSesi() {
  const accessToken = useSesiStore((state) => state.accessToken)
  const pengguna = useSesiStore((state) => state.pengguna)
  const sudahDipulihkan = useSesiStore((state) => state.sudahDipulihkan)

  return { accessToken, pengguna, sudahDipulihkan }
}

export function usePemulihanSesi() {
  const tandaiSudahDipulihkan = useSesiStore((state) => state.tandaiSudahDipulihkan)
  const sudahDipulihkan = useSesiStore((state) => state.sudahDipulihkan)

  useEffect(() => {
    if (sudahDipulihkan) {
      return
    }

    perbaruiSesi().finally(tandaiSudahDipulihkan)
  }, [sudahDipulihkan, tandaiSudahDipulihkan])

  return sudahDipulihkan
}

export function useMasuk() {
  const simpanSesi = useSesiStore((state) => state.simpanSesi)

  return useMutation({
    mutationFn: (permintaan: PermintaanMasuk) => authApi.masuk(permintaan),
    onSuccess: (sesi) => simpanSesi(sesi.accessToken, sesi.user),
  })
}

export function useDaftar() {
  return useMutation({
    mutationFn: (permintaan: PermintaanDaftar) => authApi.daftar(permintaan),
  })
}

export function useKeluar() {
  const hapusSesi = useSesiStore((state) => state.hapusSesi)
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => authApi.keluar(),
    onSettled: () => {
      hapusSesi()
      queryClient.clear()
    },
  })
}

export function useProfil() {
  const accessToken = useSesiStore((state) => state.accessToken)

  return useQuery({
    queryKey: ['profil'],
    queryFn: () => authApi.profil(),
    enabled: accessToken !== null,
    staleTime: 60_000,
  })
}
