import { useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useAnchorBatch, useCreateEvidenceBatch, useProjectEvidence } from '../../controllers/useEvidence'
import type { AnchorResult } from '../../models/evidence'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { Card, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { EvidenceList, EvidenceSummaryLine } from '../evidence/EvidenceList'
import { TxLink } from '../pay/PayParts'
import { injectedWallet, sendCalldata, walletErrorMessage } from '../pay/wallet'

const POLL_MS = 4_000
const POLL_LIMIT = 45

type Step = 'idle' | 'preparing' | 'wallet' | 'confirming'

const STEP_LABEL: Record<Step, string> = {
  idle: 'Catat bukti ke blockchain',
  preparing: 'Menyiapkan sidik jari',
  wallet: 'Menunggu dompet',
  confirming: 'Menunggu konfirmasi jaringan',
}

function wait(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

// Sidik jari (SHA-256) laporan harian, foto lapangan, progres, tagihan termin,
// dan pembayaran semua metode dicatat ke blockchain sebagai satu akar Merkle.
// Isinya tidak pernah keluar dari server; yang tercatat hanya sidik jarinya.
export function EvidenceCard({ project }: { project: Project }) {
  const evidence = useProjectEvidence(project.id)
  const createBatch = useCreateEvidenceBatch()
  const anchorBatch = useAnchorBatch()
  const canWrite = useScopedCan('write')
  const [step, setStep] = useState<Step>('idle')
  const [problem, setProblem] = useState<string | null>(null)
  const [done, setDone] = useState<AnchorResult | null>(null)

  async function anchor() {
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
      const txHash = await sendCalldata(wallet, batch.chainId, batch.calldata)
      phase = 'confirming'
      setStep(phase)
      for (let attempt = 0; attempt < POLL_LIMIT; attempt++) {
        const result = await anchorBatch.mutateAsync({ batchId: batch.id, txHash })
        if (result.status === 'anchored') {
          setDone(result)
          return
        }
        await wait(POLL_MS)
      }
      setProblem('Transaksi belum terkonfirmasi. Tunggu sebentar lalu klik lagi; transaksi yang sama akan dipakai.')
    } catch (error) {
      setProblem(phase === 'wallet' || isWalletError(error) ? walletErrorMessage(error) : errorMessage(error))
    } finally {
      setStep('idle')
    }
  }

  return (
    <Card
      title="Jejak bukti di blockchain"
      description="Sidik jari laporan, foto, progres, tagihan termin, dan pembayaran (transfer, tunai, maupun on-chain) dicatat ke Sepolia. Bukti yang diubah setelah dicatat langsung ketahuan."
      action={
        canWrite ? (
          <Button isPending={step !== 'idle'} pendingLabel={STEP_LABEL[step]} onClick={() => void anchor()}>
            {STEP_LABEL.idle}
          </Button>
        ) : undefined
      }
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
        </div>
      ) : null}
    </Card>
  )
}

function isWalletError(error: unknown): boolean {
  return typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'number'
}
