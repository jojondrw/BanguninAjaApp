import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import { useJournalEntries, useJournalEntry } from '../../controllers/useFinance'
import { useAccounts } from '../../controllers/useProjectWorkspace'
import {
  JOURNAL_NOTE_MAX_LENGTH,
  isDateRangeValid,
  journalDetailRows,
  journalTotals,
} from '../../models/accounting'
import type { Account } from '../../models/master'
import { accountOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { ErrorNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { JOURNAL_PARAM, ledgerLink, useSearchParam } from './financeTabs'
import { JournalForm } from './JournalForm'
import { DateRangeFilter, Notice } from './parts'

const ALL_ACCOUNTS = accountOptions()
const PAGE_SIZE = 20
const LINK_CLASS = 'text-slate-900 underline-offset-4 hover:text-navy-600 hover:underline'

function JournalDetailCard({ id, accounts, onClose }: { id: string; accounts: Account[]; onClose: () => void }) {
  const entry = useJournalEntry(id)
  const cardRef = useRef<HTMLDivElement>(null)

  // Rincian muncul di bawah daftar, jadi digulir ke pandangan saat dipilih,
  // termasuk saat dibuka dari tautan di buku besar.
  useEffect(() => {
    cardRef.current?.scrollIntoView({ block: 'nearest' })
  }, [id])

  const account = (accountId: string) => accounts.find((item) => item.id === accountId)
  const close = <RowAction label="Tutup rincian" onClick={onClose} />

  if (entry.isPending || entry.isError) {
    return (
      <div ref={cardRef}>
        <Card title="Rincian jurnal" action={close}>
          {entry.isPending ? <Loading /> : <ErrorNote message={errorMessage(entry.error)} />}
        </Card>
      </div>
    )
  }

  const detail = entry.data
  const totals = journalTotals(detail.lines)

  return (
    <div ref={cardRef}>
      <Card
        title={`Jurnal ${detail.number}`}
        description={`${shortDate(detail.date)} · sumber ${detail.source || 'tidak diisi'} · dicatat ${shortDate(detail.createdAt)}`}
        action={close}
      >
        <div className="mb-4 flex flex-wrap items-center gap-3">
          {totals.debit === totals.credit ? (
            <Chip tone="bg-green-100 text-green-800">Seimbang</Chip>
          ) : (
            <Chip tone="bg-red-100 text-red-800">Tidak seimbang</Chip>
          )}
          <p className="text-sm text-slate-600">{detail.note || 'Tanpa keterangan.'}</p>
        </div>
        <Table
          rows={journalDetailRows(detail.lines)}
          emptyMessage="Jurnal ini tidak punya baris."
          columns={[
            {
              header: 'Akun',
              cell: (row) => {
                if (row.kind === 'total') {
                  return <span className="font-semibold text-slate-900">Total</span>
                }
                const item = account(row.line.accountId)
                return (
                  <Link
                    to={ledgerLink(row.line.accountId)}
                    className={`${LINK_CLASS} ${row.line.credit > 0 ? 'pl-6' : ''} inline-block`}
                    title="Lihat buku besar akun ini"
                  >
                    {item ? `${item.code} ${item.name}` : row.line.accountId.slice(0, 8)}
                  </Link>
                )
              },
            },
            {
              header: 'Debit',
              align: 'right',
              cell: (row) =>
                row.kind === 'total' ? (
                  <span className="font-semibold text-slate-900">{rupiah(row.debit)}</span>
                ) : row.line.debit > 0 ? (
                  rupiah(row.line.debit)
                ) : (
                  ''
                ),
            },
            {
              header: 'Kredit',
              align: 'right',
              cell: (row) =>
                row.kind === 'total' ? (
                  <span className="font-semibold text-slate-900">{rupiah(row.credit)}</span>
                ) : row.line.credit > 0 ? (
                  rupiah(row.line.credit)
                ) : (
                  ''
                ),
            },
          ]}
        />
        <p className="mt-3 text-xs text-slate-500">Pilih nama akun untuk membuka buku besarnya.</p>
      </Card>
    </div>
  )
}

export function JournalTab() {
  const [search, setSearch] = useState('')
  const [accountId, setAccountId] = useState('')
  const [range, setRange] = useState({ dateFrom: '', dateTo: '' })
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useSearchParam(JOURNAL_PARAM)
  const accounts = useAccounts()
  const accountItems = accounts.data?.items ?? []
  const isRangeValid = isDateRangeValid(range.dateFrom, range.dateTo)

  const entries = useJournalEntries(
    {
      search: search.trim() === '' ? undefined : search.trim(),
      accountId: accountId === '' ? undefined : accountId,
      dateFrom: range.dateFrom === '' ? undefined : range.dateFrom,
      dateTo: range.dateTo === '' ? undefined : range.dateTo,
      page,
      pageSize: PAGE_SIZE,
    },
    isRangeValid,
  )

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  const hasFilter = search.trim() !== '' || accountId !== '' || range.dateFrom !== '' || range.dateTo !== ''

  return (
    <div className="space-y-6">
      <Card
        title="Jurnal umum"
        description="Setiap jurnal minimal dua baris dengan total debit sama dengan total kredit. Jurnal yang sudah dicatat tidak bisa diubah atau dihapus; koreksi dengan jurnal pembalik."
      >
        <Toolbar>
          <ToolbarInput
            id="journal-search"
            label="Cari nomor atau keterangan jurnal"
            type="search"
            placeholder="Cari nomor atau keterangan"
            className="w-60"
            maxLength={JOURNAL_NOTE_MAX_LENGTH}
            value={search}
            onChange={(event) => filterChanged(setSearch)(event.target.value)}
          />
          <SearchSelect
            {...ALL_ACCOUNTS}
            id="journal-filter-account"
            label="Saring menurut akun"
            compact
            allowEmpty
            emptyLabel="Semua akun"
            className="w-56"
            value={accountId}
            onChange={(value) => filterChanged(setAccountId)(value)}
          />
          <DateRangeFilter
            idPrefix="journal-filter"
            dateFrom={range.dateFrom}
            dateTo={range.dateTo}
            onChange={filterChanged(setRange)}
          />
          <FormToggle isOpen={isFormOpen} openLabel="Catat jurnal" onToggle={() => setIsFormOpen((open) => !open)} />
        </Toolbar>

        {isFormOpen ? (
          <JournalForm accounts={accountItems} isLoadingAccounts={accounts.isPending} onRecorded={(entry) => setSelectedId(entry.id)} />
        ) : null}
        {accounts.isError ? (
          <Notice>
            <ErrorNote message="Daftar akun gagal dimuat, jadi pilihan akun belum tersedia." />
          </Notice>
        ) : null}

        {!isRangeValid ? <ErrorNote message="Tanggal sampai tidak boleh lebih awal dari tanggal dari." /> : null}
        {isRangeValid && entries.isPending ? <Loading /> : null}
        {isRangeValid && entries.isError ? <LoadFailed onRetry={() => entries.refetch()} /> : null}
        {isRangeValid && entries.data ? (
          <>
            <Table
              rows={entries.data.items}
              emptyMessage={
                hasFilter
                  ? 'Tidak ada jurnal yang cocok dengan penyaringan ini.'
                  : 'Belum ada jurnal. Catat yang pertama lewat tombol Catat jurnal.'
              }
              columns={[
                { header: 'Nomor', cell: (row) => <span className="font-medium text-slate-900">{row.number}</span> },
                { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                { header: 'Keterangan', cell: (row) => row.note || '-' },
                { header: 'Sumber', cell: (row) => row.source || '-' },
                { header: 'Total', align: 'right', cell: (row) => rupiah(row.total) },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Rincian"
                      isActive={row.id === selectedId}
                      onClick={() => setSelectedId(row.id === selectedId ? '' : row.id)}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={entries.data.page}
              totalPages={entries.data.totalPages}
              totalItems={entries.data.totalItems}
              unit="jurnal"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId !== '' ? (
        <JournalDetailCard id={selectedId} accounts={accountItems} onClose={() => setSelectedId('')} />
      ) : null}
    </div>
  )
}
