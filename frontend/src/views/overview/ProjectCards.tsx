import { Link } from 'react-router-dom'

import type { Project } from '../../models/project'
import { rupiahShort, shortDate } from '../../shared/format'
import { Money } from '../components/Money'
import { Ring } from '../components/Ring'
import { StatusChip } from '../project/StatusChip'

export function ProjectCards({ projects }: { projects: Project[] }) {
  return (
    <ul className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {projects.map((project) => (
        <li key={project.id} className="min-w-0">
          <Link
            to={`/proyek/${project.id}`}
            className="flex h-full items-center gap-4 rounded-2xl bg-slate-50/70 p-4 transition-colors hover:bg-navy-50/60"
          >
            <Ring value={project.progress} size={64} label={`${project.progress}%`} />
            <div className="min-w-0 flex-1 space-y-1">
              <p className="truncate text-[11px] font-medium text-slate-500">{project.code}</p>
              <p className="line-clamp-2 text-[14px] leading-5 font-semibold text-slate-900">{project.name}</p>
              <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                <StatusChip project={project} />
                <span className="text-xs text-slate-500">
                  <Money>{rupiahShort(project.contractValue)}</Money>
                </span>
              </div>
              <p className="text-xs text-slate-500">Target {shortDate(project.targetEndDate)}</p>
            </div>
          </Link>
        </li>
      ))}
    </ul>
  )
}
