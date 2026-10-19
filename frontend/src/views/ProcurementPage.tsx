import { useMemo, useState } from 'react'

import {
  useActiveVendorCount,
  usePurchaseOrderSummary,
  usePurchaseRequestCount,
} from '../controllers/useProcurement'
import type { PurchaseOrderFormValues, PurchaseOrderStatus, PurchaseRequestStatus } from '../models/procurement'
import { number, rupiahShort } from '../shared/format'
import { kpiText } from '../shared/kpiText'
import { AppShell } from './components/AppShell'
import { Kpi, KpiRow } from './components/Data'
import { SectionTabs } from './components/ListTools'
import type { OrderPrefill } from './procurement/lookup'
import { OrdersSection } from './procurement/OrdersSection'
import { ReceiptsSection } from './procurement/ReceiptsSection'
import { RequestsSection } from './procurement/RequestsSection'
import { ProcurementOverview, type ProcurementDrill } from './procurement/ProcurementOverview'
import { VendorsSection } from './procurement/VendorsSection'

type ProcurementTab = 'overview' | 'orders' | 'receipts' | 'vendors' | 'requests'

const TABS: { value: ProcurementTab; label: string }[] = [
  { value: 'overview', label: 'Ringkasan' },
  { value: 'orders', label: 'Pesanan pembelian' },
  { value: 'receipts', label: 'Penerimaan barang' },
  { value: 'vendors', label: 'Vendor' },
  { value: 'requests', label: 'Permintaan pembelian' },
]

function ProcurementKpis() {
  const orders = usePurchaseOrderSummary()
  const activeVendors = useActiveVendorCount()
  const pendingRequests = usePurchaseRequestCount('submitted')

  return (
    <KpiRow>
      <Kpi
        label="PO terbuka"
        value={kpiText(orders, (summary) => number(summary.openCount))}
        note={orders.data ? 'draf, dikirim, atau diterima sebagian' : undefined}
      />
      <Kpi
        label="Nilai PO terbuka"
        value={kpiText(orders, (summary) => rupiahShort(summary.openValue))}
        note={orders.data ? `total ${rupiahShort(orders.data.totalValue)} di luar yang batal` : undefined}
      />
      <Kpi label="Vendor aktif" value={kpiText(activeVendors, number)} note="bisa menerima pesanan baru" />
      <Kpi label="Permintaan menunggu" value={kpiText(pendingRequests, number)} note="diajukan, belum disetujui" />
    </KpiRow>
  )
}

export function ProcurementPage() {
  const [tab, setTab] = useState<ProcurementTab>('overview')
  const [orderPrefill, setOrderPrefill] = useState<OrderPrefill | null>(null)
  const [orderStatus, setOrderStatus] = useState<PurchaseOrderStatus | ''>('')
  const [requestStatus, setRequestStatus] = useState<PurchaseRequestStatus | ''>('')
  const [visit, setVisit] = useState(0)

  const drill = useMemo<ProcurementDrill>(() => {
    const go = (next: ProcurementTab) => {
      setOrderPrefill(null)
      setVisit((current) => current + 1)
      setTab(next)
    }
    return {
      openRequests: (status) => {
        setRequestStatus(status)
        go('requests')
      },
      openOrders: (status) => {
        setOrderStatus(status)
        go('orders')
      },
      openReceipts: () => go('receipts'),
      openVendors: () => go('vendors'),
    }
  }, [])

  // "Buat PO dari permintaan" pindah ke tab pesanan dengan formulir yang sudah
  // terisi dari permintaan itu. Isian itu dibuang begitu pengguna pindah tab,
  // jadi kembali ke tab pesanan membuka formulir kosong.
  const createOrderFromRequest = (values: PurchaseOrderFormValues, requestNumber: string) => {
    setOrderPrefill({ values, requestNumber })
    setTab('orders')
  }

  const changeTab = (next: ProcurementTab) => {
    if (next !== tab) {
      setOrderPrefill(null)
      setOrderStatus('')
      setRequestStatus('')
    }
    setTab(next)
  }

  return (
    <AppShell
      title="Pengadaan"
      description="Vendor, permintaan, pesanan pembelian, dan penerimaan barang material proyek"
      tabs={<SectionTabs tabs={TABS} active={tab} onChange={changeTab} label="Bagian pengadaan" />}
    >
      <ProcurementKpis />

      <div className="mb-6" />

      {tab === 'overview' ? <ProcurementOverview drill={drill} /> : null}
      {tab === 'orders' ? <OrdersSection key={`pesanan-${visit}`} prefill={orderPrefill} initialStatus={orderStatus} /> : null}
      {tab === 'receipts' ? <ReceiptsSection /> : null}
      {tab === 'vendors' ? <VendorsSection /> : null}
      {tab === 'requests' ? (
        <RequestsSection key={`permintaan-${visit}`} onCreateOrder={createOrderFromRequest} initialStatus={requestStatus} />
      ) : null}
    </AppShell>
  )
}
