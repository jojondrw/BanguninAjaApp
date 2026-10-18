// Jejak bukti: sidik jari laporan harian, foto lapangan, progres, tagihan
// termin, dan pembayaran (semua metode) dicatat ke blockchain sebagai satu
// akar Merkle per kumpulan. Bentuknya mengikuti internal/evidence di backend.

export type EvidenceKind = 'daily_report' | 'site_photo' | 'progress' | 'owner_claim' | 'payment' | 'wage_root'
export type EvidenceStatus = 'anchored' | 'pending' | 'changed' | 'deleted'

export interface EvidenceItem {
  kind: EvidenceKind
  subjectId: string
  label: string
  date: string | null
  status: EvidenceStatus
  anchoredAt: string | null
  txHash: string
  txUrl: string
}

export interface EvidenceSummary {
  anchored: number
  pending: number
  changed: number
  deleted: number
}

export interface ClaimApproval {
  termId: string
  token: string
  signer: string
  approvedAt: string | null
  stillSame: boolean
}

export interface ProjectEvidence {
  items: EvidenceItem[]
  summary: EvidenceSummary
  evidenceRoot: string
  approverAddress: string
  claims: ClaimApproval[]
  chainId: number
}

export interface EvidenceBatch {
  id: string
  root: string
  calldata: string
  recordCount: number
  chainId: number
  txHash: string
  txUrl: string
  anchoredAt: string | null
  createdAt: string
}

export interface AnchorResult {
  status: 'anchored' | 'waiting'
  message: string
  confirmations: number
  batch: EvidenceBatch
}

export interface TypedData {
  types: Record<string, { name: string; type: string }[]>
  primaryType: string
  domain: { name: string; version: string; chainId: number }
  message: Record<string, string>
}

export interface PublicClaim {
  projectCode: string
  projectName: string
  contractNumber: string
  ownerName: string
  label: string
  progress: number
  grossAmount: number
  retentionAmount: number
  netAmount: number
  claimedOn: string
  items: EvidenceItem[]
  summary: EvidenceSummary
  typedData: TypedData
  approverAddress: string
  approval: ClaimApproval
  explorerUrl: string
}

export const EVIDENCE_KIND_LABEL: Record<EvidenceKind, string> = {
  daily_report: 'Laporan harian',
  site_photo: 'Foto lapangan',
  progress: 'Progres proyek',
  owner_claim: 'Tagihan termin',
  payment: 'Pembayaran',
  wage_root: 'Akar penggajian',
}

export const EVIDENCE_STATUS_LABEL: Record<EvidenceStatus, string> = {
  anchored: 'Tercatat, utuh',
  pending: 'Belum dicatat',
  changed: 'Berubah setelah dicatat',
  deleted: 'Dihapus setelah dicatat',
}

export const EVIDENCE_STATUS_TONE: Record<EvidenceStatus, string> = {
  anchored: 'bg-green-100 text-green-800',
  pending: 'bg-slate-100 text-slate-700',
  changed: 'bg-red-100 text-red-700',
  deleted: 'bg-red-100 text-red-700',
}

export function verifyLink(origin: string, token: string): string {
  return `${origin}/verifikasi/${encodeURIComponent(token)}`
}

export function tamperedCount(summary: EvidenceSummary): number {
  return summary.changed + summary.deleted
}
