import { type FormEvent, useState } from 'react'

import { useChangePassword, useCurrentUser, useUpdateProfile } from '../../controllers/useAuth'
import {
  NAME_MAX_LENGTH,
  NAME_MIN_LENGTH,
  PASSWORD_MAX_LENGTH,
  PASSWORD_MIN_LENGTH,
  type User,
} from '../../models/auth'
import { errorMessage } from '../../shared/errorMessage'
import { AppShell } from '../components/AppShell'
import { Card } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'

export function AccountPage() {
  const user = useCurrentUser()

  return (
    <AppShell title="Akun saya" description="Nama tampilan dan kata sandi akun Anda">
      <div className="grid items-start gap-8 lg:grid-cols-2">
        {user ? <ProfileCard user={user} /> : null}
        <PasswordCard />
      </div>
    </AppShell>
  )
}

function ProfileCard({ user }: { user: User }) {
  const [name, setName] = useState(user.name)
  const [formError, setFormError] = useState<string | null>(null)
  const update = useUpdateProfile()
  const trimmed = name.trim()
  const shortBy = NAME_MIN_LENGTH - trimmed.length

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.reset()
    if (trimmed.length < NAME_MIN_LENGTH || trimmed.length > NAME_MAX_LENGTH) {
      setFormError(`Nama harus ${NAME_MIN_LENGTH} sampai ${NAME_MAX_LENGTH} karakter.`)
      return
    }
    if (trimmed === user.name) {
      setFormError('Nama masih sama dengan yang tersimpan.')
      return
    }
    setFormError(null)
    update.mutate({ name: trimmed }, { onSuccess: (saved) => setName(saved.name) })
  }

  return (
    <Card title="Profil" description="Nama ini tampil di menu samping, riwayat laporan, dan daftar pengguna.">
      <form onSubmit={submit} className="grid gap-4" noValidate>
        <Field
          id="account-name"
          label="Nama"
          autoComplete="name"
          required
          minLength={NAME_MIN_LENGTH}
          maxLength={NAME_MAX_LENGTH}
          value={name}
          onChange={(event) => {
            setName(event.target.value)
            setFormError(null)
            update.reset()
          }}
          hint={
            name.length > 0 && shortBy > 0
              ? `Kurang ${shortBy} karakter lagi`
              : `${NAME_MIN_LENGTH} sampai ${NAME_MAX_LENGTH} karakter`
          }
        />

        <div className="rounded-xl bg-slate-50 px-4 py-3 text-[13px]">
          <dl className="grid gap-3 sm:grid-cols-2">
            <div className="min-w-0">
              <dt className="text-slate-500">Email</dt>
              <dd className="mt-0.5 font-medium break-all text-slate-900">{user.email}</dd>
            </div>
            <div className="min-w-0">
              <dt className="text-slate-500">Peran</dt>
              <dd className="mt-0.5 font-medium text-slate-900">{user.role?.name ?? 'Menunggu peran'}</dd>
            </div>
          </dl>
          <p className="mt-3 text-xs text-slate-500">
            Email dipakai untuk masuk dan tidak bisa diubah di sini. Peran hanya bisa diganti admin di halaman Pengguna.
          </p>
        </div>

        {formError ? <ErrorNote message={formError} /> : null}
        {update.isError ? <ErrorNote message={errorMessage(update.error)} /> : null}
        {update.isSuccess ? <SuccessNote message="Nama tersimpan." /> : null}

        <div>
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan">
            Simpan nama
          </Button>
        </div>
      </form>
    </Card>
  )
}

const EMPTY_PASSWORDS = { current: '', next: '', confirm: '' }

function PasswordCard() {
  const [values, setValues] = useState(EMPTY_PASSWORDS)
  const [formError, setFormError] = useState<string | null>(null)
  const change = useChangePassword()
  const shortBy = PASSWORD_MIN_LENGTH - values.next.length
  const isMismatch = values.confirm.length > 0 && values.confirm !== values.next

  const update = (changes: Partial<typeof EMPTY_PASSWORDS>) => {
    setValues((current) => ({ ...current, ...changes }))
    setFormError(null)
    change.reset()
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    change.reset()
    const problem = passwordProblem(values)
    if (problem) {
      setFormError(problem)
      return
    }
    setFormError(null)
    change.mutate(
      { currentPassword: values.current, newPassword: values.next },
      { onSuccess: () => setValues(EMPTY_PASSWORDS) },
    )
  }

  return (
    <Card title="Ganti kata sandi" description="Setelah diganti, perangkat lain yang masih masuk harus masuk lagi.">
      <form onSubmit={submit} className="grid gap-4" noValidate>
        <Field
          id="account-current-password"
          label="Kata sandi saat ini"
          type="password"
          autoComplete="current-password"
          required
          maxLength={PASSWORD_MAX_LENGTH}
          value={values.current}
          onChange={(event) => update({ current: event.target.value })}
        />
        <Field
          id="account-new-password"
          label="Kata sandi baru"
          type="password"
          autoComplete="new-password"
          required
          minLength={PASSWORD_MIN_LENGTH}
          maxLength={PASSWORD_MAX_LENGTH}
          value={values.next}
          onChange={(event) => update({ next: event.target.value })}
          hint={
            values.next.length > 0 && shortBy > 0
              ? `Kurang ${shortBy} karakter lagi`
              : `Minimal ${PASSWORD_MIN_LENGTH} karakter, berbeda dari kata sandi saat ini`
          }
        />
        <Field
          id="account-confirm-password"
          label="Ulangi kata sandi baru"
          type="password"
          autoComplete="new-password"
          required
          maxLength={PASSWORD_MAX_LENGTH}
          value={values.confirm}
          onChange={(event) => update({ confirm: event.target.value })}
          hint={isMismatch ? 'Belum sama dengan kata sandi baru' : undefined}
        />

        {formError ? <ErrorNote message={formError} /> : null}
        {change.isError ? <ErrorNote message={errorMessage(change.error)} /> : null}
        {change.isSuccess ? (
          <SuccessNote message="Kata sandi diganti. Sesi ini tetap berjalan; perangkat lain perlu masuk lagi." />
        ) : null}

        <div>
          <Button type="submit" isPending={change.isPending} pendingLabel="Menyimpan">
            Ganti kata sandi
          </Button>
        </div>
      </form>
    </Card>
  )
}

function passwordProblem(values: typeof EMPTY_PASSWORDS): string | null {
  if (!values.current) {
    return 'Isi kata sandi saat ini.'
  }
  if (values.next.length < PASSWORD_MIN_LENGTH || values.next.length > PASSWORD_MAX_LENGTH) {
    return `Kata sandi baru harus ${PASSWORD_MIN_LENGTH} sampai ${PASSWORD_MAX_LENGTH} karakter.`
  }
  if (values.next === values.current) {
    return 'Kata sandi baru harus berbeda dari kata sandi saat ini.'
  }
  if (values.confirm !== values.next) {
    return 'Ulangan kata sandi baru belum sama.'
  }
  return null
}
