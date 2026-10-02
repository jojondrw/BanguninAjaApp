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
function toApiDate(value: string): string | undefined {
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
