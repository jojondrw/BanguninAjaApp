import { describe, expect, it } from 'vitest'

import { journalBalance, toJournalRequest, type JournalLineValues } from '../models/accounting'
import { billingTotals, dueNote, paidPercent, paymentFormOf, toPaymentRequest, type BillingBalance } from '../models/billing'
import { resolveLabel, stepIndex } from '../views/components/searchSelectLogic'

function line(side: JournalLineValues['side'], amount: string, accountId = 'akun'): JournalLineValues {
  return { key: `${side}-${amount}`, accountId, side, amount }
}

describe('aturan saldo jurnal', () => {
  it('seimbang kalau total debit sama dengan total kredit', () => {
    const balance = journalBalance([line('debit', '1500000'), line('credit', '1000000'), line('credit', '500000')])
    expect(balance).toEqual({ debit: 1_500_000, credit: 1_500_000, difference: 0, isBalanced: true })
  })

  it('belum seimbang dan menunjukkan selisihnya', () => {
    const balance = journalBalance([line('debit', '2000000'), line('credit', '500000')])
    expect(balance.isBalanced).toBe(false)
    expect(balance.difference).toBe(1_500_000)
  })

  it('jurnal kosong atau bernilai nol tidak dianggap seimbang', () => {
    expect(journalBalance([line('debit', ''), line('credit', '')]).isBalanced).toBe(false)
    expect(journalBalance([line('debit', '0'), line('credit', '0')]).isBalanced).toBe(false)
  })

  it('isian bukan angka atau negatif tidak ikut dijumlah', () => {
    const balance = journalBalance([line('debit', 'abc'), line('debit', '-5'), line('credit', '10')])
    expect(balance.debit).toBe(0)
    expect(balance.credit).toBe(10)
  })

  it('setiap baris dikirim satu sisi saja', () => {
    const request = toJournalRequest({
      number: ' JU-1 ',
      date: '2026-10-03',
      note: '',
      source: '',
      lines: [line('debit', '750', 'kas'), line('credit', '750', 'modal')],
    })
    expect(request.number).toBe('JU-1')
    expect(request.date).toBe('2026-10-03T00:00:00Z')
    expect(request.lines).toEqual([
      { accountId: 'kas', debit: 750, credit: 0 },
      { accountId: 'modal', debit: 0, credit: 750 },
    ])
  })
})

function balance(overrides: Partial<BillingBalance>): BillingBalance {
  return {
    dueDate: '2026-10-10',
    amount: 1000,
    paidAmount: 0,
    outstanding: 1000,
    status: 'not_due',
    daysOverdue: 0,
    ...overrides,
  }
}

describe('tagihan', () => {
  it('formulir bayar terisi sisa tagihan dan nominal dikirim sebagai angka', () => {
    const values = paymentFormOf(balance({ paidAmount: 400, outstanding: 600 }), '2026-10-03')
    expect(values).toMatchObject({ amount: '600', paidAt: '2026-10-03', method: 'transfer' })

    const request = toPaymentRequest({ ...values, reference: ' TRX-9 ', note: ' termin 2 ' })
    expect(request).toEqual({
      amount: 600,
      paidAt: '2026-10-03T00:00:00Z',
      method: 'transfer',
      reference: 'TRX-9',
      note: 'termin 2',
    })
  })

  it('persentase terbayar dibulatkan ke bawah', () => {
    expect(paidPercent(balance({ amount: 3, paidAmount: 2 }))).toBe(66)
    expect(paidPercent(balance({ amount: 0 }))).toBe(0)
  })

  it('keterangan tenggat ditulis dalam kalimat', () => {
    expect(dueNote(balance({ status: 'overdue', daysOverdue: 12 }), '2026-10-03')).toBe('Terlambat 12 hari')
    expect(dueNote(balance({ status: 'due', dueDate: '2026-10-04' }), '2026-10-03')).toBe('Jatuh tempo besok')
    expect(dueNote(balance({ status: 'due', dueDate: '2026-10-08' }), '2026-10-03')).toBe('5 hari lagi')
    expect(dueNote(balance({ status: 'paid' }), '2026-10-03')).toBeNull()
  })

  it('ringkasan tagihan menghitung yang terbuka, terlambat, dan menandai daftar terpotong', () => {
    const totals = billingTotals(
      [
        balance({ status: 'overdue', outstanding: 300 }),
        balance({ status: 'due', outstanding: 200 }),
        balance({ status: 'paid', paidAmount: 1000, outstanding: 0 }),
      ],
      120,
    )
    expect(totals).toMatchObject({
      count: 3,
      openCount: 2,
      outstanding: 500,
      overdueCount: 1,
      overdueOutstanding: 300,
      dueSoonCount: 1,
      paidAmount: 1000,
      isPartial: true,
    })
  })
})

describe('logika SearchSelect', () => {
  it('stepIndex berputar di ujung daftar', () => {
    expect(stepIndex(-1, 3, 1)).toBe(0)
    expect(stepIndex(-1, 3, -1)).toBe(2)
    expect(stepIndex(2, 3, 1)).toBe(0)
    expect(stepIndex(0, 3, -1)).toBe(2)
    expect(stepIndex(0, 0, 1)).toBe(-1)
  })

  it('resolveLabel mendahulukan pilihan yang sudah diketahui', () => {
    const results = [{ value: 'b', label: 'Dari hasil' }]
    expect(resolveLabel('a', [null, { value: 'a', label: 'Awal' }], results)).toBe('Awal')
    expect(resolveLabel('b', [], results)).toBe('Dari hasil')
    expect(resolveLabel('c', [], results)).toBe('')
    expect(resolveLabel('', [{ value: '', label: 'x' }], results)).toBe('')
  })
})
