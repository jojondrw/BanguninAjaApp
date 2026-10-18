import { Camera } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useSitePhotoUrl } from '../../controllers/useSitePhotos'
import type { Project } from '../../models/project'
import { rupiahShort, shortDate } from '../../shared/format'
import { Money } from '../components/Money'
import { ProjectTypeIcon } from './ProjectTypeIcon'
import { StatusChip } from './StatusChip'

// Kartu proyek dengan foto lapangan terbaru memenuhi seluruh kartu. Teksnya ada
// di panel kaca buram di bagian bawah, jadi tetap terbaca di atas foto apa pun.
// Proyek tanpa foto memakai ikon jenisnya supaya tetap mudah dibedakan.
export function ProjectCard({ project }: { project: Project }) {
  const photoUrl = useSitePhotoUrl(project.id, project.coverPhotoId)
  const progress = Math.max(0, Math.min(100, Math.round(project.progress)))

  return (
    <Link
      to={`/proyek/${project.id}`}
      className="group relative isolate flex aspect-[4/5] min-h-[300px] flex-col justify-between overflow-hidden rounded-2xl bg-slate-100
                 shadow-hairline transition-shadow hover:shadow-[0_0_0_1px_rgb(0_0_0/0.06),0_10px_28px_rgb(15_23_42/0.14)]"
    >
      <div className="absolute inset-0 -z-10">
        {photoUrl ? (
          <img
            src={photoUrl}
            alt={`Kondisi lapangan ${project.name}`}
            className="size-full object-cover transition-transform duration-500 group-hover:scale-[1.03] motion-safe:animate-fade"
          />
        ) : project.coverPhotoId ? (
          <div aria-hidden="true" className="size-full animate-pulse bg-slate-200" />
        ) : (
          <div className="grid size-full place-items-center bg-gradient-to-br from-navy-50 via-white to-slate-100 pb-28">
            <div className="flex flex-col items-center gap-2 text-navy-600/80">
              <ProjectTypeIcon type={project.type} className="size-10" />
              <span className="inline-flex items-center gap-1 text-[11px] text-slate-500">
                <Camera aria-hidden="true" className="size-3" strokeWidth={1.75} />
                Belum ada foto lapangan
              </span>
            </div>
          </div>
        )}
      </div>

      <div className="flex items-start justify-between gap-2 p-3">
        <span className="inline-flex min-w-0 items-center gap-1.5 rounded-full bg-white/80 px-2.5 py-1 text-[12px] font-medium text-slate-800 shadow-sm backdrop-blur-md">
          <ProjectTypeIcon type={project.type} className="size-3.5 shrink-0 text-navy-600" />
          <span className="truncate">{project.type || 'Proyek'}</span>
        </span>
        <StatusChip project={project} />
      </div>

      <div className="m-2 space-y-3 rounded-xl bg-white/70 p-3.5 shadow-[0_1px_2px_rgb(0_0_0/0.06)] ring-1 ring-white/50 backdrop-blur-xl backdrop-saturate-150">
        <div className="min-w-0">
          <p className="truncate text-[11px] font-medium text-slate-600">{project.code}</p>
          <p className="mt-0.5 line-clamp-2 text-[15px] leading-5 font-semibold text-slate-900">{project.name}</p>
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2.5">
            <span
              role="progressbar"
              aria-label={`Progres ${project.name}`}
              aria-valuenow={progress}
              aria-valuemin={0}
              aria-valuemax={100}
              className="block h-1.5 flex-1 overflow-hidden rounded-[2px] bg-slate-900/10"
            >
              <span className="bar-fill block h-full rounded-[2px]" style={{ width: `${progress}%` }} />
            </span>
            <span className="w-9 text-right text-xs font-medium text-slate-800 tabular-nums">{progress}%</span>
          </div>
          <div className="flex items-center justify-between gap-2 text-xs text-slate-600">
            <span>Target {shortDate(project.targetEndDate)}</span>
            <Money>{rupiahShort(project.contractValue)}</Money>
          </div>
        </div>
      </div>
    </Link>
  )
}
