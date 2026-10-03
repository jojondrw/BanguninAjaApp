import { useState } from 'react'

import {
  useGoodsReceipt,
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
  receiptNeedsMovementCheck,
  receiptStockLines,
  receiptStockRequest,
  receiptStockRequests,
  type ReceiptStockLine,
  type ReceiptStockState,
} from '../../models/procurement'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { Chip } from '../components/ListTools'
import { codeOf, nameOf } from './lookup'
import { ManualStockLine } from './ManualStockLine'
import { ConfirmAction, Facts } from './parts'

const STOCK_STATE_LABEL: Record<ReceiptStockState, string> = {
  posted: 'Masuk stok otomatis',
  recorded: 'Dicatat manual',
  pending: 'Belum masuk stok',
  convert: 'Satuan beda, catat manual',
  none: 'Tidak ada yang diterima',
}

const STOCK_STATE_TONE: Record<ReceiptStockState, string> = {
  posted: 'bg-green-100 text-green-800',
  recorded: 'bg-green-100 text-green-800',
  pending: 'bg-amber-100 text-amber-800',
  convert: 'bg-amber-100 text-amber-800',
  none: 'bg-slate-100 text-slate-700',
}

function countOf(lines: ReceiptStockLine[], state: ReceiptStockState): number {
  return lines.filter((line) => line.state === state).length
}

export function ReceiptDetailCard({ id, onClose }: { id: string; onClose?: () => void }) {
  const receipt = useGoodsReceipt(id)
  const order = usePurchaseOrder(receipt.data?.purchaseOrderId ?? null)
  const units = useUnitsOfMeasure()
  const warehouses = useWarehouseOptions()
  const movements = useReceiptStockMovements(receipt.data)
  const recordStock = useRecordReceiptStock()
  const recordManualStock = useRecordReceiptStock()
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
  const unitItems = units.data?.items ?? []
  const warehouseName = nameOf(warehouses.data?.items ?? [], detail.warehouseId)
  const needsCheck = receiptNeedsMovementCheck(detail)
  const isStockKnown = !needsCheck || movements.data !== undefined
  const lines = receiptStockLines(detail, movements.data?.items ?? [])
  const pending = receiptStockRequests(detail, lines)
  const convertLines = lines.filter((line) => line.state === 'convert')
  const posted = countOf(lines, 'posted')
  const recorded = countOf(lines, 'recorded')
  const manualRequest = recordManualStock.variables?.[0]
  const manualLine = lines.find((line) => line.materialId === manualRequest?.materialId)

  const recordPending = () =>
    recordStock.mutate(pending, { onSettled: () => setIsConfirming(false) })

  const recordConverted = (line: ReceiptStockLine) => (quantity: number, onSettled: () => void) =>
    recordManualStock.mutate([receiptStockRequest(detail, line, quantity)], { onSettled })

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
          { header: 'Material', cell: (row) => row.materialName },
          { header: 'Diterima baik', align: 'right', cell: (row) => number(row.acceptedQuantity) },
          { header: 'Ditolak', align: 'right', cell: (row) => number(row.rejectedQuantity) },
          { header: 'Satuan', cell: (row) => codeOf(unitItems, row.unitOfMeasureId) },
          {
            header: 'Stok gudang',
            cell: (row) =>
              isStockKnown || row.state === 'posted' || row.state === 'none' ? (
                <Chip label={STOCK_STATE_LABEL[row.state]} tone={STOCK_STATE_TONE[row.state]} />
              ) : (
                '...'
              ),
          },
        ]}
      />

      <div className="mt-5 space-y-3 border-t border-slate-100 pt-4">
        <h3 className="text-sm font-semibold text-slate-900">Stok gudang</h3>
        {posted > 0 ? (
          <p className="text-[13px] text-slate-600">
            Stok {warehouseName} bertambah otomatis untuk {number(posted)} material saat penerimaan disimpan, dengan
            referensi {detail.number}.
          </p>
        ) : null}
        {needsCheck && movements.isError ? (
          <LoadFailed onRetry={() => movements.refetch()} />
        ) : !isStockKnown ? (
          <Loading label="Memeriksa mutasi stok penerimaan ini..." />
        ) : (
          <>
            {recorded > 0 ? (
              <p className="text-[13px] text-slate-600">
                {number(recorded)} material sudah dicatat manual sebagai stok masuk di {warehouseName} dengan
                referensi {detail.number}.
              </p>
            ) : null}
            {pending.length > 0 ? (
              <div className="flex flex-wrap items-center justify-between gap-3">
                <p className="max-w-xl text-[13px] text-slate-600">
                  Penerimaan ini dicatat sebelum stok bertambah otomatis. {number(pending.length)} material belum
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
            ) : null}
            {convertLines.length > 0 ? (
              <div className="space-y-3">
                <p className="text-[13px] text-slate-600">
                  {number(convertLines.length)} baris belum masuk stok karena satuan pesanannya beda dengan satuan stok
                  material. Isi jumlahnya dalam satuan stok, lalu catat. Mutasi dicatat dengan referensi{' '}
                  {detail.number} dan tidak bisa dihapus.
                </p>
                <ul className="space-y-3">
                  {convertLines.map((line) => (
                    <ManualStockLine
                      key={line.receiptItemId}
                      receiptId={detail.id}
                      line={line}
                      orderUnit={codeOf(unitItems, line.unitOfMeasureId)}
                      stockUnit={codeOf(unitItems, line.stockUnitOfMeasureId)}
                      warehouseName={warehouseName}
                      isPending={recordManualStock.isPending && manualLine?.receiptItemId === line.receiptItemId}
                      error={
                        recordManualStock.isError && manualLine?.receiptItemId === line.receiptItemId
                          ? errorMessage(recordManualStock.error)
                          : null
                      }
                      onRecord={recordConverted(line)}
                    />
                  ))}
                </ul>
              </div>
            ) : null}
            {lines.every((line) => line.state === 'none') ? (
              <p className="text-[13px] text-slate-600">
                Tidak ada barang yang diterima baik, jadi tidak ada stok yang perlu dicatat.
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
        {recordManualStock.isSuccess && manualRequest && manualLine ? (
          <SuccessNote
            message={`Stok ${manualLine.materialName} bertambah ${number(manualRequest.quantity)} ${codeOf(unitItems, manualLine.stockUnitOfMeasureId)} di ${warehouseName}.`}
          />
        ) : null}
      </div>
    </Card>
  )
}
