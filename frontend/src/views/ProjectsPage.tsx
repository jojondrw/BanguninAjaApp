import { type ChangeEvent, type FormEvent, useState } from 'react'
import { ChevronRight, Plus, X } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'

import { useCan } from '../controllers/useAccess'
import { useCreateProject, useProjects } from '../controllers/useErp'
import {
  deletedProjectName,
  EMPTY_PROJECT_FORM,
  PROJECT_CODE_MAX_LENGTH,
  PROJECT_NAME_MAX_LENGTH,
  PROJECT_STATUS_LABEL,
  PROJECT_TYPE_MAX_LENGTH,
  PROJECT_TYPE_SUGGESTIONS,
  type ProjectFormValues,
  type ProjectStatus,
} from '../models/project'
import { errorMessage } from '../shared/errorMessage'
import { rupiah, rupiahShort, shortDate } from '../shared/format'
import { AppShell } from './components/AppShell'
import { Card, LoadFailed, Loading, Table } from './components/Data'
import { Button, CONTROL_CLASS, ErrorNote, Field, SuccessNote } from './components/Form'
import { StatusChip } from './OverviewPage'
import { ProgressMeter } from './project/parts'

const TYPE_OPTIONS_ID = 'project-type-options'

const FILTERS: { value: ProjectStatus | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'planning', label: PROJECT_STATUS_LABEL.planning },
  { value: 'ongoing', label: PROJECT_STATUS_LABEL.ongoing },
  { value: 'on_hold', label: PROJECT_STATUS_LABEL.on_hold },
  { value: 'completed', label: PROJECT_STATUS_LABEL.completed },
  { value: 'cancelled', label: PROJECT_STATUS_LABEL.cancelled },
]

function filterChipClass(isActive: boolean): string {
  return `h-8 rounded-full px-3 text-[13px] font-medium transition-[background-color,color,transform]
          motion-safe:active:scale-[0.97] ${
            isActive ? 'bg-slate-900 text-white' : 'bg-white text-slate-600 shadow-hairline hover:bg-slate-50 hover:text-slate-900'
          }`
}

function emptyListMessage(onlyMine: boolean): string {
  if (onlyMine) {
    return 'Anda belum masuk tim proyek yang cocok dengan penyaringan ini. Anggota tim diatur di tab Tim ruang kerja proyek.'
  }
  return 'Tidak ada proyek yang cocok dengan penyaringan ini.'
}

const OPEN_LINK_CLASS =
  'inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium whitespace-nowrap text-navy-600 ' +
  'transition-[background-color,transform] hover:bg-navy-50 motion-safe:active:scale-95'

function contractValueHint(value: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return 'Dalam Rupiah. Kosongkan kalau belum ada kontrak.'
  }
  return rupiah(amount)
}

function NewProjectCard() {
  const [values, setValues] = useState<ProjectFormValues>(EMPTY_PROJECT_FORM)
  const createProject = useCreateProject()

  const update = (key: keyof ProjectFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createProject.mutate(values, { onSuccess: () => setValues(EMPTY_PROJECT_FORM) })
  }

  return (
    <Card
      title="Proyek baru"
      description="Proyek baru berstatus Perencanaan. Isian bertanda opsional boleh dikosongkan."
    >
      <form onSubmit={submit} className="space-y-5">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="project-code"
            label="Kode"
            placeholder="PRJ-001"
            autoComplete="off"
            maxLength={PROJECT_CODE_MAX_LENGTH}
            hint={`Unik, maksimal ${PROJECT_CODE_MAX_LENGTH} karakter`}
            required
            value={values.code}
            onChange={update('code')}
          />
          <Field
            id="project-name"
            label="Nama proyek"
            placeholder="Perumahan Griya Asri"
            autoComplete="off"
            maxLength={PROJECT_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="project-type"
            label="Jenis"
            placeholder="Pilih atau ketik jenis"
            list={TYPE_OPTIONS_ID}
            autoComplete="off"
            maxLength={PROJECT_TYPE_MAX_LENGTH}
            hint={`Pilih dari saran atau ketik sendiri, maksimal ${PROJECT_TYPE_MAX_LENGTH} karakter`}
            required
            value={values.type}
            onChange={update('type')}
          />
          <Field
            id="project-contract-value"
            label="Nilai kontrak (opsional)"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={contractValueHint(values.contractValue)}
            value={values.contractValue}
            onChange={update('contractValue')}
          />
          <Field
            id="project-start-date"
            label="Tanggal mulai (opsional)"
            type="date"
            value={values.startDate}
            onChange={update('startDate')}
          />
          <Field
            id="project-target-end-date"
            label="Target selesai (opsional)"
            type="date"
            min={values.startDate || undefined}
            hint="Tidak boleh lebih awal dari tanggal mulai"
            value={values.targetEndDate}
            onChange={update('targetEndDate')}
          />
        </div>

        <datalist id={TYPE_OPTIONS_ID}>
          {PROJECT_TYPE_SUGGESTIONS.map((type) => (
            <option key={type} value={type} />
          ))}
        </datalist>

        <Button type="submit" isPending={createProject.isPending} pendingLabel="Menyimpan proyek">
          Simpan proyek
        </Button>

        {createProject.isError ? <ErrorNote message={errorMessage(createProject.error)} /> : null}
        {createProject.isSuccess ? (
          <SuccessNote
            message={`Proyek ${createProject.data.name} (${createProject.data.code}) berhasil dibuat.`}
          />
        ) : null}
      </form>
    </Card>
  )
}

export function ProjectsPage() {
  const [status, setStatus] = useState<ProjectStatus | ''>('')
  const [search, setSearch] = useState('')
  const [onlyMine, setOnlyMine] = useState(false)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const canCreate = useCan('projects', 'write')
  const location = useLocation()
  const deletedProject = deletedProjectName(location.state)
  const projects = useProjects({
    pageSize: 50,
    status: status === '' ? undefined : status,
    search: search.trim() === '' ? undefined : search.trim(),
    mine: onlyMine ? 'true' : undefined,
  })

  return (
    <AppShell
      title="Proyek"
      description="Daftar proyek, disaring dan dicari langsung di server. Buka ruang kerja proyek untuk mengubah data, status, tahap, izin, dan RAB."
      actions={
        canCreate ? (
          <Button
            variant={isFormOpen ? 'subtle' : 'primary'}
            onClick={() => setIsFormOpen((open) => !open)}
          >
            {isFormOpen ? (
              <X aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
            ) : (
              <Plus aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
            )}
            {isFormOpen ? 'Tutup formulir' : 'Tambah proyek'}
          </Button>
        ) : undefined
      }
    >
      {isFormOpen && canCreate ? (
        <div className="mb-6">
          <NewProjectCard />
        </div>
      ) : null}

      {deletedProject ? (
        <div className="mb-6">
          <SuccessNote message={`Proyek ${deletedProject} sudah dihapus.`} />
        </div>
      ) : null}

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
            className={`${CONTROL_CLASS} h-8 w-56 px-3 text-[13px]`}
          />
          <button
            type="button"
            onClick={() => setOnlyMine((current) => !current)}
            aria-pressed={onlyMine}
            className={filterChipClass(onlyMine)}
          >
            Proyek saya
          </button>
          <span aria-hidden="true" className="h-5 w-px bg-slate-200" />
          <div className="flex flex-wrap gap-1.5" role="group" aria-label="Saring menurut status">
            {FILTERS.map((filter) => (
              <button
                key={filter.label}
                type="button"
                onClick={() => setStatus(filter.value)}
                aria-pressed={status === filter.value}
                className={filterChipClass(status === filter.value)}
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
            emptyMessage={emptyListMessage(onlyMine)}
            columns={[
              { header: 'Kode', cell: (row) => row.code },
              {
                header: 'Nama',
                cell: (row) => (
                  <Link to={`/proyek/${row.id}`} className="font-medium text-slate-900 underline-offset-4 hover:text-navy-600 hover:underline">
                    {row.name}
                  </Link>
                ),
              },
              { header: 'Jenis', cell: (row) => row.type },
              { header: 'Status', cell: (row) => <StatusChip project={row} /> },
              { header: 'Mulai', cell: (row) => shortDate(row.startDate) },
              { header: 'Target', cell: (row) => shortDate(row.targetEndDate) },
              { header: 'Nilai', align: 'right', cell: (row) => rupiahShort(row.contractValue) },
              {
                header: 'Progres',
                cell: (row) => (
                  <div className="w-32">
                    <ProgressMeter percent={row.progress} label={`Progres ${row.name}`} />
                  </div>
                ),
              },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <Link to={`/proyek/${row.id}`} className={OPEN_LINK_CLASS} aria-label={`Buka ruang kerja ${row.name}`}>
                    Buka ruang kerja
                    <ChevronRight aria-hidden="true" className="size-3.5" strokeWidth={1.8} />
                  </Link>
                ),
              },
            ]}
          />
        ) : null}
      </Card>
    </AppShell>
  )
}
