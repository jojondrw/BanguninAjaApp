import { useState } from 'react'
import { Link } from 'react-router-dom'

import { useLedger } from '../../controllers/useFinance'
import { useAccounts } from '../../controllers/useProjectWorkspace'
import {
  CASH_ACCOUNT_CODE,
  NORMAL_BALANCE,
  isDateRangeValid,
  ledgerRows,
  type LedgerPage,
} from '../../models/accounting'
import { ACCOUNT_TYPE_LABEL, type Account } from '../../models/master'
import { accountOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, shortDate } from '../../shared/format'
import { Card, Empty, Loading, Table } from '../components/Data'
import { Button, ErrorNote } from '../components/Form'
import { Pager, Toolbar } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { ACCOUNT_PARAM, balanceText, journalLink, useSearchParam } from './financeTabs'
import { DateRangeFilter } from './parts'

const ALL_ACCOUNTS = accountOptions()
const PAGE_SIZE = 20

function inputDate(value: string): string {
  return shortDate(`${value}T00:00:00`)
}

function periodNote(dateFrom: string, dateTo: string): string {
  if (dateFrom === '' && dateTo === '') {
    return 'semua tanggal'
  }
  if (dateFrom === '') {
    return `sampai ${inputDate(dateTo)}`
  }
  return dateTo === '' ? `sejak ${inputDate(dateFrom)}` : `${inputDate(dateFrom)} sampai ${inputDate(dateTo)}`
}

function LedgerSummary({ ledger, dateFrom, dateTo }: { ledger: LedgerPage; dateFrom: string; dateTo: string }) {
  const period = periodNote(dateFrom, dateTo)
  const cells = [
    {
      label: 'Saldo awal',
      value: balanceText(ledger.openingBalance),
      note: dateFrom === '' ? 'belum ada mutasi sebelumnya' : `sebelum ${inputDate(dateFrom)}`,
    },
    { label: 'Total debit', value: rupiah(ledger.totalDebit), note: period },
    { label: 'Total kredit', value: rupiah(ledger.totalCredit), note: period },
    { label: 'Saldo akhir', value: balanceText(ledger.closingBalance), note: 'saldo awal + debit − kredit' },
  ]

  return (
    <dl className="mb-5 grid grid-cols-2 gap-x-6 gap-y-4 rounded-xl bg-slate-50 px-4 py-3.5 lg:grid-cols-4">
      {cells.map((cell) => (
        <div key={cell.label} className="min-w-0">
          <dt className="text-xs font-medium text-slate-500">{cell.label}</dt>
          <dd className="mt-1 text-base font-semibold text-slate-900 tabular-nums">{cell.value}</dd>
          <dd className="text-xs text-slate-500">{cell.note}</dd>
        </div>
      ))}
    </dl>
  )
}

// Akun Kas diisi jurnal otomatis dari setiap transaksi kas, jadi buku besarnya
// memuat mutasi kas di samping jurnal manual yang memakai akun Kas.
function ledgerDescription(account: Account | undefined): string {
  if (!account) {
    return 'Mutasi satu akun dari semua jurnal, termasuk jurnal otomatis dari transaksi kas, urut tanggal, dengan saldo berjalan.'
  }
  const normalSide = NORMAL_BALANCE[account.type] === 'D' ? 'debit (D)' : 'kredit (K)'
  const base = `Akun ${ACCOUNT_TYPE_LABEL[account.type].toLowerCase()}, normalnya bersaldo ${normalSide}. Saldo berjalan = saldo awal + debit − kredit.`
  if (account.code !== CASH_ACCOUNT_CODE) {
    return base
  }
  return `${base} Setiap transaksi di tab Kas ikut tercatat di sini lewat jurnal otomatisnya (kas masuk di debit, kas keluar di kredit), ditambah jurnal manual yang memakai akun Kas.`
}

function emphasis(text: string) {
  return <span className="font-semibold text-slate-900">{text}</span>
}

export function LedgerTab() {
  const [accountId, setAccountId] = useSearchParam(ACCOUNT_PARAM)
  const [range, setRange] = useState({ dateFrom: '', dateTo: '' })
  const [page, setPage] = useState(1)
  const accounts = useAccounts()
  const accountItems = accounts.data?.items ?? []
  const account = accountItems.find((item) => item.id === accountId)
  const isRangeValid = isDateRangeValid(range.dateFrom, range.dateTo)

  const ledger = useLedger(
    {
      accountId,
      dateFrom: range.dateFrom === '' ? undefined : range.dateFrom,
      dateTo: range.dateTo === '' ? undefined : range.dateTo,
      page,
      pageSize: PAGE_SIZE,
    },
    isRangeValid,
  )

  return (
    <Card
      title={account ? `Buku besar ${account.code} ${account.name}` : 'Buku besar'}
      description={ledgerDescription(account)}
    >
      <Toolbar>
        <SearchSelect
          {...ALL_ACCOUNTS}
          id="ledger-account"
          label="Akun buku besar"
          placeholder="Cari kode atau nama akun"
          compact
          className="w-64"
          value={accountId}
          onChange={(value) => {
            setAccountId(value)
            setPage(1)
          }}
        />
        <DateRangeFilter
          idPrefix="ledger"
          dateFrom={range.dateFrom}
          dateTo={range.dateTo}
          onChange={(next) => {
            setRange(next)
            setPage(1)
          }}
        />
      </Toolbar>

      {accounts.isError ? (
        <div className="mb-4 space-y-3">
          <ErrorNote message="Daftar akun gagal dimuat." />
          <Button variant="subtle" onClick={() => accounts.refetch()}>
            Muat ulang akun
          </Button>
        </div>
      ) : null}

      {accountId === '' ? (
        <Empty message="Pilih akun untuk melihat mutasi dan saldonya. Buku besar diisi dari semua jurnal, termasuk jurnal otomatis dari setiap transaksi kas." />
      ) : null}
      {accountId !== '' && !isRangeValid ? (
        <ErrorNote message="Tanggal sampai tidak boleh lebih awal dari tanggal dari." />
      ) : null}
      {accountId !== '' && isRangeValid && ledger.isPending ? <Loading /> : null}
      {accountId !== '' && isRangeValid && ledger.isError ? (
        <div className="space-y-3">
          <ErrorNote message={errorMessage(ledger.error)} />
          <Button variant="subtle" onClick={() => ledger.refetch()}>
            Coba lagi
          </Button>
        </div>
      ) : null}
      {accountId !== '' && isRangeValid && ledger.data ? (
        <div
          aria-busy={ledger.isPlaceholderData}
          className={`transition-opacity ${ledger.isPlaceholderData ? 'opacity-60' : ''}`}
        >
          <LedgerSummary ledger={ledger.data} dateFrom={range.dateFrom} dateTo={range.dateTo} />
          <Table
            rows={ledgerRows(ledger.data)}
            emptyMessage="Belum ada jurnal yang memakai akun ini pada periode tersebut."
            columns={[
              {
                header: 'Tanggal',
                cell: (row) => (row.kind === 'line' ? shortDate(row.line.date) : ''),
              },
              {
                header: 'Jurnal',
                cell: (row) => {
                  if (row.kind === 'opening') {
                    return emphasis('Saldo awal')
                  }
                  if (row.kind === 'closing') {
                    return emphasis('Total dan saldo akhir')
                  }
                  return (
                    <Link
                      to={journalLink(row.line.journalEntryId)}
                      className="font-medium text-slate-900 underline-offset-4 hover:text-navy-600 hover:underline"
                      title="Lihat rincian jurnal"
                    >
                      {row.line.number}
                    </Link>
                  )
                },
              },
              { header: 'Keterangan', cell: (row) => (row.kind === 'line' ? row.line.note || '-' : '') },
              {
                header: 'Debit',
                align: 'right',
                cell: (row) => {
                  if (row.kind === 'closing') {
                    return emphasis(rupiah(row.debit))
                  }
                  return row.kind === 'line' && row.line.debit > 0 ? rupiah(row.line.debit) : ''
                },
              },
              {
                header: 'Kredit',
                align: 'right',
                cell: (row) => {
                  if (row.kind === 'closing') {
                    return emphasis(rupiah(row.credit))
                  }
                  return row.kind === 'line' && row.line.credit > 0 ? rupiah(row.line.credit) : ''
                },
              },
              {
                header: 'Saldo',
                align: 'right',
                cell: (row) =>
                  row.kind === 'line' ? balanceText(row.line.balance) : emphasis(balanceText(row.balance)),
              },
            ]}
          />
          <Pager
            page={ledger.data.page}
            totalPages={ledger.data.totalPages}
            totalItems={ledger.data.totalItems}
            unit="baris jurnal"
            onChange={setPage}
          />
        </div>
      ) : null}
    </Card>
  )
}
