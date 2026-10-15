import { Hourglass } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useCurrentUser, useLogout, useProfile } from '../../controllers/useAuth'
import { errorMessage } from '../../shared/errorMessage'
import { AuthLayout } from '../components/BrandPanel'
import { Button, ErrorNote } from '../components/Form'

// Pendaftar baru belum punya peran, jadi semua modul tertutup. Daripada
// menampilkan aplikasi kosong penuh penolakan, layar ini menjelaskan apa yang
// harus terjadi dan memberi cara memeriksa ulang setelah admin memberi peran.
export function PendingApprovalPage() {
  const user = useCurrentUser()
  const profile = useProfile()
  const logout = useLogout()
  const navigate = useNavigate()
  const [hasChecked, setHasChecked] = useState(false)

  const checkAgain = () => {
    setHasChecked(false)
    void profile.refetch().then(() => setHasChecked(true))
  }

  const signOut = () => {
    logout.mutate(undefined, { onSettled: () => navigate('/masuk', { replace: true }) })
  }

  return (
    <AuthLayout pageTitle="Menunggu persetujuan">
      <span className="grid size-11 place-items-center rounded-full bg-amber-50 text-amber-700">
        <Hourglass aria-hidden="true" className="size-5" strokeWidth={1.8} />
      </span>
      <h1 className="mt-5 text-[28px] leading-9 font-semibold tracking-[-0.025em] text-slate-900">
        Akun menunggu persetujuan
      </h1>
      <p className="mt-2 text-[15px] text-slate-500">
        Akun Anda sudah terdaftar, tetapi belum diberi peran. Supaya data perusahaan tidak terbuka untuk sembarang
        pendaftar, admin perlu memilih peran Anda dulu di halaman Pengguna. Setelah itu menu sesuai peran langsung
        muncul.
      </p>

      <dl className="mt-6 rounded-xl bg-slate-50 px-4 py-3 text-[13px]">
        <dt className="text-slate-500">Masuk sebagai</dt>
        <dd className="mt-0.5 font-medium break-all text-slate-900">{user?.email ?? '-'}</dd>
        {user?.name ? <dd className="text-slate-600">{user.name}</dd> : null}
      </dl>

      <div className="mt-6 flex flex-col gap-3">
        {profile.isError ? <ErrorNote message={errorMessage(profile.error)} /> : null}
        {hasChecked && !profile.isFetching && !profile.isError && !profile.data?.role ? (
          <p role="status" className="rounded-xl bg-amber-50 px-3.5 py-2.5 text-[13px] text-amber-800">
            Belum ada peran untuk akun ini. Hubungi admin tim Anda, lalu periksa lagi.
          </p>
        ) : null}
        <div className="flex flex-wrap gap-2">
          <Button onClick={checkAgain} isPending={profile.isFetching} pendingLabel="Memeriksa">
            Periksa lagi
          </Button>
          <Button variant="subtle" onClick={signOut} isPending={logout.isPending} pendingLabel="Sedang keluar">
            Keluar
          </Button>
        </div>
      </div>
    </AuthLayout>
  )
}
