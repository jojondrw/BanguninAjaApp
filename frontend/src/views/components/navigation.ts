import {
  Boxes,
  FolderKanban,
  Handshake,
  LayoutDashboard,
  type LucideIcon,
  MapPinned,
  ReceiptText,
  ShoppingCart,
  Users,
  Wallet,
} from 'lucide-react'

export interface MenuItem {
  to: string
  label: string
  icon: LucideIcon
  keywords?: string
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
      { to: '/proyek', label: 'Proyek', icon: FolderKanban, keywords: 'project tahap izin rab' },
      { to: '/lokasi', label: 'Analisis Lokasi', icon: MapPinned, keywords: 'peta skor site evaluasi gis' },
      { to: '/keuangan', label: 'Keuangan', icon: Wallet, keywords: 'kas anggaran finance budget' },
      { to: '/tagihan', label: 'Tagihan', icon: ReceiptText, keywords: 'faktur invoice piutang utang pembayaran billing' },
    ],
  },
  {
    label: 'Operasional',
    items: [
      { to: '/penjualan', label: 'Penjualan', icon: Handshake, keywords: 'unit pelanggan kontrak cicilan sales' },
      { to: '/pengadaan', label: 'Pengadaan', icon: ShoppingCart, keywords: 'vendor po pembelian procurement' },
      { to: '/inventaris', label: 'Inventaris', icon: Boxes, keywords: 'stok gudang material inventory' },
      { to: '/sdm', label: 'SDM', icon: Users, keywords: 'karyawan absensi gaji hr payroll' },
    ],
  },
]

export const MENU_ITEMS = MENU.flatMap((group) => group.items)
