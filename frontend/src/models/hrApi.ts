import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  Attendance,
  AttendanceFilter,
  AttendanceRequest,
  Employee,
  EmployeeFilter,
  EmployeeRequest,
  EmploymentCount,
  Payroll,
  PayrollFilter,
  PayrollRequest,
  PayrollSummary,
  PayrollSummaryFilter,
} from './hr'

export const hrApi = {
  employees: (filter: EmployeeFilter = {}) =>
    request<Page<Employee>>(`/hr/employees${toQueryString({ ...filter })}`),

  employeeSummary: () => request<EmploymentCount[]>('/hr/employees/summary'),

  createEmployee: (body: EmployeeRequest) =>
    request<Employee>('/hr/employees', { method: 'POST', body }),

  updateEmployee: (id: string, body: EmployeeRequest) =>
    request<Employee>(`/hr/employees/${id}`, { method: 'PUT', body }),

  deleteEmployee: (id: string) => request<void>(`/hr/employees/${id}`, { method: 'DELETE' }),

  attendances: (filter: AttendanceFilter = {}) =>
    request<Page<Attendance>>(`/hr/attendances${toQueryString({ ...filter })}`),

  createAttendance: (body: AttendanceRequest) =>
    request<Attendance>('/hr/attendances', { method: 'POST', body }),

  updateAttendance: (id: string, body: AttendanceRequest) =>
    request<Attendance>(`/hr/attendances/${id}`, { method: 'PUT', body }),

  deleteAttendance: (id: string) => request<void>(`/hr/attendances/${id}`, { method: 'DELETE' }),

  payrolls: (filter: PayrollFilter = {}) =>
    request<Page<Payroll>>(`/hr/payrolls${toQueryString({ ...filter })}`),

  payrollSummary: (filter: PayrollSummaryFilter = {}) =>
    request<PayrollSummary>(`/hr/payrolls/summary${toQueryString({ ...filter })}`),

  createPayroll: (body: PayrollRequest) =>
    request<Payroll>('/hr/payrolls', { method: 'POST', body }),

  updatePayroll: (id: string, body: PayrollRequest) =>
    request<Payroll>(`/hr/payrolls/${id}`, { method: 'PUT', body }),

  payPayroll: (id: string) => request<Payroll>(`/hr/payrolls/${id}/pay`, { method: 'PATCH' }),

  deletePayroll: (id: string) => request<void>(`/hr/payrolls/${id}`, { method: 'DELETE' }),
}
