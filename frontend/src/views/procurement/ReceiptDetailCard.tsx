import { useState } from 'react'

import {
  useGoodsReceipt,
  useMaterials,
  usePurchaseOrder,
  useReceiptStockMovements,
  useRecordReceiptStock,
  useUnitsOfMeasure,
  useWarehouseOptions,
} from '../../controllers/useProcurement'
import {
  ORDER_STATUS_LABEL,
  ORDER_STATUS_TONE,
  RECEIPT_CONDITION_LABEL,
  RECEIPT_CONDITION_TONE,
  receiptStockLines,
  receiptStockRequests,
  type ReceiptStockState,
} from '../../models/procurement'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { Chip } from '../components/ListTools'
import { codeOf, nameOf } from './lookup'
import { ConfirmAction, Facts } from './parts'

const STOCK_STATE_LABEL: Record<ReceiptStockState, string> = {
  recorded: 'Sudah masuk stok',
  pending: 'Belum masuk stok',
  unit_mismatch: 'Satuan beda, catat manual',
  none: 'Tidak ada yang diterima',
}

const STOCK_STATE_TONE: Record<ReceiptStockState, string> = {
  recorded: 'bg-green-100 text-green-800',
  pending: 'bg-amber-100 text-amber-800',
  unit_mismatch: 'bg-slate-100 text-slate-700',
  none: 'bg-slate-100 text-slate-700',
}

export function ReceiptDetailCard({ id, onClose }: { id: string; onClose?: () => void }) {
  const receipt = useGoodsReceipt(id)
  const order = usePurchaseOrder(receipt.data?.purchaseOrderId ?? null)
  const materials = useMaterials()
  const units = useUnitsOfMeasure()
  const warehouses = useWarehouseOptions()
  const movements = useReceiptStockMovements(receipt.data)
  const recordStock = useRecordReceiptStock()
  const [isConfirming, setIsConfirming] = useState(false)

  const closeAction = onClose ? (
    <Button variant="subtle" onClick={onClose}>
      Tutup rincian
    </Button>
  ) : undefined

  if (receipt.isPending) {
    return (
      <Card title="Rincian penerimaan" action={closeAction}>
        <Loading />
      </Card>
    )
  }

  if (receipt.isError) {
    return (
      <Card title="Rincian penerimaan" action={closeAction}>
        <LoadFailed onRetry={() => receipt.refetch()} />
      </Card>
    )
  }

  const detail = receipt.data
  const materialItems = materials.data?.items ?? []
  const unitItems = units.data?.items ?? []
  const warehouseName = nameOf(warehouses.data?.items ?? [], detail.warehouseId)
  const isStockKnown = order.data !== undefined && movements.data !== undefined && materials.data !== undefined
  const lines = receiptStockLines(detail, order.data?.items ?? [], materialItems, movements.data?.items ?? [])
  const pending = receiptStockRequests(detail, lines)
  const mismatched = lines.filter((line) => line.state === 'unit_mismatch').length
  const recorded = lines.filter((line) => line.state === 'recorded').length

  const recordPending = () =>
    recordStock.mutate(pending, { onSettled: () => setIsConfirming(false) })

  return (
    <Card
      title={`Penerimaan ${detail.number}`}
      description={`Pesanan ${order.data?.number ?? '-'}, diterima ${shortDate(detail.date)}`}
      action={closeAction}
    >
      <Facts
        items={[
          { label: 'Gudang penerima', value: warehouseName },
          {
            label: 'Kondisi barang',
            value: (
              <Chip label={RECEIPT_CONDITION_LABEL[detail.condition]} tone={RECEIPT_CONDITION_TONE[detail.condition]} />
            ),
          },
          {
            label: 'Status pesanan sekarang',
            value: order.data ? (
              <Chip label={ORDER_STATUS_LABEL[order.data.status]} tone={ORDER_STATUS_TONE[order.data.status]} />
            ) : (
              '-'
            ),
          },
          { label: 'Catatan', value: detail.note || '-' },
        ]}
      />

      <Table
        rows={lines}
        emptyMessage="Penerimaan ini tidak punya baris barang."
        columns={[
          { header: 'Material', cell: (row) => nameOf(materialItems, row.materialId) },
          { header: 'Diterima baik', align: 'right', cell: (row) => number(row.acceptedQuantity) },
          { header: 'Ditolak', align: 'right', cell: (row) => number(row.rejectedQuantity) },
          { header: 'Satuan', cell: (row) => codeOf(unitItems, row.unitOfMeasureId) },
          {
            header: 'Stok gudang',
            cell: (row) =>
              isStockKnown ? <Chip label={STOCK_STATE_LABEL[row.state]} tone={STOCK_STATE_TONE[row.state]} /> : '...',
          },
        ]}
      />

      <div className="mt-5 space-y-3 border-t border-slate-100 pt-4">
        <h3 className="text-sm font-semibold text-slate-900">Stok gudang</h3>
        {movements.isError || order.isError ? (
          <LoadFailed
            onRetry={() => {
              void movements.refetch()
              void order.refetch()
            }}
          />
        ) : !isStockKnown ? (
          <Loading label="Memeriksa mutasi stok penerimaan ini..." />
        ) : (
          <>
            {pending.length > 0 ? (
              <div className="flex flex-wrap items-center justify-between gap-3">
                <p className="max-w-xl text-[13px] text-slate-600">
                  Penerimaan tidak menambah stok gudang dengan sendirinya. {number(pending.length)} material belum
                  tercatat sebagai stok masuk di {warehouseName}. Mutasi dicatat dengan referensi {detail.number} dan
                  tidak bisa dihapus.
                </p>
                <ConfirmAction
                  size="md"
                  label="Catat stok masuk"
                  prompt={`Tambah stok ${number(pending.length)} material ke ${warehouseName}?`}
                  confirmLabel="Ya, catat stok masuk"
                  pendingLabel="Mencatat stok masuk"
                  isAsking={isConfirming}
                  isPending={recordStock.isPending}
                  onAsk={() => setIsConfirming(true)}
                  onCancel={() => setIsConfirming(false)}
                  onConfirm={recordPending}
                />
              </div>
            ) : recorded > 0 ? (
              <p className="text-[13px] text-slate-600">
                Semua barang yang diterima baik sudah tercatat sebagai stok masuk di {warehouseName} dengan referensi{' '}
                {detail.number}.
              </p>
            ) : mismatched === 0 ? (
              <p className="text-[13px] text-slate-600">
                Tidak ada barang yang diterima baik, jadi tidak ada stok yang perlu dicatat.
              </p>
            ) : null}
            {mismatched > 0 ? (
              <p className="text-[13px] text-slate-600">
                {number(mismatched)} baris memakai satuan yang berbeda dari satuan stok materialnya, jadi tidak dicatat
                otomatis. Catat manual di halaman Persediaan setelah dikonversi, dengan referensi {detail.number}.
              </p>
            ) : null}
          </>
        )}
        {recordStock.isError ? <ErrorNote message={errorMessage(recordStock.error)} /> : null}
        {recordStock.isSuccess ? (
          <SuccessNote
            message={`${number(recordStock.data)} mutasi stok masuk dicatat. Stok di halaman Persediaan ikut diperbarui.`}
          />
        ) : null}
      </div>
    </Card>
  )
}
