import { type FormEvent, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { useRegister } from '../controllers/useAuth'
import { errorMessage } from '../shared/errorMessage'
import { Button, ErrorNote, Field } from './components/Form'
import { BrandPanel } from './components/BrandPanel'

const MINIMUM_PASSWORD_LENGTH = 8

export function RegisterPage() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const register = useRegister()
  const navigate = useNavigate()

  const passwordTooShort = password.length > 0 && password.length < MINIMUM_PASSWORD_LENGTH

  const submit = (event: FormEvent) => {
    event.preventDefault()
    register.mutate(
      { name, email, password },
      {
        onSuccess: () =>
          navigate('/masuk', {
            replace: true,
            state: { message: 'Akun berhasil dibuat. Silakan masuk.' },
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

          <form onSubmit={submit} className="mt-8 flex flex-col gap-4" noValidate>
            {register.isError ? <ErrorNote message={errorMessage(register.error)} /> : null}

            <Field
              id="name"
              label="Nama lengkap"
              autoComplete="name"
              placeholder="Jonathan Andrew"
              minLength={2}
              required
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
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
              autoComplete="new-password"
              placeholder="Minimal 8 karakter"
              minLength={MINIMUM_PASSWORD_LENGTH}
              required
              hint={
                passwordTooShort
                  ? `Kurang ${MINIMUM_PASSWORD_LENGTH - password.length} karakter lagi`
                  : 'Minimal 8 karakter'
              }
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />

            <Button type="submit" isPending={register.isPending} pendingLabel="Membuat akun">
              Daftar
            </Button>
          </form>

          <p className="mt-6 text-sm text-slate-600">
            Sudah punya akun?{' '}
            <Link to="/masuk" className="font-medium text-navy-600 underline-offset-4 hover:underline">
              Masuk saja
            </Link>
          </p>
        </div>
      </main>

      <BrandPanel />
    </div>
  )
}
