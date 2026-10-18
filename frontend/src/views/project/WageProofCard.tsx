import { ShieldCheck } from 'lucide-react'
import { useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useCreateWageProof, useWageProofs } from '../../controllers/useEvidence'
import type { Project } from '../../models/project'
import { periodLabel, receiptLink, recentPeriods, wageProofLink } from '../../models/wageProof'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, SelectField } from '../components/Form'
import { Money } from '../components/Money'
import { CopyButton } from '../pay/PayParts'

const PERIOD_CHOICES = 6

// Bukti zero-knowledge: semua pekerja proyek pada satu periode dibayar paling
// sedikit upah minimum kota proyek. Pemberi kerja bisa memverifikasinya tanpa
// melihat gaji siapa pun, dan tiap pekerja dapat kuitansi untuk barisnya sendiri.
export function WageProofCard({ project }: { project: Project }) {
  const proofs = useWageProofs(project.id)
  const create = useCreateWageProof(project.id)
  const canWrite = useScopedCan('write')
  const periods = recentPeriods(new Date(), PERIOD_CHOICES)
  const [period, setPeriod] = useState(periods[1])
  const below = create.data?.below ?? []

  return (
    <Card
      title="Bukti upah minimum"
      description="Membuktikan ke pemberi kerja bahwa semua pekerja proyek dibayar paling sedikit UMK, tanpa membuka gaji siapa pun (zero-knowledge proof)."
    >
      {canWrite ? (
        <div className="mb-5 flex flex-wrap items-end gap-3">
          <div className="w-52">
            <SelectField id="periode-upah" label="Periode penggajian" value={period} onChange={(event) => setPeriod(event.target.value)}>
              {periods.map((value) => (
                <option key={value} value={value}>
                  {periodLabel(value)}
                </option>
              ))}
            </SelectField>
          </div>
          <Button isPending={create.isPending} pendingLabel="Membuat bukti" onClick={() => create.mutate(period)}>
            <ShieldCheck aria-hidden="true" className="size-3.5" strokeWidth={2} />
            Buat bukti
          </Button>
        </div>
      ) : null}

      {create.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(create.error)} />
        </div>
      ) : null}
      {below.length > 0 ? (
        <div className="mb-5 space-y-2 rounded-xl bg-red-50 px-4 py-3 text-[13px] text-red-800">
          <p className="font-semibold">Bukti tidak bisa dibuat: {number(below.length)} pekerja dibayar di bawah UMK.</p>
          <ul className="space-y-1">
            {below.map((worker) => (
              <li key={worker.name}>
                {worker.name}: <Money>{rupiah(worker.pay)}</Money> untuk {number(worker.days)} hari, minimal{' '}
                <Money>{rupiah(worker.required)}</Money>
              </li>
            ))}
          </ul>
          <p className="text-xs text-red-700">Hanya Anda yang melihat daftar ini. Perbaiki penggajiannya, lalu buat bukti lagi.</p>
        </div>
      ) : null}

      {proofs.isPending ? <Loading /> : null}
      {proofs.isError ? <LoadFailed onRetry={() => proofs.refetch()} /> : null}
      {proofs.data && proofs.data.length === 0 ? (
        <p className="text-[13px] text-slate-500">Belum ada bukti upah minimum untuk proyek ini.</p>
      ) : null}
      {proofs.data && proofs.data.length > 0 ? (
        <ul className="divide-y divide-slate-100">
          {proofs.data.map((proof) => (
            <li key={proof.id} className="space-y-3 py-4 first:pt-0">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-[14px] font-semibold text-slate-900">{periodLabel(proof.period)}</p>
                  <p className="text-xs text-slate-500">
                    {number(proof.workerCount)} pekerja, semua ≥ UMK {proof.regionName} <Money>{rupiah(proof.minimumWage)}</Money>
                  </p>
                </div>
                <CopyButton value={wageProofLink(window.location.origin, proof.token)} label={`Salin tautan bukti ${proof.period} untuk pemberi kerja`} />
              </div>
              <details className="text-[13px]">
                <summary className="cursor-pointer font-medium text-navy-600">Kuitansi untuk tiap pekerja</summary>
                <ul className="mt-2 space-y-1.5">
                  {proof.receipts.map((receipt) => (
                    <li key={receipt.token} className="flex items-center justify-between gap-3">
                      <span className="text-slate-700">{receipt.name}</span>
                      <CopyButton value={receiptLink(window.location.origin, receipt.token)} label={`Salin tautan kuitansi ${receipt.name}`} />
                    </li>
                  ))}
                </ul>
              </details>
            </li>
          ))}
        </ul>
      ) : null}
    </Card>
  )
}
