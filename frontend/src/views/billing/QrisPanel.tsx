import { ExternalLink } from 'lucide-react'
import { type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import {
  useCancelGatewayRequest,
  useCreateGatewayRequest,
  useGatewayConfig,
  useGatewayRequests,
} from '../../controllers/useGateway'
import type { BillingRecord } from '../../models/billing'
import {
  defaultQrisAmount,
  providerLabel,
  qrisAmountError,
  qrisPayLink,
  type GatewayConfig,
  type GatewayRequest,
} from '../../models/gateway'
import { ONCHAIN_TARGET_LABEL, TARGET_KIND_OF, timeLabel, type OnchainBillingKind } from '../../models/onchain'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { LoadFailed, Loading } from '../components/Data'
import { BUTTON_BASE, BUTTON_SUBTLE, Button, CONTROL_CLASS, ErrorNote, Field, LABEL_CLASS, SuccessNote } from '../components/Form'
import { Money } from '../components/Money'
import { FormPanel, Pager } from '../components/RecordControls'
import { CopyButton, OnchainStatusChip } from '../pay/PayParts'
import { QrImage } from '../pay/QrImage'

function CreateQrisForm({ kind, record, config, onCreated }: {
  kind: OnchainBillingKind
  record: BillingRecord
  config: GatewayConfig
  onCreated: () => void
}) {
  const create = useCreateGatewayRequest(kind, record.id)
  const [amount, setAmount] = useState(String(defaultQrisAmount(record.outstanding, config.maxAmount)))
  const problem = qrisAmountError(Number(amount), record.outstanding, config.maxAmount)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (problem === null) {
      create.mutate(Number(amount), { onSuccess: onCreated })
    }
  }

  return (
    <form onSubmit={submit} className="space-y-3">
      <Field
        id="qris-amount"
        label="Nominal yang ditagih"
        type="number"
        inputMode="numeric"
        min={1}
        max={Math.min(record.outstanding, config.maxAmount)}
        step={1}
        required
        hint={`Sisa tagihan ${rupiah(record.outstanding)}. QRIS paling banyak ${rupiah(config.maxAmount)} per transaksi.`}
        value={amount}
        onChange={(event) => setAmount(event.target.value)}
      />
      <p className="text-xs text-slate-500">
        Kode QR berlaku {config.requestTtlMinutes} menit dan bisa dibayar dari aplikasi bank atau dompet digital mana pun.
        Pembayaran dicatat otomatis begitu {providerLabel(config.provider)} mengonfirmasinya.
      </p>
      <Button type="submit" isPending={create.isPending} pendingLabel="Membuat QRIS">
        Buat QRIS
      </Button>
      {problem !== null && amount !== '' ? <p className="text-xs text-red-700">{problem}</p> : null}
      {create.isError ? <ErrorNote message={errorMessage(create.error)} /> : null}
    </form>
  )
}

function CancelQris({ request }: { request: GatewayRequest }) {
  const cancel = useCancelGatewayRequest()
  const [isConfirming, setIsConfirming] = useState(false)

  if (!isConfirming) {
    return (
      <Button variant="subtle" onClick={() => setIsConfirming(true)}>
        Batalkan QRIS
      </Button>
    )
  }

  return (
    <span className="flex flex-col gap-2">
      <span className="flex flex-wrap items-center gap-2 rounded-xl bg-amber-50 px-3 py-2">
        <span className="text-[13px] text-amber-900">Batalkan QRIS ini? Kodenya tidak bisa dibayar lagi.</span>
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

function ActiveQris({ request, canWrite }: { request: GatewayRequest; canWrite: boolean }) {
  const link = qrisPayLink(window.location.origin, request.token)

  return (
    <div className="mt-4 rounded-xl bg-slate-50 p-4 shadow-[inset_0_0_0_1px_rgb(0_0_0/0.04)]">
      <div className="flex flex-col gap-4 sm:flex-row">
        <QrImage value={request.qrString} label="Kode QRIS" className="size-40 shrink-0 self-center sm:self-start" />
        <div className="min-w-0 flex-1 space-y-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <OnchainStatusChip status={request.status} />
            <span className="text-xs text-slate-500">Berlaku sampai {timeLabel(request.expiresAt)}</span>
          </div>
          <p className="text-2xl font-semibold tracking-[-0.02em] text-slate-900 tabular-nums">
            <Money>{rupiah(request.amountIdr)}</Money>
          </p>
          <div className="flex flex-col gap-1.5">
            <label htmlFor={`qris-link-${request.id}`} className={LABEL_CLASS}>
              Tautan halaman bayar untuk pembeli
            </label>
            <div className="flex flex-wrap items-center gap-2">
              <input
                id={`qris-link-${request.id}`}
                readOnly
                value={link}
                onFocus={(event) => event.target.select()}
                className={`${CONTROL_CLASS} h-8 min-w-0 flex-1 px-2.5 font-mono text-xs`}
              />
              <CopyButton value={link} label="Salin tautan halaman bayar QRIS" />
              <a href={link} target="_blank" rel="noreferrer" className={`h-8 ${BUTTON_BASE} ${BUTTON_SUBTLE}`}>
                <ExternalLink aria-hidden="true" className="size-4" strokeWidth={1.8} />
                Buka
              </a>
            </div>
          </div>
          <p className="text-xs text-slate-500">Status di sini berubah sendiri begitu pembayaran dikonfirmasi.</p>
          {canWrite ? <CancelQris request={request} /> : null}
        </div>
      </div>
    </div>
  )
}

function QrisItem({ request, kind }: { request: GatewayRequest; kind: OnchainBillingKind }) {
  const fromLinked = request.targetKind !== TARGET_KIND_OF[kind]

  return (
    <li className="flex flex-wrap items-start justify-between gap-x-4 gap-y-1 py-3">
      <div className="min-w-0 space-y-0.5">
        <p className="flex flex-wrap items-center gap-2 text-sm text-slate-900">
          <span className="font-medium">{timeLabel(request.createdAt)}</span>
          <OnchainStatusChip status={request.status} />
        </p>
        <p className="text-xs text-slate-500">
          {providerLabel(request.provider)}, ref. <span className="font-mono">{request.providerId}</span>
        </p>
        {request.failureReason ? <p className="text-xs break-words text-red-700">{request.failureReason}</p> : null}
        {fromLinked ? (
          <p className="text-xs text-navy-700">Dibuat dari {ONCHAIN_TARGET_LABEL[request.targetKind]} yang tertaut</p>
        ) : null}
        {request.createdByName ? <p className="text-xs text-slate-500">Dibuat oleh {request.createdByName}</p> : null}
      </div>
      <span className="text-sm font-medium text-slate-900 tabular-nums">
        <Money>{rupiah(request.amountIdr)}</Money>
      </span>
    </li>
  )
}

// Bagian "Bayar dengan QRIS" di rincian faktur pelanggan dan piutang. Tidak
// tampil kalau server belum dihubungkan ke gerbang pembayaran.
export function QrisSection({ kind, record }: { kind: OnchainBillingKind; record: BillingRecord }) {
  const config = useGatewayConfig()
  const canWrite = useScopedCan('write')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const requests = useGatewayRequests(kind, record.id, page, config.data?.enabled === true)

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
    <section className="mt-6 border-t border-slate-100 pt-5" aria-labelledby="qris-title">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 id="qris-title" className="text-sm font-semibold text-slate-900">
            Bayar dengan QRIS
          </h3>
          <p className="mt-0.5 text-xs text-slate-500">
            Lewat {providerLabel(config.data.provider)}
            {config.data.testMode ? ' dalam mode uji, tanpa uang sungguhan' : ''}. Server menanyakan sendiri status
            pembayarannya, lalu mencatatnya ke tagihan dan buku kas.
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
            {isFormOpen ? 'Tutup formulir' : 'Minta pembayaran QRIS'}
          </Button>
        ) : null}
      </div>

      {isFormOpen && canCreate ? (
        <div className="mt-4">
          <FormPanel>
            <CreateQrisForm
              kind={kind}
              record={record}
              config={config.data}
              onCreated={() => {
                setIsFormOpen(false)
                setPage(1)
                setNotice('QRIS dibuat. Kirim tautannya atau tunjukkan kodenya ke pembeli.')
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

      {active ? <ActiveQris key={active.id} request={active} canWrite={canWrite} /> : null}

      {requests.isPending ? <Loading label="Mengambil riwayat QRIS..." /> : null}
      {requests.isError ? <LoadFailed onRetry={() => requests.refetch()} /> : null}
      {requests.data && items.length > 0 ? (
        <>
          <ul className="mt-2 divide-y divide-slate-100" aria-label="Riwayat permintaan QRIS">
            {items.map((request) => (
              <QrisItem key={request.id} request={request} kind={kind} />
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
