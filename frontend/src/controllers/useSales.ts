import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  toContractRequest,
  toCustomerRequest,
  toInstallmentRequest,
  toUnitRequest,
  type ContractFilter,
  type ContractFormValues,
  type ContractStatus,
  type CustomerFilter,
  type CustomerFormValues,
  type InstallmentFilter,
  type InstallmentFormValues,
  type InstallmentStatus,
  type UnitFilter,
  type UnitFormValues,
} from '../models/sales'
import { salesApi } from '../models/salesApi'

const ONE_MINUTE = 60_000
const SALES_KEY = 'sales'

const DATA_QUERY = {
  staleTime: ONE_MINUTE,
  retry: 0,
}

// Daftar yang dipaginasi tetap menampilkan halaman lama selama halaman baru
// dimuat, supaya tabel tidak berkedip kosong setiap pindah halaman.
const PAGED_QUERY = {
  ...DATA_QUERY,
  placeholderData: keepPreviousData,
}

export function useUnits(filter: UnitFilter = {}) {
  return useQuery({
    queryKey: [SALES_KEY, 'units', filter],
    queryFn: () => salesApi.units(filter),
    ...PAGED_QUERY,
  })
}

export function useUnitSummary(projectId?: string) {
  return useQuery({
    queryKey: [SALES_KEY, 'unit-summary', projectId ?? 'all'],
    queryFn: () => salesApi.unitSummary(projectId),
    ...DATA_QUERY,
  })
}

export function useCustomers(filter: CustomerFilter = {}) {
  return useQuery({
    queryKey: [SALES_KEY, 'customers', filter],
    queryFn: () => salesApi.customers(filter),
    ...PAGED_QUERY,
  })
}

export function useContracts(filter: ContractFilter = {}) {
  return useQuery({
    queryKey: [SALES_KEY, 'contracts', filter],
    queryFn: () => salesApi.contracts(filter),
    ...PAGED_QUERY,
  })
}

export function useInstallments(filter: InstallmentFilter, enabled = true) {
  return useQuery({
    queryKey: [SALES_KEY, 'installments', filter],
    queryFn: () => salesApi.installments(filter),
    enabled,
    ...PAGED_QUERY,
  })
}

// Hanya butuh totalItems, jadi cukup satu baris per permintaan.
export function useInstallmentCount(status: InstallmentStatus) {
  return useQuery({
    queryKey: [SALES_KEY, 'installment-count', status],
    queryFn: () => salesApi.installments({ status, pageSize: 1 }),
    select: (page) => page.totalItems,
    ...DATA_QUERY,
  })
}

// Kontrak mengubah status unit, dan cicilan bergantung pada kontrak. Semua
// mutasi penjualan membatalkan seluruh query penjualan supaya KPI ikut segar.
function useSalesMutation<TVariables, TData>(mutationFn: (variables: TVariables) => Promise<TData>) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [SALES_KEY] }),
  })
}

export function useCreateUnit() {
  return useSalesMutation((values: UnitFormValues) => salesApi.createUnit(toUnitRequest(values)))
}

export function useCreateCustomer() {
  return useSalesMutation((values: CustomerFormValues) =>
    salesApi.createCustomer(toCustomerRequest(values)),
  )
}

export function useCreateContract() {
  return useSalesMutation((values: ContractFormValues) =>
    salesApi.createContract(toContractRequest(values)),
  )
}

export function useUpdateContractStatus() {
  return useSalesMutation(({ id, status }: { id: string; status: Exclude<ContractStatus, 'draft'> }) =>
    salesApi.updateContractStatus(id, status),
  )
}

export function useCreateInstallment() {
  return useSalesMutation(({ contractId, values }: { contractId: string; values: InstallmentFormValues }) =>
    salesApi.createInstallment(contractId, toInstallmentRequest(values)),
  )
}
