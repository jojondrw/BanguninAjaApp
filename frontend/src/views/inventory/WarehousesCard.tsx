import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useCreateWarehouse, useDeleteWarehouse, useUpdateWarehouse, useWarehouses } from '../../controllers/useInventory'
import {
  EMPTY_WAREHOUSE_FORM,
  WAREHOUSE_CODE_MAX_LENGTH,
  WAREHOUSE_NAME_MAX_LENGTH,
  warehouseFormValues,
  type Warehouse,
  type WarehouseFormValues,
} from '../../models/inventory'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, SelectField, Toolbar } from '../components/RecordControls'
import { pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'
import { PAGE_SIZE, projectLabel, type Lookups } from './inventoryShared'

type ChangeWarehouse = (key: keyof WarehouseFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

function useWarehouseValues(initial: WarehouseFormValues) {
  const [values, setValues] = useState<WarehouseFormValues>(initial)

  const change: ChangeWarehouse = (key) => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

function WarehouseFields({ idPrefix, values, onChange, projects }: {
  idPrefix: string
  values: WarehouseFormValues
  onChange: ChangeWarehouse
  projects: Project[]
}) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Field
        id={`${idPrefix}-code`}
        label="Kode"
        placeholder="GDG-01"
        autoComplete="off"
        maxLength={WAREHOUSE_CODE_MAX_LENGTH}
        hint={`Unik, maksimal ${WAREHOUSE_CODE_MAX_LENGTH} karakter`}
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama gudang"
        placeholder="Gudang Lapangan Blok A"
        autoComplete="off"
        maxLength={WAREHOUSE_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <SelectField
        id={`${idPrefix}-project`}
        label="Proyek (opsional)"
        hint="Kosongkan untuk gudang pusat"
        value={values.projectId}
        onChange={onChange('projectId')}
      >
        <option value="">Gudang pusat, tanpa proyek</option>
        {projects.map((project) => (
          <option key={project.id} value={project.id}>
            {project.name}
          </option>
        ))}
      </SelectField>
    </div>
  )
}

function NewWarehouseForm({ projects }: { projects: Project[] }) {
  const { values, setValues, change } = useWarehouseValues(EMPTY_WAREHOUSE_FORM)
  const createWarehouse = useCreateWarehouse()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createWarehouse.mutate(values, { onSuccess: () => setValues(EMPTY_WAREHOUSE_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <WarehouseFields idPrefix="warehouse-new" values={values} onChange={change} projects={projects} />

        <Button type="submit" isPending={createWarehouse.isPending} pendingLabel="Menyimpan gudang">
          Simpan gudang
        </Button>

        {createWarehouse.isError ? <ErrorNote message={errorMessage(createWarehouse.error)} /> : null}
        {createWarehouse.isSuccess ? (
          <SuccessNote message={`Gudang ${createWarehouse.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditWarehouseForm({ warehouse, projects, onSaved, onCancel }: {
  warehouse: Warehouse
  projects: Project[]
  onSaved: (saved: Warehouse) => void
  onCancel: () => void
}) {
  const { values, change } = useWarehouseValues(warehouseFormValues(warehouse))
  const updateWarehouse = useUpdateWarehouse()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateWarehouse.mutate({ warehouse, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">Ubah gudang {warehouse.code}</h3>
        <WarehouseFields idPrefix="warehouse-edit" values={values} onChange={change} projects={projects} />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateWarehouse.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan gudang
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateWarehouse.isError ? <ErrorNote message={errorMessage(updateWarehouse.error)} /> : null}
      </form>
    </FormPanel>
  )
}

export function WarehousesCard({ lookups }: { lookups: Lookups }) {
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Warehouse | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteWarehouse = useDeleteWarehouse()
  const isFormOpen = isCreating || editing !== null
  const warehouses = useWarehouses({
    projectId: projectId === '' ? undefined : projectId,
    page,
    pageSize: PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    deleteWarehouse.reset()
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

  const startEdit = (warehouse: Warehouse) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === warehouse.id ? null : warehouse))
  }

  const remove = (warehouse: Warehouse) => {
    deleteWarehouse.mutate(warehouse, {
      onSuccess: () => {
        setNotice(`Gudang ${warehouse.name} dihapus.`)
        setPage((current) => pageAfterRemoval(current, warehouses.data?.items.length ?? 0))
        if (editing?.id === warehouse.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card
      title="Gudang"
      description="Gudang pusat dan gudang lapangan per proyek. Gudang yang sudah punya mutasi stok atau penerimaan barang tidak bisa dihapus."
    >
      <Toolbar>
        <FilterSelect
          id="warehouse-filter-project"
          label="Saring menurut proyek"
          value={projectId}
          onChange={(event) => {
            setProjectId(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Semua proyek</option>
          {lookups.projects.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name}
            </option>
          ))}
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Tambah gudang" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewWarehouseForm projects={lookups.projects} /> : null}
      {editing ? (
        <EditWarehouseForm
          key={editing.id}
          warehouse={editing}
          projects={lookups.projects}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Gudang ${saved.name} berhasil diperbarui.`)
          }}
        />
      ) : null}

      <RowNotice
        success={notice}
        error={
          deleteWarehouse.isError
            ? refusalText(`Gudang ${deleteWarehouse.variables.name} tidak bisa dihapus.`, deleteWarehouse.error)
            : null
        }
      />

      {warehouses.isPending ? <Loading /> : null}
      {warehouses.isError ? <LoadFailed onRetry={() => warehouses.refetch()} /> : null}
      {warehouses.data ? (
        <>
          <Table
            rows={warehouses.data.items}
            emptyMessage="Belum ada gudang untuk penyaringan ini."
            columns={[
              { header: 'Kode', cell: (row) => row.code },
              { header: 'Nama', cell: (row) => row.name },
              { header: 'Proyek', cell: (row) => projectLabel(lookups, row.projectId, 'Gudang pusat') },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <ActionGroup>
                    {askingId === row.id ? null : (
                      <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                    )}
                    <ConfirmAction
                      label="Hapus"
                      srLabel={row.name}
                      question={`Hapus ${row.name}?`}
                      confirmLabel="Ya, hapus"
                      pendingLabel="Menghapus"
                      isAsking={askingId === row.id}
                      isPending={deleteWarehouse.isPending && deleteWarehouse.variables.id === row.id}
                      onAsk={() => {
                        clearNotes()
                        setAskingId(row.id)
                      }}
                      onCancel={() => setAskingId(null)}
                      onConfirm={() => remove(row)}
                    />
                  </ActionGroup>
                ),
              },
            ]}
          />
          <Pager
            page={warehouses.data.page}
            totalPages={warehouses.data.totalPages}
            totalItems={warehouses.data.totalItems}
            unit="gudang"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
