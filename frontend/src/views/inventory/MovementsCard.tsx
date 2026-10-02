import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useRecordStockMovement, useStockMovements } from '../../controllers/useInventory'
import {
  MOVEMENT_REFERENCE_MAX_LENGTH,
  MOVEMENT_TYPES,
  MOVEMENT_TYPE_LABEL,
  MOVEMENT_TYPE_TONE,
  emptyMovementForm,
  movementNeedsSource,
  movementNeedsTarget,
  type AdjustmentDirection,
  type MovementFormValues,
  type MovementType,
  type StockMovement,
} from '../../models/inventory'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import {
  Chip,
  FilterSelect,
  FormPanel,
  FormToggle,
  Pager,
  SelectField,
  Toolbar,
} from '../components/RecordControls'
import { quantityText, unitCode, warehouseText, type Lookups } from './inventoryShared'
import { materialOptions, warehouseOptions } from '../../models/lookupApi'
import { SearchSelect } from '../components/SearchSelect'

const ALL_WAREHOUSES = warehouseOptions()

const MOVEMENT_PAGE_SIZE = 10

function materialText(movement: StockMovement): string {
  if (!movement.materialName) {
    return movement.materialId.slice(0, 8)
  }
  return `${movement.materialName} (${movement.materialCode})`
}

function movementQuantity(movement: StockMovement): string {
  return quantityText(movement.quantity, movement.unitOfMeasureCode ?? '')
}

function NewMovementForm({ lookups, today }: { lookups: Lookups; today: string }) {
  const [values, setValues] = useState<MovementFormValues>(() => emptyMovementForm(today))
  const recordMovement = useRecordStockMovement()
  const needsSource = movementNeedsSource(values)
  const needsTarget = movementNeedsTarget(values)
  const [materialUnitId, setMaterialUnitId] = useState<string>()
  const unit = values.materialId === '' ? '' : unitCode(lookups, materialUnitId)

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
            onChange={(event) => setValues((current) => ({ ...current, type: event.target.value as MovementType }))}
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
          <SearchSelect
            {...materialOptions}
            id="movement-material"
            label="Material"
            placeholder="Cari nama atau kode material"
            required
            value={values.materialId}
            onChange={(materialId, material) => {
              setMaterialUnitId(material?.unitOfMeasureId)
              setValues((current) => ({ ...current, materialId }))
            }}
          />
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
            <SearchSelect
              {...ALL_WAREHOUSES}
              id="movement-source"
              label="Gudang asal"
              placeholder="Cari gudang asal"
              hint="Stok di gudang ini harus mencukupi"
              required
              value={values.sourceWarehouseId}
              onChange={(sourceWarehouseId) =>
                setValues((current) => ({
                  ...current,
                  sourceWarehouseId,
                  targetWarehouseId: current.targetWarehouseId === sourceWarehouseId ? '' : current.targetWarehouseId,
                }))
              }
            />
          ) : null}
          {needsTarget ? (
            <SearchSelect
              {...ALL_WAREHOUSES}
              id="movement-target"
              label="Gudang tujuan"
              placeholder="Cari gudang tujuan"
              required
              exclude={targetsTransferSource ? [values.sourceWarehouseId] : undefined}
              value={values.targetWarehouseId}
              onChange={(targetWarehouseId) => setValues((current) => ({ ...current, targetWarehouseId }))}
            />
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
            message={`Mutasi ${MOVEMENT_TYPE_LABEL[recordMovement.data.type].toLowerCase()} ${movementQuantity(
              recordMovement.data,
            )} ${materialText(recordMovement.data)} tersimpan.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}

export function MovementsCard({ lookups, today }: { lookups: Lookups; today: string }) {
  const [type, setType] = useState<MovementType | ''>('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const movements = useStockMovements({
    type: type === '' ? undefined : type,
    page,
    pageSize: MOVEMENT_PAGE_SIZE,
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
              { header: 'Material', cell: materialText },
              { header: 'Jumlah', align: 'right', cell: movementQuantity },
              { header: 'Dari', cell: (row) => warehouseText(row.sourceWarehouseId, row.sourceWarehouseName) },
              { header: 'Ke', cell: (row) => warehouseText(row.targetWarehouseId, row.targetWarehouseName) },
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
