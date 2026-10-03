import { useSearchParams } from 'react-router-dom'

import { rupiah } from '../../shared/format'
import { balanceSide } from '../../models/accounting'

export type FinanceTab = 'ringkasan' | 'anggaran' | 'kas' | 'jurnal' | 'buku-besar'

export const FINANCE_TABS: { value: FinanceTab; label: string }[] = [
  { value: 'ringkasan', label: 'Ringkasan' },
  { value: 'anggaran', label: 'Anggaran' },
  { value: 'kas', label: 'Kas' },
  { value: 'jurnal', label: 'Jurnal' },
  { value: 'buku-besar', label: 'Buku besar' },
]

export const DEFAULT_FINANCE_TAB: FinanceTab = 'ringkasan'
export const TAB_PARAM = 'tab'
export const JOURNAL_PARAM = 'jurnal'
export const ACCOUNT_PARAM = 'akun'
export const CASH_PARAM = 'transaksi'

export function toFinanceTab(value: string | null): FinanceTab {
  return FINANCE_TABS.find((tab) => tab.value === value)?.value ?? DEFAULT_FINANCE_TAB
}

// Tautan antar tab di halaman yang sama, misalnya dari baris jurnal ke buku
// besar akunnya. Hanya query yang berubah, jadi path halaman tetap.
export function ledgerLink(accountId: string) {
  return { search: `?${new URLSearchParams({ [TAB_PARAM]: 'buku-besar', [ACCOUNT_PARAM]: accountId })}` }
}

export function journalLink(journalEntryId: string) {
  return { search: `?${new URLSearchParams({ [TAB_PARAM]: 'jurnal', [JOURNAL_PARAM]: journalEntryId })}` }
}

// Dari jurnal otomatis kembali ke transaksi kas asalnya.
export function cashLink(cashTransactionId: string) {
  return { search: `?${new URLSearchParams({ [TAB_PARAM]: 'kas', [CASH_PARAM]: cashTransactionId })}` }
}

// Satu parameter query sebagai state. replace dipakai supaya tombol kembali
// tidak menelusuri setiap pilihan.
export function useSearchParam(name: string): [string, (value: string) => void] {
  const [params, setParams] = useSearchParams()

  const update = (value: string) =>
    setParams(
      (current) => {
        const next = new URLSearchParams(current)
        if (value === '') {
          next.delete(name)
        } else {
          next.set(name, value)
        }
        return next
      },
      { replace: true },
    )

  return [params.get(name) ?? '', update]
}

export function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount) || amount <= 0) {
    return fallback
  }
  return rupiah(amount)
}

// Saldo buku besar ditulis tanpa tanda minus, diikuti sisinya: D untuk debit,
// K untuk kredit.
export function balanceText(balance: number): string {
  const side = balanceSide(balance)
  return side === '' ? rupiah(0) : `${rupiah(Math.abs(balance))} ${side}`
}
