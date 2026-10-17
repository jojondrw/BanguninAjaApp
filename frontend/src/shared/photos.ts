// Foto stok berlisensi Unsplash License (bebas dipakai tanpa kredit), sudah
// dikecilkan ke WebP di public/images.

export interface StockPhoto {
  src: string
  alt: string
}

export const PHOTOS = {
  lokasi: { src: '/images/lokasi.webp', alt: 'Jakarta dari udara pada malam hari' },
  konstruksi: { src: '/images/konstruksi.webp', alt: 'Gedung bertingkat yang sedang dibangun dengan tower crane' },
  perumahan: { src: '/images/perumahan.webp', alt: 'Kawasan perumahan dari udara' },
} satisfies Record<string, StockPhoto>

export function greeting(hour: number): string {
  if (hour >= 4 && hour < 11) {
    return 'Selamat pagi'
  }
  if (hour >= 11 && hour < 15) {
    return 'Selamat siang'
  }
  if (hour >= 15 && hour < 18) {
    return 'Selamat sore'
  }
  return 'Selamat malam'
}
