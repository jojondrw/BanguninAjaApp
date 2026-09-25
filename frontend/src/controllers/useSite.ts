import { useMutation, useQuery } from '@tanstack/react-query'

import { siteApi } from '../models/siteApi'
import type { EvaluateRequest } from '../models/site'

export function useBuildingProfiles() {
  return useQuery({
    queryKey: ['building-profiles'],
    queryFn: () => siteApi.buildingProfiles(),
    staleTime: 5 * 60_000,
    retry: 0,
  })
}

export function useEvaluateSite() {
  return useMutation({
    mutationFn: (payload: EvaluateRequest) => siteApi.evaluate(payload),
  })
}
