import { ExternalLink } from 'lucide-react'
import { type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import {
  useCancelOnchainRequest,
  useCreateOnchainRequest,
  useOnchainConfig,
  useOnchainRequests,
} from '../../controllers/useOnchain'
import type { BillingRecord } from '../../models/billing'
import {
  ONCHAIN_TARGET_LABEL,
  TARGET_KIND_OF,
  etherLabel,
  formatEther,
  networkName,
  payLink,
  rateLabel,
  shortHex,
  timeLabel,
  weiForRupiah,
  type OnchainBillingKind,
  type OnchainConfig,
  type OnchainRequest,
} from '../../models/onchain'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { LoadFailed, Loading } from '../components/Data'
import { BUTTON_BASE, BUTTON_SUBTLE, Button, CONTROL_CLASS, ErrorNote, Field, LABEL_CLASS, SuccessNote } from '../components/Form'
import { FormPanel, Pager } from '../components/RecordControls'
import { AmountSummary, CopyButton, OnchainStatusChip, TxLink } from '../pay/PayParts'
import { QrImage } from '../pay/QrImage'
import { Money } from '../components/Money'

function CreateRequestForm({ kind, record, config, onCreated }: {
  kind: OnchainBillingKind
  record: BillingRecord
  config: OnchainConfig
  onCreated: () => void
}) {
  const create = useCreateOnchainRequest(kind, record.id)
  const [amount, setAmount] = useState(String(record.outstanding))
  const value = Number(amount)
  const isValid = Number.isInteger(value) && value > 0 && value <= record.outstanding
  const preview = isValid ? etherLabel(formatEther(weiForRupiah(value, config.idrPerEth)), config.chainId) : null

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (isValid) {
      create.mutate(value, { onSuccess: onCreated })
    }
  }

  return (
    <form onSubmit={submit} className="space-y-3">
      <div className="grid gap-4 sm:grid-cols-2">
        <Field
          id="onchain-amount"
          label="Nominal yang ditagih"
          type="number"
          inputMode="numeric"
          min={1}
          max={record.outstanding}
          step={1}
          required
          hint={`Terisi sisa ${rupiah(record.outstanding)}. Boleh sebagian.`}
          value={amount}
          onChange={(event) => setAmount(event.target.value)}
        />
        <div className="flex flex-col gap-1.5">
          <span className={LABEL_CLASS}>Setara</span>
          <p className="flex h-9 items-center text-sm text-slate-900 tabular-nums">{preview ?? '-'}</p>
          <p className="text-xs text-slate-500">{rateLabel(config.idrPerEth, config.chainId)}, bukan harga pasar</p>
        </div>
      </div>
      <p className="text-xs text-slate-500">
        Tautan berlaku {config.requestTtlMinutes} menit. Pembayaran dicatat otomatis setelah transaksi mendapat{' '}
        {config.minConfirmations} konfirmasi di jaringan {networkName(config.chainId)}.
      </p>
      <Button type="submit" isPending={create.isPending} pendingLabel="Membuat permintaan">
        Buat permintaan bayar
      </Button>
      {create.isError ? <ErrorNote message={errorMessage(create.error)} /> : null}
    </form>
  )
}

function CancelRequest({ request }: { request: OnchainRequest }) {
  const cancel = useCancelOnchainRequest()
  const [isConfirming, setIsConfirming] = useState(false)

  if (!isConfirming) {
    return (
      <Button variant="subtle" onClick={() => setIsConfirming(true)}>
        Batalkan permintaan
      </Button>
    )
  }

  return (
    <span className="flex flex-col gap-2">
      <span className="flex flex-wrap items-center gap-2 rounded-xl bg-amber-50 px-3 py-2">
        <span className="text-[13px] text-amber-900">Batalkan permintaan ini? Tautannya tidak bisa dipakai lagi.</span>
        <Button
          isPending={cancel.isPending}
          pendingLabel="Membatalkan"
          onClick={() => cancel.mutate(request.id, { onSettled: () => setIsConfirming(false) })}
        >
          Ya, batalkan
        </Button>
        {cancel.isPending ? null : (
          <Button variant="subtle" onClick={() => setIsConfirming(false)}>
            Tidak jadi
          </Button>
        )}
      </span>
      {cancel.isError ? <ErrorNote message={errorMessage(cancel.error)} /> : null}
    </span>
  )
}

function ActiveRequest({ request, canWrite }: { request: OnchainRequest; canWrite: boolean }) {
  const link = payLink(window.location.origin, request.token)

  return (
    <div className="mt-4 rounded-xl bg-slate-50 p-4 shadow-[inset_0_0_0_1px_rgb(0_0_0/0.04)]">
      <div className="flex flex-col gap-4 sm:flex-row">
        <QrImage value={request.paymentUri} label="Kode QR pembayaran on-chain" className="size-40 shrink-0 self-center sm:self-start" />
        <div className="min-w-0 flex-1 space-y-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <OnchainStatusChip status={request.status} />
            <span className="text-xs text-slate-500">Berlaku sampai {timeLabel(request.expiresAt)}</span>
          </div>
          <AmountSummary
            amountIdr={request.amountIdr}
            amountEth={request.amountEth}
            idrPerEth={request.idrPerEth}
            chainId={request.chainId}
          />
          <div className="flex flex-col gap-1.5">
            <label htmlFor={`onchain-link-${request.id}`} className={LABEL_CLASS}>
              Tautan halaman bayar untuk pembeli
            </label>
            <div className="flex flex-wrap items-center gap-2">
              <input
                id={`onchain-link-${request.id}`}
                readOnly
                value={link}
                onFocus={(event) => event.target.select()}
                className={`${CONTROL_CLASS} h-8 min-w-0 flex-1 px-2.5 font-mono text-xs`}
              />
              <CopyButton value={link} label="Salin tautan halaman bayar" />
              <a href={link} target="_blank" rel="noreferrer" className={`h-8 ${BUTTON_BASE} ${BUTTON_SUBTLE}`}>
                <ExternalLink aria-hidden="true" className="size-4" strokeWidth={1.8} />
                Buka
              </a>
            </div>
          </div>
          <p className="text-xs text-slate-500">Status di sini berubah sendiri begitu pembeli selesai membayar.</p>
          {canWrite ? <CancelRequest request={request} /> : null}
        </div>
      </div>
    </div>
  )
}

function RequestItem({ request, kind }: { request: OnchainRequest; kind: OnchainBillingKind }) {
  const fromLinked = request.targetKind !== TARGET_KIND_OF[kind]

  return (
    <li className="flex flex-wrap items-start justify-between gap-x-4 gap-y-1 py-3">
      <div className="min-w-0 space-y-0.5">
        <p className="flex flex-wrap items-center gap-2 text-sm text-slate-900">
          <span className="font-medium">{timeLabel(request.createdAt)}</span>
          <OnchainStatusChip status={request.status} />
        </p>
        {request.txHash && request.txUrl ? (
          <p className="text-xs text-slate-600">
            Transaksi <TxLink url={request.txUrl} hash={request.txHash} />
            {request.payerAddress ? ` dari ${shortHex(request.payerAddress)}` : ''}
          </p>
        ) : null}
        {request.failureReason ? <p className="text-xs break-words text-red-700">{request.failureReason}</p> : null}
        {fromLinked ? (
          <p className="text-xs text-navy-700">Dibuat dari {ONCHAIN_TARGET_LABEL[request.targetKind]} yang tertaut</p>
        ) : null}
        {request.createdByName ? <p className="text-xs text-slate-500">Dibuat oleh {request.createdByName}</p> : null}
      </div>
      <span className="text-right text-sm tabular-nums">
        <span className="block font-medium text-slate-900"><Money>{rupiah(request.amountIdr)}</Money></span>
        <span className="block text-xs text-slate-500">{etherLabel(request.amountEth, request.chainId)}</span>
      </span>
    </li>
  )
}

// Bagian "Pembayaran on-chain" di rincian faktur pelanggan dan piutang. Tidak
// tampil sama sekali kalau server belum diisi alamat penerima.
export function OnchainSection({ kind, record }: { kind: OnchainBillingKind; record: BillingRecord }) {
  const config = useOnchainConfig()
  const canWrite = useScopedCan('write')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const requests = useOnchainRequests(kind, record.id, page, config.data?.enabled === true)

  if (!config.data?.enabled) {
    return null
  }
  const items = requests.data?.items ?? []
  if (!canWrite && items.length === 0) {
    return null
  }
  const active = items.find((request) => request.status === 'pending') ?? null
  const canCreate = canWrite && record.outstanding > 0 && active === null

  return (
    <section className="mt-6 border-t border-slate-100 pt-5" aria-labelledby="onchain-title">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 id="onchain-title" className="text-sm font-semibold text-slate-900">
            Pembayaran on-chain
          </h3>
          <p className="mt-0.5 text-xs text-slate-500">
            Pembeli membayar SepoliaETH di jaringan uji {networkName(config.data.chainId)}. Server memeriksa transaksinya
            sendiri lalu mencatat pembayaran.
          </p>
        </div>
        {canCreate ? (
          <Button
            variant="subtle"
            onClick={() => {
              setIsFormOpen((open) => !open)
              setNotice(null)
            }}
          >
            {isFormOpen ? 'Tutup formulir' : 'Bayar on-chain'}
          </Button>
        ) : null}
      </div>

      {isFormOpen && canCreate ? (
        <div className="mt-4">
          <FormPanel>
            <CreateRequestForm
              kind={kind}
              record={record}
              config={config.data}
              onCreated={() => {
                setIsFormOpen(false)
                setPage(1)
                setNotice('Permintaan bayar dibuat. Kirim tautan atau tunjukkan QR-nya ke pembeli.')
              }}
            />
          </FormPanel>
        </div>
      ) : null}

      {notice ? (
        <div className="mt-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}

      {active ? <ActiveRequest key={active.id} request={active} canWrite={canWrite} /> : null}

      {requests.isPending ? <Loading label="Mengambil permintaan bayar on-chain..." /> : null}
      {requests.isError ? <LoadFailed onRetry={() => requests.refetch()} /> : null}
      {requests.data && items.length > 0 ? (
        <>
          <ul className="mt-2 divide-y divide-slate-100" aria-label="Riwayat permintaan bayar on-chain">
            {items.map((request) => (
              <RequestItem key={request.id} request={request} kind={kind} />
            ))}
          </ul>
          <Pager
            page={requests.data.page}
            totalPages={requests.data.totalPages}
            totalItems={requests.data.totalItems}
            unit="permintaan"
            onChange={setPage}
          />
        </>
      ) : null}
    </section>
  )
}
