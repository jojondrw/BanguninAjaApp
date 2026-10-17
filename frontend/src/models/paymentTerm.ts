export type PaymentTermKind = 'progress' | 'retention'

export type PaymentTermState = 'planned' | 'ready' | 'claimed' | 'paid'

export const TERM_STATE_LABEL: Record<PaymentTermState, string> = {
  planned: 'Menunggu progres',
  ready: 'Siap diajukan',
  claimed: 'Diajukan',
  paid: 'Lunas',
}

export const TERM_STATE_TONE: Record<PaymentTermState, string> = {
  planned: 'bg-slate-100 text-slate-700',
  ready: 'bg-amber-100 text-amber-800',
  claimed: 'bg-navy-100 text-navy-800',
  paid: 'bg-green-100 text-green-800',
}

// Batas mengikuti tag binding ContractRequest dan PaymentTermRequest di backend.
export const CONTRACT_NUMBER_MAX_LENGTH = 60
export const RETENTION_MAX_PERCENT = 10
export const DEFAULT_RETENTION_PERCENT = 5
const FULL_PERCENT = 100
const NEXT_TERM_PROGRESS_STEP = 25

export interface ConstructionContract {
  vendorId: string
  vendorName: string
  paymentTermDays: number
  number: string
  value: number
  currentValue: number
  approvedChange: number
  extensionDays: number
  retentionPercent: number
}

export type ContractChangeStatus = 'proposed' | 'approved' | 'rejected'

export const CHANGE_STATUS_LABEL: Record<ContractChangeStatus, string> = {
  proposed: 'Diusulkan',
  approved: 'Disetujui',
  rejected: 'Ditolak',
}

export const CHANGE_STATUS_TONE: Record<ContractChangeStatus, string> = {
  proposed: 'bg-amber-100 text-amber-800',
  approved: 'bg-green-100 text-green-800',
  rejected: 'bg-slate-100 text-slate-700',
}

// Batas mengikuti tag binding ContractChangeRequest di backend.
export const CHANGE_DESCRIPTION_MAX_LENGTH = 300
export const CHANGE_DAYS_MAX = 3650

export interface ContractChange {
  id: string
  label: string
  description: string
  valueChange: number
  daysChange: number
  status: ContractChangeStatus
  proposerName: string | null
  deciderName: string | null
  decidedOn: string | null
  createdAt: string
}

export type ChangeKind = 'add' | 'deduct'

export interface ContractChangeFormValues {
  kind: ChangeKind
  amount: string
  days: string
  description: string
}

export interface ContractChangeRequest {
  description: string
  valueChange: number
  daysChange: number
}

export const EMPTY_CHANGE_FORM: ContractChangeFormValues = { kind: 'add', amount: '', days: '', description: '' }

// Pekerjaan kurang dikirim sebagai nilai negatif; pengguna cukup memilih jenis
// dan mengisi angka positif supaya tidak salah tanda.
export function toContractChangeRequest(values: ContractChangeFormValues): ContractChangeRequest {
  const amount = values.amount === '' ? 0 : Math.abs(Number(values.amount))
  return {
    description: values.description.trim(),
    valueChange: values.kind === 'deduct' ? -amount : amount,
    daysChange: values.days === '' ? 0 : Number(values.days),
  }
}

export function signedRupiahShort(amount: number, format: (value: number) => string): string {
  return amount > 0 ? `+${format(amount)}` : format(amount)
}

// Siapa dan kapan, dibaca dari kiri ke kanan seperti jejak persetujuan.
export function changeTrail(change: ContractChange, dateText: (iso: string) => string): string {
  const proposed = `diusulkan ${change.proposerName ?? 'pengguna terhapus'}, ${dateText(change.createdAt)}`
  if (change.status === 'proposed' || !change.decidedOn) {
    return proposed
  }
  const verb = change.status === 'approved' ? 'disetujui' : 'ditolak'
  return `${proposed}; ${verb} ${change.deciderName ?? 'pengguna terhapus'}, ${dateText(change.decidedOn)}`
}

export interface PaymentTerm {
  id: string
  label: string
  kind: PaymentTermKind
  state: PaymentTermState
  requiredProgress: number
  percent: number
  grossAmount: number
  retentionAmount: number
  netAmount: number
  paidAmount: number
  payableId: string | null
  claimedOn: string | null
  claimedProgress: number | null
}

export interface PaymentSchedule {
  contract: ConstructionContract | null
  progress: number
  terms: PaymentTerm[]
  changes: ContractChange[]
  plannedPercent: number
  claimedAmount: number
  paidAmount: number
  retentionHeld: number
  retentionReleased: boolean
  contractLocked: boolean
  canReleaseRetention: boolean
}

export interface ContractRequest {
  vendorId: string
  number: string
  value: number
  retentionPercent: number
}

export interface PaymentTermRequest {
  requiredProgress: number
  percent: number
}

export interface ContractFormValues {
  vendorId: string
  number: string
  value: string
  retentionPercent: string
}

export interface PaymentTermFormValues {
  requiredProgress: string
  percent: string
}

export function contractFormValues(contract: ConstructionContract | null): ContractFormValues {
  if (!contract) {
    return { vendorId: '', number: '', value: '', retentionPercent: String(DEFAULT_RETENTION_PERCENT) }
  }
  return {
    vendorId: contract.vendorId,
    number: contract.number,
    value: String(contract.value),
    retentionPercent: String(contract.retentionPercent),
  }
}

export function toContractRequest(values: ContractFormValues): ContractRequest {
  return {
    vendorId: values.vendorId,
    number: values.number.trim(),
    value: Number(values.value),
    retentionPercent: values.retentionPercent === '' ? 0 : Number(values.retentionPercent),
  }
}

export function remainingPercent(schedule: Pick<PaymentSchedule, 'plannedPercent'>): number {
  return Math.max(0, Math.round((FULL_PERCENT - schedule.plannedPercent) * 100) / 100)
}

// Isian awal termin berikutnya: sisa persen kontrak, dan syarat progres di atas
// termin terakhir supaya urutannya langsung masuk akal.
export function nextTermForm(schedule: {
  plannedPercent: number
  terms: Pick<PaymentTerm, 'kind' | 'requiredProgress'>[]
}): PaymentTermFormValues {
  const progressTerms = schedule.terms.filter((term) => term.kind === 'progress')
  const last = progressTerms.at(-1)
  const remaining = remainingPercent(schedule)
  return {
    requiredProgress: last ? String(Math.min(FULL_PERCENT, last.requiredProgress + NEXT_TERM_PROGRESS_STEP)) : '0',
    percent: remaining > 0 ? String(remaining) : '',
  }
}

export function toPaymentTermRequest(values: PaymentTermFormValues): PaymentTermRequest {
  return {
    requiredProgress: Number(values.requiredProgress),
    percent: Number(values.percent),
  }
}

export function formatPercent(value: number): string {
  return `${value.toLocaleString('id-ID', { maximumFractionDigits: 2 })}%`
}

export function shareOfContract(part: number, whole: number): string {
  return whole > 0 ? formatPercent(Math.round((part / whole) * 1000) / 10) : '0%'
}

// Kalimat di bawah status termin. Termin yang sudah diajukan memakai progres
// saat diajukan, bukan progres sekarang, karena itulah dasar pembayarannya.
export function termNote(term: PaymentTerm, progress: number, claimedText: (iso: string) => string): string {
  if (term.claimedOn) {
    const at = term.claimedProgress === null ? '' : ` pada progres ${term.claimedProgress}%`
    return `diajukan ${claimedText(term.claimedOn)}${at}`
  }
  if (term.state === 'planned' && progress < term.requiredProgress) {
    return `progres sekarang ${progress}%`
  }
  if (term.state === 'planned') {
    return 'setelah termin sebelumnya'
  }
  return ''
}
