import type { ProjectUnitSummary } from '../../models/sales'
import { number } from '../../shared/format'

// Warna mengikuti chip status unit: tersedia hijau, dipesan kuning, terjual
// biru, ditahan abu.
export const UNIT_SEGMENTS: { key: 'sold' | 'reserved' | 'available' | 'onHold'; label: string; color: string }[] = [
  { key: 'sold', label: 'terjual', color: 'bg-navy-600' },
  { key: 'reserved', label: 'dipesan', color: 'bg-amber-400' },
  { key: 'available', label: 'tersedia', color: 'bg-green-500' },
  { key: 'onHold', label: 'ditahan', color: 'bg-slate-300' },
]

export function unitBreakdown(summary: ProjectUnitSummary): string {
  return UNIT_SEGMENTS.filter((segment) => summary[segment.key] > 0)
    .map((segment) => `${number(summary[segment.key])} ${segment.label}`)
    .join(', ')
}
