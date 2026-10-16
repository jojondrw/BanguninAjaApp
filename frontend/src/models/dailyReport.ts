import { toApiDate, todayDate } from '../shared/localDate'
import type { Page } from './common'

export type Weather = 'sunny' | 'cloudy' | 'drizzle' | 'rain' | 'heavy_rain'

export const WEATHER_OPTIONS: readonly Weather[] = ['sunny', 'cloudy', 'drizzle', 'rain', 'heavy_rain']

export const WEATHER_LABEL: Record<Weather, string> = {
  sunny: 'Cerah',
  cloudy: 'Berawan',
  drizzle: 'Gerimis',
  rain: 'Hujan',
  heavy_rain: 'Hujan lebat',
}

// Batas mengikuti tag binding DailyReportRequest di backend.
export const DAILY_REPORT_EQUIPMENT_MAX_LENGTH = 300
export const DAILY_REPORT_TEXT_MAX_LENGTH = 2000
export const DAILY_REPORT_WORKERS_MAX = 10_000

export interface DailyReport {
  id: string
  projectId: string
  reportDate: string
  weatherMorning: Weather
  weatherAfternoon: Weather
  workers: number
  equipment: string
  workDone: string
  issues: string
  createdBy: string | null
  createdAt: string
  updatedAt: string
}

export interface DailyReportPage extends Page<DailyReport> {
  reportsLast30: number
  rainDaysLast30: number
}

export interface DailyReportRequest {
  reportDate: string
  weatherMorning: Weather
  weatherAfternoon: Weather
  workers: number
  equipment: string
  workDone: string
  issues: string
}

export interface DailyReportFormValues {
  reportDate: string
  weatherMorning: Weather
  weatherAfternoon: Weather
  workers: string
  equipment: string
  workDone: string
  issues: string
}

export function emptyDailyReportForm(): DailyReportFormValues {
  return {
    reportDate: todayDate(),
    weatherMorning: 'sunny',
    weatherAfternoon: 'sunny',
    workers: '',
    equipment: '',
    workDone: '',
    issues: '',
  }
}

export function dailyReportFormValues(report: DailyReport): DailyReportFormValues {
  return {
    reportDate: report.reportDate.slice(0, 10),
    weatherMorning: report.weatherMorning,
    weatherAfternoon: report.weatherAfternoon,
    workers: String(report.workers),
    equipment: report.equipment,
    workDone: report.workDone,
    issues: report.issues,
  }
}

export function toDailyReportRequest(values: DailyReportFormValues): DailyReportRequest {
  return {
    reportDate: toApiDate(values.reportDate),
    weatherMorning: values.weatherMorning,
    weatherAfternoon: values.weatherAfternoon,
    workers: values.workers === '' ? 0 : Number(values.workers),
    equipment: values.equipment.trim(),
    workDone: values.workDone.trim(),
    issues: values.issues.trim(),
  }
}

// Satu kata kalau pagi dan siang sama, supaya kolom cuaca tetap ringkas.
export function weatherSummary(report: Pick<DailyReport, 'weatherMorning' | 'weatherAfternoon'>): string {
  if (report.weatherMorning === report.weatherAfternoon) {
    return WEATHER_LABEL[report.weatherMorning]
  }
  return `${WEATHER_LABEL[report.weatherMorning]}, siang ${WEATHER_LABEL[report.weatherAfternoon].toLowerCase()}`
}

export function isRainy(weather: Weather): boolean {
  return weather === 'rain' || weather === 'heavy_rain'
}

export function recentSummary(page: Pick<DailyReportPage, 'reportsLast30' | 'rainDaysLast30'>): string {
  if (page.reportsLast30 === 0) {
    return 'Belum ada laporan dalam 30 hari terakhir'
  }
  const rain = page.rainDaysLast30 === 0 ? 'tanpa hari hujan' : `${page.rainDaysLast30} hari hujan`
  return `${page.reportsLast30} laporan dalam 30 hari terakhir, ${rain}`
}
