import { ChevronRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useBudgets, useCashByAccount, useCashFlow, useCashTransactions } from '../../controllers/useErp'
import { busiestBudgets, topAccounts, type Budget, type CashAccountTotal } from '../../models/finance'
import { number, rupiahShort, shortDate } from '../../shared/format'
import { monthsAgo, todayDate } from '../../shared/localDate'
import { CashFlowChart } from '../components/CashFlowChart'
import { DonutChart } from '../components/DonutChart'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Money } from '../components/Money'
import type { FinanceTab } from './financeTabs'
import { SignedAmount } from './parts'

const CHART_MONTHS = 12
const TOP_ACCOUNTS = 5
const TOP_BUDGETS = 6
const RECENT = 5

function LinkButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex items-center gap-0.5 text-[13px] font-medium text-navy-600 hover:text-navy-800"
    >
      {label}
      <ChevronRight aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
    </button>
  )
}

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

function MoneyMix({ from, to }: { from: string; to: string }) {
  const accounts = useCashByAccount({ dateFrom: from, dateTo: to })
  if (accounts.isPending) {
    return <Loading />
  }
  if (accounts.isError) {
    return <LoadFailed onRetry={() => accounts.refetch()} />
  }
  return (
    <div className="grid gap-6 sm:grid-cols-2">
      <div>
        <p className="mb-3 text-xs font-medium text-slate-500">Keluar untuk</p>
        <AccountDonut accounts={accounts.data} type="out" colors={OUT_COLORS} caption="Kas keluar per akun" />
      </div>
      <div>
        <p className="mb-3 text-xs font-medium text-slate-500">Masuk dari</p>
        <AccountDonut accounts={accounts.data} type="in" colors={IN_COLORS} caption="Kas masuk per akun" />
      </div>
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

// Ringkasan keuangan dalam satu layar: arus kas, ke mana uang pergi, serapan
// anggaran, dan transaksi terakhir. Rincian lengkap ada di tab lainnya.
export function FinanceSummaryTab({ onOpenTab }: { onOpenTab: (tab: FinanceTab) => void }) {
  const from = monthsAgo(CHART_MONTHS - 1)
  const to = todayDate()
  const cashFlow = useCashFlow({ dateFrom: from, dateTo: to })
  const budgets = useBudgets()
  const transactions = useCashTransactions(RECENT)

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
      <Card title="Arus kas 12 bulan" description="Masuk ke atas, keluar ke bawah, garis kuning adalah selisihnya">
        {cashFlow.isPending ? <Loading /> : null}
        {cashFlow.isError ? <LoadFailed onRetry={() => cashFlow.refetch()} /> : null}
        {cashFlow.data ? <CashFlowChart periods={cashFlow.data.periods} /> : null}
      </Card>

      <Card
        title="Ke mana uangnya"
        description="Kas per akun dalam 12 bulan"
        action={<LinkButton label="Buku besar" onClick={() => onOpenTab('buku-besar')} />}
      >
        <MoneyMix from={from} to={to} />
      </Card>

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
  )
}
