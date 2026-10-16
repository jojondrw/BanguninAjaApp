import { useState } from 'react'

import { useBuildingProfiles, useSavedLocationPage, useSavedLocationSummary } from '../../controllers/useErp'
import { useDownloadSiteReport } from '../../controllers/useLocationReports'
import type { SavedLocation } from '../../models/erpApi'
import type { Project } from '../../models/project'
import { kpiValue, number, shortDate } from '../../shared/format'
import { Card, Kpi, KpiRow, LoadFailed, Loading, Table } from '../components/Data'
import { Button } from '../components/Form'
import { Pager } from '../components/RecordControls'
import { SITE_EVALUATION_ID, SiteEvaluation } from '../components/SiteEvaluation'
import type { MapPoint } from '../components/SiteMap'
import { DownloadNotes } from '../location/ReportDownload'
import { ProgressMeter } from './parts'

const COORDINATE_DIGITS = 4
const PAGE_SIZE = 10

// KPI dihitung backend dari semua lokasi proyek, sedangkan tabelnya berhalaman
// dengan urutan skor tertinggi dulu.
export function ProjectSiteTab({ project }: { project: Project }) {
  const [focus, setFocus] = useState<MapPoint | null>(null)
  const [page, setPage] = useState(1)
  const sites = useSavedLocationPage(project.id, page, PAGE_SIZE)
  const totals = useSavedLocationSummary(project.id)
  const profiles = useBuildingProfiles()
  const report = useDownloadSiteReport()

  const summary = totals.data
  const hasSites = (summary?.count ?? 0) > 0
  const kpi = (value: string) => kpiValue(totals.isPending, totals.isError, hasSites ? value : '—')

  const profileName = (id: string | null) =>
    profiles.data?.find((profile) => profile.id === id)?.name ?? '-'

  function showOnMap(site: SavedLocation) {
    setFocus({ latitude: site.latitude, longitude: site.longitude })
    document.getElementById(SITE_EVALUATION_ID)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <div className="space-y-8">
      <KpiRow>
        <Kpi
          label="Lokasi dievaluasi"
          value={kpiValue(totals.isPending, totals.isError, number(summary?.count ?? 0))}
          note="kandidat lokasi untuk proyek ini"
        />
        <Kpi
          label="Skor tertinggi"
          value={kpi(String(summary?.best?.score ?? '—'))}
          note={summary?.best?.name ?? 'belum ada evaluasi'}
        />
        <Kpi label="Rata-rata skor" value={kpi(String(summary?.averageScore ?? '—'))} note="skala 0 sampai 100" />
        <Kpi
          label="Evaluasi terakhir"
          value={kpi(shortDate(summary?.latest?.savedAt ?? null))}
          note={summary?.latest?.name ?? 'belum ada evaluasi'}
        />
      </KpiRow>

      <SiteEvaluation project={project} focus={focus} onFocus={setFocus} />

      <Card
        title="Perbandingan lokasi"
        description="Semua lokasi yang pernah dievaluasi untuk proyek ini, diurutkan dari skor kelayakan tertinggi"
      >
        {sites.isPending ? <Loading /> : null}
        {sites.isError ? <LoadFailed onRetry={() => sites.refetch()} /> : null}
        {report.isError || report.isSuccess ? (
          <div className="mb-4">
            <DownloadNotes fileName={report.data} error={report.error} />
          </div>
        ) : null}
        {sites.data ? (
          <>
            <Table
              rows={sites.data.items}
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
                {
                  header: 'Laporan',
                  align: 'right',
                  cell: (row) => (
                    <Button
                      variant="subtle"
                      isPending={report.isPending && report.variables === row.id}
                      pendingLabel="Mengunduh..."
                      onClick={() => report.mutate(row.id)}
                    >
                      Unduh PDF<span className="sr-only">: laporan {row.name}</span>
                    </Button>
                  ),
                },
              ]}
            />
            <Pager
              page={sites.data.page}
              totalPages={sites.data.totalPages}
              totalItems={sites.data.totalItems}
              unit="lokasi"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>
    </div>
  )
}
