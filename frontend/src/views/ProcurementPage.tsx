import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../controllers/useErp'
import {
  useActiveVendorCount,
  useCreatePurchaseOrder,
  useCreateVendor,
  useMaterials,
  usePurchaseOrder,
  usePurchaseOrders,
  usePurchaseRequestCount,
  usePurchaseRequests,
  useUnitsOfMeasure,
  useUpdatePurchaseOrderStatus,
  useVendorOptions,
  useVendors,
} from '../controllers/useProcurement'
import {
  EMPTY_VENDOR_FORM,
  OPEN_ORDER_STATUSES,
  ORDER_NUMBER_MAX_LENGTH,
  ORDER_STATUS_LABEL,
  ORDER_STATUS_TONE,
  REQUEST_STATUS_LABEL,
  REQUEST_STATUS_TONE,
  VENDOR_CATEGORY_MAX_LENGTH,
  VENDOR_CATEGORY_SUGGESTIONS,
  VENDOR_CODE_MAX_LENGTH,
  VENDOR_CONTACT_MAX_LENGTH,
  VENDOR_NAME_MAX_LENGTH,
  VENDOR_RATING_LABEL,
  VENDOR_RATING_TONE,
  VENDOR_TAX_NUMBER_MAX_LENGTH,
  emptyOrderLine,
  emptyPurchaseOrderForm,
  orderLineTotal,
  type Material,
  type OrderLineValues,
  type PurchaseOrderFormValues,
  type PurchaseOrderStatus,
  type PurchaseRequestStatus,
  type UnitOfMeasure,
  type Vendor,
  type VendorFormValues,
  type VendorRating,
} from '../models/procurement'
import type { Project } from '../models/project'
import { errorMessage } from '../shared/errorMessage'
import { number, rupiah, rupiahShort, shortDate } from '../shared/format'
import { kpiText } from '../shared/kpiText'
import { AppShell } from './components/AppShell'
import { Card, Empty, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'
import { Button, CONTROL_CLASS, ErrorNote, Field, SuccessNote } from './components/Form'
import { Chip, FilterChips, Pager, RowAction, SectionTabs, SelectField } from './components/ListTools'

const PAGE_SIZE = 10
const OPTION_LIMIT = 100
const CATEGORY_OPTIONS_ID = 'vendor-category-options'

type ProcurementTab = 'orders' | 'vendors' | 'requests'

const TABS: { value: ProcurementTab; label: string }[] = [
  { value: 'orders', label: 'Pesanan pembelian' },
  { value: 'vendors', label: 'Vendor' },
  { value: 'requests', label: 'Permintaan pembelian' },
]

const ORDER_FILTERS: { value: PurchaseOrderStatus | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'draft', label: ORDER_STATUS_LABEL.draft },
  { value: 'sent', label: ORDER_STATUS_LABEL.sent },
  { value: 'partially_received', label: ORDER_STATUS_LABEL.partially_received },
  { value: 'completed', label: ORDER_STATUS_LABEL.completed },
  { value: 'cancelled', label: ORDER_STATUS_LABEL.cancelled },
]

const REQUEST_FILTERS: { value: PurchaseRequestStatus | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'draft', label: REQUEST_STATUS_LABEL.draft },
  { value: 'submitted', label: REQUEST_STATUS_LABEL.submitted },
  { value: 'approved', label: REQUEST_STATUS_LABEL.approved },
  { value: 'rejected', label: REQUEST_STATUS_LABEL.rejected },
  { value: 'completed', label: REQUEST_STATUS_LABEL.completed },
]

type ActiveFilter = 'all' | 'active' | 'inactive'

const ACTIVE_FILTERS: { value: ActiveFilter; label: string }[] = [
  { value: 'all', label: 'Semua' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Nonaktif' },
]

const VENDOR_RATINGS: VendorRating[] = ['new', 'good', 'fair', 'poor']

function nameOf(items: { id: string; name: string }[], id: string): string {
  return items.find((item) => item.id === id)?.name ?? '-'
}

function codeOf(items: { id: string; code: string }[], id: string): string {
  return items.find((item) => item.id === id)?.code ?? '-'
}

function ProcurementKpis() {
  const orders = usePurchaseOrders({ pageSize: OPTION_LIMIT })
  const activeVendors = useActiveVendorCount()
  const pendingRequests = usePurchaseRequestCount('submitted')

  const items = orders.data?.items ?? []
  const open = items.filter((order) => OPEN_ORDER_STATUSES.includes(order.status))
  const openValue = open.reduce((total, order) => total + order.value, 0)
  const totalValue = items
    .filter((order) => order.status !== 'cancelled')
    .reduce((total, order) => total + order.value, 0)
  const isPartial = (orders.data?.totalItems ?? 0) > items.length
  const scope = isPartial ? `, dari ${OPTION_LIMIT} PO terbaru` : ''

  return (
    <KpiRow>
      <Kpi
        label="PO terbuka"
        value={kpiText(orders, () => number(open.length))}
        note={orders.data ? `draf, dikirim, atau diterima sebagian${scope}` : undefined}
      />
      <Kpi
        label="Nilai PO terbuka"
        value={kpiText(orders, () => rupiahShort(openValue))}
        note={orders.data ? `total ${rupiahShort(totalValue)} di luar yang batal${scope}` : undefined}
      />
      <Kpi label="Vendor aktif" value={kpiText(activeVendors, number)} note="bisa menerima pesanan baru" />
      <Kpi label="Permintaan menunggu" value={kpiText(pendingRequests, number)} note="diajukan, belum disetujui" />
    </KpiRow>
  )
}

function OrderLineFields({ line, index, materials, units, usedMaterials, canRemove, onChange, onRemove }: {
  line: OrderLineValues
  index: number
  materials: Material[]
  units: UnitOfMeasure[]
  usedMaterials: string[]
  canRemove: boolean
  onChange: (line: OrderLineValues) => void
  onRemove: () => void
}) {
  const id = `po-line-${line.key}`

  // Memilih material mengisi satuan bawaan dan harga terakhirnya. Keduanya
  // tetap boleh diubah sesudahnya.
  const chooseMaterial = (event: ChangeEvent<HTMLSelectElement>) => {
    const material = materials.find((item) => item.id === event.target.value)
    onChange({
      ...line,
      materialId: event.target.value,
      unitOfMeasureId: material?.unitOfMeasureId ?? line.unitOfMeasureId,
      unitPrice: line.unitPrice === '' && material ? String(material.lastPrice) : line.unitPrice,
    })
  }

  const update = (key: 'quantity' | 'unitOfMeasureId' | 'unitPrice') =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => onChange({ ...line, [key]: event.target.value })

  return (
    <fieldset className="rounded-xl bg-white p-4 shadow-hairline">
      <legend className="float-left mb-3 w-full text-xs font-medium text-slate-500">Baris {index + 1}</legend>
      <div className="clear-both grid gap-4 md:grid-cols-[2fr_1fr_1fr_1fr]">
        <SelectField id={`${id}-material`} label="Material" required value={line.materialId} onChange={chooseMaterial}>
          <option value="">Pilih material</option>
          {materials.map((material) => (
            <option
              key={material.id}
              value={material.id}
              disabled={material.id !== line.materialId && usedMaterials.includes(material.id)}
            >
              {material.name} ({material.code})
            </option>
          ))}
        </SelectField>
        <Field
          id={`${id}-quantity`}
          label="Jumlah"
          type="number"
          inputMode="decimal"
          min={0.01}
          step={0.01}
          required
          value={line.quantity}
          onChange={update('quantity')}
        />
        <SelectField
          id={`${id}-unit`}
          label="Satuan"
          required
          value={line.unitOfMeasureId}
          onChange={update('unitOfMeasureId')}
        >
          <option value="">Pilih satuan</option>
          {units.map((unit) => (
            <option key={unit.id} value={unit.id}>
              {unit.name}
            </option>
          ))}
        </SelectField>
        <Field
          id={`${id}-price`}
          label="Harga satuan"
          type="number"
          inputMode="numeric"
          min={0}
          step={1}
          placeholder="0"
          value={line.unitPrice}
          onChange={update('unitPrice')}
        />
      </div>
      <div className="mt-3 flex items-center justify-between gap-3">
        <p className="text-sm text-slate-600 tabular-nums">Subtotal {rupiah(orderLineTotal(line))}</p>
        {canRemove ? (
          <button
            type="button"
            onClick={onRemove}
            className="rounded-md px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
          >
            Hapus baris {index + 1}
          </button>
        ) : null}
      </div>
    </fieldset>
  )
}

function NewPurchaseOrderCard({ vendors, projects, isLoadingOptions }: {
  vendors: Vendor[]
  projects: Project[]
  isLoadingOptions: boolean
}) {
  const [values, setValues] = useState<PurchaseOrderFormValues>(emptyPurchaseOrderForm)
  const createOrder = useCreatePurchaseOrder()
  const materials = useMaterials()
  const units = useUnitsOfMeasure()
  const materialItems = materials.data?.items ?? []
  const unitItems = units.data?.items ?? []
  const activeVendors = vendors.filter((vendor) => vendor.active)
  const total = values.items.reduce((sum, line) => sum + orderLineTotal(line), 0)
  const usedMaterials = values.items.map((line) => line.materialId).filter((id) => id !== '')

  const update = (key: 'number' | 'vendorId' | 'projectId' | 'date' | 'dueDate') =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const changeLine = (key: string) => (line: OrderLineValues) =>
    setValues((current) => ({ ...current, items: current.items.map((item) => (item.key === key ? line : item)) }))

  const removeLine = (key: string) => () =>
    setValues((current) => ({ ...current, items: current.items.filter((item) => item.key !== key) }))

  const addLine = () => setValues((current) => ({ ...current, items: [...current.items, emptyOrderLine()] }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createOrder.mutate(values, { onSuccess: () => setValues(emptyPurchaseOrderForm()) })
  }

  if (materials.isPending || units.isPending || isLoadingOptions) {
    return (
      <Card title="Pesanan pembelian baru">
        <Loading label="Mengambil vendor, proyek, material, dan satuan..." />
      </Card>
    )
  }

  if (materials.isError || units.isError) {
    return (
      <Card title="Pesanan pembelian baru">
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
      : activeVendors.length === 0
        ? 'Belum ada vendor aktif. Tambahkan vendor di tab Vendor.'
        : projects.length === 0
          ? 'Belum ada proyek. Buat proyek dulu di menu Proyek.'
          : null

  return (
    <Card
      title="Pesanan pembelian baru"
      description="Pesanan baru berstatus Draf. Nilai pesanan dihitung server dari jumlah dikali harga satuan."
    >
      {missing ? (
        <Empty message={missing} />
      ) : (
        <form onSubmit={submit} className="space-y-5">
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
              {activeVendors.map((vendor) => (
                <option key={vendor.id} value={vendor.id}>
                  {vendor.name} ({vendor.code})
                </option>
              ))}
            </SelectField>
            <SelectField id="po-project" label="Proyek" required value={values.projectId} onChange={update('projectId')}>
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

          <div className="space-y-3">
            {values.items.map((line, index) => (
              <OrderLineFields
                key={line.key}
                line={line}
                index={index}
                materials={materialItems}
                units={unitItems}
                usedMaterials={usedMaterials}
                canRemove={values.items.length > 1}
                onChange={changeLine(line.key)}
                onRemove={removeLine(line.key)}
              />
            ))}
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Button variant="subtle" onClick={addLine}>
                Tambah baris material
              </Button>
              <p className="text-sm font-medium text-slate-900 tabular-nums">Total {rupiah(total)}</p>
            </div>
          </div>

          <Button type="submit" isPending={createOrder.isPending} pendingLabel="Menyimpan pesanan">
            Simpan pesanan
          </Button>

          {createOrder.isError ? <ErrorNote message={errorMessage(createOrder.error)} /> : null}
          {createOrder.isSuccess ? (
            <SuccessNote
              message={`Pesanan ${createOrder.data.number} senilai ${rupiah(createOrder.data.value)} berhasil dibuat.`}
            />
          ) : null}
        </form>
      )}
    </Card>
  )
}

function PurchaseOrderDetailCard({ id, vendors, projects }: { id: string; vendors: Vendor[]; projects: Project[] }) {
  const order = usePurchaseOrder(id)
  const materials = useMaterials()
  const units = useUnitsOfMeasure()
  const updateStatus = useUpdatePurchaseOrderStatus()
  const materialItems = materials.data?.items ?? []
  const unitItems = units.data?.items ?? []

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

  return (
    <Card
      title={`Rincian pesanan ${detail.number}`}
      description={`${nameOf(vendors, detail.vendorId)} untuk ${nameOf(projects, detail.projectId)}, ${shortDate(detail.date)}`}
    >
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <span className="flex items-center gap-3">
          <Chip label={ORDER_STATUS_LABEL[detail.status]} tone={ORDER_STATUS_TONE[detail.status]} />
          <span className="text-sm text-slate-600">Jatuh tempo {shortDate(detail.dueDate)}</span>
        </span>
        {detail.status === 'draft' ? (
          <Button
            isPending={updateStatus.isPending}
            pendingLabel="Mengirim"
            onClick={() => updateStatus.mutate({ id: detail.id, status: 'sent' })}
          >
            Kirim ke vendor
          </Button>
        ) : null}
      </div>
      {updateStatus.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(updateStatus.error)} />
        </div>
      ) : null}

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
    </Card>
  )
}

function OrdersSection() {
  const [status, setStatus] = useState<PurchaseOrderStatus | ''>('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const orders = usePurchaseOrders({ status: status === '' ? undefined : status, page, pageSize: PAGE_SIZE })
  const vendors = useVendorOptions()
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const vendorItems = vendors.data?.items ?? []
  const projectItems = projects.data?.items ?? []

  return (
    <div className="space-y-6">
      {isFormOpen ? (
        <NewPurchaseOrderCard
          vendors={vendorItems}
          projects={projectItems}
          isLoadingOptions={vendors.isPending || projects.isPending}
        />
      ) : null}

      <Card
        title="Daftar pesanan pembelian"
        description={orders.data ? `${number(orders.data.totalItems)} pesanan` : undefined}
      >
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <FilterChips
            filters={ORDER_FILTERS}
            active={status}
            label="Saring status pesanan"
            onChange={(value) => {
              setStatus(value)
              setPage(1)
            }}
          />
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={() => setIsFormOpen((open) => !open)}>
            {isFormOpen ? 'Tutup formulir' : 'Buat pesanan'}
          </Button>
        </div>

        {orders.isPending ? <Loading /> : null}
        {orders.isError ? <LoadFailed onRetry={() => orders.refetch()} /> : null}
        {orders.data ? (
          <>
            <Table
              rows={orders.data.items}
              emptyMessage={
                status === ''
                  ? 'Belum ada pesanan pembelian. Buat pesanan pertama lewat tombol Buat pesanan.'
                  : 'Tidak ada pesanan dengan status ini.'
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
        <PurchaseOrderDetailCard key={selectedId} id={selectedId} vendors={vendorItems} projects={projectItems} />
      ) : null}
    </div>
  )
}

function NewVendorCard() {
  const [values, setValues] = useState<VendorFormValues>(EMPTY_VENDOR_FORM)
  const createVendor = useCreateVendor()

  const update = (key: keyof VendorFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createVendor.mutate(values, { onSuccess: () => setValues(EMPTY_VENDOR_FORM) })
  }

  return (
    <Card title="Vendor baru" description="Vendor baru langsung aktif dan bisa dipilih saat membuat pesanan pembelian.">
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

        <Button type="submit" isPending={createVendor.isPending} pendingLabel="Menyimpan vendor">
          Simpan vendor
        </Button>

        {createVendor.isError ? <ErrorNote message={errorMessage(createVendor.error)} /> : null}
        {createVendor.isSuccess ? (
          <SuccessNote message={`Vendor ${createVendor.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </Card>
  )
}

function VendorsSection() {
  const [search, setSearch] = useState('')
  const [activeFilter, setActiveFilter] = useState<ActiveFilter>('all')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const vendors = useVendors({
    search: search.trim() === '' ? undefined : search.trim(),
    active: activeFilter === 'all' ? undefined : activeFilter === 'active',
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <div className="space-y-6">
      {isFormOpen ? <NewVendorCard /> : null}

      <Card title="Daftar vendor" description={vendors.data ? `${number(vendors.data.totalItems)} vendor` : undefined}>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <label htmlFor="vendor-search" className="sr-only">
              Cari vendor
            </label>
            <input
              id="vendor-search"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
              placeholder="Cari nama vendor"
              className={`${CONTROL_CLASS} h-8 w-56 px-3 text-[13px]`}
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
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={() => setIsFormOpen((open) => !open)}>
            {isFormOpen ? 'Tutup formulir' : 'Tambah vendor'}
          </Button>
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

function RequestsSection() {
  const [status, setStatus] = useState<PurchaseRequestStatus | ''>('')
  const [page, setPage] = useState(1)
  const requests = usePurchaseRequests({ status: status === '' ? undefined : status, page, pageSize: PAGE_SIZE })
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const projectItems = projects.data?.items ?? []

  return (
    <Card
      title="Permintaan pembelian"
      description="Kebutuhan material dari proyek. Pesanan bisa dibuat dari permintaan yang sudah disetujui."
    >
      <div className="mb-4">
        <FilterChips
          filters={REQUEST_FILTERS}
          active={status}
          label="Saring status permintaan"
          onChange={(value) => {
            setStatus(value)
            setPage(1)
          }}
        />
      </div>

      {requests.isPending ? <Loading /> : null}
      {requests.isError ? <LoadFailed onRetry={() => requests.refetch()} /> : null}
      {requests.data ? (
        <>
          <Table
            rows={requests.data.items}
            emptyMessage={
              status === '' ? 'Belum ada permintaan pembelian.' : 'Tidak ada permintaan dengan status ini.'
            }
            columns={[
              { header: 'Nomor', cell: (row) => row.number },
              { header: 'Proyek', cell: (row) => nameOf(projectItems, row.projectId) },
              { header: 'Tanggal', cell: (row) => shortDate(row.date) },
              { header: 'Jumlah barang', align: 'right', cell: (row) => number(row.itemCount) },
              { header: 'Catatan', cell: (row) => row.note || '-' },
              {
                header: 'Status',
                cell: (row) => (
                  <Chip label={REQUEST_STATUS_LABEL[row.status]} tone={REQUEST_STATUS_TONE[row.status]} />
                ),
              },
            ]}
          />
          <Pager
            page={requests.data.page}
            totalPages={requests.data.totalPages}
            totalItems={requests.data.totalItems}
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

export function ProcurementPage() {
  const [tab, setTab] = useState<ProcurementTab>('orders')

  return (
    <AppShell title="Pengadaan" description="Vendor, permintaan, dan pesanan pembelian material proyek">
      <ProcurementKpis />

      <div className="mt-6 mb-4">
        <SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian pengadaan" />
      </div>

      {tab === 'orders' ? <OrdersSection /> : null}
      {tab === 'vendors' ? <VendorsSection /> : null}
      {tab === 'requests' ? <RequestsSection /> : null}
    </AppShell>
  )
}
