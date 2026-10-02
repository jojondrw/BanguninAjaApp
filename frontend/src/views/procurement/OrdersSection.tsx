import { useState } from 'react'

import { useProjects } from '../../controllers/useErp'
import {
  useDeletePurchaseOrder,
  useGoodsReceipts,
  useMaterials,
  usePurchaseOrder,
  usePurchaseOrders,
  usePurchaseRequest,
  useUnitsOfMeasure,
  useUpdatePurchaseOrderStatus,
  useVendorOptions,
  useWarehouseOptions,
} from '../../controllers/useProcurement'
import {
  ORDER_NUMBER_MAX_LENGTH,
  ORDER_STATUS_LABEL,
  ORDER_STATUS_TONE,
  RECEIPT_CONDITION_LABEL,
  RECEIPT_CONDITION_TONE,
  emptyPurchaseOrderForm,
  purchaseOrderFormFrom,
  type PurchaseOrderStatus,
  type Vendor,
} from '../../models/procurement'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiah, rupiahShort, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { Chip, FilterChips, Pager, RowAction } from '../components/ListTools'
import { ToolbarInput } from '../components/RecordControls'
import { OPTION_LIMIT, PAGE_SIZE, codeOf, nameOf, type OrderPrefill } from './lookup'
import { ConfirmAction, Facts } from './parts'
import { PurchaseOrderFormCard } from './PurchaseOrderForm'
import { ReceiptDetailCard } from './ReceiptDetailCard'
import { ReceiveGoodsForm } from './ReceiveGoodsForm'

const ORDER_FILTERS: { value: PurchaseOrderStatus | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'draft', label: ORDER_STATUS_LABEL.draft },
  { value: 'sent', label: ORDER_STATUS_LABEL.sent },
  { value: 'partially_received', label: ORDER_STATUS_LABEL.partially_received },
  { value: 'completed', label: ORDER_STATUS_LABEL.completed },
  { value: 'cancelled', label: ORDER_STATUS_LABEL.cancelled },
]

const STATUS_HINT: Record<PurchaseOrderStatus, string> = {
  draft: 'Draf masih bisa diubah atau dihapus. Setelah dikirim ke vendor, isinya terkunci.',
  sent: 'Menunggu barang datang. Catat penerimaan saat barang tiba.',
  partially_received: 'Sebagian barang sudah diterima. Catat penerimaan berikutnya saat sisanya datang.',
  completed: 'Semua barang sudah diterima.',
  cancelled: 'Pesanan dibatalkan dan tidak bisa diproses lagi.',
}

type DetailMode = 'view' | 'edit' | 'receive'
type OrderConfirm = 'cancel' | 'delete'

function PurchaseOrderDetailCard({ id, vendors, projects, isLoadingOptions, onDeleted }: {
  id: string
  vendors: Vendor[]
  projects: Project[]
  isLoadingOptions: boolean
  onDeleted: (number: string) => void
}) {
  const [mode, setMode] = useState<DetailMode>('view')
  const [confirming, setConfirming] = useState<OrderConfirm | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [receiptId, setReceiptId] = useState<string | null>(null)
  const order = usePurchaseOrder(id)
  const sourceRequest = usePurchaseRequest(order.data?.purchaseRequestId ?? null)
  const receipts = useGoodsReceipts({ purchaseOrderId: id, pageSize: OPTION_LIMIT })
  const materials = useMaterials()
  const units = useUnitsOfMeasure()
  const warehouses = useWarehouseOptions()
  const updateStatus = useUpdatePurchaseOrderStatus()
  const deleteOrder = useDeletePurchaseOrder()
  const materialItems = materials.data?.items ?? []
  const unitItems = units.data?.items ?? []
  const warehouseItems = warehouses.data?.items ?? []

  if (order.isPending) {
    return (
      <Card title="Rincian pesanan">
        <Loading />
      </Card>
    )
  }

  if (order.isError) {
    return (
      <Card title="Rincian pesanan">
        <LoadFailed onRetry={() => order.refetch()} />
      </Card>
    )
  }

  const detail = order.data
  const requestNumber = sourceRequest.data?.number ?? null

  if (mode === 'edit') {
    return (
      <PurchaseOrderFormCard
        editingId={detail.id}
        initial={purchaseOrderFormFrom(detail)}
        requestNumber={requestNumber}
        vendors={vendors}
        projects={projects}
        isLoadingOptions={isLoadingOptions}
        onSaved={(saved) => {
          setMode('view')
          setNotice(`Perubahan pesanan ${saved.number} tersimpan, nilainya sekarang ${rupiah(saved.value)}.`)
        }}
        onCancel={() => setMode('view')}
      />
    )
  }

  const startAction = () => {
    setNotice(null)
    updateStatus.reset()
    deleteOrder.reset()
  }

  const send = () => {
    startAction()
    updateStatus.mutate(
      { id: detail.id, status: 'sent' },
      { onSuccess: (saved) => setNotice(`Pesanan ${saved.number} dikirim ke vendor. Catat penerimaan saat barang datang.`) },
    )
  }

  const cancel = () =>
    updateStatus.mutate(
      { id: detail.id, status: 'cancelled' },
      {
        onSuccess: (saved) => setNotice(`Pesanan ${saved.number} dibatalkan.`),
        onSettled: () => setConfirming(null),
      },
    )

  const remove = () =>
    deleteOrder.mutate(detail.id, {
      onSuccess: () => onDeleted(detail.number),
      onSettled: () => setConfirming(null),
    })

  const ask = (action: OrderConfirm) => () => {
    startAction()
    setConfirming(action)
  }

  const cancelConfirm = (
    <ConfirmAction
      size="md"
      tone="danger"
      label="Batalkan pesanan"
      prompt="Batalkan pesanan ini? Statusnya tidak bisa dikembalikan."
      confirmLabel="Ya, batalkan"
      pendingLabel="Membatalkan"
      isAsking={confirming === 'cancel'}
      isPending={updateStatus.isPending && updateStatus.variables?.status === 'cancelled'}
      onAsk={ask('cancel')}
      onCancel={() => setConfirming(null)}
      onConfirm={cancel}
    />
  )

  const deleteConfirm = (
    <ConfirmAction
      size="md"
      tone="danger"
      label="Hapus pesanan"
      prompt="Hapus pesanan draf ini beserta semua barisnya?"
      confirmLabel="Ya, hapus"
      pendingLabel="Menghapus"
      isAsking={confirming === 'delete'}
      isPending={deleteOrder.isPending}
      onAsk={ask('delete')}
      onCancel={() => setConfirming(null)}
      onConfirm={remove}
    />
  )

  const isCancelling = confirming === 'cancel' || (updateStatus.isPending && updateStatus.variables?.status === 'cancelled')
  const isDeleting = confirming === 'delete' || deleteOrder.isPending
  const canReceive = detail.status === 'sent' || detail.status === 'partially_received'

  let actions = null
  if (isCancelling) {
    actions = cancelConfirm
  } else if (isDeleting) {
    actions = deleteConfirm
  } else if (detail.status === 'draft') {
    actions = (
      <>
        <Button variant="subtle" onClick={() => setMode('edit')}>
          Ubah pesanan
        </Button>
        <Button isPending={updateStatus.isPending} pendingLabel="Mengirim" onClick={send}>
          Kirim ke vendor
        </Button>
        {cancelConfirm}
        {deleteConfirm}
      </>
    )
  } else if (canReceive) {
    actions = (
      <>
        <Button
          variant={mode === 'receive' ? 'subtle' : 'primary'}
          onClick={() => {
            startAction()
            setMode((current) => (current === 'receive' ? 'view' : 'receive'))
          }}
        >
          {mode === 'receive' ? 'Tutup formulir penerimaan' : 'Terima barang'}
        </Button>
        {detail.status === 'sent' ? cancelConfirm : null}
      </>
    )
  }

  const receiptRows = receipts.data?.items ?? []

  return (
    <div className="space-y-6">
      <Card
        title={`Rincian pesanan ${detail.number}`}
        description={`${nameOf(vendors, detail.vendorId)} untuk ${nameOf(projects, detail.projectId)}, ${shortDate(detail.date)}`}
      >
        <Facts
          items={[
            {
              label: 'Status',
              value: <Chip label={ORDER_STATUS_LABEL[detail.status]} tone={ORDER_STATUS_TONE[detail.status]} />,
            },
            { label: 'Jatuh tempo', value: shortDate(detail.dueDate) },
            { label: 'Dari permintaan', value: detail.purchaseRequestId ? (requestNumber ?? '...') : '-' },
            { label: 'Nilai pesanan', value: rupiah(detail.value) },
          ]}
        />

        <div className="mb-4 space-y-3">
          <p className="text-[13px] text-slate-600">{STATUS_HINT[detail.status]}</p>
          {actions ? <div className="flex flex-wrap items-center gap-2">{actions}</div> : null}
          {updateStatus.isError ? <ErrorNote message={errorMessage(updateStatus.error)} /> : null}
          {deleteOrder.isError ? <ErrorNote message={errorMessage(deleteOrder.error)} /> : null}
          {notice ? <SuccessNote message={notice} /> : null}
        </div>

        <Table
          rows={detail.items}
          emptyMessage="Pesanan ini tidak punya baris material."
          columns={[
            { header: 'Material', cell: (row) => nameOf(materialItems, row.materialId) },
            { header: 'Jumlah', align: 'right', cell: (row) => number(row.quantity) },
            { header: 'Satuan', cell: (row) => codeOf(unitItems, row.unitOfMeasureId) },
            { header: 'Harga satuan', align: 'right', cell: (row) => rupiah(row.unitPrice) },
            { header: 'Subtotal', align: 'right', cell: (row) => rupiah(row.total) },
            { header: 'Diterima', align: 'right', cell: (row) => number(row.receivedQuantity) },
            { header: 'Sisa', align: 'right', cell: (row) => number(row.remainingQuantity) },
          ]}
        />
        <p className="mt-4 text-right text-sm font-medium text-slate-900 tabular-nums">Total {rupiah(detail.value)}</p>

        {detail.status !== 'draft' ? (
          <div className="mt-5 border-t border-slate-100 pt-4">
            <h3 className="mb-3 text-sm font-semibold text-slate-900">Riwayat penerimaan</h3>
            {receipts.isPending ? <Loading /> : null}
            {receipts.isError ? <LoadFailed onRetry={() => receipts.refetch()} /> : null}
            {receipts.data ? (
              <Table
                rows={receiptRows}
                emptyMessage="Belum ada barang yang diterima untuk pesanan ini."
                columns={[
                  { header: 'Nomor', cell: (row) => row.number },
                  { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                  { header: 'Gudang', cell: (row) => nameOf(warehouseItems, row.warehouseId) },
                  {
                    header: 'Kondisi',
                    cell: (row) => (
                      <Chip label={RECEIPT_CONDITION_LABEL[row.condition]} tone={RECEIPT_CONDITION_TONE[row.condition]} />
                    ),
                  },
                  {
                    header: 'Aksi',
                    align: 'right',
                    cell: (row) => (
                      <RowAction
                        label="Rincian"
                        isActive={row.id === receiptId}
                        onClick={() => setReceiptId((current) => (current === row.id ? null : row.id))}
                      />
                    ),
                  },
                ]}
              />
            ) : null}
          </div>
        ) : null}
      </Card>

      {mode === 'receive' && canReceive ? (
        <Card
          title={`Terima barang untuk ${detail.number}`}
          description={`Dari ${nameOf(vendors, detail.vendorId)}. Jumlah diterima sudah diisi sisa pesanan per baris.`}
        >
          <ReceiveGoodsForm
            orderId={detail.id}
            onSaved={(result) => {
              setMode('view')
              setReceiptId(result.receipt.id)
              setNotice(
                `Penerimaan ${result.receipt.number} tersimpan${
                  result.order ? `, status pesanan sekarang ${ORDER_STATUS_LABEL[result.order.status]}` : ''
                }. Stok gudang belum bertambah, catat stok masuk dari rincian penerimaan di bawah.`,
              )
            }}
          />
        </Card>
      ) : null}

      {receiptId ? <ReceiptDetailCard key={receiptId} id={receiptId} onClose={() => setReceiptId(null)} /> : null}
    </div>
  )
}

// prefill hanya dibaca saat bagian ini dibuka. Menutup formulir membuang isian
// dari permintaan, jadi membukanya lagi memberi formulir kosong.
export function OrdersSection({ prefill }: { prefill: OrderPrefill | null }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<PurchaseOrderStatus | ''>('')
  const [page, setPage] = useState(1)
  const [formSeed, setFormSeed] = useState<OrderPrefill | null>(prefill)
  const [isFormOpen, setIsFormOpen] = useState(prefill !== null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const orders = usePurchaseOrders({
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    page,
    pageSize: PAGE_SIZE,
  })
  const vendors = useVendorOptions()
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const vendorItems = vendors.data?.items ?? []
  const projectItems = projects.data?.items ?? []
  const isLoadingOptions = vendors.isPending || projects.isPending

  const toggleForm = () => {
    if (isFormOpen) {
      setFormSeed(null)
    }
    setIsFormOpen((open) => !open)
  }

  return (
    <div className="space-y-6">
      {isFormOpen ? (
        <PurchaseOrderFormCard
          editingId={null}
          initial={formSeed?.values ?? emptyPurchaseOrderForm()}
          requestNumber={formSeed?.requestNumber ?? null}
          vendors={vendorItems}
          projects={projectItems}
          isLoadingOptions={isLoadingOptions}
          onSaved={(order) => {
            setNotice(null)
            setSelectedId(order.id)
          }}
        />
      ) : null}

      <Card
        title="Daftar pesanan pembelian"
        description={orders.data ? `${number(orders.data.totalItems)} pesanan` : undefined}
      >
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <ToolbarInput
              id="order-search"
              label="Cari nomor PO"
              type="search"
              placeholder="Cari nomor PO"
              maxLength={ORDER_NUMBER_MAX_LENGTH}
              className="w-48"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
            />
            <FilterChips
              filters={ORDER_FILTERS}
              active={status}
              label="Saring status pesanan"
              onChange={(value) => {
                setStatus(value)
                setPage(1)
              }}
            />
          </div>
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={toggleForm}>
            {isFormOpen ? 'Tutup formulir' : 'Buat pesanan'}
          </Button>
        </div>

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {orders.isPending ? <Loading /> : null}
        {orders.isError ? <LoadFailed onRetry={() => orders.refetch()} /> : null}
        {orders.data ? (
          <>
            <Table
              rows={orders.data.items}
              emptyMessage={
                status === '' && search.trim() === ''
                  ? 'Belum ada pesanan pembelian. Buat pesanan pertama lewat tombol Buat pesanan.'
                  : 'Tidak ada pesanan yang cocok dengan penyaringan ini.'
              }
              columns={[
                { header: 'Nomor', cell: (row) => row.number },
                { header: 'Vendor', cell: (row) => nameOf(vendorItems, row.vendorId) },
                { header: 'Proyek', cell: (row) => nameOf(projectItems, row.projectId) },
                { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                { header: 'Jatuh tempo', cell: (row) => shortDate(row.dueDate) },
                { header: 'Nilai', align: 'right', cell: (row) => rupiahShort(row.value) },
                {
                  header: 'Status',
                  cell: (row) => <Chip label={ORDER_STATUS_LABEL[row.status]} tone={ORDER_STATUS_TONE[row.status]} />,
                },
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
              page={orders.data.page}
              totalPages={orders.data.totalPages}
              totalItems={orders.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <PurchaseOrderDetailCard
          key={selectedId}
          id={selectedId}
          vendors={vendorItems}
          projects={projectItems}
          isLoadingOptions={isLoadingOptions}
          onDeleted={(deletedNumber) => {
            setSelectedId(null)
            setNotice(`Pesanan ${deletedNumber} dihapus.`)
          }}
        />
      ) : null}
    </div>
  )
}
