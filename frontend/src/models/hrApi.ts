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
} from './hr'

export const hrApi = {
  employees: (filter: EmployeeFilter = {}) =>
    request<Page<Employee>>(`/hr/employees${toQueryString({ ...filter })}`),

  employeeSummary: () => request<EmploymentCount[]>('/hr/employees/summary'),

  createEmployee: (body: EmployeeRequest) =>
    request<Employee>('/hr/employees', { method: 'POST', body }),

  attendances: (filter: AttendanceFilter = {}) =>
    request<Page<Attendance>>(`/hr/attendances${toQueryString({ ...filter })}`),

  createAttendance: (body: AttendanceRequest) =>
    request<Attendance>('/hr/attendances', { method: 'POST', body }),

  payrolls: (filter: PayrollFilter = {}) =>
    request<Page<Payroll>>(`/hr/payrolls${toQueryString({ ...filter })}`),

  createPayroll: (body: PayrollRequest) =>
    request<Payroll>('/hr/payrolls', { method: 'POST', body }),

  payPayroll: (id: string) => request<Payroll>(`/hr/payrolls/${id}/pay`, { method: 'PATCH' }),
}
