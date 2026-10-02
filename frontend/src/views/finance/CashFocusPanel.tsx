import { BookOpen } from 'lucide-react'
import type { ReactNode } from 'react'

import { useCashTransaction } from '../../controllers/useFinance'
import { accountLabel } from '../../models/accounting'
import { CASH_TYPE_LABEL, CASH_TYPE_TONE, type CashTransaction } from '../../models/finance'
import type { Account } from '../../models/master'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { Loading } from '../components/Data'
import { ErrorNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip } from '../components/RecordControls'
import { journalLink } from './financeTabs'
import { RowActions, RowLink, SignedAmount } from './parts'

function Detail({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs font-medium text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm text-slate-900">{children}</dd>
    </div>
  )
}

// Transaksi yang dibuka dari tautan "Lihat transaksi kas" di rincian jurnal.
// Daftar kas berhalaman dan bisa tersaring, jadi transaksinya ditampilkan
// sendiri di atas daftar, bukan dicari di halaman tabel.
export function CashFocusPanel({ id, accounts, projectName, isEditing, onEdit, onClose }: {
  id: string
  accounts: Account[]
  projectName: (projectId: string | null) => string
  isEditing: boolean
  onEdit: (transaction: CashTransaction) => void
  onClose: () => void
}) {
  const transaction = useCashTransaction(id)
  const data = transaction.data

  return (
    <section aria-labelledby="cash-focus-title" className="mb-5 rounded-xl bg-slate-50 px-4 py-3.5">
      <div className="mb-3 flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 id="cash-focus-title" className="text-sm font-semibold text-slate-900">
            Transaksi kas dari jurnal
          </h3>
          <p className="text-xs text-slate-500">Jurnal otomatis yang Anda buka berasal dari transaksi ini.</p>
        </div>
        <RowActions>
          {data ? <RowAction label="Ubah" isActive={isEditing} onClick={() => onEdit(data)} /> : null}
          {data?.journalEntryId ? (
            <RowLink to={journalLink(data.journalEntryId)} label="Lihat jurnal" icon={BookOpen} />
          ) : null}
          <RowAction label="Tutup" onClick={onClose} />
        </RowActions>
      </div>

      {transaction.isPending ? <Loading /> : null}
      {transaction.isError ? <ErrorNote message={errorMessage(transaction.error)} /> : null}
      {data ? (
        <dl className="grid grid-cols-2 gap-x-6 gap-y-4 lg:grid-cols-3">
          <Detail label="Tanggal">{shortDate(data.date)}</Detail>
          <Detail label="Jenis">
            <Chip tone={CASH_TYPE_TONE[data.type]}>{CASH_TYPE_LABEL[data.type]}</Chip>
          </Detail>
          <Detail label="Jumlah">
            <SignedAmount type={data.type} amount={data.amount} />
          </Detail>
          <Detail label="Akun lawan">{accountLabel(accounts, data.accountId)}</Detail>
          <Detail label="Proyek">{projectName(data.projectId)}</Detail>
          <Detail label="Keterangan">{data.note || '-'}</Detail>
        </dl>
      ) : null}
    </section>
  )
}
