import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useAssets, useCreateAsset, useDeleteAsset, useUpdateAsset } from '../../controllers/useInventory'
import {
  ASSET_CATEGORY_MAX_LENGTH,
  ASSET_CATEGORY_SUGGESTIONS,
  ASSET_CODE_MAX_LENGTH,
  ASSET_NAME_MAX_LENGTH,
  ASSET_STATUSES,
  ASSET_STATUS_LABEL,
  ASSET_STATUS_TONE,
  EMPTY_ASSET_FORM,
  assetFormValues,
  type Asset,
  type AssetFormValues,
  type AssetStatus,
} from '../../models/inventory'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, rupiahShort, shortDate } from '../../shared/format'
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
import { amountHint, pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'

const ASSET_PAGE_SIZE = 10
const CATEGORY_OPTIONS_ID = 'asset-category-options'

type ChangeAsset = (key: keyof AssetFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

function useAssetValues(initial: AssetFormValues) {
  const [values, setValues] = useState<AssetFormValues>(initial)

  const change: ChangeAsset = (key) => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

// Penyusutan per bulan dihitung backend dengan cara yang sama: nilai
// perolehan dibagi umur manfaat.
function depreciationHint(values: AssetFormValues): string {
  const value = Number(values.acquisitionValue)
  const months = Number(values.usefulLifeMonths)
  if (values.usefulLifeMonths === '' || !Number.isFinite(value) || !Number.isFinite(months) || months <= 0) {
    return 'Kosongkan kalau aset tidak disusutkan'
  }
  return `Penyusutan sekitar ${rupiah(Math.floor(value / months))} per bulan`
}

function AssetFields({ idPrefix, values, onChange }: {
  idPrefix: string
  values: AssetFormValues
  onChange: ChangeAsset
}) {
  const acquisitionValue = values.acquisitionValue === '' ? 0 : Number(values.acquisitionValue)

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <Field
        id={`${idPrefix}-code`}
        label="Kode"
        placeholder="AST-001"
        autoComplete="off"
        maxLength={ASSET_CODE_MAX_LENGTH}
        hint={`Unik, maksimal ${ASSET_CODE_MAX_LENGTH} karakter`}
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama aset"
        placeholder="Excavator Komatsu PC200"
        autoComplete="off"
        maxLength={ASSET_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <Field
        id={`${idPrefix}-category`}
        label="Kategori"
        placeholder="Pilih atau ketik kategori"
        list={CATEGORY_OPTIONS_ID}
        autoComplete="off"
        maxLength={ASSET_CATEGORY_MAX_LENGTH}
        required
        value={values.category}
        onChange={onChange('category')}
      />
      <SelectField id={`${idPrefix}-status`} label="Status" value={values.status} onChange={onChange('status')}>
        {ASSET_STATUSES.map((status) => (
          <option key={status} value={status}>
            {ASSET_STATUS_LABEL[status]}
          </option>
        ))}
      </SelectField>
      <Field
        id={`${idPrefix}-acquired`}
        label="Tanggal perolehan (opsional)"
        type="date"
        value={values.acquisitionDate}
        onChange={onChange('acquisitionDate')}
      />
      <Field
        id={`${idPrefix}-value`}
        label="Nilai perolehan"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        hint={amountHint(values.acquisitionValue, 'Dalam Rupiah')}
        value={values.acquisitionValue}
        onChange={onChange('acquisitionValue')}
      />
      <Field
        id={`${idPrefix}-depreciation`}
        label="Akumulasi penyusutan"
        type="number"
        inputMode="numeric"
        min={0}
        max={Number.isFinite(acquisitionValue) ? acquisitionValue : undefined}
        step={1}
        placeholder="0"
        hint={amountHint(values.accumulatedDepreciation, 'Tidak boleh melebihi nilai perolehan')}
        value={values.accumulatedDepreciation}
        onChange={onChange('accumulatedDepreciation')}
      />
      <Field
        id={`${idPrefix}-life`}
        label="Umur manfaat (bulan)"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        hint={depreciationHint(values)}
        value={values.usefulLifeMonths}
        onChange={onChange('usefulLifeMonths')}
      />
    </div>
  )
}

function NewAssetForm() {
  const { values, setValues, change } = useAssetValues(EMPTY_ASSET_FORM)
  const createAsset = useCreateAsset()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createAsset.mutate(values, { onSuccess: () => setValues(EMPTY_ASSET_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <AssetFields idPrefix="asset-new" values={values} onChange={change} />

        <Button type="submit" isPending={createAsset.isPending} pendingLabel="Menyimpan aset">
          Simpan aset
        </Button>

        {createAsset.isError ? <ErrorNote message={errorMessage(createAsset.error)} /> : null}
        {createAsset.isSuccess ? (
          <SuccessNote message={`Aset ${createAsset.data.name} (${createAsset.data.code}) berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditAssetForm({ asset, onSaved, onCancel }: {
  asset: Asset
  onSaved: (saved: Asset) => void
  onCancel: () => void
}) {
  const { values, change } = useAssetValues(assetFormValues(asset))
  const updateAsset = useUpdateAsset()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateAsset.mutate({ asset, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">Ubah aset {asset.code}</h3>
        <AssetFields idPrefix="asset-edit" values={values} onChange={change} />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateAsset.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan aset
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateAsset.isError ? <ErrorNote message={errorMessage(updateAsset.error)} /> : null}
      </form>
    </FormPanel>
  )
}

export function AssetsCard() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<AssetStatus | ''>('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Asset | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteAsset = useDeleteAsset()
  const isFormOpen = isCreating || editing !== null
  const assets = useAssets({
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    page,
    pageSize: ASSET_PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    deleteAsset.reset()
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

  const startEdit = (asset: Asset) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === asset.id ? null : asset))
  }

  const remove = (asset: Asset) => {
    deleteAsset.mutate(asset, {
      onSuccess: () => {
        setNotice(`Aset ${asset.name} (${asset.code}) dihapus.`)
        setPage((current) => pageAfterRemoval(current, assets.data?.items.length ?? 0))
        if (editing?.id === asset.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card title="Aset" description="Nilai buku dihitung server dari nilai perolehan dikurangi akumulasi penyusutan.">
      <Toolbar>
        <ToolbarInput
          id="asset-search"
          label="Cari nama atau kode aset"
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
          id="asset-filter-status"
          label="Saring menurut status aset"
          value={status}
          onChange={(event) => {
            setStatus(event.target.value as AssetStatus | '')
            setPage(1)
          }}
        >
          <option value="">Semua status</option>
          {ASSET_STATUSES.map((option) => (
            <option key={option} value={option}>
              {ASSET_STATUS_LABEL[option]}
            </option>
          ))}
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Tambah aset" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewAssetForm /> : null}
      {editing ? (
        <EditAssetForm
          key={editing.id}
          asset={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Aset ${saved.name} (${saved.code}) berhasil diperbarui.`)
          }}
        />
      ) : null}

      <datalist id={CATEGORY_OPTIONS_ID}>
        {ASSET_CATEGORY_SUGGESTIONS.map((category) => (
          <option key={category} value={category} />
        ))}
      </datalist>

      <RowNotice
        success={notice}
        error={deleteAsset.isError ? refusalText(`Aset ${deleteAsset.variables.name} tidak bisa dihapus.`, deleteAsset.error) : null}
      />

      {assets.isPending ? <Loading /> : null}
      {assets.isError ? <LoadFailed onRetry={() => assets.refetch()} /> : null}
      {assets.data ? (
        <>
          <dl className="mb-4 grid gap-3 sm:grid-cols-3">
            {[
              { label: 'Nilai perolehan', value: assets.data.totalAcquisitionValue },
              { label: 'Akumulasi penyusutan', value: assets.data.totalAccumulatedDepreciation },
              { label: 'Nilai buku', value: assets.data.totalBookValue },
            ].map((item) => (
              <div key={item.label} className="rounded-lg bg-slate-50 px-3 py-2">
                <dt className="text-xs text-slate-500">{item.label}</dt>
                <dd className="text-sm font-semibold text-slate-900 tabular-nums">{rupiahShort(item.value)}</dd>
              </div>
            ))}
          </dl>
          <Table
            rows={assets.data.items}
            emptyMessage="Belum ada aset yang cocok. Tambahkan lewat tombol Tambah aset."
            columns={[
              {
                header: 'Aset',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="font-medium text-slate-900">{row.name}</span>
                    <span className="text-xs text-slate-500">{row.code}</span>
                  </span>
                ),
              },
              { header: 'Kategori', cell: (row) => row.category },
              { header: 'Perolehan', cell: (row) => shortDate(row.acquisitionDate) },
              { header: 'Nilai perolehan', align: 'right', cell: (row) => rupiahShort(row.acquisitionValue) },
              { header: 'Nilai buku', align: 'right', cell: (row) => rupiahShort(row.bookValue) },
              {
                header: 'Status',
                cell: (row) => <Chip tone={ASSET_STATUS_TONE[row.status]}>{ASSET_STATUS_LABEL[row.status]}</Chip>,
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
                      question={`Hapus ${row.name}? Alat yang tertaut dilepas dari aset ini.`}
                      confirmLabel="Ya, hapus"
                      pendingLabel="Menghapus"
                      isAsking={askingId === row.id}
                      isPending={deleteAsset.isPending && deleteAsset.variables.id === row.id}
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
            page={assets.data.page}
            totalPages={assets.data.totalPages}
            totalItems={assets.data.totalItems}
            unit="aset"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
