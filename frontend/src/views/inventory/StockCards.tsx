import { useState } from 'react'

import { useLowStockMaterials, useStocks } from '../../controllers/useInventory'
import { shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { materialOptions, warehouseOptions } from '../../models/lookupApi'
import { Pager, Toolbar } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { PAGE_SIZE, quantityText, unitCode, type Lookups } from './inventoryShared'

const ALL_WAREHOUSES = warehouseOptions()

function NameWithCode({ name, code }: { name: string; code: string }) {
  return (
    <span className="flex flex-col">
      <span className="text-slate-900">{name}</span>
      <span className="text-xs text-slate-500">{code}</span>
    </span>
  )
}

export function StockCard() {
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
        <SearchSelect
          {...ALL_WAREHOUSES}
          id="stock-filter-warehouse"
          label="Saring menurut gudang"
          compact
          allowEmpty
          emptyLabel="Semua gudang"
          className="w-52"
          value={warehouseId}
          onChange={(value) => {
            setWarehouseId(value)
            setPage(1)
          }}
        />
        <SearchSelect
          {...materialOptions}
          id="stock-filter-material"
          label="Saring menurut material"
          compact
          allowEmpty
          emptyLabel="Semua material"
          className="w-52"
          value={materialId}
          onChange={(value) => {
            setMaterialId(value)
            setPage(1)
          }}
        />
      </Toolbar>

      {stocks.isPending ? <Loading /> : null}
      {stocks.isError ? <LoadFailed onRetry={() => stocks.refetch()} /> : null}
      {stocks.data ? (
        <>
          <Table
            rows={stocks.data.items}
            emptyMessage="Belum ada stok. Catat mutasi masuk untuk mengisi gudang."
            columns={[
              { header: 'Gudang', cell: (row) => <NameWithCode name={row.warehouseName} code={row.warehouseCode} /> },
              { header: 'Material', cell: (row) => <NameWithCode name={row.materialName} code={row.materialCode} /> },
              {
                header: 'Jumlah',
                align: 'right',
                cell: (row) => (
                  <span className="font-medium text-slate-900">
                    {quantityText(row.quantity, row.unitOfMeasureCode ?? '')}
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

export function LowStockCard({ lookups }: { lookups: Lookups }) {
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
