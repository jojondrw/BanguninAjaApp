import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useCreateMaterial, useDeleteMaterial, useMaterials, useUpdateMaterial } from '../../controllers/useInventory'
import {
  EMPTY_MATERIAL_FORM,
  MATERIAL_CATEGORY_MAX_LENGTH,
  MATERIAL_CATEGORY_SUGGESTIONS,
  MATERIAL_CODE_MAX_LENGTH,
  MATERIAL_NAME_MAX_LENGTH,
  materialFormValues,
  type Material,
  type MaterialFormValues,
  type UnitOfMeasure,
} from '../../models/inventory'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FormPanel, FormToggle, Pager, SelectField, Toolbar, ToolbarInput } from '../components/RecordControls'
import { amountHint, pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'
import { PAGE_SIZE, quantityText, unitCode, type Lookups } from './inventoryShared'

const CATEGORY_OPTIONS_ID = 'material-category-options'

type ChangeMaterial = (key: keyof MaterialFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

function useMaterialValues(initial: MaterialFormValues) {
  const [values, setValues] = useState<MaterialFormValues>(initial)

  const change: ChangeMaterial = (key) => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

function MaterialFields({ idPrefix, values, onChange, units }: {
  idPrefix: string
  values: MaterialFormValues
  onChange: ChangeMaterial
  units: UnitOfMeasure[]
}) {
  const unit = units.find((item) => item.id === values.unitOfMeasureId)?.code ?? ''

  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Field
        id={`${idPrefix}-code`}
        label="Kode"
        placeholder="MAT-001"
        autoComplete="off"
        maxLength={MATERIAL_CODE_MAX_LENGTH}
        hint={`Unik, maksimal ${MATERIAL_CODE_MAX_LENGTH} karakter`}
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama material"
        placeholder="Semen Portland 50 kg"
        autoComplete="off"
        maxLength={MATERIAL_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <Field
        id={`${idPrefix}-category`}
        label="Kategori (opsional)"
        placeholder="Pilih atau ketik kategori"
        list={CATEGORY_OPTIONS_ID}
        autoComplete="off"
        maxLength={MATERIAL_CATEGORY_MAX_LENGTH}
        value={values.category}
        onChange={onChange('category')}
      />
      <SelectField
        id={`${idPrefix}-unit`}
        label="Satuan"
        required
        value={values.unitOfMeasureId}
        onChange={onChange('unitOfMeasureId')}
      >
        <option value="">Pilih satuan</option>
        {units.map((item) => (
          <option key={item.id} value={item.id}>
            {item.name} ({item.code})
          </option>
        ))}
      </SelectField>
      <Field
        id={`${idPrefix}-minimum`}
        label={unit === '' ? 'Stok minimum' : `Stok minimum (${unit})`}
        type="number"
        inputMode="decimal"
        min={0}
        step={0.01}
        placeholder="0"
        hint="Di bawah angka ini material masuk daftar stok menipis"
        value={values.minimumStock}
        onChange={onChange('minimumStock')}
      />
      <Field
        id={`${idPrefix}-price`}
        label="Harga terakhir (opsional)"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        hint={amountHint(values.lastPrice, unit === '' ? 'Dalam Rupiah per satuan' : `Dalam Rupiah per ${unit}`)}
        value={values.lastPrice}
        onChange={onChange('lastPrice')}
      />
    </div>
  )
}

function NewMaterialForm({ units }: { units: UnitOfMeasure[] }) {
  const { values, setValues, change } = useMaterialValues(EMPTY_MATERIAL_FORM)
  const createMaterial = useCreateMaterial()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createMaterial.mutate(values, { onSuccess: () => setValues(EMPTY_MATERIAL_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <MaterialFields idPrefix="material-new" values={values} onChange={change} units={units} />

        <Button type="submit" isPending={createMaterial.isPending} pendingLabel="Menyimpan material">
          Simpan material
        </Button>

        {createMaterial.isError ? <ErrorNote message={errorMessage(createMaterial.error)} /> : null}
        {createMaterial.isSuccess ? (
          <SuccessNote message={`Material ${createMaterial.data.name} (${createMaterial.data.code}) berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditMaterialForm({ material, units, onSaved, onCancel }: {
  material: Material
  units: UnitOfMeasure[]
  onSaved: (saved: Material) => void
  onCancel: () => void
}) {
  const { values, change } = useMaterialValues(materialFormValues(material))
  const updateMaterial = useUpdateMaterial()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateMaterial.mutate({ material, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">Ubah material {material.code}</h3>
        <MaterialFields idPrefix="material-edit" values={values} onChange={change} units={units} />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateMaterial.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan material
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateMaterial.isError ? <ErrorNote message={errorMessage(updateMaterial.error)} /> : null}
      </form>
    </FormPanel>
  )
}

export function MaterialsCard({ lookups }: { lookups: Lookups }) {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Material | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteMaterial = useDeleteMaterial()
  const isFormOpen = isCreating || editing !== null
  const materials = useMaterials({
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    deleteMaterial.reset()
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

  const startEdit = (material: Material) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === material.id ? null : material))
  }

  const remove = (material: Material) => {
    deleteMaterial.mutate(material, {
      onSuccess: () => {
        setNotice(`Material ${material.name} (${material.code}) dihapus.`)
        setPage((current) => pageAfterRemoval(current, materials.data?.items.length ?? 0))
        if (editing?.id === material.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card
      title="Material"
      description="Katalog material beserta satuan dan stok minimumnya. Material yang sudah punya mutasi stok atau dipakai dokumen pengadaan tidak bisa dihapus."
    >
      <Toolbar>
        <ToolbarInput
          id="material-search"
          label="Cari nama atau kode material"
          type="search"
          placeholder="Cari nama atau kode"
          className="w-56"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        <FormToggle isOpen={isFormOpen} openLabel="Tambah material" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewMaterialForm units={lookups.units} /> : null}
      {editing ? (
        <EditMaterialForm
          key={editing.id}
          material={editing}
          units={lookups.units}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Material ${saved.name} (${saved.code}) berhasil diperbarui.`)
          }}
        />
      ) : null}

      <datalist id={CATEGORY_OPTIONS_ID}>
        {MATERIAL_CATEGORY_SUGGESTIONS.map((category) => (
          <option key={category} value={category} />
        ))}
      </datalist>

      <RowNotice
        success={notice}
        error={
          deleteMaterial.isError
            ? refusalText(`Material ${deleteMaterial.variables.name} tidak bisa dihapus.`, deleteMaterial.error)
            : null
        }
      />

      {materials.isPending ? <Loading /> : null}
      {materials.isError ? <LoadFailed onRetry={() => materials.refetch()} /> : null}
      {materials.data ? (
        <>
          <Table
            rows={materials.data.items}
            emptyMessage="Tidak ada material yang cocok."
            columns={[
              { header: 'Kode', cell: (row) => row.code },
              { header: 'Nama', cell: (row) => row.name },
              { header: 'Kategori', cell: (row) => row.category || '-' },
              { header: 'Satuan', cell: (row) => unitCode(lookups, row.unitOfMeasureId) || '-' },
              {
                header: 'Stok minimum',
                align: 'right',
                cell: (row) => quantityText(row.minimumStock, unitCode(lookups, row.unitOfMeasureId)),
              },
              { header: 'Harga terakhir', align: 'right', cell: (row) => rupiah(row.lastPrice) },
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
                      isPending={deleteMaterial.isPending && deleteMaterial.variables.id === row.id}
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
            page={materials.data.page}
            totalPages={materials.data.totalPages}
            totalItems={materials.data.totalItems}
            unit="material"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
