import { type ReactNode, useEffect, useState } from 'react'

import { PHOTOS } from '../../shared/photos'
import { BrandName } from './Brand'

const figures = [
  { value: '2,89 jt', label: 'Zona harga tanah' },
  { value: '458 rb', label: 'Objek lahan' },
  { value: '12', label: 'Indeks bahaya' },
]

// Urutannya mengikuti alur produk: memilih lokasi, membangun, lalu menjual unit.
const SLIDES = [PHOTOS.lokasi, PHOTOS.konstruksi, PHOTOS.perumahan]
const SLIDE_MS = 7000

function useSlideIndex(count: number): number {
  const [index, setIndex] = useState(0)

  useEffect(() => {
    // Tanpa matchMedia (misalnya jsdom) foto tidak berganti, sama seperti saat
    // pengguna meminta gerakan dikurangi.
    const reduceMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? true
    if (reduceMotion) {
      return
    }
    const timer = window.setInterval(() => setIndex((current) => (current + 1) % count), SLIDE_MS)
    return () => window.clearInterval(timer)
  }, [count])

  return index
}

export function AuthLayout({ pageTitle, children }: { pageTitle: string; children: ReactNode }) {
  useEffect(() => {
    document.title = `${pageTitle} · BanguninAja`
  }, [pageTitle])

  return (
    <div className="grid min-h-screen bg-white lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)]">
      <main className="flex flex-col px-6 py-6 sm:px-10">
        <BrandName />
        <div className="flex flex-1 items-center justify-center py-12">
          <div className="w-full max-w-[360px] motion-safe:animate-enter">{children}</div>
        </div>
      </main>

      <BrandPanel />
    </div>
  )
}

export function BrandPanel() {
  const index = useSlideIndex(SLIDES.length)

  return (
    <aside className="relative m-3 hidden overflow-hidden rounded-[28px] bg-navy-900 text-white lg:flex lg:flex-col lg:justify-end">
      {SLIDES.map((photo, slide) => (
        <img
          key={photo.src}
          src={photo.src}
          alt=""
          aria-hidden="true"
          decoding="async"
          fetchPriority={slide === 0 ? 'high' : 'low'}
          className={`absolute inset-0 size-full object-cover transition-[opacity,scale] ease-out
                      [transition-duration:1400ms,8000ms] ${slide === index ? 'scale-100 opacity-100' : 'scale-105 opacity-0'}`}
        />
      ))}
      <div
        aria-hidden="true"
        className="absolute inset-0 bg-[linear-gradient(to_top,rgb(7_28_70/0.96)_0%,rgb(7_28_70/0.82)_38%,rgb(7_28_70/0.25)_75%,rgb(7_28_70/0.1)_100%)]"
      />

      <div className="relative px-12 pb-12">
        <p className="max-w-lg text-[26px] leading-[1.25] font-semibold tracking-[-0.022em]">
          Developer memilih lokasi dari intuisi dan harga tanah murah. Datanya sebenarnya sudah terbuka.
        </p>
        <p className="mt-4 max-w-md text-[15px] leading-relaxed text-white/70">
          BanguninAja menyatukan harga tanah resmi ATR/BPN, indeks bahaya BNPB, dan sebaran penduduk jadi satu skor
          yang bisa Anda telusuri asal usulnya.
        </p>

        <dl className="mt-10 grid max-w-lg grid-cols-3 gap-px overflow-hidden rounded-2xl bg-white/10 ring-1 ring-white/10 backdrop-blur-md">
          {figures.map((figure) => (
            <div key={figure.label} className="bg-white/[0.04] px-5 py-4">
              <dt className="text-xs text-white/60">{figure.label}</dt>
              <dd className="mt-1 text-xl font-semibold tracking-[-0.02em] tabular-nums">{figure.value}</dd>
            </div>
          ))}
        </dl>
      </div>
    </aside>
  )
}
