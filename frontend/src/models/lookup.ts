// Satu pilihan pada kotak cari-dan-pilih. `hint` tampil kecil di kanan, misalnya kode.
export interface SearchOption {
  value: string
  label: string
  hint?: string
}

// Sumber pilihan dari server: pencarian per halaman kecil plus pengambilan satu
// data lewat id, supaya label pilihan yang tersimpan tetap tampil walau tidak
// masuk hasil pencarian.
export interface OptionSource<O extends SearchOption = SearchOption> {
  cacheKey: string
  fetchOptions: (search: string) => Promise<O[]>
  fetchSelected: (value: string) => Promise<O>
}
