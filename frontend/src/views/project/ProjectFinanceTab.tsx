import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useBudgets } from '../../controllers/useErp'
import {
  CASH_FLOW_MONTHS,
  useAccounts,
  useCreateCashTransaction,
  useProjectCashFlow,
  useProjectCashTransactions,
} from '../../controllers/useProjectWorkspace'
import {
  CASH_NOTE_MAX_LENGTH,
  CASH_TYPE_LABEL,
  CASH_TYPE_TONE,
  emptyCashTransactionForm,
  summarizeBudgets,
  type Budget,
  type CashFlowPeriod,
  type CashTransactionFormValues,
} from '../../models/finance'
import { accountGroups, type Account } from '../../models/master'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { kpiValue, monthLabel, rupiah, rupiahShort, shortDate } from '../../shared/format'
import { Bar, Card, Empty, Kpi, KpiRow, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { Chip, FormSection } from './parts'

const NO_BUDGETS: Budget[] = []
const NO_ACCOUNTS: Account[] = []
const MIN_VISIBLE_BAR = 2

export function ProjectFinanceTab({ project }: { project: Project }) {
  const budgets = useBudgets(project.id)
  const cashFlow = useProjectCashFlow(project.id)
  const summary = summarizeBudgets(budgets.data?.items ?? NO_BUDGETS)

  return (
    <div className="space-y-6">
      <KpiRow>
        <Kpi
          label="Anggaran"
          value={kpiValue(budgets.isPending, budgets.isError, rupiahShort(summary.value))}
          note={summary.years > 0 ? `${summary.years} tahun anggaran` : 'belum ada anggaran'}
        />
        <Kpi
          label="Realisasi"
          value={kpiValue(budgets.isPending, budgets.isError, rupiahShort(summary.realized))}
          note={summary.value > 0 ? `serapan ${summary.absorption}% dari anggaran` : 'kas keluar di tahun anggaran'}
        />
        <Kpi
          label="Sisa anggaran"
          value={kpiValue(budgets.isPending, budgets.isError, rupiahShort(summary.remaining))}
          note={summary.remaining < 0 ? 'realisasi melebihi anggaran' : 'belum terserap'}
        />
        <Kpi
          label="Saldo kas proyek"
          value={kpiValue(cashFlow.isPending, cashFlow.isError, rupiahShort(cashFlow.data?.balance ?? 0))}
          note={
            cashFlow.data
              ? `masuk ${rupiahShort(cashFlow.data.totalIn)}, keluar ${rupiahShort(cashFlow.data.totalOut)} dalam ${CASH_FLOW_MONTHS} bulan`
              : undefined
          }
        />
      </KpiRow>

      <div className="grid gap-6 xl:grid-cols-[1.4fr_1fr]">
        <Card
          title="Pengeluaran per bulan"
          description={`Kas keluar proyek ${CASH_FLOW_MONTHS} bulan terakhir, dikelompokkan oleh server`}
        >
          {cashFlow.isPending ? <Loading /> : null}
          {cashFlow.isError ? <LoadFailed onRetry={() => cashFlow.refetch()} /> : null}
          {cashFlow.data ? <SpendByMonth periods={cashFlow.data.periods} /> : null}
        </Card>

        <Card title="Anggaran dan realisasi" description="Realisasi adalah kas keluar proyek di tahun anggaran yang sama">
          {budgets.isPending ? <Loading /> : null}
          {budgets.isError ? <LoadFailed onRetry={() => budgets.refetch()} /> : null}
          {budgets.data ? (
            <Table
              rows={budgets.data.items}
              emptyMessage="Belum ada anggaran untuk proyek ini."
              columns={[
                { header: 'Tahun', cell: (row) => String(row.year) },
                { header: 'Anggaran', align: 'right', cell: (row) => rupiahShort(row.value) },
                { header: 'Realisasi', align: 'right', cell: (row) => rupiahShort(row.realized) },
                { header: 'Serapan', align: 'right', cell: (row) => <Bar percent={row.absorption} /> },
              ]}
            />
          ) : null}
        </Card>
      </div>

      <TransactionsCard project={project} />
    </div>
  )
}

function SpendByMonth({ periods }: { periods: CashFlowPeriod[] }) {
  const highest = Math.max(0, ...periods.map((period) => period.cashOut))

  if (highest === 0) {
    return <Empty message={`Belum ada kas keluar dalam ${CASH_FLOW_MONTHS} bulan terakhir.`} />
  }

  return (
    <ul className="space-y-2">
      {periods.map((period) => {
        const width = period.cashOut > 0 ? Math.max(MIN_VISIBLE_BAR, (period.cashOut / highest) * 100) : 0
        return (
          <li key={period.period} className="grid grid-cols-[5rem_1fr_5.5rem] items-center gap-3">
            <span className="text-xs text-slate-600">{monthLabel(period.period)}</span>
            <span aria-hidden="true" className="h-2.5 overflow-hidden rounded-full bg-slate-100">
              <span className="block h-full rounded-full bg-navy-600" style={{ width: `${width}%` }} />
            </span>
            <span className="text-right text-xs tabular-nums text-slate-700">
              {period.cashOut > 0 ? rupiahShort(period.cashOut) : '-'}
            </span>
          </li>
        )
      })}
    </ul>
  )
}

function TransactionsCard({ project }: { project: Project }) {
  const transactions = useProjectCashTransactions(project.id)
  const accounts = useAccounts()
  const [isAdding, setIsAdding] = useState(false)

  const accountName = (id: string) => {
    const account = accounts.data?.items.find((item) => item.id === id)
    return account ? `${account.code} ${account.name}` : '-'
  }
  const page = transactions.data

  return (
    <Card
      title="Transaksi kas proyek"
      description={page ? transactionsDescription(page.items.length, page.totalItems) : undefined}
    >
      {transactions.isPending ? <Loading /> : null}
      {transactions.isError ? <LoadFailed onRetry={() => transactions.refetch()} /> : null}
      {page ? (
        <Table
          rows={page.items}
          emptyMessage="Belum ada transaksi kas untuk proyek ini."
          columns={[
            { header: 'Tanggal', cell: (row) => shortDate(row.date) },
            { header: 'Keterangan', cell: (row) => row.note || '-' },
            { header: 'Akun', cell: (row) => accountName(row.accountId) },
            { header: 'Jenis', cell: (row) => <Chip tone={CASH_TYPE_TONE[row.type]}>{CASH_TYPE_LABEL[row.type]}</Chip> },
            { header: 'Jumlah', align: 'right', cell: (row) => rupiah(row.amount) },
          ]}
        />
      ) : null}

      <div className="mt-5">
        <Button variant="subtle" onClick={() => setIsAdding((open) => !open)}>
          {isAdding ? 'Tutup formulir transaksi' : 'Catat transaksi'}
        </Button>
      </div>

      {isAdding ? <CashTransactionForm projectId={project.id} /> : null}
    </Card>
  )
}

function transactionsDescription(shown: number, total: number): string {
  if (shown < total) {
    return `Menampilkan ${shown} transaksi terbaru dari ${total}`
  }
  return `${total} transaksi, terbaru di atas`
}

function amountHint(value: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount) || amount <= 0) {
    return 'Dalam Rupiah, lebih dari nol'
  }
  return rupiah(amount)
}

function CashTransactionForm({ projectId }: { projectId: string }) {
  const [values, setValues] = useState<CashTransactionFormValues>(() => emptyCashTransactionForm(new Date()))
  const accounts = useAccounts()
  const createTransaction = useCreateCashTransaction(projectId)
  const groups = accountGroups(accounts.data?.items ?? NO_ACCOUNTS, values.type)

  const update =
    (key: keyof CashTransactionFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createTransaction.mutate(values, {
      onSuccess: () => setValues((current) => ({ ...current, amount: '', note: '' })),
    })
  }

  return (
    <FormSection title="Transaksi kas baru">
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="cash-date"
            label="Tanggal"
            type="date"
            required
            value={values.date}
            onChange={update('date')}
          />
          <SelectField id="cash-type" label="Jenis" value={values.type} onChange={update('type')}>
            <option value="out">Kas keluar</option>
            <option value="in">Kas masuk</option>
          </SelectField>
          <SelectField
            id="cash-account"
            label="Akun"
            required
            disabled={accounts.isPending}
            hint="Kas keluar biasanya ke akun Beban, kas masuk ke akun Pendapatan"
            value={values.accountId}
            onChange={update('accountId')}
          >
            <option value="">{accounts.isPending ? 'Memuat akun...' : 'Pilih akun'}</option>
            {groups.map((group) => (
              <optgroup key={group.type} label={group.label}>
                {group.accounts.map((account) => (
                  <option key={account.id} value={account.id}>
                    {account.code} {account.name}
                  </option>
                ))}
              </optgroup>
            ))}
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
            hint={amountHint(values.amount)}
            value={values.amount}
            onChange={update('amount')}
          />
          <div className="md:col-span-2">
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
        </div>

        {accounts.isError ? <LoadFailed onRetry={() => accounts.refetch()} /> : null}

        <Button type="submit" isPending={createTransaction.isPending} pendingLabel="Menyimpan transaksi">
          Simpan transaksi
        </Button>

        {createTransaction.isError ? <ErrorNote message={errorMessage(createTransaction.error)} /> : null}
        {createTransaction.isSuccess ? (
          <SuccessNote
            message={`Kas ${CASH_TYPE_LABEL[createTransaction.data.type].toLowerCase()} ${rupiah(createTransaction.data.amount)} tercatat untuk proyek ini.`}
          />
        ) : null}
      </form>
    </FormSection>
  )
}
