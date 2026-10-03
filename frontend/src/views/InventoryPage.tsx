import { useState } from 'react'

import {
  useLowStockMaterials,
  useMaterials,
  useStockMovementCount,
  useUnitsOfMeasure,
  useWarehouses,
} from '../controllers/useInventory'
import { number } from '../shared/format'
import { firstDayOfMonth, todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { Kpi, KpiRow } from './components/Data'
import { SectionTabs } from './components/ListTools'
import { kpiText } from './components/kpiText'
import { AssetsCard } from './inventory/AssetsCard'
import { EquipmentCard } from './inventory/EquipmentCard'
import type { Lookups } from './inventory/inventoryShared'
import { MaterialsCard } from './inventory/MaterialsCard'
import { MovementsCard } from './inventory/MovementsCard'
import { LowStockCard, StockCard } from './inventory/StockCards'
import { WarehousesCard } from './inventory/WarehousesCard'

type InventoryTab = 'stock' | 'catalog' | 'assets'

const TABS: { value: InventoryTab; label: string }[] = [
  { value: 'stock', label: 'Stok dan mutasi' },
  { value: 'catalog', label: 'Material dan gudang' },
  { value: 'assets', label: 'Aset dan alat' },
]

export function InventoryPage() {
  const [today] = useState(todayDate)
  const [tab, setTab] = useState<InventoryTab>('stock')
  const materials = useMaterials({ pageSize: 1 })
  const warehouses = useWarehouses({ pageSize: 1 })
  const units = useUnitsOfMeasure()
  const lowStock = useLowStockMaterials(1)
  const movementsThisMonth = useStockMovementCount({ dateFrom: firstDayOfMonth(today), dateTo: today })

  const lookups: Lookups = {
    units: units.data?.items ?? [],
  }

  const lowStockCount = lowStock.data?.totalItems ?? 0

  return (
    <AppShell title="Inventaris" description="Material, gudang, stok, mutasi barang, aset, dan alat">
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

      <div className="mt-6 mb-4">
        <SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian inventaris" />
      </div>

      {tab === 'stock' ? (
        <div className="grid gap-6">
          <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
            <StockCard />
            <LowStockCard lookups={lookups} />
          </div>
          <MovementsCard lookups={lookups} today={today} />
        </div>
      ) : null}

      {tab === 'catalog' ? (
        <div className="grid gap-6">
          <MaterialsCard />
          <WarehousesCard />
        </div>
      ) : null}

      {tab === 'assets' ? (
        <div className="grid gap-6">
          <AssetsCard />
          <EquipmentCard today={today} />
        </div>
      ) : null}
    </AppShell>
  )
}
