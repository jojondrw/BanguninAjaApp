import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { evidenceApi, verifyApi } from '../models/evidenceApi'
import { wageProofApi } from '../models/wageProof'
import { DATA_QUERY } from './useErp'

const EVIDENCE_KEY = 'evidence'

export function useProjectEvidence(projectId: string) {
  return useQuery({
    queryKey: [EVIDENCE_KEY, 'project', projectId],
    queryFn: () => evidenceApi.project(projectId),
    ...DATA_QUERY,
  })
}

function useInvalidateEvidence() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: [EVIDENCE_KEY] })
}

export function useCreateEvidenceBatch() {
  return useMutation({ mutationFn: evidenceApi.createBatch })
}

export function useEvidenceConfig() {
  return useQuery({ queryKey: [EVIDENCE_KEY, 'config'], queryFn: evidenceApi.config, staleTime: 5 * 60_000, retry: 0 })
}

export function useServerAnchor() {
  return useMutation({ mutationFn: evidenceApi.serverAnchor })
}

export function useAnchorBatch() {
  const invalidate = useInvalidateEvidence()
  return useMutation({
    mutationFn: ({ batchId, txHash }: { batchId: string; txHash: string }) => evidenceApi.anchor(batchId, txHash),
    onSuccess: (result) => (result.status === 'anchored' ? invalidate() : undefined),
  })
}

export function useSetApprover(projectId: string) {
  const invalidate = useInvalidateEvidence()
  return useMutation({
    mutationFn: (address: string) => evidenceApi.setApprover(projectId, address),
    onSuccess: invalidate,
  })
}

export function useCreateClaimLink(projectId: string) {
  const invalidate = useInvalidateEvidence()
  return useMutation({
    mutationFn: (termId: string) => evidenceApi.createClaimLink(projectId, termId),
    onSuccess: invalidate,
  })
}

export function usePublicClaim(token: string) {
  return useQuery({
    queryKey: [EVIDENCE_KEY, 'public', token],
    queryFn: () => verifyApi.claim(token),
    retry: false,
  })
}

export function useApproveClaim(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (signature: string) => verifyApi.approve(token, signature),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [EVIDENCE_KEY, 'public', token] }),
  })
}

const WAGE_KEY = 'wage-proofs'

export function useWageProofs(projectId: string) {
  return useQuery({
    queryKey: [EVIDENCE_KEY, WAGE_KEY, projectId],
    queryFn: () => wageProofApi.list(projectId),
    ...DATA_QUERY,
  })
}

export function useCreateWageProof(projectId: string) {
  const invalidate = useInvalidateEvidence()
  return useMutation({
    mutationFn: (period: string) => wageProofApi.create(projectId, period),
    onSuccess: (result) => (result.proof ? invalidate() : undefined),
  })
}

export function usePublicWageProof(token: string) {
  return useQuery({ queryKey: [EVIDENCE_KEY, 'public-wage', token], queryFn: () => wageProofApi.publicProof(token), retry: false })
}

export function usePublicReceipt(token: string) {
  return useQuery({ queryKey: [EVIDENCE_KEY, 'receipt', token], queryFn: () => wageProofApi.publicReceipt(token), retry: false })
}
