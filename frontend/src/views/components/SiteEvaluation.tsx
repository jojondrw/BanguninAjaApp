import { lazy, Suspense, useState } from 'react'

import {
  useBuildingProfiles,
  useProjects,
  useSavedLocations,
  useSiteEvaluate,
} from '../../controllers/useErp'
import type { SavedLocation, SiteEvaluateResponse } from '../../models/erpApi'
import { suggestedProfileCode, type Project } from '../../models/project'
import { Button, CONTROL_CLASS, ErrorNote, Field } from './Form'
import { Card, Empty, Kpi, KpiRow, Loading } from './Data'
import type { MapPoint } from './SiteMap'

const SiteMap = lazy(() =>
  import('./SiteMap').then((module) => ({ default: module.SiteMap })),
)

const PROJECT_OPTION_LIMIT = 100
const MAX_LATITUDE = 90
const MAX_LONGITUDE = 180
const NO_SITES: SavedLocation[] = []

export const SITE_EVALUATION_ID = 'evaluasi-lokasi'

const DIMENSION_LABEL: Record<string, string> = {
  fisik_lingkungan: 'Fisik & Lingkungan',
  infrastruktur: 'Infrastruktur & Aksesibilitas',
  demografi_sosial: 'Demografi & Sosial',
  pasar_kompetisi: 'Pasar & Kompetisi',
  finansial_proyek: 'Finansial Proyek',
}

const SELECT_CLASS = `${CONTROL_CLASS} h-10 w-full px-3 disabled:cursor-not-allowed`

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

interface SiteEvaluationProps {
  // Kalau diisi, evaluasi selalu dikaitkan ke proyek ini dan pilihan proyek
  // disembunyikan. Dipakai di ruang kerja proyek.
  project?: Project
  focus: MapPoint | null
  onFocus: (point: MapPoint) => void
}

// Peta, formulir evaluasi, dan hasilnya. Dipakai bersama oleh /lokasi dan tab
// Analisis Lokasi di ruang kerja proyek.
export function SiteEvaluation({ project, focus, onFocus }: SiteEvaluationProps) {
  const evaluate = useSiteEvaluate()
  const profiles = useBuildingProfiles()

  const [name, setName] = useState('')
  const [latitude, setLatitude] = useState('')
  const [longitude, setLongitude] = useState('')
  const [chosenProfileId, setChosenProfileId] = useState('')
  const [chosenProjectId, setChosenProjectId] = useState('')

  const projectId = project ? project.id : chosenProjectId
  const suggestedCode = project ? suggestedProfileCode(project.type) : null
  const suggestedProfileId =
    profiles.data?.find((profile) => profile.code === suggestedCode)?.id ?? ''
  const buildingProfileId = chosenProfileId || suggestedProfileId

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
      { onSuccess: () => onFocus(target) },
    )
  }

  return (
    <div className="space-y-6">
      <div id={SITE_EVALUATION_ID} className="scroll-mt-6">
        <Card
          title="Evaluasi lokasi"
          description="Klik peta atau ketik koordinat, lalu pilih profil bangunan untuk menjalankan analisis."
        >
          <form onSubmit={handleSubmit} className="space-y-5">
            <Suspense
              fallback={
                <div className="flex h-80 items-center justify-center rounded-xl bg-slate-100 sm:h-96">
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
                  Profil bangunan
                </label>

                <select
                  id="building-profile"
                  value={buildingProfileId}
                  onChange={(event) =>
                    setChosenProfileId(event.target.value)
                  }
                  required
                  disabled={profiles.isLoading}
                  aria-describedby={suggestedProfileId ? 'building-profile-hint' : undefined}
                  className={SELECT_CLASS}
                >
                  <option value="">
                    {profiles.isLoading
                      ? 'Memuat profil bangunan...'
                      : 'Pilih profil bangunan'}
                  </option>

                  {profiles.data?.map((profile) => (
                    <option key={profile.id} value={profile.id}>
                      {profile.name} ({profile.code})
                    </option>
                  ))}
                </select>

                {suggestedProfileId && project ? (
                  <p id="building-profile-hint" className="mt-1.5 text-xs text-slate-500">
                    Dipilih otomatis dari jenis proyek ({project.type}). Boleh diganti.
                  </p>
                ) : null}

                {profiles.isError ? (
                  <p className="mt-1.5 text-xs text-red-600">
                    Gagal mengambil daftar profil bangunan.
                  </p>
                ) : null}
              </div>

              <Field
                id="latitude"
                label="Lintang (latitude)"
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
                label="Bujur (longitude)"
                type="number"
                step="any"
                min={-MAX_LONGITUDE}
                max={MAX_LONGITUDE}
                placeholder="106.816666"
                value={longitude}
                onChange={(event) => setLongitude(event.target.value)}
                required
              />

              {project ? (
                <div className="md:col-span-2">
                  <p className="text-sm font-medium text-slate-700">Proyek</p>
                  <p className="mt-1.5 text-sm text-slate-900">
                    {project.name} ({project.code})
                  </p>
                  <p className="mt-1.5 text-xs text-slate-500">
                    Hasil evaluasi otomatis dikaitkan ke proyek ini, dan peta
                    hanya menampilkan lokasi milik proyek ini.
                  </p>
                </div>
              ) : (
                <ProjectSelect value={chosenProjectId} onChange={setChosenProjectId} />
              )}
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
      </div>

      {evaluate.isPending ? (
        <Card title="Analisis">
          <Loading label="Mengambil skor, regulasi, dan konteks berita..." />
        </Card>
      ) : null}

      {result ? <EvaluationResult result={result} /> : null}
    </div>
  )
}

// Komponen terpisah supaya daftar proyek hanya diambil di /lokasi, bukan di
// ruang kerja proyek yang proyeknya sudah pasti.
function ProjectSelect({ value, onChange }: { value: string; onChange: (projectId: string) => void }) {
  const projects = useProjects({ pageSize: PROJECT_OPTION_LIMIT })

  return (
    <div className="md:col-span-2">
      <label
        htmlFor="project"
        className="mb-1.5 block text-sm font-medium text-slate-700"
      >
        Proyek (opsional)
      </label>

      <select
        id="project"
        value={value}
        onChange={(event) => onChange(event.target.value)}
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
  )
}

function EvaluationResult({ result }: { result: SiteEvaluateResponse }) {
  return (
    <>
      <KpiRow>
        <Kpi
          label="Skor keseluruhan"
          value={`${result.predictive.overall_score}`}
          note="Skor kelayakan lokasi"
        />

        <Kpi
          label="Penanda risiko"
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
          description="Skor 0 sampai 100 dari lima dimensi data GIS, dibobot per profil bangunan."
        >
          <div className="space-y-4">
            {result.predictive.dimension_scores.length === 0 ? (
              <Empty message="Belum ada skor dimensi." />
            ) : (
              result.predictive.dimension_scores.map((dimension) => (
                <div
                  key={dimension.dimension_code}
                  className="rounded-xl bg-slate-50 p-4"
                >
                  <div className="flex items-center justify-between gap-4">
                    <p className="text-sm font-medium text-slate-900">
                      {DIMENSION_LABEL[dimension.dimension_code] ?? dimension.dimension_code}
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
          description="Informasi deskriptif dari RDTR. Tidak ikut dihitung dalam skor."
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
        description="Berita terbaru tentang wilayah ini. Tidak ikut dihitung dalam skor."
      >
        {result.descriptive?.news.length ? (
          <div className="space-y-3">
            {result.descriptive.news.map((article) => (
              <a
                key={`${article.url}-${article.title}`}
                href={article.url}
                target="_blank"
                rel="noreferrer"
                className="block rounded-xl p-4 shadow-hairline transition-[background-color,transform] hover:bg-slate-50 motion-safe:active:scale-[0.99]"
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
  )
}
