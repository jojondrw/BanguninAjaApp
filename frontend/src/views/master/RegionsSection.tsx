import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useCreateRegion, useDeleteRegion, useRegions, useUpdateRegion } from '../../controllers/useMaster'
import type { OptionSource } from '../../models/lookup'
import {
  EMPTY_REGION_FORM,
  REGION_CODE_MAX_LENGTH,
  REGION_NAME_MAX_LENGTH,
  REGION_PARENT_TYPES,
  REGION_SEARCH_MAX_LENGTH,
  REGION_TYPE_LABEL,
  REGION_TYPES,
  regionFormValues,
  regionNeedsParent,
  regionParentHint,
  regionParentText,
  withRegionType,
  type Region,
  type RegionFormValues,
  type RegionType,
} from '../../models/master'
import { regionOptions } from '../../models/masterApi'
import { errorMessage } from '../../shared/errorMessage'
import { type Column, Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'
import { PAGE_SIZE } from './masterShared'

// Pilihan induk mengikuti tingkat wilayah di formulir (kota di bawah provinsi,
// kecamatan di bawah kota atau kabupaten).
const PARENT_SOURCES = Object.fromEntries(
  REGION_TYPES.map((type) => [type, regionOptions(REGION_PARENT_TYPES[type])]),
) as Record<RegionType, OptionSource>

// Penyaring induk hanya menawarkan tingkat yang bisa punya turunan.
const FILTER_PARENT_SOURCE = regionOptions(['province', 'city', 'regency'])

type ChangeRegion = (key: 'code' | 'name') => (event: ChangeEvent<HTMLInputElement>) => void

function useRegionValues(initial: RegionFormValues) {
  const [values, setValues] = useState<RegionFormValues>(initial)

  const change: ChangeRegion = (key) => (event) => setValues((current) => ({ ...current, [key]: event.target.value }))
  const pickType = (type: RegionType) => setValues((current) => withRegionType(current, type))
  const pickParent = (parentId: string) => setValues((current) => ({ ...current, parentId }))

  return { values, setValues, change, pickType, pickParent }
}

function RegionFields({ idPrefix, values, onChange, onType, onParent }: {
  idPrefix: string
  values: RegionFormValues
  onChange: ChangeRegion
  onType: (type: RegionType) => void
  onParent: (parentId: string) => void
}) {
  const needsParent = regionNeedsParent(values.type)
  const parentLabel = REGION_PARENT_TYPES[values.type].map((type) => REGION_TYPE_LABEL[type].toLowerCase()).join(' atau ')

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Field
        id={`${idPrefix}-code`}
        label="Kode wilayah"
        placeholder="3201"
        autoComplete="off"
        maxLength={REGION_CODE_MAX_LENGTH}
        hint={`Unik, maksimal ${REGION_CODE_MAX_LENGTH} karakter`}
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama wilayah"
        placeholder="Bogor"
        autoComplete="off"
        maxLength={REGION_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <SelectField
        id={`${idPrefix}-type`}
        label="Tingkat"
        hint="Tingkat wilayah yang sudah punya turunan tidak bisa diubah"
        value={values.type}
        onChange={(event) => onType(event.target.value as RegionType)}
      >
        {REGION_TYPES.map((type) => (
          <option key={type} value={type}>
            {REGION_TYPE_LABEL[type]}
          </option>
        ))}
      </SelectField>
      <SearchSelect
        key={values.type}
        {...PARENT_SOURCES[values.type]}
        id={`${idPrefix}-parent`}
        label="Wilayah induk"
        placeholder={needsParent ? `Cari ${parentLabel}` : undefined}
        hint={regionParentHint(values.type)}
        allowEmpty={!needsParent}
        emptyLabel="Tanpa induk"
        required={needsParent}
        disabled={!needsParent}
        value={values.parentId}
        onChange={onParent}
      />
    </div>
  )
}

function NewRegionForm() {
  const { values, setValues, change, pickType, pickParent } = useRegionValues(EMPTY_REGION_FORM)
  const createRegion = useCreateRegion()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createRegion.mutate(values, {
      onSuccess: () => setValues({ ...EMPTY_REGION_FORM, type: values.type, parentId: values.parentId }),
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <RegionFields idPrefix="region-new" values={values} onChange={change} onType={pickType} onParent={pickParent} />

        <Button type="submit" isPending={createRegion.isPending} pendingLabel="Menyimpan wilayah">
          Simpan wilayah
        </Button>

        {createRegion.isError ? <ErrorNote message={errorMessage(createRegion.error)} /> : null}
        {createRegion.isSuccess ? (
          <SuccessNote
            message={`${REGION_TYPE_LABEL[createRegion.data.type]} ${createRegion.data.name} berhasil ditambahkan.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditRegionForm({ region, onSaved, onCancel }: {
  region: Region
  onSaved: (saved: Region) => void
  onCancel: () => void
}) {
  const { values, change, pickType, pickParent } = useRegionValues(regionFormValues(region))
  const updateRegion = useUpdateRegion()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateRegion.mutate({ region, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">
          Ubah wilayah {region.code} {region.name}
        </h3>
        <RegionFields idPrefix="region-edit" values={values} onChange={change} onType={pickType} onParent={pickParent} />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateRegion.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan wilayah
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateRegion.isError ? <ErrorNote message={errorMessage(updateRegion.error)} /> : null}
      </form>
    </FormPanel>
  )
}

export function RegionsSection() {
  const canWrite = useScopedCan('write')
  const [search, setSearch] = useState('')
  const [type, setType] = useState<RegionType | ''>('')
  const [parentId, setParentId] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Region | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteRegion = useDeleteRegion()
  const isFormOpen = isCreating || editing !== null

  const regions = useRegions({
    search: search.trim() === '' ? undefined : search.trim(),
    type: type === '' ? undefined : type,
    parentId: parentId === '' ? undefined : parentId,
    page,
    pageSize: PAGE_SIZE,
  })

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  const clearNotes = () => {
    setNotice(null)
    deleteRegion.reset()
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

  const startEdit = (region: Region) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === region.id ? null : region))
  }

  const remove = (region: Region) => {
    deleteRegion.mutate(region, {
      onSuccess: () => {
        setNotice(`Wilayah ${region.name} dihapus.`)
        setPage((current) => pageAfterRemoval(current, regions.data?.items.length ?? 0))
        if (editing?.id === region.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  const columns: Column<Region>[] = [
    { header: 'Kode', cell: (row) => <span className="font-medium text-slate-900 tabular-nums">{row.code}</span> },
    { header: 'Nama', cell: (row) => row.name },
    { header: 'Tingkat', cell: (row) => REGION_TYPE_LABEL[row.type] ?? row.type },
    { header: 'Induk', cell: (row) => regionParentText(row) },
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
            srLabel={row.name}
            question={`Hapus ${row.name}?`}
            confirmLabel="Ya, hapus"
            pendingLabel="Menghapus"
            isAsking={askingId === row.id}
            isPending={deleteRegion.isPending && deleteRegion.variables.id === row.id}
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
      description="Provinsi, kota, kabupaten, dan kecamatan untuk proyek dan lokasi tersimpan. Wilayah yang masih punya turunan atau sudah dipakai tidak bisa dihapus."
    >
      <Toolbar>
        <ToolbarInput
          id="region-search"
          label="Cari kode atau nama wilayah"
          type="search"
          placeholder="Cari kode atau nama wilayah"
          className="w-56"
          maxLength={REGION_SEARCH_MAX_LENGTH}
          value={search}
          onChange={(event) => filterChanged(setSearch)(event.target.value)}
        />
        <FilterSelect
          id="region-filter-type"
          label="Saring menurut tingkat"
          value={type}
          onChange={(event) => filterChanged(setType)(event.target.value as RegionType | '')}
        >
          <option value="">Semua tingkat</option>
          {REGION_TYPES.map((regionType) => (
            <option key={regionType} value={regionType}>
              {REGION_TYPE_LABEL[regionType]}
            </option>
          ))}
        </FilterSelect>
        <SearchSelect
          {...FILTER_PARENT_SOURCE}
          id="region-filter-parent"
          label="Saring menurut wilayah induk"
          placeholder="Cari wilayah induk"
          compact
          allowEmpty
          emptyLabel="Semua induk"
          className="w-52"
          value={parentId}
          onChange={(value) => filterChanged(setParentId)(value)}
        />
        <FormToggle isOpen={isFormOpen} openLabel="Tambah wilayah" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewRegionForm /> : null}
      {editing ? (
        <EditRegionForm
          key={editing.id}
          region={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Wilayah ${saved.name} berhasil diperbarui.`)
          }}
        />
      ) : null}

      <RowNotice
        success={notice}
        error={
          deleteRegion.isError
            ? refusalText(`Wilayah ${deleteRegion.variables.name} tidak bisa dihapus.`, deleteRegion.error)
            : null
        }
      />

      {regions.isPending ? <Loading /> : null}
      {regions.isError ? <LoadFailed onRetry={() => regions.refetch()} /> : null}
      {regions.data ? (
        <>
          <Table rows={regions.data.items} emptyMessage="Tidak ada wilayah yang cocok dengan penyaringan ini." columns={columns} />
          <Pager
            page={regions.data.page}
            totalPages={regions.data.totalPages}
            totalItems={regions.data.totalItems}
            unit="wilayah"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
