import { useState } from 'react'

import { useBuildingProfiles, useSavedLocations } from '../../controllers/useErp'
import type { SavedLocation } from '../../models/erpApi'
import { rankSites, summarizeSites } from '../../models/location'
import type { Project } from '../../models/project'
import { kpiValue, shortDate } from '../../shared/format'
import { Card, Kpi, KpiRow, LoadFailed, Loading, Table } from '../components/Data'
import { Button } from '../components/Form'
import { SITE_EVALUATION_ID, SiteEvaluation } from '../components/SiteEvaluation'
import type { MapPoint } from '../components/SiteMap'
import { ProgressMeter } from './parts'

const NO_SITES: SavedLocation[] = []
const COORDINATE_DIGITS = 4

export function ProjectSiteTab({ project }: { project: Project }) {
  const [focus, setFocus] = useState<MapPoint | null>(null)
  const sites = useSavedLocations(project.id)
  const profiles = useBuildingProfiles()

  const items = sites.data?.items ?? NO_SITES
  const summary = summarizeSites(items)
  const hasSites = summary.count > 0

  const profileName = (id: string | null) =>
    profiles.data?.find((profile) => profile.id === id)?.name ?? '-'

  function showOnMap(site: SavedLocation) {
    setFocus({ latitude: site.latitude, longitude: site.longitude })
    document.getElementById(SITE_EVALUATION_ID)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <div className="space-y-6">
      <KpiRow>
        <Kpi
          label="Lokasi dievaluasi"
          value={kpiValue(sites.isPending, sites.isError, String(sites.data?.totalItems ?? 0))}
          note="kandidat lokasi untuk proyek ini"
        />
        <Kpi
          label="Skor tertinggi"
          value={kpiValue(sites.isPending, sites.isError, hasSites ? String(summary.best?.score) : '—')}
          note={summary.best?.name ?? 'belum ada evaluasi'}
        />
        <Kpi
          label="Rata-rata skor"
          value={kpiValue(sites.isPending, sites.isError, hasSites ? String(summary.averageScore) : '—')}
          note="skala 0 sampai 100"
        />
        <Kpi
          label="Evaluasi terakhir"
          value={kpiValue(sites.isPending, sites.isError, hasSites ? shortDate(summary.latest?.savedAt ?? null) : '—')}
          note={summary.latest?.name ?? 'belum ada evaluasi'}
        />
      </KpiRow>

      <SiteEvaluation project={project} focus={focus} onFocus={setFocus} />

      <Card
        title="Perbandingan lokasi"
        description="Semua lokasi yang pernah dievaluasi untuk proyek ini, diurutkan dari skor kelayakan tertinggi"
      >
        {sites.isPending ? <Loading /> : null}
        {sites.isError ? <LoadFailed onRetry={() => sites.refetch()} /> : null}
        {sites.data ? (
          <Table
            rows={rankSites(items)}
            emptyMessage="Belum ada lokasi yang dievaluasi. Klik peta di atas untuk mengevaluasi lokasi pertama."
            columns={[
              { header: 'Nama', cell: (row) => <span className="font-medium text-slate-900">{row.name}</span> },
              { header: 'Profil bangunan', cell: (row) => profileName(row.buildingProfileId) },
              {
                header: 'Koordinat',
                cell: (row) => (
                  <span className="tabular-nums">
                    {row.latitude.toFixed(COORDINATE_DIGITS)}, {row.longitude.toFixed(COORDINATE_DIGITS)}
                  </span>
                ),
              },
              {
                header: 'Skor',
                cell: (row) => (
                  <div className="min-w-32">
                    <ProgressMeter percent={row.score} label={`Skor kelayakan ${row.name}`} suffix="" />
                  </div>
                ),
              },
              { header: 'Dievaluasi', cell: (row) => shortDate(row.savedAt) },
              {
                header: 'Peta',
                align: 'right',
                cell: (row) => (
                  <Button variant="subtle" onClick={() => showOnMap(row)}>
                    Lihat di peta<span className="sr-only">: {row.name}</span>
                  </Button>
                ),
              },
            ]}
          />
        ) : null}
      </Card>
    </div>
  )
}
