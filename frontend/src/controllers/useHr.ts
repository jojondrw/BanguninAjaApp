import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  employeeWithLeftDate,
  inputDate,
  toAttendanceRequest,
  toEmployeeRequest,
  toPayrollRequest,
  type Attendance,
  type AttendanceFilter,
  type AttendanceFormValues,
  type Employee,
  type EmployeeFilter,
  type EmployeeFormValues,
  type EmploymentCount,
  type Payroll,
  type PayrollFilter,
  type PayrollFormValues,
  type PayrollSummaryFilter,
} from '../models/hr'
import { hrApi } from '../models/hrApi'

const ONE_MINUTE = 60_000

const HR_KEY = 'hr'
const EMPLOYEES_KEY = [HR_KEY, 'employees'] as const
const SUMMARY_KEY = [HR_KEY, 'employee-summary'] as const
const ATTENDANCES_KEY = [HR_KEY, 'attendances'] as const
const PAYROLLS_KEY = [HR_KEY, 'payrolls'] as const

const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

export function useEmployees(filter: EmployeeFilter) {
  return useQuery({
    queryKey: [...EMPLOYEES_KEY, filter],
    queryFn: () => hrApi.employees(filter),
    ...PAGED_QUERY,
  })
}

export interface EmployeeHeadcount {
  total: number
  byType: EmploymentCount[]
}

export function useEmployeeHeadcount() {
  return useQuery({
    queryKey: SUMMARY_KEY,
    queryFn: hrApi.employeeSummary,
    select: (counts): EmployeeHeadcount => ({
      total: counts.reduce((sum, count) => sum + count.total, 0),
      byType: counts,
    }),
    ...DATA_QUERY,
  })
}

// Nama, proyek, dan status kerja karyawan ikut tampil di absensi, penggajian,
// dan hitungan karyawan aktif, jadi semua data SDM disegarkan.
function useInvalidateHr() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: [HR_KEY] })
}

export function useCreateEmployee() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: EmployeeFormValues) => hrApi.createEmployee(toEmployeeRequest(values)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: EMPLOYEES_KEY })
      void queryClient.invalidateQueries({ queryKey: SUMMARY_KEY })
    },
  })
}

export interface EmployeeUpdate {
  employee: Employee
  values: EmployeeFormValues
}

export function useUpdateEmployee() {
  const invalidateHr = useInvalidateHr()

  return useMutation({
    mutationFn: ({ employee, values }: EmployeeUpdate) =>
      hrApi.updateEmployee(employee.id, toEmployeeRequest(values, employee.userId)),
    onSuccess: invalidateHr,
  })
}

export interface EmployeeLeftDate {
  employee: Employee
  leftDate: string
}

// Nonaktifkan (isi tanggal keluar) dan aktifkan lagi (kosongkan tanggal keluar).
export function useSetEmployeeLeftDate() {
  const invalidateHr = useInvalidateHr()

  return useMutation({
    mutationFn: ({ employee, leftDate }: EmployeeLeftDate) =>
      hrApi.updateEmployee(employee.id, employeeWithLeftDate(employee, leftDate)),
    onSuccess: invalidateHr,
  })
}

export function useDeleteEmployee() {
  const invalidateHr = useInvalidateHr()

  return useMutation({
    mutationFn: (employee: Employee) => hrApi.deleteEmployee(employee.id),
    onSuccess: invalidateHr,
  })
}

export function useAttendances(filter: AttendanceFilter) {
  return useQuery({
    queryKey: [...ATTENDANCES_KEY, filter],
    queryFn: () => hrApi.attendances(filter),
    ...PAGED_QUERY,
  })
}

// Cukup totalItems dari halaman berukuran satu, jadi hitungannya dikerjakan
// server, bukan dengan mengambil semua baris.
export function useAttendanceCount(filter: Omit<AttendanceFilter, 'page' | 'pageSize'>) {
  return useQuery({
    queryKey: [...ATTENDANCES_KEY, 'count', filter],
    queryFn: () => hrApi.attendances({ ...filter, pageSize: 1 }),
    select: (page) => page.totalItems,
    ...DATA_QUERY,
  })
}

// Upah karyawan harian dihitung dari hari hadir, jadi perubahan absensi ikut
// menyegarkan penggajian.
function useInvalidateAttendances() {
  const queryClient = useQueryClient()

  return () => {
    void queryClient.invalidateQueries({ queryKey: ATTENDANCES_KEY })
    void queryClient.invalidateQueries({ queryKey: PAYROLLS_KEY })
  }
}

export function useCreateAttendance(date: string) {
  const invalidateAttendances = useInvalidateAttendances()

  return useMutation({
    mutationFn: (values: AttendanceFormValues) =>
      hrApi.createAttendance(toAttendanceRequest(values, date)),
    onSuccess: invalidateAttendances,
  })
}

export interface AttendanceUpdate {
  attendance: Attendance
  values: AttendanceFormValues
}

export function useUpdateAttendance() {
  const invalidateAttendances = useInvalidateAttendances()

  return useMutation({
    mutationFn: ({ attendance, values }: AttendanceUpdate) =>
      hrApi.updateAttendance(attendance.id, toAttendanceRequest(values, inputDate(attendance.date))),
    onSuccess: invalidateAttendances,
  })
}

export function useDeleteAttendance() {
  const invalidateAttendances = useInvalidateAttendances()

  return useMutation({
    mutationFn: (attendance: Attendance) => hrApi.deleteAttendance(attendance.id),
    onSuccess: invalidateAttendances,
  })
}

export function usePayrolls(filter: PayrollFilter) {
  return useQuery({
    queryKey: [...PAYROLLS_KEY, filter],
    queryFn: () => hrApi.payrolls(filter),
    ...PAGED_QUERY,
  })
}

// Total periode dijumlah server dari semua slip, jadi tidak terbatas satu
// halaman daftar.
export function usePayrollSummary(filter: PayrollSummaryFilter) {
  return useQuery({
    queryKey: [...PAYROLLS_KEY, 'summary', filter],
    queryFn: () => hrApi.payrollSummary(filter),
    ...DATA_QUERY,
  })
}

// Dipakai daftar "Perlu perhatian" di Ringkasan: total slip satu periode.
export function usePayrollTotals(period: string) {
  return usePayrollSummary({ period })
}

function useInvalidatePayrolls() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: PAYROLLS_KEY })
}

export function useCreatePayroll(period: string) {
  const invalidatePayrolls = useInvalidatePayrolls()

  return useMutation({
    mutationFn: (values: PayrollFormValues) => hrApi.createPayroll(toPayrollRequest(values, period)),
    onSuccess: invalidatePayrolls,
  })
}

export interface PayrollUpdate {
  payroll: Payroll
  values: PayrollFormValues
}

export function useUpdatePayroll() {
  const invalidatePayrolls = useInvalidatePayrolls()

  return useMutation({
    mutationFn: ({ payroll, values }: PayrollUpdate) =>
      hrApi.updatePayroll(payroll.id, toPayrollRequest(values, payroll.period)),
    onSuccess: invalidatePayrolls,
  })
}

export function usePayPayroll() {
  const invalidatePayrolls = useInvalidatePayrolls()

  return useMutation({
    mutationFn: hrApi.payPayroll,
    onSuccess: invalidatePayrolls,
  })
}

export function useDeletePayroll() {
  const invalidatePayrolls = useInvalidatePayrolls()

  return useMutation({
    mutationFn: (payroll: Payroll) => hrApi.deletePayroll(payroll.id),
    onSuccess: invalidatePayrolls,
  })
}
