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
