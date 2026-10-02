import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useBudgetItems,
  useCreatePermit,
  useCreatePhase,
  usePermits,
  usePhases,
  useUnitsOfMeasure,
} from '../../controllers/useProjectWorkspace'
import {
  EMPTY_PERMIT_FORM,
  EMPTY_PHASE_FORM,
  isProjectClosed,
  nextPhaseOrder,
  PERMIT_NUMBER_MAX_LENGTH,
  PERMIT_STATUS_LABEL,
  PERMIT_STATUS_TONE,
  PERMIT_TYPE_MAX_LENGTH,
  PERMIT_TYPE_SUGGESTIONS,
  PHASE_NAME_MAX_LENGTH,
  PHASE_STATUS_LABEL,
  PHASE_STATUS_TONE,
  sortPhases,
  summarizePermits,
  summarizePhases,
  type Permit,
  type PermitFormValues,
  type PermitStatus,
  type PhaseFormValues,
  type Project,
  type ProjectPhase,
} from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { dateRange, kpiValue, number, rupiah, rupiahShort, shortDate } from '../../shared/format'
import { Card, Empty, Kpi, KpiRow, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { Chip, FormSection, ProgressMeter } from './parts'

const NO_PHASES: ProjectPhase[] = []
const NO_PERMITS: Permit[] = []
const PERMIT_TYPE_OPTIONS_ID = 'permit-type-options'
const PERMIT_STATUSES: PermitStatus[] = ['submitted', 'issued', 'expired', 'rejected']

export function ProjectSummaryTab({ project }: { project: Project }) {
  const phases = usePhases(project.id)
  const permits = usePermits(project.id)
  const budgetItems = useBudgetItems(project.id)

  const phaseSummary = summarizePhases(phases.data ?? NO_PHASES)
  const permitSummary = summarizePermits(permits.data ?? NO_PERMITS)
  const rabTotal = budgetItems.data?.grandTotal ?? 0

  return (
    <div className="space-y-6">
      <KpiRow>
        <Kpi
          label="Progres proyek"
          value={`${project.progress}%`}
          note={phaseSummary.total > 0 ? `rata-rata dari ${phaseSummary.total} tahap` : 'belum ada tahap'}
        />
        <Kpi
          label="Tahap selesai"
          value={kpiValue(phases.isPending, phases.isError, `${phaseSummary.completed} dari ${phaseSummary.total}`)}
          note={`${phaseSummary.inProgress} tahap sedang berjalan`}
        />
        <Kpi
          label="Izin terbit"
          value={kpiValue(permits.isPending, permits.isError, `${permitSummary.issued} dari ${permitSummary.total}`)}
          note={`${permitSummary.submitted} izin masih diajukan`}
        />
        <Kpi
          label="Nilai RAB"
          value={kpiValue(budgetItems.isPending, budgetItems.isError, rupiahShort(rabTotal))}
          note={rabNote(rabTotal, project.contractValue, budgetItems.data?.totalItems ?? 0)}
        />
      </KpiRow>

      <div className="grid gap-6 xl:grid-cols-2">
        <PhasesCard project={project} />
        <PermitsCard project={project} />
      </div>

      <BudgetItemsCard project={project} />
    </div>
  )
}

function rabNote(total: number, contractValue: number, items: number): string {
  if (items === 0) {
    return 'belum ada item RAB'
  }
  if (contractValue <= 0) {
    return `${items} item pekerjaan`
  }
  return `${items} item, ${Math.round((total / contractValue) * 100)}% dari nilai kontrak`
}

function PhasesCard({ project }: { project: Project }) {
  const phases = usePhases(project.id)
  const [isAdding, setIsAdding] = useState(false)
  const isClosed = isProjectClosed(project)

  return (
    <Card title="Tahap proyek" description="Progres proyek dihitung server dari rata-rata progres tahap">
      {phases.isPending ? <Loading /> : null}
      {phases.isError ? <LoadFailed onRetry={() => phases.refetch()} /> : null}
      {phases.data ? (
        phases.data.length === 0 ? (
          <Empty message="Belum ada tahap. Mulai dengan tahap pertama, misalnya Pekerjaan persiapan." />
        ) : (
          <ol className="space-y-4">
            {sortPhases(phases.data).map((phase) => (
              <li key={phase.id}>
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm font-medium text-slate-900">
                    {phase.sortOrder}. {phase.name}
                  </p>
                  <Chip tone={PHASE_STATUS_TONE[phase.status]}>{PHASE_STATUS_LABEL[phase.status]}</Chip>
                </div>
                <p className="mt-0.5 mb-1.5 text-xs text-slate-500">
                  {dateRange(phase.startDate, phase.targetEndDate)}
                </p>
                <ProgressMeter percent={phase.progress} label={`Progres ${phase.name}`} />
              </li>
            ))}
          </ol>
        )
      ) : null}

      {isClosed ? (
        <p className="mt-5 text-xs text-slate-500">
          Proyek sudah selesai atau dibatalkan, jadi tahap tidak bisa ditambah.
        </p>
      ) : (
        <div className="mt-5">
          <Button variant="subtle" onClick={() => setIsAdding((open) => !open)}>
            {isAdding ? 'Tutup formulir tahap' : 'Tambah tahap'}
          </Button>
        </div>
      )}

      {isAdding && !isClosed && phases.data ? (
        <PhaseForm projectId={project.id} sortOrder={nextPhaseOrder(phases.data)} />
      ) : null}
    </Card>
  )
}

function PhaseForm({ projectId, sortOrder }: { projectId: string; sortOrder: number }) {
  const [values, setValues] = useState<PhaseFormValues>(EMPTY_PHASE_FORM)
  const createPhase = useCreatePhase(projectId)

  const update = (key: keyof PhaseFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createPhase.mutate({ values, sortOrder }, { onSuccess: () => setValues(EMPTY_PHASE_FORM) })
  }

  return (
    <FormSection title={`Tahap ke-${sortOrder}`}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field
            id="phase-name"
            label="Nama tahap"
            placeholder="Pekerjaan struktur"
            autoComplete="off"
            maxLength={PHASE_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="phase-progress"
            label="Progres (%)"
            type="number"
            inputMode="numeric"
            min={0}
            max={100}
            step={1}
            required
            hint="0 sampai 100. Status tahap mengikuti angka ini."
            value={values.progress}
            onChange={update('progress')}
          />
          <Field
            id="phase-start-date"
            label="Mulai (opsional)"
            type="date"
            value={values.startDate}
            onChange={update('startDate')}
          />
          <Field
            id="phase-target-end-date"
            label="Target selesai (opsional)"
            type="date"
            min={values.startDate || undefined}
            hint="Tidak boleh lebih awal dari tanggal mulai"
            value={values.targetEndDate}
            onChange={update('targetEndDate')}
          />
        </div>

        <Button type="submit" isPending={createPhase.isPending} pendingLabel="Menyimpan tahap">
          Simpan tahap
        </Button>

        {createPhase.isError ? <ErrorNote message={errorMessage(createPhase.error)} /> : null}
        {createPhase.isSuccess ? (
          <SuccessNote message={`Tahap ${createPhase.data.name} ditambahkan. Progres proyek diperbarui.`} />
        ) : null}
      </form>
    </FormSection>
  )
}

function PermitsCard({ project }: { project: Project }) {
  const permits = usePermits(project.id)
  const [isAdding, setIsAdding] = useState(false)

  return (
    <Card title="Perizinan" description="Izin yang diajukan dan yang sudah terbit untuk proyek ini">
      {permits.isPending ? <Loading /> : null}
      {permits.isError ? <LoadFailed onRetry={() => permits.refetch()} /> : null}
      {permits.data ? (
        <Table
          rows={permits.data}
          emptyMessage="Belum ada izin yang dicatat untuk proyek ini."
          columns={[
            { header: 'Jenis', cell: (row) => row.type },
            { header: 'Nomor', cell: (row) => row.number },
            {
              header: 'Status',
              cell: (row) => <Chip tone={PERMIT_STATUS_TONE[row.status]}>{PERMIT_STATUS_LABEL[row.status]}</Chip>,
            },
            { header: 'Terbit', cell: (row) => shortDate(row.issuedDate) },
            { header: 'Berlaku sampai', cell: (row) => shortDate(row.validUntil) },
          ]}
        />
      ) : null}

      <div className="mt-5">
        <Button variant="subtle" onClick={() => setIsAdding((open) => !open)}>
          {isAdding ? 'Tutup formulir izin' : 'Tambah izin'}
        </Button>
      </div>

      {isAdding ? <PermitForm projectId={project.id} /> : null}
    </Card>
  )
}

function PermitForm({ projectId }: { projectId: string }) {
  const [values, setValues] = useState<PermitFormValues>(EMPTY_PERMIT_FORM)
  const createPermit = useCreatePermit(projectId)
  const isIssued = values.status === 'issued'

  const update =
    (key: keyof PermitFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createPermit.mutate(values, { onSuccess: () => setValues(EMPTY_PERMIT_FORM) })
  }

  return (
    <FormSection title="Izin baru">
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field
            id="permit-type"
            label="Jenis izin"
            placeholder="Pilih atau ketik jenis"
            list={PERMIT_TYPE_OPTIONS_ID}
            autoComplete="off"
            maxLength={PERMIT_TYPE_MAX_LENGTH}
            required
            value={values.type}
            onChange={update('type')}
          />
          <Field
            id="permit-number"
            label="Nomor"
            placeholder="PBG-3273-2026-0142"
            autoComplete="off"
            maxLength={PERMIT_NUMBER_MAX_LENGTH}
            hint="Nomor izin, atau nomor pengajuan kalau belum terbit"
            required
            value={values.number}
            onChange={update('number')}
          />
          <SelectField id="permit-status" label="Status" value={values.status} onChange={update('status')}>
            {PERMIT_STATUSES.map((status) => (
              <option key={status} value={status}>
                {PERMIT_STATUS_LABEL[status]}
              </option>
            ))}
          </SelectField>
          <Field
            id="permit-issued-date"
            label={isIssued ? 'Tanggal terbit' : 'Tanggal terbit (opsional)'}
            type="date"
            required={isIssued}
            hint={isIssued ? 'Wajib diisi untuk izin yang sudah terbit' : undefined}
            value={values.issuedDate}
            onChange={update('issuedDate')}
          />
          <Field
            id="permit-valid-until"
            label="Berlaku sampai (opsional)"
            type="date"
            min={values.issuedDate || undefined}
            value={values.validUntil}
            onChange={update('validUntil')}
          />
        </div>

        <datalist id={PERMIT_TYPE_OPTIONS_ID}>
          {PERMIT_TYPE_SUGGESTIONS.map((type) => (
            <option key={type} value={type} />
          ))}
        </datalist>

        <Button type="submit" isPending={createPermit.isPending} pendingLabel="Menyimpan izin">
          Simpan izin
        </Button>

        {createPermit.isError ? <ErrorNote message={errorMessage(createPermit.error)} /> : null}
        {createPermit.isSuccess ? (
          <SuccessNote message={`Izin ${createPermit.data.type} nomor ${createPermit.data.number} dicatat.`} />
        ) : null}
      </form>
    </FormSection>
  )
}

function BudgetItemsCard({ project }: { project: Project }) {
  const budgetItems = useBudgetItems(project.id)
  const units = useUnitsOfMeasure()

  const unitCode = (id: string) => units.data?.items.find((unit) => unit.id === id)?.code ?? ''
  const page = budgetItems.data

  return (
    <Card
      title="Rencana anggaran biaya (RAB)"
      description={page ? budgetItemsDescription(page.items.length, page.totalItems, page.grandTotal) : undefined}
    >
      {budgetItems.isPending ? <Loading /> : null}
      {budgetItems.isError ? <LoadFailed onRetry={() => budgetItems.refetch()} /> : null}
      {page ? (
        <Table
          rows={page.items}
          emptyMessage="Belum ada item RAB untuk proyek ini."
          columns={[
            { header: 'Kode', cell: (row) => row.code },
            { header: 'Uraian', cell: (row) => row.description },
            { header: 'Volume', align: 'right', cell: (row) => `${number(row.volume)} ${unitCode(row.unitOfMeasureId)}` },
            { header: 'Harga satuan', align: 'right', cell: (row) => rupiah(row.unitPrice) },
            { header: 'Jumlah', align: 'right', cell: (row) => rupiah(row.total) },
          ]}
        />
      ) : null}
    </Card>
  )
}

function budgetItemsDescription(shown: number, total: number, grandTotal: number): string {
  const sum = `total ${rupiah(grandTotal)}`
  if (shown < total) {
    return `Menampilkan ${shown} dari ${total} item, ${sum}`
  }
  return `${total} item, ${sum}`
}
