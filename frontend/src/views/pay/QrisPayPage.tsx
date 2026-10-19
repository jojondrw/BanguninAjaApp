import { CircleCheck, FlaskConical } from 'lucide-react'
import { type ReactNode, useEffect } from 'react'
import { useParams } from 'react-router-dom'

import { useQrisPayment, useSimulateQrisPayment } from '../../controllers/useGateway'
import { providerLabel, type GatewayPublicPayment } from '../../models/gateway'
import { ONCHAIN_TARGET_LABEL, timeLabel } from '../../models/onchain'
import { ApiError } from '../../shared/apiClient'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { BrandName } from '../components/Brand'
import { Loading } from '../components/Data'
import { Button, ErrorNote } from '../components/Form'
import { Money } from '../components/Money'
import { OnchainStatusChip } from './PayParts'
import { QrImage } from './QrImage'

function isRejection(error: unknown): boolean {
  return error instanceof ApiError && error.status > 0 && error.status < 500 && error.status !== 429
}

function QrisLayout({ children }: { children: ReactNode }) {
  useEffect(() => {
    document.title = 'Bayar dengan QRIS · BanguninAja'
  }, [])

  return (
    <div className="min-h-screen bg-slate-100">
      <header className="mx-auto flex max-w-md items-center px-4 pt-5 sm:px-6">
        <BrandName />
      </header>
      <main className="mx-auto max-w-md px-4 py-6 sm:px-6">
        <div className="rounded-2xl bg-white p-5 shadow-hairline sm:p-6 motion-safe:animate-enter">{children}</div>
        <p className="mt-4 text-center text-xs text-slate-500">
          Pastikan nama penerima dan nominal di aplikasi Anda sama sebelum membayar. BanguninAja tidak pernah meminta PIN
          atau kode OTP.
        </p>
      </main>
    </div>
  )
}

function SimulateButton({ token }: { token: string }) {
  const simulate = useSimulateQrisPayment(token)
  return (
    <div className="space-y-2 rounded-xl bg-amber-50 px-3.5 py-3">
      <p className="flex items-start gap-2 text-[13px] text-amber-900">
        <FlaskConical aria-hidden="true" className="mt-0.5 size-4 shrink-0" strokeWidth={1.8} />
        Mode uji: tidak ada uang sungguhan. Tombol ini meminta gerbang pembayaran berpura-pura QRIS sudah dibayar.
      </p>
      <Button variant="subtle" isPending={simulate.isPending} pendingLabel="Mensimulasikan" onClick={() => simulate.mutate()}>
        Simulasikan pembayaran
      </Button>
      {simulate.isError ? <ErrorNote message={errorMessage(simulate.error)} /> : null}
    </div>
  )
}

function Waiting({ token, payment }: { token: string; payment: GatewayPublicPayment }) {
  return (
    <div className="space-y-4">
      <QrImage value={payment.qrString} label="Kode QRIS" className="mx-auto size-60" />
      <ol className="list-decimal space-y-1 pl-5 text-[13px] text-slate-600">
        <li>Buka aplikasi bank atau dompet digital, pilih Scan QR atau Bayar.</li>
        <li>Pindai kode di atas dan pastikan nominalnya sama.</li>
        <li>Setujui pembayaran. Halaman ini berubah sendiri begitu dananya diterima.</li>
      </ol>
      <p className="text-xs text-slate-500">Berlaku sampai {timeLabel(payment.expiresAt)}.</p>
      {payment.testMode ? <SimulateButton token={token} /> : null}
    </div>
  )
}

function Outcome({ payment }: { payment: GatewayPublicPayment }) {
  if (payment.status === 'paid') {
    return (
      <div className="flex items-start gap-3 rounded-xl bg-green-50 px-4 py-3.5 text-green-800">
        <CircleCheck aria-hidden="true" className="mt-0.5 size-5 shrink-0" strokeWidth={1.8} />
        <div className="text-[13px]">
          <p className="font-semibold">Pembayaran diterima</p>
          <p>
            Sudah tercatat di tagihan {payment.paidAt ? `pada ${timeLabel(payment.paidAt)}` : ''}. Halaman ini boleh
            ditutup.
          </p>
        </div>
      </div>
    )
  }
  const reasons: Record<string, string> = {
    expired: 'Kode QRIS ini sudah kedaluwarsa. Minta kode baru ke bagian keuangan.',
    cancelled: 'Kode QRIS ini sudah dibatalkan. Minta kode baru ke bagian keuangan.',
    failed: payment.failureReason || 'Pembayaran tidak bisa dicatat. Hubungi bagian keuangan.',
  }
  return <ErrorNote message={reasons[payment.status] ?? 'Kode ini tidak bisa dibayar lagi.'} />
}

// Halaman publik /bayar/qris/:token. Statusnya ditanyakan ke server tiap tiga
// detik; server sendiri yang menanyakan ke gerbang pembayaran, jadi pembeli
// tidak perlu menekan "sudah bayar".
export function QrisPayPage() {
  const token = useParams().token ?? ''
  const payment = useQrisPayment(token)

  return (
    <QrisLayout>
      {payment.isPending ? <Loading label="Mengambil data tagihan..." /> : null}
      {payment.isError ? (
        <div className="space-y-3">
          <h1 className="text-lg font-semibold text-slate-900">Halaman bayar tidak bisa dibuka</h1>
          <ErrorNote message={errorMessage(payment.error)} />
          {isRejection(payment.error) ? null : (
            <Button variant="subtle" onClick={() => payment.refetch()}>
              Coba lagi
            </Button>
          )}
        </div>
      ) : null}
      {payment.data ? (
        <div className="space-y-5">
          <div>
            <div className="flex items-center justify-between gap-3">
              <p className="text-xs font-medium text-slate-500">
                Bayar {ONCHAIN_TARGET_LABEL[payment.data.targetKind]} {payment.data.targetNumber}
              </p>
              <OnchainStatusChip status={payment.data.status} />
            </div>
            <h1 className="mt-1 text-3xl font-semibold tracking-[-0.03em] text-slate-900 tabular-nums">
              <Money>{rupiah(payment.data.amountIdr)}</Money>
            </h1>
            <p className="mt-0.5 text-[13px] text-slate-500">
              {payment.data.projectName ? `${payment.data.projectName}, ` : ''}QRIS lewat {providerLabel(payment.data.provider)}
            </p>
          </div>
          {payment.data.status === 'pending' ? <Waiting token={token} payment={payment.data} /> : <Outcome payment={payment.data} />}
        </div>
      ) : null}
    </QrisLayout>
  )
}
