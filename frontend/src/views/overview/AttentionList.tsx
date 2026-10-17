import { AlarmClock, CalendarClock, ChevronRight, ClipboardCheck, PackageMinus, Wallet, type LucideIcon } from 'lucide-react'
import { Link } from 'react-router-dom'

import { canAccess, usePermissions } from '../../controllers/useAccess'
import { usePayrollTotals } from '../../controllers/useHr'
import { useLowStockMaterials } from '../../controllers/useInventory'
import { usePurchaseRequestCount } from '../../controllers/useProcurement'
import { useInstallmentCount } from '../../controllers/useSales'
import { number, rupiahShort } from '../../shared/format'
import { currentPeriod } from '../../shared/localDate'
import { moduleForPath } from '../components/navigation'
import { MoneyText } from '../components/Money'

type Severity = 'urgent' | 'pending'

interface Attention {
  key: string
  icon: LucideIcon
  to: string
  label: string
  count: number | null | undefined
  detail: string
  severity: Severity
}

const SALES_PATH = '/penjualan'
const PROCUREMENT_PATH = '/pengadaan'
const INVENTORY_PATH = '/inventaris'
const HR_PATH = '/sdm'

const SEVERITY_TONE: Record<Severity, string> = {
  urgent: 'bg-red-50 text-red-600',
  pending: 'bg-amber-50 text-amber-700',
}

export function AttentionList() {
  const permissions = usePermissions()
  // Hanya hal dari modul yang boleh dibaca peran ini, dan permintaannya pun
  // tidak dikirim untuk modul lain supaya tidak ada 403 di latar belakang.
  const canRead = (path: string) => {
    const module = moduleForPath(path)
    return module === null || canAccess(permissions, module, 'read')
  }
  const overdue = useInstallmentCount('overdue', { enabled: canRead(SALES_PATH) })
  const due = useInstallmentCount('due', { enabled: canRead(SALES_PATH) })
  const requests = usePurchaseRequestCount('submitted', { enabled: canRead(PROCUREMENT_PATH) })
  const lowStock = useLowStockMaterials(1, { enabled: canRead(INVENTORY_PATH) })
  const payroll = usePayrollTotals(currentPeriod(), { enabled: canRead(HR_PATH) })

  const allItems: Attention[] = [
    {
      key: 'overdue',
      icon: AlarmClock,
      to: SALES_PATH,
      label: 'Cicilan terlambat',
      count: countOf(overdue.data, overdue.isError),
      detail: 'Pembeli melewati jatuh tempo',
      severity: 'urgent',
    },
    {
      key: 'due',
      icon: CalendarClock,
      to: SALES_PATH,
      label: 'Cicilan jatuh tempo',
      count: countOf(due.data, due.isError),
      detail: 'Perlu ditagih sekarang',
      severity: 'pending',
    },
    {
      key: 'requests',
      icon: ClipboardCheck,
      to: PROCUREMENT_PATH,
      label: 'Permintaan pembelian',
      count: countOf(requests.data, requests.isError),
      detail: 'Menunggu persetujuan',
      severity: 'pending',
    },
    {
      key: 'stock',
      icon: PackageMinus,
      to: INVENTORY_PATH,
      label: 'Stok menipis',
      count: countOf(lowStock.data?.totalItems, lowStock.isError),
      detail: 'Material di bawah stok minimum',
      severity: 'urgent',
    },
    {
      key: 'payroll',
      icon: Wallet,
      to: HR_PATH,
      label: 'Gaji belum dibayar',
      count: countOf(payroll.data?.unpaidCount, payroll.isError),
      detail: payroll.data ? `${rupiahShort(payroll.data.unpaidNetPay)} bulan ini` : 'Slip gaji bulan ini',
      severity: 'pending',
    },
  ]
  const items = allItems.filter((item) => canRead(item.to))

  const open = items.filter((item) => (item.count ?? 0) > 0)
  const isLoading = items.some((item) => item.count === undefined)

  return (
    <div>
      {!isLoading && open.length === 0 ? (
        <p role="status" className="rounded-xl bg-green-50 px-4 py-3 text-[13px] text-green-700">
          Semua beres. Tidak ada tagihan terlambat, permintaan tertunda, atau stok menipis.
        </p>
      ) : null}

      <ul className="-mx-2 flex flex-col">
        {items.map((item) => (
          <li key={item.key}>
            <AttentionRow item={item} />
          </li>
        ))}
      </ul>
    </div>
  )
}

function AttentionRow({ item }: { item: Attention }) {
  const Icon = item.icon
  const hasItems = (item.count ?? 0) > 0

  return (
    <Link
      to={item.to}
      className="group flex items-center gap-3 rounded-xl px-2 py-2.5 transition-[background-color,transform] hover:bg-slate-50
                 motion-safe:active:scale-[0.99]"
    >
      <span
        aria-hidden="true"
        className={`grid size-8 shrink-0 place-items-center rounded-lg ${
          hasItems ? SEVERITY_TONE[item.severity] : 'bg-slate-100 text-slate-400'
        }`}
      >
        <Icon className="size-4" strokeWidth={1.8} />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block text-[13px] font-medium text-slate-900">{item.label}</span>
        <span className="block truncate text-xs text-slate-500"><MoneyText>{item.detail}</MoneyText></span>
      </span>
      <span
        className={`text-[15px] font-semibold tabular-nums ${hasItems ? 'text-slate-900' : 'text-slate-400'}`}
        aria-label={countLabel(item.count)}
      >
        {countText(item.count)}
      </span>
      <ChevronRight
        aria-hidden="true"
        className="size-4 text-slate-300 transition-transform group-hover:translate-x-0.5 group-hover:text-slate-500"
        strokeWidth={1.8}
      />
    </Link>
  )
}

function countOf(value: number | undefined, isError: boolean): number | null | undefined {
  return isError ? null : value
}

function countText(count: number | null | undefined): string {
  if (count === undefined) {
    return '…'
  }
  return count === null ? '—' : number(count)
}

function countLabel(count: number | null | undefined): string {
  if (count === undefined) {
    return 'memuat'
  }
  return count === null ? 'gagal dimuat' : `${count} item`
}
