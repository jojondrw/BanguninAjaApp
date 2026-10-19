import { useAssetOverview, useInventoryOverview } from '../../controllers/useInventory'
import { EQUIPMENT_STATUSES, EQUIPMENT_STATUS_LABEL, EQUIPMENT_STATUS_TONE } from '../../models/inventory'
import { number, rupiahShort } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { DonutChart } from '../components/DonutChart'
import { LinkButton } from '../components/LinkButton'
import { Money } from '../components/Money'

const WAREHOUSE_COLORS = ['#0b2b6b', '#1f5fd1', '#5b86d9', '#94a3b8', '#cbd5e1', '#e2e8f0', '#f1f5f9']
const CATEGORY_COLORS = ['#0f766e', '#14b8a6', '#5eead4', '#fbbb16', '#cbd5e1', '#e2e8f0']
const TOP_SLICES = 5

export interface InventoryDrill {
  openStock: () => void
  openAssets: () => void
}

function topSlices<T>(rows: T[], label: (row: T) => string, value: (row: T) => number) {
  const sorted = [...rows].sort((a, b) => value(b) - value(a))
  const shown = sorted.slice(0, TOP_SLICES).map((row) => ({ label: label(row), value: value(row) }))
  const rest = sorted.slice(TOP_SLICES).reduce((sum, row) => sum + value(row), 0)
  return rest > 0 ? [...shown, { label: 'Lainnya', value: rest }] : shown
}

function WeeklyMovements({ weeks }: { weeks: { week: string; inValue: number; outValue: number }[] }) {
  const highest = Math.max(1, ...weeks.flatMap((week) => [week.inValue, week.outValue]))
  if (weeks.length === 0) {
    return <p className="text-[13px] text-slate-500">Belum ada mutasi dalam delapan minggu terakhir.</p>
  }
  return (
    <div>
      <div className="mb-3 flex gap-4 text-xs text-slate-500">
        <span className="inline-flex items-center gap-1.5">
          <span aria-hidden="true" className="size-2 rounded-full bg-green-500" />
          Masuk
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span aria-hidden="true" className="size-2 rounded-full bg-slate-400" />
          Keluar
        </span>
      </div>
      <ul className="flex h-40 items-end gap-2" aria-label="Nilai mutasi stok per minggu">
        {weeks.map((week) => (
          <li key={week.week} className="flex h-full min-w-0 flex-1 flex-col items-center justify-end gap-1">
            <span className="flex h-full w-full items-end justify-center gap-0.5" title={`Masuk ${rupiahShort(week.inValue)}, keluar ${rupiahShort(week.outValue)}`}>
              <span className="w-1/3 max-w-3 rounded-t-[3px] bg-green-500" style={{ height: `${(week.inValue / highest) * 100}%` }} />
              <span className="w-1/3 max-w-3 rounded-t-[3px] bg-slate-400" style={{ height: `${(week.outValue / highest) * 100}%` }} />
            </span>
            <span className="text-[10px] text-slate-400 tabular-nums">{week.week.slice(8, 10)}/{week.week.slice(5, 7)}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function Tile({ label, value, note, tone, onClick }: {
  label: string
  value: number
  note: string
  tone: string
  onClick: () => void
}) {
  return (
    <button type="button" onClick={onClick} className={`rounded-xl px-3.5 py-3 text-left transition-colors ${tone}`}>
      <span className="block text-xs">{label}</span>
      <span className="mt-0.5 block text-xl font-semibold tracking-[-0.02em] tabular-nums">{number(value)}</span>
      <span className="block text-xs opacity-80">{note}</span>
    </button>
  )
}

// Ringkasan inventaris: nilai stok, arus barang, hal mendesak, dan aset.
export function InventoryOverview({ drill }: { drill: InventoryDrill }) {
  const stock = useInventoryOverview()
  const assets = useAssetOverview()

  if (stock.isPending || assets.isPending) {
    return <Loading />
  }
  if (stock.isError || assets.isError) {
    return (
      <LoadFailed
        onRetry={() => {
          void stock.refetch()
          void assets.refetch()
        }}
      />
    )
  }
  const inventory = stock.data
  const equipment = assets.data

  return (
    <div className="grid gap-6 xl:grid-cols-2">
      <Card
        title="Nilai stok per gudang"
        description="Jumlah barang dikali harga terakhir material"
        action={<LinkButton label="Stok" onClick={drill.openStock} />}
      >
        <DonutChart
          caption="Nilai stok per gudang"
          colors={WAREHOUSE_COLORS}
          slices={topSlices(inventory.warehouses, (row) => row.name, (row) => row.value)}
        />
      </Card>

      <div className="grid gap-6">
        <Card title="Perlu ditindaklanjuti">
          <div className="grid gap-2 sm:grid-cols-3">
            <Tile
              label="Stok menipis"
              value={inventory.lowStock}
              note="di bawah stok minimum"
              tone="bg-amber-50 text-amber-800 hover:bg-amber-100/70"
              onClick={drill.openStock}
            />
            <Tile
              label="Servis terlambat"
              value={equipment.serviceOverdue}
              note="alat lewat jadwal servis"
              tone="bg-red-50 text-red-700 hover:bg-red-100/70"
              onClick={drill.openAssets}
            />
            <Tile
              label="Servis segera"
              value={equipment.serviceSoon}
              note={`dalam ${equipment.serviceWindow} hari`}
              tone="bg-slate-50 text-slate-700 hover:bg-slate-100"
              onClick={drill.openAssets}
            />
          </div>
        </Card>
        <Card title="Mutasi 8 minggu" description="Nilai barang masuk dan keluar gudang per minggu">
          <WeeklyMovements weeks={inventory.weeks} />
        </Card>
      </div>

      <Card
        title="Aset"
        description={`Nilai buku ${rupiahShort(equipment.bookValue)} di luar aset yang dilepas`}
        action={<LinkButton label="Aset dan alat" onClick={drill.openAssets} />}
      >
        <DonutChart
          caption="Nilai buku aset per kategori"
          colors={CATEGORY_COLORS}
          slices={topSlices(equipment.categories, (row) => row.category, (row) => row.bookValue)}
        />
      </Card>

      <Card title="Alat menurut status" action={<LinkButton label="Semua alat" onClick={drill.openAssets} />}>
        <ul className="grid grid-cols-2 gap-2">
          {EQUIPMENT_STATUSES.map((status) => (
            <li key={status}>
              <button
                type="button"
                onClick={drill.openAssets}
                className="w-full rounded-xl bg-slate-50 px-3.5 py-3 text-left hover:bg-slate-100"
              >
                <span className={`inline-block rounded-full px-2 py-0.5 text-[11px] font-medium ${EQUIPMENT_STATUS_TONE[status]}`}>
                  {EQUIPMENT_STATUS_LABEL[status]}
                </span>
                <span className="mt-1 block text-xl font-semibold text-slate-900 tabular-nums">
                  {number(equipment.equipmentStatus[status] ?? 0)}
                </span>
              </button>
            </li>
          ))}
        </ul>
        <p className="mt-3 text-xs text-slate-500">
          Total stok <Money>{rupiahShort(inventory.stockValue)}</Money> di {number(inventory.warehouses.length)} gudang.
        </p>
      </Card>
    </div>
  )
}
