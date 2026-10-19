// Tanggal dihitung dari jam peramban, bukan UTC, supaya "hari ini" di Jakarta
// tetap hari ini walaupun di UTC masih kemarin.

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

export function todayDate(): string {
  const now = new Date()
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

// Tanggal pertama bulan, count bulan sebelum bulan ini.
export function monthsAgo(count: number): string {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth() - count, 1)
  return `${start.getFullYear()}-${pad(start.getMonth() + 1)}-01`
}

export function currentPeriod(): string {
  return todayDate().slice(0, 7)
}

export function firstDayOfMonth(date: string): string {
  return `${date.slice(0, 7)}-01`
}

// Isian tanggal memberi "YYYY-MM-DD", sedangkan backend membaca time.Time yang
// butuh RFC 3339 dan menolak tanggal tanpa jam.
export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
