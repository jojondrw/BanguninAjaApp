import { useNavigate } from 'react-router-dom'

import { useLogout, useProfile, useSession } from '../controllers/useAuth'
import { Button, ErrorNote } from './components/Form'

export function DashboardPage() {
  const { user } = useSession()
  const profile = useProfile()
  const logout = useLogout()
  const navigate = useNavigate()

  const endSession = () => {
    logout.mutate(undefined, { onSettled: () => navigate('/masuk', { replace: true }) })
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <span className="text-sm font-semibold tracking-tight">BanguninAja</span>
          <Button variant="subtle" isPending={logout.isPending} pendingLabel="Keluar" onClick={endSession}>
            Keluar
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-10">
        <h1 className="text-2xl font-semibold tracking-tight">Halo, {user?.name ?? 'pengguna'}</h1>
        <p className="mt-2 text-sm text-slate-600">
          Sesi kamu aktif. Modul proyek, pengadaan, dan peta lokasi menyusul setelah endpointnya siap.
        </p>

        <section className="mt-8 rounded-xl border border-slate-200 bg-white p-6">
          <h2 className="text-sm font-semibold text-slate-900">Data akun</h2>

          {profile.isPending ? (
            <p className="mt-3 text-sm text-slate-500">Mengambil data akun...</p>
          ) : null}

          {profile.isError ? (
            <div className="mt-3">
              <ErrorNote message="Gagal mengambil data akun." />
              <div className="mt-3">
                <Button variant="subtle" onClick={() => profile.refetch()}>
                  Coba lagi
                </Button>
              </div>
            </div>
          ) : null}

          {profile.data ? (
            <dl className="mt-4 grid gap-4 sm:grid-cols-3">
              <Item label="Nama" value={profile.data.name} />
              <Item label="Email" value={profile.data.email} />
              <Item label="Bergabung" value={formatDate(profile.data.createdAt)} />
            </dl>
          ) : null}
        </section>
      </main>
    </div>
  )
}

function Item({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm font-medium text-slate-900">{value}</dd>
    </div>
  )
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}
