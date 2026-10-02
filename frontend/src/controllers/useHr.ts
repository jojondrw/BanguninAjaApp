import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  toAttendanceRequest,
  toEmployeeRequest,
  toPayrollRequest,
  type AttendanceFilter,
  type AttendanceFormValues,
  type EmployeeFilter,
  type EmployeeFormValues,
  type EmploymentCount,
  type Payroll,
  type PayrollFilter,
  type PayrollFormValues,
} from '../models/hr'
import { hrApi } from '../models/hrApi'
import type { Page } from '../models/common'

const ONE_MINUTE = 60_000
const MAX_PAGE_SIZE = 100

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

// Pilihan karyawan di formulir absensi dan penggajian: hanya yang masih aktif.
export function useActiveEmployeeOptions() {
  return useEmployees({ active: 'true', pageSize: MAX_PAGE_SIZE })
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

export function useCreateAttendance(date: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: AttendanceFormValues) =>
      hrApi.createAttendance(toAttendanceRequest(values, date)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ATTENDANCES_KEY })
      void queryClient.invalidateQueries({ queryKey: PAYROLLS_KEY })
    },
  })
}

export function usePayrolls(filter: PayrollFilter) {
  return useQuery({
    queryKey: [...PAYROLLS_KEY, filter],
    queryFn: () => hrApi.payrolls(filter),
    ...PAGED_QUERY,
  })
}

export interface PayrollTotals {
  count: number
  netPay: number
  paidNetPay: number
  unpaidNetPay: number
  unpaidCount: number
  isPartial: boolean
}

function summarizePayrolls(page: Page<Payroll>): PayrollTotals {
  const unpaid = page.items.filter((payroll) => !payroll.paid)
  const netPay = page.items.reduce((sum, payroll) => sum + payroll.netPay, 0)
  const unpaidNetPay = unpaid.reduce((sum, payroll) => sum + payroll.netPay, 0)

  return {
    count: page.totalItems,
    netPay,
    paidNetPay: netPay - unpaidNetPay,
    unpaidNetPay,
    unpaidCount: unpaid.length,
    isPartial: page.totalItems > page.items.length,
  }
}

// Backend belum punya ringkasan penggajian, jadi totalnya dijumlah dari satu
// halaman terbesar yang diizinkan (100 baris). isPartial menandai kalau periode
// itu punya lebih dari 100 slip.
export function usePayrollTotals(period: string) {
  return useQuery({
    queryKey: [...PAYROLLS_KEY, 'totals', period],
    queryFn: () => hrApi.payrolls({ period, pageSize: MAX_PAGE_SIZE }),
    select: summarizePayrolls,
    ...DATA_QUERY,
  })
}

export function useCreatePayroll(period: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (values: PayrollFormValues) => hrApi.createPayroll(toPayrollRequest(values, period)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: PAYROLLS_KEY }),
  })
}

export function usePayPayroll() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: hrApi.payPayroll,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: PAYROLLS_KEY }),
  })
}
