import { describe, expect, it } from 'vitest'

import {
  ALIGNMENT_CENTERS,
  LEVEL_M_BLOCKS,
  encodeQr,
  formatBits,
  maskApplies,
  reedSolomonDivisor,
  reedSolomonRemainder,
  splitBlocks,
  versionBits,
  type QrMatrix,
} from '../views/pay/qrCode'

const PAY_URI = 'ethereum:0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed@11155111?value=50000000000000'

// Pembaca QR kecil khusus tes: menandai area pola tetap dengan aturan standar,
// membaca bit zig-zag, membuka mask, memisah blok, lalu membaca mode byte.
function reservedModules(version: number, size: number): boolean[][] {
  const reserved = Array.from({ length: size }, () => Array<boolean>(size).fill(false))
  const mark = (row: number, column: number) => {
    reserved[row][column] = true
  }
  for (let index = 0; index < size; index++) {
    mark(6, index)
    mark(index, 6)
  }
  for (let row = 0; row < 9; row++) {
    for (let column = 0; column < 9; column++) {
      mark(row, column)
    }
  }
  for (let row = 0; row < 9; row++) {
    for (let column = size - 8; column < size; column++) {
      mark(row, column)
    }
  }
  for (let row = size - 8; row < size; row++) {
    for (let column = 0; column < 9; column++) {
      mark(row, column)
    }
  }
  const centers = ALIGNMENT_CENTERS[version]
  const first = centers[0]
  const last = centers[centers.length - 1]
  const nearFinder = (row: number, column: number) =>
    (row === first && column === first) || (row === first && column === last) || (row === last && column === first)
  for (const row of centers) {
    for (const column of centers) {
      if (nearFinder(row, column)) {
        continue
      }
      for (let dy = -2; dy <= 2; dy++) {
        for (let dx = -2; dx <= 2; dx++) {
          mark(row + dy, column + dx)
        }
      }
    }
  }
  if (version >= 7) {
    for (let near = 0; near < 6; near++) {
      for (let far = size - 11; far < size - 8; far++) {
        mark(near, far)
        mark(far, near)
      }
    }
  }
  return reserved
}

function readCodewords(qr: QrMatrix): number[] {
  const reserved = reservedModules(qr.version, qr.size)
  const bits: number[] = []
  let column = qr.size - 1
  let upward = true
  while (column > 0) {
    if (column === 6) {
      column--
    }
    for (let step = 0; step < qr.size; step++) {
      const row = upward ? qr.size - 1 - step : step
      for (const current of [column, column - 1]) {
        if (!reserved[row][current]) {
          bits.push(Number(qr.modules[row][current] !== maskApplies(qr.mask, row, current)))
        }
      }
    }
    column -= 2
    upward = !upward
  }
  const bytes: number[] = []
  for (let index = 0; index + 8 <= bits.length; index += 8) {
    bytes.push(bits.slice(index, index + 8).reduce((value, current) => (value << 1) | current, 0))
  }
  return bytes
}

function deinterleave(codewords: number[], version: number): { data: number[][]; corrections: number[][] } {
  const layout = LEVEL_M_BLOCKS[version]
  const sizes = layout.groups.flatMap(([count, size]) => Array<number>(count).fill(size))
  const data = sizes.map(() => [] as number[])
  let position = 0
  for (let index = 0; index < Math.max(...sizes); index++) {
    sizes.forEach((size, block) => {
      if (index < size) {
        data[block].push(codewords[position++])
      }
    })
  }
  const corrections = sizes.map(() => [] as number[])
  for (let index = 0; index < layout.ecPerBlock; index++) {
    corrections.forEach((correction) => correction.push(codewords[position++]))
  }
  return { data, corrections }
}

function decodeByteMode(data: number[], version: number): string {
  const bits = data.flatMap((value) => Array.from({ length: 8 }, (_, index) => (value >>> (7 - index)) & 1))
  const read = (start: number, length: number) => bits.slice(start, start + length).reduce((value, bit) => (value << 1) | bit, 0)
  expect(read(0, 4)).toBe(0b0100)
  const countLength = version < 10 ? 8 : 16
  const length = read(4, countLength)
  const bytes = Array.from({ length }, (_, index) => read(4 + countLength + index * 8, 8))
  return new TextDecoder().decode(new Uint8Array(bytes))
}

function readFormatCopies(qr: QrMatrix): [number, number] {
  const at = (row: number, column: number) => Number(qr.modules[row][column])
  const first = [0, 1, 2, 3, 4, 5].map((row) => at(row, 8))
  first.push(at(7, 8), at(8, 8), at(8, 7))
  for (let index = 9; index < 15; index++) {
    first.push(at(8, 14 - index))
  }
  const second: number[] = []
  for (let index = 0; index < 8; index++) {
    second.push(at(8, qr.size - 1 - index))
  }
  for (let index = 8; index < 15; index++) {
    second.push(at(qr.size - 15 + index, 8))
  }
  const toNumber = (bitsLowFirst: number[]) => bitsLowFirst.reduce((value, bit, index) => value | (bit << index), 0)
  return [toNumber(first), toNumber(second)]
}

function decode(qr: QrMatrix): string {
  const codewords = readCodewords(qr)
  const { data, corrections } = deinterleave(codewords, qr.version)
  const divisor = reedSolomonDivisor(LEVEL_M_BLOCKS[qr.version].ecPerBlock)
  data.forEach((block, index) => {
    expect(reedSolomonRemainder(block, divisor)).toEqual(corrections[index])
  })
  return decodeByteMode(data.flat(), qr.version)
}

describe('penyandi QR', () => {
  it('menghitung kode Reed-Solomon sesuai contoh standar HELLO WORLD 1-M', () => {
    const data = [32, 91, 11, 120, 209, 114, 220, 77, 67, 64, 236, 17, 236, 17, 236, 17]
    expect(reedSolomonRemainder(data, reedSolomonDivisor(10))).toEqual([196, 35, 39, 119, 235, 215, 231, 226, 93, 23])
  })

  it('menulis bit format level M dan bit versi sesuai tabel standar', () => {
    const levelM = [
      '101010000010010', '101000100100101', '101111001111100', '101101101001011',
      '100010111111001', '100000011001110', '100111110010111', '100101010100000',
    ]
    levelM.forEach((expected, mask) => expect(formatBits(mask).toString(2).padStart(15, '0')).toBe(expected))
    expect(versionBits(7).toString(2).padStart(18, '0')).toBe('000111110010010100')
    expect(versionBits(10).toString(2).padStart(18, '0')).toBe('001010010011010011')
  })

  it('memilih versi terkecil yang muat untuk URI EIP-681', () => {
    const qr = encodeQr(PAY_URI)
    expect(qr.version).toBe(5)
    expect(qr.size).toBe(37)
    expect(qr.modules).toHaveLength(37)
  })

  it('menggambar pola pencari, pola waktu, modul gelap, dan dua salinan format yang sama', () => {
    const qr = encodeQr(PAY_URI)
    const finderRow = [true, true, true, true, true, true, true]
    expect(qr.modules[0].slice(0, 7)).toEqual(finderRow)
    expect(qr.modules[0].slice(qr.size - 7)).toEqual(finderRow)
    expect(qr.modules[qr.size - 1].slice(0, 7)).toEqual(finderRow)
    expect(qr.modules[3].slice(0, 7)).toEqual([true, false, true, true, true, false, true])
    for (let index = 8; index < qr.size - 8; index++) {
      expect(qr.modules[6][index]).toBe(index % 2 === 0)
      expect(qr.modules[index][6]).toBe(index % 2 === 0)
    }
    expect(qr.modules[qr.size - 8][8]).toBe(true)
    const [first, second] = readFormatCopies(qr)
    expect(first).toBe(formatBits(qr.mask))
    expect(second).toBe(formatBits(qr.mask))
  })

  it.each([
    ['URI pembayaran', PAY_URI],
    ['tautan halaman bayar', 'https://banguninaja.vercel.app/bayar/AbCdEfGhIjKlMnOpQrStUvWxYz012345'],
    ['teks pendek', 'HELLO'],
    ['teks dengan huruf non-ASCII', 'Bayar tagihan — proyek Griya Asri'],
    ['versi 8 dengan dua kelompok blok', 'x'.repeat(150)],
    ['versi 10 penuh', 'y'.repeat(213)],
  ])('membaca kembali %s dengan isi yang sama', (_, text) => {
    expect(decode(encodeQr(text))).toBe(text)
  })

  it('tetap terbaca dengan setiap mask', () => {
    for (let mask = 0; mask < 8; mask++) {
      const qr = encodeQr(PAY_URI, mask)
      expect(qr.mask).toBe(mask)
      expect(decode(qr)).toBe(PAY_URI)
    }
  })

  it('memotong data per blok sesuai tabel', () => {
    const data = Array.from({ length: 154 }, (_, index) => index)
    expect(splitBlocks(data, 8).map((block) => block.length)).toEqual([38, 38, 39, 39])
  })

  it('menolak teks yang melebihi kapasitas versi 10', () => {
    expect(() => encodeQr('z'.repeat(214))).toThrow(/terlalu panjang/)
  })
})
