import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreateVendor,
  useDeleteVendor,
  useUpdateVendor,
  useVendors,
} from '../../controllers/useProcurement'
import {
  EMPTY_VENDOR_FORM,
  VENDOR_CATEGORY_MAX_LENGTH,
  VENDOR_CATEGORY_SUGGESTIONS,
  VENDOR_CODE_MAX_LENGTH,
  VENDOR_CONTACT_MAX_LENGTH,
  VENDOR_NAME_MAX_LENGTH,
  VENDOR_RATING_LABEL,
  VENDOR_RATING_TONE,
  VENDOR_TAX_NUMBER_MAX_LENGTH,
  vendorFormFrom,
  type Vendor,
  type VendorFormValues,
  type VendorRating,
} from '../../models/procurement'
import { ApiError } from '../../shared/apiClient'
import { errorMessage } from '../../shared/errorMessage'
import { number } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { Chip, FilterChips, Pager, RowAction, SelectField } from '../components/ListTools'
import { ToolbarInput } from '../components/RecordControls'
import { PAGE_SIZE } from './lookup'
import { ConfirmAction, SmallButton } from './parts'

const CATEGORY_OPTIONS_ID = 'vendor-category-options'
const VENDOR_RATINGS: VendorRating[] = ['new', 'good', 'fair', 'poor']

type ActiveFilter = 'all' | 'active' | 'inactive'

const ACTIVE_FILTERS: { value: ActiveFilter; label: string }[] = [
  { value: 'all', label: 'Semua' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Nonaktif' },
]

// Formulir vendor untuk menambah (vendor null) dan mengubah. Status aktif
// tidak diubah di sini, tapi lewat tombol Nonaktifkan/Aktifkan di daftar.
function VendorFormCard({ vendor, onSaved, onCancel }: {
  vendor: Vendor | null
  onSaved: (vendor: Vendor) => void
  onCancel: () => void
}) {
  const [values, setValues] = useState<VendorFormValues>(() => (vendor ? vendorFormFrom(vendor) : EMPTY_VENDOR_FORM))
  const createVendor = useCreateVendor()
  const updateVendor = useUpdateVendor()
  const isEditing = vendor !== null
  const isSaving = createVendor.isPending || updateVendor.isPending
  const saveError = isEditing ? updateVendor.error : createVendor.error

  const update = (key: keyof VendorFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (vendor) {
      updateVendor.mutate({ id: vendor.id, values, active: vendor.active }, { onSuccess: onSaved })
      return
    }
    createVendor.mutate(values, { onSuccess: () => setValues(EMPTY_VENDOR_FORM) })
  }

  return (
    <Card
      title={vendor ? `Ubah vendor ${vendor.name}` : 'Vendor baru'}
      description={
        vendor
          ? 'Kode vendor harus tetap unik. Status aktif diatur lewat tombol Nonaktifkan atau Aktifkan di daftar.'
          : 'Vendor baru langsung aktif dan bisa dipilih saat membuat pesanan pembelian.'
      }
      action={
        isEditing ? (
          <Button variant="subtle" onClick={onCancel}>
            Batal ubah
          </Button>
        ) : undefined
      }
    >
      <form onSubmit={submit} className="space-y-5">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="vendor-code"
            label="Kode vendor"
            placeholder="VND-001"
            autoComplete="off"
            maxLength={VENDOR_CODE_MAX_LENGTH}
            hint={`Unik, maksimal ${VENDOR_CODE_MAX_LENGTH} karakter`}
            required
            value={values.code}
            onChange={update('code')}
          />
          <Field
            id="vendor-name"
            label="Nama vendor"
            autoComplete="off"
            maxLength={VENDOR_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="vendor-category"
            label="Kategori"
            placeholder="Pilih atau ketik kategori"
            list={CATEGORY_OPTIONS_ID}
            autoComplete="off"
            maxLength={VENDOR_CATEGORY_MAX_LENGTH}
            required
            value={values.category}
            onChange={update('category')}
          />
          <Field
            id="vendor-contact"
            label="Kontak (opsional)"
            autoComplete="off"
            maxLength={VENDOR_CONTACT_MAX_LENGTH}
            value={values.contact}
            onChange={update('contact')}
          />
          <Field
            id="vendor-tax-number"
            label="NPWP (opsional)"
            autoComplete="off"
            maxLength={VENDOR_TAX_NUMBER_MAX_LENGTH}
            value={values.taxNumber}
            onChange={update('taxNumber')}
          />
          <Field
            id="vendor-payment-term"
            label="Termin bayar (hari)"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            value={values.paymentTermDays}
            onChange={update('paymentTermDays')}
          />
          <SelectField id="vendor-rating" label="Penilaian" value={values.rating} onChange={update('rating')}>
            {VENDOR_RATINGS.map((rating) => (
              <option key={rating} value={rating}>
                {VENDOR_RATING_LABEL[rating]}
              </option>
            ))}
          </SelectField>
        </div>

        <datalist id={CATEGORY_OPTIONS_ID}>
          {VENDOR_CATEGORY_SUGGESTIONS.map((category) => (
            <option key={category} value={category} />
          ))}
        </datalist>

        <Button type="submit" isPending={isSaving} pendingLabel={isEditing ? 'Menyimpan perubahan' : 'Menyimpan vendor'}>
          {isEditing ? 'Simpan perubahan' : 'Simpan vendor'}
        </Button>

        {saveError ? <ErrorNote message={errorMessage(saveError)} /> : null}
        {!isEditing && createVendor.isSuccess ? (
          <SuccessNote message={`Vendor ${createVendor.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </Card>
  )
}

// Vendor yang sudah dipakai pesanan tidak bisa dihapus. Pesan backend
// dilengkapi jalan keluarnya: nonaktifkan saja.
function deleteErrorMessage(error: unknown): string {
  const message = errorMessage(error)
  if (error instanceof ApiError && error.code === 'vendor_in_use') {
    return `${message}. Nonaktifkan vendor ini supaya tidak bisa dipilih untuk pesanan baru.`
  }
  return message
}

type VendorForm = { vendor: Vendor | null } | null

export function VendorsSection() {
  const [search, setSearch] = useState('')
  const [activeFilter, setActiveFilter] = useState<ActiveFilter>('all')
  const [page, setPage] = useState(1)
  const [form, setForm] = useState<VendorForm>(null)
  const [confirmingId, setConfirmingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const toggleVendor = useUpdateVendor()
  const deleteVendor = useDeleteVendor()
  const vendors = useVendors({
    search: search.trim() === '' ? undefined : search.trim(),
    active: activeFilter === 'all' ? undefined : activeFilter === 'active',
    page,
    pageSize: PAGE_SIZE,
  })
  const editingId = form?.vendor?.id ?? null

  const startAction = () => {
    setNotice(null)
    toggleVendor.reset()
    deleteVendor.reset()
  }

  const toggleActive = (vendor: Vendor) => {
    startAction()
    setConfirmingId(null)
    toggleVendor.mutate(
      { id: vendor.id, values: vendorFormFrom(vendor), active: !vendor.active },
      {
        onSuccess: (saved) =>
          setNotice(
            saved.active
              ? `Vendor ${saved.name} aktif lagi dan bisa dipilih untuk pesanan baru.`
              : `Vendor ${saved.name} dinonaktifkan. Pesanan lamanya tetap ada, tapi vendor ini tidak bisa dipilih untuk pesanan baru.`,
          ),
      },
    )
  }

  const remove = (vendor: Vendor) =>
    deleteVendor.mutate(vendor.id, {
      onSuccess: () => {
        setNotice(`Vendor ${vendor.name} dihapus.`)
        if (editingId === vendor.id) {
          setForm(null)
        }
      },
      onSettled: () => setConfirmingId(null),
    })

  const toggleForm = () => {
    startAction()
    setForm((current) => (current === null ? { vendor: null } : null))
  }

  return (
    <div className="space-y-6">
      {form ? (
        <VendorFormCard
          key={form.vendor?.id ?? 'new'}
          vendor={form.vendor}
          onSaved={(saved) => {
            setForm(null)
            setNotice(`Perubahan vendor ${saved.name} tersimpan.`)
          }}
          onCancel={() => setForm(null)}
        />
      ) : null}

      <Card title="Daftar vendor" description={vendors.data ? `${number(vendors.data.totalItems)} vendor` : undefined}>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <ToolbarInput
              id="vendor-search"
              label="Cari vendor"
              type="search"
              placeholder="Cari nama atau kode vendor"
              maxLength={VENDOR_NAME_MAX_LENGTH}
              className="w-56"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
            />
            <FilterChips
              filters={ACTIVE_FILTERS}
              active={activeFilter}
              label="Saring status vendor"
              onChange={(value) => {
                setActiveFilter(value)
                setPage(1)
              }}
            />
          </div>
          <Button variant={form ? 'subtle' : 'primary'} onClick={toggleForm}>
            {form ? 'Tutup formulir' : 'Tambah vendor'}
          </Button>
        </div>

        <div className="mb-4 space-y-3 empty:hidden">
          {toggleVendor.isError ? <ErrorNote message={errorMessage(toggleVendor.error)} /> : null}
          {deleteVendor.isError ? <ErrorNote message={deleteErrorMessage(deleteVendor.error)} /> : null}
          {notice ? <SuccessNote message={notice} /> : null}
        </div>

        {vendors.isPending ? <Loading /> : null}
        {vendors.isError ? <LoadFailed onRetry={() => vendors.refetch()} /> : null}
        {vendors.data ? (
          <>
            <Table
              rows={vendors.data.items}
              emptyMessage={
                search.trim() === '' && activeFilter === 'all'
                  ? 'Belum ada vendor. Tambahkan lewat tombol Tambah vendor.'
                  : 'Tidak ada vendor yang cocok dengan penyaringan ini.'
              }
              columns={[
                { header: 'Kode', cell: (row) => row.code },
                { header: 'Nama', cell: (row) => row.name },
                { header: 'Kategori', cell: (row) => row.category },
                { header: 'Kontak', cell: (row) => row.contact || '-' },
                { header: 'Termin', align: 'right', cell: (row) => `${number(row.paymentTermDays)} hari` },
                {
                  header: 'Penilaian',
                  cell: (row) => <Chip label={VENDOR_RATING_LABEL[row.rating]} tone={VENDOR_RATING_TONE[row.rating]} />,
                },
                {
                  header: 'Status',
                  cell: (row) =>
                    row.active ? (
                      <Chip label="Aktif" tone="bg-green-100 text-green-800" />
                    ) : (
                      <Chip label="Nonaktif" tone="bg-slate-100 text-slate-700" />
                    ),
                },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => {
                    const isDeleting = deleteVendor.isPending && deleteVendor.variables === row.id
                    const deleteAction = (
                      <ConfirmAction
                        tone="danger"
                        label="Hapus"
                        prompt={`Hapus ${row.code}?`}
                        confirmLabel="Ya, hapus"
                        pendingLabel="Menghapus"
                        isAsking={confirmingId === row.id}
                        isPending={isDeleting}
                        onAsk={() => {
                          startAction()
                          setConfirmingId(row.id)
                        }}
                        onCancel={() => setConfirmingId(null)}
                        onConfirm={() => remove(row)}
                      />
                    )

                    if (confirmingId === row.id || isDeleting) {
                      return deleteAction
                    }

                    return (
                      <span className="inline-flex items-center justify-end gap-1.5">
                        <RowAction
                          label="Ubah"
                          isActive={editingId === row.id}
                          onClick={() => {
                            startAction()
                            setForm(editingId === row.id ? null : { vendor: row })
                          }}
                        />
                        <SmallButton
                          isPending={toggleVendor.isPending && toggleVendor.variables?.id === row.id}
                          pendingLabel="Menyimpan"
                          onClick={() => toggleActive(row)}
                        >
                          {row.active ? 'Nonaktifkan' : 'Aktifkan'}
                        </SmallButton>
                        {deleteAction}
                      </span>
                    )
                  },
                },
              ]}
            />
            <Pager
              page={vendors.data.page}
              totalPages={vendors.data.totalPages}
              totalItems={vendors.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>
    </div>
  )
}
