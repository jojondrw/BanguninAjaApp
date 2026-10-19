import { ChevronRight } from 'lucide-react'

// Tautan teks ke tab atau rincian lain di halaman yang sama, dipakai di kartu
// ringkasan untuk "lihat semua".
export function LinkButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex items-center gap-0.5 text-[13px] font-medium whitespace-nowrap text-navy-600 hover:text-navy-800"
    >
      {label}
      <ChevronRight aria-hidden="true" className="size-3.5" strokeWidth={2.2} />
    </button>
  )
}
