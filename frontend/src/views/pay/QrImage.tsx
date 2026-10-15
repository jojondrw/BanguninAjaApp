import { encodeQr } from './qrCode'

const QUIET_ZONE = 4

// QR digambar sebagai SVG dari penyandi lokal, jadi isi tagihan tidak pernah
// dikirim ke layanan gambar QR pihak ketiga.
export function QrImage({ value, label, className = 'size-56' }: { value: string; label: string; className?: string }) {
  const matrix = encodeQr(value)
  const dimension = matrix.size + QUIET_ZONE * 2
  const path = matrix.modules
    .flatMap((row, rowIndex) =>
      row.map((dark, columnIndex) => (dark ? `M${columnIndex + QUIET_ZONE} ${rowIndex + QUIET_ZONE}h1v1h-1z` : '')),
    )
    .join('')

  return (
    <svg
      role="img"
      aria-label={label}
      viewBox={`0 0 ${dimension} ${dimension}`}
      shapeRendering="crispEdges"
      className={`rounded-xl bg-white ${className}`}
    >
      <rect width={dimension} height={dimension} fill="#ffffff" />
      <path d={path} fill="#111113" />
    </svg>
  )
}
