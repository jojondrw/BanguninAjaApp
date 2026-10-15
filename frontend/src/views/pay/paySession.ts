// Hash yang sudah dikirim disimpan per token di sessionStorage, supaya halaman
// yang dimuat ulang tetap melanjutkan pengecekan transaksi yang sama.
// Penyimpanan bisa tidak tersedia (mode privat), jadi semua akses dibungkus.
const PREFIX = 'bayar:'

export function readSubmittedHash(token: string): string | null {
  try {
    return window.sessionStorage.getItem(PREFIX + token)
  } catch {
    return null
  }
}

export function storeSubmittedHash(token: string, hash: string | null) {
  try {
    if (hash === null) {
      window.sessionStorage.removeItem(PREFIX + token)
      return
    }
    window.sessionStorage.setItem(PREFIX + token, hash)
  } catch {
    // Tanpa penyimpanan, pengecekan tetap jalan selama halaman terbuka.
  }
}
