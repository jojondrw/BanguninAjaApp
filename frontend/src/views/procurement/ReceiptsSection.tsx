import { type ChangeEvent, useState } from 'react'

import { useGoodsReceipts } from '../../controllers/useProcurement'
import { purchaseOrderOptions, receivableOrderOptions, warehouseOptions } from '../../models/lookupApi'
import {
  ORDER_STATUS_LABEL,
  RECEIPT_CONDITION_LABEL,
  RECEIPT_CONDITION_TONE,
  RECEIPT_NUMBER_MAX_LENGTH,
} from '../../models/procurement'
import { number, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { Chip, Pager, RowAction } from '../components/ListTools'
import { FormToggle, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { PAGE_SIZE } from './lookup'
import { ReceiptDetailCard } from './ReceiptDetailCard'
import { ReceiveGoodsForm } from './ReceiveGoodsForm'

const ALL_ORDERS = purchaseOrderOptions()
const ALL_WAREHOUSES = warehouseOptions()

function optional(value: string): string | undefined {
  const trimmed = value.trim()
  return trimmed === '' ? undefined : trimmed
}

export function ReceiptsSection() {
  const [search, setSearch] = useState('')
  const [orderId, setOrderId] = useState('')
  const [warehouseId, setWarehouseId] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [page, setPage] = useState(1)
  const [isReceiving, setIsReceiving] = useState(false)
  const [receivingOrderId, setReceivingOrderId] = useState('')
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const receipts = useGoodsReceipts({
    search: optional(search),
    purchaseOrderId: optional(orderId),
    warehouseId: optional(warehouseId),
    dateFrom: optional(dateFrom),
    dateTo: optional(dateTo),
    page,
    pageSize: PAGE_SIZE,
  })
  const isFiltered = [search, orderId, warehouseId, dateFrom, dateTo].some((value) => value.trim() !== '')

  const changeFilter = (setter: (value: string) => void) =>
    (event: ChangeEvent<HTMLInputElement>) => {
      setter(event.target.value)
      setPage(1)
    }

  const pickFilter = (setter: (value: string) => void) => (value: string) => {
    setter(value)
    setPage(1)
  }

  return (
    <div className="space-y-6">
      {isReceiving ? (
        <Card
          title="Terima barang"
          description="Pilih pesanan yang barangnya datang. Hanya pesanan yang sudah dikirim ke vendor atau baru diterima sebagian."
        >
              <div className="space-y-5">
                <div className="max-w-md">
                  <SearchSelect
                    {...receivableOrderOptions}
                    id="receive-order"
                    label="Pesanan pembelian"
                    placeholder="Cari nomor PO"
                    hint="Pesanan yang belum dikirim ke vendor tidak muncul di sini."
                    required
                    value={receivingOrderId}
                    onChange={setReceivingOrderId}
                  />
                </div>
                {receivingOrderId ? (
                  <ReceiveGoodsForm
                    key={receivingOrderId}
                    orderId={receivingOrderId}
                    onSaved={(result) => {
                      setIsReceiving(false)
                      setReceivingOrderId('')
                      setSelectedId(result.receipt.id)
                      setNotice(
                        `Penerimaan ${result.receipt.number} tersimpan${
                          result.order
                            ? `, pesanan ${result.order.number} sekarang ${ORDER_STATUS_LABEL[result.order.status]}`
                            : ''
                        }. Stok gudang belum bertambah, catat stok masuk dari rincian penerimaan.`,
                      )
                    }}
                  />
                ) : null}
              </div>
        </Card>
      ) : null}

      <Card
        title="Penerimaan barang"
        description="Bukti barang datang dari vendor. Menyimpan penerimaan mengubah status pesanan menjadi Diterima sebagian atau Selesai. Stok gudang dicatat terpisah dari rincian penerimaan."
      >
        <Toolbar>
          <ToolbarInput
            id="receipt-search"
            label="Cari nomor penerimaan"
            type="search"
            placeholder="Cari nomor penerimaan"
            maxLength={RECEIPT_NUMBER_MAX_LENGTH}
            className="w-52"
            value={search}
            onChange={changeFilter(setSearch)}
          />
          <SearchSelect
            {...ALL_ORDERS}
            id="receipt-filter-order"
            label="Saring menurut pesanan"
            compact
            allowEmpty
            emptyLabel="Semua pesanan"
            className="w-48"
            value={orderId}
            onChange={pickFilter(setOrderId)}
          />
          <SearchSelect
            {...ALL_WAREHOUSES}
            id="receipt-filter-warehouse"
            label="Saring menurut gudang"
            compact
            allowEmpty
            emptyLabel="Semua gudang"
            className="w-48"
            value={warehouseId}
            onChange={pickFilter(setWarehouseId)}
          />
          <ToolbarInput
            id="receipt-date-from"
            label="Dari"
            showLabel
            type="date"
            max={dateTo || undefined}
            value={dateFrom}
            onChange={changeFilter(setDateFrom)}
          />
          <ToolbarInput
            id="receipt-date-to"
            label="Sampai"
            showLabel
            type="date"
            min={dateFrom || undefined}
            value={dateTo}
            onChange={changeFilter(setDateTo)}
          />
          <FormToggle
            isOpen={isReceiving}
            openLabel="Terima barang"
            onToggle={() => {
              setNotice(null)
              setReceivingOrderId('')
              setIsReceiving((open) => !open)
            }}
          />
        </Toolbar>

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {receipts.isPending ? <Loading /> : null}
        {receipts.isError ? <LoadFailed onRetry={() => receipts.refetch()} /> : null}
        {receipts.data ? (
          <>
            <Table
              rows={receipts.data.items}
              emptyMessage={
                isFiltered
                  ? 'Tidak ada penerimaan yang cocok dengan penyaringan ini.'
                  : 'Belum ada penerimaan barang. Catat lewat tombol Terima barang saat kiriman vendor datang.'
              }
              columns={[
                { header: 'Nomor', cell: (row) => row.number },
                { header: 'Pesanan', cell: (row) => row.purchaseOrderNumber ?? '-' },
                { header: 'Gudang', cell: (row) => row.warehouseName ?? '-' },
                { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                {
                  header: 'Kondisi',
                  cell: (row) => (
                    <Chip label={RECEIPT_CONDITION_LABEL[row.condition]} tone={RECEIPT_CONDITION_TONE[row.condition]} />
                  ),
                },
                { header: 'Baris', align: 'right', cell: (row) => number(row.itemCount) },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Rincian"
                      isActive={row.id === selectedId}
                      onClick={() => setSelectedId((current) => (current === row.id ? null : row.id))}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={receipts.data.page}
              totalPages={receipts.data.totalPages}
              totalItems={receipts.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? <ReceiptDetailCard key={selectedId} id={selectedId} onClose={() => setSelectedId(null)} /> : null}
    </div>
  )
}
