import { panggilApi } from '../shared/apiClient'
import type { PermintaanDaftar, PermintaanMasuk, Pengguna, Sesi } from './auth'

export const authApi = {
  daftar: (permintaan: PermintaanDaftar) =>
    panggilApi<Pengguna>('/auth/register', { method: 'POST', body: permintaan }),

  masuk: (permintaan: PermintaanMasuk) =>
    panggilApi<Sesi>('/auth/login', { method: 'POST', body: permintaan, tanpaPembaruanToken: true }),

  keluar: () => panggilApi<void>('/auth/logout', { method: 'POST', tanpaPembaruanToken: true }),

  profil: () => panggilApi<Pengguna>('/auth/me'),
}
