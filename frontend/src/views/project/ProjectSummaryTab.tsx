import { useBudgetItems, usePermits, usePhases, useProgressCurve } from '../../controllers/useProjectWorkspace'
import {
  progressBasis,
  scheduleStatus,
  summarizePermits,
  summarizePhases,
  type Permit,
  type ProgressCurve,
  type Project,
  type ProjectPhase,
} from '../../models/project'
import { kpiValue, rupiahShort } from '../../shared/format'
import { Kpi, KpiRow } from '../components/Data'
import { BudgetItemsCard } from './BudgetItemsCard'
import { PermitsCard } from './PermitsCard'
import { PhasesCard } from './PhasesCard'
import { ProgressCurveCard } from './ProgressCurveCard'

const NO_PHASES: ProjectPhase[] = []
const NO_PERMITS: Permit[] = []

export function ProjectSummaryTab({ project }: { project: Project }) {
  const phases = usePhases(project.id)
  const permits = usePermits(project.id)
  const budgetItems = useBudgetItems(project.id)
  const curve = useProgressCurve(project.id)

  const phaseSummary = summarizePhases(phases.data ?? NO_PHASES)
  const permitSummary = summarizePermits(permits.data ?? NO_PERMITS)
  const rabTotal = budgetItems.data?.grandTotal ?? 0

  return (
    <div className="space-y-8">
      <KpiRow>
        <Kpi
          label="Progres proyek"
          value={`${project.progress}%`}
          note={progressNote(curve.data, phaseSummary.total)}
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

      <ProgressCurveCard project={project} />

      <div className="grid gap-8 2xl:grid-cols-2">
        <PhasesCard project={project} />
        <PermitsCard project={project} />
      </div>

      <BudgetItemsCard project={project} />
    </div>
  )
}

function progressNote(curve: ProgressCurve | undefined, phaseCount: number): string {
  const basis = progressBasis(curve, phaseCount)
  if (!curve || curve.points.length === 0) {
    return basis
  }
  return `${basis}, ${scheduleStatus(curve.today.deviation).label}`
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
