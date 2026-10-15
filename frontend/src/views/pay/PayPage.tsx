import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Wallet } from 'lucide-react'
import { type FormEvent, type ReactNode, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'

import { usePaymentConfirmation, usePublicPayment } from '../../controllers/useOnchain'
import {
  ONCHAIN_TARGET_LABEL,
  currencyName,
  networkName,
  normalizeTxHash,
  timeLabel,
  type ConfirmOutcome,
  type PublicPayment,
} from '../../models/onchain'
import { ApiError } from '../../shared/apiClient'
import { errorMessage } from '../../shared/errorMessage'
import { BrandName } from '../components/Brand'
import { Loading } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { AmountSummary, CopyButton, OnchainStatusChip, TestnetNote, TxLink } from './PayParts'
import { QrImage } from './QrImage'
import { readSubmittedHash, storeSubmittedHash } from './paySession'
import { injectedWallet, metamaskDeepLink, payWithWallet, walletErrorMessage } from './wallet'

const MINUTE_MS = 60_000

function useNow(intervalMs: number): number {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), intervalMs)
    return () => window.clearInterval(timer)
  }, [intervalMs])
  return now
}

function explorerTxUrl(payment: PublicPayment, hash: string): string {
  return `${payment.explorerUrl}/tx/${hash}`
}

function isRejection(error: unknown): boolean {
  return error instanceof ApiError && error.status > 0 && error.status < 500 && error.status !== 429
}

function PayLayout({ children }: { children: ReactNode }) {
  useEffect(() => {
    document.title = 'Bayar tagihan · BanguninAja'
  }, [])

  return (
    <div className="min-h-screen bg-slate-100">
      <header className="mx-auto flex max-w-xl items-center px-4 pt-5 sm:px-6">
        <BrandName />
      </header>
      <main className="mx-auto max-w-xl px-4 py-6 sm:px-6">
        <div className="rounded-2xl bg-white p-5 shadow-hairline sm:p-6 motion-safe:animate-enter">{children}</div>
        <p className="mt-4 text-center text-xs text-slate-500">
          Halaman ini hanya meminta dompet Anda mengirim transaksi. BanguninAja tidak pernah meminta kunci privat atau frasa
          pemulihan.
        </p>
      </main>
    </div>
  )
}

function ManualHashForm({ isChecking, onSubmit, intro }: {
  isChecking: boolean
  onSubmit: (hash: string) => void
  intro: string
}) {
  const [raw, setRaw] = useState('')
  const [invalid, setInvalid] = useState(false)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const hash = normalizeTxHash(raw)
    setInvalid(hash === null)
    if (hash !== null) {
      onSubmit(hash)
    }
  }

  return (
    <form onSubmit={submit} noValidate className="space-y-3">
      <p className="text-[13px] text-slate-600">{intro}</p>
      <Field
        id="pay-tx-hash"
        label="Hash transaksi"
        placeholder="0x diikuti 64 karakter"
        autoComplete="off"
        spellCheck={false}
        maxLength={80}
        hint="Salin dari riwayat aktivitas di dompet atau dari Etherscan"
        value={raw}
        onChange={(event) => setRaw(event.target.value)}
      />
      {invalid ? <ErrorNote message="Hash transaksi harus 0x diikuti 64 karakter heksadesimal (0-9, a-f)." /> : null}
      <Button type="submit" variant="subtle" isPending={isChecking} pendingLabel="Memeriksa transaksi">
        Periksa transaksi
      </Button>
    </form>
  )
}

function WalletPay({ payment, onHash }: { payment: PublicPayment; onHash: (hash: string) => void }) {
  const wallet = injectedWallet()
  const send = useMutation({
    mutationFn: () => {
      if (wallet === null) {
        throw new Error('Dompet tidak ditemukan')
      }
      return payWithWallet(wallet, payment)
    },
    onSuccess: onHash,
  })

  if (wallet === null) {
    return (
      <div className="space-y-2 rounded-xl bg-slate-50 px-3.5 py-3 text-[13px] text-slate-700">
        <p>MetaMask tidak terdeteksi di peramban ini. Pindai QR di atas dengan dompet di HP, atau buka halaman ini di aplikasi MetaMask.</p>
        <a
          href={metamaskDeepLink(window.location.href)}
          className="inline-flex font-medium text-navy-600 underline-offset-4 hover:underline"
        >
          Buka di aplikasi MetaMask
        </a>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <Button isPending={send.isPending} pendingLabel="Menunggu persetujuan di MetaMask" onClick={() => send.mutate()}>
        <Wallet aria-hidden="true" className="size-4" strokeWidth={1.8} />
        Bayar dengan MetaMask
      </Button>
      <p className="text-xs text-slate-500">
        MetaMask akan pindah ke jaringan {networkName(payment.chainId)} lalu meminta persetujuan mengirim{' '}
        {payment.amountEth.replace('.', ',')} {currencyName(payment.chainId)}. Biaya gas juga dibayar dengan SepoliaETH.
      </p>
      {send.isError ? <ErrorNote message={walletErrorMessage(send.error)} /> : null}
    </div>
  )
}

function PayOptions({ payment, now, onHash, isChecking }: {
  payment: PublicPayment
  now: number
  onHash: (hash: string) => void
  isChecking: boolean
}) {
  const minutesLeft = Math.max(0, Math.ceil((new Date(payment.expiresAt).getTime() - now) / MINUTE_MS))
  const [isManualOpen, setIsManualOpen] = useState(false)

  return (
    <div className="space-y-5">
      <p className="text-[13px] text-slate-600">
        Berlaku sampai {timeLabel(payment.expiresAt)}
        {minutesLeft > 0 ? ` (sisa sekitar ${minutesLeft} menit)` : ''}.
      </p>

      <div className="flex flex-col items-center gap-3 rounded-xl bg-slate-50 p-4 sm:flex-row sm:items-start">
        <QrImage value={payment.paymentUri} label="Kode QR pembayaran untuk dompet di HP" className="size-48 shrink-0" />
        <div className="min-w-0 space-y-2 text-[13px] text-slate-700">
          <p>Pindai dengan MetaMask atau dompet Ethereum lain di HP. Alamat, jaringan, dan nilainya sudah terisi.</p>
          <p className="text-xs font-medium text-slate-500">Alamat penerima</p>
          <p className="font-mono text-xs break-all text-slate-900">{payment.receiver}</p>
          <CopyButton value={payment.receiver} label="Salin alamat penerima" />
        </div>
      </div>

      <WalletPay payment={payment} onHash={onHash} />

      <div className="border-t border-slate-100 pt-4">
        {isManualOpen ? (
          <ManualHashForm
            isChecking={isChecking}
            onSubmit={onHash}
            intro="Sudah membayar dari HP atau dompet lain? Tempel hash transaksinya supaya pembayaran langsung dicatat."
          />
        ) : (
          <Button variant="subtle" onClick={() => setIsManualOpen(true)}>
            Saya sudah bayar, tempel hash transaksi
          </Button>
        )}
      </div>
    </div>
  )
}

function WaitingPanel({ payment, hash, message, onReset, onRetry }: {
  payment: PublicPayment
  hash: string
  message: string
  onReset: () => void
  onRetry: (() => void) | null
}) {
  return (
    <div role="status" className="space-y-3 rounded-xl bg-amber-50 px-4 py-3.5">
      <p className="flex items-center gap-2 text-sm font-medium text-amber-900">
        <span aria-hidden="true" className="size-3.5 rounded-full border-2 border-current border-r-transparent motion-safe:animate-spin" />
        Menunggu konfirmasi jaringan
      </p>
      <p className="text-[13px] text-amber-900">{message}. Halaman ini memeriksa ulang setiap beberapa detik.</p>
      <p className="text-[13px] text-amber-900">
        Transaksi: <TxLink url={explorerTxUrl(payment, hash)} hash={hash} />
      </p>
      <span className="flex flex-wrap gap-2">
        {onRetry ? <Button onClick={onRetry}>Periksa lagi</Button> : null}
        <Button variant="subtle" onClick={onReset}>
          Ganti hash transaksi
        </Button>
      </span>
    </div>
  )
}

function FinalPanel({ payment }: { payment: PublicPayment }) {
  if (payment.status === 'paid') {
    return (
      <div className="space-y-3">
        <SuccessNote message={`Pembayaran diterima dan sudah tercatat${payment.paidAt ? ` pada ${timeLabel(payment.paidAt)}` : ''}. Terima kasih.`} />
        {payment.txHash && payment.txUrl ? (
          <p className="text-[13px] text-slate-700">
            Bukti di Etherscan: <TxLink url={payment.txUrl} hash={payment.txHash} />
          </p>
        ) : null}
      </div>
    )
  }
  if (payment.status === 'failed') {
    return (
      <div className="space-y-3">
        <ErrorNote
          message={`Transaksi diterima di jaringan, tetapi tidak bisa dicatat ke tagihan: ${payment.failureReason}. Hubungi bagian keuangan.`}
        />
        {payment.txHash && payment.txUrl ? (
          <p className="text-[13px] text-slate-700">
            Transaksi: <TxLink url={payment.txUrl} hash={payment.txHash} />
          </p>
        ) : null}
      </div>
    )
  }
  return <ErrorNote message="Permintaan bayar ini sudah dibatalkan. Minta tautan baru ke bagian keuangan." />
}

function PaymentBody({ token, payment, now }: { token: string; payment: PublicPayment; now: number }) {
  const queryClient = useQueryClient()
  const [hash, setHash] = useState<string | null>(() => readSubmittedHash(token))
  const confirmation = usePaymentConfirmation(token, hash)
  const outcome: ConfirmOutcome | undefined = confirmation.data
  const verification = outcome?.verification
  const isFinal = verification === 'paid' || verification === 'failed'
  const view = isFinal && outcome ? outcome : payment
  const isRejected = confirmation.isError && isRejection(confirmation.error)

  useEffect(() => {
    if (isFinal) {
      void queryClient.invalidateQueries({ queryKey: ['pay', token], exact: true })
    }
  }, [isFinal, queryClient, token])

  useEffect(() => {
    if (isRejected) {
      storeSubmittedHash(token, null)
    }
  }, [isRejected, token])

  const submitHash = (next: string) => {
    if (next === hash) {
      void confirmation.refetch()
      return
    }
    storeSubmittedHash(token, next)
    setHash(next)
  }

  const reset = () => {
    storeSubmittedHash(token, null)
    setHash(null)
  }

  const isOpen = view.status === 'pending' || view.status === 'expired'
  const isWaiting = isOpen && hash !== null && !isRejected && (confirmation.isFetching || verification === 'waiting' || confirmation.isError)

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-lg font-semibold text-slate-900">
            Bayar {ONCHAIN_TARGET_LABEL[view.targetKind]} {view.targetNumber}
          </h1>
          {view.projectName ? <p className="text-[13px] text-slate-500">Proyek {view.projectName}</p> : null}
        </div>
        <OnchainStatusChip status={view.status} />
      </div>

      <AmountSummary amountIdr={view.amountIdr} amountEth={view.amountEth} idrPerEth={view.idrPerEth} chainId={view.chainId} />
      <TestnetNote chainId={view.chainId} />

      {isOpen ? null : <FinalPanel payment={view} />}

      {isWaiting && hash !== null ? (
        <WaitingPanel
          payment={view}
          hash={hash}
          message={outcome?.message ?? (confirmation.isError ? errorMessage(confirmation.error) : 'Memeriksa transaksi di jaringan')}
          onReset={reset}
          onRetry={confirmation.isError && !confirmation.isFetching ? () => void confirmation.refetch() : null}
        />
      ) : null}

      {isRejected ? <ErrorNote message={errorMessage(confirmation.error)} /> : null}

      {isOpen && !isWaiting && view.status === 'pending' ? (
        <PayOptions payment={view} now={now} onHash={submitHash} isChecking={confirmation.isFetching} />
      ) : null}

      {isOpen && !isWaiting && view.status === 'expired' ? (
        <div className="space-y-4">
          <ErrorNote message="Permintaan bayar ini sudah kedaluwarsa. Minta tautan baru ke bagian keuangan." />
          <ManualHashForm
            isChecking={confirmation.isFetching}
            onSubmit={submitHash}
            intro="Sudah mengirim sebelum batas waktu? Tempel hash transaksinya, pembayaran tetap dicatat kalau masuk blok tepat waktu."
          />
        </div>
      ) : null}
    </div>
  )
}

export function PayPage() {
  const token = useParams().token ?? ''
  const payment = usePublicPayment(token)
  const now = useNow(MINUTE_MS / 2)

  return (
    <PayLayout>
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
      {payment.data ? <PaymentBody token={token} payment={payment.data} now={now} /> : null}
    </PayLayout>
  )
}
