import { CheckCircle2, ShieldAlert, ShieldCheck } from 'lucide-react'
import { type ReactNode, useEffect } from 'react'
import { useParams } from 'react-router-dom'

import { usePublicReceipt, usePublicWageProof } from '../../controllers/useEvidence'
import { shortHex } from '../../models/onchain'
import { periodLabel, type AnchorStatus } from '../../models/wageProof'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiah, shortDate } from '../../shared/format'
import { BrandName } from '../components/Brand'
import { Loading } from '../components/Data'
import { ErrorNote } from '../components/Form'
import { Money } from '../components/Money'
import { TxLink } from '../pay/PayParts'

function PublicLayout({ title, children }: { title: string; children: ReactNode }) {
  useEffect(() => {
    document.title = `${title} · BanguninAja`
  }, [title])

  return (
    <div className="min-h-screen bg-slate-100">
      <header className="mx-auto flex max-w-xl items-center px-4 pt-5 sm:px-6">
        <BrandName />
      </header>
      <main className="mx-auto max-w-xl px-4 py-6 sm:px-6">
        <div className="space-y-5 rounded-2xl bg-white p-5 shadow-hairline sm:p-6 motion-safe:animate-enter">{children}</div>
      </main>
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

function AnchorLine({ anchor }: { anchor: AnchorStatus }) {
  if (anchor.status === 'anchored' && anchor.anchoredAt) {
    return (
      <p className="flex flex-wrap items-center gap-2 text-[13px] text-slate-600">
        Akar penggajian tercatat di blockchain {shortDate(anchor.anchoredAt)}: <TxLink url={anchor.txUrl} hash={anchor.txHash} />
      </p>
    )
  }
  return <p className="text-[13px] text-slate-500">Akar penggajian belum dicatat ke blockchain oleh pengembang.</p>
}

function Verdict({ ok, yes, no }: { ok: boolean; yes: string; no: string }) {
  const Icon = ok ? ShieldCheck : ShieldAlert
  return (
    <div className={`flex items-start gap-2.5 rounded-xl px-3.5 py-3 text-[14px] ${ok ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-700'}`}>
      <Icon aria-hidden="true" className="mt-0.5 size-5 shrink-0" strokeWidth={1.8} />
      <p className="font-medium">{ok ? yes : no}</p>
    </div>
  )
}

// Halaman pemberi kerja: memeriksa bukti tanpa melihat gaji siapa pun.
export function VerifyWagePage() {
  const token = useParams().token ?? ''
  const proof = usePublicWageProof(token)

  return (
    <PublicLayout title="Bukti upah minimum">
      {proof.isPending ? <Loading label="Memeriksa bukti..." /> : null}
      {proof.isError ? <ErrorNote message={errorMessage(proof.error)} /> : null}
      {proof.data ? (
        <>
          <div>
            <p className="text-[13px] text-slate-500">{proof.data.projectCode}</p>
            <h1 className="mt-0.5 text-xl font-semibold tracking-[-0.02em] text-slate-900">
              Upah pekerja {proof.data.projectName}
            </h1>
            <p className="mt-1 text-[13px] text-slate-600">Penggajian {periodLabel(proof.data.period)}</p>
          </div>
          <Verdict
            ok={proof.data.verified}
            yes={`Terbukti: semua ${number(proof.data.workerCount)} pekerja dibayar paling sedikit UMK ${proof.data.regionName}.`}
            no="Bukti tidak valid. Hubungi pengembang."
          />
          <dl className="grid grid-cols-2 gap-3">
            <Fact label={`UMK ${proof.data.regionName} per bulan`}>
              <Money>{rupiah(proof.data.minimumWage)}</Money>
            </Fact>
            <Fact label="Setara per hari kerja (÷ 25)">
              <Money>{rupiah(proof.data.dailyMinimum)}</Money>
            </Fact>
          </dl>
          <p className="text-[13px] leading-relaxed text-slate-600">
            Bukti ini dibuat dengan zero-knowledge proof (Groth16). Ia membuktikan setiap baris penggajian di akar{' '}
            <span className="font-mono">{shortHex(proof.data.root)}</span> memenuhi upah minimum, tanpa membuka gaji, nama, atau
            jumlah hari kerja siapa pun. Pekerja harian dibandingkan per hari kerja.
          </p>
          <AnchorLine anchor={proof.data.anchor} />
          <p className="text-xs text-slate-500">Sumber UMK: {proof.data.wageSource}.</p>
          <details className="text-xs text-slate-500">
            <summary className="cursor-pointer font-medium text-navy-600">Data untuk verifikasi mandiri</summary>
            <p className="mt-2 break-all font-mono">Bukti: {proof.data.proofHex}</p>
            <p className="mt-2 break-all font-mono">Kunci verifikasi: {proof.data.verifyingKeyHex.slice(0, 200)}…</p>
          </details>
        </>
      ) : null}
    </PublicLayout>
  )
}

// Halaman pekerja: memastikan upahnya sendiri ikut dihitung dalam bukti.
export function WageReceiptPage() {
  const token = useParams().token ?? ''
  const receipt = usePublicReceipt(token)

  return (
    <PublicLayout title="Kuitansi upah">
      {receipt.isPending ? <Loading label="Memeriksa kuitansi..." /> : null}
      {receipt.isError ? <ErrorNote message={errorMessage(receipt.error)} /> : null}
      {receipt.data ? (
        <>
          <div>
            <p className="text-[13px] text-slate-500">{receipt.data.projectCode}</p>
            <h1 className="mt-0.5 text-xl font-semibold tracking-[-0.02em] text-slate-900">Kuitansi upah {receipt.data.name}</h1>
            <p className="mt-1 text-[13px] text-slate-600">
              {receipt.data.projectName}, penggajian {periodLabel(receipt.data.period)}
            </p>
          </div>
          <dl className="grid grid-cols-2 gap-3">
            <Fact label="Upah yang tercatat">
              <Money>{rupiah(receipt.data.pay)}</Money>
            </Fact>
            <Fact label="Hari kerja dihitung">{number(receipt.data.days)}</Fact>
          </dl>
          <Verdict
            ok={receipt.data.included}
            yes="Upah Anda ikut dihitung dalam bukti upah minimum proyek ini."
            no="Upah Anda tidak cocok dengan bukti. Laporkan ke pengawas atau pemberi kerja."
          />
          <p className="text-[13px] leading-relaxed text-slate-600">
            Kalau upah yang tercatat di atas berbeda dengan yang Anda terima, laporkan. Kuitansi ini bukti bahwa angka inilah yang
            dipakai pengembang.
          </p>
          <AnchorLine anchor={receipt.data.anchor} />
          <p className="flex items-center gap-1.5 text-xs text-slate-500">
            <CheckCircle2 aria-hidden="true" className="size-3.5" strokeWidth={2} />
            Simpan tautan ini. Hanya Anda yang menerimanya.
          </p>
        </>
      ) : null}
    </PublicLayout>
  )
}
