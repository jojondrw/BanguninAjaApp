import { useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import {
  useAnchorBatch,
  useCreateEvidenceBatch,
  useEvidenceConfig,
  useProjectEvidence,
  useServerAnchor,
} from '../../controllers/useEvidence'
import type { AnchorResult, EvidenceConfig } from '../../models/evidence'
import { shortHex } from '../../models/onchain'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { Card, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { EvidenceList, EvidenceSummaryLine } from '../evidence/EvidenceList'
import { TxLink } from '../pay/PayParts'
import { injectedWallet, sendCalldata, walletErrorMessage } from '../pay/wallet'

const POLL_MS = 4_000
const POLL_LIMIT = 45

type Step = 'idle' | 'preparing' | 'wallet' | 'server' | 'confirming'

const STEP_LABEL: Record<Step, string> = {
  idle: 'Catat bukti ke blockchain',
  preparing: 'Menyiapkan sidik jari',
  wallet: 'Menunggu dompet',
  server: 'Server mengirim transaksi',
  confirming: 'Menunggu konfirmasi jaringan',
}

function wait(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

function RegistryNote({ config }: { config: EvidenceConfig }) {
  if (!config.registryAddress && !config.serverSigner) {
    return null
  }
  return (
    <p className="text-xs text-slate-500">
      {config.registryAddress ? (
        <>
          Dicatat ke kontrak{' '}
          <a
            href={config.registryUrl}
            target="_blank"
            rel="noreferrer"
            className="font-mono text-navy-600 underline-offset-4 hover:underline"
          >
            EvidenceRegistry {shortHex(config.registryAddress)}
          </a>
          .{' '}
        </>
      ) : null}
      {config.serverSigner ? `Dompet server ${shortHex(config.signerAddress)}` : ''}
      {config.serverSigner && config.autoAnchorHours > 0 ? `, mencatat otomatis tiap ${config.autoAnchorHours} jam.` : config.serverSigner ? '.' : ''}
    </p>
  )
}

// Sidik jari (SHA-256) laporan harian, foto lapangan, progres, tagihan termin,
// dan pembayaran semua metode dicatat ke blockchain sebagai satu akar Merkle.
// Bila dompet server terpasang, server menandatangani transaksinya sendiri;
// MetaMask tetap bisa dipakai sebagai jalur manual.
export function EvidenceCard({ project }: { project: Project }) {
  const evidence = useProjectEvidence(project.id)
  const config = useEvidenceConfig()
  const createBatch = useCreateEvidenceBatch()
  const serverAnchor = useServerAnchor()
  const anchorBatch = useAnchorBatch()
  const canWrite = useScopedCan('write')
  const [step, setStep] = useState<Step>('idle')
  const [problem, setProblem] = useState<string | null>(null)
  const [done, setDone] = useState<AnchorResult | null>(null)
  const hasServer = config.data?.serverSigner === true

  async function confirm(batchId: string, txHash: string) {
    setStep('confirming')
    for (let attempt = 0; attempt < POLL_LIMIT; attempt++) {
      const result = await anchorBatch.mutateAsync({ batchId, txHash })
      if (result.status === 'anchored') {
        setDone(result)
        return
      }
      await wait(POLL_MS)
    }
    setProblem('Transaksi belum terkonfirmasi. Tunggu sebentar lalu klik lagi; transaksi yang sama akan dipakai.')
  }

  async function anchorWithServer() {
    setProblem(null)
    setDone(null)
    try {
      setStep('server')
      const sent = await serverAnchor.mutateAsync()
      await confirm(sent.batch.id, sent.txHash)
    } catch (error) {
      setProblem(errorMessage(error))
    } finally {
      setStep('idle')
    }
  }

  async function anchorWithWallet() {
    setProblem(null)
    setDone(null)
    const wallet = injectedWallet()
    if (!wallet) {
      setProblem('Dompet tidak terdeteksi. Pasang MetaMask di peramban ini, lalu coba lagi.')
      return
    }
    let phase: Step = 'preparing'
    try {
      setStep(phase)
      const batch = await createBatch.mutateAsync()
      phase = 'wallet'
      setStep(phase)
      const txHash = await sendCalldata(wallet, batch.chainId, batch.calldata, batch.anchorTo)
      phase = 'confirming'
      await confirm(batch.id, txHash)
    } catch (error) {
      setProblem(phase === 'wallet' || isWalletError(error) ? walletErrorMessage(error) : errorMessage(error))
    } finally {
      setStep('idle')
    }
  }

  const busy = step !== 'idle'
  function renderActions() {
    if (busy) {
      return (
        <Button isPending pendingLabel={STEP_LABEL[step]}>
          {STEP_LABEL.idle}
        </Button>
      )
    }
    return (
      <span className="flex flex-wrap items-center justify-end gap-2">
        {hasServer ? (
          <Button variant="subtle" onClick={() => void anchorWithWallet()}>
            Pakai MetaMask
          </Button>
        ) : null}
        <Button onClick={() => void (hasServer ? anchorWithServer() : anchorWithWallet())}>{STEP_LABEL.idle}</Button>
      </span>
    )
  }


  return (
    <Card
      title="Jejak bukti di blockchain"
      description="Sidik jari laporan, foto, progres, tagihan termin, dan pembayaran (transfer, tunai, maupun on-chain) dicatat ke Sepolia. Bukti yang diubah setelah dicatat langsung ketahuan."
      action={canWrite ? renderActions() : undefined}
    >
      {evidence.isPending ? <Loading /> : null}
      {evidence.isError ? <LoadFailed onRetry={() => evidence.refetch()} /> : null}
      {problem ? (
        <div className="mb-4">
          <ErrorNote message={problem} />
        </div>
      ) : null}
      {done ? (
        <div className="mb-4 space-y-1.5">
          <SuccessNote message={`${done.batch.recordCount} sidik jari bukti (semua proyek) tercatat dalam satu transaksi.`} />
          <TxLink url={done.batch.txUrl} hash={done.batch.txHash} />
        </div>
      ) : null}
      {evidence.data ? (
        <div className="space-y-4">
          <EvidenceSummaryLine summary={evidence.data.summary} />
          <EvidenceList items={evidence.data.items} />
          <p className="text-xs text-slate-500">
            Satu transaksi mencatat bukti semua proyek sekaligus. Yang tercatat hanya sidik jari, isi laporan dan foto tetap di
            server.
          </p>
          {config.data ? <RegistryNote config={config.data} /> : null}
        </div>
      ) : null}
    </Card>
  )
}

function isWalletError(error: unknown): boolean {
  return typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'number'
}
