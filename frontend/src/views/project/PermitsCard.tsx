import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreatePermit,
  useDeletePermit,
  usePermits,
  useUpdatePermit,
} from '../../controllers/useProjectWorkspace'
import { toInputDate } from '../../models/finance'
import {
  EMPTY_PERMIT_FORM,
  issuedPermitValues,
  PERMIT_NUMBER_MAX_LENGTH,
  PERMIT_STATUS_LABEL,
  PERMIT_STATUS_TONE,
  PERMIT_TYPE_MAX_LENGTH,
  PERMIT_TYPE_SUGGESTIONS,
  permitFormValues,
  permitIssueValues,
  type Permit,
  type PermitFormValues,
  type PermitIssueValues,
  type PermitStatus,
  type Project,
} from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FormSection, InlineConfirm } from './parts'

const PERMIT_STATUSES: PermitStatus[] = ['submitted', 'issued', 'expired', 'rejected']

type UpdatePermit = ReturnType<typeof useUpdatePermit>

// Satu formulir terbuka pada satu waktu: izin baru, ubah izin, atau jalan
// pintas "Tandai terbit" untuk izin yang masih diajukan.
type PermitPanel = { kind: 'new' } | { kind: 'edit' | 'issue'; permitId: string } | null

export function PermitsCard({ project }: { project: Project }) {
  const permits = usePermits(project.id)
  const updatePermit = useUpdatePermit(project.id)
  const deletePermit = useDeletePermit(project.id)
  const [panel, setPanel] = useState<PermitPanel>(null)

  const selected =
    panel && panel.kind !== 'new' ? permits.data?.find((permit) => permit.id === panel.permitId) : undefined

  const isOpen = (kind: 'edit' | 'issue', permitId: string) =>
    panel !== null && panel.kind === kind && panel.permitId === permitId

  const toggle = (kind: 'edit' | 'issue', permitId: string) => {
    updatePermit.reset()
    setPanel(isOpen(kind, permitId) ? null : { kind, permitId })
  }

  const toggleNew = () => setPanel((current) => (current?.kind === 'new' ? null : { kind: 'new' }))
  const close = () => setPanel(null)

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
            {
              header: 'Aksi',
              align: 'right',
              cell: (row) => (
                <span className="inline-flex items-center justify-end gap-1">
                  {row.status === 'submitted' ? (
                    <RowAction
                      label="Tandai terbit"
                      isActive={isOpen('issue', row.id)}
                      onClick={() => toggle('issue', row.id)}
                    />
                  ) : null}
                  <RowAction label="Ubah" isActive={isOpen('edit', row.id)} onClick={() => toggle('edit', row.id)} />
                  <InlineConfirm
                    label="Hapus"
                    confirmLabel="Ya, hapus izin"
                    pendingLabel="Menghapus"
                    isPending={deletePermit.isPending && deletePermit.variables === row.id}
                    onConfirm={() => deletePermit.mutate(row.id)}
                  />
                </span>
              ),
            },
          ]}
        />
      ) : null}

      {deletePermit.isError ? (
        <div className="mt-4">
          <ErrorNote message={errorMessage(deletePermit.error)} />
        </div>
      ) : null}
      {updatePermit.isSuccess && panel === null ? (
        <div className="mt-4">
          <SuccessNote
            message={`Izin ${updatePermit.data.type} nomor ${updatePermit.data.number} sekarang ${PERMIT_STATUS_LABEL[updatePermit.data.status].toLowerCase()}.`}
          />
        </div>
      ) : null}

      <div className="mt-5">
        <Button variant="subtle" onClick={toggleNew}>
          {panel?.kind === 'new' ? 'Tutup formulir izin' : 'Tambah izin'}
        </Button>
      </div>

      {panel?.kind === 'new' ? <NewPermitForm projectId={project.id} /> : null}
      {panel?.kind === 'edit' && selected ? (
        <EditPermitForm key={selected.id} permit={selected} update={updatePermit} onDone={close} />
      ) : null}
      {panel?.kind === 'issue' && selected ? (
        <IssuePermitForm key={selected.id} permit={selected} update={updatePermit} onDone={close} />
      ) : null}
    </Card>
  )
}

function PermitFields({ idPrefix, values, onChange, autoFocus = false }: {
  idPrefix: string
  values: PermitFormValues
  onChange: (key: keyof PermitFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  autoFocus?: boolean
}) {
  const isIssued = values.status === 'issued'
  const optionsId = `${idPrefix}-type-options`

  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Field
        id={`${idPrefix}-type`}
        label="Jenis izin"
        placeholder="Pilih atau ketik jenis"
        list={optionsId}
        autoComplete="off"
        autoFocus={autoFocus}
        maxLength={PERMIT_TYPE_MAX_LENGTH}
        required
        value={values.type}
        onChange={onChange('type')}
      />
      <Field
        id={`${idPrefix}-number`}
        label="Nomor"
        placeholder="PBG-3273-2026-0142"
        autoComplete="off"
        maxLength={PERMIT_NUMBER_MAX_LENGTH}
        hint="Nomor izin, atau nomor pengajuan kalau belum terbit"
        required
        value={values.number}
        onChange={onChange('number')}
      />
      <SelectField id={`${idPrefix}-status`} label="Status" value={values.status} onChange={onChange('status')}>
        {PERMIT_STATUSES.map((status) => (
          <option key={status} value={status}>
            {PERMIT_STATUS_LABEL[status]}
          </option>
        ))}
      </SelectField>
      <Field
        id={`${idPrefix}-issued-date`}
        label={isIssued ? 'Tanggal terbit' : 'Tanggal terbit (opsional)'}
        type="date"
        required={isIssued}
        hint={isIssued ? 'Wajib diisi untuk izin yang sudah terbit' : undefined}
        value={values.issuedDate}
        onChange={onChange('issuedDate')}
      />
      <Field
        id={`${idPrefix}-valid-until`}
        label="Berlaku sampai (opsional)"
        type="date"
        min={values.issuedDate || undefined}
        hint="Tidak boleh lebih awal dari tanggal terbit"
        value={values.validUntil}
        onChange={onChange('validUntil')}
      />

      <datalist id={optionsId}>
        {PERMIT_TYPE_SUGGESTIONS.map((type) => (
          <option key={type} value={type} />
        ))}
      </datalist>
    </div>
  )
}

function usePermitValues(initial: PermitFormValues) {
  const [values, setValues] = useState<PermitFormValues>(initial)

  const change =
    (key: keyof PermitFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

function NewPermitForm({ projectId }: { projectId: string }) {
  const { values, setValues, change } = usePermitValues(EMPTY_PERMIT_FORM)
  const createPermit = useCreatePermit(projectId)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createPermit.mutate(values, { onSuccess: () => setValues(EMPTY_PERMIT_FORM) })
  }

  return (
    <FormSection title="Izin baru">
      <form onSubmit={submit} className="space-y-4">
        <PermitFields idPrefix="permit-new" values={values} onChange={change} />

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

function EditPermitForm({ permit, update, onDone }: { permit: Permit; update: UpdatePermit; onDone: () => void }) {
  const { values, change } = usePermitValues(permitFormValues(permit))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.mutate({ permitId: permit.id, values }, { onSuccess: onDone })
  }

  return (
    <FormSection title={`Ubah izin ${permit.type}`}>
      <form onSubmit={submit} className="space-y-4">
        <PermitFields idPrefix="permit-edit" values={values} onChange={change} autoFocus />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan izin">
            Simpan perubahan izin
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

// Izin yang terbit wajib punya tanggal terbit (permit_issued_date_required),
// jadi tanggal hari ini diisikan lebih dulu dan nomor resmi bisa diganti.
function IssuePermitForm({ permit, update, onDone }: { permit: Permit; update: UpdatePermit; onDone: () => void }) {
  const [values, setValues] = useState<PermitIssueValues>(() => permitIssueValues(permit, toInputDate(new Date())))
  const idPrefix = 'permit-issue'

  const change = (key: keyof PermitIssueValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.mutate({ permitId: permit.id, values: issuedPermitValues(permit, values) }, { onSuccess: onDone })
  }

  return (
    <FormSection title={`Tandai izin ${permit.type} terbit`}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-3">
          <Field
            id={`${idPrefix}-number`}
            label="Nomor izin"
            autoComplete="off"
            autoFocus
            maxLength={PERMIT_NUMBER_MAX_LENGTH}
            hint="Ganti dengan nomor resmi kalau berbeda dari nomor pengajuan"
            required
            value={values.number}
            onChange={change('number')}
          />
          <Field
            id={`${idPrefix}-issued-date`}
            label="Tanggal terbit"
            type="date"
            required
            value={values.issuedDate}
            onChange={change('issuedDate')}
          />
          <Field
            id={`${idPrefix}-valid-until`}
            label="Berlaku sampai (opsional)"
            type="date"
            min={values.issuedDate || undefined}
            value={values.validUntil}
            onChange={change('validUntil')}
          />
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan izin">
            Simpan izin terbit
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
