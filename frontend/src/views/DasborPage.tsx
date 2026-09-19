import { useNavigate } from 'react-router-dom'

import { useKeluar, useProfil, useSesi } from '../controllers/useAuth'
import { Peringatan, Tombol } from './komponen/Formulir'

export function DasborPage() {
  const { pengguna } = useSesi()
  const profil = useProfil()
  const keluar = useKeluar()
  const navigate = useNavigate()

  const akhiriSesi = () => {
    keluar.mutate(undefined, { onSettled: () => navigate('/masuk', { replace: true }) })
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <span className="text-sm font-semibold tracking-tight">BanguninAja</span>
          <Tombol varian="halus" sedangProses={keluar.isPending} labelProses="Keluar" onClick={akhiriSesi}>
            Keluar
          </Tombol>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-10">
        <h1 className="text-2xl font-semibold tracking-tight">
          Halo, {pengguna?.name ?? 'pengguna'}
        </h1>
        <p className="mt-2 text-sm text-slate-600">
          Sesi kamu aktif. Modul proyek, pengadaan, dan peta lokasi menyusul setelah endpointnya siap.
        </p>

        <section className="mt-8 rounded-xl border border-slate-200 bg-white p-6">
          <h2 className="text-sm font-semibold text-slate-900">Data akun</h2>

          {profil.isPending ? (
            <p className="mt-3 text-sm text-slate-500">Mengambil data akun...</p>
          ) : null}

          {profil.isError ? (
            <div className="mt-3">
              <Peringatan pesan="Gagal mengambil data akun." />
              <div className="mt-3">
                <Tombol varian="halus" onClick={() => profil.refetch()}>
                  Coba lagi
                </Tombol>
              </div>
            </div>
          ) : null}

          {profil.data ? (
            <dl className="mt-4 grid gap-4 sm:grid-cols-3">
              <Butir label="Nama" nilai={profil.data.name} />
              <Butir label="Email" nilai={profil.data.email} />
              <Butir label="Bergabung" nilai={tanggalIndonesia(profil.data.createdAt)} />
            </dl>
          ) : null}
        </section>
      </main>
    </div>
  )
}

function Butir({ label, nilai }: { label: string; nilai: string }) {
  return (
    <div>
      <dt className="text-xs text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm font-medium text-slate-900">{nilai}</dd>
    </div>
  )
}

function tanggalIndonesia(iso: string): string {
  return new Date(iso).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}
