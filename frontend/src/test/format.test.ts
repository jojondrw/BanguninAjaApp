import { afterEach, describe, expect, it, vi } from 'vitest'

import { dateRange, kpiValue, monthLabel, number, rupiah, rupiahShort, shortDate } from '../shared/format'
import { currentPeriod, firstDayOfMonth, toApiDate, todayDate } from '../shared/localDate'

// Intl memakai spasi tak terputus di antara "Rp" dan angka.
const plain = (text: string) => text.replace(/ /g, ' ')

describe('rupiah', () => {
  it('memakai titik ribuan tanpa desimal', () => {
    expect(plain(rupiah(1_500_000))).toBe('Rp 1.500.000')
    expect(plain(rupiah(0))).toBe('Rp 0')
  })

  it('membulatkan sen', () => {
    expect(plain(rupiah(1234.6))).toBe('Rp 1.235')
  })
})

describe('rupiahShort', () => {
  it('menyingkat miliar dengan satu desimal', () => {
    expect(plain(rupiahShort(2_460_000_000))).toBe('Rp 2,5 M')
    expect(plain(rupiahShort(1_000_000_000))).toBe('Rp 1 M')
  })

  it('menyingkat juta tanpa desimal', () => {
    expect(plain(rupiahShort(12_400_000))).toBe('Rp 12 jt')
  })

  it('memberi tanda minus di depan untuk nilai negatif', () => {
    expect(plain(rupiahShort(-12_400_000))).toBe('−Rp 12 jt')
    expect(plain(rupiahShort(-3_000_000_000))).toBe('−Rp 3 M')
  })

  it('menulis lengkap nilai di bawah sejuta', () => {
    expect(plain(rupiahShort(750_000))).toBe('Rp 750.000')
  })
})

describe('angka dan tanggal', () => {
  it('number memakai pemisah ribuan Indonesia', () => {
    expect(number(1_234_567)).toBe('1.234.567')
  })

  it('shortDate menulis tanggal pendek dan strip untuk kosong', () => {
    expect(shortDate('2026-10-03T05:00:00Z')).toBe('3 Okt 2026')
    expect(shortDate(null)).toBe('-')
  })

  it('dateRange menjelaskan jadwal yang belum diisi', () => {
    expect(dateRange(null, null)).toBe('Jadwal belum diisi')
    expect(dateRange('2026-01-05T05:00:00Z', null)).toBe('5 Jan 2026 sampai -')
  })

  it('monthLabel mengubah periode YYYY-MM jadi bulan pendek', () => {
    expect(monthLabel('2026-10')).toBe('Okt 2026')
    expect(monthLabel('2026')).toBe('2026')
  })

  it('kpiValue tidak menampilkan angka selama memuat atau gagal', () => {
    expect(kpiValue(true, false, '12')).toBe('...')
    expect(kpiValue(false, true, '12')).toBe('—')
    expect(kpiValue(false, false, '12')).toBe('12')
  })
})

describe('localDate', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('hari ini mengikuti jam lokal, bukan UTC', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 0, 5, 23, 30))
    expect(todayDate()).toBe('2026-01-05')
    expect(currentPeriod()).toBe('2026-01')
  })

  it('awal bulan dan format tanggal untuk backend', () => {
    expect(firstDayOfMonth('2026-10-17')).toBe('2026-10-01')
    expect(toApiDate('2026-10-03')).toBe('2026-10-03T00:00:00Z')
  })
})
