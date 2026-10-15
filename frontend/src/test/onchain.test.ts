import { describe, expect, it } from 'vitest'

import {
  etherLabel,
  formatEther,
  normalizeTxHash,
  payLink,
  rateLabel,
  shortHex,
  toHexQuantity,
  weiForRupiah,
} from '../models/onchain'

const HASH = '0xc9be6b0adf4a05bb83cd9768fabf03e46e489ee6193c7b841668a9236ca62b0f'

describe('model pembayaran on-chain', () => {
  it('mengonversi rupiah ke wei dengan kurs simulasi yang sama seperti backend', () => {
    expect(weiForRupiah(50_000_000, 1e12)).toBe(50_000_000_000_000n)
    expect(weiForRupiah(1, 3)).toBe(333_333_333_333_333_334n)
    expect(formatEther(weiForRupiah(50_000_000, 1e12))).toBe('0.00005')
    expect(formatEther(10n ** 18n)).toBe('1')
    expect(formatEther(15n * 10n ** 17n)).toBe('1.5')
  })

  it('menulis ETH dengan koma desimal dan label kurs simulasi', () => {
    expect(etherLabel('0.00005', 11155111)).toBe('0,00005 SepoliaETH')
    expect(etherLabel('2', 1)).toBe('2 ETH')
    expect(rateLabel(1e12, 11155111)).toMatch(/^Kurs simulasi: 1 SepoliaETH = Rp\s?1\.000\.000\.000\.000$/)
  })

  it('menerima hash transaksi yang ditempel dengan spasi atau huruf besar', () => {
    expect(normalizeTxHash(`  ${HASH.toUpperCase().replace('0X', '0x')}\n`)).toBe(HASH)
    expect(normalizeTxHash(HASH.slice(2))).toBeNull()
    expect(normalizeTxHash(HASH.slice(0, 65))).toBeNull()
    expect(normalizeTxHash(`0x${'g'.repeat(64)}`)).toBeNull()
  })

  it('menyusun tautan bayar, nilai heksadesimal untuk dompet, dan hash ringkas', () => {
    expect(payLink('https://banguninaja.vercel.app', 'abc')).toBe('https://banguninaja.vercel.app/bayar/abc')
    expect(toHexQuantity('50000000000000')).toBe('0x2d79883d2000')
    expect(shortHex(HASH)).toBe('0xc9be6b…a62b0f')
  })
})
