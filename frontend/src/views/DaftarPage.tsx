import { type FormEvent, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { useDaftar } from '../controllers/useAuth'
import { Kolom, Peringatan, Tombol } from './komponen/Formulir'
import { PanelMerek } from './komponen/PanelMerek'
import { pesanKesalahan } from '../shared/pesanKesalahan'

const PANJANG_SANDI_MINIMAL = 8

export function DaftarPage() {
  const [nama, setNama] = useState('')
  const [email, setEmail] = useState('')
  const [sandi, setSandi] = useState('')
  const daftar = useDaftar()
  const navigate = useNavigate()

  const sandiTerlaluPendek = sandi.length > 0 && sandi.length < PANJANG_SANDI_MINIMAL

  const kirim = (peristiwa: FormEvent) => {
    peristiwa.preventDefault()
    daftar.mutate(
      { name: nama, email, password: sandi },
      {
        onSuccess: () =>
          navigate('/masuk', {
            replace: true,
            state: { pesan: 'Akun berhasil dibuat. Silakan masuk.' },
          }),
      },
    )
  }

  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      <main className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">Buat akun</h1>
          <p className="mt-2 text-sm text-slate-600">Cukup tiga isian, tidak sampai satu menit.</p>

          <form onSubmit={kirim} className="mt-8 flex flex-col gap-4" noValidate>
            {daftar.isError ? <Peringatan pesan={pesanKesalahan(daftar.error)} /> : null}

            <Kolom
              id="nama"
              label="Nama lengkap"
              autoComplete="name"
              placeholder="Jonathan Andrew"
              minLength={2}
              required
              value={nama}
              onChange={(peristiwa) => setNama(peristiwa.target.value)}
            />
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
              autoComplete="new-password"
              placeholder="Minimal 8 karakter"
              minLength={PANJANG_SANDI_MINIMAL}
              required
              bantuan={
                sandiTerlaluPendek
                  ? `Kurang ${PANJANG_SANDI_MINIMAL - sandi.length} karakter lagi`
                  : 'Minimal 8 karakter'
              }
              value={sandi}
              onChange={(peristiwa) => setSandi(peristiwa.target.value)}
            />

            <Tombol type="submit" sedangProses={daftar.isPending} labelProses="Membuat akun">
              Daftar
            </Tombol>
          </form>

          <p className="mt-6 text-sm text-slate-600">
            Sudah punya akun?{' '}
            <Link to="/masuk" className="font-medium text-navy-600 underline-offset-4 hover:underline">
              Masuk saja
            </Link>
          </p>
        </div>
      </main>

      <PanelMerek />
    </div>
  )
}
