export type UnitStatus = 'available' | 'reserved' | 'sold' | 'on_hold'
export type NewUnitStatus = 'available' | 'on_hold'
export type ContractType = 'cash' | 'mortgage' | 'installment'
export type ContractStatus = 'draft' | 'active' | 'paid' | 'cancelled'
export type ContractTarget = Exclude<ContractStatus, 'draft'>
export type InstallmentStatus = 'not_due' | 'due' | 'paid' | 'overdue'
export type LeadStage = 'new' | 'interested' | 'negotiating' | 'won' | 'cancelled'

export interface PropertyUnit {
  id: string
  code: string
  projectId: string
  unitType: string
  areaSqm: number
  price: number
  status: UnitStatus
  createdAt: string
  updatedAt: string
}

export interface UnitStatusCount {
  status: UnitStatus
  total: number
}

export interface UnitFilter {
  search?: string
  projectId?: string
  status?: UnitStatus
  page?: number
  pageSize?: number
}

export interface UnitRequest {
  code: string
  projectId: string
  unitType: string
  areaSqm: number
  price: number
  status: NewUnitStatus
}

export interface Customer {
  id: string
  name: string
  contact: string
  email: string
  identityNumber: string
  address: string
  createdAt: string
  updatedAt: string
}

export interface CustomerFilter {
  search?: string
  page?: number
  pageSize?: number
}

export interface CustomerRequest {
  name: string
  contact: string
  email: string
  identityNumber: string
  address: string
}

export interface Lead {
  id: string
  name: string
  contact: string
  projectId: string | null
  source: string
  stage: LeadStage
  lastContactedAt: string | null
  customerId: string | null
  createdAt: string
  updatedAt: string
}

export interface LeadConversion {
  lead: Lead
  customer: Customer
}

export interface LeadFilter {
  search?: string
  projectId?: string
  stage?: LeadStage
  page?: number
  pageSize?: number
}

// Backend menolak string kosong untuk projectId dan tanggal, jadi isian yang
// dikosongkan dikirim sebagai null.
export interface LeadRequest {
  name: string
  contact: string
  projectId: string | null
  source: string
  stage: LeadStage
  lastContactedAt: string | null
}

export type LeadStageCounts = Record<LeadStage, number>

export interface Contract {
  id: string
  number: string
  customerId: string
  customerName: string
  unitId: string
  unitCode: string
  type: ContractType
  value: number
  date: string
  status: ContractStatus
  createdAt: string
  updatedAt: string
}

export interface ContractFilter {
  search?: string
  customerId?: string
  unitId?: string
  status?: ContractStatus
  type?: ContractType
  page?: number
  pageSize?: number
}

export interface ContractRequest {
  number: string
  customerId: string
  unitId: string
  type: ContractType
  value: number
  date: string
}

export interface Installment {
  id: string
  contractId: string
  contractNumber: string
  customerId: string
  customerName: string
  installmentNumber: number
  dueDate: string
  amount: number
  paidDate: string | null
  status: InstallmentStatus
  daysOverdue: number
  createdAt: string
  updatedAt: string
}

export interface InstallmentFilter {
  contractId?: string
  customerId?: string
  status?: InstallmentStatus
  page?: number
  pageSize?: number
}

export interface InstallmentRequest {
  installmentNumber: number
  dueDate: string
  amount: number
}

export interface InstallmentPaymentRequest {
  paidDate: string
}

// Batas panjang mengikuti tag binding di backend/internal/sales/dto.go.
export const UNIT_CODE_MAX_LENGTH = 20
export const UNIT_TYPE_MAX_LENGTH = 60
export const CUSTOMER_NAME_MAX_LENGTH = 160
export const CUSTOMER_CONTACT_MAX_LENGTH = 60
export const CUSTOMER_EMAIL_MAX_LENGTH = 160
export const CUSTOMER_IDENTITY_MAX_LENGTH = 20
export const CUSTOMER_ADDRESS_MAX_LENGTH = 2000
export const CONTRACT_NUMBER_MAX_LENGTH = 40
export const LEAD_NAME_MAX_LENGTH = 160
export const LEAD_CONTACT_MAX_LENGTH = 60
export const LEAD_SOURCE_MAX_LENGTH = 60

export const LEAD_STAGES: LeadStage[] = ['new', 'interested', 'negotiating', 'won', 'cancelled']

// Urutan maju prospek. Batal bisa dipilih dari tahap mana pun lewat formulir
// ubah, karena backend tidak membatasi perpindahan tahap.
const LEAD_PIPELINE: LeadStage[] = ['new', 'interested', 'negotiating', 'won']

export function nextLeadStage(stage: LeadStage): LeadStage | null {
  const index = LEAD_PIPELINE.indexOf(stage)
  if (index === -1 || index === LEAD_PIPELINE.length - 1) {
    return null
  }
  return LEAD_PIPELINE[index + 1]
}

// Salinan contractTransitions di backend/internal/sales/service.go. Lunas dan
// batal adalah status akhir.
export const CONTRACT_TRANSITIONS: Record<ContractStatus, ContractTarget[]> = {
  draft: ['active', 'cancelled'],
  active: ['paid', 'cancelled'],
  paid: [],
  cancelled: [],
}

// Unit berstatus Dipesan atau Terjual dikunci backend (unit_locked).
export function isUnitEditable(unit: PropertyUnit): boolean {
  return unit.status === 'available' || unit.status === 'on_hold'
}

// Isian tanggal memberi "YYYY-MM-DD", sedangkan backend membaca time.Time yang
// butuh RFC 3339. Tanggal saja akan ditolak sebagai payload tidak valid.
export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}

export function todayInput(): string {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${now.getFullYear()}-${month}-${day}`
}

function toAmount(value: string): number {
  return value === '' ? 0 : Number(value)
}

// Kolom tanggal backend dikirim sebagai "YYYY-MM-DDT00:00:00Z", jadi sepuluh
// karakter pertama sudah sesuai untuk isian tanggal.
export function dateInput(iso: string | null): string {
  return iso ? iso.slice(0, 10) : ''
}

export interface UnitFormValues {
  code: string
  projectId: string
  unitType: string
  areaSqm: string
  price: string
  status: NewUnitStatus
}

export const EMPTY_UNIT_FORM: UnitFormValues = {
  code: '',
  projectId: '',
  unitType: '',
  areaSqm: '',
  price: '',
  status: 'available',
}

export function toUnitRequest(values: UnitFormValues): UnitRequest {
  return {
    code: values.code.trim(),
    projectId: values.projectId,
    unitType: values.unitType.trim(),
    areaSqm: toAmount(values.areaSqm),
    price: toAmount(values.price),
    status: values.status,
  }
}

export function unitToForm(unit: PropertyUnit): UnitFormValues {
  return {
    code: unit.code,
    projectId: unit.projectId,
    unitType: unit.unitType,
    areaSqm: String(unit.areaSqm),
    price: String(unit.price),
    status: unit.status === 'on_hold' ? 'on_hold' : 'available',
  }
}

export interface CustomerFormValues {
  name: string
  identityNumber: string
  contact: string
  email: string
  address: string
}

export const EMPTY_CUSTOMER_FORM: CustomerFormValues = {
  name: '',
  identityNumber: '',
  contact: '',
  email: '',
  address: '',
}

export function toCustomerRequest(values: CustomerFormValues): CustomerRequest {
  return {
    name: values.name.trim(),
    identityNumber: values.identityNumber.trim(),
    contact: values.contact.trim(),
    email: values.email.trim(),
    address: values.address.trim(),
  }
}

export function customerToForm(customer: Customer): CustomerFormValues {
  return {
    name: customer.name,
    identityNumber: customer.identityNumber,
    contact: customer.contact,
    email: customer.email,
    address: customer.address,
  }
}

// Prospek hanya punya satu kolom kontak. Kalau isinya alamat email, dipindah
// ke kolom email pelanggan, selain itu dianggap nomor telepon.
export function leadToCustomerForm(lead: Lead): CustomerFormValues {
  const isEmail = lead.contact.includes('@')
  return {
    ...EMPTY_CUSTOMER_FORM,
    name: lead.name,
    contact: isEmail ? '' : lead.contact,
    email: isEmail ? lead.contact : '',
  }
}

export interface LeadFormValues {
  name: string
  contact: string
  projectId: string
  source: string
  stage: LeadStage
  lastContactedAt: string
}

export const EMPTY_LEAD_FORM: LeadFormValues = {
  name: '',
  contact: '',
  projectId: '',
  source: '',
  stage: 'new',
  lastContactedAt: '',
}

export function leadToForm(lead: Lead): LeadFormValues {
  return {
    name: lead.name,
    contact: lead.contact,
    projectId: lead.projectId ?? '',
    source: lead.source,
    stage: lead.stage,
    lastContactedAt: dateInput(lead.lastContactedAt),
  }
}

export function toLeadRequest(values: LeadFormValues): LeadRequest {
  return {
    name: values.name.trim(),
    contact: values.contact.trim(),
    projectId: values.projectId === '' ? null : values.projectId,
    source: values.source.trim(),
    stage: values.stage,
    lastContactedAt: values.lastContactedAt === '' ? null : toApiDate(values.lastContactedAt),
  }
}

// PUT prospek butuh semua kolom, jadi pindah tahap mengirim ulang data lama
// dengan tahap yang baru.
export function leadWithStage(lead: Lead, stage: LeadStage): LeadRequest {
  return {
    name: lead.name,
    contact: lead.contact,
    projectId: lead.projectId,
    source: lead.source,
    stage,
    lastContactedAt: lead.lastContactedAt,
  }
}

export interface ContractFormValues {
  number: string
  customerId: string
  unitId: string
  type: ContractType
  value: string
  date: string
}

export function emptyContractForm(): ContractFormValues {
  return {
    number: '',
    customerId: '',
    unitId: '',
    type: 'installment',
    value: '',
    date: todayInput(),
  }
}

export function toContractRequest(values: ContractFormValues): ContractRequest {
  return {
    number: values.number.trim(),
    customerId: values.customerId,
    unitId: values.unitId,
    type: values.type,
    value: toAmount(values.value),
    date: toApiDate(values.date),
  }
}

export function contractToForm(contract: Contract): ContractFormValues {
  return {
    number: contract.number,
    customerId: contract.customerId,
    unitId: contract.unitId,
    type: contract.type,
    value: String(contract.value),
    date: dateInput(contract.date),
  }
}

export interface InstallmentFormValues {
  installmentNumber: string
  dueDate: string
  amount: string
}

export const EMPTY_INSTALLMENT_FORM: InstallmentFormValues = { installmentNumber: '', dueDate: '', amount: '' }

export function toInstallmentRequest(values: InstallmentFormValues): InstallmentRequest {
  return {
    installmentNumber: toAmount(values.installmentNumber),
    dueDate: toApiDate(values.dueDate),
    amount: toAmount(values.amount),
  }
}

export function installmentToForm(installment: Installment): InstallmentFormValues {
  return {
    installmentNumber: String(installment.installmentNumber),
    dueDate: dateInput(installment.dueDate),
    amount: String(installment.amount),
  }
}

export const UNIT_STATUS_LABEL: Record<UnitStatus, string> = {
  available: 'Tersedia',
  reserved: 'Dipesan',
  sold: 'Terjual',
  on_hold: 'Ditahan',
}

export const UNIT_STATUS_TONE: Record<UnitStatus, string> = {
  available: 'bg-green-100 text-green-800',
  reserved: 'bg-amber-100 text-amber-800',
  sold: 'bg-blue-100 text-blue-800',
  on_hold: 'bg-slate-100 text-slate-700',
}

export const CONTRACT_TYPE_LABEL: Record<ContractType, string> = {
  cash: 'Tunai',
  mortgage: 'KPR',
  installment: 'Angsuran',
}

export const CONTRACT_STATUS_LABEL: Record<ContractStatus, string> = {
  draft: 'Draf',
  active: 'Aktif',
  paid: 'Lunas',
  cancelled: 'Batal',
}

export const CONTRACT_STATUS_TONE: Record<ContractStatus, string> = {
  draft: 'bg-slate-100 text-slate-700',
  active: 'bg-blue-100 text-blue-800',
  paid: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
}

export const INSTALLMENT_STATUS_LABEL: Record<InstallmentStatus, string> = {
  not_due: 'Belum jatuh tempo',
  due: 'Jatuh tempo 7 hari',
  paid: 'Lunas',
  overdue: 'Terlambat',
}

export const INSTALLMENT_STATUS_TONE: Record<InstallmentStatus, string> = {
  not_due: 'bg-slate-100 text-slate-700',
  due: 'bg-amber-100 text-amber-800',
  paid: 'bg-green-100 text-green-800',
  overdue: 'bg-red-100 text-red-800',
}

export const LEAD_STAGE_LABEL: Record<LeadStage, string> = {
  new: 'Baru',
  interested: 'Tertarik',
  negotiating: 'Negosiasi',
  won: 'Berhasil',
  cancelled: 'Batal',
}

export const LEAD_STAGE_TONE: Record<LeadStage, string> = {
  new: 'bg-slate-100 text-slate-700',
  interested: 'bg-blue-100 text-blue-800',
  negotiating: 'bg-amber-100 text-amber-800',
  won: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
}
