import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useUpdateProject } from '../../controllers/useProjectWorkspace'
import {
  PROJECT_CODE_MAX_LENGTH,
  PROJECT_NAME_MAX_LENGTH,
  PROJECT_TYPE_MAX_LENGTH,
  PROJECT_TYPE_SUGGESTIONS,
  projectFormValues,
  type Project,
  type ProjectFormValues,
} from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { Card } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RegionPicker } from './RegionPicker'

const TYPE_OPTIONS_ID = 'project-edit-type-options'

function contractValueHint(value: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return 'Dalam Rupiah. Kosongkan kalau belum ada kontrak.'
  }
  return rupiah(amount)
}

// Isian dimulai dari data proyek saat panel dibuka. Status dan progres tidak
// ada di sini karena diatur lewat panel status dan tahap proyek.
export function ProjectEditCard({ project, onClose }: { project: Project; onClose: () => void }) {
  const [values, setValues] = useState<ProjectFormValues>(() => projectFormValues(project))
  const [regionId, setRegionId] = useState(project.regionId ?? '')
  const updateProject = useUpdateProject(project.id)

  const update = (key: keyof ProjectFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateProject.mutate({ values, regionId })
  }

  return (
    <Card
      title="Ubah data proyek"
      description="Semua isian dikirim ulang ke server. Status dan progres diatur terpisah."
    >
      <form onSubmit={submit} className="space-y-5">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="project-edit-code"
            label="Kode"
            autoComplete="off"
            maxLength={PROJECT_CODE_MAX_LENGTH}
            hint={`Unik, maksimal ${PROJECT_CODE_MAX_LENGTH} karakter`}
            required
            value={values.code}
            onChange={update('code')}
          />
          <Field
            id="project-edit-name"
            label="Nama proyek"
            autoComplete="off"
            maxLength={PROJECT_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="project-edit-type"
            label="Jenis"
            list={TYPE_OPTIONS_ID}
            autoComplete="off"
            maxLength={PROJECT_TYPE_MAX_LENGTH}
            hint={`Pilih dari saran atau ketik sendiri, maksimal ${PROJECT_TYPE_MAX_LENGTH} karakter`}
            required
            value={values.type}
            onChange={update('type')}
          />
          <Field
            id="project-edit-contract-value"
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
            id="project-edit-start-date"
            label="Tanggal mulai (opsional)"
            type="date"
            value={values.startDate}
            onChange={update('startDate')}
          />
          <Field
            id="project-edit-target-end-date"
            label="Target selesai (opsional)"
            type="date"
            min={values.startDate || undefined}
            hint="Tidak boleh lebih awal dari tanggal mulai"
            value={values.targetEndDate}
            onChange={update('targetEndDate')}
          />
          <RegionPicker idPrefix="project-edit" value={regionId} onChange={setRegionId} />
        </div>

        <datalist id={TYPE_OPTIONS_ID}>
          {PROJECT_TYPE_SUGGESTIONS.map((type) => (
            <option key={type} value={type} />
          ))}
        </datalist>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateProject.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan
          </Button>
          <Button variant="subtle" onClick={onClose}>
            Tutup formulir
          </Button>
        </div>

        {updateProject.isError ? <ErrorNote message={errorMessage(updateProject.error)} /> : null}
        {updateProject.isSuccess ? (
          <SuccessNote message={`Data proyek ${updateProject.data.name} (${updateProject.data.code}) disimpan.`} />
        ) : null}
      </form>
    </Card>
  )
}
