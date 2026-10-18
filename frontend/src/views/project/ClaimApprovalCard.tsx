import { CheckCircle2, Link2 } from 'lucide-react'
import { type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useCreateClaimLink, useProjectEvidence, useSetApprover } from '../../controllers/useEvidence'
import { useOwnerSchedule } from '../../controllers/useProjectWorkspace'
import { verifyLink, type ClaimApproval } from '../../models/evidence'
import { shortHex } from '../../models/onchain'
import type { OwnerTerm } from '../../models/ownerBilling'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field } from '../components/Form'
import { Money } from '../components/Money'
import { CopyButton } from '../pay/PayParts'

// Persetujuan pemberi kerja atas tagihan termin, ditandatangani dengan dompetnya
// (EIP-712, gratis). Tanda tangan mengikat nilai tagihan dan sidik jari bukti
// proyek saat itu, jadi tidak bisa disangkal atau dipalsukan pengembang.
export function ClaimApprovalCard({ project }: { project: Project }) {
  const schedule = useOwnerSchedule(project.id)
  const evidence = useProjectEvidence(project.id)
  const claimed = schedule.data?.terms.filter((term) => term.claimedOn) ?? []

  if (schedule.data && !schedule.data.contract) {
    return null
  }

  return (
    <Card
      title="Persetujuan pemberi kerja"
      description="Pemberi kerja membuka tautan verifikasi, memeriksa bukti, lalu menyetujui tagihan termin dengan tanda tangan dompetnya."
    >
      {schedule.isPending || evidence.isPending ? <Loading /> : null}
      {schedule.isError || evidence.isError ? (
        <LoadFailed onRetry={() => void Promise.all([schedule.refetch(), evidence.refetch()])} />
      ) : null}
      {schedule.data && evidence.data ? (
        <div className="space-y-5">
          <ApproverForm project={project} current={evidence.data.approverAddress} />
          {claimed.length === 0 ? (
            <p className="text-[13px] text-slate-500">Belum ada termin yang ditagih.</p>
          ) : (
            <ul className="divide-y divide-slate-100">
              {claimed.map((term) => (
                <ClaimRow
                  key={term.id}
                  project={project}
                  term={term}
                  approval={evidence.data.claims.find((claim) => claim.termId === term.id)}
                  hasApprover={evidence.data.approverAddress !== ''}
                />
              ))}
            </ul>
          )}
        </div>
      ) : null}
    </Card>
  )
}

function ApproverForm({ project, current }: { project: Project; current: string }) {
  const canWrite = useScopedCan('write')
  const save = useSetApprover(project.id)
  const [address, setAddress] = useState(current)

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    save.mutate(address.trim())
  }

  if (!canWrite) {
    return (
      <p className="text-[13px] text-slate-600">
        Dompet pemberi kerja: <span className="font-mono">{current ? shortHex(current) : 'belum didaftarkan'}</span>
      </p>
    )
  }

  return (
    <form onSubmit={submit} className="flex flex-wrap items-end gap-3">
      <div className="min-w-0 flex-1 basis-72">
        <Field
          id="dompet-pemberi-kerja"
          label="Alamat dompet pemberi kerja"
          hint="Hanya tanda tangan dari alamat ini yang diterima sebagai persetujuan."
          value={address}
          placeholder="0x…"
          onChange={(event) => setAddress(event.target.value)}
        />
      </div>
      <Button type="submit" variant="subtle" isPending={save.isPending} pendingLabel="Menyimpan">
        Simpan
      </Button>
      {save.isError ? (
        <div className="w-full">
          <ErrorNote message={errorMessage(save.error)} />
        </div>
      ) : null}
    </form>
  )
}

function ClaimRow({ project, term, approval, hasApprover }: {
  project: Project
  term: OwnerTerm
  approval: ClaimApproval | undefined
  hasApprover: boolean
}) {
  const canWrite = useScopedCan('write')
  const createLink = useCreateClaimLink(project.id)

  return (
    <li className="flex flex-wrap items-center justify-between gap-3 py-3">
      <div className="min-w-0">
        <p className="text-[13px] font-medium text-slate-900">
          {term.label} · <Money>{rupiah(term.netAmount)}</Money>
        </p>
        {approval?.approvedAt ? (
          <p className="flex items-center gap-1.5 text-xs text-green-800">
            <CheckCircle2 aria-hidden="true" className="size-3.5" strokeWidth={2} />
            Disetujui {shortDate(approval.approvedAt)} oleh <span className="font-mono">{shortHex(approval.signer)}</span>
            {approval.stillSame ? ', bukti masih sama' : ''}
          </p>
        ) : (
          <p className="text-xs text-slate-500">Menunggu persetujuan pemberi kerja</p>
        )}
        {approval?.approvedAt && !approval.stillSame ? (
          <p className="text-xs text-red-700">Bukti proyek berubah sejak disetujui.</p>
        ) : null}
      </div>
      {approval?.approvedAt ? null : approval ? (
        <CopyButton value={verifyLink(window.location.origin, approval.token)} label={`Salin tautan verifikasi ${term.label}`} />
      ) : canWrite ? (
        <Button
          variant="subtle"
          isPending={createLink.isPending}
          pendingLabel="Membuat"
          onClick={() => createLink.mutate(term.id)}
        >
          <Link2 aria-hidden="true" className="size-3.5" strokeWidth={2} />
          Buat tautan verifikasi
        </Button>
      ) : null}
      {!hasApprover && !approval?.approvedAt ? (
        <p className="w-full text-xs text-amber-700">Daftarkan dulu alamat dompet pemberi kerja supaya tautannya bisa dipakai menyetujui.</p>
      ) : null}
      {createLink.isError ? (
        <div className="w-full">
          <ErrorNote message={errorMessage(createLink.error)} />
        </div>
      ) : null}
    </li>
  )
}
