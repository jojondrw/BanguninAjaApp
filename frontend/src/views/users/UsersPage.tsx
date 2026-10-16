import { Hourglass } from 'lucide-react'
import { useState } from 'react'

import { useCurrentUser } from '../../controllers/useAuth'
import { useChangeUserActive, useChangeUserRole, useUsers } from '../../controllers/useUsers'
import {
  type ManagedUser,
  PENDING_ROLE_FILTER,
  ROLE_OPTIONS,
  type RoleCode,
  roleLabel,
  type UserFilter,
} from '../../models/user'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { AppShell } from '../components/AppShell'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { BUTTON_BASE, ErrorNote, SuccessNote } from '../components/Form'
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
  const pending = useUsers({ role: PENDING_ROLE_FILTER, pageSize: 1 })
  const pendingCount = pending.data?.totalItems ?? 0
  const isPendingView = role === PENDING_ROLE_FILTER
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

  const togglePendingView = () => {
    setRole(isPendingView ? '' : PENDING_ROLE_FILTER)
    setPage(1)
  }

  const failure = changeRole.isError ? changeRole.error : changeActive.isError ? changeActive.error : null

  return (
    <AppShell title="Pengguna" description="Atur peran dan status akun tim. Hak akses tiap modul mengikuti peran.">
      <Card
        title="Daftar pengguna"
        description="Pendaftar baru tidak bisa membuka data apa pun sampai Anda memberi peran. Harus selalu ada minimal satu admin aktif."
      >
        {pendingCount > 0 || isPendingView ? (
          <PendingBanner count={pendingCount} isShowing={isPendingView} onToggle={togglePendingView} />
        ) : null}
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
            <option value={PENDING_ROLE_FILTER}>Menunggu peran</option>
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
              emptyMessage={emptyMessage(filter, isPendingView)}
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
                  cell: (row) =>
                    row.id === me?.id ? null : (
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

function emptyMessage(filter: UserFilter, isPendingView: boolean): string {
  if (isPendingView && !filter.search && !filter.active) {
    return 'Tidak ada pendaftar yang menunggu peran.'
  }
  if (filter.search || filter.role || filter.active) {
    return 'Tidak ada pengguna yang cocok dengan saringan itu.'
  }
  return 'Belum ada pengguna terdaftar.'
}

// Pendaftar baru tidak punya peran, jadi admin perlu tahu ada yang menunggu.
// Tombolnya menyaring daftar ke akun tanpa peran, sama dengan pilihan
// "Menunggu peran" di saringan peran.
function PendingBanner({ count, isShowing, onToggle }: { count: number; isShowing: boolean; onToggle: () => void }) {
  return (
    <div className="mb-4 flex flex-wrap items-center gap-x-3 gap-y-2 rounded-xl bg-amber-50 px-3.5 py-2.5 text-[13px] text-amber-800">
      <Hourglass aria-hidden="true" className="size-4 shrink-0" strokeWidth={1.8} />
      <p role="status" className="min-w-0 flex-1">
        {count > 0
          ? `${number(count)} pendaftar menunggu peran. Pilih perannya di kolom Peran, atau nonaktifkan akun yang tidak dikenal.`
          : 'Tidak ada lagi pendaftar yang menunggu peran.'}
      </p>
      <button
        type="button"
        onClick={onToggle}
        aria-pressed={isShowing}
        className={`h-8 ${BUTTON_BASE} ${
          isShowing ? 'bg-slate-900 text-white hover:bg-slate-800' : 'bg-white text-amber-900 shadow-hairline hover:bg-amber-100/60'
        }`}
      >
        Menunggu peran
      </button>
    </div>
  )
}
