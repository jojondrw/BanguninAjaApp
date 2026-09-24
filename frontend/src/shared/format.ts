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
  if (Math.abs(amount) >= BILLION) {
    return `Rp ${NUMBER.format(Number((amount / BILLION).toFixed(1)))} M`
  }
  if (Math.abs(amount) >= MILLION) {
    return `Rp ${NUMBER.format(Math.round(amount / MILLION))} jt`
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
