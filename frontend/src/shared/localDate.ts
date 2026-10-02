// Tanggal dihitung dari jam peramban, bukan UTC, supaya "hari ini" di Jakarta
// tetap hari ini walaupun di UTC masih kemarin.

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

export function todayDate(): string {
  const now = new Date()
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
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
