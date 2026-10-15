import {
  Boxes,
  FolderKanban,
  Handshake,
  LayoutDashboard,
  type LucideIcon,
  MapPinned,
  ReceiptText,
  ShoppingCart,
  UserCog,
  Users,
  Wallet,
} from 'lucide-react'

import type { ModuleKey } from '../../models/auth'

export interface MenuItem {
  to: string
  label: string
  icon: LucideIcon
  keywords?: string
  // Modul yang wajib bisa dibaca untuk melihat menu dan membuka rutenya.
  // Tanpa modul berarti terbuka untuk semua yang sudah masuk.
  module?: ModuleKey
}

export interface MenuGroup {
  label: string
  items: MenuItem[]
}

export const MENU: MenuGroup[] = [
  {
    label: 'Ruang kerja',
    items: [
      { to: '/', label: 'Ringkasan', icon: LayoutDashboard, keywords: 'dasbor beranda home kpi' },
      { to: '/proyek', label: 'Proyek', icon: FolderKanban, keywords: 'project tahap izin rab', module: 'projects' },
      { to: '/lokasi', label: 'Analisis Lokasi', icon: MapPinned, keywords: 'peta skor site evaluasi gis', module: 'locations' },
      { to: '/keuangan', label: 'Keuangan', icon: Wallet, keywords: 'kas anggaran finance budget', module: 'finance' },
      { to: '/tagihan', label: 'Tagihan', icon: ReceiptText, keywords: 'faktur invoice piutang utang pembayaran billing', module: 'billing' },
    ],
  },
  {
    label: 'Operasional',
    items: [
      { to: '/penjualan', label: 'Penjualan', icon: Handshake, keywords: 'unit pelanggan kontrak cicilan sales', module: 'sales' },
      { to: '/pengadaan', label: 'Pengadaan', icon: ShoppingCart, keywords: 'vendor po pembelian procurement', module: 'procurement' },
      { to: '/inventaris', label: 'Inventaris', icon: Boxes, keywords: 'stok gudang material inventory', module: 'inventory' },
      { to: '/sdm', label: 'SDM', icon: Users, keywords: 'karyawan absensi gaji hr payroll', module: 'hr' },
    ],
  },
  {
    label: 'Pengaturan',
    items: [
      { to: '/pengguna', label: 'Pengguna', icon: UserCog, keywords: 'user peran role akses admin', module: 'users' },
    ],
  },
]

export const MENU_ITEMS = MENU.flatMap((group) => group.items)

type CanRead = (module: ModuleKey) => boolean

export function visibleMenu(canRead: CanRead): MenuGroup[] {
  return MENU.map((group) => ({ ...group, items: group.items.filter((item) => isVisible(item, canRead)) })).filter(
    (group) => group.items.length > 0,
  )
}

export function isVisible(item: MenuItem, canRead: CanRead): boolean {
  return item.module === undefined || canRead(item.module)
}

// Modul milik sebuah alamat, termasuk sub-halaman seperti /proyek/:id.
export function moduleForPath(pathname: string): ModuleKey | null {
  const item = MENU_ITEMS.find(
    (candidate) => candidate.to !== '/' && (pathname === candidate.to || pathname.startsWith(`${candidate.to}/`)),
  )
  return item?.module ?? null
}
