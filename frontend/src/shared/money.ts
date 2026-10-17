// Nilai uang di dalam teks, sesuai keluaran rupiah() dan rupiahShort():
// "Rp 1.250.000", "-Rp 500", "−Rp 2,5 M", "+Rp 1,3 M", "Rp 48 jt". Angka harus berakhir
// dengan digit supaya titik penutup kalimat tidak ikut terambil.
const MONEY = /[-−+]?Rp\s?\d(?:[\d.,]*\d)?(?:\s(?:M|jt)\b)?/g

export interface TextPart {
  text: string
  money: boolean
}

export function splitMoney(text: string): TextPart[] {
  const parts: TextPart[] = []
  let last = 0
  for (const match of text.matchAll(MONEY)) {
    const start = match.index
    if (start > last) {
      parts.push({ text: text.slice(last, start), money: false })
    }
    parts.push({ text: match[0], money: true })
    last = start + match[0].length
  }
  if (last < text.length) {
    parts.push({ text: text.slice(last), money: false })
  }
  return parts
}
