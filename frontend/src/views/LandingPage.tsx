import {
  ArrowRight,
  Boxes,
  Check,
  FolderKanban,
  Handshake,
  MapPinned,
  ShieldCheck,
  ShoppingCart,
  Users,
  type LucideIcon,
} from 'lucide-react'
import { useEffect } from 'react'
import { Link } from 'react-router-dom'

import { PHOTOS } from '../shared/photos'
import { BrandName } from './components/Brand'

const STAGES: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: MapPinned, title: 'Lokasi', text: 'Skor lahan dari data resmi, lengkap dengan asal-usul tiap angkanya.' },
  { icon: FolderKanban, title: 'Proyek', text: 'Tahap, RAB, kurva S, laporan harian, dan foto kondisi lapangan.' },
  { icon: ShoppingCart, title: 'Pengadaan', text: 'Permintaan, purchase order, dan penerimaan barang yang langsung jadi hutang.' },
  { icon: Boxes, title: 'Inventaris', text: 'Stok per gudang, mutasi material, dan peringatan saat stok menipis.' },
  { icon: Users, title: 'SDM', text: 'Karyawan tetap, kontrak, dan harian, absensi, sampai penggajian.' },
  { icon: Handshake, title: 'Penjualan', text: 'Unit, kontrak, cicilan, sampai kunci diserahkan ke pembeli.' },
]

const SOURCES = ['BPS', 'BNPB InaRISK', 'ATR/BPN: zona nilai tanah dan RDTR', 'Pusiknas Polri', 'OpenStreetMap']

const FIELD_POINTS = [
  'Progres berbobot RAB dan kurva S rencana lawan realisasi',
  'Laporan harian: cuaca, tenaga kerja, pekerjaan, kendala',
  'Foto lapangan terbaru jadi sampul tiap proyek',
  'Izin proyek seperti KKPR, PBG, dan SLF lengkap dengan masa berlakunya',
]

const FINANCE_POINTS = [
  'Pembayaran tagihan langsung tercatat di kas dan jurnal',
  'Termin pemberi kerja dan kontraktor, lengkap dengan retensi',
  'Pekerjaan tambah-kurang yang mengubah nilai kontrak',
  'Pembayaran on-chain di jaringan uji Ethereum',
]

const ROLES = ['Admin', 'Manajer Proyek', 'Keuangan', 'Pengadaan', 'Penjualan', 'SDM', 'Pengamat']

// Halaman depan untuk pengunjung yang belum masuk. Pengguna yang sudah masuk
// langsung melihat Ringkasan di alamat yang sama. Foto proyek membawa suasana
// konstruksi; maskot Bani hanya muncul kecil di ajakan penutup.
export function LandingPage() {
  useEffect(() => {
    document.title = 'BanguninAja · Dari memilih lahan sampai menyerahkan kunci'
  }, [])

  return (
    <div className="min-h-screen bg-white text-slate-900">
      <header className="sticky top-0 z-20 border-b border-slate-200/60 bg-white/80 backdrop-blur-xl">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4 sm:px-6">
          <Link to="/" aria-label="BanguninAja">
            <BrandName />
          </Link>
          <nav className="flex items-center gap-1.5" aria-label="Akun">
            <Link
              to="/masuk"
              className="rounded-full px-3.5 py-1.5 text-[13px] font-medium text-slate-700 transition-colors hover:bg-slate-100"
            >
              Masuk
            </Link>
            <Link
              to="/daftar"
              className="rounded-full bg-navy-700 px-3.5 py-1.5 text-[13px] font-medium text-white transition-colors hover:bg-navy-800"
            >
              Daftar
            </Link>
          </nav>
        </div>
      </header>

      <main>
        <section className="px-3 pt-3 sm:px-4">
          <div className="relative mx-auto flex min-h-[560px] max-w-[1400px] items-end overflow-hidden rounded-[28px] bg-navy-900 text-white md:min-h-[640px]">
            <img
              src={PHOTOS.konstruksi.src}
              alt={PHOTOS.konstruksi.alt}
              fetchPriority="high"
              className="absolute inset-0 size-full object-cover object-[50%_30%]"
            />
            <div
              aria-hidden="true"
              className="absolute inset-0 bg-[linear-gradient(to_top,rgb(7_28_70/0.95)_0%,rgb(7_28_70/0.75)_40%,rgb(7_28_70/0.15)_80%)]"
            />
            <div className="relative w-full px-6 pb-10 sm:px-12 sm:pb-14 motion-safe:animate-enter">
              <p className="text-[13px] font-semibold text-[#FFD60A]">ERP untuk developer properti dan kontraktor</p>
              <h1 className="mt-3 max-w-3xl text-[40px] leading-[1.05] font-semibold tracking-[-0.035em] sm:text-[60px]">
                Dari memilih lahan sampai menyerahkan kunci.
              </h1>
              <p className="mt-5 max-w-2xl text-[17px] leading-relaxed text-white/75">
                BanguninAja menilai lokasi dengan data resmi, lalu mengelola proyek, pengadaan, gudang, tenaga kerja,
                keuangan, dan penjualan unit dalam satu aplikasi.
              </p>
              <div className="mt-8 flex flex-wrap gap-3">
                <Link
                  to="/daftar"
                  className="inline-flex items-center gap-1.5 rounded-full bg-white px-5 py-2.5 text-[14px] font-semibold text-navy-800
                             transition-[background-color,transform] hover:bg-navy-50 motion-safe:active:scale-[0.98]"
                >
                  Mulai sekarang
                  <ArrowRight aria-hidden="true" className="size-4" strokeWidth={2} />
                </Link>
                <Link
                  to="/masuk"
                  className="inline-flex items-center rounded-full px-5 py-2.5 text-[14px] font-semibold text-white ring-1 ring-white/35
                             backdrop-blur-sm transition-colors hover:bg-white/10"
                >
                  Masuk ke akun
                </Link>
              </div>
            </div>
          </div>
        </section>

        <section className="mx-auto grid max-w-6xl items-center gap-12 px-4 py-20 sm:px-6 md:grid-cols-2 md:py-28">
          <div>
            <p className="text-[13px] font-semibold text-navy-600">Analisis lokasi</p>
            <h2 className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
              Lahan dipilih dari data, bukan dari firasat.
            </h2>
            <p className="mt-5 text-[16px] leading-relaxed text-slate-600">
              Developer sering memilih lokasi dari intuisi dan harga tanah yang murah. BanguninAja menggabungkan data
              pemerintah dan peta terbuka menjadi satu skor lokasi, dan setiap angkanya bisa ditelusuri ke sumbernya.
            </p>
            <ul className="mt-8 flex flex-wrap gap-2" aria-label="Sumber data">
              {SOURCES.map((source) => (
                <li key={source} className="rounded-full bg-slate-100 px-3.5 py-1.5 text-[13px] font-medium text-slate-700">
                  {source}
                </li>
              ))}
            </ul>
          </div>
          <PhotoPanel photo={PHOTOS.lokasi} className="aspect-[4/3]" />
        </section>

        <section className="bg-slate-50" aria-labelledby="alur">
          <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 md:py-28">
            <div className="max-w-2xl">
              <p className="text-[13px] font-semibold text-navy-600">Satu alur</p>
              <h2 id="alur" className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
                Enam tahap bisnis, satu tempat.
              </h2>
              <p className="mt-4 text-[16px] leading-relaxed text-slate-600">
                Data dari satu tahap langsung dipakai tahap berikutnya. Tidak ada lagi spreadsheet terpisah per divisi.
              </p>
            </div>
            <ol className="mt-12 grid gap-px overflow-hidden rounded-3xl bg-slate-200/80 sm:grid-cols-2 lg:grid-cols-3">
              {STAGES.map((stage, index) => {
                const Icon = stage.icon
                return (
                  <li key={stage.title} className="bg-white p-6">
                    <div className="flex items-center justify-between">
                      <span className="grid size-10 place-items-center rounded-xl bg-navy-700 text-[#FFD60A]">
                        <Icon aria-hidden="true" className="size-5" strokeWidth={1.75} />
                      </span>
                      <span className="text-[13px] font-semibold text-slate-300 tabular-nums">
                        {String(index + 1).padStart(2, '0')}
                      </span>
                    </div>
                    <h3 className="mt-5 text-[17px] font-semibold tracking-[-0.01em]">{stage.title}</h3>
                    <p className="mt-1.5 text-[14px] leading-snug text-slate-600">{stage.text}</p>
                  </li>
                )
              })}
            </ol>
          </div>
        </section>

        <section className="mx-auto grid max-w-6xl items-center gap-12 px-4 py-20 sm:px-6 md:grid-cols-2 md:py-28">
          <PhotoPanel photo={PHOTOS.konstruksi} className="aspect-[4/3] md:aspect-[4/5]" position="object-[50%_60%]" />
          <div>
            <p className="text-[13px] font-semibold text-navy-600">Lapangan</p>
            <h2 className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
              Kantor tahu keadaan proyek tanpa harus ke lokasi.
            </h2>
            <CheckList points={FIELD_POINTS} />
          </div>
        </section>

        <section className="bg-navy-900 text-white">
          <div className="mx-auto grid max-w-6xl items-center gap-12 px-4 py-20 sm:px-6 md:grid-cols-2 md:py-28">
            <div>
              <p className="text-[13px] font-semibold text-[#FFD60A]">Keuangan</p>
              <h2 className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
                Uang proyek yang selalu cocok dengan bukunya.
              </h2>
              <CheckList points={FINANCE_POINTS} tone="dark" />
            </div>
            <PhotoPanel photo={PHOTOS.perumahan} className="aspect-[4/3] md:aspect-[4/5]" />
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-4 py-20 sm:px-6 md:py-28" aria-labelledby="peran">
          <div className="grid gap-10 md:grid-cols-2 md:items-end">
            <div>
              <p className="text-[13px] font-semibold text-navy-600">Akses</p>
              <h2 id="peran" className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
                Tujuh peran, masing-masing melihat yang perlu saja.
              </h2>
            </div>
            <p className="flex items-start gap-3 text-[15px] leading-relaxed text-slate-600">
              <ShieldCheck aria-hidden="true" className="mt-0.5 size-5 shrink-0 text-navy-600" strokeWidth={1.75} />
              Hak baca dan tulis diatur per modul, proyek bisa dibatasi untuk anggota timnya, dan NIK pelanggan maupun
              karyawan disamarkan untuk peran yang tidak mengurusnya.
            </p>
          </div>
          <ul className="mt-10 flex flex-wrap gap-2" aria-label="Peran pengguna">
            {ROLES.map((role) => (
              <li key={role} className="rounded-full bg-slate-100 px-4 py-2 text-[14px] font-medium text-slate-700">
                {role}
              </li>
            ))}
          </ul>
        </section>

        <section className="px-4 pb-20 sm:px-6 md:pb-28">
          <div className="mx-auto flex max-w-6xl flex-col items-center gap-6 rounded-[32px] bg-slate-100 px-6 py-12 text-center md:flex-row md:px-12 md:text-left">
            <img
              src="/images/bani/pose-blueprint.webp"
              alt="Bani, maskot BanguninAja, membawa gambar kerja"
              loading="lazy"
              className="h-auto w-24 shrink-0 md:w-28"
            />
            <div className="flex-1">
              <h2 className="text-[26px] leading-[1.15] font-semibold tracking-[-0.03em] sm:text-[32px]">
                Siap membangun dengan data?
              </h2>
              <p className="mt-2 max-w-xl text-[15px] leading-relaxed text-slate-600">
                Buat akun untuk tim Anda. Admin perusahaan menentukan peran tiap anggota, lalu semua bekerja di data yang
                sama.
              </p>
            </div>
            <div className="flex flex-wrap justify-center gap-3">
              <Link
                to="/daftar"
                className="inline-flex items-center gap-1.5 rounded-full bg-navy-700 px-5 py-2.5 text-[14px] font-semibold text-white transition-colors hover:bg-navy-800"
              >
                Buat akun
                <ArrowRight aria-hidden="true" className="size-4" strokeWidth={2} />
              </Link>
              <Link
                to="/masuk"
                className="inline-flex items-center rounded-full bg-white px-5 py-2.5 text-[14px] font-semibold text-navy-700 ring-1 ring-slate-300 transition-colors hover:bg-slate-50"
              >
                Masuk
              </Link>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-slate-200/70">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-6 text-[13px] text-slate-500 sm:px-6">
          <BrandName />
          <p>© 2026 BanguninAja</p>
        </div>
      </footer>
    </div>
  )
}

function PhotoPanel({ photo, className, position = 'object-center' }: {
  photo: { src: string; alt: string }
  className: string
  position?: string
}) {
  return (
    <div className={`overflow-hidden rounded-[28px] bg-slate-200 ${className}`}>
      <img src={photo.src} alt={photo.alt} loading="lazy" className={`size-full object-cover ${position}`} />
    </div>
  )
}

function CheckList({ points, tone = 'light' }: { points: string[]; tone?: 'light' | 'dark' }) {
  return (
    <ul className="mt-8 space-y-3.5">
      {points.map((point) => (
        <li key={point} className={`flex items-start gap-3 text-[15px] ${tone === 'dark' ? 'text-white/85' : 'text-slate-700'}`}>
          <span
            className={`mt-0.5 grid size-5 shrink-0 place-items-center rounded-full ${
              tone === 'dark' ? 'bg-[#FFD60A] text-navy-900' : 'bg-navy-700 text-white'
            }`}
          >
            <Check aria-hidden="true" className="size-3" strokeWidth={3} />
          </span>
          {point}
        </li>
      ))}
    </ul>
  )
}
