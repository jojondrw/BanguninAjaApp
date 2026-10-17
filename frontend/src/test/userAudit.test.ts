import { describe, expect, it } from 'vitest'

import { auditSentence, type UserAuditEntry } from '../models/user'

const ENTRY: UserAuditEntry = {
  id: 'a-1', actorName: 'Budi Santoso', userName: 'Rina Wulandari', userEmail: 'rina@contoh.id',
  action: 'role', oldValue: 'penjualan', newValue: 'pengadaan', createdAt: '2026-10-17T07:00:00Z',
}

describe('jejak audit pengguna', () => {
  it('menulis perubahan peran dengan nama peran, bukan kodenya', () => {
    expect(auditSentence(ENTRY)).toBe('Budi Santoso mengubah peran Rina Wulandari dari Penjualan menjadi Pengadaan')
  })

  it('membedakan menonaktifkan dan mengaktifkan lagi', () => {
    expect(auditSentence({ ...ENTRY, action: 'active', oldValue: 'true', newValue: 'false' })).toBe(
      'Budi Santoso menonaktifkan akun Rina Wulandari',
    )
    expect(auditSentence({ ...ENTRY, action: 'active', oldValue: 'false', newValue: 'true' })).toBe(
      'Budi Santoso mengaktifkan lagi akun Rina Wulandari',
    )
  })

  it('menyebut sistem dan pendaftar baru dengan jelas', () => {
    expect(auditSentence({ ...ENTRY, actorName: null, oldValue: '', newValue: 'admin' })).toBe(
      'Sistem mengubah peran Rina Wulandari dari Tanpa peran menjadi Admin',
    )
  })
})
