import { useState } from 'react'

import { useProjects } from '../controllers/useErp'
import { PROJECT_STATUS_LABEL, type ProjectStatus } from '../models/project'
import { rupiahShort, shortDate } from '../shared/format'
import { AppShell } from './components/AppShell'
import { Bar, Card, LoadFailed, Loading, Table } from './components/Data'
import { StatusChip } from './OverviewPage'

const FILTERS: { value: ProjectStatus | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'planning', label: PROJECT_STATUS_LABEL.planning },
  { value: 'ongoing', label: PROJECT_STATUS_LABEL.ongoing },
  { value: 'on_hold', label: PROJECT_STATUS_LABEL.on_hold },
  { value: 'completed', label: PROJECT_STATUS_LABEL.completed },
]

export function ProjectsPage() {
  const [status, setStatus] = useState<ProjectStatus | ''>('')
  const [search, setSearch] = useState('')
  const projects = useProjects({
    pageSize: 50,
    status: status === '' ? undefined : status,
    search: search.trim() === '' ? undefined : search.trim(),
  })

  return (
    <AppShell title="Proyek" description="Daftar proyek, disaring dan dicari langsung di server">
      <Card title="Daftar proyek" description={projects.data ? `${projects.data.totalItems} proyek` : undefined}>
        <div className="mb-4 flex flex-wrap items-center gap-2">
          <label htmlFor="cari" className="sr-only">
            Cari proyek
          </label>
          <input
            id="cari"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Cari nama proyek"
            className="w-56 rounded-lg border border-slate-300 px-3 py-1.5 text-sm placeholder:text-slate-400
                       focus:border-navy-600 focus:ring-2 focus:ring-navy-100"
          />
          <div className="flex flex-wrap gap-1">
            {FILTERS.map((filter) => (
              <button
                key={filter.label}
                type="button"
                onClick={() => setStatus(filter.value)}
                aria-pressed={status === filter.value}
                className={`rounded-lg px-3 py-1.5 text-sm transition ${
                  status === filter.value
                    ? 'bg-navy-700 text-white'
                    : 'border border-slate-300 text-slate-600 hover:bg-slate-50'
                }`}
              >
                {filter.label}
              </button>
            ))}
          </div>
        </div>

        {projects.isPending ? <Loading /> : null}
        {projects.isError ? <LoadFailed onRetry={() => projects.refetch()} /> : null}
        {projects.data ? (
          <Table
            rows={projects.data.items}
            emptyMessage="Tidak ada proyek yang cocok dengan penyaringan ini."
            columns={[
              { header: 'Kode', cell: (row) => row.code },
              { header: 'Nama', cell: (row) => row.name },
              { header: 'Jenis', cell: (row) => row.type },
              { header: 'Status', cell: (row) => <StatusChip project={row} /> },
              { header: 'Mulai', cell: (row) => shortDate(row.startDate) },
              { header: 'Target', cell: (row) => shortDate(row.targetEndDate) },
              { header: 'Nilai', align: 'right', cell: (row) => rupiahShort(row.contractValue) },
              { header: 'Progres', align: 'right', cell: (row) => <Bar percent={row.progress} /> },
            ]}
          />
        ) : null}
      </Card>
    </AppShell>
  )
}
