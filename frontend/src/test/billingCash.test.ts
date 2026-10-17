import { describe, expect, it } from 'vitest'

import { isFromBilling } from '../models/finance'

describe('kas dari pembayaran tagihan', () => {
  it('dikenali dari tautan ke pembayaran tagihan', () => {
    expect(isFromBilling({ billingPaymentId: 'b5c2f0f4-0d1e-4f39-9f3a-0c2f1f6c8e11' })).toBe(true)
  })

  it('kas yang dicatat manual tetap bisa diubah dari Keuangan', () => {
    expect(isFromBilling({ billingPaymentId: null })).toBe(false)
  })
})
