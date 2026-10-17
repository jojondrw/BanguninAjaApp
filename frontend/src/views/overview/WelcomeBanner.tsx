import { ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useCan } from '../../controllers/useAccess'
import { useCurrentUser } from '../../controllers/useAuth'
import { PHOTOS, greeting } from '../../shared/photos'

const PHOTO = PHOTOS.lokasi

export function WelcomeBanner() {
  const user = useCurrentUser()
  const canReadLocations = useCan('locations', 'read')
  const canReadProjects = useCan('projects', 'read')
  const firstName = user?.name?.trim().split(/\s+/)[0]
  const salam = greeting(new Date().getHours())

  return (
    <section className="relative mb-8 overflow-hidden rounded-2xl bg-navy-900 text-white">
      <img
        src={PHOTO.src}
        alt=""
        aria-hidden="true"
        decoding="async"
        className="absolute inset-0 size-full object-cover object-[center_45%]"
      />
      <div
        aria-hidden="true"
        className="absolute inset-0 max-sm:bg-[rgb(7_28_70/0.82)] sm:bg-[linear-gradient(to_right,rgb(7_28_70/0.94)_0%,rgb(7_28_70/0.78)_45%,rgb(7_28_70/0.2)_100%)]"
      />

      <div className="relative px-6 py-8 sm:px-8">
        <h2 className="text-[22px] leading-7 font-semibold tracking-[-0.022em]">
          {firstName ? `${salam}, ${firstName}` : salam}
        </h2>
        <p className="mt-1.5 max-w-md text-[14px] text-white/75">
          Mulai dari lokasi berikutnya, atau lanjutkan proyek yang sedang berjalan.
        </p>

        {canReadLocations || canReadProjects ? (
          <div className="mt-5 flex flex-wrap gap-2">
            {canReadLocations ? (
              <Link
                to="/lokasi"
                className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-white px-3.5 text-[13px] font-medium text-navy-900
                           transition-[background-color,transform] hover:bg-navy-50 motion-safe:active:scale-[0.98]"
              >
                Evaluasi lokasi
                <ArrowRight aria-hidden="true" className="size-3.5" strokeWidth={2} />
              </Link>
            ) : null}
            {canReadProjects ? (
              <Link
                to="/proyek"
                className="inline-flex h-9 items-center rounded-lg px-3.5 text-[13px] font-medium text-white ring-1 ring-white/30
                           transition-[background-color,transform] hover:bg-white/10 motion-safe:active:scale-[0.98]"
              >
                Lihat proyek
              </Link>
            ) : null}
          </div>
        ) : null}
      </div>
    </section>
  )
}
