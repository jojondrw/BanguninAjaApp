import { Link } from 'react-router-dom'

import { usePortfolio } from '../../controllers/useErp'
import { PROJECT_STATUS_LABEL, scheduleGaps, type ProjectStatus } from '../../models/project'
import { number } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { DonutChart } from '../components/DonutChart'
import { LinkButton } from '../components/LinkButton'
import { ProjectTypeIcon } from './ProjectTypeIcon'

const STATUS_ORDER: ProjectStatus[] = ['ongoing', 'planning', 'on_hold', 'completed', 'cancelled']
const STATUS_COLORS = ['#1f5fd1', '#94a3b8', '#fbbb16', '#16a34a', '#dc2626']
const TYPE_COLORS = ['#0b2b6b', '#1f5fd1', '#5b86d9', '#0f766e', '#14b8a6', '#fbbb16', '#cbd5e1']
const SCHEDULE_ROWS = 8
const TOP_TYPES = 6

function GapBadge({ gap }: { gap: number }) {
  if (Math.abs(gap) < 3) {
    return <span className="text-xs font-medium text-slate-500">sesuai rencana</span>
  }
  const behind = gap < 0
  return (
    <span className={`text-xs font-medium tabular-nums ${behind ? 'text-red-600' : 'text-green-700'}`}>
      {behind ? '' : '+'}
      {number(gap)} poin
    </span>
  )
}

// Ringkasan portofolio: komposisi status, proyek yang paling tertinggal dari
// rencana, dan nilai kontrak per jenis bangunan.
export function PortfolioOverview({ onOpenList }: { onOpenList: () => void }) {
  const portfolio = usePortfolio()

  if (portfolio.isPending) {
    return <Loading />
  }
  if (portfolio.isError) {
    return <LoadFailed onRetry={() => portfolio.refetch()} />
  }
  const { projects, byType } = portfolio.data
  if (projects.length === 0) {
    return <Empty message="Belum ada proyek." />
  }
  const gaps = scheduleGaps(projects)
  const behind = gaps.filter((project) => project.gap <= -3).length
  const types = [...byType].sort((a, b) => b.contractValue - a.contractValue)
  const shownTypes = types.slice(0, TOP_TYPES).map((type) => ({ label: type.type || 'Lainnya', value: type.contractValue }))
  const otherTypes = types.slice(TOP_TYPES).reduce((sum, type) => sum + type.contractValue, 0)

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
      <Card
        title="Jadwal proyek berjalan"
        description={`${number(behind)} dari ${number(gaps.length)} proyek tertinggal dari rencana hari ini. Garis tipis adalah rencana.`}
        action={<LinkButton label="Semua proyek" onClick={onOpenList} />}
      >
        <ul className="divide-y divide-slate-100">
          {gaps.slice(0, SCHEDULE_ROWS).map((project) => (
            <li key={project.id}>
              <Link to={`/proyek/${project.id}`} className="grid gap-1.5 py-2.5 hover:bg-slate-50/60">
                <span className="flex items-center justify-between gap-3 text-[13px]">
                  <span className="flex min-w-0 items-center gap-1.5">
                    <ProjectTypeIcon type={project.type} className="size-3.5 shrink-0 text-navy-600" />
                    <span className="truncate font-medium text-slate-900">{project.name}</span>
                  </span>
                  <GapBadge gap={project.gap} />
                </span>
                <span aria-hidden="true" className="relative block h-2 rounded-full bg-slate-100">
                  <span
                    className={`block h-full rounded-full ${project.gap <= -3 ? 'bg-red-400' : 'bar-fill'}`}
                    style={{ width: `${project.progress}%` }}
                  />
                  <span
                    className="absolute -top-1 h-4 w-0.5 rounded-full bg-slate-900"
                    style={{ left: `calc(${Math.min(100, project.planned ?? 0)}% - 1px)` }}
                  />
                </span>
                <span className="text-xs text-slate-500 tabular-nums">
                  progres {number(project.progress)}%, rencana {number(project.planned ?? 0)}%
                  {project.status === 'on_hold' ? ', tertunda' : ''}
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </Card>

      <div className="grid content-start gap-6">
        <Card title="Status proyek" action={<LinkButton label="Daftar" onClick={onOpenList} />}>
          <DonutChart
            caption="Jumlah proyek per status"
            colors={STATUS_COLORS}
            slices={STATUS_ORDER.map((status) => ({
              label: PROJECT_STATUS_LABEL[status],
              value: projects.filter((project) => project.status === status).length,
            }))}
            format={(value) => `${number(value)} proyek`}
          />
        </Card>
        <Card title="Nilai kontrak per jenis" description="Di luar proyek yang batal">
          <DonutChart
            caption="Nilai kontrak per jenis proyek"
            colors={TYPE_COLORS}
            slices={otherTypes > 0 ? [...shownTypes, { label: 'Lainnya', value: otherTypes }] : shownTypes}
          />
        </Card>
      </div>
    </div>
  )
}
