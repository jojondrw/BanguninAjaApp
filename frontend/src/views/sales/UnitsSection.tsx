import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../../controllers/useErp'
import { useCreateUnit, useDeleteUnit, useUnits, useUpdateUnit } from '../../controllers/useSales'
import type { Project } from '../../models/project'
import {
  EMPTY_UNIT_FORM,
  UNIT_CODE_MAX_LENGTH,
  UNIT_STATUS_LABEL,
  UNIT_STATUS_TONE,
  UNIT_TYPE_MAX_LENGTH,
  isUnitEditable,
  unitToForm,
  type NewUnitStatus,
  type PropertyUnit,
  type UnitFormValues,
} from '../../models/sales'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiahShort } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FilterSelect, FormPanel, FormToggle, Pager, SelectField, Toolbar } from '../components/RecordControls'
import { ActionGroup, ConfirmAction } from './RowActions'
import { OPTION_LIMIT, PAGE_SIZE, amountHint, pageAfterRemoval, projectLabel, refusalText } from './salesShared'

const NEW_UNIT_STATUSES: NewUnitStatus[] = ['available', 'on_hold']

function UnitForm({ unit, projects, isLoadingProjects, onSaved, onCancel }: {
  unit: PropertyUnit | null
  projects: Project[]
  isLoadingProjects: boolean
  onSaved: (message: string) => void
  onCancel: () => void
}) {
  const [values, setValues] = useState<UnitFormValues>(() => (unit ? unitToForm(unit) : EMPTY_UNIT_FORM))
  const createUnit = useCreateUnit()
  const updateUnit = useUpdateUnit()
  const saving = unit ? updateUnit : createUnit

  const update = (key: keyof UnitFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (unit) {
      updateUnit.mutate(
        { id: unit.id, values },
        { onSuccess: (saved) => onSaved(`Unit ${saved.code} berhasil diperbarui.`) },
      )
      return
    }
    createUnit.mutate(values, { onSuccess: () => setValues(EMPTY_UNIT_FORM) })
  }

  if (isLoadingProjects) {
    return (
      <FormPanel>
        <Loading label="Mengambil proyek..." />
      </FormPanel>
    )
  }

  if (projects.length === 0) {
    return (
      <FormPanel>
        <Empty message="Belum ada proyek. Buat proyek dulu di menu Proyek, lalu tambahkan unitnya di sini." />
      </FormPanel>
    )
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-sm font-semibold text-slate-900">{unit ? `Ubah unit ${unit.code}` : 'Unit baru'}</h3>
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="unit-code"
            label="Kode unit"
            placeholder="A-01"
            autoComplete="off"
            autoFocus={unit !== null}
            maxLength={UNIT_CODE_MAX_LENGTH}
            hint={`Unik, maksimal ${UNIT_CODE_MAX_LENGTH} karakter`}
            required
            value={values.code}
            onChange={update('code')}
          />
          <SelectField id="unit-project" label="Proyek" required value={values.projectId} onChange={update('projectId')}>
            <option value="">Pilih proyek</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name} ({project.code})
              </option>
            ))}
          </SelectField>
          <Field
            id="unit-type"
            label="Tipe"
            placeholder="Tipe 36/72"
            autoComplete="off"
            maxLength={UNIT_TYPE_MAX_LENGTH}
            required
            value={values.unitType}
            onChange={update('unitType')}
          />
          <Field
            id="unit-area"
            label="Luas (m²)"
            type="number"
            inputMode="decimal"
            min={0.01}
            step={0.01}
            placeholder="72"
            required
            value={values.areaSqm}
            onChange={update('areaSqm')}
          />
          <Field
            id="unit-price"
            label="Harga jual"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={amountHint(values.price, 'Dalam Rupiah')}
            value={values.price}
            onChange={update('price')}
          />
          <SelectField
            id="unit-status"
            label={unit ? 'Status' : 'Status awal'}
            hint="Dipesan dan Terjual diatur otomatis oleh kontrak"
            value={values.status}
            onChange={update('status')}
          >
            {NEW_UNIT_STATUSES.map((status) => (
              <option key={status} value={status}>
                {UNIT_STATUS_LABEL[status]}
              </option>
            ))}
          </SelectField>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={saving.isPending} pendingLabel="Menyimpan unit">
            {unit ? 'Simpan perubahan' : 'Simpan unit'}
          </Button>
          {unit ? (
            <Button variant="subtle" onClick={onCancel}>
              Batal ubah
            </Button>
          ) : null}
        </div>

        {saving.isError ? <ErrorNote message={errorMessage(saving.error)} /> : null}
        {createUnit.isSuccess && !unit ? (
          <SuccessNote message={`Unit ${createUnit.data.code} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

export function UnitsSection() {
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<PropertyUnit | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const units = useUnits({ projectId: projectId === '' ? undefined : projectId, page, pageSize: PAGE_SIZE })
  const deleteUnit = useDeleteUnit()
  const projectItems = projects.data?.items ?? []
  const isFormOpen = isCreating || editing !== null

  const closeForm = () => {
    setIsCreating(false)
    setEditing(null)
  }

  const toggleForm = () => {
    setNotice(null)
    if (isFormOpen) {
      closeForm()
      return
    }
    setIsCreating(true)
  }

  const startEdit = (unit: PropertyUnit) => {
    setNotice(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === unit.id ? null : unit))
  }

  const askDelete = (unit: PropertyUnit) => {
    setNotice(null)
    deleteUnit.reset()
    setAskingId(unit.id)
  }

  const confirmDelete = (unit: PropertyUnit) => {
    deleteUnit.mutate(unit, {
      onSuccess: () => {
        setNotice(`Unit ${unit.code} berhasil dihapus.`)
        setPage((current) => pageAfterRemoval(current, units.data?.items.length ?? 0))
        if (editing?.id === unit.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card title="Daftar unit" description="Unit yang sudah dipesan atau terjual terkunci oleh kontraknya.">
      <Toolbar>
        <FilterSelect
          id="unit-project-filter"
          label="Saring per proyek"
          value={projectId}
          disabled={projects.isPending}
          onChange={(event) => {
            setProjectId(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Semua proyek</option>
          {projectItems.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name}
            </option>
          ))}
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Tambah unit" onToggle={toggleForm} />
      </Toolbar>

      {isFormOpen ? (
        <UnitForm
          key={editing?.id ?? 'new'}
          unit={editing}
          projects={projectItems}
          isLoadingProjects={projects.isPending}
          onCancel={closeForm}
          onSaved={(message) => {
            closeForm()
            setNotice(message)
          }}
        />
      ) : null}

      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}
      {deleteUnit.isError ? (
        <div className="mb-4">
          <ErrorNote message={refusalText(`Unit ${deleteUnit.variables.code} tidak bisa dihapus.`, deleteUnit.error)} />
        </div>
      ) : null}

      {units.isPending ? <Loading /> : null}
      {units.isError ? <LoadFailed onRetry={() => units.refetch()} /> : null}
      {units.data ? (
        <>
          <Table
            rows={units.data.items}
            emptyMessage={
              projectId === ''
                ? 'Belum ada unit. Tambahkan unit pertama lewat tombol Tambah unit.'
                : 'Proyek ini belum punya unit.'
            }
            columns={[
              { header: 'Kode', cell: (row) => <span className="font-medium text-slate-900">{row.code}</span> },
              { header: 'Proyek', cell: (row) => projectLabel(projectItems, row.projectId) },
              { header: 'Tipe', cell: (row) => row.unitType },
              { header: 'Luas', align: 'right', cell: (row) => `${number(row.areaSqm)} m²` },
              { header: 'Harga', align: 'right', cell: (row) => rupiahShort(row.price) },
              {
                header: 'Status',
                cell: (row) => <Chip tone={UNIT_STATUS_TONE[row.status]}>{UNIT_STATUS_LABEL[row.status]}</Chip>,
              },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <ActionGroup>
                    {isUnitEditable(row) && askingId !== row.id ? (
                      <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                    ) : null}
                    <ConfirmAction
                      label="Hapus"
                      question={`Hapus unit ${row.code}?`}
                      confirmLabel="Ya, hapus"
                      pendingLabel="Menghapus"
                      isAsking={askingId === row.id}
                      isPending={deleteUnit.isPending && deleteUnit.variables.id === row.id}
                      onAsk={() => askDelete(row)}
                      onCancel={() => setAskingId(null)}
                      onConfirm={() => confirmDelete(row)}
                    />
                  </ActionGroup>
                ),
              },
            ]}
          />
          <Pager
            page={units.data.page}
            totalPages={units.data.totalPages}
            totalItems={units.data.totalItems}
            unit="unit"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
