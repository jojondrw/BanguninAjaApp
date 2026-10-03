import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useCreatePurchaseOrder, useUpdatePurchaseOrder } from '../../controllers/useProcurement'
import { activeVendorOptions, projectOptions } from '../../models/lookupApi'
import {
  ORDER_NUMBER_MAX_LENGTH,
  emptyPurchaseOrderForm,
  type PurchaseOrderDetail,
  type PurchaseOrderFormValues,
} from '../../models/procurement'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { Card } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { SearchSelect } from '../components/SearchSelect'
import { MaterialLinesEditor } from './MaterialLines'

// Formulir pesanan untuk membuat (editingId null) dan mengubah pesanan draf.
// Pesanan dari permintaan membawa purchaseRequestId, dan proyeknya dikunci
// karena backend menolak proyek yang berbeda dari permintaannya.
export function PurchaseOrderFormCard({
  editingId,
  initial,
  requestNumber,
  onSaved,
  onCancel,
}: {
  editingId: string | null
  initial: PurchaseOrderFormValues
  requestNumber: string | null
  onSaved?: (order: PurchaseOrderDetail) => void
  onCancel?: () => void
}) {
  const [values, setValues] = useState<PurchaseOrderFormValues>(initial)
  const createOrder = useCreatePurchaseOrder()
  const updateOrder = useUpdatePurchaseOrder()
  const isEditing = editingId !== null
  const isFromRequest = values.purchaseRequestId !== ''
  const requestLabel = requestNumber ? ` ${requestNumber}` : ''
  const title = isEditing ? `Ubah pesanan ${initial.number}` : 'Pesanan pembelian baru'

  const isSaving = createOrder.isPending || updateOrder.isPending
  const saveError = isEditing ? updateOrder.error : createOrder.error

  const update = (key: 'number' | 'date' | 'dueDate') =>
    (event: ChangeEvent<HTMLInputElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const choose = (key: 'vendorId' | 'projectId') => (value: string) =>
    setValues((current) => ({ ...current, [key]: value }))

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
          <SearchSelect
            {...activeVendorOptions}
            id="po-vendor"
            label="Vendor"
            placeholder="Cari nama atau kode vendor"
            hint="Hanya vendor aktif"
            required
            value={values.vendorId}
            onChange={choose('vendorId')}
          />
          <SearchSelect
            {...projectOptions}
            id="po-project"
            label="Proyek"
            placeholder="Cari nama atau kode proyek"
            hint={isFromRequest ? 'Mengikuti proyek permintaan' : undefined}
            required
            disabled={isFromRequest}
            value={values.projectId}
            onChange={choose('projectId')}
          />
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
    </Card>
  )
}
