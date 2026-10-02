import { useBillingTotals } from '../../controllers/useBilling'
import { number, rupiahShort } from '../../shared/format'
import { kpiText } from '../../shared/kpiText'
import { Kpi, KpiRow } from '../components/Data'

const SCOPE_LIMIT = 100

function scopeNote(isPartial: boolean): string {
  return isPartial ? `, dari ${SCOPE_LIMIT} catatan pertama` : ''
}

// Backend belum punya endpoint ringkasan, jadi angka di sini dihitung dari
// daftar faktur, piutang, dan utang masing masing.
export function BillingKpis() {
  const invoices = useBillingTotals('invoices')
  const receivables = useBillingTotals('receivables')
  const payables = useBillingTotals('payables')

  const loaded =
    invoices.data && receivables.data && payables.data ? [invoices.data, receivables.data, payables.data] : undefined
  const urgent = {
    isPending: invoices.isPending || receivables.isPending || payables.isPending,
    isError: invoices.isError || receivables.isError || payables.isError,
    data: loaded,
  }
  const overdueCount = loaded?.reduce((sum, totals) => sum + totals.overdueCount, 0) ?? 0
  const overdueValue = loaded?.reduce((sum, totals) => sum + totals.overdueOutstanding, 0) ?? 0
  const urgentPartial = loaded?.some((totals) => totals.isPartial) ?? false

  return (
    <KpiRow>
      <Kpi
        label="Piutang belum diterima"
        value={kpiText(receivables, (data) => rupiahShort(data.outstanding))}
        note={
          receivables.data
            ? `${number(receivables.data.openCount)} piutang terbuka${scopeNote(receivables.data.isPartial)}`
            : undefined
        }
      />
      <Kpi
        label="Utang belum dibayar"
        value={kpiText(payables, (data) => rupiahShort(data.outstanding))}
        note={
          payables.data
            ? `${number(payables.data.openCount)} utang ke vendor${scopeNote(payables.data.isPartial)}`
            : undefined
        }
      />
      <Kpi
        label="Jatuh tempo dan terlambat"
        value={kpiText(urgent, (data) =>
          number(data.reduce((sum, totals) => sum + totals.overdueCount + totals.dueSoonCount, 0)),
        )}
        note={
          loaded
            ? `${number(overdueCount)} terlambat (${rupiahShort(overdueValue)}) dari faktur, piutang, dan utang${scopeNote(urgentPartial)}`
            : undefined
        }
      />
      <Kpi
        label="Piutang sudah diterima"
        value={kpiText(receivables, (data) => rupiahShort(data.paidAmount))}
        note="total semua waktu, server belum mencatat tanggal tiap pembayaran"
      />
    </KpiRow>
  )
}
