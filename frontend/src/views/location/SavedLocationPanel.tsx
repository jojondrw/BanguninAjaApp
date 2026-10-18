import { useEffect, useRef } from 'react'

import { useDimensions, useProjectNames, useSavedLocationDetails } from '../../controllers/useLocations'
import { rupiah, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { Button } from '../components/Form'
import { Money } from '../components/Money'
import { Ring } from '../components/Ring'
import { SiteReportPanel } from './ReportDownload'

const COORDINATE_DIGITS = 5
const INDEX_DIGITS = 2

// Rincian lokasi yang pernah dievaluasi, dibuka dari penanda di peta. Konteks
// regulasi dan berita tidak disimpan saat evaluasi, jadi yang tampil adalah
// skor, risiko, dan data lokasi yang tersimpan, ditambah laporan PDF.
export function SavedLocationPanel({ id, onClose }: { id: string; onClose: () => void }) {
  const panelRef = useRef<HTMLDivElement>(null)
  const details = useSavedLocationDetails([id])
  const dimensions = useDimensions()
  const site = details.locations[0]
  const projectNames = useProjectNames(site?.projectId ? [site.projectId] : [])

  useEffect(() => {
    panelRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [id])

  const dimensionName = (dimensionId: string) =>
    dimensions.data?.find((dimension) => dimension.id === dimensionId)?.name ?? 'Dimensi'

  return (
    <div ref={panelRef} className="scroll-mt-6 space-y-6">
      <Card
        title={site ? site.name : 'Rincian lokasi'}
        description={site ? `Dievaluasi ${shortDate(site.savedAt)}` : undefined}
        action={
          <Button variant="subtle" onClick={onClose}>
            Tutup rincian
          </Button>
        }
      >
        {details.isPending ? <Loading /> : null}
        {details.error ? <LoadFailed onRetry={details.retry} /> : null}
        {site ? (
          <div className="space-y-8">
            {site.scoreSimulated ? (
              <p role="alert" className="rounded-2xl bg-amber-50 px-4 py-3 text-[13px] text-amber-900">
                <span className="font-semibold">Skor simulasi.</span> Angka di bawah adalah contoh, bukan hasil data GIS.
                Evaluasi ulang titik ini sebelum dipakai untuk keputusan.
              </p>
            ) : null}
            <div className="flex flex-wrap items-center gap-8">
              <Ring value={site.score} size={120} label={String(site.score)} caption="dari 100" />
              <dl className="grid flex-1 grid-cols-2 gap-x-8 gap-y-4 sm:grid-cols-3">
                <Fact label="Proyek">{site.projectId ? (projectNames.get(site.projectId) ?? '...') : 'Tanpa proyek'}</Fact>
                <Fact label="Koordinat">
                  <span className="tabular-nums">
                    {site.latitude.toFixed(COORDINATE_DIGITS)}, {site.longitude.toFixed(COORDINATE_DIGITS)}
                  </span>
                </Fact>
                <Fact label="Harga tanah per m²">
                  {site.landPricePerSqm > 0 ? <Money>{rupiah(site.landPricePerSqm)}</Money> : 'Tidak ada data'}
                </Fact>
                <Fact label="Indeks bahaya banjir">{site.floodIndex.toFixed(INDEX_DIGITS)}</Fact>
                <Fact label="Indeks bahaya gempa">{site.earthquakeIndex.toFixed(INDEX_DIGITS)}</Fact>
                <Fact label="Sumber skor">{site.scoreSimulated ? 'Simulasi' : 'Data GIS'}</Fact>
              </dl>
            </div>

            <div className="grid gap-8 lg:grid-cols-2">
              <section>
                <h3 className="text-sm font-semibold text-slate-900">Skor per dimensi</h3>
                <ul className="mt-3 space-y-3">
                  {site.dimensionScores.map((score) => (
                    <li key={score.dimensionId}>
                      <div className="flex items-baseline justify-between gap-3 text-[13px]">
                        <span className="text-slate-700">{dimensionName(score.dimensionId)}</span>
                        <span className="font-medium text-slate-900 tabular-nums">{score.value}</span>
                      </div>
                      <span aria-hidden="true" className="mt-1.5 block h-1.5 overflow-hidden rounded-[2px] bg-slate-100">
                        <span
                          className="bar-fill block h-full rounded-[2px]"
                          style={{ width: `${Math.max(0, Math.min(100, score.value))}%` }}
                        />
                      </span>
                    </li>
                  ))}
                </ul>
              </section>

              <section>
                <h3 className="text-sm font-semibold text-slate-900">Penanda risiko</h3>
                {site.riskFlags && site.riskFlags.length > 0 ? (
                  <ul className="mt-3 space-y-2">
                    {site.riskFlags.map((flag) => (
                      <li key={`${flag.code}-${flag.message}`} className="rounded-xl bg-red-50 px-3.5 py-2.5">
                        <p className="text-[13px] font-medium text-red-800">
                          {flag.severity === 'high' ? 'Tinggi' : 'Sedang'}: {flag.code}
                        </p>
                        <p className="mt-0.5 text-xs text-red-700">{flag.message}</p>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="mt-3 text-[13px] text-slate-500">Tidak ada risiko yang terdeteksi untuk lokasi ini.</p>
                )}
                {site.note ? <p className="mt-4 text-[13px] text-slate-600">Catatan: {site.note}</p> : null}
              </section>
            </div>
          </div>
        ) : null}
      </Card>

      {site ? <SiteReportPanel savedLocationId={site.id} /> : null}
    </div>
  )
}

function Fact({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm font-medium text-slate-900">{children}</dd>
    </div>
  )
}
