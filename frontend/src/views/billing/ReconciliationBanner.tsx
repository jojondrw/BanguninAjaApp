import { TriangleAlert } from 'lucide-react'
import { useState } from 'react'

import { useBillingReconciliation } from '../../controllers/useBilling'
import { number, rupiah } from '../../shared/format'
import { Pager } from '../components/RecordControls'

const PAGE_SIZE = 10

// Peringatan baca-saja untuk faktur dan piutang tertaut yang total bayarnya
// berbeda. Data ini berasal dari sebelum riwayat pembayaran bersama ada, jadi
// tidak bisa dibetulkan otomatis. Tidak tampil apa pun kalau semua cocok atau
// pemeriksaannya gagal, supaya halaman tagihan tetap bisa dipakai.
export function ReconciliationBanner() {
  const [page, setPage] = useState(1)
  const [isOpen, setOpen] = useState(false)
  const pairs = useBillingReconciliation(page, PAGE_SIZE)
  const total = pairs.data?.totalItems ?? 0

  if (!pairs.data || total === 0) {
    return null
  }

  const listId = 'billing-reconciliation-list'

  return (
    <section
      aria-labelledby="billing-reconciliation-title"
      className="mt-6 rounded-2xl bg-amber-50 px-4 py-3 text-[13px] text-amber-900"
    >
      <div className="flex flex-wrap items-start gap-3">
        <TriangleAlert aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-amber-700" strokeWidth={1.8} />
        <div className="min-w-0 flex-1">
          <h2 id="billing-reconciliation-title" className="font-semibold">
            {number(total)} pasangan faktur dan piutang punya total bayar berbeda
          </h2>
          <p className="mt-0.5 text-amber-800">
            Data ini tercatat sebelum riwayat pembayaran bersama dipakai. Cocokkan pembayarannya secara manual di
            rincian faktur atau piutang.
          </p>
        </div>
        <button
          type="button"
          aria-expanded={isOpen}
          aria-controls={listId}
          onClick={() => setOpen((open) => !open)}
          className="h-8 shrink-0 rounded-lg px-3 font-medium text-amber-900 hover:bg-amber-100"
        >
          {isOpen ? 'Sembunyikan daftar' : 'Lihat daftar'}
        </button>
      </div>

      {isOpen ? (
        <div id={listId} className="relative mt-3 overflow-x-auto rounded-xl bg-white">
          <table className="w-full text-[13px] text-slate-700">
            <thead>
              <tr className="text-left text-xs text-slate-500">
                <th scope="col" className="px-3 py-2 font-medium">Faktur</th>
                <th scope="col" className="px-3 py-2 font-medium">Piutang</th>
                <th scope="col" className="px-3 py-2 font-medium">Pelanggan</th>
                <th scope="col" className="px-3 py-2 text-right font-medium">Dibayar di faktur</th>
                <th scope="col" className="px-3 py-2 text-right font-medium">Dibayar di piutang</th>
                <th scope="col" className="px-3 py-2 text-right font-medium">Selisih</th>
              </tr>
            </thead>
            <tbody>
              {pairs.data.items.map((pair) => (
                <tr key={pair.invoiceId} className="border-t border-slate-100">
                  <td className="px-3 py-2 font-medium text-slate-900">{pair.invoiceNumber}</td>
                  <td className="px-3 py-2">{pair.receivableReference}</td>
                  <td className="px-3 py-2">{pair.customerName ?? '-'}</td>
                  <td className="px-3 py-2 text-right tabular-nums">{rupiah(pair.invoicePaidAmount)}</td>
                  <td className="px-3 py-2 text-right tabular-nums">{rupiah(pair.receivablePaidAmount)}</td>
                  <td className="px-3 py-2 text-right font-medium tabular-nums text-amber-800">
                    {rupiah(Math.abs(pair.difference))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="px-3 pb-3">
            <Pager
              page={pairs.data.page}
              totalPages={pairs.data.totalPages}
              totalItems={pairs.data.totalItems}
              unit="pasangan"
              onChange={setPage}
            />
          </div>
        </div>
      ) : null}
    </section>
  )
}
