import { type FormEvent, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { useLogin } from '../controllers/useAuth'
import { errorMessage } from '../shared/errorMessage'
import { Button, ErrorNote, Field, SuccessNote } from './components/Form'
import { BrandPanel } from './components/BrandPanel'

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const login = useLogin()
  const navigate = useNavigate()
  const location = useLocation()
  const successMessage = (location.state as { message?: string } | null)?.message

  const submit = (event: FormEvent) => {
    event.preventDefault()
    login.mutate({ email, password }, { onSuccess: () => navigate('/', { replace: true }) })
  }

  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      <main className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">Masuk ke BanguninAja</h1>
          <p className="mt-2 text-sm text-slate-600">
            Simpan lokasi incaran, riwayat analisis, dan rencana pembangunanmu.
          </p>

          <form onSubmit={submit} className="mt-8 flex flex-col gap-4" noValidate>
            {successMessage ? <SuccessNote message={successMessage} /> : null}
            {login.isError ? <ErrorNote message={errorMessage(login.error)} /> : null}

            <Field
              id="email"
              label="Email"
              type="email"
              autoComplete="email"
              placeholder="nama@kampus.ac.id"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />
            <Field
              id="password"
              label="Kata sandi"
              type="password"
              autoComplete="current-password"
              placeholder="Minimal 8 karakter"
              minLength={8}
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />

            <Button type="submit" isPending={login.isPending} pendingLabel="Sedang masuk">
              Masuk
            </Button>
          </form>

          <p className="mt-6 text-sm text-slate-600">
            Belum punya akun?{' '}
            <Link to="/daftar" className="font-medium text-navy-600 underline-offset-4 hover:underline">
              Daftar dulu
            </Link>
          </p>
        </div>
      </main>

      <BrandPanel />
    </div>
  )
}
