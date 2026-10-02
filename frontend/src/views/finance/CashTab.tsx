import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../../controllers/useErp'
import {
  useCashTransactionList,
  useDeleteCashTransaction,
  useRecordCashTransaction,
  useUpdateCashTransaction,
} from '../../controllers/useFinance'
import { useAccounts } from '../../controllers/useProjectWorkspace'
import {
  accountLabel,
  cashFormFrom,
  emptyCashForm,
  isDateRangeValid,
  type CashFormValues,
} from '../../models/accounting'
import {
  CASH_NOTE_MAX_LENGTH,
  CASH_TYPE_LABEL,
  CASH_TYPE_TONE,
  type CashTransaction,
  type CashType,
} from '../../models/finance'
import { accountGroups, type Account } from '../../models/master'
import { accountOptions, projectOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, shortDate } from '../../shared/format'
import { todayDate } from '../../shared/localDate'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { FilterChips, RowAction } from '../components/ListTools'
import { Chip, FormPanel, FormToggle, Pager, SelectField, Toolbar } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { amountHint } from './financeTabs'
import { AccountOptions, ConfirmDelete, DateRangeFilter, Notice, RowActions, SignedAmount } from './parts'

const ALL_ACCOUNTS = accountOptions()
const PAGE_SIZE = 20
const OPTION_LIMIT = 100

const TYPE_FILTERS: { value: CashType | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'in', label: 'Kas masuk' },
  { value: 'out', label: 'Kas keluar' },
]

function transactionName(transaction: CashTransaction): string {
  const direction = transaction.type === 'in' ? 'kas masuk' : 'kas keluar'
  return `${direction} ${rupiah(transaction.amount)} tanggal ${shortDate(transaction.date)}`
}

function CashForm({ transaction, accounts, isLoadingAccounts, onUpdated }: {
  transaction: CashTransaction | null
  accounts: Account[]
  isLoadingAccounts: boolean
  onUpdated?: (message: string) => void
}) {
  const [values, setValues] = useState<CashFormValues>(() =>
    transaction ? cashFormFrom(transaction) : emptyCashForm(todayDate()),
  )
  const recordTransaction = useRecordCashTransaction()
  const updateTransaction = useUpdateCashTransaction()
  const mutation = transaction ? updateTransaction : recordTransaction
  const groups = accountGroups(accounts, values.type)

  const update =
    (key: keyof CashFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (transaction) {
      updateTransaction.mutate(
        { id: transaction.id, values },
        { onSuccess: (saved) => onUpdated?.(`Transaksi ${transactionName(saved)} diperbarui.`) },
      )
      return
    }
    recordTransaction.mutate(values, {
      onSuccess: () => setValues((current) => ({ ...current, amount: '', note: '' })),
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        {transaction ? (
          <p className="text-sm text-slate-600">
            Mengubah transaksi <span className="font-medium text-slate-900">{transactionName(transaction)}</span>.
          </p>
        ) : null}
        <div className="grid gap-4 md:grid-cols-3">
          <Field id="cash-date" label="Tanggal" type="date" required value={values.date} onChange={update('date')} />
          <SelectField id="cash-type" label="Jenis" value={values.type} onChange={update('type')}>
            <option value="out">Kas keluar</option>
            <option value="in">Kas masuk</option>
          </SelectField>
          <Field
            id="cash-amount"
            label="Jumlah"
            type="number"
            inputMode="numeric"
            min={1}
            step={1}
            required
            placeholder="0"
            hint={amountHint(values.amount, 'Dalam Rupiah, lebih dari nol')}
            value={values.amount}
            onChange={update('amount')}
          />
          <SelectField
            id="cash-account"
            label="Akun lawan"
            required
            disabled={isLoadingAccounts}
            hint="Kas keluar biasanya ke akun Beban, kas masuk ke akun Pendapatan"
            value={values.accountId}
            onChange={update('accountId')}
          >
            <AccountOptions groups={groups} placeholder={isLoadingAccounts ? 'Memuat akun...' : 'Pilih akun'} />
          </SelectField>
          <SearchSelect
            {...projectOptions}
            id="cash-project"
            label="Proyek (opsional)"
            placeholder="Cari proyek"
            hint="Kas keluar proyek dihitung sebagai realisasi anggaran"
            allowEmpty
            emptyLabel="Tanpa proyek, kantor pusat"
            value={values.projectId}
            onChange={(projectId) => setValues((current) => ({ ...current, projectId }))}
          />
          <Field
            id="cash-note"
            label="Keterangan (opsional)"
            placeholder="Pembayaran termin 2 kontraktor"
            autoComplete="off"
            maxLength={CASH_NOTE_MAX_LENGTH}
            value={values.note}
            onChange={update('note')}
          />
        </div>

        <Button
          type="submit"
          isPending={mutation.isPending}
          pendingLabel={transaction ? 'Menyimpan perubahan' : 'Menyimpan transaksi'}
        >
          {transaction ? 'Simpan perubahan' : 'Simpan transaksi'}
        </Button>

        {mutation.isError ? <ErrorNote message={errorMessage(mutation.error)} /> : null}
        {recordTransaction.isSuccess ? (
          <SuccessNote message={`Transaksi ${transactionName(recordTransaction.data)} tercatat.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

export function CashTab() {
  const [type, setType] = useState<CashType | ''>('')
  const [projectId, setProjectId] = useState('')
  const [accountId, setAccountId] = useState('')
  const [range, setRange] = useState({ dateFrom: '', dateTo: '' })
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<CashTransaction | null>(null)
  const [confirmingId, setConfirmingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const accounts = useAccounts()
  const deleteTransaction = useDeleteCashTransaction()
  const projectItems = projects.data?.items ?? []
  const accountItems = accounts.data?.items ?? []
  const isRangeValid = isDateRangeValid(range.dateFrom, range.dateTo)

  const transactions = useCashTransactionList(
    {
      type: type === '' ? undefined : type,
      projectId: projectId === '' ? undefined : projectId,
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

  const projectName = (id: string | null) =>
    id === null ? 'Kantor pusat' : (projectItems.find((project) => project.id === id)?.name ?? id.slice(0, 8))

  const startEdit = (transaction: CashTransaction) => {
    setIsCreating(false)
    setNotice(null)
    setEditing((current) => (current?.id === transaction.id ? null : transaction))
  }

  const remove = (transaction: CashTransaction) => {
    setNotice(null)
    deleteTransaction.mutate(transaction.id, {
      onSuccess: () => {
        setNotice(`Transaksi ${transactionName(transaction)} dihapus.`)
        if (editing?.id === transaction.id) {
          setEditing(null)
        }
        if (transactions.data?.items.length === 1 && page > 1) {
          setPage(page - 1)
        }
      },
      onSettled: () => setConfirmingId(null),
    })
  }

  const hasFilter = type !== '' || projectId !== '' || accountId !== '' || range.dateFrom !== '' || range.dateTo !== ''

  return (
    <Card
      title="Transaksi kas"
      description="Kas masuk bertanda +, kas keluar bertanda −. Transaksi kas tidak membuat jurnal otomatis; catat jurnalnya di tab Jurnal bila perlu."
    >
      <div className="mb-3">
        <FilterChips filters={TYPE_FILTERS} active={type} onChange={filterChanged(setType)} label="Saring menurut jenis kas" />
      </div>
      <Toolbar>
        <SearchSelect
          {...projectOptions}
          id="cash-filter-project"
          label="Saring menurut proyek"
          compact
          allowEmpty
          emptyLabel="Semua proyek"
          className="w-52"
          value={projectId}
          onChange={(value) => filterChanged(setProjectId)(value)}
        />
        <SearchSelect
          {...ALL_ACCOUNTS}
          id="cash-filter-account"
          label="Saring menurut akun"
          compact
          allowEmpty
          emptyLabel="Semua akun"
          className="w-56"
          value={accountId}
          onChange={(value) => filterChanged(setAccountId)(value)}
        />
        <DateRangeFilter
          idPrefix="cash-filter"
          dateFrom={range.dateFrom}
          dateTo={range.dateTo}
          onChange={filterChanged(setRange)}
        />
        <FormToggle
          isOpen={isCreating || editing !== null}
          openLabel="Catat transaksi"
          onToggle={() => {
            setNotice(null)
            if (editing !== null) {
              setEditing(null)
              return
            }
            setIsCreating((open) => !open)
          }}
        />
      </Toolbar>

      {isCreating ? (
        <CashForm
          transaction={null}
          accounts={accountItems}
          isLoadingAccounts={accounts.isPending}
        />
      ) : null}
      {editing ? (
        <CashForm
          key={editing.id}
          transaction={editing}
          accounts={accountItems}
          isLoadingAccounts={accounts.isPending}
          onUpdated={(message) => {
            setEditing(null)
            setNotice(message)
          }}
        />
      ) : null}
      {accounts.isError ? (
        <Notice>
          <ErrorNote message="Daftar akun gagal dimuat, jadi nama akun dan pilihan akun belum tersedia." />
        </Notice>
      ) : null}

      {deleteTransaction.isError ? (
        <Notice>
          <ErrorNote message={errorMessage(deleteTransaction.error)} />
        </Notice>
      ) : null}
      {notice ? (
        <Notice>
          <SuccessNote message={notice} />
        </Notice>
      ) : null}

      {!isRangeValid ? <ErrorNote message="Tanggal sampai tidak boleh lebih awal dari tanggal dari." /> : null}
      {isRangeValid && transactions.isPending ? <Loading /> : null}
      {isRangeValid && transactions.isError ? <LoadFailed onRetry={() => transactions.refetch()} /> : null}
      {isRangeValid && transactions.data ? (
        <>
          <Table
            rows={transactions.data.items}
            emptyMessage={
              hasFilter
                ? 'Tidak ada transaksi kas yang cocok dengan penyaringan ini.'
                : 'Belum ada transaksi kas. Catat yang pertama lewat tombol Catat transaksi.'
            }
            columns={[
              { header: 'Tanggal', cell: (row) => shortDate(row.date) },
              { header: 'Keterangan', cell: (row) => row.note || '-' },
              { header: 'Proyek', cell: (row) => projectName(row.projectId) },
              { header: 'Akun lawan', cell: (row) => accountLabel(accountItems, row.accountId) },
              {
                header: 'Jenis',
                cell: (row) => <Chip tone={CASH_TYPE_TONE[row.type]}>{CASH_TYPE_LABEL[row.type]}</Chip>,
              },
              { header: 'Jumlah', align: 'right', cell: (row) => <SignedAmount type={row.type} amount={row.amount} /> },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <RowActions>
                    <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                    <ConfirmDelete
                      subject={`transaksi ${transactionName(row)}`}
                      isConfirming={confirmingId === row.id}
                      isPending={deleteTransaction.isPending && deleteTransaction.variables === row.id}
                      onAsk={() => setConfirmingId(row.id)}
                      onCancel={() => setConfirmingId(null)}
                      onConfirm={() => remove(row)}
                    />
                  </RowActions>
                ),
              },
            ]}
          />
          <Pager
            page={transactions.data.page}
            totalPages={transactions.data.totalPages}
            totalItems={transactions.data.totalItems}
            unit="transaksi"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
