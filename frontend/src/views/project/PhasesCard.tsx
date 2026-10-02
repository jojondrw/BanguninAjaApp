import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreatePhase,
  useDeletePhase,
  usePhases,
  useUpdatePhase,
} from '../../controllers/useProjectWorkspace'
import {
  EMPTY_PHASE_FORM,
  isProjectClosed,
  isValidProgress,
  nextPhaseOrder,
  PHASE_NAME_MAX_LENGTH,
  PHASE_STATUS_LABEL,
  PHASE_STATUS_TONE,
  phaseEditValues,
  PROGRESS_MAX,
  sortPhases,
  type PhaseEditValues,
  type PhaseFormValues,
  type Project,
  type ProjectPhase,
} from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { dateRange } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Button, CONTROL_CLASS, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FormSection, InlineConfirm, ProgressMeter } from './parts'

type UpdatePhase = ReturnType<typeof useUpdatePhase>

export function PhasesCard({ project }: { project: Project }) {
  const phases = usePhases(project.id)
  const updatePhase = useUpdatePhase(project.id)
  const [isAdding, setIsAdding] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const isClosed = isProjectClosed(project)

  const toggleEdit = (phaseId: string) => {
    updatePhase.reset()
    setIsAdding(false)
    setEditingId((current) => (current === phaseId ? null : phaseId))
  }

  const toggleAdd = () => {
    setEditingId(null)
    setIsAdding((open) => !open)
  }

  return (
    <Card title="Tahap proyek" description="Progres proyek dihitung server dari rata-rata progres tahap">
      {phases.isPending ? <Loading /> : null}
      {phases.isError ? <LoadFailed onRetry={() => phases.refetch()} /> : null}
      {phases.data ? (
        phases.data.length === 0 ? (
          <Empty message="Belum ada tahap. Mulai dengan tahap pertama, misalnya Pekerjaan persiapan." />
        ) : (
          <ol className="divide-y divide-slate-100">
            {sortPhases(phases.data).map((phase) => (
              <PhaseRow
                key={phase.id}
                projectId={project.id}
                phase={phase}
                isClosed={isClosed}
                isEditing={editingId === phase.id}
                onToggleEdit={() => toggleEdit(phase.id)}
                update={updatePhase}
                onEditDone={() => setEditingId(null)}
              />
            ))}
          </ol>
        )
      ) : null}

      {updatePhase.isSuccess && editingId === null ? (
        <div className="mt-4">
          <SuccessNote message={`Tahap ${updatePhase.data.name} disimpan. Progres proyek ikut diperbarui.`} />
        </div>
      ) : null}

      {isClosed ? (
        <p className="mt-5 text-xs text-slate-500">
          Proyek sudah selesai atau dibatalkan, jadi tahap tidak bisa ditambah, diubah, atau dihapus.
        </p>
      ) : (
        <div className="mt-5">
          <Button variant="subtle" onClick={toggleAdd}>
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

function PhaseRow({ projectId, phase, isClosed, isEditing, onToggleEdit, update, onEditDone }: {
  projectId: string
  phase: ProjectPhase
  isClosed: boolean
  isEditing: boolean
  onToggleEdit: () => void
  update: UpdatePhase
  onEditDone: () => void
}) {
  const deletePhase = useDeletePhase(projectId)
  const saveProgress = useUpdatePhase(projectId)

  return (
    <li className="py-4 first:pt-0 last:pb-0">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-medium text-slate-900">
          {phase.sortOrder}. {phase.name}
        </p>
        <Chip tone={PHASE_STATUS_TONE[phase.status]}>{PHASE_STATUS_LABEL[phase.status]}</Chip>
      </div>
      <p className="mt-0.5 mb-1.5 text-xs text-slate-500">{dateRange(phase.startDate, phase.targetEndDate)}</p>
      <ProgressMeter percent={phase.progress} label={`Progres ${phase.name}`} />

      {isClosed ? null : (
        <div className="mt-3 flex flex-wrap items-center justify-between gap-x-3 gap-y-2">
          {isEditing ? (
            <span className="text-xs text-slate-500">Ubah progres lewat formulir di bawah.</span>
          ) : (
            <PhaseProgressForm key={`${phase.id}-${phase.progress}`} phase={phase} save={saveProgress} />
          )}
          <div className="flex items-center gap-1">
            <RowAction label={isEditing ? 'Tutup' : 'Ubah'} isActive={isEditing} onClick={onToggleEdit} />
            <InlineConfirm
              label="Hapus"
              confirmLabel="Ya, hapus tahap"
              pendingLabel="Menghapus"
              isPending={deletePhase.isPending}
              onConfirm={() => deletePhase.mutate(phase.id)}
            />
          </div>
        </div>
      )}

      {saveProgress.isError ? (
        <div className="mt-2">
          <ErrorNote message={errorMessage(saveProgress.error)} />
        </div>
      ) : null}
      {deletePhase.isError ? (
        <div className="mt-2">
          <ErrorNote message={errorMessage(deletePhase.error)} />
        </div>
      ) : null}

      {isEditing && !isClosed ? <PhaseEditForm phase={phase} update={update} onDone={onEditDone} /> : null}
    </li>
  )
}

// Jalan cepat untuk laporan harian: cukup angka progres baru. Kolom lain
// dikirim ulang apa adanya karena PUT tahap menimpa semua kolom.
function PhaseProgressForm({ phase, save }: { phase: ProjectPhase; save: UpdatePhase }) {
  const [progress, setProgress] = useState(String(phase.progress))
  const inputId = `phase-quick-progress-${phase.id}`
  const isSaved = save.isSuccess && save.data.progress === phase.progress && progress === String(phase.progress)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (isValidProgress(progress)) {
      save.mutate({ phaseId: phase.id, values: { ...phaseEditValues(phase), progress } })
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-wrap items-center gap-2">
      <label htmlFor={inputId} className="text-xs text-slate-600">
        Progres baru (%)
      </label>
      <input
        id={inputId}
        type="number"
        inputMode="numeric"
        min={0}
        max={PROGRESS_MAX}
        step={1}
        required
        value={progress}
        onChange={(event) => setProgress(event.target.value)}
        className={`${CONTROL_CLASS} h-9 w-20 px-2.5 tabular-nums`}
      />
      <Button type="submit" variant="subtle" isPending={save.isPending} pendingLabel="Menyimpan">
        Simpan progres
      </Button>
      {isSaved ? (
        <span role="status" className="text-xs text-green-700">
          Progres disimpan
        </span>
      ) : null}
    </form>
  )
}

function PhaseEditForm({ phase, update, onDone }: { phase: ProjectPhase; update: UpdatePhase; onDone: () => void }) {
  const [values, setValues] = useState<PhaseEditValues>(() => phaseEditValues(phase))
  const idPrefix = `phase-edit-${phase.id}`

  const change = (key: keyof PhaseEditValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.mutate({ phaseId: phase.id, values }, { onSuccess: onDone })
  }

  return (
    <FormSection title={`Ubah tahap ${phase.name}`}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field
            id={`${idPrefix}-name`}
            label="Nama tahap"
            autoComplete="off"
            autoFocus
            maxLength={PHASE_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={change('name')}
          />
          <Field
            id={`${idPrefix}-sort-order`}
            label="Urutan"
            type="number"
            inputMode="numeric"
            min={1}
            step={1}
            required
            hint="Menentukan posisi tahap di daftar, mulai dari 1"
            value={values.sortOrder}
            onChange={change('sortOrder')}
          />
          <Field
            id={`${idPrefix}-start-date`}
            label="Mulai (opsional)"
            type="date"
            value={values.startDate}
            onChange={change('startDate')}
          />
          <Field
            id={`${idPrefix}-target-end-date`}
            label="Target selesai (opsional)"
            type="date"
            min={values.startDate || undefined}
            hint="Tidak boleh lebih awal dari tanggal mulai"
            value={values.targetEndDate}
            onChange={change('targetEndDate')}
          />
          <Field
            id={`${idPrefix}-progress`}
            label="Progres (%)"
            type="number"
            inputMode="numeric"
            min={0}
            max={PROGRESS_MAX}
            step={1}
            required
            hint="0 sampai 100. Status tahap mengikuti angka ini."
            value={values.progress}
            onChange={change('progress')}
          />
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan tahap">
            Simpan tahap
          </Button>
          <Button variant="subtle" onClick={onDone}>
            Batal
          </Button>
        </div>

        {update.isError ? <ErrorNote message={errorMessage(update.error)} /> : null}
      </form>
    </FormSection>
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
            max={PROGRESS_MAX}
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
