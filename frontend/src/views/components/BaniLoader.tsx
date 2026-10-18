import { useId, useState } from 'react'

const BASE = '/images/bani'
const LOOPS = ['bricks', 'wheelbarrow', 'trowel', 'blueprint'] as const

type Loop = (typeof LOOPS)[number] | 'build-house'

// Animasi Bani, maskot BanguninAja. Gambar diam frame pertama (beberapa KB)
// tampil sebagai latar selama animasinya masih diunduh, dan menjadi satu-satunya
// gambar bagi pengguna yang memilih mengurangi gerak.
function BaniAnimation({ name, height }: { name: Loop; height: number }) {
  const [isLoaded, setIsLoaded] = useState(false)

  return (
    <picture>
      <source srcSet={`${BASE}/${name}-still.webp`} media="(prefers-reduced-motion: reduce)" />
      <img
        src={`${BASE}/${name}.webp`}
        alt=""
        aria-hidden="true"
        draggable={false}
        onLoad={() => setIsLoaded(true)}
        style={{ height, backgroundImage: isLoaded ? undefined : `url(${BASE}/${name}-still.webp)` }}
        className="w-auto bg-contain bg-center bg-no-repeat select-none"
      />
    </picture>
  )
}

// Layar penuh saat sesi diperiksa atau halaman pertama kali dimuat: Bani
// membangun rumah dari fondasi sampai atap.
export function BaniScreen({ label }: { label: string }) {
  return (
    <div role="status" className="flex min-h-screen flex-col items-center justify-center gap-4 bg-slate-100 px-4">
      <BaniAnimation name="build-house" height={176} />
      <p className="text-[13px] text-slate-500">{label}</p>
    </div>
  )
}

// Versi kecil di dalam kartu. Tiap kartu memakai salah satu dari empat
// animasi, dipilih tetap dari id komponen supaya tidak berganti saat render ulang.
export function BaniLoop({ label }: { label: string }) {
  const id = useId()
  const loop = LOOPS[[...id].reduce((sum, char) => sum + char.charCodeAt(0), 0) % LOOPS.length]

  return (
    <div role="status" className="flex items-center gap-3 py-3 motion-safe:animate-[fade_200ms_ease-out_300ms_both]">
      <BaniAnimation name={loop} height={64} />
      <span className="text-[13px] text-slate-500">{label}</span>
    </div>
  )
}
