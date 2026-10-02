import { type ReactNode, useState } from 'react'
import { ChevronLeft } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'

import { useProject } from '../controllers/useProjectWorkspace'
import type { Project } from '../models/project'
import { ApiError } from '../shared/apiClient'
import { rupiah, shortDate } from '../shared/format'
import { AppShell } from './components/AppShell'
import { Card, Empty, LoadFailed, Loading } from './components/Data'
import { BUTTON_BASE, BUTTON_SUBTLE } from './components/Form'
import { Tabs, type TabItem } from './components/Tabs'
import { StatusChip } from './OverviewPage'
import { PanelToggle, ProgressMeter } from './project/parts'
import { ProjectDeleteCard } from './project/ProjectDeleteCard'
import { ProjectEditCard } from './project/ProjectEditCard'
import { ProjectFinanceTab } from './project/ProjectFinanceTab'
import { ProjectSiteTab } from './project/ProjectSiteTab'
import { ProjectStatusCard } from './project/ProjectStatusCard'
import { ProjectSummaryTab } from './project/ProjectSummaryTab'

type WorkspaceTab = 'ringkasan' | 'lokasi' | 'keuangan'

const TABS: readonly TabItem<WorkspaceTab>[] = [
  { id: 'ringkasan', label: 'Ringkasan' },
  { id: 'lokasi', label: 'Analisis Lokasi' },
  { id: 'keuangan', label: 'Keuangan' },
]

const DEFAULT_TAB: WorkspaceTab = 'ringkasan'
const TAB_PARAM = 'tab'
const NOT_FOUND_STATUSES = [400, 404]

function toTab(value: string | null): WorkspaceTab {
  return TABS.find((tab) => tab.id === value)?.id ?? DEFAULT_TAB
}

function isMissingProject(error: unknown): boolean {
  return error instanceof ApiError && NOT_FOUND_STATUSES.includes(error.status)
}

type ManagePanel = 'status' | 'edit' | 'delete'

const PANEL_ID = 'proyek-panel-kelola'

const PANEL_TOGGLES: { panel: ManagePanel; label: string }[] = [
  { panel: 'status', label: 'Ubah status' },
  { panel: 'edit', label: 'Ubah proyek' },
  { panel: 'delete', label: 'Hapus proyek' },
]

const BACK_LINK_CLASS = `h-9 pl-2.5 ${BUTTON_BASE} ${BUTTON_SUBTLE}`

function BackLink() {
  return (
    <Link to="/proyek" className={BACK_LINK_CLASS}>
      <ChevronLeft aria-hidden="true" className="size-4" strokeWidth={1.8} />
      Semua proyek
    </Link>
  )
}

export function ProjectWorkspacePage() {
  const { id = '' } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const project = useProject(id)
  const tab = toTab(searchParams.get(TAB_PARAM))
  const [panel, setPanel] = useState<ManagePanel | null>(null)

  const togglePanel = (next: ManagePanel) => setPanel((current) => (current === next ? null : next))
  const closePanel = () => setPanel(null)

  // Tab disimpan di URL supaya muat ulang dan tautan yang dibagikan membuka tab
  // yang sama. replace dipakai agar tombol kembali tidak menelusuri tiap tab.
  function changeTab(next: WorkspaceTab) {
    setSearchParams(next === DEFAULT_TAB ? {} : { [TAB_PARAM]: next }, { replace: true })
  }

  if (project.isPending) {
    return (
      <AppShell title="Ruang kerja proyek" actions={<BackLink />}>
        <Loading label="Mengambil data proyek..." />
      </AppShell>
    )
  }

  if (project.isError) {
    return (
      <AppShell title="Ruang kerja proyek" actions={<BackLink />}>
        <Card title={isMissingProject(project.error) ? 'Proyek tidak ditemukan' : 'Gagal memuat proyek'}>
          {isMissingProject(project.error) ? (
            <Empty message="Proyek ini tidak ada atau sudah dihapus. Pilih proyek lain dari daftar proyek." />
          ) : (
            <LoadFailed onRetry={() => project.refetch()} />
          )}
        </Card>
      </AppShell>
    )
  }

  const data = project.data

  // Satu panel kelola terbuka pada satu waktu, tepat di bawah kepala halaman.
  // Panel diberi key id proyek supaya isian ubah selalu mulai dari data proyek
  // yang sedang dibuka.
  return (
    <AppShell
      title={data.name}
      description={`${data.code} · ${data.type}`}
      actions={
        <div className="flex flex-wrap items-center gap-2">
          <BackLink />
          {PANEL_TOGGLES.map((toggle) => (
            <PanelToggle
              key={toggle.panel}
              label={toggle.label}
              panelId={PANEL_ID}
              isOpen={panel === toggle.panel}
              onToggle={() => togglePanel(toggle.panel)}
            />
          ))}
        </div>
      }
    >
      <div className="space-y-6">
        {panel ? (
          <div id={PANEL_ID} key={`${data.id}-${panel}`}>
            {panel === 'status' ? <ProjectStatusCard project={data} onClose={closePanel} /> : null}
            {panel === 'edit' ? <ProjectEditCard project={data} onClose={closePanel} /> : null}
            {panel === 'delete' ? <ProjectDeleteCard project={data} onClose={closePanel} /> : null}
          </div>
        ) : null}

        <ProjectFacts project={data} />

        <Tabs idPrefix="proyek" label="Bagian ruang kerja proyek" tabs={TABS} active={tab} onChange={changeTab}>
          {tab === 'ringkasan' ? <ProjectSummaryTab project={data} /> : null}
          {tab === 'lokasi' ? <ProjectSiteTab project={data} /> : null}
          {tab === 'keuangan' ? <ProjectFinanceTab project={data} /> : null}
        </Tabs>
      </div>
    </AppShell>
  )
}

function ProjectFacts({ project }: { project: Project }) {
  return (
    <section aria-label="Data proyek" className="rounded-2xl bg-white px-5 py-4 shadow-hairline">
      <dl className="grid gap-x-6 gap-y-4 sm:grid-cols-2 lg:grid-cols-5">
        <Fact label="Status">
          <StatusChip project={project} />
        </Fact>
        <Fact label="Nilai kontrak">
          <span className="tabular-nums">{rupiah(project.contractValue)}</span>
        </Fact>
        <Fact label="Mulai">{shortDate(project.startDate)}</Fact>
        <Fact label="Target selesai">{shortDate(project.targetEndDate)}</Fact>
        <Fact label="Progres">
          <ProgressMeter percent={project.progress} label="Progres proyek" />
        </Fact>
      </dl>
    </section>
  )
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-medium text-slate-500">{label}</dt>
      <dd className="mt-1.5 text-sm font-medium text-slate-900">{children}</dd>
    </div>
  )
}
