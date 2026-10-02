import { lazy, Suspense, useState } from 'react'

import {
  useBuildingProfiles,
  useProjects,
  useSavedLocations,
  useSiteEvaluate,
} from '../controllers/useErp'
import type { SavedLocation } from '../models/erpApi'
import { AppShell } from './components/AppShell'
import { Button, ErrorNote, Field } from './components/Form'
import { Card, Empty, Kpi, KpiRow, Loading } from './components/Data'
import type { MapPoint } from './components/SiteMap'

const SiteMap = lazy(() =>
  import('./components/SiteMap').then((module) => ({ default: module.SiteMap })),
)

const PROJECT_OPTION_LIMIT = 100
const MAX_LATITUDE = 90
const MAX_LONGITUDE = 180
const NO_SITES: SavedLocation[] = []

const SELECT_CLASS =
  'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:cursor-not-allowed disabled:bg-slate-100'

function toMapPoint(latitude: string, longitude: string): MapPoint | null {
  if (latitude.trim() === '' || longitude.trim() === '') {
    return null
  }
  const point = { latitude: Number(latitude), longitude: Number(longitude) }
  const isValid =
    Number.isFinite(point.latitude) &&
    Number.isFinite(point.longitude) &&
    Math.abs(point.latitude) <= MAX_LATITUDE &&
    Math.abs(point.longitude) <= MAX_LONGITUDE

  return isValid ? point : null
}

function describeSites(shown: number, total: number, isProjectChosen: boolean): string {
  const scope = isProjectChosen ? ' untuk proyek ini' : ''
  if (total === 0) {
    return `Belum ada lokasi tersimpan${scope}.`
  }
  return `Menampilkan ${shown} dari ${total} lokasi tersimpan${scope}.`
}

export function LocationPage() {
  const evaluate = useSiteEvaluate()
  const profiles = useBuildingProfiles()
  const projects = useProjects({ pageSize: PROJECT_OPTION_LIMIT })

  const [name, setName] = useState('')
  const [latitude, setLatitude] = useState('')
  const [longitude, setLongitude] = useState('')
  const [buildingProfileId, setBuildingProfileId] = useState('')
  const [projectId, setProjectId] = useState('')
  const [focus, setFocus] = useState<MapPoint | null>(null)

  const savedLocations = useSavedLocations(projectId || undefined)
  const sites = savedLocations.data?.items ?? NO_SITES
  const picked = toMapPoint(latitude, longitude)
  const result = evaluate.data

  function handlePick(point: MapPoint) {
    setLatitude(String(point.latitude))
    setLongitude(String(point.longitude))
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const target = { latitude: Number(latitude), longitude: Number(longitude) }
    evaluate.mutate(
      {
        name: name.trim(),
        ...target,
        building_profile_id: buildingProfileId,
        project_id: projectId || undefined,
      },
      { onSuccess: () => setFocus(target) },
    )
  }

  return (
    <AppShell
      title="Analisis Lokasi"
      description="Evaluasi kelayakan lokasi dan konteks regulasi serta berita sekitar"
    >
      <div className="space-y-6">
        <Card
          title="Evaluasi lokasi"
          description="Klik peta atau ketik koordinat, lalu pilih profil bangunan untuk menjalankan analisis."
        >
          <form onSubmit={handleSubmit} className="space-y-5">
            <Suspense
              fallback={
                <div className="flex h-80 items-center justify-center rounded-lg border border-slate-200 bg-slate-100 sm:h-96">
                  <p className="text-sm text-slate-500">Memuat peta...</p>
                </div>
              }
            >
              <SiteMap
                picked={picked}
                sites={sites}
                focus={focus}
                onPick={handlePick}
              />
            </Suspense>

            {savedLocations.isPending ? (
              <p className="text-xs text-slate-500">Memuat lokasi tersimpan...</p>
            ) : null}

            {savedLocations.isError ? (
              <ErrorNote message="Gagal memuat lokasi tersimpan di peta." />
            ) : null}

            {savedLocations.data ? (
              <p role="status" className="text-xs text-slate-500">
                {describeSites(sites.length, savedLocations.data.totalItems, projectId !== '')}
              </p>
            ) : null}

            <div className="grid gap-4 md:grid-cols-2">
              <Field
                id="location-name"
                label="Nama lokasi"
                placeholder="Contoh: Lokasi Proyek A"
                value={name}
                onChange={(event) => setName(event.target.value)}
                required
              />

              <div>
                <label
                  htmlFor="building-profile"
                  className="mb-1.5 block text-sm font-medium text-slate-700"
                >
                  Building Profile
                </label>

                <select
                  id="building-profile"
                  value={buildingProfileId}
                  onChange={(event) =>
                    setBuildingProfileId(event.target.value)
                  }
                  required
                  disabled={profiles.isLoading}
                  className={SELECT_CLASS}
                >
                  <option value="">
                    {profiles.isLoading
                      ? 'Memuat building profile...'
                      : 'Pilih building profile'}
                  </option>

                  {profiles.data?.map((profile) => (
                    <option key={profile.id} value={profile.id}>
                      {profile.name} ({profile.code})
                    </option>
                  ))}
                </select>

                {profiles.isError ? (
                  <p className="mt-1.5 text-xs text-red-600">
                    Gagal mengambil building profile.
                  </p>
                ) : null}
              </div>

              <Field
                id="latitude"
                label="Latitude"
                type="number"
                step="any"
                min={-MAX_LATITUDE}
                max={MAX_LATITUDE}
                placeholder="-6.200000"
                value={latitude}
                onChange={(event) => setLatitude(event.target.value)}
                required
              />

              <Field
                id="longitude"
                label="Longitude"
                type="number"
                step="any"
                min={-MAX_LONGITUDE}
                max={MAX_LONGITUDE}
                placeholder="106.816666"
                value={longitude}
                onChange={(event) => setLongitude(event.target.value)}
                required
              />

              <div className="md:col-span-2">
                <label
                  htmlFor="project"
                  className="mb-1.5 block text-sm font-medium text-slate-700"
                >
                  Proyek (opsional)
                </label>

                <select
                  id="project"
                  value={projectId}
                  onChange={(event) => setProjectId(event.target.value)}
                  disabled={projects.isLoading}
                  aria-describedby="project-hint"
                  className={SELECT_CLASS}
                >
                  <option value="">
                    {projects.isLoading ? 'Memuat proyek...' : 'Tanpa proyek'}
                  </option>

                  {projects.data?.items.map((project) => (
                    <option key={project.id} value={project.id}>
                      {project.name} ({project.code})
                    </option>
                  ))}
                </select>

                <p id="project-hint" className="mt-1.5 text-xs text-slate-500">
                  Hasil evaluasi dikaitkan ke proyek ini. Saat proyek dipilih,
                  peta hanya menampilkan lokasi milik proyek itu.
                </p>

                {projects.isError ? (
                  <p role="alert" className="mt-1.5 text-xs text-red-600">
                    Gagal mengambil daftar proyek.
                  </p>
                ) : null}
              </div>
            </div>

            <Button
              type="submit"
              isPending={evaluate.isPending}
              pendingLabel="Menganalisis..."
            >
              Evaluasi lokasi
            </Button>

            {evaluate.isError ? (
              <ErrorNote message={evaluate.error.message} />
            ) : null}
          </form>
        </Card>

        {evaluate.isPending ? (
          <Card title="Analisis">
            <Loading label="Mengambil skor, regulasi, dan konteks berita..." />
          </Card>
        ) : null}

        {result ? (
          <>
            <KpiRow>
              <Kpi
                label="Overall Score"
                value={`${result.predictive.overall_score}`}
                note="Skor kelayakan lokasi"
              />

              <Kpi
                label="Risk Flags"
                value={`${result.predictive.risk_flags.length}`}
                note="Risiko yang terdeteksi"
              />

              <Kpi
                label="Berita"
                value={`${result.descriptive?.news.length ?? 0}`}
                note="Artikel konteks lokasi"
              />

              <Kpi
                label="Status Regulasi"
                value={
                  result.descriptive?.regulasi?.is_simulated
                    ? 'Simulasi'
                    : 'RDTR'
                }
                note="Sumber data regulasi"
              />
            </KpiRow>

            <div className="grid gap-6 lg:grid-cols-2">
              <Card
                title="Skor kelayakan"
                description="Hasil predictive dari Site Intelligence."
              >
                <div className="space-y-4">
                  {result.predictive.dimension_scores.length === 0 ? (
                    <Empty message="Belum ada dimension score." />
                  ) : (
                    result.predictive.dimension_scores.map((dimension) => (
                      <div
                        key={dimension.dimension_code}
                        className="rounded-lg border border-slate-200 p-4"
                      >
                        <div className="flex items-center justify-between gap-4">
                          <p className="text-sm font-medium text-slate-900">
                            {dimension.dimension_code}
                          </p>

                          <p className="text-lg font-semibold tabular-nums text-slate-900">
                            {dimension.value}
                          </p>
                        </div>

                        {dimension.explanation ? (
                          <p className="mt-1 text-xs text-slate-500">
                            {dimension.explanation}
                          </p>
                        ) : null}
                      </div>
                    ))
                  )}

                  {result.predictive.risk_flags.length > 0 ? (
                    <div className="pt-2">
                      <p className="mb-2 text-sm font-semibold text-slate-900">
                        Risk flags
                      </p>

                      <div className="space-y-2">
                        {result.predictive.risk_flags.map((flag) => (
                          <div
                            key={`${flag.code}-${flag.message}`}
                            className="rounded-lg border border-red-200 bg-red-50 p-3"
                          >
                            <p className="text-sm font-medium text-red-800">
                              {flag.code} · {flag.severity}
                            </p>

                            <p className="mt-1 text-xs text-red-700">
                              {flag.message}
                            </p>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </div>
              </Card>

              <Card
                title="Konteks regulasi"
                description="Informasi deskriptif dari T9. Tidak masuk ke scoring."
              >
                {result.descriptive?.regulasi ? (
                  <div className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <p className="text-xs text-slate-500">Zona</p>

                      <p className="mt-1 text-sm font-medium text-slate-900">
                        {result.descriptive.regulasi.zona || '-'}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-slate-500">Status data</p>

                      <p className="mt-1 text-sm font-medium text-slate-900">
                        {result.descriptive.regulasi.is_simulated
                          ? 'Simulasi'
                          : 'RDTR'}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-slate-500">KDB</p>

                      <p className="mt-1 text-sm font-medium text-slate-900">
                        {result.descriptive.regulasi.kdb}%
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-slate-500">KLB</p>

                      <p className="mt-1 text-sm font-medium text-slate-900">
                        {result.descriptive.regulasi.klb}
                      </p>
                    </div>
                  </div>
                ) : (
                  <Empty message="Data regulasi tidak tersedia." />
                )}
              </Card>
            </div>

            <Card
              title="Berita sekitar lokasi"
              description="Konteks berita terbaru. Tidak digunakan dalam scoring."
            >
              {result.descriptive?.news.length ? (
                <div className="space-y-3">
                  {result.descriptive.news.map((article) => (
                    <a
                      key={`${article.url}-${article.title}`}
                      href={article.url}
                      target="_blank"
                      rel="noreferrer"
                      className="block rounded-lg border border-slate-200 p-4 transition hover:bg-slate-50"
                    >
                      <p className="text-sm font-medium text-slate-900">
                        {article.title}
                      </p>

                      <p className="mt-1 text-xs text-slate-500">
                        {article.source}
                        {article.published_at
                          ? ` · ${article.published_at}`
                          : ''}
                      </p>
                    </a>
                  ))}
                </div>
              ) : (
                <Empty message="Belum ada berita relevan untuk konteks lokasi ini." />
              )}
            </Card>
          </>
        ) : null}
      </div>
    </AppShell>
  )
}