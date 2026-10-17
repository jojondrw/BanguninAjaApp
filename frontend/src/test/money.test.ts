import { describe, expect, it } from 'vitest'

import { rupiah, rupiahShort } from '../shared/format'
import { splitMoney } from '../shared/money'

function moneyIn(text: string): string[] {
  return splitMoney(text)
    .filter((part) => part.money)
    .map((part) => part.text)
}

describe('nilai uang di dalam teks', () => {
  it('mengenali keluaran rupiah dan rupiahShort, termasuk yang negatif', () => {
    for (const amount of [0, 1_250_000, -500, 48_000_000, 2_500_000_000, -2_500_000_000]) {
      expect(moneyIn(rupiah(amount))).toEqual([rupiah(amount)])
      expect(moneyIn(rupiahShort(amount))).toEqual([rupiahShort(amount)])
    }
  })

  it('memisahkan beberapa nilai uang dari kalimat di sekitarnya', () => {
    const text = `masuk ${rupiahShort(10_200_000_000)}, keluar ${rupiahShort(18_900_000_000)}`
    expect(splitMoney(text).map((part) => [part.text, part.money])).toEqual([
      ['masuk ', false],
      [rupiahShort(10_200_000_000), true],
      [', keluar ', false],
      [rupiahShort(18_900_000_000), true],
    ])
  })

  it('tidak mengambil titik penutup kalimat atau kata yang diawali M', () => {
    expect(moneyIn(`Total ${rupiah(1_000)}.`)).toEqual([rupiah(1_000)])
    expect(moneyIn('Rp 5 Masuk')).toEqual(['Rp 5'])
  })

  it('ikut mengambil tanda plus pada perubahan nilai', () => {
    expect(moneyIn('adendum +Rp 1,3 M')).toEqual(['+Rp 1,3 M'])
  })

  it('membiarkan tanggal, persen, dan angka biasa', () => {
    expect(moneyIn('1 Jun 2027, progres 46%, 12 unit')).toEqual([])
  })
})
