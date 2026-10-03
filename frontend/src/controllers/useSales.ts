import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  leadWithStage,
  toApiDate,
  toContractRequest,
  toCustomerRequest,
  toInstallmentRequest,
  toLeadRequest,
  toUnitRequest,
  type Contract,
  type ContractFilter,
  type ContractFormValues,
  type ContractTarget,
  type Customer,
  type CustomerFilter,
  type CustomerFormValues,
  type Installment,
  type InstallmentFilter,
  type InstallmentFormValues,
  type InstallmentStatus,
  type Lead,
  type LeadFilter,
  type LeadFormValues,
  type LeadStage,
  type PropertyUnit,
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

export function useLeads(filter: LeadFilter = {}) {
  return useQuery({
    queryKey: [SALES_KEY, 'leads', filter],
    queryFn: () => salesApi.leads(filter),
    ...PAGED_QUERY,
  })
}

export function useLeadStageCounts(projectId?: string) {
  return useQuery({
    queryKey: [SALES_KEY, 'lead-stage-counts', projectId ?? 'all'],
    queryFn: () => salesApi.leadStageCounts(projectId),
    ...PAGED_QUERY,
  })
}

export function useContractSummary(projectId?: string) {
  return useQuery({
    queryKey: [SALES_KEY, 'contract-summary', projectId ?? 'all'],
    queryFn: () => salesApi.contractSummary(projectId),
    ...DATA_QUERY,
  })
}

export function useContracts(filter: ContractFilter = {}) {
  return useQuery({
    queryKey: [SALES_KEY, 'contracts', filter],
    queryFn: () => salesApi.contracts(filter),
    ...PAGED_QUERY,
  })
}

// Rincian kontrak diambil sendiri, bukan dari baris tabel, supaya tetap
// tampil walau kontraknya keluar dari halaman atau saringan yang sedang dibuka.
// Selama kontraknya dihapus, query dimatikan supaya tidak ikut diambil ulang
// dan berbalik menjadi 404 sebelum kartunya ditutup.
export function useContract(id: string, enabled = true) {
  return useQuery({
    queryKey: [SALES_KEY, 'contract', id],
    queryFn: () => salesApi.contract(id),
    enabled,
    ...DATA_QUERY,
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

export function useUpdateUnit() {
  return useSalesMutation(({ id, values }: { id: string; values: UnitFormValues }) =>
    salesApi.updateUnit(id, toUnitRequest(values)),
  )
}

export function useDeleteUnit() {
  return useSalesMutation((unit: PropertyUnit) => salesApi.deleteUnit(unit.id))
}

export function useConvertLead() {
  return useSalesMutation(({ leadId, values }: { leadId: string; values: CustomerFormValues }) =>
    salesApi.convertLead(leadId, toCustomerRequest(values)),
  )
}

export function useCreateCustomer() {
  return useSalesMutation((values: CustomerFormValues) =>
    salesApi.createCustomer(toCustomerRequest(values)),
  )
}

export function useUpdateCustomer() {
  return useSalesMutation(({ id, values }: { id: string; values: CustomerFormValues }) =>
    salesApi.updateCustomer(id, toCustomerRequest(values)),
  )
}

export function useDeleteCustomer() {
  return useSalesMutation((customer: Customer) => salesApi.deleteCustomer(customer.id))
}

export function useCreateLead() {
  return useSalesMutation((values: LeadFormValues) => salesApi.createLead(toLeadRequest(values)))
}

export function useUpdateLead() {
  return useSalesMutation(({ id, values }: { id: string; values: LeadFormValues }) =>
    salesApi.updateLead(id, toLeadRequest(values)),
  )
}

export function useMoveLeadStage() {
  return useSalesMutation(({ lead, stage }: { lead: Lead; stage: LeadStage }) =>
    salesApi.updateLead(lead.id, leadWithStage(lead, stage)),
  )
}

export function useDeleteLead() {
  return useSalesMutation((lead: Lead) => salesApi.deleteLead(lead.id))
}

export function useCreateContract() {
  return useSalesMutation((values: ContractFormValues) =>
    salesApi.createContract(toContractRequest(values)),
  )
}

export function useUpdateContract() {
  return useSalesMutation(({ id, values }: { id: string; values: ContractFormValues }) =>
    salesApi.updateContract(id, toContractRequest(values)),
  )
}

export function useUpdateContractStatus() {
  return useSalesMutation(({ id, status }: { id: string; status: ContractTarget }) =>
    salesApi.updateContractStatus(id, status),
  )
}

export function useDeleteContract() {
  return useSalesMutation((contract: Contract) => salesApi.deleteContract(contract.id))
}

export function useCreateInstallment() {
  return useSalesMutation(({ contractId, values }: { contractId: string; values: InstallmentFormValues }) =>
    salesApi.createInstallment(contractId, toInstallmentRequest(values)),
  )
}

export function useUpdateInstallment() {
  return useSalesMutation(({ installment, values }: { installment: Installment; values: InstallmentFormValues }) =>
    salesApi.updateInstallment(installment.contractId, installment.id, toInstallmentRequest(values)),
  )
}

export function useDeleteInstallment() {
  return useSalesMutation((installment: Installment) =>
    salesApi.deleteInstallment(installment.contractId, installment.id),
  )
}

export function usePayInstallment() {
  return useSalesMutation(({ installment, paidDate }: { installment: Installment; paidDate: string }) =>
    salesApi.payInstallment(installment.contractId, installment.id, { paidDate: toApiDate(paidDate) }),
  )
}
