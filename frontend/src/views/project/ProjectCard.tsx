import { Camera } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useSitePhotoUrl } from '../../controllers/useSitePhotos'
import type { Project } from '../../models/project'
import { rupiahShort, shortDate } from '../../shared/format'
import { Money } from '../components/Money'
import { ProjectTypeIcon } from './ProjectTypeIcon'
import { StatusChip } from './StatusChip'

// Kartu proyek dengan foto lapangan terbaru sebagai sampul. Proyek tanpa foto
// memakai ikon jenisnya, jadi rumah sakit, ruko, dan perumahan tetap mudah
// dibedakan sekilas.
export function ProjectCard({ project }: { project: Project }) {
  const photoUrl = useSitePhotoUrl(project.id, project.coverPhotoId)
  const progress = Math.max(0, Math.min(100, Math.round(project.progress)))

  return (
    <Link
      to={`/proyek/${project.id}`}
      className="group flex h-full flex-col overflow-hidden rounded-2xl bg-white shadow-hairline transition-shadow
                 hover:shadow-[0_0_0_1px_rgb(0_0_0/0.06),0_10px_28px_rgb(15_23_42/0.10)]"
    >
      <div className="relative aspect-[16/10] overflow-hidden bg-white">
        {photoUrl ? (
          <img
            src={photoUrl}
            alt={`Kondisi lapangan ${project.name}`}
            className="size-full object-cover transition-transform duration-500 group-hover:scale-[1.03] motion-safe:animate-fade"
          />
        ) : project.coverPhotoId ? (
          <div aria-hidden="true" className="size-full animate-pulse bg-slate-100" />
        ) : (
          <div className="grid size-full place-items-center bg-gradient-to-br from-navy-50 via-white to-slate-100">
            <div className="flex flex-col items-center gap-2 text-navy-600/80">
              <ProjectTypeIcon type={project.type} className="size-10" />
              <span className="inline-flex items-center gap-1 text-[11px] text-slate-500">
                <Camera aria-hidden="true" className="size-3" strokeWidth={1.75} />
                Belum ada foto lapangan
              </span>
            </div>
          </div>
        )}
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 bottom-0 h-2/5 bg-gradient-to-b from-white/0 via-white/60 to-white"
        />
        <div className="absolute inset-x-3 top-3 flex items-start justify-between gap-2">
          <span className="inline-flex min-w-0 items-center gap-1.5 rounded-full bg-white/90 px-2.5 py-1 text-[12px] font-medium text-slate-800 shadow-sm backdrop-blur">
            <ProjectTypeIcon type={project.type} className="size-3.5 shrink-0 text-navy-600" />
            <span className="truncate">{project.type || 'Proyek'}</span>
          </span>
          <StatusChip project={project} />
        </div>
      </div>

      <div className="flex flex-1 flex-col gap-3 p-4">
        <div className="min-w-0">
          <p className="truncate text-[11px] font-medium text-slate-500">{project.code}</p>
          <p className="mt-0.5 line-clamp-2 text-[15px] leading-5 font-semibold text-slate-900">{project.name}</p>
        </div>
        <div className="mt-auto space-y-2">
          <div className="flex items-center gap-2.5">
            <span
              role="progressbar"
              aria-label={`Progres ${project.name}`}
              aria-valuenow={progress}
              aria-valuemin={0}
              aria-valuemax={100}
              className="block h-1.5 flex-1 overflow-hidden rounded-[2px] bg-slate-100"
            >
              <span className="bar-fill block h-full rounded-[2px]" style={{ width: `${progress}%` }} />
            </span>
            <span className="w-9 text-right text-xs font-medium text-slate-700 tabular-nums">{progress}%</span>
          </div>
          <div className="flex items-center justify-between gap-2 text-xs text-slate-500">
            <span>Target {shortDate(project.targetEndDate)}</span>
            <Money>{rupiahShort(project.contractValue)}</Money>
          </div>
        </div>
      </div>
    </Link>
  )
}
