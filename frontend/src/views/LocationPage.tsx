import { useState } from 'react'

import {
  useBuildingProfiles,
  useSiteEvaluate,
} from '../controllers/useErp'
import { AppShell } from './components/AppShell'
import { Button, ErrorNote, Field } from './components/Form'
import { Card, Empty, Kpi, KpiRow, Loading } from './components/Data'

export function LocationPage() {
  const evaluate = useSiteEvaluate()
  const profiles = useBuildingProfiles()

  const [name, setName] = useState('')
  const [latitude, setLatitude] = useState('')
  const [longitude, setLongitude] = useState('')
  const [buildingProfileId, setBuildingProfileId] = useState('')

  const result = evaluate.data

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    evaluate.mutate({
      name: name.trim(),
      latitude: Number(latitude),
      longitude: Number(longitude),
      building_profile_id: buildingProfileId,
    })
  }

  return (
    <AppShell
      title="Analisis Lokasi"
      description="Evaluasi kelayakan lokasi dan konteks regulasi serta berita sekitar"
    >
      <div className="space-y-6">
        <Card
          title="Evaluasi lokasi"
          description="Masukkan koordinat dan profil bangunan untuk menjalankan analisis."
        >
          <form onSubmit={handleSubmit} className="space-y-5">
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
                  className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:cursor-not-allowed disabled:bg-slate-100"
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
                placeholder="106.816666"
                value={longitude}
                onChange={(event) => setLongitude(event.target.value)}
                required
              />
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