import { request } from '../shared/apiClient'
import { publicCall } from './onchainApi'

// Bukti zero-knowledge bahwa semua pekerja proyek dibayar paling sedikit upah
// minimum kota proyek, tanpa membuka gaji siapa pun. Bentuknya mengikuti
// internal/wageproof di backend.

export interface AnchorStatus {
  status: 'anchored' | 'pending' | 'changed' | 'deleted'
  anchoredAt: string | null
  txHash: string
  txUrl: string
}

export interface ReceiptLink {
  name: string
  token: string
}

export interface WageProof {
  id: string
  period: string
  regionName: string
  minimumWage: number
  wageSource: string
  workerCount: number
  root: string
  token: string
  receipts: ReceiptLink[]
  createdAt: string
}

export interface BelowMinimum {
  name: string
  pay: number
  days: number
  required: number
}

export interface WageProofResult {
  proof: WageProof | null
  below: BelowMinimum[]
}

export interface PublicWageProof {
  projectCode: string
  projectName: string
  period: string
  regionName: string
  minimumWage: number
  dailyMinimum: number
  wageSource: string
  workerCount: number
  root: string
  verified: boolean
  anchor: AnchorStatus
  proofHex: string
  verifyingKeyHex: string
}

export interface PublicReceipt {
  name: string
  projectCode: string
  projectName: string
  period: string
  pay: number
  days: number
  leafIndex: number
  salt: string
  path: string[]
  root: string
  included: boolean
  proofToken: string
  anchor: AnchorStatus
}

export function wageProofLink(origin: string, token: string): string {
  return `${origin}/verifikasi-upah/${encodeURIComponent(token)}`
}

export function receiptLink(origin: string, token: string): string {
  return `${origin}/kuitansi-upah/${encodeURIComponent(token)}`
}

// Bulan-bulan terakhir yang bisa dipilih, terbaru dulu, dalam format YYYY-MM.
export function recentPeriods(today: Date, count: number): string[] {
  return Array.from({ length: count }, (_, offset) => {
    const date = new Date(today.getFullYear(), today.getMonth() - offset, 1)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
  })
}

export function periodLabel(period: string): string {
  const [year, month] = period.split('-').map(Number)
  return new Date(year, month - 1, 1).toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
}

export const wageProofApi = {
  list: (projectId: string) => request<WageProof[]>(`/evidence/projects/${projectId}/wage-proofs`),
  create: (projectId: string, period: string) =>
    request<WageProofResult>(`/evidence/projects/${projectId}/wage-proofs`, { method: 'POST', body: { period } }),
  publicProof: (token: string) => publicCall<PublicWageProof>(`/verify/wages/${encodeURIComponent(token)}`),
  publicReceipt: (token: string) => publicCall<PublicReceipt>(`/verify/receipts/${encodeURIComponent(token)}`),
}
