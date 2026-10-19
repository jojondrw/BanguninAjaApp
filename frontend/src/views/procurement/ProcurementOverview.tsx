import { ArrowRight } from 'lucide-react'

import { useProcurementOverview } from '../../controllers/useProcurement'
import {
  VENDOR_RATING_LABEL,
  type PurchaseOrderStatus,
  type PurchaseRequestStatus,
  type VendorRating,
} from '../../models/procurement'
import { number, rupiahShort } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { DonutChart } from '../components/DonutChart'
import { LinkButton } from '../components/LinkButton'
import { Money } from '../components/Money'

const RATING_ORDER: VendorRating[] = ['good', 'fair', 'new', 'poor']
const RATING_COLORS = ['#16a34a', '#fbbb16', '#94a3b8', '#dc2626']

export interface ProcurementDrill {
  openRequests: (status: PurchaseRequestStatus | '') => void
  openOrders: (status: PurchaseOrderStatus | '') => void
  openReceipts: () => void
  openVendors: () => void
}

function FlowStep({ label, value, note, onClick }: { label: string; value: number; note: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="min-w-0 flex-1 rounded-xl bg-slate-50 px-3.5 py-3 text-left transition-colors hover:bg-slate-100"
    >
      <span className="block text-xs text-slate-500">{label}</span>
      <span className="mt-0.5 block text-2xl font-semibold tracking-[-0.02em] text-slate-900 tabular-nums">{number(value)}</span>
      <span className="block truncate text-xs text-slate-500">{note}</span>
    </button>
  )
}

function Arrow() {
  return <ArrowRight aria-hidden="true" className="hidden size-4 shrink-0 self-center text-slate-300 sm:block" strokeWidth={2} />
}

function AttentionTile({ label, value, note, tone, onClick }: {
  label: string
  value: string
  note: string
  tone: 'red' | 'amber' | 'slate'
  onClick: () => void
}) {
  const tones = {
    red: 'bg-red-50 text-red-700 hover:bg-red-100/70',
    amber: 'bg-amber-50 text-amber-800 hover:bg-amber-100/70',
    slate: 'bg-slate-50 text-slate-700 hover:bg-slate-100',
  }
  return (
    <button type="button" onClick={onClick} className={`rounded-xl px-3.5 py-3 text-left transition-colors ${tones[tone]}`}>
      <span className="block text-xs">{label}</span>
      <span className="mt-0.5 block text-xl font-semibold tracking-[-0.02em] tabular-nums">{value}</span>
      <span className="block text-xs opacity-80">{note}</span>
    </button>
  )
}

// Ringkasan pengadaan: alur permintaan sampai selesai, hal yang perlu
// ditindaklanjuti, dan kesehatan vendor. Setiap angka membuka tab rinciannya.
export function ProcurementOverview({ drill }: { drill: ProcurementDrill }) {
  const overview = useProcurementOverview()

  if (overview.isPending) {
    return <Loading />
  }
  if (overview.isError) {
    return <LoadFailed onRetry={() => overview.refetch()} />
  }
  const data = overview.data
  const requests = (status: PurchaseRequestStatus) => data.requests[status] ?? 0
  const orders = data.orders.byStatus
  const damagedShare = data.recentReceipts > 0 ? Math.round((data.damagedRecent / data.recentReceipts) * 100) : 0

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
      <div className="grid content-start gap-6">
        <Card title="Alur pengadaan" description="Dari permintaan proyek sampai barang diterima lengkap">
          <div className="flex flex-col gap-2 sm:flex-row">
            <FlowStep
              label="Permintaan menunggu"
              value={requests('submitted')}
              note={`${number(requests('draft'))} masih draf`}
              onClick={() => drill.openRequests('submitted')}
            />
            <Arrow />
            <FlowStep
              label="Disetujui"
              value={requests('approved')}
              note="siap dibuat pesanan"
              onClick={() => drill.openRequests('approved')}
            />
            <Arrow />
            <FlowStep
              label="Pesanan berjalan"
              value={orders.sent + orders.partially_received}
              note={`${number(orders.partially_received)} diterima sebagian`}
              onClick={() => drill.openOrders('sent')}
            />
            <Arrow />
            <FlowStep
              label="Selesai"
              value={orders.completed}
              note="barang diterima lengkap"
              onClick={() => drill.openOrders('completed')}
            />
          </div>
          <p className="mt-3 text-xs text-slate-500">
            Nilai pesanan terbuka <Money>{rupiahShort(data.orders.openValue)}</Money> dari total{' '}
            <Money>{rupiahShort(data.orders.totalValue)}</Money>.
          </p>
        </Card>

        <Card title="Perlu ditindaklanjuti">
          <div className="grid gap-2 sm:grid-cols-3">
            <AttentionTile
              label="Pesanan terlambat"
              value={number(data.lateOrders)}
              note={`${rupiahShort(data.lateValue)} lewat jatuh tempo`}
              tone="red"
              onClick={() => drill.openOrders('')}
            />
            <AttentionTile
              label="Penerimaan bermasalah"
              value={number(data.damagedRecent)}
              note={`${damagedShare}% dari ${number(data.recentReceipts)} penerimaan ${data.receiptDays} hari terakhir`}
              tone="amber"
              onClick={drill.openReceipts}
            />
            <AttentionTile
              label="Permintaan menunggu"
              value={number(requests('submitted'))}
              note="perlu disetujui atau ditolak"
              tone="slate"
              onClick={() => drill.openRequests('submitted')}
            />
          </div>
        </Card>
      </div>

      <Card title="Vendor" action={<LinkButton label="Semua vendor" onClick={drill.openVendors} />}>
        <div className="grid gap-6 sm:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)] xl:grid-cols-1">
          <div>
            <p className="mb-3 text-xs font-medium text-slate-500">Vendor aktif menurut penilaian</p>
            <DonutChart
              caption="Vendor aktif menurut penilaian"
              colors={RATING_COLORS}
              slices={RATING_ORDER.map((rating) => ({ label: VENDOR_RATING_LABEL[rating], value: data.vendorRatings[rating] ?? 0 }))}
              format={(value) => `${number(value)} vendor`}
            />
          </div>
          <div>
            <p className="mb-2 text-xs font-medium text-slate-500">Nilai pesanan terbesar</p>
            <ul className="divide-y divide-slate-100">
              {data.largestVendors.map((vendor) => (
                <li key={vendor.name} className="flex items-center justify-between gap-3 py-2 text-[13px]">
                  <span className="min-w-0">
                    <span className="block truncate text-slate-900">{vendor.name}</span>
                    <span className="text-xs text-slate-500">
                      {number(vendor.orders)} pesanan, {VENDOR_RATING_LABEL[vendor.rating].toLowerCase()}
                    </span>
                  </span>
                  <span className="shrink-0 text-slate-900 tabular-nums">
                    <Money>{rupiahShort(vendor.value)}</Money>
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </Card>
    </div>
  )
}
