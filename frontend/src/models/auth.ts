export interface Pengguna {
  id: string
  name: string
  email: string
  createdAt: string
}

export interface Sesi {
  accessToken: string
  accessTokenExpiresAt: string
  user: Pengguna
}

export interface PermintaanMasuk {
  email: string
  password: string
}

export interface PermintaanDaftar {
  name: string
  email: string
  password: string
}

export interface KesalahanApi {
  code: string
  message: string
}
