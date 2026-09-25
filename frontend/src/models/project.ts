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
