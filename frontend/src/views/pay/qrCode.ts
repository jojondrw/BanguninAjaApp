// Penyandi QR Code kecil tanpa pustaka luar: mode byte, koreksi galat M,
// versi 1 sampai 10 (cukup untuk URI EIP-681 sekitar 90 karakter). Mengikuti
// ISO/IEC 18004: data dipecah per blok, diberi kode Reed-Solomon, disusun
// zig-zag, lalu dipilih mask dengan penalti terkecil.

export interface QrMatrix {
  version: number
  size: number
  mask: number
  // modules[baris][kolom], true berarti modul gelap.
  modules: boolean[][]
}

interface BlockLayout {
  ecPerBlock: number
  // [jumlah blok, codeword data per blok]
  groups: [number, number][]
}

export const MAX_VERSION = 10

// Tabel koreksi galat level M untuk versi 1 sampai 10 (indeks 0 tidak dipakai).
export const LEVEL_M_BLOCKS: BlockLayout[] = [
  { ecPerBlock: 0, groups: [] },
  { ecPerBlock: 10, groups: [[1, 16]] },
  { ecPerBlock: 16, groups: [[1, 28]] },
  { ecPerBlock: 26, groups: [[1, 44]] },
  { ecPerBlock: 18, groups: [[2, 32]] },
  { ecPerBlock: 24, groups: [[2, 43]] },
  { ecPerBlock: 16, groups: [[4, 27]] },
  { ecPerBlock: 18, groups: [[4, 31]] },
  { ecPerBlock: 22, groups: [[2, 38], [2, 39]] },
  { ecPerBlock: 22, groups: [[3, 36], [2, 37]] },
  { ecPerBlock: 26, groups: [[4, 43], [1, 44]] },
]

export const ALIGNMENT_CENTERS: number[][] = [
  [],
  [],
  [6, 18],
  [6, 22],
  [6, 26],
  [6, 30],
  [6, 34],
  [6, 22, 38],
  [6, 24, 42],
  [6, 26, 46],
  [6, 28, 50],
]

const BYTE_MODE = 0b0100
const LEVEL_M_FORMAT = 0b00
const FORMAT_MASK = 0x5412
const FORMAT_GENERATOR = 0x537
const VERSION_GENERATOR = 0x1f25
const GF_POLYNOMIAL = 0x11d
const PAD_BYTES = [0xec, 0x11]
const FINDER_PATTERN = [true, false, true, true, true, false, true]

function bit(value: number, index: number): boolean {
  return ((value >>> index) & 1) !== 0
}

export function dataCapacity(version: number): number {
  return LEVEL_M_BLOCKS[version].groups.reduce((sum, [count, size]) => sum + count * size, 0)
}

function countBits(version: number): number {
  return version < 10 ? 8 : 16
}

function chooseVersion(length: number): number {
  for (let version = 1; version <= MAX_VERSION; version++) {
    if (4 + countBits(version) + length * 8 <= dataCapacity(version) * 8) {
      return version
    }
  }
  throw new Error(`Teks terlalu panjang untuk QR versi ${MAX_VERSION} (${length} byte)`)
}

function encodeData(bytes: Uint8Array, version: number): number[] {
  const bits: number[] = []
  const push = (value: number, length: number) => {
    for (let index = length - 1; index >= 0; index--) {
      bits.push(bit(value, index) ? 1 : 0)
    }
  }
  const capacityBits = dataCapacity(version) * 8
  push(BYTE_MODE, 4)
  push(bytes.length, countBits(version))
  bytes.forEach((value) => push(value, 8))
  push(0, Math.min(4, capacityBits - bits.length))
  push(0, (8 - (bits.length % 8)) % 8)

  const codewords: number[] = []
  for (let index = 0; index < bits.length; index += 8) {
    codewords.push(bits.slice(index, index + 8).reduce((value, current) => (value << 1) | current, 0))
  }
  for (let pad = 0; codewords.length < dataCapacity(version); pad++) {
    codewords.push(PAD_BYTES[pad % 2])
  }
  return codewords
}

export function gfMultiply(left: number, right: number): number {
  let product = 0
  for (let index = 7; index >= 0; index--) {
    product = (product << 1) ^ ((product >>> 7) * GF_POLYNOMIAL)
    product ^= ((right >>> index) & 1) * left
  }
  return product
}

export function reedSolomonDivisor(degree: number): number[] {
  const divisor = Array.from({ length: degree }, (_, index): number => (index === degree - 1 ? 1 : 0))
  let root = 1
  for (let step = 0; step < degree; step++) {
    for (let index = 0; index < degree; index++) {
      divisor[index] = gfMultiply(divisor[index], root)
      if (index + 1 < degree) {
        divisor[index] ^= divisor[index + 1]
      }
    }
    root = gfMultiply(root, 0x02)
  }
  return divisor
}

export function reedSolomonRemainder(data: number[], divisor: number[]): number[] {
  const remainder = divisor.map(() => 0)
  for (const value of data) {
    const factor = value ^ (remainder.shift() ?? 0)
    remainder.push(0)
    divisor.forEach((coefficient, index) => {
      remainder[index] ^= gfMultiply(coefficient, factor)
    })
  }
  return remainder
}

export function splitBlocks(data: number[], version: number): number[][] {
  const blocks: number[][] = []
  let offset = 0
  for (const [count, size] of LEVEL_M_BLOCKS[version].groups) {
    for (let index = 0; index < count; index++) {
      blocks.push(data.slice(offset, offset + size))
      offset += size
    }
  }
  return blocks
}

function interleave(data: number[], version: number): number[] {
  const { ecPerBlock } = LEVEL_M_BLOCKS[version]
  const blocks = splitBlocks(data, version)
  const divisor = reedSolomonDivisor(ecPerBlock)
  const corrections = blocks.map((block) => reedSolomonRemainder(block, divisor))
  const longest = Math.max(...blocks.map((block) => block.length))
  const result: number[] = []
  for (let index = 0; index < longest; index++) {
    blocks.forEach((block) => {
      if (index < block.length) {
        result.push(block[index])
      }
    })
  }
  for (let index = 0; index < ecPerBlock; index++) {
    corrections.forEach((correction) => result.push(correction[index]))
  }
  return result
}

export function formatBits(mask: number): number {
  const data = (LEVEL_M_FORMAT << 3) | mask
  let remainder = data
  for (let step = 0; step < 10; step++) {
    remainder = (remainder << 1) ^ ((remainder >>> 9) * FORMAT_GENERATOR)
  }
  return ((data << 10) | remainder) ^ FORMAT_MASK
}

export function versionBits(version: number): number {
  let remainder = version
  for (let step = 0; step < 12; step++) {
    remainder = (remainder << 1) ^ ((remainder >>> 11) * VERSION_GENERATOR)
  }
  return (version << 12) | remainder
}

class Grid {
  readonly size: number
  readonly modules: boolean[][]
  readonly reserved: boolean[][]

  constructor(size: number) {
    this.size = size
    this.modules = Array.from({ length: size }, () => Array<boolean>(size).fill(false))
    this.reserved = Array.from({ length: size }, () => Array<boolean>(size).fill(false))
  }

  setFunction(row: number, column: number, dark: boolean) {
    this.modules[row][column] = dark
    this.reserved[row][column] = true
  }
}

function drawFinder(grid: Grid, centerRow: number, centerColumn: number) {
  for (let dy = -4; dy <= 4; dy++) {
    for (let dx = -4; dx <= 4; dx++) {
      const row = centerRow + dy
      const column = centerColumn + dx
      const distance = Math.max(Math.abs(dx), Math.abs(dy))
      if (row >= 0 && row < grid.size && column >= 0 && column < grid.size) {
        grid.setFunction(row, column, distance !== 2 && distance !== 4)
      }
    }
  }
}

function drawAlignment(grid: Grid, centerRow: number, centerColumn: number) {
  for (let dy = -2; dy <= 2; dy++) {
    for (let dx = -2; dx <= 2; dx++) {
      grid.setFunction(centerRow + dy, centerColumn + dx, Math.max(Math.abs(dx), Math.abs(dy)) !== 1)
    }
  }
}

function drawFormat(grid: Grid, mask: number) {
  const bits = formatBits(mask)
  const last = grid.size - 1
  for (let index = 0; index <= 5; index++) {
    grid.setFunction(index, 8, bit(bits, index))
  }
  grid.setFunction(7, 8, bit(bits, 6))
  grid.setFunction(8, 8, bit(bits, 7))
  grid.setFunction(8, 7, bit(bits, 8))
  for (let index = 9; index < 15; index++) {
    grid.setFunction(8, 14 - index, bit(bits, index))
  }
  for (let index = 0; index < 8; index++) {
    grid.setFunction(8, last - index, bit(bits, index))
  }
  for (let index = 8; index < 15; index++) {
    grid.setFunction(grid.size - 15 + index, 8, bit(bits, index))
  }
  grid.setFunction(grid.size - 8, 8, true)
}

function drawVersion(grid: Grid, version: number) {
  if (version < 7) {
    return
  }
  const bits = versionBits(version)
  for (let index = 0; index < 18; index++) {
    const far = grid.size - 11 + (index % 3)
    const near = Math.floor(index / 3)
    grid.setFunction(near, far, bit(bits, index))
    grid.setFunction(far, near, bit(bits, index))
  }
}

function drawFunctionPatterns(grid: Grid, version: number) {
  for (let index = 0; index < grid.size; index++) {
    grid.setFunction(6, index, index % 2 === 0)
    grid.setFunction(index, 6, index % 2 === 0)
  }
  drawFinder(grid, 3, 3)
  drawFinder(grid, 3, grid.size - 4)
  drawFinder(grid, grid.size - 4, 3)

  const centers = ALIGNMENT_CENTERS[version]
  const last = centers.length - 1
  centers.forEach((row, rowIndex) => {
    centers.forEach((column, columnIndex) => {
      const overlapsFinder =
        (rowIndex === 0 && columnIndex === 0) || (rowIndex === 0 && columnIndex === last) || (rowIndex === last && columnIndex === 0)
      if (!overlapsFinder) {
        drawAlignment(grid, row, column)
      }
    })
  })
  drawFormat(grid, 0)
  drawVersion(grid, version)
}

function drawCodewords(grid: Grid, codewords: number[]) {
  const totalBits = codewords.length * 8
  let position = 0
  for (let right = grid.size - 1; right >= 1; right -= 2) {
    const pair = right <= 6 ? right - 1 : right
    const upward = ((pair + 1) & 2) === 0
    for (let step = 0; step < grid.size; step++) {
      const row = upward ? grid.size - 1 - step : step
      for (const column of [pair, pair - 1]) {
        if (!grid.reserved[row][column] && position < totalBits) {
          grid.modules[row][column] = bit(codewords[position >>> 3], 7 - (position & 7))
          position++
        }
      }
    }
  }
}

export function maskApplies(mask: number, row: number, column: number): boolean {
  const product = row * column
  switch (mask) {
    case 0:
      return (row + column) % 2 === 0
    case 1:
      return row % 2 === 0
    case 2:
      return column % 3 === 0
    case 3:
      return (row + column) % 3 === 0
    case 4:
      return (Math.floor(row / 2) + Math.floor(column / 3)) % 2 === 0
    case 5:
      return (product % 2) + (product % 3) === 0
    case 6:
      return ((product % 2) + (product % 3)) % 2 === 0
    default:
      return (((row + column) % 2) + (product % 3)) % 2 === 0
  }
}

function applyMask(grid: Grid, mask: number) {
  for (let row = 0; row < grid.size; row++) {
    for (let column = 0; column < grid.size; column++) {
      if (!grid.reserved[row][column] && maskApplies(mask, row, column)) {
        grid.modules[row][column] = !grid.modules[row][column]
      }
    }
  }
}

function lines(modules: boolean[][]): boolean[][] {
  const columns = modules[0].map((_, column) => modules.map((row) => row[column]))
  return [...modules, ...columns]
}

function runPenalty(line: boolean[]): number {
  let penalty = 0
  let run = 1
  for (let index = 1; index <= line.length; index++) {
    if (index < line.length && line[index] === line[index - 1]) {
      run++
      continue
    }
    if (run >= 5) {
      penalty += 3 + (run - 5)
    }
    run = 1
  }
  return penalty
}

function finderLikePenalty(line: boolean[]): number {
  const quiet = [false, false, false, false]
  const patterns = [[...FINDER_PATTERN, ...quiet], [...quiet, ...FINDER_PATTERN]]
  let penalty = 0
  for (let start = 0; start + 11 <= line.length; start++) {
    for (const pattern of patterns) {
      if (pattern.every((dark, offset) => line[start + offset] === dark)) {
        penalty += 40
      }
    }
  }
  return penalty
}

function blockPenalty(modules: boolean[][]): number {
  let penalty = 0
  for (let row = 0; row + 1 < modules.length; row++) {
    for (let column = 0; column + 1 < modules.length; column++) {
      const color = modules[row][column]
      if (modules[row][column + 1] === color && modules[row + 1][column] === color && modules[row + 1][column + 1] === color) {
        penalty += 3
      }
    }
  }
  return penalty
}

function balancePenalty(modules: boolean[][]): number {
  const total = modules.length * modules.length
  const dark = modules.reduce((sum, row) => sum + row.filter(Boolean).length, 0)
  return Math.floor(Math.abs((dark * 100) / total - 50) / 5) * 10
}

export function penaltyScore(modules: boolean[][]): number {
  const linePenalty = lines(modules).reduce((sum, line) => sum + runPenalty(line) + finderLikePenalty(line), 0)
  return linePenalty + blockPenalty(modules) + balancePenalty(modules)
}

function cloneGrid(source: Grid): Grid {
  const copy = new Grid(source.size)
  source.modules.forEach((row, index) => {
    copy.modules[index] = [...row]
    copy.reserved[index] = [...source.reserved[index]]
  })
  return copy
}

// forcedMask hanya untuk pengujian; tanpa itu mask dipilih dari penalti terkecil.
export function encodeQr(text: string, forcedMask?: number): QrMatrix {
  const bytes = new TextEncoder().encode(text)
  const version = chooseVersion(bytes.length)
  const base = new Grid(version * 4 + 17)
  drawFunctionPatterns(base, version)
  drawCodewords(base, interleave(encodeData(bytes, version), version))

  const masks = forcedMask === undefined ? [0, 1, 2, 3, 4, 5, 6, 7] : [forcedMask]
  let best: { grid: Grid; mask: number; score: number } | null = null
  for (const mask of masks) {
    const candidate = cloneGrid(base)
    applyMask(candidate, mask)
    drawFormat(candidate, mask)
    const score = penaltyScore(candidate.modules)
    if (best === null || score < best.score) {
      best = { grid: candidate, mask, score }
    }
  }
  const chosen = best as { grid: Grid; mask: number }
  return { version, size: chosen.grid.size, mask: chosen.mask, modules: chosen.grid.modules }
}
