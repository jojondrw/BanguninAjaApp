import { type ChangeEvent, useState } from 'react'

import {
  useGoodsReceipts,
  usePurchaseOrderOptions,
  usePurchaseOrders,
  useVendorOptions,
  useWarehouseOptions,
} from '../../controllers/useProcurement'
import {
  ORDER_STATUS_LABEL,
  RECEIPT_CONDITION_LABEL,
  RECEIPT_CONDITION_TONE,
  RECEIPT_NUMBER_MAX_LENGTH,
} from '../../models/procurement'
import { number, shortDate } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { Chip, Pager, RowAction, SelectField } from '../components/ListTools'
import { FilterSelect, FormToggle, Toolbar, ToolbarInput } from '../components/RecordControls'
import { OPTION_LIMIT, PAGE_SIZE, nameOf, numberOf } from './lookup'
import { ReceiptDetailCard } from './ReceiptDetailCard'
import { ReceiveGoodsForm } from './ReceiveGoodsForm'

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
  const orderOptions = usePurchaseOrderOptions()
  const sentOrders = usePurchaseOrders({ status: 'sent', pageSize: OPTION_LIMIT })
  const partialOrders = usePurchaseOrders({ status: 'partially_received', pageSize: OPTION_LIMIT })
  const warehouses = useWarehouseOptions()
  const vendors = useVendorOptions()

  const orderItems = orderOptions.data?.items ?? []
  const warehouseItems = warehouses.data?.items ?? []
  const vendorItems = vendors.data?.items ?? []
  const receivedOrders = orderItems.filter((order) => order.status === 'partially_received' || order.status === 'completed')
  const receivableOrders = [...(sentOrders.data?.items ?? []), ...(partialOrders.data?.items ?? [])]
  const isFiltered = [search, orderId, warehouseId, dateFrom, dateTo].some((value) => value.trim() !== '')

  const changeFilter = (setter: (value: string) => void) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      setter(event.target.value)
      setPage(1)
    }

  return (
    <div className="space-y-6">
      {isReceiving ? (
        <Card
          title="Terima barang"
          description="Pilih pesanan yang barangnya datang. Hanya pesanan yang sudah dikirim ke vendor atau baru diterima sebagian."
        >
          {sentOrders.isPending || partialOrders.isPending ? <Loading label="Mengambil pesanan yang menunggu barang..." /> : null}
          {sentOrders.isError || partialOrders.isError ? (
            <LoadFailed
              onRetry={() => {
                void sentOrders.refetch()
                void partialOrders.refetch()
              }}
            />
          ) : null}
          {sentOrders.data && partialOrders.data ? (
            receivableOrders.length === 0 ? (
              <Empty message="Tidak ada pesanan yang menunggu barang. Kirim pesanan ke vendor dulu di tab Pesanan pembelian." />
            ) : (
              <div className="space-y-5">
                <div className="max-w-md">
                  <SelectField
                    id="receive-order"
                    label="Pesanan pembelian"
                    required
                    value={receivingOrderId}
                    onChange={(event) => setReceivingOrderId(event.target.value)}
                  >
                    <option value="">Pilih pesanan</option>
                    {receivableOrders.map((order) => (
                      <option key={order.id} value={order.id}>
                        {order.number}, {nameOf(vendorItems, order.vendorId)} ({ORDER_STATUS_LABEL[order.status]})
                      </option>
                    ))}
                  </SelectField>
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
            )
          ) : null}
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
          <FilterSelect
            id="receipt-filter-order"
            label="Saring menurut pesanan"
            value={orderId}
            onChange={changeFilter(setOrderId)}
          >
            <option value="">Semua pesanan</option>
            {receivedOrders.map((order) => (
              <option key={order.id} value={order.id}>
                {order.number}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect
            id="receipt-filter-warehouse"
            label="Saring menurut gudang"
            value={warehouseId}
            onChange={changeFilter(setWarehouseId)}
          >
            <option value="">Semua gudang</option>
            {warehouseItems.map((warehouse) => (
              <option key={warehouse.id} value={warehouse.id}>
                {warehouse.name}
              </option>
            ))}
          </FilterSelect>
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
                { header: 'Pesanan', cell: (row) => numberOf(orderItems, row.purchaseOrderId) },
                { header: 'Gudang', cell: (row) => nameOf(warehouseItems, row.warehouseId) },
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
