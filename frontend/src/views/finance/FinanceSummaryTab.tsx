import { Link } from 'react-router-dom'

import { useBudgets, useCashByAccount, useCashFlow, useCashTransactions } from '../../controllers/useErp'
import { busiestBudgets, topAccounts, type Budget, type CashAccountTotal, type CashFlow } from '../../models/finance'
import { monthEndBalances } from '../../models/overview'
import { number, rupiahShort, shortDate } from '../../shared/format'
import { monthsAgo, todayDate } from '../../shared/localDate'
import { CashFlowChart } from '../components/CashFlowChart'
import { DonutChart } from '../components/DonutChart'
import { LinkButton } from '../components/LinkButton'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Money, MoneyText } from '../components/Money'
import { Sparkline } from '../components/Sparkline'
import type { FinanceTab } from './financeTabs'
import { SignedAmount } from './parts'

const CHART_MONTHS = 12
const TOP_ACCOUNTS = 5
const TOP_BUDGETS = 6
const RECENT = 5

const OUT_COLORS = ['#0b2b6b', '#1f5fd1', '#5b86d9', '#94a3b8', '#cbd5e1', '#e2e8f0']
const IN_COLORS = ['#15803d', '#22c55e', '#86efac', '#fbbb16', '#cbd5e1', '#e2e8f0']

function AccountDonut({ accounts, type, colors, caption }: {
  accounts: CashAccountTotal[]
  type: 'in' | 'out'
  colors: string[]
  caption: string
}) {
  const { items } = topAccounts(accounts, type, TOP_ACCOUNTS)
  return <DonutChart slices={items.map((item) => ({ label: item.label, value: item.total }))} colors={colors} caption={caption} />
}

function Total({ label, value, note, trend, tone = 'navy' }: {
  label: string
  value: string
  note: string
  trend: number[]
  tone?: 'navy' | 'red'
}) {
  return (
    <div className="flex min-w-0 flex-col bg-white px-5 pt-4 pb-3 sm:px-6 sm:pt-5">
      <p className="text-[13px] font-medium text-slate-500">{label}</p>
      <p className="mt-1.5 truncate text-[22px] leading-7 font-semibold tracking-[-0.03em] text-slate-900 tabular-nums">
        <MoneyText>{value}</MoneyText>
      </p>
      <p className="mt-0.5 text-xs text-slate-500">{note}</p>
      <div className="mt-auto pt-4">
        <Sparkline values={trend} tone={tone} />
      </div>
    </div>
  )
}

function CashTotals({ cashFlow }: { cashFlow: CashFlow | undefined }) {
  const value = (amount: number | undefined) => (cashFlow ? rupiahShort(amount ?? 0) : '...')
  const periods = cashFlow?.periods ?? []
  return (
    <div className="grid h-full grid-cols-2 gap-px overflow-hidden rounded-2xl bg-slate-200/70 shadow-hairline">
      <Total
        label="Kas masuk"
        value={value(cashFlow?.totalIn)}
        note="12 bulan terakhir"
        trend={periods.map((period) => period.cashIn)}
      />
      <Total
        label="Kas keluar"
        value={value(cashFlow?.totalOut)}
        note="12 bulan terakhir"
        trend={periods.map((period) => period.cashOut)}
      />
      <Total
        label="Selisih"
        value={value(cashFlow?.net)}
        note="masuk dikurangi keluar"
        trend={periods.map((period) => period.net)}
        tone={(cashFlow?.net ?? 0) < 0 ? 'red' : 'navy'}
      />
      <Total
        label="Saldo"
        value={value(cashFlow?.balance)}
        note="per hari ini"
        trend={cashFlow ? monthEndBalances(periods, cashFlow.balance) : []}
      />
    </div>
  )
}

function BudgetBullets({ budgets }: { budgets: Budget[] }) {
  if (budgets.length === 0) {
    return <Empty message="Belum ada anggaran. Tambahkan di tab Anggaran." />
  }
  return (
    <ul className="space-y-3">
      {busiestBudgets(budgets, TOP_BUDGETS).map((budget) => {
        const over = budget.absorption > 100
        return (
          <li key={budget.id}>
            <div className="flex items-baseline justify-between gap-3 text-[13px]">
              <Link
                to={`/proyek/${budget.projectId}?tab=keuangan`}
                className="truncate font-medium text-slate-900 hover:text-navy-600"
              >
                {budget.projectName || '-'} <span className="font-normal text-slate-400">{budget.year}</span>
              </Link>
              <span className={`shrink-0 tabular-nums ${over ? 'font-medium text-red-600' : 'text-slate-600'}`}>
                {number(budget.absorption)}%
              </span>
            </div>
            <span aria-hidden="true" className="relative mt-1 block h-2 overflow-hidden rounded-full bg-slate-100">
              <span
                className={`block h-full rounded-full ${over ? 'bg-red-500' : 'bar-fill'}`}
                style={{ width: `${Math.min(100, budget.absorption)}%` }}
              />
            </span>
            <p className="mt-0.5 text-xs text-slate-500">
              <Money>{rupiahShort(budget.realized)}</Money> dari <Money>{rupiahShort(budget.value)}</Money>
            </p>
          </li>
        )
      })}
    </ul>
  )
}

// Ringkasan keuangan: angka kas 12 bulan bersama asal dan tujuannya, grafik
// arus kas selebar halaman, lalu serapan anggaran dan transaksi terakhir.
export function FinanceSummaryTab({ onOpenTab }: { onOpenTab: (tab: FinanceTab) => void }) {
  const from = monthsAgo(CHART_MONTHS - 1)
  const to = todayDate()
  const cashFlow = useCashFlow({ dateFrom: from, dateTo: to })
  const accounts = useCashByAccount({ dateFrom: from, dateTo: to })
  const budgets = useBudgets()
  const transactions = useCashTransactions(RECENT)

  function accountDonut(type: 'in' | 'out') {
    if (accounts.isPending) {
      return <Loading />
    }
    if (accounts.isError) {
      return <LoadFailed onRetry={() => accounts.refetch()} />
    }
    return type === 'in' ? (
      <AccountDonut accounts={accounts.data} type="in" colors={IN_COLORS} caption="Kas masuk per akun" />
    ) : (
      <AccountDonut accounts={accounts.data} type="out" colors={OUT_COLORS} caption="Kas keluar per akun" />
    )
  }

  return (
    <div className="grid gap-6">
      <div className="grid gap-6 lg:grid-cols-2 xl:grid-cols-[minmax(0,4fr)_minmax(0,3fr)_minmax(0,3fr)]">
        <div className="lg:col-span-2 xl:col-span-1">
          <CashTotals cashFlow={cashFlow.data} />
        </div>
        <Card title="Masuk dari" description="Kas masuk per akun, 12 bulan">
          {accountDonut('in')}
        </Card>
        <Card title="Keluar untuk" description="Kas keluar per akun, 12 bulan">
          {accountDonut('out')}
        </Card>
      </div>

      <Card title="Arus kas 12 bulan" description="Masuk ke atas, keluar ke bawah, garis kuning adalah selisihnya">
        {cashFlow.isPending ? <Loading /> : null}
        {cashFlow.isError ? <LoadFailed onRetry={() => cashFlow.refetch()} /> : null}
        {cashFlow.data ? <CashFlowChart periods={cashFlow.data.periods} /> : null}
      </Card>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card
          title="Serapan anggaran"
          description="Proyek dengan serapan tertinggi, merah bila melewati anggaran"
          action={<LinkButton label="Semua anggaran" onClick={() => onOpenTab('anggaran')} />}
        >
          {budgets.isPending ? <Loading /> : null}
          {budgets.isError ? <LoadFailed onRetry={() => budgets.refetch()} /> : null}
          {budgets.data ? <BudgetBullets budgets={budgets.data.items} /> : null}
        </Card>

        <Card title="Transaksi terakhir" action={<LinkButton label="Semua transaksi" onClick={() => onOpenTab('kas')} />}>
          {transactions.isPending ? <Loading /> : null}
          {transactions.isError ? <LoadFailed onRetry={() => transactions.refetch()} /> : null}
          {transactions.data ? (
            transactions.data.items.length === 0 ? (
              <Empty message="Belum ada transaksi kas." />
            ) : (
              <ul className="divide-y divide-slate-100">
                {transactions.data.items.map((row) => (
                  <li key={row.id} className="flex items-center justify-between gap-3 py-2.5 text-[13px]">
                    <span className="min-w-0">
                      <span className="block truncate text-slate-900">{row.note || '-'}</span>
                      <span className="text-xs text-slate-500">{shortDate(row.date)}</span>
                    </span>
                    <span className="shrink-0">
                      <SignedAmount type={row.type} amount={row.amount} />
                    </span>
                  </li>
                ))}
              </ul>
            )
          ) : null}
        </Card>
      </div>
    </div>
  )
}
