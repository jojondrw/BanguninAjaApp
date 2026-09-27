import { useState } from 'react'

import { useBuildingProfiles, useEvaluateSite } from '../controllers/useSite'
import type { LatLng } from '../models/site'
import { errorMessage } from '../shared/errorMessage'
import { AppShell } from './components/AppShell'
import { Card, Empty, Loading } from './components/Data'
import { Button, ErrorNote, Field } from './components/Form'
import { ScorePanel } from './components/ScorePanel'
import { SiteMap } from './components/SiteMap'

export function LocationPage() {
  const [picked, setPicked] = useState<LatLng | null>(null)
  const [name, setName] = useState('')
  const [profileId, setProfileId] = useState('')

  const profiles = useBuildingProfiles()
  const evaluate = useEvaluateSite()

  const canEvaluate = picked !== null && name.trim() !== '' && profileId !== '' && !evaluate.isPending

  const runEvaluation = () => {
    if (!picked || !canEvaluate) {
      return
    }
    evaluate.mutate({
      latitude: picked.latitude,
      longitude: picked.longitude,
      building_profile_id: profileId,
      name: name.trim(),
    })
  }

  const pickPoint = (point: LatLng) => {
    setPicked(point)
    evaluate.reset()
  }

  return (
    <AppShell
      title="Analisis Lokasi"
      description="Klik titik di peta, pilih profil bangunan, lalu nilai kelayakan lokasi."
    >
      <div className="grid gap-6 xl:grid-cols-[1.6fr_1fr]">
        <Card title="Peta lokasi" description="Klik peta untuk menempatkan penanda">
          <SiteMap onPick={pickPoint} selected={picked} />
        </Card>

        <div className="space-y-6">
          <Card title="Nilai lokasi" description="Lengkapi data lalu jalankan penilaian">
            <div className="space-y-4">
              <dl className="grid grid-cols-2 gap-3 rounded-lg bg-slate-50 p-3">
                <div>
                  <dt className="text-xs text-slate-500">Latitude</dt>
                  <dd className="mt-0.5 text-sm font-semibold tabular-nums text-slate-900">
                    {picked ? picked.latitude.toFixed(6) : '—'}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-slate-500">Longitude</dt>
                  <dd className="mt-0.5 text-sm font-semibold tabular-nums text-slate-900">
                    {picked ? picked.longitude.toFixed(6) : '—'}
                  </dd>
                </div>
              </dl>

              <Field
                id="site-name"
                label="Nama lokasi"
                placeholder="mis. Kavling Buah Batu"
                value={name}
                maxLength={160}
                onChange={(event) => setName(event.target.value)}
              />

              <div className="flex flex-col gap-1.5">
                <label htmlFor="building-profile" className="text-sm font-medium text-slate-700">
                  Profil bangunan
                </label>
                <select
                  id="building-profile"
                  value={profileId}
                  disabled={profiles.isPending || profiles.isError}
                  onChange={(event) => setProfileId(event.target.value)}
                  className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900
                             transition focus:border-navy-600 focus:ring-2 focus:ring-navy-100 disabled:bg-slate-100"
                >
                  <option value="">
                    {profiles.isPending ? 'Memuat profil...' : 'Pilih profil bangunan'}
                  </option>
                  {(profiles.data ?? []).map((profile) => (
                    <option key={profile.id} value={profile.id}>
                      {profile.name}
                    </option>
                  ))}
                </select>
                {profiles.isError ? (
                  <p className="text-xs text-red-600">Gagal memuat profil bangunan dari server.</p>
                ) : null}
                {profiles.data && profiles.data.length === 0 ? (
                  <p className="text-xs text-slate-500">
                    Belum ada profil bangunan. Tambahkan lewat API POST /api/scoring/building-profiles.
                  </p>
                ) : null}
              </div>

              <Button
                onClick={runEvaluation}
                isPending={evaluate.isPending}
                pendingLabel="Menilai..."
                variant="primary"
              >
                Nilai lokasi
              </Button>

              {!picked ? (
                <p className="text-xs text-slate-400">Klik di peta untuk memilih titik terlebih dahulu.</p>
              ) : null}
              {evaluate.isError ? <ErrorNote message={errorMessage(evaluate.error)} /> : null}
            </div>
          </Card>

          <Card title="Hasil penilaian" description="Skor prediktif dari 5 dimensi (regulasi tidak dinilai)">
            {evaluate.isPending ? <Loading label="Menghitung skor..." /> : null}
            {evaluate.data ? (
              <ScorePanel predictive={evaluate.data.predictive} />
            ) : evaluate.isPending ? null : (
              <Empty message="Belum ada hasil. Pilih titik dan profil, lalu jalankan penilaian." />
            )}
          </Card>
        </div>
      </div>
    </AppShell>
  )
}
