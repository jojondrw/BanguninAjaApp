import { ShieldAlert, ShieldCheck, ShieldQuestion } from 'lucide-react'
import { useState } from 'react'

import {
  EVIDENCE_KIND_LABEL,
  EVIDENCE_STATUS_LABEL,
  EVIDENCE_STATUS_TONE,
  tamperedCount,
  type EvidenceItem,
  type EvidenceSummary,
} from '../../models/evidence'
import { number, shortDate } from '../../shared/format'
import { Chip } from '../components/RecordControls'
import { TxLink } from '../pay/PayParts'

const COLLAPSED_ROWS = 8

export function EvidenceSummaryLine({ summary }: { summary: EvidenceSummary }) {
  const tampered = tamperedCount(summary)
  const Icon = tampered > 0 ? ShieldAlert : summary.pending > 0 ? ShieldQuestion : ShieldCheck
  const tone = tampered > 0 ? 'bg-red-50 text-red-700' : summary.pending > 0 ? 'bg-slate-50 text-slate-700' : 'bg-green-50 text-green-800'

  return (
    <div className={`flex items-start gap-2.5 rounded-xl px-3.5 py-3 text-[13px] ${tone}`}>
      <Icon aria-hidden="true" className="mt-0.5 size-4 shrink-0" strokeWidth={1.8} />
      <p>
        <span className="font-semibold">{number(summary.anchored)} bukti tercatat dan utuh</span>
        {summary.pending > 0 ? `, ${number(summary.pending)} belum dicatat` : ''}
        {tampered > 0 ? `, ${number(tampered)} berubah atau dihapus setelah dicatat` : ''}.
      </p>
    </div>
  )
}

export function EvidenceList({ items }: { items: EvidenceItem[] }) {
  const [showAll, setShowAll] = useState(false)
  if (items.length === 0) {
    return <p className="text-[13px] text-slate-500">Belum ada bukti untuk proyek ini.</p>
  }
  const ordered = [...items].sort((left, right) => rank(left) - rank(right))
  const visible = showAll ? ordered : ordered.slice(0, COLLAPSED_ROWS)

  return (
    <div>
      <ul className="divide-y divide-slate-100">
        {visible.map((item) => (
          <li key={`${item.kind}-${item.subjectId}`} className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1.5 py-2.5">
            <div className="min-w-0">
              <p className="text-[13px] font-medium text-slate-900">
                {EVIDENCE_KIND_LABEL[item.kind]}
                {item.date ? <span className="font-normal text-slate-500"> · {shortDate(item.date)}</span> : null}
              </p>
              {item.label ? <p className="line-clamp-1 text-xs text-slate-500">{item.label}</p> : null}
            </div>
            <div className="flex items-center gap-3">
              {item.txHash ? <TxLink url={item.txUrl} hash={item.txHash} /> : null}
              <Chip tone={EVIDENCE_STATUS_TONE[item.status]}>{EVIDENCE_STATUS_LABEL[item.status]}</Chip>
            </div>
          </li>
        ))}
      </ul>
      {items.length > COLLAPSED_ROWS ? (
        <button
          type="button"
          onClick={() => setShowAll((current) => !current)}
          className="mt-2 text-[13px] font-medium text-navy-600 hover:underline"
        >
          {showAll ? 'Tampilkan lebih sedikit' : `Tampilkan semua ${number(items.length)} bukti`}
        </button>
      ) : null}
    </div>
  )
}

function rank(item: EvidenceItem): number {
  return { changed: 0, deleted: 0, pending: 1, anchored: 2 }[item.status]
}
