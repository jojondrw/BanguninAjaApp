import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import {
  useCreateUnitOfMeasure,
  useDeleteUnitOfMeasure,
  useUnitsOfMeasure,
  useUpdateUnitOfMeasure,
} from '../../controllers/useMaster'
import {
  EMPTY_UNIT_FORM,
  UNIT_CODE_MAX_LENGTH,
  UNIT_NAME_MAX_LENGTH,
  UNIT_SEARCH_MAX_LENGTH,
  unitFormValues,
  type UnitOfMeasure,
  type UnitOfMeasureFormValues,
} from '../../models/master'
import { errorMessage } from '../../shared/errorMessage'
import { type Column, Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'
import { PAGE_SIZE } from './masterShared'

type ChangeUnit = (key: keyof UnitOfMeasureFormValues) => (event: ChangeEvent<HTMLInputElement>) => void

function useUnitValues(initial: UnitOfMeasureFormValues) {
  const [values, setValues] = useState<UnitOfMeasureFormValues>(initial)

  const change: ChangeUnit = (key) => (event) => setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

function UnitFields({ idPrefix, values, onChange }: {
  idPrefix: string
  values: UnitOfMeasureFormValues
  onChange: ChangeUnit
}) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Field
        id={`${idPrefix}-code`}
        label="Kode satuan"
        placeholder="m3"
        autoComplete="off"
        maxLength={UNIT_CODE_MAX_LENGTH}
        hint="Unik, ditulis seperti di dokumen, misalnya m3 atau zak"
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama satuan"
        placeholder="Meter kubik"
        autoComplete="off"
        maxLength={UNIT_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
    </div>
  )
}

function NewUnitForm() {
  const { values, setValues, change } = useUnitValues(EMPTY_UNIT_FORM)
  const createUnit = useCreateUnitOfMeasure()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createUnit.mutate(values, { onSuccess: () => setValues(EMPTY_UNIT_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <UnitFields idPrefix="unit-new" values={values} onChange={change} />

        <Button type="submit" isPending={createUnit.isPending} pendingLabel="Menyimpan satuan">
          Simpan satuan
        </Button>

        {createUnit.isError ? <ErrorNote message={errorMessage(createUnit.error)} /> : null}
        {createUnit.isSuccess ? (
          <SuccessNote message={`Satuan ${createUnit.data.code} (${createUnit.data.name}) berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditUnitForm({ unit, onSaved, onCancel }: {
  unit: UnitOfMeasure
  onSaved: (saved: UnitOfMeasure) => void
  onCancel: () => void
}) {
  const { values, change } = useUnitValues(unitFormValues(unit))
  const updateUnit = useUpdateUnitOfMeasure()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateUnit.mutate({ unit, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">Ubah satuan {unit.code}</h3>
        <UnitFields idPrefix="unit-edit" values={values} onChange={change} />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateUnit.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan satuan
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateUnit.isError ? <ErrorNote message={errorMessage(updateUnit.error)} /> : null}
      </form>
    </FormPanel>
  )
}

export function UnitsSection() {
  const canWrite = useScopedCan('write')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<UnitOfMeasure | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteUnit = useDeleteUnitOfMeasure()
  const isFormOpen = isCreating || editing !== null

  const units = useUnitsOfMeasure({
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    deleteUnit.reset()
  }

  const closeForm = () => {
    setIsCreating(false)
    setEditing(null)
  }

  const toggleForm = () => {
    clearNotes()
    if (isFormOpen) {
      closeForm()
      return
    }
    setIsCreating(true)
  }

  const startEdit = (unit: UnitOfMeasure) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === unit.id ? null : unit))
  }

  const remove = (unit: UnitOfMeasure) => {
    deleteUnit.mutate(unit, {
      onSuccess: () => {
        setNotice(`Satuan ${unit.code} dihapus.`)
        setPage((current) => pageAfterRemoval(current, units.data?.items.length ?? 0))
        if (editing?.id === unit.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  const columns: Column<UnitOfMeasure>[] = [
    { header: 'Kode', cell: (row) => <span className="font-medium text-slate-900">{row.code}</span> },
    { header: 'Nama', cell: (row) => row.name },
  ]

  if (canWrite) {
    columns.push({
      header: 'Aksi',
      align: 'right',
      cell: (row) => (
        <ActionGroup>
          {askingId === row.id ? null : (
            <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
          )}
          <ConfirmAction
            label="Hapus"
            srLabel={row.code}
            question={`Hapus satuan ${row.code}?`}
            confirmLabel="Ya, hapus"
            pendingLabel="Menghapus"
            isAsking={askingId === row.id}
            isPending={deleteUnit.isPending && deleteUnit.variables.id === row.id}
            onAsk={() => {
              clearNotes()
              setAskingId(row.id)
            }}
            onCancel={() => setAskingId(null)}
            onConfirm={() => remove(row)}
          />
        </ActionGroup>
      ),
    })
  }

  return (
    <Card
      title="Satuan"
      description="Satuan untuk material, RAB proyek, dan pengadaan. Satuan yang sudah dipakai tidak bisa dihapus."
    >
      <Toolbar>
        <ToolbarInput
          id="unit-search"
          label="Cari kode atau nama satuan"
          type="search"
          placeholder="Cari kode atau nama satuan"
          className="w-56"
          maxLength={UNIT_SEARCH_MAX_LENGTH}
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        <FormToggle isOpen={isFormOpen} openLabel="Tambah satuan" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewUnitForm /> : null}
      {editing ? (
        <EditUnitForm
          key={editing.id}
          unit={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Satuan ${saved.code} berhasil diperbarui.`)
          }}
        />
      ) : null}

      <RowNotice
        success={notice}
        error={
          deleteUnit.isError
            ? refusalText(`Satuan ${deleteUnit.variables.code} tidak bisa dihapus.`, deleteUnit.error)
            : null
        }
      />

      {units.isPending ? <Loading /> : null}
      {units.isError ? <LoadFailed onRetry={() => units.refetch()} /> : null}
      {units.data ? (
        <>
          <Table rows={units.data.items} emptyMessage="Tidak ada satuan yang cocok dengan pencarian ini." columns={columns} />
          <Pager
            page={units.data.page}
            totalPages={units.data.totalPages}
            totalItems={units.data.totalItems}
            unit="satuan"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
