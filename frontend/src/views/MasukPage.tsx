import { type FormEvent, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { useMasuk } from '../controllers/useAuth'
import { pesanKesalahan } from '../shared/pesanKesalahan'
import { Berhasil, Kolom, Peringatan, Tombol } from './komponen/Formulir'
import { PanelMerek } from './komponen/PanelMerek'

export function MasukPage() {
  const [email, setEmail] = useState('')
  const [sandi, setSandi] = useState('')
  const masuk = useMasuk()
  const navigate = useNavigate()
  const lokasi = useLocation()
  const pesanSukses = (lokasi.state as { pesan?: string } | null)?.pesan

  const kirim = (peristiwa: FormEvent) => {
    peristiwa.preventDefault()
    masuk.mutate(
      { email, password: sandi },
      { onSuccess: () => navigate('/', { replace: true }) },
    )
  }

  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      <main className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">Masuk ke BanguninAja</h1>
          <p className="mt-2 text-sm text-slate-600">
            Simpan lokasi incaran, riwayat analisis, dan rencana pembangunanmu.
          </p>

          <form onSubmit={kirim} className="mt-8 flex flex-col gap-4" noValidate>
            {pesanSukses ? <Berhasil pesan={pesanSukses} /> : null}
            {masuk.isError ? <Peringatan pesan={pesanKesalahan(masuk.error)} /> : null}

            <Kolom
              id="email"
              label="Email"
              type="email"
              autoComplete="email"
              placeholder="nama@kampus.ac.id"
              required
              value={email}
              onChange={(peristiwa) => setEmail(peristiwa.target.value)}
            />
            <Kolom
              id="sandi"
              label="Kata sandi"
              type="password"
              autoComplete="current-password"
              placeholder="Minimal 8 karakter"
              minLength={8}
              required
              value={sandi}
              onChange={(peristiwa) => setSandi(peristiwa.target.value)}
            />

            <Tombol type="submit" sedangProses={masuk.isPending} labelProses="Sedang masuk">
              Masuk
            </Tombol>
          </form>

          <p className="mt-6 text-sm text-slate-600">
            Belum punya akun?{' '}
            <Link to="/daftar" className="font-medium text-navy-600 underline-offset-4 hover:underline">
              Daftar dulu
            </Link>
          </p>
        </div>
      </main>

      <PanelMerek />
    </div>
  )
}
