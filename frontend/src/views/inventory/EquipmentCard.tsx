import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreateEquipment,
  useDeleteEquipment,
  useEquipment,
  useUpdateEquipment,
} from '../../controllers/useInventory'
import {
  EMPTY_EQUIPMENT_FORM,
  EQUIPMENT_CODE_MAX_LENGTH,
  EQUIPMENT_NAME_MAX_LENGTH,
  EQUIPMENT_STATUSES,
  EQUIPMENT_STATUS_LABEL,
  EQUIPMENT_STATUS_TONE,
  equipmentFormValues,
  type Equipment,
  type EquipmentFormValues,
  type EquipmentStatus,
} from '../../models/inventory'
import { assetOptions, parentAssetOptions, projectOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import {
  Chip,
  FilterSelect,
  FormPanel,
  FormToggle,
  Pager,
  SelectField,
  Toolbar,
  ToolbarInput,
} from '../components/RecordControls'
import { pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'
import { LookupName } from '../components/LookupName'
import { SearchSelect } from '../components/SearchSelect'

const EQUIPMENT_PAGE_SIZE = 10

type ChangeEquipment = (
  key: keyof EquipmentFormValues,
) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

function useEquipmentValues(initial: EquipmentFormValues) {
  const [values, setValues] = useState<EquipmentFormValues>(initial)

  const change: ChangeEquipment = (key) => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const pick = (key: 'projectId' | 'assetId') => (value: string) =>
    setValues((current) => ({ ...current, [key]: value }))

  return { values, setValues, change, pick }
}

// Backend menolak alat yang beroperasi tanpa proyek (equipment_project_required),
// jadi proyek wajib diisi begitu status Beroperasi dipilih.
function EquipmentFields({ idPrefix, values, onChange, onPick }: {
  idPrefix: string
  values: EquipmentFormValues
  onChange: ChangeEquipment
  onPick: (key: 'projectId' | 'assetId') => (value: string) => void
}) {
  const isOperating = values.status === 'operating'

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <Field
        id={`${idPrefix}-code`}
        label="Kode"
        placeholder="ALT-001"
        autoComplete="off"
        maxLength={EQUIPMENT_CODE_MAX_LENGTH}
        hint={`Unik, maksimal ${EQUIPMENT_CODE_MAX_LENGTH} karakter`}
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama alat"
        placeholder="Concrete mixer 350 L"
        autoComplete="off"
        maxLength={EQUIPMENT_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <SelectField id={`${idPrefix}-status`} label="Status" value={values.status} onChange={onChange('status')}>
        {EQUIPMENT_STATUSES.map((status) => (
          <option key={status} value={status}>
            {EQUIPMENT_STATUS_LABEL[status]}
          </option>
        ))}
      </SelectField>
      <SearchSelect
        {...projectOptions}
        id={`${idPrefix}-project`}
        label={isOperating ? 'Proyek' : 'Proyek (opsional)'}
        placeholder="Cari proyek"
        hint={isOperating ? 'Wajib untuk alat yang beroperasi' : undefined}
        required={isOperating}
        allowEmpty={!isOperating}
        emptyLabel="Belum ditugaskan"
        value={values.projectId}
        onChange={onPick('projectId')}
      />
      <SearchSelect
        {...parentAssetOptions}
        id={`${idPrefix}-asset`}
        label="Aset induk (opsional)"
        placeholder="Cari aset"
        hint="Aset yang sudah dilepas tidak bisa dipilih"
        allowEmpty
        emptyLabel="Tanpa aset induk"
        value={values.assetId}
        onChange={onPick('assetId')}
      />
      <Field
        id={`${idPrefix}-hours`}
        label="Jam operasi"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        value={values.operatingHours}
        onChange={onChange('operatingHours')}
      />
      <Field
        id={`${idPrefix}-service`}
        label="Servis berikutnya (opsional)"
        type="date"
        value={values.nextServiceDate}
        onChange={onChange('nextServiceDate')}
      />
    </div>
  )
}

function NewEquipmentForm() {
  const { values, setValues, change, pick } = useEquipmentValues(EMPTY_EQUIPMENT_FORM)
  const createEquipment = useCreateEquipment()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createEquipment.mutate(values, { onSuccess: () => setValues(EMPTY_EQUIPMENT_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <EquipmentFields idPrefix="equipment-new" values={values} onChange={change} onPick={pick} />

        <Button type="submit" isPending={createEquipment.isPending} pendingLabel="Menyimpan alat">
          Simpan alat
        </Button>

        {createEquipment.isError ? <ErrorNote message={errorMessage(createEquipment.error)} /> : null}
        {createEquipment.isSuccess ? (
          <SuccessNote message={`Alat ${createEquipment.data.name} (${createEquipment.data.code}) berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditEquipmentForm({ equipment, onSaved, onCancel }: {
  equipment: Equipment
  onSaved: (saved: Equipment) => void
  onCancel: () => void
}) {
  const { values, change, pick } = useEquipmentValues(equipmentFormValues(equipment))
  const updateEquipment = useUpdateEquipment()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateEquipment.mutate({ equipment, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">Ubah alat {equipment.code}</h3>
        <EquipmentFields idPrefix="equipment-edit" values={values} onChange={change} onPick={pick} />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateEquipment.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan alat
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateEquipment.isError ? <ErrorNote message={errorMessage(updateEquipment.error)} /> : null}
      </form>
    </FormPanel>
  )
}

function ServiceDate({ date, today }: { date: string | null; today: string }) {
  if (date === null) {
    return <>-</>
  }
  if (date.slice(0, 10) < today) {
    return <span className="text-red-700">{shortDate(date)}, terlewat</span>
  }
  return <>{shortDate(date)}</>
}

export function EquipmentCard({ today }: { today: string }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<EquipmentStatus | ''>('')
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Equipment | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteEquipment = useDeleteEquipment()
  const isFormOpen = isCreating || editing !== null
  const equipment = useEquipment({
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    projectId: projectId === '' ? undefined : projectId,
    page,
    pageSize: EQUIPMENT_PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    deleteEquipment.reset()
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

  const startEdit = (item: Equipment) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === item.id ? null : item))
  }

  const remove = (item: Equipment) => {
    deleteEquipment.mutate(item, {
      onSuccess: () => {
        setNotice(`Alat ${item.name} (${item.code}) dihapus.`)
        setPage((current) => pageAfterRemoval(current, equipment.data?.items.length ?? 0))
        if (editing?.id === item.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card title="Alat berat dan peralatan" description="Alat yang beroperasi wajib ditugaskan ke proyek.">
      <Toolbar>
        <ToolbarInput
          id="equipment-search"
          label="Cari nama atau kode alat"
          type="search"
          placeholder="Cari nama atau kode"
          className="w-56"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        <FilterSelect
          id="equipment-filter-status"
          label="Saring menurut status alat"
          value={status}
          onChange={(event) => {
            setStatus(event.target.value as EquipmentStatus | '')
            setPage(1)
          }}
        >
          <option value="">Semua status</option>
          {EQUIPMENT_STATUSES.map((option) => (
            <option key={option} value={option}>
              {EQUIPMENT_STATUS_LABEL[option]}
            </option>
          ))}
        </FilterSelect>
        <SearchSelect
          {...projectOptions}
          id="equipment-filter-project"
          label="Saring menurut proyek"
          compact
          allowEmpty
          emptyLabel="Semua proyek"
          className="w-52"
          value={projectId}
          onChange={(value) => {
            setProjectId(value)
            setPage(1)
          }}
        />
        <FormToggle isOpen={isFormOpen} openLabel="Tambah alat" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewEquipmentForm /> : null}
      {editing ? (
        <EditEquipmentForm
          key={editing.id}
          equipment={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Alat ${saved.name} (${saved.code}) berhasil diperbarui.`)
          }}
        />
      ) : null}

      <RowNotice
        success={notice}
        error={
          deleteEquipment.isError
            ? refusalText(`Alat ${deleteEquipment.variables.name} tidak bisa dihapus.`, deleteEquipment.error)
            : null
        }
      />

      {equipment.isPending ? <Loading /> : null}
      {equipment.isError ? <LoadFailed onRetry={() => equipment.refetch()} /> : null}
      {equipment.data ? (
        <>
          <Table
            rows={equipment.data.items}
            emptyMessage="Belum ada alat yang cocok. Tambahkan lewat tombol Tambah alat."
            columns={[
              {
                header: 'Alat',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="font-medium text-slate-900">{row.name}</span>
                    <span className="text-xs text-slate-500">{row.code}</span>
                  </span>
                ),
              },
              {
                header: 'Proyek',
                cell: (row) => (row.projectId ? <LookupName source={projectOptions} value={row.projectId} /> : '-'),
              },
              {
                header: 'Aset induk',
                cell: (row) => (row.assetId ? <LookupName source={assetOptions} value={row.assetId} /> : '-'),
              },
              { header: 'Jam operasi', align: 'right', cell: (row) => number(row.operatingHours) },
              {
                header: 'Servis berikutnya',
                align: 'right',
                cell: (row) => <ServiceDate date={row.nextServiceDate} today={today} />,
              },
              {
                header: 'Status',
                cell: (row) => <Chip tone={EQUIPMENT_STATUS_TONE[row.status]}>{EQUIPMENT_STATUS_LABEL[row.status]}</Chip>,
              },
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
                      isPending={deleteEquipment.isPending && deleteEquipment.variables.id === row.id}
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
            page={equipment.data.page}
            totalPages={equipment.data.totalPages}
            totalItems={equipment.data.totalItems}
            unit="alat"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
