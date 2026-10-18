import { Building, Building2, Hospital, House, Layers, Store, type LucideIcon } from 'lucide-react'
import { createElement } from 'react'

// Ikon dari jenis proyek yang diketik bebas. "Rumah Sakit" dicek sebelum
// "rumah" supaya tidak terbaca sebagai perumahan.
const ICON_BY_TYPE: [RegExp, LucideIcon][] = [
  [/rumah sakit|klinik|kesehatan/i, Hospital],
  [/apartemen|rusun|menara|tower/i, Building],
  [/ruko|komersial|toko|mall|belanja|kantor/i, Store],
  [/campuran|mixed/i, Layers],
  [/perumahan|hunian|rumah|cluster|klaster/i, House],
]

function projectTypeIcon(type: string): LucideIcon {
  return ICON_BY_TYPE.find(([pattern]) => pattern.test(type))?.[1] ?? Building2
}

export function ProjectTypeIcon({ type, className = 'size-4' }: { type: string; className?: string }) {
  return createElement(projectTypeIcon(type), { 'aria-hidden': true, className, strokeWidth: 1.75 })
}
