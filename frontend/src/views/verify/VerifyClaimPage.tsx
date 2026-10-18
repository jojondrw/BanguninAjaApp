import { CheckCircle2, PenLine } from 'lucide-react'
import { type ReactNode, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'

import { useApproveClaim, usePublicClaim } from '../../controllers/useEvidence'
import { tamperedCount, type PublicClaim } from '../../models/evidence'
import { shortHex } from '../../models/onchain'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, shortDate } from '../../shared/format'
import { BrandName } from '../components/Brand'
import { Loading } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { Money } from '../components/Money'
import { EvidenceList, EvidenceSummaryLine } from '../evidence/EvidenceList'
import { injectedWallet, metamaskDeepLink, signTypedData, walletErrorMessage } from '../pay/wallet'

// Halaman yang dibuka pemberi kerja tanpa akun: memeriksa bukti proyek di
// blockchain, lalu menyetujui tagihan termin dengan tanda tangan dompet.
export function VerifyClaimPage() {
  const token = useParams().token ?? ''
  const claim = usePublicClaim(token)

  useEffect(() => {
    document.title = 'Verifikasi tagihan · BanguninAja'
  }, [])

  return (
    <div className="min-h-screen bg-slate-100">
      <header className="mx-auto flex max-w-2xl items-center px-4 pt-5 sm:px-6">
        <BrandName />
      </header>
      <main className="mx-auto max-w-2xl px-4 py-6 sm:px-6">
        <div className="rounded-2xl bg-white p-5 shadow-hairline sm:p-6 motion-safe:animate-enter">
          {claim.isPending ? <Loading label="Mengambil data tagihan..." /> : null}
          {claim.isError ? (
            <div className="space-y-3">
              <h1 className="text-lg font-semibold text-slate-900">Tautan verifikasi tidak bisa dibuka</h1>
              <ErrorNote message={errorMessage(claim.error)} />
            </div>
          ) : null}
          {claim.data ? <ClaimBody token={token} claim={claim.data} /> : null}
        </div>
        <p className="mt-4 text-center text-xs text-slate-500">
          Menyetujui hanya meminta tanda tangan dompet, tanpa transaksi dan tanpa biaya. BanguninAja tidak pernah meminta kunci
          privat atau frasa pemulihan.
        </p>
      </main>
    </div>
  )
}

function ClaimBody({ token, claim }: { token: string; claim: PublicClaim }) {
  return (
    <div className="space-y-6">
      <div>
        <p className="text-[13px] text-slate-500">
          {claim.projectCode} · {claim.contractNumber}
        </p>
        <h1 className="mt-0.5 text-xl font-semibold tracking-[-0.02em] text-slate-900">
          {claim.label} {claim.projectName}
        </h1>
        <p className="mt-1 text-[13px] text-slate-600">Ditagihkan kepada {claim.ownerName}</p>
      </div>

      <dl className="grid grid-cols-2 gap-3">
        <Fact label="Progres saat ditagih">{claim.progress}%</Fact>
        <Fact label="Tanggal tagihan">{shortDate(claim.claimedOn)}</Fact>
        <Fact label="Nilai termin">
          <Money>{rupiah(claim.grossAmount)}</Money>
        </Fact>
        <Fact label="Dibayar sekarang">
          <Money>{rupiah(claim.netAmount)}</Money>
        </Fact>
      </dl>

      <section aria-labelledby="bukti" className="space-y-3">
        <h2 id="bukti" className="text-[15px] font-semibold text-slate-900">
          Bukti proyek
        </h2>
        <p className="text-[13px] text-slate-600">
          Setiap laporan harian, foto lapangan, progres, tagihan, dan pembayaran punya sidik jari yang dicatat di blockchain.
          Kalau ada yang diubah atau dihapus setelah dicatat, statusnya merah di bawah ini.
        </p>
        <EvidenceSummaryLine summary={claim.summary} />
        <EvidenceList items={claim.items} />
      </section>

      <Approval token={token} claim={claim} />
    </div>
  )
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="rounded-xl bg-slate-50 px-3.5 py-3">
      <dt className="text-xs text-slate-500">{label}</dt>
      <dd className="mt-0.5 text-[15px] font-semibold text-slate-900 tabular-nums">{children}</dd>
    </div>
  )
}

function Approval({ token, claim }: { token: string; claim: PublicClaim }) {
  const approve = useApproveClaim(token)
  const [problem, setProblem] = useState<string | null>(null)
  const [isSigning, setIsSigning] = useState(false)
  const wallet = injectedWallet()
  const tampered = tamperedCount(claim.summary)

  if (claim.approval.approvedAt) {
    return (
      <div className="flex items-start gap-2.5 rounded-xl bg-green-50 px-3.5 py-3 text-[13px] text-green-800">
        <CheckCircle2 aria-hidden="true" className="mt-0.5 size-4 shrink-0" strokeWidth={1.8} />
        <p>
          Disetujui {shortDate(claim.approval.approvedAt)} dengan tanda tangan dompet{' '}
          <span className="font-mono">{shortHex(claim.approval.signer)}</span>.
          {claim.approval.stillSame ? ' Bukti proyek masih sama seperti saat disetujui.' : ' Bukti proyek sudah berubah sejak disetujui.'}
        </p>
      </div>
    )
  }

  async function sign() {
    setProblem(null)
    if (!wallet) {
      return
    }
    setIsSigning(true)
    try {
      const signature = await signTypedData(wallet, claim.typedData)
      await approve.mutateAsync(signature)
    } catch (error) {
      setProblem(approve.isError ? errorMessage(error) : walletErrorMessage(error))
    } finally {
      setIsSigning(false)
    }
  }

  return (
    <section aria-labelledby="setujui" className="space-y-3 border-t border-slate-100 pt-5">
      <h2 id="setujui" className="text-[15px] font-semibold text-slate-900">
        Setujui tagihan
      </h2>
      {claim.approverAddress ? (
        <p className="text-[13px] text-slate-600">
          Tanda tangan dari dompet <span className="font-mono">{shortHex(claim.approverAddress)}</span> akan mengikat nilai
          tagihan dan sidik jari seluruh bukti di atas.
        </p>
      ) : (
        <ErrorNote message="Pengembang belum mendaftarkan dompet pemberi kerja untuk proyek ini, jadi tagihan belum bisa disetujui." />
      )}
      {tampered > 0 ? (
        <ErrorNote message="Ada bukti yang berubah atau dihapus setelah dicatat. Tanyakan ke pengembang sebelum menyetujui." />
      ) : null}
      {problem ? <ErrorNote message={problem} /> : null}
      {approve.isSuccess ? <SuccessNote message="Tagihan disetujui." /> : null}
      {wallet ? (
        <Button isPending={isSigning} pendingLabel="Menunggu dompet" onClick={() => void sign()}>
          <PenLine aria-hidden="true" className="size-3.5" strokeWidth={2} />
          Setujui dengan dompet
        </Button>
      ) : (
        <p className="text-[13px] text-slate-600">
          Buka halaman ini di peramban yang punya MetaMask, atau{' '}
          <a href={metamaskDeepLink(window.location.href)} className="font-medium text-navy-600 underline underline-offset-4">
            buka di aplikasi MetaMask
          </a>
          .
        </p>
      )}
    </section>
  )
}
