import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../controllers/useErp'
import {
  useAssets,
  useCreateMaterial,
  useCreateWarehouse,
  useEquipment,
  useLowStockMaterials,
  useMaterialOptions,
  useMaterials,
  useRecordStockMovement,
  useStockMovementCount,
  useStockMovements,
  useStocks,
  useUnitsOfMeasure,
  useWarehouseOptions,
  useWarehouses,
} from '../controllers/useInventory'
import {
  ASSET_STATUS_LABEL,
  EMPTY_MATERIAL_FORM,
  EMPTY_WAREHOUSE_FORM,
  EQUIPMENT_STATUS_LABEL,
  EQUIPMENT_STATUS_TONE,
  MATERIAL_CATEGORY_MAX_LENGTH,
  MATERIAL_CATEGORY_SUGGESTIONS,
  MATERIAL_CODE_MAX_LENGTH,
  MATERIAL_NAME_MAX_LENGTH,
  MOVEMENT_REFERENCE_MAX_LENGTH,
  MOVEMENT_TYPES,
  MOVEMENT_TYPE_LABEL,
  MOVEMENT_TYPE_TONE,
  WAREHOUSE_CODE_MAX_LENGTH,
  WAREHOUSE_NAME_MAX_LENGTH,
  emptyMovementForm,
  movementNeedsSource,
  movementNeedsTarget,
  type AdjustmentDirection,
  type Material,
  type MaterialFormValues,
  type MovementFormValues,
  type MovementType,
  type UnitOfMeasure,
  type Warehouse,
  type WarehouseFormValues,
} from '../models/inventory'
import type { Project } from '../models/project'
import { errorMessage } from '../shared/errorMessage'
import { number, rupiah, rupiahShort, shortDate } from '../shared/format'
import { firstDayOfMonth, todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { Card, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'
import { Button, ErrorNote, Field, SuccessNote } from './components/Form'
import {
  Chip,
  FilterSelect,
  FormPanel,
  FormToggle,
  Pager,
  SelectField,
  Toolbar,
  ToolbarInput,
} from './components/RecordControls'
import { kpiText } from './components/kpiText'

const PAGE_SIZE = 20
const CATEGORY_OPTIONS_ID = 'material-category-options'

interface Lookups {
  materials: Material[]
  warehouses: Warehouse[]
  units: UnitOfMeasure[]
  projects: Project[]
}

function materialLabel(lookups: Lookups, id: string): string {
  const material = lookups.materials.find((item) => item.id === id)
  return material ? `${material.name} (${material.code})` : id.slice(0, 8)
}

function unitCode(lookups: Lookups, unitId: string | undefined): string {
  return lookups.units.find((unit) => unit.id === unitId)?.code ?? ''
}

function materialUnit(lookups: Lookups, materialId: string): string {
  return unitCode(lookups, lookups.materials.find((item) => item.id === materialId)?.unitOfMeasureId)
}

function warehouseLabel(lookups: Lookups, id: string | null): string {
  if (id === null) {
    return '-'
  }
  return lookups.warehouses.find((item) => item.id === id)?.name ?? id.slice(0, 8)
}

function projectLabel(lookups: Lookups, id: string | null): string {
  if (id === null) {
    return 'Gudang pusat'
  }
  return lookups.projects.find((project) => project.id === id)?.name ?? id.slice(0, 8)
}

function quantityText(quantity: number, unit: string): string {
  return unit === '' ? number(quantity) : `${number(quantity)} ${unit}`
}

function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return fallback
  }
  return rupiah(amount)
}

function LowStockCard({ lookups }: { lookups: Lookups }) {
  const [page, setPage] = useState(1)
  const lowStock = useLowStockMaterials(page)

  return (
    <Card title="Stok menipis" description="Total stok semua gudang di bawah stok minimum">
      {lowStock.isPending ? <Loading /> : null}
      {lowStock.isError ? <LoadFailed onRetry={() => lowStock.refetch()} /> : null}
      {lowStock.data ? (
        <>
          <Table
            rows={lowStock.data.items}
            emptyMessage="Semua material masih di atas stok minimum."
            columns={[
              {
                header: 'Material',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="font-medium text-slate-900">{row.name}</span>
                    <span className="text-xs text-slate-500">{row.code}</span>
                  </span>
                ),
              },
              {
                header: 'Stok',
                align: 'right',
                cell: (row) => (
                  <span className="text-red-700">
                    {quantityText(row.totalQuantity, unitCode(lookups, row.unitOfMeasureId))}
                  </span>
                ),
              },
              {
                header: 'Minimum',
                align: 'right',
                cell: (row) => quantityText(row.minimumStock, unitCode(lookups, row.unitOfMeasureId)),
              },
            ]}
          />
          <Pager
            page={lowStock.data.page}
            totalPages={lowStock.data.totalPages}
            totalItems={lowStock.data.totalItems}
            unit="material menipis"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

function WarehouseOptions({ warehouses, placeholder }: { warehouses: Warehouse[]; placeholder: string }) {
  return (
    <>
      <option value="">{placeholder}</option>
      {warehouses.map((warehouse) => (
        <option key={warehouse.id} value={warehouse.id}>
          {warehouse.name} ({warehouse.code})
        </option>
      ))}
    </>
  )
}

function NewMovementForm({ lookups, today }: { lookups: Lookups; today: string }) {
  const [values, setValues] = useState<MovementFormValues>(() => emptyMovementForm(today))
  const recordMovement = useRecordStockMovement()
  const needsSource = movementNeedsSource(values)
  const needsTarget = movementNeedsTarget(values)
  const unit = materialUnit(lookups, values.materialId)

  const update =
    (key: keyof MovementFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    recordMovement.mutate(values, {
      onSuccess: () => setValues((current) => ({ ...current, quantity: '', reference: '' })),
    })
  }

  const targetsTransferSource = values.type === 'transfer' && values.sourceWarehouseId !== ''
  const targetWarehouses = targetsTransferSource
    ? lookups.warehouses.filter((warehouse) => warehouse.id !== values.sourceWarehouseId)
    : lookups.warehouses

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <p className="text-sm text-slate-600">
          Stok gudang langsung berubah setelah mutasi disimpan. Mutasi tidak bisa diubah, koreksi lewat penyesuaian.
        </p>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <SelectField
            id="movement-type"
            label="Jenis mutasi"
            value={values.type}
            onChange={(event) =>
              setValues((current) => ({ ...current, type: event.target.value as MovementType }))
            }
          >
            {MOVEMENT_TYPES.map((type) => (
              <option key={type} value={type}>
                {MOVEMENT_TYPE_LABEL[type]}
              </option>
            ))}
          </SelectField>
          <Field
            id="movement-date"
            label="Tanggal"
            type="date"
            required
            max={today}
            value={values.date}
            onChange={update('date')}
          />
          {values.type === 'adjustment' ? (
            <SelectField
              id="movement-direction"
              label="Arah penyesuaian"
              hint="Dipakai setelah stok opname"
              value={values.direction}
              onChange={(event) =>
                setValues((current) => ({ ...current, direction: event.target.value as AdjustmentDirection }))
              }
            >
              <option value="increase">Tambah stok</option>
              <option value="decrease">Kurangi stok</option>
            </SelectField>
          ) : null}
          <SelectField
            id="movement-material"
            label="Material"
            required
            value={values.materialId}
            onChange={update('materialId')}
          >
            <option value="">Pilih material</option>
            {lookups.materials.map((material) => (
              <option key={material.id} value={material.id}>
                {material.name} ({material.code})
              </option>
            ))}
          </SelectField>
          <Field
            id="movement-quantity"
            label={unit === '' ? 'Jumlah' : `Jumlah (${unit})`}
            type="number"
            inputMode="decimal"
            min={0.01}
            step={0.01}
            placeholder="0"
            hint="Lebih dari nol, maksimal dua angka di belakang koma"
            required
            value={values.quantity}
            onChange={update('quantity')}
          />
          {needsSource ? (
            <SelectField
              id="movement-source"
              label="Gudang asal"
              hint="Stok di gudang ini harus mencukupi"
              required
              value={values.sourceWarehouseId}
              onChange={(event) => {
                const sourceWarehouseId = event.target.value
                setValues((current) => ({
                  ...current,
                  sourceWarehouseId,
                  targetWarehouseId:
                    current.targetWarehouseId === sourceWarehouseId ? '' : current.targetWarehouseId,
                }))
              }}
            >
              <WarehouseOptions warehouses={lookups.warehouses} placeholder="Pilih gudang asal" />
            </SelectField>
          ) : null}
          {needsTarget ? (
            <SelectField
              id="movement-target"
              label="Gudang tujuan"
              required
              value={values.targetWarehouseId}
              onChange={update('targetWarehouseId')}
            >
              <WarehouseOptions warehouses={targetWarehouses} placeholder="Pilih gudang tujuan" />
            </SelectField>
          ) : null}
          <Field
            id="movement-reference"
            label="Referensi (opsional)"
            placeholder="PO-001 atau nomor surat jalan"
            autoComplete="off"
            maxLength={MOVEMENT_REFERENCE_MAX_LENGTH}
            value={values.reference}
            onChange={update('reference')}
          />
        </div>

        <Button type="submit" isPending={recordMovement.isPending} pendingLabel="Menyimpan mutasi">
          Simpan mutasi
        </Button>

        {recordMovement.isError ? <ErrorNote message={errorMessage(recordMovement.error)} /> : null}
        {recordMovement.isSuccess ? (
          <SuccessNote
            message={`Mutasi ${MOVEMENT_TYPE_LABEL[recordMovement.data.type].toLowerCase()} ${quantityText(
              recordMovement.data.quantity,
              materialUnit(lookups, recordMovement.data.materialId),
            )} ${materialLabel(lookups, recordMovement.data.materialId)} tersimpan.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}

function StockCard({ lookups }: { lookups: Lookups }) {
  const [warehouseId, setWarehouseId] = useState('')
  const [materialId, setMaterialId] = useState('')
  const [page, setPage] = useState(1)
  const stocks = useStocks({
    warehouseId: warehouseId === '' ? undefined : warehouseId,
    materialId: materialId === '' ? undefined : materialId,
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <Card title="Stok per gudang" description="Jumlah tersedia untuk setiap pasangan material dan gudang">
      <Toolbar>
        <FilterSelect
          id="stock-filter-warehouse"
          label="Saring menurut gudang"
          value={warehouseId}
          onChange={(event) => {
            setWarehouseId(event.target.value)
            setPage(1)
          }}
        >
          <WarehouseOptions warehouses={lookups.warehouses} placeholder="Semua gudang" />
        </FilterSelect>
        <FilterSelect
          id="stock-filter-material"
          label="Saring menurut material"
          value={materialId}
          onChange={(event) => {
            setMaterialId(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Semua material</option>
          {lookups.materials.map((material) => (
            <option key={material.id} value={material.id}>
              {material.name}
            </option>
          ))}
        </FilterSelect>
      </Toolbar>

      {stocks.isPending ? <Loading /> : null}
      {stocks.isError ? <LoadFailed onRetry={() => stocks.refetch()} /> : null}
      {stocks.data ? (
        <>
          <Table
            rows={stocks.data.items}
            emptyMessage="Belum ada stok. Catat mutasi masuk untuk mengisi gudang."
            columns={[
              {
                header: 'Gudang',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="text-slate-900">{row.warehouseName}</span>
                    <span className="text-xs text-slate-500">{row.warehouseCode}</span>
                  </span>
                ),
              },
              {
                header: 'Material',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="text-slate-900">{row.materialName}</span>
                    <span className="text-xs text-slate-500">{row.materialCode}</span>
                  </span>
                ),
              },
              {
                header: 'Jumlah',
                align: 'right',
                cell: (row) => (
                  <span className="font-medium text-slate-900">
                    {quantityText(row.quantity, materialUnit(lookups, row.materialId))}
                  </span>
                ),
              },
              { header: 'Diperbarui', align: 'right', cell: (row) => shortDate(row.updatedAt) },
            ]}
          />
          <Pager
            page={stocks.data.page}
            totalPages={stocks.data.totalPages}
            totalItems={stocks.data.totalItems}
            unit="baris stok"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

function MovementsCard({ lookups, today }: { lookups: Lookups; today: string }) {
  const [type, setType] = useState<MovementType | ''>('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const movements = useStockMovements({
    type: type === '' ? undefined : type,
    page,
    pageSize: 10,
  })

  return (
    <Card title="Mutasi stok" description="Barang masuk, keluar, pindah gudang, dan penyesuaian. Terbaru di atas.">
      <Toolbar>
        <FilterSelect
          id="movement-filter-type"
          label="Saring menurut jenis mutasi"
          value={type}
          onChange={(event) => {
            setType(event.target.value as MovementType | '')
            setPage(1)
          }}
        >
          <option value="">Semua jenis</option>
          {MOVEMENT_TYPES.map((option) => (
            <option key={option} value={option}>
              {MOVEMENT_TYPE_LABEL[option]}
            </option>
          ))}
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Catat mutasi" onToggle={() => setIsFormOpen((open) => !open)} />
      </Toolbar>

      {isFormOpen ? <NewMovementForm lookups={lookups} today={today} /> : null}

      {movements.isPending ? <Loading /> : null}
      {movements.isError ? <LoadFailed onRetry={() => movements.refetch()} /> : null}
      {movements.data ? (
        <>
          <Table
            rows={movements.data.items}
            emptyMessage="Belum ada mutasi stok."
            columns={[
              { header: 'Tanggal', cell: (row) => shortDate(row.date) },
              {
                header: 'Jenis',
                cell: (row) => <Chip tone={MOVEMENT_TYPE_TONE[row.type]}>{MOVEMENT_TYPE_LABEL[row.type]}</Chip>,
              },
              { header: 'Material', cell: (row) => materialLabel(lookups, row.materialId) },
              {
                header: 'Jumlah',
                align: 'right',
                cell: (row) => quantityText(row.quantity, materialUnit(lookups, row.materialId)),
              },
              { header: 'Dari', cell: (row) => warehouseLabel(lookups, row.sourceWarehouseId) },
              { header: 'Ke', cell: (row) => warehouseLabel(lookups, row.targetWarehouseId) },
              { header: 'Referensi', cell: (row) => row.reference || '-' },
            ]}
          />
          <Pager
            page={movements.data.page}
            totalPages={movements.data.totalPages}
            totalItems={movements.data.totalItems}
            unit="mutasi"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

function NewMaterialForm({ units }: { units: UnitOfMeasure[] }) {
  const [values, setValues] = useState<MaterialFormValues>(EMPTY_MATERIAL_FORM)
  const createMaterial = useCreateMaterial()
  const unit = units.find((item) => item.id === values.unitOfMeasureId)?.code ?? ''

  const update =
    (key: keyof MaterialFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createMaterial.mutate(values, { onSuccess: () => setValues(EMPTY_MATERIAL_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="material-code"
            label="Kode"
            placeholder="MAT-001"
            autoComplete="off"
            maxLength={MATERIAL_CODE_MAX_LENGTH}
            hint={`Unik, maksimal ${MATERIAL_CODE_MAX_LENGTH} karakter`}
            required
            value={values.code}
            onChange={update('code')}
          />
          <Field
            id="material-name"
            label="Nama material"
            placeholder="Semen Portland 50 kg"
            autoComplete="off"
            maxLength={MATERIAL_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="material-category"
            label="Kategori (opsional)"
            placeholder="Pilih atau ketik kategori"
            list={CATEGORY_OPTIONS_ID}
            autoComplete="off"
            maxLength={MATERIAL_CATEGORY_MAX_LENGTH}
            value={values.category}
            onChange={update('category')}
          />
          <SelectField
            id="material-unit"
            label="Satuan"
            required
            value={values.unitOfMeasureId}
            onChange={update('unitOfMeasureId')}
          >
            <option value="">Pilih satuan</option>
            {units.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name} ({item.code})
              </option>
            ))}
          </SelectField>
          <Field
            id="material-minimum"
            label={unit === '' ? 'Stok minimum' : `Stok minimum (${unit})`}
            type="number"
            inputMode="decimal"
            min={0}
            step={0.01}
            placeholder="0"
            hint="Di bawah angka ini material masuk daftar stok menipis"
            value={values.minimumStock}
            onChange={update('minimumStock')}
          />
          <Field
            id="material-price"
            label="Harga terakhir (opsional)"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={amountHint(values.lastPrice, unit === '' ? 'Dalam Rupiah per satuan' : `Dalam Rupiah per ${unit}`)}
            value={values.lastPrice}
            onChange={update('lastPrice')}
          />
        </div>

        <datalist id={CATEGORY_OPTIONS_ID}>
          {MATERIAL_CATEGORY_SUGGESTIONS.map((category) => (
            <option key={category} value={category} />
          ))}
        </datalist>

        <Button type="submit" isPending={createMaterial.isPending} pendingLabel="Menyimpan material">
          Simpan material
        </Button>

        {createMaterial.isError ? <ErrorNote message={errorMessage(createMaterial.error)} /> : null}
        {createMaterial.isSuccess ? (
          <SuccessNote message={`Material ${createMaterial.data.name} (${createMaterial.data.code}) berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function MaterialsCard({ lookups }: { lookups: Lookups }) {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const materials = useMaterials({
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <Card title="Material" description="Katalog material bangunan beserta satuan dan stok minimumnya">
      <Toolbar>
        <ToolbarInput
          id="material-search"
          label="Cari nama material"
          type="search"
          placeholder="Cari nama material"
          className="w-56"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        <FormToggle isOpen={isFormOpen} openLabel="Tambah material" onToggle={() => setIsFormOpen((open) => !open)} />
      </Toolbar>

      {isFormOpen ? <NewMaterialForm units={lookups.units} /> : null}

      {materials.isPending ? <Loading /> : null}
      {materials.isError ? <LoadFailed onRetry={() => materials.refetch()} /> : null}
      {materials.data ? (
        <>
          <Table
            rows={materials.data.items}
            emptyMessage="Tidak ada material yang cocok."
            columns={[
              { header: 'Kode', cell: (row) => row.code },
              { header: 'Nama', cell: (row) => row.name },
              { header: 'Kategori', cell: (row) => row.category || '-' },
              { header: 'Satuan', cell: (row) => unitCode(lookups, row.unitOfMeasureId) || '-' },
              {
                header: 'Stok minimum',
                align: 'right',
                cell: (row) => quantityText(row.minimumStock, unitCode(lookups, row.unitOfMeasureId)),
              },
              { header: 'Harga terakhir', align: 'right', cell: (row) => rupiah(row.lastPrice) },
            ]}
          />
          <Pager
            page={materials.data.page}
            totalPages={materials.data.totalPages}
            totalItems={materials.data.totalItems}
            unit="material"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

function NewWarehouseForm({ projects }: { projects: Project[] }) {
  const [values, setValues] = useState<WarehouseFormValues>(EMPTY_WAREHOUSE_FORM)
  const createWarehouse = useCreateWarehouse()

  const update =
    (key: keyof WarehouseFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createWarehouse.mutate(values, { onSuccess: () => setValues(EMPTY_WAREHOUSE_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="warehouse-code"
            label="Kode"
            placeholder="GDG-01"
            autoComplete="off"
            maxLength={WAREHOUSE_CODE_MAX_LENGTH}
            hint={`Unik, maksimal ${WAREHOUSE_CODE_MAX_LENGTH} karakter`}
            required
            value={values.code}
            onChange={update('code')}
          />
          <Field
            id="warehouse-name"
            label="Nama gudang"
            placeholder="Gudang Lapangan Blok A"
            autoComplete="off"
            maxLength={WAREHOUSE_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <SelectField
            id="warehouse-project"
            label="Proyek (opsional)"
            hint="Kosongkan untuk gudang pusat"
            value={values.projectId}
            onChange={update('projectId')}
          >
            <option value="">Gudang pusat, tanpa proyek</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name}
              </option>
            ))}
          </SelectField>
        </div>

        <Button type="submit" isPending={createWarehouse.isPending} pendingLabel="Menyimpan gudang">
          Simpan gudang
        </Button>

        {createWarehouse.isError ? <ErrorNote message={errorMessage(createWarehouse.error)} /> : null}
        {createWarehouse.isSuccess ? (
          <SuccessNote message={`Gudang ${createWarehouse.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function WarehousesCard({ lookups }: { lookups: Lookups }) {
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const warehouses = useWarehouses({
    projectId: projectId === '' ? undefined : projectId,
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <Card title="Gudang" description="Gudang pusat dan gudang lapangan per proyek">
      <Toolbar>
        <FilterSelect
          id="warehouse-filter-project"
          label="Saring menurut proyek"
          value={projectId}
          onChange={(event) => {
            setProjectId(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Semua proyek</option>
          {lookups.projects.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name}
            </option>
          ))}
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Tambah gudang" onToggle={() => setIsFormOpen((open) => !open)} />
      </Toolbar>

      {isFormOpen ? <NewWarehouseForm projects={lookups.projects} /> : null}

      {warehouses.isPending ? <Loading /> : null}
      {warehouses.isError ? <LoadFailed onRetry={() => warehouses.refetch()} /> : null}
      {warehouses.data ? (
        <>
          <Table
            rows={warehouses.data.items}
            emptyMessage="Belum ada gudang untuk penyaringan ini."
            columns={[
              { header: 'Kode', cell: (row) => row.code },
              { header: 'Nama', cell: (row) => row.name },
              { header: 'Proyek', cell: (row) => projectLabel(lookups, row.projectId) },
            ]}
          />
          <Pager
            page={warehouses.data.page}
            totalPages={warehouses.data.totalPages}
            totalItems={warehouses.data.totalItems}
            unit="gudang"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

function AssetsCard({ lookups }: { lookups: Lookups }) {
  const [assetPage, setAssetPage] = useState(1)
  const [equipmentPage, setEquipmentPage] = useState(1)
  const assets = useAssets(assetPage)
  const equipment = useEquipment(equipmentPage)

  return (
    <Card title="Aset & alat" description="Hanya baca. Nilai buku dihitung server dari nilai perolehan dan penyusutan.">
      <div className="space-y-6">
        <div>
          {assets.isPending ? <Loading /> : null}
          {assets.isError ? <LoadFailed onRetry={() => assets.refetch()} /> : null}
          {assets.data ? (
            <>
              <dl className="mb-4 grid gap-3 sm:grid-cols-3">
                {[
                  { label: 'Nilai perolehan', value: assets.data.totalAcquisitionValue },
                  { label: 'Akumulasi penyusutan', value: assets.data.totalAccumulatedDepreciation },
                  { label: 'Nilai buku', value: assets.data.totalBookValue },
                ].map((item) => (
                  <div key={item.label} className="rounded-lg bg-slate-50 px-3 py-2">
                    <dt className="text-xs text-slate-500">{item.label}</dt>
                    <dd className="text-sm font-semibold text-slate-900 tabular-nums">{rupiahShort(item.value)}</dd>
                  </div>
                ))}
              </dl>
              <Table
                rows={assets.data.items}
                emptyMessage="Belum ada aset tercatat."
                columns={[
                  { header: 'Aset', cell: (row) => `${row.name} (${row.code})` },
                  { header: 'Kategori', cell: (row) => row.category },
                  { header: 'Status', cell: (row) => ASSET_STATUS_LABEL[row.status] },
                  { header: 'Nilai buku', align: 'right', cell: (row) => rupiahShort(row.bookValue) },
                ]}
              />
              <Pager
                page={assets.data.page}
                totalPages={assets.data.totalPages}
                totalItems={assets.data.totalItems}
                unit="aset"
                onChange={setAssetPage}
              />
            </>
          ) : null}
        </div>

        <div>
          <h3 className="mb-2 text-xs font-medium tracking-wide text-slate-500 uppercase">Alat berat dan peralatan</h3>
          {equipment.isPending ? <Loading /> : null}
          {equipment.isError ? <LoadFailed onRetry={() => equipment.refetch()} /> : null}
          {equipment.data ? (
            <>
              <Table
                rows={equipment.data.items}
                emptyMessage="Belum ada alat tercatat."
                columns={[
                  { header: 'Alat', cell: (row) => `${row.name} (${row.code})` },
                  { header: 'Proyek', cell: (row) => (row.projectId ? projectLabel(lookups, row.projectId) : '-') },
                  {
                    header: 'Status',
                    cell: (row) => (
                      <Chip tone={EQUIPMENT_STATUS_TONE[row.status]}>{EQUIPMENT_STATUS_LABEL[row.status]}</Chip>
                    ),
                  },
                  { header: 'Servis berikutnya', align: 'right', cell: (row) => shortDate(row.nextServiceDate) },
                ]}
              />
              <Pager
                page={equipment.data.page}
                totalPages={equipment.data.totalPages}
                totalItems={equipment.data.totalItems}
                unit="alat"
                onChange={setEquipmentPage}
              />
            </>
          ) : null}
        </div>
      </div>
    </Card>
  )
}

export function InventoryPage() {
  const [today] = useState(todayDate)
  const materials = useMaterialOptions()
  const warehouses = useWarehouseOptions()
  const units = useUnitsOfMeasure()
  const projects = useProjects({ pageSize: 100 })
  const lowStock = useLowStockMaterials(1)
  const movementsThisMonth = useStockMovementCount({ dateFrom: firstDayOfMonth(today), dateTo: today })

  const lookups: Lookups = {
    materials: materials.data?.items ?? [],
    warehouses: warehouses.data?.items ?? [],
    units: units.data?.items ?? [],
    projects: projects.data?.items ?? [],
  }

  const lowStockCount = lowStock.data?.totalItems ?? 0

  return (
    <AppShell title="Inventaris" description="Material, gudang, stok, dan mutasi barang">
      <KpiRow>
        <Kpi
          label="Jenis material"
          value={kpiText(materials, (data) => number(data.totalItems))}
          note="tercatat di katalog"
        />
        <Kpi
          label="Gudang"
          value={kpiText(warehouses, (data) => number(data.totalItems))}
          note="gudang pusat dan lapangan proyek"
        />
        <Kpi
          label="Stok menipis"
          value={kpiText(lowStock, (data) => number(data.totalItems))}
          note={lowStockCount > 0 ? 'perlu diadakan ulang' : 'semua di atas minimum'}
        />
        <Kpi
          label="Mutasi bulan ini"
          value={kpiText(movementsThisMonth, number)}
          note="masuk, keluar, pindah, dan penyesuaian"
        />
      </KpiRow>

      <div className="mt-6 grid gap-6">
        <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
          <StockCard lookups={lookups} />
          <LowStockCard lookups={lookups} />
        </div>

        <MovementsCard lookups={lookups} today={today} />

        <MaterialsCard lookups={lookups} />

        <div className="grid gap-6 xl:grid-cols-2">
          <WarehousesCard lookups={lookups} />
          <AssetsCard lookups={lookups} />
        </div>
      </div>
    </AppShell>
  )
}
