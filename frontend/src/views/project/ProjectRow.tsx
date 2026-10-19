import { Camera, ChevronRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useSitePhotoUrl } from '../../controllers/useSitePhotos'
import type { Project } from '../../models/project'
import { rupiahShort, shortDate } from '../../shared/format'
import { Money } from '../components/Money'
import { ProjectTypeIcon } from './ProjectTypeIcon'
import { StatusChip } from './StatusChip'

// Baris tampilan daftar: foto lapangan di kiri, keterangan proyek di kanan.
// Lebih rapat dari kartu, jadi banyak proyek bisa dibandingkan sekali gulir.
export function ProjectRow({ project }: { project: Project }) {
  const photoUrl = useSitePhotoUrl(project.id, project.coverPhotoId)
  const progress = Math.max(0, Math.min(100, Math.round(project.progress)))

  return (
    <Link
      to={`/proyek/${project.id}`}
      className="group flex items-center gap-4 rounded-2xl bg-white p-2.5 shadow-hairline transition-shadow
                 hover:shadow-[0_0_0_1px_rgb(0_0_0/0.06),0_8px_22px_rgb(15_23_42/0.10)] sm:gap-5"
    >
      <div className="relative aspect-[4/3] w-28 shrink-0 self-center overflow-hidden rounded-xl bg-navy-900 sm:w-44">
        {photoUrl ? (
          <img
            src={photoUrl}
            alt={`Kondisi lapangan ${project.name}`}
            className="size-full object-cover transition-transform duration-500 group-hover:scale-[1.04] motion-safe:animate-fade"
          />
        ) : project.coverPhotoId ? (
          <div aria-hidden="true" className="size-full animate-pulse bg-slate-200" />
        ) : (
          <div className="grid size-full place-items-center bg-gradient-to-br from-navy-600 via-navy-700 to-navy-900 text-white/85">
            <div className="flex flex-col items-center gap-1">
              <ProjectTypeIcon type={project.type} className="size-7" />
              <span className="inline-flex items-center gap-1 text-[10px] text-white/65">
                <Camera aria-hidden="true" className="size-3" strokeWidth={1.75} />
                Belum ada foto
              </span>
            </div>
          </div>
        )}
      </div>

      <div className="flex min-w-0 flex-1 flex-col justify-center gap-2 py-1 pr-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="inline-flex items-center gap-1 text-[12px] font-medium text-slate-600">
            <ProjectTypeIcon type={project.type} className="size-3.5 shrink-0 text-navy-600" />
            {project.type || 'Proyek'}
          </span>
          <span aria-hidden="true" className="hidden text-slate-300 sm:inline">
            ·
          </span>
          <span className="hidden text-[12px] text-slate-500 sm:inline">{project.code}</span>
          <span className="ml-auto">
            <StatusChip project={project} />
          </span>
        </div>

        <p title={project.name} className="line-clamp-2 text-[15px] leading-5 font-semibold tracking-[-0.01em] text-slate-900 sm:line-clamp-1 sm:text-base">
          {project.name}
        </p>

        <div className="flex items-center gap-2.5">
          <span
            role="progressbar"
            aria-label={`Progres ${project.name}`}
            aria-valuenow={progress}
            aria-valuemin={0}
            aria-valuemax={100}
            className="block h-1.5 max-w-xs flex-1 overflow-hidden rounded-[2px] bg-slate-100"
          >
            <span className="bar-fill block h-full rounded-[2px]" style={{ width: `${progress}%` }} />
          </span>
          <span className="w-9 text-xs font-medium tabular-nums text-slate-700">{progress}%</span>
        </div>

        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500">
          <span>Target {shortDate(project.targetEndDate)}</span>
          <span>
            Nilai <Money>{rupiahShort(project.contractValue)}</Money>
          </span>
        </div>
      </div>

      <ChevronRight
        aria-hidden="true"
        className="hidden size-4 shrink-0 self-center text-slate-300 transition-transform group-hover:translate-x-0.5 group-hover:text-slate-500 sm:block"
        strokeWidth={2}
      />
    </Link>
  )
}
