import type { OptionSource } from '../../models/lookup'
import { useLookupLabel } from './searchSelectLogic'

// Nama data untuk sel tabel atau teks rincian yang hanya membawa id.
export function LookupName({ source, value, fallback = '-' }: { source: OptionSource; value: string | null; fallback?: string }) {
  const label = useLookupLabel(source, value)
  return <>{label ?? fallback}</>
}
