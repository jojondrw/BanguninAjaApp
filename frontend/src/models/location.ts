import type { SavedLocation } from './erpApi'

export interface SiteSummary {
  count: number
  best: SavedLocation | null
  latest: SavedLocation | null
  averageScore: number
}

export function rankSites(sites: SavedLocation[]): SavedLocation[] {
  return [...sites].sort((a, b) => b.score - a.score || a.name.localeCompare(b.name))
}

export function summarizeSites(sites: SavedLocation[]): SiteSummary {
  if (sites.length === 0) {
    return { count: 0, best: null, latest: null, averageScore: 0 }
  }

  const total = sites.reduce((sum, site) => sum + site.score, 0)
  const latest = sites.reduce((newest, site) => (site.savedAt > newest.savedAt ? site : newest))

  return {
    count: sites.length,
    best: rankSites(sites)[0],
    latest,
    averageScore: Math.round(total / sites.length),
  }
}
