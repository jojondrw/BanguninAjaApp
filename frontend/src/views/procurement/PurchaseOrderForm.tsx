import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreatePurchaseOrder,
  useMaterials,
  useUnitsOfMeasure,
  useUpdatePurchaseOrder,
} from '../../controllers/useProcurement'
import {
  ORDER_NUMBER_MAX_LENGTH,
  emptyPurchaseOrderForm,
  type PurchaseOrderDetail,
  type PurchaseOrderFormValues,
  type Vendor,
} from '../../models/procurement'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { SelectField } from '../components/ListTools'
import { MaterialLinesEditor } from './MaterialLines'

// Formulir pesanan untuk membuat (editingId null) dan mengubah pesanan draf.
// Pesanan dari permintaan membawa purchaseRequestId, dan proyeknya dikunci
// karena backend menolak proyek yang berbeda dari permintaannya.
export function PurchaseOrderFormCard({
  editingId,
  initial,
  requestNumber,
  vendors,
  projects,
  isLoadingOptions,
  onSaved,
  onCancel,
}: {
  editingId: string | null
  initial: PurchaseOrderFormValues
  requestNumber: string | null
  vendors: Vendor[]
  projects: Project[]
  isLoadingOptions: boolean
  onSaved?: (order: PurchaseOrderDetail) => void
  onCancel?: () => void
}) {
  const [values, setValues] = useState<PurchaseOrderFormValues>(initial)
  const createOrder = useCreatePurchaseOrder()
  const updateOrder = useUpdatePurchaseOrder()
  const materials = useMaterials()
  const units = useUnitsOfMeasure()
  const materialItems = materials.data?.items ?? []
  const unitItems = units.data?.items ?? []
  const isEditing = editingId !== null
  const isFromRequest = values.purchaseRequestId !== ''
  const requestLabel = requestNumber ? ` ${requestNumber}` : ''
  const title = isEditing ? `Ubah pesanan ${initial.number}` : 'Pesanan pembelian baru'

  // Vendor nonaktif tidak bisa dipilih, tapi vendor pesanan yang sedang diubah
  // tetap ditampilkan supaya pilihannya tidak tiba-tiba kosong.
  const vendorChoices = vendors.filter((vendor) => vendor.active || vendor.id === values.vendorId)
  const isSaving = createOrder.isPending || updateOrder.isPending
  const saveError = isEditing ? updateOrder.error : createOrder.error

  const update = (key: 'number' | 'vendorId' | 'projectId' | 'date' | 'dueDate') =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (editingId) {
      updateOrder.mutate({ id: editingId, values }, { onSuccess: (order) => onSaved?.(order) })
      return
    }
    createOrder.mutate(values, {
      onSuccess: (order) => {
        setValues(emptyPurchaseOrderForm())
        onSaved?.(order)
      },
    })
  }

  const cancelAction = onCancel ? (
    <Button variant="subtle" onClick={onCancel}>
      Batal ubah
    </Button>
  ) : undefined

  if (materials.isPending || units.isPending || isLoadingOptions) {
    return (
      <Card title={title}>
        <Loading label="Mengambil vendor, proyek, material, dan satuan..." />
      </Card>
    )
  }

  if (materials.isError || units.isError) {
    return (
      <Card title={title}>
        <LoadFailed
          onRetry={() => {
            void materials.refetch()
            void units.refetch()
          }}
        />
      </Card>
    )
  }

  const missing =
    materialItems.length === 0
      ? 'Belum ada material di modul Persediaan. Pesanan pembelian butuh minimal satu material, jadi tambahkan materialnya dulu. Formulir ini langsung bisa dipakai begitu material tersedia.'
      : vendorChoices.length === 0
        ? 'Belum ada vendor aktif. Tambahkan vendor di tab Vendor.'
        : projects.length === 0
          ? 'Belum ada proyek. Buat proyek dulu di menu Proyek.'
          : null

  return (
    <Card
      title={title}
      description={
        isEditing
          ? 'Hanya pesanan berstatus Draf yang bisa diubah. Semua baris material diganti dengan isi formulir ini.'
          : 'Pesanan baru berstatus Draf. Nilai pesanan dihitung server dari jumlah dikali harga satuan.'
      }
      action={cancelAction}
    >
      {missing ? (
        <Empty message={missing} />
      ) : (
        <form onSubmit={submit} className="space-y-5">
          {isFromRequest ? (
            <p className="rounded-xl bg-slate-50 px-3.5 py-2.5 text-[13px] text-slate-700">
              {isEditing
                ? `Pesanan ini terhubung ke permintaan pembelian${requestLabel}, jadi proyeknya mengikuti permintaan.`
                : `Dibuat dari permintaan pembelian${requestLabel}. Proyek mengikuti permintaan, material dan jumlah boleh disesuaikan, harga satuan diisi dari harga terakhir material.`}
            </p>
          ) : null}

          <div className="grid gap-4 md:grid-cols-3">
            <Field
              id="po-number"
              label="Nomor PO"
              placeholder="PO-2026-001"
              autoComplete="off"
              maxLength={ORDER_NUMBER_MAX_LENGTH}
              required
              value={values.number}
              onChange={update('number')}
            />
            <SelectField
              id="po-vendor"
              label="Vendor"
              hint="Hanya vendor aktif"
              required
              value={values.vendorId}
              onChange={update('vendorId')}
            >
              <option value="">Pilih vendor</option>
              {vendorChoices.map((vendor) => (
                <option key={vendor.id} value={vendor.id}>
                  {vendor.name} ({vendor.code}){vendor.active ? '' : ', nonaktif'}
                </option>
              ))}
            </SelectField>
            <SelectField
              id="po-project"
              label="Proyek"
              hint={isFromRequest ? 'Mengikuti proyek permintaan' : undefined}
              required
              disabled={isFromRequest}
              value={values.projectId}
              onChange={update('projectId')}
            >
              <option value="">Pilih proyek</option>
              {projects.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name} ({project.code})
                </option>
              ))}
            </SelectField>
            <Field id="po-date" label="Tanggal pesanan" type="date" required value={values.date} onChange={update('date')} />
            <Field
              id="po-due-date"
              label="Jatuh tempo (opsional)"
              type="date"
              min={values.date || undefined}
              hint="Tidak boleh lebih awal dari tanggal pesanan"
              value={values.dueDate}
              onChange={update('dueDate')}
            />
          </div>

          <MaterialLinesEditor
            idPrefix="po"
            lines={values.items}
            materials={materialItems}
            units={unitItems}
            showPrice
            onChange={(items) => setValues((current) => ({ ...current, items }))}
          />

          <Button type="submit" isPending={isSaving} pendingLabel={isEditing ? 'Menyimpan perubahan' : 'Menyimpan pesanan'}>
            {isEditing ? 'Simpan perubahan' : 'Simpan pesanan'}
          </Button>

          {saveError ? <ErrorNote message={errorMessage(saveError)} /> : null}
          {!isEditing && createOrder.isSuccess ? (
            <SuccessNote
              message={`Pesanan ${createOrder.data.number} senilai ${rupiah(createOrder.data.value)} berhasil dibuat${
                createOrder.data.purchaseRequestId ? ` dari permintaan pembelian${requestLabel}` : ''
              }. Kirim ke vendor lewat rincian pesanan.`}
            />
          ) : null}
        </form>
      )}
    </Card>
  )
}
