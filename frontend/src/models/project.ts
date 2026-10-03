import type { Page } from './common'

export type ProjectStatus = 'planning' | 'ongoing' | 'on_hold' | 'completed' | 'cancelled'

export interface Project {
  id: string
  code: string
  name: string
  regionId: string | null
  type: string
  status: ProjectStatus
  progress: number
  startDate: string | null
  targetEndDate: string | null
  contractValue: number
  createdAt: string
  updatedAt: string
}

// Ringkasan seluruh proyek yang dihitung backend, bukan dari satu halaman daftar.
export interface ProjectSummary {
  count: number
  byStatus: Record<ProjectStatus, number>
  contractValue: number
  averageProgress: number
}

export interface ProjectFilter {
  search?: string
  status?: ProjectStatus
  page?: number
  pageSize?: number
}

export interface ProjectRequest {
  code: string
  name: string
  type: string
  regionId?: string
  startDate?: string
  targetEndDate?: string
  contractValue: number
}

// Batas panjang mengikuti tag binding ProjectRequest di backend.
export const PROJECT_CODE_MAX_LENGTH = 20
export const PROJECT_NAME_MAX_LENGTH = 160
export const PROJECT_TYPE_MAX_LENGTH = 40

export const PROJECT_TYPE_SUGGESTIONS = ['Perumahan', 'Apartemen', 'Komersial', 'Rumah Sakit', 'Campuran']

export interface ProjectFormValues {
  code: string
  name: string
  type: string
  contractValue: string
  startDate: string
  targetEndDate: string
}

export const EMPTY_PROJECT_FORM: ProjectFormValues = {
  code: '',
  name: '',
  type: '',
  contractValue: '',
  startDate: '',
  targetEndDate: '',
}

// Isian tanggal memberi "YYYY-MM-DD", sedangkan backend membaca time.Time yang
// butuh format RFC 3339. Tanggal kosong tidak dikirim sama sekali.
export function toApiDate(value: string): string | undefined {
  return value === '' ? undefined : `${value}T00:00:00Z`
}

export function toProjectRequest(values: ProjectFormValues): ProjectRequest {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    type: values.type.trim(),
    contractValue: values.contractValue === '' ? 0 : Number(values.contractValue),
    startDate: toApiDate(values.startDate),
    targetEndDate: toApiDate(values.targetEndDate),
  }
}

export const PROJECT_STATUS_LABEL: Record<ProjectStatus, string> = {
  planning: 'Perencanaan',
  ongoing: 'Berjalan',
  on_hold: 'Tertunda',
  completed: 'Selesai',
  cancelled: 'Batal',
}

export const PROJECT_STATUS_TONE: Record<ProjectStatus, string> = {
  planning: 'bg-slate-100 text-slate-700',
  ongoing: 'bg-blue-100 text-blue-800',
  on_hold: 'bg-amber-100 text-amber-800',
  completed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
}

// Profil bangunan yang paling cocok dengan jenis proyek, dipakai sebagai pilihan
// awal di formulir evaluasi lokasi. "Rumah Sakit" dicek sebelum "rumah".
const PROFILE_BY_PROJECT_TYPE: [RegExp, string][] = [
  [/rumah sakit|klinik/i, 'hospital'],
  [/perumahan|apartemen|hunian|rumah/i, 'housing'],
  [/mall|komersial|ruko|belanja/i, 'mall'],
  [/hiburan/i, 'entertainment'],
]

export function suggestedProfileCode(projectType: string): string | null {
  const match = PROFILE_BY_PROJECT_TYPE.find(([pattern]) => pattern.test(projectType))
  return match ? match[1] : null
}

export function isProjectClosed(project: Project): boolean {
  return project.status === 'completed' || project.status === 'cancelled'
}

export type PhaseStatus = 'not_started' | 'in_progress' | 'completed'

export interface ProjectPhase {
  id: string
  projectId: string
  name: string
  sortOrder: number
  startDate: string | null
  targetEndDate: string | null
  progress: number
  status: PhaseStatus
  createdAt: string
  updatedAt: string
}

export interface PhaseRequest {
  name: string
  sortOrder: number
  startDate?: string
  targetEndDate?: string
  progress: number
}

export interface PhaseFormValues {
  name: string
  startDate: string
  targetEndDate: string
  progress: string
}

export const EMPTY_PHASE_FORM: PhaseFormValues = {
  name: '',
  startDate: '',
  targetEndDate: '',
  progress: '0',
}

export const PHASE_NAME_MAX_LENGTH = 160

export const PHASE_STATUS_LABEL: Record<PhaseStatus, string> = {
  not_started: 'Belum mulai',
  in_progress: 'Berjalan',
  completed: 'Selesai',
}

export const PHASE_STATUS_TONE: Record<PhaseStatus, string> = {
  not_started: 'bg-slate-100 text-slate-700',
  in_progress: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
}

export function sortPhases(phases: ProjectPhase[]): ProjectPhase[] {
  return [...phases].sort((a, b) => a.sortOrder - b.sortOrder)
}

export function nextPhaseOrder(phases: ProjectPhase[]): number {
  return phases.reduce((highest, phase) => Math.max(highest, phase.sortOrder), 0) + 1
}

export function toPhaseRequest(values: PhaseFormValues, sortOrder: number): PhaseRequest {
  return {
    name: values.name.trim(),
    sortOrder,
    startDate: toApiDate(values.startDate),
    targetEndDate: toApiDate(values.targetEndDate),
    progress: values.progress === '' ? 0 : Number(values.progress),
  }
}

export interface PhaseSummary {
  total: number
  completed: number
  inProgress: number
}

export function summarizePhases(phases: ProjectPhase[]): PhaseSummary {
  return {
    total: phases.length,
    completed: phases.filter((phase) => phase.status === 'completed').length,
    inProgress: phases.filter((phase) => phase.status === 'in_progress').length,
  }
}

export type PermitStatus = 'submitted' | 'issued' | 'expired' | 'rejected'

export interface Permit {
  id: string
  projectId: string
  type: string
  number: string
  issuedDate: string | null
  validUntil: string | null
  status: PermitStatus
  createdAt: string
  updatedAt: string
}

export interface PermitRequest {
  type: string
  number: string
  issuedDate?: string
  validUntil?: string
  status: PermitStatus
}

export interface PermitFormValues {
  type: string
  number: string
  status: PermitStatus
  issuedDate: string
  validUntil: string
}

export const EMPTY_PERMIT_FORM: PermitFormValues = {
  type: '',
  number: '',
  status: 'submitted',
  issuedDate: '',
  validUntil: '',
}

// Batas panjang mengikuti tag binding PermitRequest di backend.
export const PERMIT_TYPE_MAX_LENGTH = 60
export const PERMIT_NUMBER_MAX_LENGTH = 80

export const PERMIT_TYPE_SUGGESTIONS = ['PBG', 'SLF', 'KKPR', 'AMDAL', 'UKL-UPL', 'Andalalin']

export const PERMIT_STATUS_LABEL: Record<PermitStatus, string> = {
  submitted: 'Diajukan',
  issued: 'Terbit',
  expired: 'Kedaluwarsa',
  rejected: 'Ditolak',
}

export const PERMIT_STATUS_TONE: Record<PermitStatus, string> = {
  submitted: 'bg-amber-100 text-amber-800',
  issued: 'bg-green-100 text-green-800',
  expired: 'bg-slate-100 text-slate-700',
  rejected: 'bg-red-100 text-red-800',
}

export function toPermitRequest(values: PermitFormValues): PermitRequest {
  return {
    type: values.type.trim(),
    number: values.number.trim(),
    status: values.status,
    issuedDate: toApiDate(values.issuedDate),
    validUntil: toApiDate(values.validUntil),
  }
}

export interface PermitSummary {
  total: number
  issued: number
  submitted: number
}

export function summarizePermits(permits: Permit[]): PermitSummary {
  return {
    total: permits.length,
    issued: permits.filter((permit) => permit.status === 'issued').length,
    submitted: permits.filter((permit) => permit.status === 'submitted').length,
  }
}

export interface BudgetItem {
  id: string
  projectId: string
  code: string
  description: string
  volume: number
  unitOfMeasureId: string
  unitPrice: number
  total: number
  createdAt: string
  updatedAt: string
}

export interface BudgetItemPage extends Page<BudgetItem> {
  grandTotal: number
}

// Perpindahan status yang diterima backend, disalin dari projectTransitions di
// internal/project/service.go. Selesai dan Batal adalah status akhir.
export const PROJECT_TRANSITIONS: Record<ProjectStatus, ProjectStatus[]> = {
  planning: ['ongoing', 'cancelled'],
  ongoing: ['on_hold', 'completed', 'cancelled'],
  on_hold: ['ongoing', 'cancelled'],
  completed: [],
  cancelled: [],
}

export const PROJECT_STATUS_DESCRIPTION: Record<ProjectStatus, string> = {
  planning: 'Proyek masih disiapkan. Tahap, izin, dan RAB disusun sebelum pekerjaan dimulai.',
  ongoing: 'Pekerjaan lapangan sedang berjalan. Progres diperbarui lewat tahap proyek.',
  on_hold: 'Pekerjaan dihentikan sementara dan bisa dilanjutkan lagi.',
  completed: 'Semua tahap sudah 100% dan proyek ditutup. Tahap dan RAB tidak bisa diubah lagi.',
  cancelled: 'Proyek dihentikan permanen. Tahap dan RAB tidak bisa diubah lagi.',
}

// Perpindahan yang tidak bisa dibalik (misalnya Perencanaan ke Berjalan, atau
// ke status akhir) perlu konfirmasi kedua sebelum dikirim.
export function isReversibleTransition(from: ProjectStatus, to: ProjectStatus): boolean {
  return PROJECT_TRANSITIONS[to].includes(from)
}

export function transitionLabel(from: ProjectStatus, to: ProjectStatus): string {
  switch (to) {
    case 'ongoing':
      return from === 'on_hold' ? 'Lanjutkan pengerjaan' : 'Mulai pengerjaan'
    case 'on_hold':
      return 'Tunda proyek'
    case 'completed':
      return 'Tandai selesai'
    case 'cancelled':
      return 'Batalkan proyek'
    default:
      return `Ubah ke ${PROJECT_STATUS_LABEL[to]}`
  }
}

// State navigasi setelah proyek dihapus, supaya /proyek bisa memberi tahu
// proyek mana yang baru saja hilang dari daftar.
export interface DeletedProjectState {
  deletedProject: string
}

export function deletedProjectName(state: unknown): string | null {
  if (typeof state === 'object' && state !== null && 'deletedProject' in state) {
    const name = (state as DeletedProjectState).deletedProject
    return typeof name === 'string' ? name : null
  }
  return null
}

// Kebalikan toApiDate: tanggal dari backend ("2026-11-01T00:00:00Z") menjadi
// nilai isian bertipe date.
export function fromApiDate(value: string | null): string {
  return value ? value.slice(0, 10) : ''
}

export function projectFormValues(project: Project): ProjectFormValues {
  return {
    code: project.code,
    name: project.name,
    type: project.type,
    contractValue: String(project.contractValue),
    startDate: fromApiDate(project.startDate),
    targetEndDate: fromApiDate(project.targetEndDate),
  }
}

// PUT /projects/:id menimpa semua kolom, termasuk wilayah. Wilayah yang tidak
// dikirim berarti dikosongkan, jadi wilayah lama selalu ikut dikirim.
export function toProjectUpdateRequest(values: ProjectFormValues, regionId: string): ProjectRequest {
  return { ...toProjectRequest(values), regionId: regionId === '' ? undefined : regionId }
}

export type RegionType = 'province' | 'city' | 'regency' | 'district'

export interface Region {
  id: string
  code: string
  name: string
  type: RegionType
  parentId: string | null
  createdAt: string
  updatedAt: string
}

export const REGION_TYPE_LABEL: Record<RegionType, string> = {
  province: 'Provinsi',
  city: 'Kota',
  regency: 'Kabupaten',
  district: 'Kecamatan',
}

export const REGION_SEARCH_MIN_LENGTH = 2

export function regionLabel(region: Region): string {
  return `${region.name} (${REGION_TYPE_LABEL[region.type] ?? region.type})`
}

export const PROGRESS_MAX = 100

export interface PhaseEditValues extends PhaseFormValues {
  sortOrder: string
}

export function phaseEditValues(phase: ProjectPhase): PhaseEditValues {
  return {
    name: phase.name,
    sortOrder: String(phase.sortOrder),
    startDate: fromApiDate(phase.startDate),
    targetEndDate: fromApiDate(phase.targetEndDate),
    progress: String(phase.progress),
  }
}

export function toPhaseEditRequest(values: PhaseEditValues): PhaseRequest {
  return toPhaseRequest(values, Number(values.sortOrder))
}

// Progres tahap wajib bilangan bulat 0 sampai 100 (binding min=0,max=100).
export function isValidProgress(value: string): boolean {
  const progress = Number(value)
  return value !== '' && Number.isInteger(progress) && progress >= 0 && progress <= PROGRESS_MAX
}

export function permitFormValues(permit: Permit): PermitFormValues {
  return {
    type: permit.type,
    number: permit.number,
    status: permit.status,
    issuedDate: fromApiDate(permit.issuedDate),
    validUntil: fromApiDate(permit.validUntil),
  }
}

export interface PermitIssueValues {
  number: string
  issuedDate: string
  validUntil: string
}

export function permitIssueValues(permit: Permit, today: string): PermitIssueValues {
  return {
    number: permit.number,
    issuedDate: fromApiDate(permit.issuedDate) || today,
    validUntil: fromApiDate(permit.validUntil),
  }
}

// Menandai izin terbit tetap lewat PUT penuh, jadi jenis izin ikut dikirim.
export function issuedPermitValues(permit: Permit, values: PermitIssueValues): PermitFormValues {
  return { ...permitFormValues(permit), ...values, status: 'issued' }
}

export interface BudgetItemFilter {
  search?: string
  page?: number
  pageSize?: number
}

export interface BudgetItemRequest {
  code: string
  description: string
  volume: number
  unitOfMeasureId: string
  unitPrice: number
}

export interface BudgetItemFormValues {
  code: string
  description: string
  volume: string
  unitOfMeasureId: string
  unitPrice: string
}

export const EMPTY_BUDGET_ITEM_FORM: BudgetItemFormValues = {
  code: '',
  description: '',
  volume: '',
  unitOfMeasureId: '',
  unitPrice: '',
}

// Batas panjang mengikuti tag binding BudgetItemRequest di backend.
export const BUDGET_ITEM_CODE_MAX_LENGTH = 20
export const BUDGET_ITEM_DESCRIPTION_MAX_LENGTH = 200

const VOLUME_PRECISION = 100

export function budgetItemFormValues(item: BudgetItem): BudgetItemFormValues {
  return {
    code: item.code,
    description: item.description,
    volume: String(item.volume),
    unitOfMeasureId: item.unitOfMeasureId,
    unitPrice: String(item.unitPrice),
  }
}

export function toBudgetItemRequest(values: BudgetItemFormValues): BudgetItemRequest {
  return {
    code: values.code.trim(),
    description: values.description.trim(),
    volume: values.volume === '' ? 0 : Number(values.volume),
    unitOfMeasureId: values.unitOfMeasureId,
    unitPrice: values.unitPrice === '' ? 0 : Number(values.unitPrice),
  }
}

// Sama dengan hitungan server: volume dibulatkan dua desimal, lalu jumlah
// adalah volume kali harga satuan dibulatkan ke rupiah terdekat.
export function budgetItemTotal(values: BudgetItemFormValues): number | null {
  const volume = Number(values.volume)
  const unitPrice = Number(values.unitPrice)
  if (values.volume === '' || values.unitPrice === '' || !Number.isFinite(volume) || !Number.isFinite(unitPrice)) {
    return null
  }
  return Math.round((Math.round(volume * VOLUME_PRECISION) / VOLUME_PRECISION) * unitPrice)
}

export function budgetShare(total: number, grandTotal: number): number {
  return grandTotal > 0 ? Math.round((total / grandTotal) * 100) : 0
}
