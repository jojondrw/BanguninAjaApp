import { describe, expect, it } from 'vitest'

import {
  ownerContractFormValues,
  ownerTermNote,
  toOwnerContractRequest,
  type OwnerContract,
  type OwnerTerm,
} from '../models/ownerBilling'

const contract: OwnerContract = {
  customerId: 'c1',
  customerName: 'Yayasan Kesehatan Ibu dan Anak Cirebon',
  number: 'KTR-RSIA-01',
  value: 67_000_000_000,
  retentionPercent: 5,
  paymentTermDays: 30,
}

function term(overrides: Partial<OwnerTerm>): OwnerTerm {
  return {
    id: 't1',
    label: 'Termin 1',
    kind: 'progress',
    state: 'planned',
    requiredProgress: 50,
    percent: 20,
    grossAmount: 0,
    retentionAmount: 0,
    netAmount: 0,
    paidAmount: 0,
    receivableId: null,
    claimedOn: null,
    claimedProgress: null,
    ...overrides,
  }
}

describe('kontrak pemberi kerja', () => {
  it('kontrak baru diisi nilai kontrak proyek, retensi 5%, dan tempo 30 hari', () => {
    expect(ownerContractFormValues(null, 67_000_000_000)).toEqual({
      customerId: '',
      number: '',
      value: '67000000000',
      retentionPercent: '5',
      paymentTermDays: '30',
    })
  })

  it('mengisi formulir dari kontrak yang sudah ada dan mengirim angkanya sebagai bilangan', () => {
    const values = { ...ownerContractFormValues(contract, 1), number: '  KTR-RSIA-01/REV ' }
    expect(toOwnerContractRequest(values)).toEqual({
      customerId: 'c1',
      number: 'KTR-RSIA-01/REV',
      value: 67_000_000_000,
      retentionPercent: 5,
      paymentTermDays: 30,
    })
  })
})

describe('keterangan termin pemberi kerja', () => {
  const date = (iso: string) => iso

  it('menyebut tanggal dan progres saat ditagih', () => {
    expect(ownerTermNote(term({ state: 'paid', claimedOn: '2026-08-02', claimedProgress: 25 }), 46, date)).toBe(
      'ditagih 2026-08-02 pada progres 25%',
    )
  })

  it('menyebut progres sekarang selama syaratnya belum tercapai', () => {
    expect(ownerTermNote(term({}), 46, date)).toBe('progres sekarang 46%')
    expect(ownerTermNote(term({ requiredProgress: 40 }), 46, date)).toBe('setelah termin sebelumnya')
  })
})
