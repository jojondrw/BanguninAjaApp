import type { ProjectUnitSummary } from '../../models/sales'
import { UNIT_SEGMENTS } from './unitSegments'

export function ProjectUnitBar({ summary }: { summary: ProjectUnitSummary }) {
  return (
    <span aria-hidden="true" className="flex h-2 w-full overflow-hidden rounded-full bg-slate-100">
      {UNIT_SEGMENTS.map((segment) =>
        summary[segment.key] > 0 ? (
          <span
            key={segment.key}
            className={`h-full ${segment.color}`}
            style={{ width: `${(summary[segment.key] / Math.max(summary.total, 1)) * 100}%` }}
          />
        ) : null,
      )}
    </span>
  )
}

export function UnitBarLegend() {
  return (
    <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-slate-500">
      {UNIT_SEGMENTS.map((segment) => (
        <span key={segment.key} className="inline-flex items-center gap-1.5">
          <span aria-hidden="true" className={`size-2 rounded-full ${segment.color}`} />
          {segment.label[0].toUpperCase() + segment.label.slice(1)}
        </span>
      ))}
    </div>
  )
}
