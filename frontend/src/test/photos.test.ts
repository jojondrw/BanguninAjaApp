import { describe, expect, it } from 'vitest'

import { PHOTOS, greeting } from '../shared/photos'

describe('salam sesuai jam', () => {
  it('memakai pagi, siang, sore, dan malam pada batas jamnya', () => {
    expect([4, 10, 11, 14, 15, 17, 18, 23].map(greeting)).toEqual([
      'Selamat pagi',
      'Selamat pagi',
      'Selamat siang',
      'Selamat siang',
      'Selamat sore',
      'Selamat sore',
      'Selamat malam',
      'Selamat malam',
    ])
  })

  it('menganggap dini hari sebelum jam 4 masih malam', () => {
    expect(greeting(0)).toBe('Selamat malam')
    expect(greeting(3)).toBe('Selamat malam')
  })
})

describe('foto stok', () => {
  it('memakai berkas WebP lokal, bukan tautan ke luar', () => {
    expect(Object.values(PHOTOS).every((photo) => /^\/images\/[a-z]+\.webp$/.test(photo.src))).toBe(true)
  })
})
