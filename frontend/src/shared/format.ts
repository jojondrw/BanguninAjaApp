const RUPIAH = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})

const NUMBER = new Intl.NumberFormat('id-ID')

const BILLION = 1_000_000_000
const MILLION = 1_000_000

export function rupiah(amount: number): string {
  return RUPIAH.format(amount)
}

export function rupiahShort(amount: number): string {
  const sign = amount < 0 ? '−' : ''
  const size = Math.abs(amount)
  if (size >= BILLION) {
    return `${sign}Rp ${NUMBER.format(Number((size / BILLION).toFixed(1)))} M`
  }
  if (size >= MILLION) {
    return `${sign}Rp ${NUMBER.format(Math.round(size / MILLION))} jt`
  }
  return RUPIAH.format(amount)
}

export function number(value: number): string {
  return NUMBER.format(value)
}

export function shortDate(iso: string | null): string {
  if (!iso) {
    return '-'
  }
  return new Date(iso).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export function dateRange(start: string | null, end: string | null): string {
  if (!start && !end) {
    return 'Jadwal belum diisi'
  }
  return `${shortDate(start)} sampai ${shortDate(end)}`
}

// Nilai kartu KPI selama data dimuat atau gagal dimuat, supaya tidak tampil
// angka nol yang menyesatkan.
export function kpiValue(isPending: boolean, isError: boolean, value: string): string {
  if (isError) {
    return '—'
  }
  return isPending ? '...' : value
}

export function monthLabel(period: string): string {
  const [year, month] = period.split('-')
  if (!month) {
    return period
  }
  return new Date(Number(year), Number(month) - 1).toLocaleDateString('id-ID', {
    month: 'short',
    year: 'numeric',
  })
}
