import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  PROJECT_WAREHOUSE_LIMIT,
  useProjectWarehouses,
  usePurchaseOrder,
  useRecordGoodsReceipt,
  useWarehouseCount,
} from '../../controllers/useProcurement'
import type { MaterialOption } from '../../models/lookupApi'
import {
  ORDER_STATUS_LABEL,
  RECEIPT_CONDITIONS,
  RECEIPT_CONDITION_LABEL,
  RECEIPT_NOTE_MAX_LENGTH,
  RECEIPT_NUMBER_MAX_LENGTH,
  RECEIVABLE_ORDER_STATUSES,
  defaultReceiptLine,
  emptyGoodsReceiptForm,
  receiptLineQuantities,
  type GoodsReceiptCondition,
  type GoodsReceiptDetail,
  type GoodsReceiptFormValues,
  type PurchaseOrderDetail,
  type PurchaseOrderItem,
  type ReceiptLineValues,
} from '../../models/procurement'
import { errorMessage } from '../../shared/errorMessage'
import { number } from '../../shared/format'
import { Empty, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field } from '../components/Form'
import { SelectField } from '../components/ListTools'
import { warehouseOptions } from '../../models/lookupApi'
import { SearchSelect } from '../components/SearchSelect'
import { useMaterialLookups, useUnitCode } from './lookup'
import { TextAreaField } from './parts'

const ALL_WAREHOUSES = warehouseOptions()

// Selisih pembulatan dua desimal, sama dengan presisi jumlah di backend.
const QUANTITY_TOLERANCE = 0.005

export interface ReceiveResult {
  receipt: GoodsReceiptDetail
  order: PurchaseOrderDetail | null
}

// Isi formulir "Terima barang" untuk satu pesanan. Bingkai kartunya diatur
// pemanggil, karena formulir ini dibuka dari rincian pesanan maupun dari tab
// Penerimaan barang.
export function ReceiveGoodsForm({ orderId, onSaved }: { orderId: string; onSaved: (result: ReceiveResult) => void }) {
  const order = usePurchaseOrder(orderId)
  const warehouseCount = useWarehouseCount()
  const projectWarehouses = useProjectWarehouses(order.data?.projectId)
  const materials = useMaterialLookups(order.data?.items.map((item) => item.materialId) ?? [])
  const recordReceipt = useRecordGoodsReceipt()
  const [values, setValues] = useState<GoodsReceiptFormValues>(emptyGoodsReceiptForm)
  const [lines, setLines] = useState<Record<string, ReceiptLineValues>>({})
  const [hasTriedSubmit, setHasTriedSubmit] = useState(false)

  if (order.isPending || warehouseCount.isPending || projectWarehouses.isPending) {
    return <Loading label="Mengambil pesanan dan gudang..." />
  }

  if (order.isError || warehouseCount.isError || projectWarehouses.isError) {
    return (
      <LoadFailed
        onRetry={() => {
          void order.refetch()
          void warehouseCount.refetch()
          void projectWarehouses.refetch()
        }}
      />
    )
  }

  const detail = order.data

  if (!RECEIVABLE_ORDER_STATUSES.includes(detail.status)) {
    return (
      <Empty
        message={`Pesanan ${detail.number} berstatus ${ORDER_STATUS_LABEL[detail.status]}. Barang hanya bisa diterima untuk pesanan yang sudah dikirim ke vendor atau baru diterima sebagian.`}
      />
    )
  }

  if (warehouseCount.data === 0) {
    return <Empty message="Belum ada gudang. Tambahkan gudang di halaman Persediaan dulu, lalu catat penerimaan ini." />
  }

  // Gudang milik proyek pesanan jadi petunjuk. Kalau proyeknya punya tepat
  // satu gudang, gudang itu langsung terpilih.
  const ownWarehouses = projectWarehouses.data.items
  const ownTotal = projectWarehouses.data.totalItems
  const warehouseId = values.warehouseId || (ownTotal === 1 ? ownWarehouses[0].id : '')
  const warehouseHint =
    ownTotal === 0
      ? undefined
      : `Gudang proyek ini: ${ownWarehouses.map((item) => item.name).join(', ')}${
          ownTotal > PROJECT_WAREHOUSE_LIMIT ? `, dan ${number(ownTotal - PROJECT_WAREHOUSE_LIMIT)} lainnya` : ''
        }`

  const resolved = detail.items.map((item) => {
    const line = lines[item.id] ?? defaultReceiptLine(item)
    return { item, line, quantities: receiptLineQuantities(line) }
  })
  const overIndex = resolved.findIndex(
    ({ item, quantities }) => quantities.accepted > item.remainingQuantity + QUANTITY_TOLERANCE,
  )
  const hasQuantity = resolved.some(({ quantities }) => quantities.accepted + quantities.rejected > 0)
  const problem =
    overIndex >= 0
      ? `Jumlah diterima baris ${overIndex + 1} melebihi sisa pesanan (${number(resolved[overIndex].item.remainingQuantity)}).`
      : !hasQuantity
        ? 'Isi jumlah diterima atau ditolak minimal di satu baris.'
        : null

  const update = (key: keyof GoodsReceiptFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const updateLine = (itemId: string, current: ReceiptLineValues, key: keyof ReceiptLineValues) =>
    (event: ChangeEvent<HTMLInputElement>) =>
      setLines((all) => ({ ...all, [itemId]: { ...current, [key]: event.target.value } }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (problem) {
      setHasTriedSubmit(true)
      return
    }
    recordReceipt.mutate(
      {
        purchaseOrderId: detail.id,
        values: { ...values, warehouseId },
        lines: resolved.map(({ item, line }) => ({ purchaseOrderItemId: item.id, values: line })),
      },
      { onSuccess: onSaved },
    )
  }

  return (
    <form onSubmit={submit} className="space-y-5">
      <p className="text-sm text-slate-600">
        Isi jumlah yang benar-benar datang. Setelah disimpan, status pesanan berubah sendiri menjadi Diterima sebagian
        atau Selesai, stok gudang penerima bertambah otomatis sebesar jumlah yang diterima baik, dan penerimaan tidak
        bisa diubah maupun dihapus. Barang yang ditolak tidak masuk stok.
      </p>

      <div className="grid gap-4 md:grid-cols-4">
        <Field
          id={`receipt-${detail.id}-number`}
          label="Nomor penerimaan"
          placeholder="BPB-2026-001"
          autoComplete="off"
          maxLength={RECEIPT_NUMBER_MAX_LENGTH}
          required
          value={values.number}
          onChange={update('number')}
        />
        <SearchSelect
          {...ALL_WAREHOUSES}
          id={`receipt-${detail.id}-warehouse`}
          label="Gudang penerima"
          placeholder="Cari nama atau kode gudang"
          hint={warehouseHint}
          required
          value={warehouseId}
          onChange={(value) => setValues((current) => ({ ...current, warehouseId: value }))}
        />
        <Field
          id={`receipt-${detail.id}-date`}
          label="Tanggal diterima"
          type="date"
          required
          value={values.date}
          onChange={update('date')}
        />
        <SelectField
          id={`receipt-${detail.id}-condition`}
          label="Kondisi barang"
          required
          value={values.condition}
          onChange={(event) =>
            setValues((current) => ({ ...current, condition: event.target.value as GoodsReceiptCondition }))
          }
        >
          {RECEIPT_CONDITIONS.map((condition) => (
            <option key={condition} value={condition}>
              {RECEIPT_CONDITION_LABEL[condition]}
            </option>
          ))}
        </SelectField>
      </div>

      <div className="space-y-3">
        {resolved.map(({ item, line }, index) => (
          <ReceiveLineFields
            key={item.id}
            id={`receipt-${detail.id}-line-${item.id}`}
            index={index}
            item={item}
            line={line}
            material={materials.byId.get(item.materialId)}
            onChange={(key) => updateLine(item.id, line, key)}
          />
        ))}
      </div>

      <TextAreaField
        id={`receipt-${detail.id}-note`}
        label="Catatan (opsional)"
        placeholder="Misalnya nomor surat jalan atau keterangan barang rusak"
        maxLength={RECEIPT_NOTE_MAX_LENGTH}
        value={values.note}
        onChange={update('note')}
      />

      {hasTriedSubmit && problem ? <ErrorNote message={problem} /> : null}

      <Button type="submit" isPending={recordReceipt.isPending} pendingLabel="Menyimpan penerimaan">
        Simpan penerimaan
      </Button>

      {recordReceipt.isError ? <ErrorNote message={errorMessage(recordReceipt.error)} /> : null}
    </form>
  )
}

function withUnit(label: string, unit: string): string {
  return unit === '' ? label : `${label} (${unit})`
}

function ReceiveLineFields({ id, index, item, line, material, onChange }: {
  id: string
  index: number
  item: PurchaseOrderItem
  line: ReceiptLineValues
  material: MaterialOption | undefined
  onChange: (key: keyof ReceiptLineValues) => (event: ChangeEvent<HTMLInputElement>) => void
}) {
  const unit = useUnitCode(item.unitOfMeasureId)
  const isComplete = item.remainingQuantity <= 0
  const stockUnitId = material?.unitOfMeasureId ?? item.unitOfMeasureId
  const stockUnit = useUnitCode(stockUnitId)

  return (
    <fieldset className="rounded-xl bg-white p-4 shadow-hairline">
      <legend className="float-left mb-3 w-full text-xs font-medium text-slate-500">
        Baris {index + 1}: {material?.label ?? '…'}
      </legend>
      <div className="clear-both grid gap-4 md:grid-cols-[1.4fr_1fr_1fr]">
        <dl className="grid grid-cols-3 gap-2 text-[13px]">
          <div>
            <dt className="text-xs text-slate-500">Dipesan</dt>
            <dd className="mt-0.5 text-slate-900 tabular-nums">
              {number(item.quantity)} {unit}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-slate-500">Sudah diterima</dt>
            <dd className="mt-0.5 text-slate-900 tabular-nums">
              {number(item.receivedQuantity)} {unit}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-slate-500">Sisa</dt>
            <dd className="mt-0.5 font-medium text-slate-900 tabular-nums">
              {number(item.remainingQuantity)} {unit}
            </dd>
          </div>
        </dl>
        {isComplete ? (
          <p className="self-center text-[13px] text-slate-600 md:col-span-2">
            Baris ini sudah diterima lengkap.
          </p>
        ) : (
          <>
            <Field
              id={`${id}-accepted`}
              label={withUnit('Diterima baik', unit)}
              type="number"
              inputMode="decimal"
              min={0}
              max={item.remainingQuantity}
              step={0.01}
              placeholder="0"
              hint="Mengurangi sisa pesanan. Kosongkan kalau belum datang."
              value={line.accepted}
              onChange={onChange('accepted')}
            />
            <Field
              id={`${id}-rejected`}
              label={withUnit('Ditolak', unit)}
              type="number"
              inputMode="decimal"
              min={0}
              step={0.01}
              placeholder="0"
              hint="Barang rusak yang dikembalikan, tidak mengurangi sisa."
              value={line.rejected}
              onChange={onChange('rejected')}
            />
          </>
        )}
      </div>
      {!isComplete && stockUnitId !== item.unitOfMeasureId ? (
        <p className="mt-3 text-xs text-amber-800">
          Satuan stok material ini {stockUnit || '-'}, beda dengan satuan pesanan ({unit || '-'}). Baris ini tidak
          menambah stok otomatis: catat manual dari rincian penerimaan setelah jumlahnya dikonversi.
        </p>
      ) : null}
    </fieldset>
  )
}
