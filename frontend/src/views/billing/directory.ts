import type { PartyType } from '../../models/billing'
import type { Vendor } from '../../models/procurement'
import type { Project } from '../../models/project'
import type { Customer } from '../../models/sales'
import { rupiah } from '../../shared/format'

// Respons tagihan hanya membawa id pelanggan, vendor, dan proyek. Daftar ini
// diambil sekali di halaman untuk menampilkan nama dan mengisi pilihan formulir.
export interface Directory {
  customers: Customer[]
  vendors: Vendor[]
  projects: Project[]
  isLoading: boolean
  isError: boolean
}

function shortId(id: string): string {
  return id.slice(0, 8)
}

export function customerName(directory: Directory, id: string): string {
  return directory.customers.find((customer) => customer.id === id)?.name ?? shortId(id)
}

export function vendorName(directory: Directory, id: string): string {
  return directory.vendors.find((vendor) => vendor.id === id)?.name ?? shortId(id)
}

export function partyName(directory: Directory, partyType: PartyType, id: string): string {
  return partyType === 'customer' ? customerName(directory, id) : vendorName(directory, id)
}

export function projectName(directory: Directory, id: string | null): string {
  if (id === null) {
    return 'Tanpa proyek'
  }
  return directory.projects.find((project) => project.id === id)?.name ?? shortId(id)
}

export function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return fallback
  }
  return rupiah(amount)
}
