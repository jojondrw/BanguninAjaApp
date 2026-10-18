// Foto kondisi lapangan terkini. Foto terbaru menjadi sampul proyek di daftar
// proyek dan Ringkasan. Batas dan format sama dengan validasi backend.

export const SITE_PHOTO_MAX_BYTES = 3 * 1024 * 1024
export const SITE_PHOTO_TYPES = ['image/jpeg', 'image/png', 'image/webp']
export const SITE_PHOTO_CAPTION_MAX_LENGTH = 160

export interface SitePhoto {
  id: string
  projectId: string
  takenOn: string
  caption: string
  contentType: string
  size: number
  createdBy: string | null
  createdAt: string
}

export interface SitePhotoUpload {
  file: File
  caption: string
  takenOn: string
}

export function sitePhotoProblem(file: File | null): string | null {
  if (!file) {
    return 'Pilih foto yang akan diunggah.'
  }
  if (!SITE_PHOTO_TYPES.includes(file.type)) {
    return 'Foto harus berformat JPG, PNG, atau WebP.'
  }
  if (file.size > SITE_PHOTO_MAX_BYTES) {
    return 'Ukuran foto paling besar 3 MB.'
  }
  return null
}
