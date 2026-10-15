import { monthLabel, number, shortDate } from '../shared/format'
import { currentPeriod } from '../shared/localDate'
import type { ModuleKey } from './auth'

// Jenis laporan sama dengan katalog backend (internal/reporting/catalog.go).
export type ReportType = 'arus_kas' | 'anggaran' | 'proyek' | 'piutang_utang' | 'stok' | 'penggajian' | 'penjualan'

export type ReportFormat = 'pdf' | 'xlsx' | 'csv'

export type ReportColumnKind = 'text' | 'money' | 'integer' | 'decimal' | 'percent' | 'date'

export type ReportValue = string | number | null

export interface ReportTypeInfo {
  value: ReportType
  label: string
  // Modul yang datanya dibaca laporan ini. Sama dengan pemetaan di backend
  // (reportKinds di catalog.go): laporan hanya tersedia kalau modul ini boleh dibaca.
  module: ModuleKey
  description: string
  usesPeriod: boolean
  // Laporan bulanan memakai bulan berjalan kalau periode dikosongkan.
  monthly: boolean
  periodHint: string
}

export const REPORT_TYPES: ReportTypeInfo[] = [
  {
    value: 'arus_kas',
    label: 'Arus kas',
    module: 'finance',
    description: 'Kas masuk, kas keluar, dan saldo per bulan beserta semua transaksi kas di periode itu.',
    usesPeriod: true,
    monthly: true,
    periodHint: 'Kosongkan untuk memakai bulan berjalan. Saldo awal dihitung dari transaksi sebelum tanggal awal.',
  },
  {
    value: 'anggaran',
    label: 'Anggaran dan realisasi',
    module: 'finance',
    description: 'Anggaran per proyek per tahun dibanding kas keluar, sisa, dan persentase serapan.',
    usesPeriod: true,
    monthly: false,
    periodHint: 'Anggaran dipilih menurut tahunnya. Kosongkan untuk semua tahun.',
  },
  {
    value: 'proyek',
    label: 'Proyek',
    module: 'projects',
    description: 'Daftar proyek dengan status, progres, jadwal, nilai kontrak, dan total RAB.',
    usesPeriod: true,
    monthly: false,
    periodHint: 'Proyek yang jadwalnya beririsan dengan periode ini. Kosongkan untuk semua proyek.',
  },
  {
    value: 'piutang_utang',
    label: 'Umur piutang dan utang',
    module: 'billing',
    description: 'Piutang dan utang yang belum lunas, dikelompokkan 0–30, 31–60, 61–90, dan lebih dari 90 hari.',
    usesPeriod: true,
    monthly: false,
    periodHint: 'Menyaring tanggal jatuh tempo. Umur dihitung sampai hari ini.',
  },
  {
    value: 'stok',
    label: 'Stok material',
    module: 'inventory',
    description: 'Stok per gudang saat ini dan material yang di bawah stok minimum.',
    usesPeriod: false,
    monthly: false,
    periodHint: 'Stok selalu memakai posisi hari ini, jadi tidak ada periode.',
  },
  {
    value: 'penggajian',
    label: 'Penggajian',
    module: 'hr',
    description: 'Ringkasan gaji per periode dan rincian slip gaji setiap karyawan.',
    usesPeriod: true,
    monthly: true,
    periodHint: 'Periode gaji bulanan yang tercakup. Kosongkan untuk bulan berjalan.',
  },
  {
    value: 'penjualan',
    label: 'Penjualan',
    module: 'sales',
    description: 'Kontrak penjualan dengan pembayaran, serta status cicilan yang jatuh tempo di periode itu.',
    usesPeriod: true,
    monthly: false,
    periodHint: 'Kontrak yang ditandatangani dan cicilan yang jatuh tempo di periode ini.',
  },
]

// Jenis laporan yang boleh dibuka peran ini, urutannya tetap seperti katalog.
export function readableReportTypes(canRead: (module: ModuleKey) => boolean): ReportTypeInfo[] {
  return REPORT_TYPES.filter((type) => canRead(type.module))
}

export const REPORT_TYPE_LABEL: Record<ReportType, string> = Object.fromEntries(
  REPORT_TYPES.map((type) => [type.value, type.label]),
) as Record<ReportType, string>

export const REPORT_FORMATS: { value: ReportFormat; label: string }[] = [
  { value: 'pdf', label: 'PDF' },
  { value: 'xlsx', label: 'Excel (XLSX)' },
  { value: 'csv', label: 'CSV' },
]

export const REPORT_FORMAT_LABEL: Record<ReportFormat, string> = { pdf: 'PDF', xlsx: 'Excel', csv: 'CSV' }

export const REPORT_NAME_MAX_LENGTH = 160

export function reportTypeInfo(type: string): ReportTypeInfo | undefined {
  return REPORT_TYPES.find((info) => info.value === type)
}

export function reportTypeLabel(type: string): string {
  return reportTypeInfo(type)?.label ?? type
}

export interface Report {
  id: string
  name: string
  type: string
  scope: string
  projectId: string | null
  projectCode: string
  projectName: string
  periodStart: string | null
  periodEnd: string | null
  format: ReportFormat
  createdBy: string
  createdByName: string
  createdAt: string
  updatedAt: string
}

export interface ReportFilter {
  type?: string
  page?: number
  pageSize?: number
}

export interface ReportRequest {
  name: string
  type: ReportType
  scope: string
  projectId: string | null
  periodStart: string | null
  periodEnd: string | null
  format: ReportFormat
}

export interface ReportCriteria {
  type: ReportType
  projectId: string
  periodStart: string
  periodEnd: string
}

export interface ReportColumn {
  header: string
  kind: ReportColumnKind
}

export interface ReportPreviewSection {
  title: string
  columns: ReportColumn[]
  rows: ReportValue[][]
  total?: ReportValue[]
  totalRows: number
}

export interface ReportPreview {
  type: ReportType
  title: string
  filters: string
  periodStart: string | null
  periodEnd: string | null
  generatedAt: string
  sections: ReportPreviewSection[]
}

export interface ReportRow {
  values: ReportValue[]
  isTotal: boolean
}

export function reportRows(section: ReportPreviewSection): ReportRow[] {
  const rows = section.rows.map((values) => ({ values, isTotal: false }))
  return section.total ? [...rows, { values: section.total, isTotal: true }] : rows
}

export function formatReportValue(kind: ReportColumnKind, value: ReportValue | undefined): string {
  if (value === null || value === undefined || value === '') {
    return ''
  }
  if (typeof value === 'string') {
    return kind === 'date' ? shortDate(value) : value
  }
  return kind === 'percent' ? `${number(value)}%` : number(value)
}

export function isNumericColumn(kind: ReportColumnKind): boolean {
  return kind === 'money' || kind === 'integer' || kind === 'decimal' || kind === 'percent'
}

export function reportPeriodLabel(report: Pick<Report, 'periodStart' | 'periodEnd'>): string {
  if (!report.periodStart && !report.periodEnd) {
    return 'Semua periode'
  }
  return `${shortDate(report.periodStart)} – ${shortDate(report.periodEnd)}`
}

export function createdAtLabel(iso: string): string {
  return new Date(iso).toLocaleString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export interface ReportFile {
  blob: Blob
  fileName: string
}

export interface ReportFormValues extends ReportCriteria {
  name: string
  format: ReportFormat
}

// Nama bawaan kalau isian nama dikosongkan, misalnya "Arus kas Okt 2026".
export function defaultReportName(type: ReportType, periodStart: string, periodEnd: string): string {
  const label = REPORT_TYPE_LABEL[type]
  if (periodStart && periodEnd) {
    return `${label} ${shortDate(periodStart)} – ${shortDate(periodEnd)}`
  }
  if (periodStart) {
    return `${label} sejak ${shortDate(periodStart)}`
  }
  if (periodEnd) {
    return `${label} sampai ${shortDate(periodEnd)}`
  }
  return reportTypeInfo(type)?.monthly ? `${label} ${monthLabel(currentPeriod())}` : label
}

export function periodIsOrdered(periodStart: string, periodEnd: string): boolean {
  return !periodStart || !periodEnd || periodStart <= periodEnd
}

export function fallbackFileName(report: Pick<Report, 'type' | 'format'>): string {
  return `laporan-${report.type.replace(/_/g, '-')}.${report.format}`
}
