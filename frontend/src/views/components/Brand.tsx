// Logo BanguninAja: tiga gedung berdiri di atas tiga baris data. Proyek
// dibangun berdasarkan data, lebar keduanya sama dan tebal garisnya sama.
const BUILDINGS = [0, 21, 42]
const ROWS = [61, 82, 103]

export function BrandMark({ className = 'h-7 w-auto text-navy-700' }: { className?: string }) {
  return (
    <svg viewBox="0 0 55 116" width={13} height={28} fill="currentColor" aria-hidden="true" className={className}>
      {BUILDINGS.map((x) => (
        <rect key={`b${x}`} x={x} y={0} width={13} height={53} />
      ))}
      {ROWS.map((y) => (
        <rect key={`r${y}`} x={0} y={y} width={55} height={13} />
      ))}
    </svg>
  )
}

export function BrandName({ className = '' }: { className?: string }) {
  return (
    <span className={`flex items-center gap-2 ${className}`}>
      <BrandMark />
      <span className="text-[15px] font-semibold tracking-[-0.02em] text-slate-900">BanguninAja</span>
    </span>
  )
}
