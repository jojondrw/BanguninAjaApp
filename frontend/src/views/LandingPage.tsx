import { ArrowRight, Check, ShieldCheck } from 'lucide-react'
import { useEffect } from 'react'
import { Link } from 'react-router-dom'

import { BrandName } from './components/Brand'

const BANI = '/images/bani'

const STAGES = [
  { pose: 'map-pin', title: 'Lokasi', text: 'Skor lahan dari data resmi, lengkap dengan asal-usul tiap angkanya.' },
  { pose: 'blueprint', title: 'Proyek', text: 'Tahap, RAB, kurva S, laporan harian, dan foto kondisi lapangan.' },
  { pose: 'cart', title: 'Pengadaan', text: 'Permintaan, purchase order, dan penerimaan barang yang langsung jadi hutang.' },
  { pose: 'boxes', title: 'Inventaris', text: 'Stok per gudang, mutasi material, dan peringatan saat stok menipis.' },
  { pose: 'team', title: 'SDM', text: 'Karyawan tetap, kontrak, dan harian, absensi, sampai penggajian.' },
  { pose: 'house-key', title: 'Penjualan', text: 'Unit, kontrak, cicilan, sampai kunci diserahkan ke pembeli.' },
]

const SOURCES = ['BPS', 'BNPB InaRISK', 'ATR/BPN: zona nilai tanah dan RDTR', 'Pusiknas Polri', 'OpenStreetMap']

const FINANCE_POINTS = [
  'Pembayaran tagihan langsung tercatat di kas dan jurnal',
  'Termin pemberi kerja dan kontraktor, lengkap dengan retensi',
  'Pekerjaan tambah-kurang yang mengubah nilai kontrak',
  'Pembayaran on-chain di jaringan uji Ethereum',
]

const ROLES = ['Admin', 'Manajer Proyek', 'Keuangan', 'Pengadaan', 'Penjualan', 'SDM', 'Pengamat']

// Halaman depan untuk pengunjung yang belum masuk. Pengguna yang sudah masuk
// langsung melihat Ringkasan di alamat yang sama.
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
        <section className="mx-auto grid max-w-6xl items-center gap-10 px-4 pt-14 pb-20 sm:px-6 md:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)] md:pt-20 md:pb-28">
          <div className="motion-safe:animate-enter">
            <p className="text-[13px] font-semibold text-navy-600">ERP untuk developer properti dan kontraktor</p>
            <h1 className="mt-3 text-[40px] leading-[1.05] font-semibold tracking-[-0.035em] sm:text-[56px]">
              Dari memilih lahan sampai menyerahkan kunci.
            </h1>
            <p className="mt-5 max-w-xl text-[17px] leading-relaxed text-slate-600">
              BanguninAja menilai lokasi dengan data resmi, lalu mengelola proyek, pengadaan, gudang, tenaga kerja,
              keuangan, dan penjualan unit dalam satu aplikasi.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <PrimaryLink to="/daftar">Mulai sekarang</PrimaryLink>
              <SecondaryLink to="/masuk">Masuk ke akun</SecondaryLink>
            </div>
          </div>
          <div className="relative mx-auto w-full max-w-[420px]">
            <div
              aria-hidden="true"
              className="absolute inset-[6%] rounded-full bg-[radial-gradient(circle_at_50%_45%,var(--color-navy-100),transparent_70%)]"
            />
            <img
              src={`${BANI}/pose-wave.webp`}
              alt="Bani, maskot BanguninAja, berang-berang berhelm proyek yang melambaikan tangan"
              width={591}
              height={640}
              fetchPriority="high"
              className="relative mx-auto h-auto w-[78%] motion-safe:animate-enter"
            />
          </div>
        </section>

        <section className="bg-navy-900 text-white">
          <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 md:py-28">
            <p className="max-w-3xl text-[28px] leading-[1.2] font-semibold tracking-[-0.025em] sm:text-[36px]">
              Developer sering memilih lokasi dari intuisi dan harga tanah yang murah. Padahal datanya sudah terbuka.
            </p>
            <p className="mt-5 max-w-2xl text-[16px] leading-relaxed text-white/70">
              BanguninAja menggabungkan data pemerintah dan peta terbuka menjadi satu skor lokasi. Setiap angka bisa
              ditelusuri kembali ke sumbernya, jadi keputusan membeli lahan bisa dipertanggungjawabkan.
            </p>
            <ul className="mt-10 flex flex-wrap gap-2" aria-label="Sumber data">
              {SOURCES.map((source) => (
                <li key={source} className="rounded-full bg-white/[0.08] px-3.5 py-1.5 text-[13px] text-white/85 ring-1 ring-white/10">
                  {source}
                </li>
              ))}
            </ul>
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-4 py-20 sm:px-6 md:py-28" aria-labelledby="alur">
          <div className="max-w-2xl">
            <p className="text-[13px] font-semibold text-navy-600">Satu alur</p>
            <h2 id="alur" className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
              Enam tahap bisnis, satu tempat.
            </h2>
            <p className="mt-4 text-[16px] leading-relaxed text-slate-600">
              Data dari satu tahap langsung dipakai tahap berikutnya. Tidak ada lagi spreadsheet terpisah per divisi.
            </p>
          </div>
          <ol className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {STAGES.map((stage, index) => (
              <li key={stage.title} className="flex items-center gap-4 rounded-3xl bg-slate-50 p-5">
                <img
                  src={`${BANI}/pose-${stage.pose}.webp`}
                  alt=""
                  aria-hidden="true"
                  loading="lazy"
                  className="size-24 shrink-0 object-contain"
                />
                <div className="min-w-0">
                  <p className="text-[12px] font-medium text-slate-500 tabular-nums">Tahap {index + 1}</p>
                  <h3 className="text-[17px] font-semibold tracking-[-0.01em]">{stage.title}</h3>
                  <p className="mt-1 text-[14px] leading-snug text-slate-600">{stage.text}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>

        <section className="bg-slate-50">
          <div className="mx-auto grid max-w-6xl items-center gap-12 px-4 py-20 sm:px-6 md:grid-cols-2 md:py-28">
            <img
              src={`${BANI}/pose-chart.webp`}
              alt=""
              aria-hidden="true"
              loading="lazy"
              className="mx-auto h-auto w-full max-w-[340px] md:order-2"
            />
            <div>
              <p className="text-[13px] font-semibold text-navy-600">Keuangan</p>
              <h2 className="mt-2 text-[32px] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-[40px]">
                Uang proyek yang selalu cocok dengan bukunya.
              </h2>
              <ul className="mt-8 space-y-3.5">
                {FINANCE_POINTS.map((point) => (
                  <li key={point} className="flex items-start gap-3 text-[15px] text-slate-700">
                    <span className="mt-0.5 grid size-5 shrink-0 place-items-center rounded-full bg-navy-700 text-white">
                      <Check aria-hidden="true" className="size-3" strokeWidth={3} />
                    </span>
                    {point}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-4 py-20 sm:px-6 md:py-28" aria-labelledby="peran">
          <div className="grid gap-10 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] md:items-end">
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
          <div className="mx-auto flex max-w-6xl flex-col items-center gap-6 rounded-[32px] bg-navy-900 px-6 py-14 text-center text-white md:flex-row md:px-14 md:text-left">
            <img
              src={`${BANI}/pose-celebrate.webp`}
              alt=""
              aria-hidden="true"
              loading="lazy"
              className="h-auto w-40 shrink-0 md:w-48"
            />
            <div className="flex-1">
              <h2 className="text-[28px] leading-[1.15] font-semibold tracking-[-0.03em] sm:text-[34px]">
                Siap membangun dengan data?
              </h2>
              <p className="mt-3 max-w-xl text-[15px] leading-relaxed text-white/70">
                Buat akun untuk tim Anda. Admin perusahaan menentukan peran tiap anggota, lalu semua bekerja di data yang
                sama.
              </p>
            </div>
            <div className="flex flex-wrap justify-center gap-3">
              <Link
                to="/daftar"
                className="inline-flex items-center gap-1.5 rounded-full bg-white px-5 py-2.5 text-[14px] font-semibold text-navy-800 transition-colors hover:bg-navy-50"
              >
                Buat akun
                <ArrowRight aria-hidden="true" className="size-4" strokeWidth={2} />
              </Link>
              <Link
                to="/masuk"
                className="inline-flex items-center rounded-full px-5 py-2.5 text-[14px] font-semibold text-white ring-1 ring-white/30 transition-colors hover:bg-white/10"
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

function PrimaryLink({ to, children }: { to: string; children: string }) {
  return (
    <Link
      to={to}
      className="inline-flex items-center gap-1.5 rounded-full bg-navy-700 px-5 py-2.5 text-[14px] font-semibold text-white
                 transition-[background-color,transform] hover:bg-navy-800 motion-safe:active:scale-[0.98]"
    >
      {children}
      <ArrowRight aria-hidden="true" className="size-4" strokeWidth={2} />
    </Link>
  )
}

function SecondaryLink({ to, children }: { to: string; children: string }) {
  return (
    <Link
      to={to}
      className="inline-flex items-center rounded-full px-5 py-2.5 text-[14px] font-semibold text-navy-700 ring-1 ring-slate-300
                 transition-colors hover:bg-slate-50"
    >
      {children}
    </Link>
  )
}
