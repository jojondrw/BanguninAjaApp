import { request } from '../shared/apiClient'
import type { AnchorResult, ClaimApproval, EvidenceBatch, ProjectEvidence, PublicClaim } from './evidence'
import { publicCall } from './onchainApi'

export const evidenceApi = {
  project: (projectId: string) => request<ProjectEvidence>(`/evidence/projects/${projectId}`),

  batches: () => request<EvidenceBatch[]>('/evidence/batches'),

  createBatch: () => request<EvidenceBatch>('/evidence/batches', { method: 'POST' }),

  anchor: (batchId: string, txHash: string) =>
    request<AnchorResult>(`/evidence/batches/${batchId}/anchor`, { method: 'POST', body: { txHash } }),

  setApprover: (projectId: string, address: string) =>
    request<{ address: string }>(`/evidence/projects/${projectId}/approver`, { method: 'PUT', body: { address } }),

  createClaimLink: (projectId: string, termId: string) =>
    request<ClaimApproval>(`/evidence/projects/${projectId}/claims/${termId}`, { method: 'POST' }),
}

// Halaman verifikasi dibuka pemberi kerja tanpa akun.
export const verifyApi = {
  claim: (token: string) => publicCall<PublicClaim>(`/verify/claims/${encodeURIComponent(token)}`),

  approve: (token: string, signature: string) =>
    publicCall<ClaimApproval>(`/verify/claims/${encodeURIComponent(token)}/approve`, { signature }),
}
