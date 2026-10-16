import { describe, expect, it } from 'vitest'

import {
  dailyReportFormValues,
  emptyDailyReportForm,
  recentSummary,
  toDailyReportRequest,
  weatherSummary,
  type DailyReport,
} from '../models/dailyReport'

const REPORT: DailyReport = {
  id: 'laporan',
  projectId: 'proyek',
  reportDate: '2026-10-16T00:00:00Z',
  weatherMorning: 'cloudy',
  weatherAfternoon: 'heavy_rain',
  workers: 64,
  equipment: '1 tower crane',
  workDone: 'Pengecoran pelat lantai 4',
  issues: 'Hujan lebat sejak 13.30',
  createdBy: null,
  createdAt: '2026-10-16T10:00:00Z',
  updatedAt: '2026-10-16T10:00:00Z',
}

describe('cuaca laporan harian', () => {
  it('cukup satu kata kalau pagi dan siang sama', () => {
    expect(weatherSummary({ weatherMorning: 'sunny', weatherAfternoon: 'sunny' })).toBe('Cerah')
  })

  it('menyebut cuaca siang kalau berbeda dari pagi', () => {
    expect(weatherSummary(REPORT)).toBe('Berawan, siang hujan lebat')
  })
})

describe('form laporan harian', () => {
  it('mengirim tanggal sebagai RFC 3339 dan memangkas teks', () => {
    const request = toDailyReportRequest({
      ...emptyDailyReportForm(),
      reportDate: '2026-10-16',
      workers: '42',
      workDone: '  Pengecoran kolom  ',
      equipment: ' 1 molen ',
    })
    expect(request.reportDate).toBe('2026-10-16T00:00:00Z')
    expect(request.workers).toBe(42)
    expect(request.workDone).toBe('Pengecoran kolom')
    expect(request.equipment).toBe('1 molen')
    expect(request.issues).toBe('')
  })

  it('memuat laporan yang ada kembali ke form tanpa jam', () => {
    const values = dailyReportFormValues(REPORT)
    expect(values.reportDate).toBe('2026-10-16')
    expect(values.workers).toBe('64')
    expect(values.weatherAfternoon).toBe('heavy_rain')
  })

  it('mengisi tanggal hari ini sebagai bawaan', () => {
    expect(emptyDailyReportForm().reportDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })
})

describe('ringkasan 30 hari', () => {
  it('menyebut jumlah laporan dan hari hujan', () => {
    expect(recentSummary({ reportsLast30: 14, rainDaysLast30: 5 })).toBe('14 laporan dalam 30 hari terakhir, 5 hari hujan')
    expect(recentSummary({ reportsLast30: 3, rainDaysLast30: 0 })).toBe('3 laporan dalam 30 hari terakhir, tanpa hari hujan')
    expect(recentSummary({ reportsLast30: 0, rainDaysLast30: 0 })).toBe('Belum ada laporan dalam 30 hari terakhir')
  })
})
