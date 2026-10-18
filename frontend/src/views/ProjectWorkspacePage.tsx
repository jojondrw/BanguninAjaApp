import { type ReactNode, useState } from 'react'
import { ChevronLeft } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'

import { useCan } from '../controllers/useAccess'
import { useProject } from '../controllers/useProjectWorkspace'
import type { Project } from '../models/project'
import { ApiError } from '../shared/apiClient'
import { rupiah, shortDate } from '../shared/format'
import { AppShell } from './components/AppShell'
import { Card, Empty, LoadFailed, Loading } from './components/Data'
import { BUTTON_BASE, BUTTON_SUBTLE } from './components/Form'
import { AccessScope } from './access/AccessScope'
import { Tabs, type TabItem } from './components/Tabs'
import { StatusChip } from './project/StatusChip'
import { PanelToggle, ProgressMeter } from './project/parts'
import { ProjectDeleteCard } from './project/ProjectDeleteCard'
import { ProjectEditCard } from './project/ProjectEditCard'
import { ProjectFieldTab } from './project/ProjectFieldTab'
import { ProjectFinanceTab } from './project/ProjectFinanceTab'
import { ProjectSiteTab } from './project/ProjectSiteTab'
import { ProjectStatusCard } from './project/ProjectStatusCard'
import { ProjectSummaryTab } from './project/ProjectSummaryTab'
import { ProjectTeamTab } from './project/ProjectTeamTab'
import { Money, MoneyText } from './components/Money'

type WorkspaceTab = 'ringkasan' | 'lapangan' | 'tim' | 'lokasi' | 'keuangan'

const TABS: readonly TabItem<WorkspaceTab>[] = [
  { id: 'ringkasan', label: 'Ringkasan' },
  { id: 'lapangan', label: 'Lapangan' },
  { id: 'tim', label: 'Tim' },
  { id: 'lokasi', label: 'Analisis Lokasi' },
  { id: 'keuangan', label: 'Keuangan' },
]

const DEFAULT_TAB: WorkspaceTab = 'ringkasan'
const TAB_PARAM = 'tab'
const NOT_FOUND_STATUSES = [400, 404]

function toTab(value: string | null): WorkspaceTab {
  return TABS.find((tab) => tab.id === value)?.id ?? DEFAULT_TAB
}

function isOutsideTeam(error: unknown): boolean {
  return error instanceof ApiError && error.code === 'project_member_required'
}

function isMissingProject(error: unknown): boolean {
  return error instanceof ApiError && NOT_FOUND_STATUSES.includes(error.status)
}

function workspaceErrorTitle(error: unknown): string {
  if (isOutsideTeam(error)) {
    return 'Bukan proyek tim Anda'
  }
  return isMissingProject(error) ? 'Proyek tidak ditemukan' : 'Gagal memuat proyek'
}

type ManagePanel = 'status' | 'edit' | 'delete'

const PANEL_ID = 'proyek-panel-kelola'

const PANEL_TOGGLES: { panel: ManagePanel; label: string; tone?: 'danger' }[] = [
  { panel: 'status', label: 'Ubah status' },
  { panel: 'edit', label: 'Ubah proyek' },
  { panel: 'delete', label: 'Hapus proyek', tone: 'danger' },
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
  const canReadLocations = useCan('locations', 'read')
  const canReadFinance = useCan('finance', 'read')
  // Tab yang datanya dari modul lain hanya tampil kalau peran boleh membacanya.
  const tabs = TABS.filter(
    (item) => (item.id !== 'lokasi' || canReadLocations) && (item.id !== 'keuangan' || canReadFinance),
  )
  const requestedTab = toTab(searchParams.get(TAB_PARAM))
  const tab = tabs.some((item) => item.id === requestedTab) ? requestedTab : DEFAULT_TAB
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
        <Card title={workspaceErrorTitle(project.error)}>
          {isOutsideTeam(project.error) ? (
            <Empty message="Proyek ini hanya bisa dibuka anggota timnya. Minta Admin atau Manajer Proyek di proyek ini menambahkan Anda di tab Tim." />
          ) : null}
          {isMissingProject(project.error) ? (
            <Empty message="Proyek ini tidak ada atau sudah dihapus. Pilih proyek lain dari daftar proyek." />
          ) : null}
          {!isOutsideTeam(project.error) && !isMissingProject(project.error) ? (
            <LoadFailed onRetry={() => project.refetch()} />
          ) : null}
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
              tone={toggle.tone}
              panelId={PANEL_ID}
              isOpen={panel === toggle.panel}
              onToggle={() => togglePanel(toggle.panel)}
            />
          ))}
        </div>
      }
    >
      <div className="space-y-8">
        {panel ? (
          <div id={PANEL_ID} key={`${data.id}-${panel}`}>
            {panel === 'status' ? <ProjectStatusCard project={data} onClose={closePanel} /> : null}
            {panel === 'edit' ? <ProjectEditCard project={data} onClose={closePanel} /> : null}
            {panel === 'delete' ? <ProjectDeleteCard project={data} onClose={closePanel} /> : null}
          </div>
        ) : null}

        <ProjectFacts project={data} />

        <Tabs idPrefix="proyek" label="Bagian ruang kerja proyek" tabs={tabs} active={tab} onChange={changeTab}>
          {tab === 'ringkasan' ? <ProjectSummaryTab project={data} /> : null}
          {tab === 'lapangan' ? <ProjectFieldTab project={data} /> : null}
          {tab === 'tim' ? <ProjectTeamTab project={data} /> : null}
          {tab === 'lokasi' ? (
            <AccessScope module="locations">
              <ProjectSiteTab project={data} />
            </AccessScope>
          ) : null}
          {tab === 'keuangan' ? (
            <AccessScope module="finance">
              <ProjectFinanceTab project={data} />
            </AccessScope>
          ) : null}
        </Tabs>
      </div>
    </AppShell>
  )
}

function ProjectFacts({ project }: { project: Project }) {
  return (
    <section aria-label="Data proyek" className="rounded-2xl bg-white px-5 py-4 shadow-hairline sm:px-6 sm:py-5">
      <dl className="grid gap-x-6 gap-y-4 sm:grid-cols-2 lg:grid-cols-5">
        <Fact label="Status">
          <StatusChip project={project} />
        </Fact>
        <Fact label="Nilai kontrak">
          <span className="tabular-nums"><Money>{rupiah(project.contractValue)}</Money></span>
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
      <dd className="mt-1.5 text-sm font-medium text-slate-900"><MoneyText>{children}</MoneyText></dd>
    </div>
  )
}
