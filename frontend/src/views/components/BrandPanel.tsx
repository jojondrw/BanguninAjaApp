import { type ReactNode, useEffect } from 'react'

import { BrandName } from './Brand'

const figures = [
  { value: '2,89 jt', label: 'zona harga tanah' },
  { value: '458 rb', label: 'objek lahan' },
  { value: '12', label: 'indeks bahaya' },
]

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
  return (
    <aside className="relative m-3 hidden overflow-hidden rounded-[28px] bg-navy-900 text-white lg:flex lg:flex-col lg:justify-end">
      <div
        aria-hidden="true"
        className="absolute inset-0 bg-[radial-gradient(120%_80%_at_100%_0%,rgb(31_95_209/0.55),transparent_60%),radial-gradient(70%_60%_at_0%_100%,rgb(251_187_22/0.18),transparent_70%)]"
      />
      <div
        aria-hidden="true"
        className="absolute inset-0 opacity-[0.07] [background-image:linear-gradient(white_1px,transparent_1px),linear-gradient(90deg,white_1px,transparent_1px)] [background-size:44px_44px] [mask-image:linear-gradient(to_bottom,black,transparent_75%)]"
      />

      <div className="relative px-12 pb-12">
        <p className="max-w-lg text-[26px] leading-[1.25] font-semibold tracking-[-0.022em]">
          Developer memilih lokasi dari intuisi dan harga tanah murah. Datanya sebenarnya sudah terbuka.
        </p>
        <p className="mt-4 max-w-md text-[15px] leading-relaxed text-white/70">
          BanguninAja menyatukan harga tanah resmi ATR/BPN, indeks bahaya BNPB, dan sebaran penduduk jadi satu skor
          yang bisa kamu telusuri asal usulnya.
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
