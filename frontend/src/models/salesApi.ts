import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import {
  LEAD_STAGES,
  type Contract,
  type ContractFilter,
  type ContractRequest,
  type ContractTarget,
  type Customer,
  type CustomerFilter,
  type CustomerRequest,
  type Installment,
  type InstallmentFilter,
  type InstallmentPaymentRequest,
  type InstallmentRequest,
  type Lead,
  type LeadConversion,
  type LeadFilter,
  type LeadRequest,
  type LeadStageCounts,
  type PropertyUnit,
  type UnitFilter,
  type UnitRequest,
  type UnitStatusCount,
} from './sales'

function leads(filter: LeadFilter = {}) {
  return request<Page<Lead>>(`/sales/leads${toQueryString({ ...filter })}`)
}

// Backend belum punya ringkasan per tahap, jadi jumlahnya diambil dari
// totalItems satu baris per tahap.
async function leadStageCounts(projectId?: string): Promise<LeadStageCounts> {
  const pages = await Promise.all(LEAD_STAGES.map((stage) => leads({ stage, projectId, pageSize: 1 })))
  return Object.fromEntries(LEAD_STAGES.map((stage, index) => [stage, pages[index].totalItems])) as LeadStageCounts
}

export const salesApi = {
  units: (filter: UnitFilter = {}) =>
    request<Page<PropertyUnit>>(`/sales/units${toQueryString({ ...filter })}`),

  unitSummary: (projectId?: string) =>
    request<UnitStatusCount[]>(`/sales/units/summary${toQueryString({ projectId })}`),

  createUnit: (body: UnitRequest) => request<PropertyUnit>('/sales/units', { method: 'POST', body }),

  updateUnit: (id: string, body: UnitRequest) =>
    request<PropertyUnit>(`/sales/units/${id}`, { method: 'PUT', body }),

  deleteUnit: (id: string) => request<void>(`/sales/units/${id}`, { method: 'DELETE' }),

  customers: (filter: CustomerFilter = {}) =>
    request<Page<Customer>>(`/sales/customers${toQueryString({ ...filter })}`),

  createCustomer: (body: CustomerRequest) =>
    request<Customer>('/sales/customers', { method: 'POST', body }),

  updateCustomer: (id: string, body: CustomerRequest) =>
    request<Customer>(`/sales/customers/${id}`, { method: 'PUT', body }),

  deleteCustomer: (id: string) => request<void>(`/sales/customers/${id}`, { method: 'DELETE' }),

  leads,

  leadStageCounts,

  createLead: (body: LeadRequest) => request<Lead>('/sales/leads', { method: 'POST', body }),

  updateLead: (id: string, body: LeadRequest) => request<Lead>(`/sales/leads/${id}`, { method: 'PUT', body }),

  deleteLead: (id: string) => request<void>(`/sales/leads/${id}`, { method: 'DELETE' }),
  convertLead: (id: string, body: CustomerRequest) =>
    request<LeadConversion>(`/sales/leads/${id}/convert`, { method: 'POST', body }),

  contracts: (filter: ContractFilter = {}) =>
    request<Page<Contract>>(`/sales/contracts${toQueryString({ ...filter })}`),

  contract: (id: string) => request<Contract>(`/sales/contracts/${id}`),

  createContract: (body: ContractRequest) =>
    request<Contract>('/sales/contracts', { method: 'POST', body }),

  updateContract: (id: string, body: ContractRequest) =>
    request<Contract>(`/sales/contracts/${id}`, { method: 'PUT', body }),

  updateContractStatus: (id: string, status: ContractTarget) =>
    request<Contract>(`/sales/contracts/${id}/status`, { method: 'PATCH', body: { status } }),

  deleteContract: (id: string) => request<void>(`/sales/contracts/${id}`, { method: 'DELETE' }),

  installments: (filter: InstallmentFilter = {}) =>
    request<Page<Installment>>(`/sales/installments${toQueryString({ ...filter })}`),

  createInstallment: (contractId: string, body: InstallmentRequest) =>
    request<Installment>(`/sales/contracts/${contractId}/installments`, { method: 'POST', body }),

  updateInstallment: (contractId: string, id: string, body: InstallmentRequest) =>
    request<Installment>(`/sales/contracts/${contractId}/installments/${id}`, { method: 'PUT', body }),

  deleteInstallment: (contractId: string, id: string) =>
    request<void>(`/sales/contracts/${contractId}/installments/${id}`, { method: 'DELETE' }),

  payInstallment: (contractId: string, id: string, body: InstallmentPaymentRequest) =>
    request<Installment>(`/sales/contracts/${contractId}/installments/${id}/payment`, { method: 'POST', body }),
}
