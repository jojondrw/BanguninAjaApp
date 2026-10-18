const BASE = '/images/bani'
const STILL = `${BASE}/build-house-still.webp`

type Surface = 'page' | 'card'

// Video dirender di atas warna latar tempat ia tampil (abu halaman atau kartu
// putih), karena video tidak punya transparansi di semua browser.
const VIDEO: Record<Surface, string> = {
  page: `${BASE}/build-house.mp4`,
  card: `${BASE}/build-house-card.mp4`,
}

function prefersReducedMotion(): boolean {
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
}

// Bani membangun rumah dari fondasi sampai atap. Frame-nya sudah distabilkan
// (ukuran Bani sama di semua frame) dan diberi frame sisipan supaya mengalir,
// lalu disimpan sebagai video pendek yang ringan. Pengguna yang memilih
// mengurangi gerak melihat gambar diam.
function BuildingAnimation({ height, surface }: { height: number; surface: Surface }) {
  if (prefersReducedMotion()) {
    return <img src={STILL} alt="" aria-hidden="true" style={{ height }} className="w-auto select-none" />
  }

  return (
    <video
      src={VIDEO[surface]}
      poster={STILL}
      autoPlay
      muted
      loop
      playsInline
      preload="auto"
      aria-hidden="true"
      tabIndex={-1}
      style={{ height }}
      className="pointer-events-none w-auto select-none"
    />
  )
}

// Layar penuh saat sesi diperiksa atau halaman pertama kali dimuat.
export function BaniScreen({ label }: { label: string }) {
  return (
    <div role="status" className="flex min-h-screen flex-col items-center justify-center gap-5 bg-slate-100 px-4">
      <BuildingAnimation height={176} surface="page" />
      <p className="text-[13px] text-slate-500">{label}</p>
    </div>
  )
}

// Versi kecil di dalam kartu putih, muncul setelah 300 ms supaya data yang
// cepat datang tidak berkedip.
export function BaniLoop({ label }: { label: string }) {
  return (
    <div role="status" className="flex items-center gap-3 py-3 motion-safe:animate-[fade_200ms_ease-out_300ms_both]">
      <BuildingAnimation height={56} surface="card" />
      <span className="text-[13px] text-slate-500">{label}</span>
    </div>
  )
}
