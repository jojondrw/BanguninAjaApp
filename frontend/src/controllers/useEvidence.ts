import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { evidenceApi, verifyApi } from '../models/evidenceApi'
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
