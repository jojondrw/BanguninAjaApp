import { Check, Copy, ExternalLink, FlaskConical } from 'lucide-react'
import { useEffect, useState } from 'react'

import {
  ONCHAIN_STATUS_LABEL,
  ONCHAIN_STATUS_TONE,
  etherLabel,
  networkName,
  rateLabel,
  shortHex,
  type OnchainStatus,
} from '../../models/onchain'
import { rupiah } from '../../shared/format'
import { BUTTON_BASE, BUTTON_SUBTLE } from '../components/Form'
import { Chip } from '../components/RecordControls'
import { Money } from '../components/Money'

const COPIED_FEEDBACK_MS = 2_000
const FAUCET_URL = 'https://cloud.google.com/application/web3/faucet/ethereum/sepolia'

export function OnchainStatusChip({ status }: { status: OnchainStatus }) {
  return <Chip tone={ONCHAIN_STATUS_TONE[status]}>{ONCHAIN_STATUS_LABEL[status]}</Chip>
}

// Salin ke papan klip. Kalau peramban menolak (misalnya bukan HTTPS), teksnya
// tetap terlihat di sebelah tombol untuk disalin manual.
export function CopyButton({ value, label }: { value: string; label: string }) {
  const [state, setState] = useState<'idle' | 'copied' | 'failed'>('idle')

  useEffect(() => {
    if (state === 'idle') {
      return
    }
    const timer = window.setTimeout(() => setState('idle'), COPIED_FEEDBACK_MS)
    return () => window.clearTimeout(timer)
  }, [state])

  const copy = () => {
    if (!navigator.clipboard) {
      setState('failed')
      return
    }
    navigator.clipboard.writeText(value).then(
      () => setState('copied'),
      () => setState('failed'),
    )
  }

  return (
    <button type="button" onClick={copy} aria-label={label} className={`h-8 ${BUTTON_BASE} ${BUTTON_SUBTLE}`}>
      {state === 'copied' ? (
        <Check aria-hidden="true" className="size-4" strokeWidth={1.8} />
      ) : (
        <Copy aria-hidden="true" className="size-4" strokeWidth={1.8} />
      )}
      {state === 'copied' ? 'Disalin' : state === 'failed' ? 'Salin manual' : 'Salin'}
    </button>
  )
}

export function TxLink({ url, hash }: { url: string; hash: string }) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      className="inline-flex items-center gap-1 font-mono text-[13px] text-navy-600 underline-offset-4 hover:underline"
    >
      {shortHex(hash)}
      <ExternalLink aria-hidden="true" className="size-3.5" strokeWidth={1.8} />
      <span className="sr-only">(lihat di Etherscan, tab baru)</span>
    </a>
  )
}

export function AmountSummary({ amountIdr, amountEth, idrPerEth, chainId }: {
  amountIdr: number
  amountEth: string
  idrPerEth: number
  chainId: number
}) {
  return (
    <div>
      <p className="text-[28px] leading-9 font-semibold tracking-[-0.025em] text-slate-900 tabular-nums"><Money>{rupiah(amountIdr)}</Money></p>
      <p className="mt-1 text-[15px] text-slate-700 tabular-nums">≈ {etherLabel(amountEth, chainId)}</p>
      <p className="mt-1 text-xs text-slate-500">{rateLabel(idrPerEth, chainId)}, bukan harga pasar.</p>
    </div>
  )
}

export function TestnetNote({ chainId }: { chainId: number }) {
  return (
    <div className="flex items-start gap-2.5 rounded-xl bg-navy-50 px-3.5 py-3 text-[13px] text-navy-800">
      <FlaskConical aria-hidden="true" className="mt-0.5 size-4 shrink-0" strokeWidth={1.8} />
      <p>
        Ini simulasi pembayaran di jaringan uji Ethereum {networkName(chainId)}. Yang dikirim adalah SepoliaETH dari
        faucet, bukan uang sungguhan. Belum punya?{' '}
        <a href={FAUCET_URL} target="_blank" rel="noreferrer" className="font-medium underline underline-offset-4">
          Ambil SepoliaETH gratis di faucet
        </a>
        .
      </p>
    </div>
  )
}
