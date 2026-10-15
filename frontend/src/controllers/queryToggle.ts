// Pilihan tambahan untuk hook baca yang juga dipakai di halaman lintas modul
// seperti Ringkasan: enabled false menahan permintaan ke server, misalnya kalau
// peran tidak boleh membaca modulnya. Bawaannya true, jadi pemanggil lain tidak berubah.
export interface QueryToggle {
  enabled?: boolean
}
