const angka = [
  { nilai: '2,89 jt', label: 'zona harga tanah' },
  { nilai: '458 rb', label: 'objek lahan' },
  { nilai: '12', label: 'indeks bahaya' },
]

export function PanelMerek() {
  return (
    <aside className="hidden flex-col justify-end bg-navy-700 px-12 py-12 text-white lg:flex">
      <p className="max-w-md text-xl leading-relaxed font-medium">
        Developer memilih lokasi dari intuisi dan harga tanah murah. Datanya sebenarnya sudah terbuka.
      </p>
      <p className="mt-4 max-w-md text-sm leading-relaxed text-navy-100">
        BanguninAja menyatukan harga tanah resmi ATR/BPN, indeks bahaya BNPB, dan sebaran penduduk
        jadi satu skor yang bisa kamu telusuri asal usulnya.
      </p>

      <dl className="mt-10 flex gap-10 border-t border-white/15 pt-6">
        {angka.map((butir) => (
          <div key={butir.label}>
            <dt className="text-xs text-navy-100">{butir.label}</dt>
            <dd className="mt-1 text-xl font-semibold tabular-nums">{butir.nilai}</dd>
          </div>
        ))}
      </dl>
    </aside>
  )
}
