import { describe, expect, it } from 'vitest'

import { seesEveryProject } from '../models/project'

describe('hak akses per proyek', () => {
  it('membiarkan Admin, Pengamat, dan Keuangan melihat semua proyek', () => {
    expect(['admin', 'pengamat', 'keuangan'].every(seesEveryProject)).toBe(true)
  })

  it('membatasi peran lain dan pengguna tanpa peran ke proyek timnya', () => {
    expect(['manajer_proyek', 'pengadaan', 'penjualan', 'sdm'].some(seesEveryProject)).toBe(false)
    expect(seesEveryProject(undefined)).toBe(false)
  })
})
