import { Camera } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useSitePhotoUrl } from '../../controllers/useSitePhotos'
import type { Project } from '../../models/project'
import { rupiahShort, shortDate } from '../../shared/format'
import { Money } from '../components/Money'
import { ProjectTypeIcon } from './ProjectTypeIcon'
import { StatusChip } from './StatusChip'

// Kartu proyek dengan foto lapangan terbaru memenuhi seluruh kartu. Bagian
// bawahnya memudar halus ke gelap (blur bertahap dan gradasi, tanpa garis tepi),
// dan teks putih di atasnya tetap terbaca di atas foto apa pun.
// Proyek tanpa foto memakai ikon jenisnya supaya tetap mudah dibedakan.
export function ProjectCard({ project }: { project: Project }) {
  const photoUrl = useSitePhotoUrl(project.id, project.coverPhotoId)
  const progress = Math.max(0, Math.min(100, Math.round(project.progress)))

  return (
    <Link
      to={`/proyek/${project.id}`}
      className="group relative isolate flex aspect-[4/5] min-h-[300px] flex-col justify-between overflow-hidden rounded-2xl bg-navy-900
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
          <div className="grid size-full place-items-center bg-gradient-to-br from-navy-600 via-navy-700 to-navy-900 pb-32">
            <div className="flex flex-col items-center gap-2 text-white/85">
              <ProjectTypeIcon type={project.type} className="size-10" />
              <span className="inline-flex items-center gap-1 text-[11px] text-white/65">
                <Camera aria-hidden="true" className="size-3" strokeWidth={1.75} />
                Belum ada foto lapangan
              </span>
            </div>
          </div>
        )}
        <div
          aria-hidden="true"
          className="absolute inset-x-0 bottom-0 h-3/5 backdrop-blur-xl [mask-image:linear-gradient(to_bottom,transparent,black_55%)]"
        />
        <div
          aria-hidden="true"
          className="absolute inset-x-0 bottom-0 h-3/4 bg-gradient-to-b from-transparent via-black/35 to-black/75"
        />
      </div>

      <div className="flex items-start justify-between gap-2 p-3">
        <span className="inline-flex min-w-0 items-center gap-1.5 rounded-full bg-white/80 px-2.5 py-1 text-[12px] font-medium text-slate-800 shadow-sm backdrop-blur-md">
          <ProjectTypeIcon type={project.type} className="size-3.5 shrink-0 text-navy-600" />
          <span className="truncate">{project.type || 'Proyek'}</span>
        </span>
        <StatusChip project={project} />
      </div>

      <div className="space-y-3 px-4 pb-4 text-white">
        <div className="min-w-0">
          <p className="truncate text-[11px] font-medium text-white/70">{project.code}</p>
          <p
            title={project.name}
            className="mt-0.5 line-clamp-2 min-h-[44px] text-[17px] leading-[22px] font-semibold tracking-[-0.01em] [text-shadow:0_1px_8px_rgb(0_0_0/0.25)]"
          >
            {project.name}
          </p>
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2.5">
            <span
              role="progressbar"
              aria-label={`Progres ${project.name}`}
              aria-valuenow={progress}
              aria-valuemin={0}
              aria-valuemax={100}
              className="block h-1.5 flex-1 overflow-hidden rounded-[2px] bg-white/25"
            >
              <span className="block h-full rounded-[2px] bg-white" style={{ width: `${progress}%` }} />
            </span>
            <span className="w-9 text-right text-xs font-medium tabular-nums">{progress}%</span>
          </div>
          <div className="flex items-center justify-between gap-2 text-xs text-white/75">
            <span>Target {shortDate(project.targetEndDate)}</span>
            <Money>{rupiahShort(project.contractValue)}</Money>
          </div>
        </div>
      </div>
    </Link>
  )
}
