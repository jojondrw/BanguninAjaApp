import type { PaymentTerm, PaymentTermState } from './paymentTerm'
import { DEFAULT_RETENTION_PERCENT } from './paymentTerm'

// Tagihan termin ke pemberi kerja memakai hitungan yang sama dengan termin
// kontraktor, tetapi setiap termin yang ditagih menjadi piutang, bukan hutang.

export const OWNER_STATE_LABEL: Record<PaymentTermState, string> = {
  planned: 'Menunggu progres',
  ready: 'Siap ditagih',
  claimed: 'Ditagih',
  paid: 'Lunas',
}

export const DEFAULT_OWNER_TERM_DAYS = 30
export const OWNER_TERM_DAYS_MAX = 180

export interface OwnerContract {
  customerId: string
  customerName: string
  number: string
  value: number
  retentionPercent: number
  paymentTermDays: number
}

export type OwnerTerm = Omit<PaymentTerm, 'payableId'> & { receivableId: string | null }

export interface OwnerSchedule {
  contract: OwnerContract | null
  progress: number
  terms: OwnerTerm[]
  plannedPercent: number
  claimedAmount: number
  paidAmount: number
  retentionHeld: number
  retentionReleased: boolean
  contractLocked: boolean
  canReleaseRetention: boolean
}

export interface OwnerContractRequest {
  customerId: string
  number: string
  value: number
  retentionPercent: number
  paymentTermDays: number
}

export interface OwnerContractFormValues {
  customerId: string
  number: string
  value: string
  retentionPercent: string
  paymentTermDays: string
}

// Kontrak baru diisi nilai kontrak proyek, karena itulah nilai yang disepakati
// dengan pemberi kerja.
export function ownerContractFormValues(contract: OwnerContract | null, projectValue: number): OwnerContractFormValues {
  if (!contract) {
    return {
      customerId: '',
      number: '',
      value: projectValue > 0 ? String(projectValue) : '',
      retentionPercent: String(DEFAULT_RETENTION_PERCENT),
      paymentTermDays: String(DEFAULT_OWNER_TERM_DAYS),
    }
  }
  return {
    customerId: contract.customerId,
    number: contract.number,
    value: String(contract.value),
    retentionPercent: String(contract.retentionPercent),
    paymentTermDays: String(contract.paymentTermDays),
  }
}

export function toOwnerContractRequest(values: OwnerContractFormValues): OwnerContractRequest {
  return {
    customerId: values.customerId,
    number: values.number.trim(),
    value: Number(values.value),
    retentionPercent: values.retentionPercent === '' ? 0 : Number(values.retentionPercent),
    paymentTermDays: values.paymentTermDays === '' ? 0 : Number(values.paymentTermDays),
  }
}

export function ownerTermNote(term: OwnerTerm, progress: number, claimedText: (iso: string) => string): string {
  if (term.claimedOn) {
    const at = term.claimedProgress === null ? '' : ` pada progres ${term.claimedProgress}%`
    return `ditagih ${claimedText(term.claimedOn)}${at}`
  }
  if (term.state === 'planned' && progress < term.requiredProgress) {
    return `progres sekarang ${progress}%`
  }
  if (term.state === 'planned') {
    return 'setelah termin sebelumnya'
  }
  return ''
}
