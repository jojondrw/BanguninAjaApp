import { request } from '../shared/apiClient'
import { toQueryString, type Page } from './common'
import type {
  Contract,
  ContractFilter,
  ContractRequest,
  ContractStatus,
  Customer,
  CustomerFilter,
  CustomerRequest,
  Installment,
  InstallmentFilter,
  InstallmentRequest,
  PropertyUnit,
  UnitFilter,
  UnitRequest,
  UnitStatusCount,
} from './sales'

export const salesApi = {
  units: (filter: UnitFilter = {}) =>
    request<Page<PropertyUnit>>(`/sales/units${toQueryString({ ...filter })}`),

  unitSummary: (projectId?: string) =>
    request<UnitStatusCount[]>(`/sales/units/summary${toQueryString({ projectId })}`),

  createUnit: (body: UnitRequest) => request<PropertyUnit>('/sales/units', { method: 'POST', body }),

  customers: (filter: CustomerFilter = {}) =>
    request<Page<Customer>>(`/sales/customers${toQueryString({ ...filter })}`),

  createCustomer: (body: CustomerRequest) =>
    request<Customer>('/sales/customers', { method: 'POST', body }),

  contracts: (filter: ContractFilter = {}) =>
    request<Page<Contract>>(`/sales/contracts${toQueryString({ ...filter })}`),

  createContract: (body: ContractRequest) =>
    request<Contract>('/sales/contracts', { method: 'POST', body }),

  updateContractStatus: (id: string, status: Exclude<ContractStatus, 'draft'>) =>
    request<Contract>(`/sales/contracts/${id}/status`, { method: 'PATCH', body: { status } }),

  installments: (filter: InstallmentFilter = {}) =>
    request<Page<Installment>>(`/sales/installments${toQueryString({ ...filter })}`),

  createInstallment: (contractId: string, body: InstallmentRequest) =>
    request<Installment>(`/sales/contracts/${contractId}/installments`, { method: 'POST', body }),
}
