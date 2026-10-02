import { toApiDate } from '../shared/localDate'

export type EmploymentType ='permanent' | 'contract' | 'daily' | 'subcontractor'
export type AttendanceStatus = 'present' | 'permitted' | 'sick' | 'absent' | 'holiday'

export interface Employee {
  id: string
  identityNumber: string
  name: string
  position: string
  projectId: string | null
  userId: string | null
  employmentType: EmploymentType
  joinedDate: string
  leftDate: string | null
  baseSalary: number
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface EmploymentCount {
  employmentType: EmploymentType
  total: number
}

export interface EmployeeFilter {
  search?: string
  employmentType?: EmploymentType
  projectId?: string
  active?: 'true' | 'false'
  page?: number
  pageSize?: number
}

export interface EmployeeRequest {
  identityNumber: string
  name: string
  position: string
  projectId?: string
  employmentType: EmploymentType
  joinedDate: string
  baseSalary: number
}

export interface Attendance {
  id: string
  employeeId: string
  employeeName: string
  date: string
  checkInTime: string | null
  checkOutTime: string | null
  status: AttendanceStatus
  createdAt: string
  updatedAt: string
}

// dateFrom dan dateTo dibaca backend dengan format "YYYY-MM-DD", beda dengan
// body JSON yang wajib RFC 3339.
export interface AttendanceFilter {
  employeeId?: string
  projectId?: string
  status?: AttendanceStatus
  dateFrom?: string
  dateTo?: string
  page?: number
  pageSize?: number
}

export interface AttendanceRequest {
  employeeId: string
  date: string
  checkInTime?: string
  checkOutTime?: string
  status: AttendanceStatus
}

export interface Payroll {
  id: string
  employeeId: string
  employeeName: string
  period: string
  basicPay: number
  allowance: number
  deduction: number
  netPay: number
  paid: boolean
  createdAt: string
  updatedAt: string
}

export interface PayrollFilter {
  employeeId?: string
  period?: string
  paid?: 'true' | 'false'
  page?: number
  pageSize?: number
}

export interface PayrollRequest {
  employeeId: string
  period: string
  allowance: number
  deduction: number
}

// Batas panjang mengikuti tag binding EmployeeRequest di backend.
export const EMPLOYEE_IDENTITY_MAX_LENGTH = 20
export const EMPLOYEE_NAME_MAX_LENGTH = 160
export const EMPLOYEE_POSITION_MAX_LENGTH = 80

export const EMPLOYMENT_TYPE_LABEL: Record<EmploymentType, string> = {
  permanent: 'Tetap',
  contract: 'Kontrak',
  daily: 'Harian',
  subcontractor: 'Subkontraktor',
}

export const EMPLOYMENT_TYPES: EmploymentType[] = ['permanent', 'contract', 'daily', 'subcontractor']

export const ATTENDANCE_STATUS_LABEL: Record<AttendanceStatus, string> = {
  present: 'Hadir',
  permitted: 'Izin',
  sick: 'Sakit',
  absent: 'Tidak hadir',
  holiday: 'Libur',
}

export const ATTENDANCE_STATUSES: AttendanceStatus[] = ['present', 'permitted', 'sick', 'absent', 'holiday']

export const ATTENDANCE_STATUS_TONE: Record<AttendanceStatus, string> = {
  present: 'bg-green-100 text-green-800',
  permitted: 'bg-blue-100 text-blue-800',
  sick: 'bg-amber-100 text-amber-800',
  absent: 'bg-red-100 text-red-800',
  holiday: 'bg-slate-100 text-slate-700',
}

export interface EmployeeFormValues {
  identityNumber: string
  name: string
  position: string
  employmentType: EmploymentType
  projectId: string
  joinedDate: string
  baseSalary: string
}

export const EMPTY_EMPLOYEE_FORM: EmployeeFormValues = {
  identityNumber: '',
  name: '',
  position: '',
  employmentType: 'permanent',
  projectId: '',
  joinedDate: '',
  baseSalary: '',
}

export interface AttendanceFormValues {
  employeeId: string
  status: AttendanceStatus
  checkInTime: string
  checkOutTime: string
}

export const EMPTY_ATTENDANCE_FORM: AttendanceFormValues = {
  employeeId: '',
  status: 'present',
  checkInTime: '07:30',
  checkOutTime: '',
}

export interface PayrollFormValues {
  employeeId: string
  allowance: string
  deduction: string
}

export const EMPTY_PAYROLL_FORM: PayrollFormValues = {
  employeeId: '',
  allowance: '',
  deduction: '',
}

function amount(value: string): number {
  return value === '' ? 0 : Number(value)
}

export function toEmployeeRequest(values: EmployeeFormValues): EmployeeRequest {
  return {
    identityNumber: values.identityNumber.trim(),
    name: values.name.trim(),
    position: values.position.trim(),
    projectId: values.projectId === '' ? undefined : values.projectId,
    employmentType: values.employmentType,
    joinedDate: toApiDate(values.joinedDate),
    baseSalary: amount(values.baseSalary),
  }
}

// Jam masuk dan pulang hanya boleh dikirim untuk karyawan yang hadir. Backend
// menolak jam untuk status lain.
export function toAttendanceRequest(values: AttendanceFormValues, date: string): AttendanceRequest {
  const isPresent = values.status === 'present'

  return {
    employeeId: values.employeeId,
    date: toApiDate(date),
    status: values.status,
    checkInTime: isPresent && values.checkInTime !== '' ? values.checkInTime : undefined,
    checkOutTime: isPresent && values.checkOutTime !== '' ? values.checkOutTime : undefined,
  }
}

export function toPayrollRequest(values: PayrollFormValues, period: string): PayrollRequest {
  return {
    employeeId: values.employeeId,
    period,
    allowance: amount(values.allowance),
    deduction: amount(values.deduction),
  }
}
