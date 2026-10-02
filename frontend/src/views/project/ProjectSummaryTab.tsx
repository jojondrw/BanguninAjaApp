import { useBudgetItems, usePermits, usePhases } from '../../controllers/useProjectWorkspace'
import {
  summarizePermits,
  summarizePhases,
  type Permit,
  type Project,
  type ProjectPhase,
} from '../../models/project'
import { kpiValue, rupiahShort } from '../../shared/format'
import { Kpi, KpiRow } from '../components/Data'
import { BudgetItemsCard } from './BudgetItemsCard'
import { PermitsCard } from './PermitsCard'
import { PhasesCard } from './PhasesCard'

const NO_PHASES: ProjectPhase[] = []
const NO_PERMITS: Permit[] = []

export function ProjectSummaryTab({ project }: { project: Project }) {
  const phases = usePhases(project.id)
  const permits = usePermits(project.id)
  const budgetItems = useBudgetItems(project.id)

  const phaseSummary = summarizePhases(phases.data ?? NO_PHASES)
  const permitSummary = summarizePermits(permits.data ?? NO_PERMITS)
  const rabTotal = budgetItems.data?.grandTotal ?? 0

  return (
    <div className="space-y-6">
      <KpiRow>
        <Kpi
          label="Progres proyek"
          value={`${project.progress}%`}
          note={phaseSummary.total > 0 ? `rata-rata dari ${phaseSummary.total} tahap` : 'belum ada tahap'}
        />
        <Kpi
          label="Tahap selesai"
          value={kpiValue(phases.isPending, phases.isError, `${phaseSummary.completed} dari ${phaseSummary.total}`)}
          note={`${phaseSummary.inProgress} tahap sedang berjalan`}
        />
        <Kpi
          label="Izin terbit"
          value={kpiValue(permits.isPending, permits.isError, `${permitSummary.issued} dari ${permitSummary.total}`)}
          note={`${permitSummary.submitted} izin masih diajukan`}
        />
        <Kpi
          label="Nilai RAB"
          value={kpiValue(budgetItems.isPending, budgetItems.isError, rupiahShort(rabTotal))}
          note={rabNote(rabTotal, project.contractValue, budgetItems.data?.totalItems ?? 0)}
        />
      </KpiRow>

      <div className="grid gap-6 xl:grid-cols-2">
        <PhasesCard project={project} />
        <PermitsCard project={project} />
      </div>

      <BudgetItemsCard project={project} />
    </div>
  )
}

function rabNote(total: number, contractValue: number, items: number): string {
  if (items === 0) {
    return 'belum ada item RAB'
  }
  if (contractValue <= 0) {
    return `${items} item pekerjaan`
  }
  return `${items} item, ${Math.round((total / contractValue) * 100)}% dari nilai kontrak`
}
