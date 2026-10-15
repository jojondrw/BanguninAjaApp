import { useState } from 'react'

import { useCurrentUser } from '../../controllers/useAuth'
import { useChangeUserActive, useChangeUserRole, useUsers } from '../../controllers/useUsers'
import { type ManagedUser, ROLE_OPTIONS, type RoleCode, roleLabel, type UserFilter } from '../../models/user'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { AppShell } from '../components/AppShell'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { ErrorNote, SuccessNote } from '../components/Form'
import { Chip, FilterSelect, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { ActionGroup, ConfirmAction } from '../hr/RowActions'
import { RoleCell } from './RoleCell'

const PAGE_SIZE = 20

export function UsersPage() {
  const me = useCurrentUser()
  const [search, setSearch] = useState('')
  const [role, setRole] = useState('')
  const [status, setStatus] = useState<'' | 'true' | 'false'>('')
  const [page, setPage] = useState(1)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const filter: UserFilter = {
    search: search.trim() || undefined,
    role: role || undefined,
    active: status || undefined,
    page,
    pageSize: PAGE_SIZE,
  }
  const users = useUsers(filter)
  const changeRole = useChangeUserRole()
  const changeActive = useChangeUserActive()

  const resetMessages = () => {
    setNotice(null)
    changeRole.reset()
    changeActive.reset()
  }

  const saveRole = (user: ManagedUser, next: RoleCode) => {
    resetMessages()
    changeRole.mutate(
      { user, role: next },
      { onSuccess: (saved) => setNotice(`Peran ${saved.name} sekarang ${roleLabel(saved.role?.code)}.`) },
    )
  }

  const askToggle = (user: ManagedUser) => {
    resetMessages()
    setAskingId(user.id)
  }

  const confirmToggle = (user: ManagedUser) => {
    changeActive.mutate(
      { user, active: !user.active },
      {
        onSuccess: (saved) =>
          setNotice(saved.active ? `Akun ${saved.name} aktif lagi.` : `Akun ${saved.name} dinonaktifkan dan sesinya dicabut.`),
        onSettled: () => setAskingId(null),
      },
    )
  }

  const failure = changeRole.isError ? changeRole.error : changeActive.isError ? changeActive.error : null

  return (
    <AppShell title="Pengguna" description="Atur peran dan status akun tim. Hak akses tiap modul mengikuti peran.">
      <Card
        title="Daftar pengguna"
        description="Pendaftar baru otomatis menjadi Pengamat. Harus selalu ada minimal satu admin aktif."
      >
        <Toolbar>
          <ToolbarInput
            id="user-search"
            label="Cari pengguna"
            type="search"
            placeholder="Cari nama atau email"
            maxLength={160}
            className="w-60"
            value={search}
            onChange={(event) => {
              setSearch(event.target.value)
              setPage(1)
            }}
          />
          <FilterSelect
            id="user-role-filter"
            label="Saring peran"
            value={role}
            onChange={(event) => {
              setRole(event.target.value)
              setPage(1)
            }}
          >
            <option value="">Semua peran</option>
            {ROLE_OPTIONS.map((option) => (
              <option key={option.code} value={option.code}>
                {option.label}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect
            id="user-status-filter"
            label="Saring status"
            value={status}
            onChange={(event) => {
              setStatus(event.target.value as '' | 'true' | 'false')
              setPage(1)
            }}
          >
            <option value="">Semua status</option>
            <option value="true">Aktif</option>
            <option value="false">Nonaktif</option>
          </FilterSelect>
        </Toolbar>

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}
        {failure ? (
          <div className="mb-4">
            <ErrorNote message={errorMessage(failure)} />
          </div>
        ) : null}

        {users.isPending ? <Loading /> : null}
        {users.isError ? <LoadFailed onRetry={() => users.refetch()} /> : null}
        {users.data ? (
          <>
            <Table
              rows={users.data.items}
              emptyMessage={
                filter.search || filter.role || filter.active
                  ? 'Tidak ada pengguna yang cocok dengan saringan itu.'
                  : 'Belum ada pengguna terdaftar.'
              }
              columns={[
                {
                  header: 'Pengguna',
                  cell: (row) => (
                    <div className="min-w-0">
                      <p className="font-medium text-slate-900">
                        {row.name}
                        {row.id === me?.id ? <span className="ml-1.5 text-xs font-normal text-slate-500">(Anda)</span> : null}
                      </p>
                      <p className="text-xs text-slate-500">{row.email}</p>
                    </div>
                  ),
                },
                {
                  header: 'Peran',
                  cell: (row) => (
                    <RoleCell
                      key={`${row.id}-${row.role?.code ?? 'none'}`}
                      user={row}
                      isPending={changeRole.isPending && changeRole.variables.user.id === row.id}
                      onSave={saveRole}
                    />
                  ),
                },
                {
                  header: 'Status',
                  cell: (row) =>
                    row.active ? (
                      <Chip tone="bg-green-100 text-green-800">Aktif</Chip>
                    ) : (
                      <Chip tone="bg-slate-100 text-slate-700">Nonaktif</Chip>
                    ),
                },
                { header: 'Terdaftar', cell: (row) => shortDate(row.createdAt) },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <ActionGroup>
                      <ConfirmAction
                        label={row.active ? 'Nonaktifkan' : 'Aktifkan lagi'}
                        srLabel={row.active ? `Nonaktifkan ${row.name}` : `Aktifkan lagi ${row.name}`}
                        question={
                          row.active
                            ? `Nonaktifkan ${row.name}? Sesinya dicabut dan ia tidak bisa masuk lagi.`
                            : `Aktifkan lagi ${row.name}?`
                        }
                        confirmLabel={row.active ? 'Ya, nonaktifkan' : 'Ya, aktifkan'}
                        pendingLabel={row.active ? 'Menonaktifkan' : 'Mengaktifkan'}
                        tone={row.active ? 'danger' : 'primary'}
                        isAsking={askingId === row.id}
                        isPending={changeActive.isPending && changeActive.variables.user.id === row.id}
                        onAsk={() => askToggle(row)}
                        onCancel={() => setAskingId(null)}
                        onConfirm={() => confirmToggle(row)}
                      />
                    </ActionGroup>
                  ),
                },
              ]}
            />
            <Pager
              page={users.data.page}
              totalPages={users.data.totalPages}
              totalItems={users.data.totalItems}
              unit="pengguna"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>
    </AppShell>
  )
}
