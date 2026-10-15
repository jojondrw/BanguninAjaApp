import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useCurrentUser } from '../../controllers/useAuth'
import {
  useAddProjectMember,
  useProjectMembers,
  useRemoveProjectMember,
  useUpdateProjectMember,
} from '../../controllers/useProjectMembers'
import type { Project } from '../../models/project'
import {
  EMPTY_MEMBER_FORM,
  MEMBER_POSITION_MAX_LENGTH,
  MEMBER_POSITION_SUGGESTIONS,
  memberEditValues,
  memberSummary,
  type MemberEditValues,
  type MemberFormValues,
  type ProjectMember,
} from '../../models/projectMember'
import { memberCandidateOptions } from '../../models/projectMemberApi'
import { ROLE_OPTIONS, type RoleCode } from '../../models/user'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table, type Column } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FormPanel, FormToggle } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { Chip, FormSection, InlineConfirm } from './parts'

const POSITION_OPTIONS_ID = 'member-position-options'
const INACTIVE_TONE = 'bg-slate-100 text-slate-600'
const ROLE_HINT = 'Peran di proyek ini. Untuk sementara hak akses tetap mengikuti peran akun.'

type UpdateMember = ReturnType<typeof useUpdateProjectMember>

// Tim proyek: siapa saja yang terlibat, posisinya, dan peran aplikasinya di
// proyek ini. Tombol ubah mengikuti hak tulis modul proyek (AppShell memasang
// cakupan akses dari rute /proyek).
export function ProjectTeamTab({ project }: { project: Project }) {
  const members = useProjectMembers(project.id)
  const updateMember = useUpdateProjectMember(project.id)
  const removeMember = useRemoveProjectMember(project.id)
  const canWrite = useScopedCan('write')
  const currentUserId = useCurrentUser()?.id
  const [isAdding, setIsAdding] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [removedName, setRemovedName] = useState<string | null>(null)

  const editing = editingId ? members.data?.find((member) => member.id === editingId) : undefined

  const toggleEdit = (memberId: string) => {
    updateMember.reset()
    setRemovedName(null)
    setEditingId((current) => (current === memberId ? null : memberId))
  }

  const remove = (member: ProjectMember) => {
    updateMember.reset()
    setRemovedName(null)
    removeMember.mutate(member.id, {
      onSuccess: () => {
        setRemovedName(member.name)
        setEditingId((current) => (current === member.id ? null : current))
      },
    })
  }

  const columns: Column<ProjectMember>[] = [
    { header: 'Nama', cell: (row) => <MemberName member={row} isCurrentUser={row.userId === currentUserId} /> },
    { header: 'Email', cell: (row) => row.email },
    { header: 'Posisi', cell: (row) => row.position ?? '-' },
    { header: 'Peran proyek', cell: (row) => row.role.name },
    { header: 'Sejak', cell: (row) => shortDate(row.joinedAt) },
  ]
  if (canWrite) {
    columns.push({
      header: 'Aksi',
      align: 'right',
      cell: (row) => (
        <span className="inline-flex items-center justify-end gap-1">
          <RowAction label="Ubah" isActive={editingId === row.id} onClick={() => toggleEdit(row.id)} />
          <InlineConfirm
            label="Keluarkan"
            confirmLabel="Ya, keluarkan dari tim"
            pendingLabel="Mengeluarkan"
            isPending={removeMember.isPending && removeMember.variables === row.id}
            onConfirm={() => remove(row)}
          />
        </span>
      ),
    })
  }

  return (
    <Card
      title="Tim proyek"
      description={members.data ? teamDescription(members.data.length) : 'Orang yang terlibat di proyek ini'}
      action={<FormToggle isOpen={isAdding} openLabel="Tambah anggota" onToggle={() => setIsAdding((open) => !open)} />}
    >
      {isAdding && canWrite ? (
        <FormPanel>
          <AddMemberForm projectId={project.id} memberUserIds={(members.data ?? []).map((member) => member.userId)} />
        </FormPanel>
      ) : null}

      {removedName ? (
        <div className="mb-4">
          <SuccessNote message={`${removedName} dikeluarkan dari tim proyek.`} />
        </div>
      ) : null}
      {removeMember.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(removeMember.error)} />
        </div>
      ) : null}

      {members.isPending ? <Loading label="Mengambil anggota tim..." /> : null}
      {members.isError ? <LoadFailed onRetry={() => members.refetch()} /> : null}
      {members.data ? (
        <Table
          rows={members.data}
          emptyMessage={
            canWrite
              ? 'Belum ada anggota tim. Tambahkan lewat tombol Tambah anggota.'
              : 'Belum ada anggota tim untuk proyek ini.'
          }
          columns={columns}
        />
      ) : null}

      {updateMember.isSuccess && editingId === null ? (
        <div className="mt-4">
          <SuccessNote message={`Perubahan untuk ${memberSummary(updateMember.data)} disimpan.`} />
        </div>
      ) : null}
      {editing ? (
        <EditMemberForm key={editing.id} member={editing} update={updateMember} onDone={() => setEditingId(null)} />
      ) : null}

      <datalist id={POSITION_OPTIONS_ID}>
        {MEMBER_POSITION_SUGGESTIONS.map((position) => (
          <option key={position} value={position} />
        ))}
      </datalist>
    </Card>
  )
}

function teamDescription(count: number): string {
  return count === 0 ? 'Belum ada orang di tim proyek ini' : `${count} anggota, urut nama`
}

function MemberName({ member, isCurrentUser }: { member: ProjectMember; isCurrentUser: boolean }) {
  return (
    <span className="inline-flex flex-wrap items-center gap-x-2 gap-y-1">
      <span className="font-medium text-slate-900">{member.name}</span>
      {isCurrentUser ? <span className="text-xs text-slate-500">(Anda)</span> : null}
      {member.active ? null : <Chip tone={INACTIVE_TONE}>Akun nonaktif</Chip>}
    </span>
  )
}

function AddMemberForm({ projectId, memberUserIds }: { projectId: string; memberUserIds: string[] }) {
  const [values, setValues] = useState<MemberFormValues>(EMPTY_MEMBER_FORM)
  const [candidates] = useState(() => memberCandidateOptions(projectId))
  const addMember = useAddProjectMember(projectId)

  const update = (key: 'position' | 'role') => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    addMember.mutate(values, { onSuccess: () => setValues(EMPTY_MEMBER_FORM) })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-3">
        <SearchSelect
          {...candidates}
          id="member-user"
          label="Pengguna"
          placeholder="Cari nama atau email"
          required
          exclude={memberUserIds}
          hint="Hanya akun aktif yang belum masuk tim ini"
          value={values.userId}
          onChange={(userId) => setValues((current) => ({ ...current, userId }))}
        />
        <Field
          id="member-position"
          label="Posisi (opsional)"
          placeholder="Pilih atau ketik posisi"
          list={POSITION_OPTIONS_ID}
          autoComplete="off"
          maxLength={MEMBER_POSITION_MAX_LENGTH}
          hint={`Misalnya Site manager atau Estimator, maksimal ${MEMBER_POSITION_MAX_LENGTH} karakter`}
          value={values.position}
          onChange={update('position')}
        />
        <SelectField id="member-role" label="Peran proyek" hint={ROLE_HINT} value={values.role} onChange={update('role')}>
          <option value="">Ikuti peran akun</option>
          <RoleOptions />
        </SelectField>
      </div>

      <Button type="submit" isPending={addMember.isPending} pendingLabel="Menambahkan anggota">
        Tambahkan ke tim
      </Button>

      {addMember.isError ? <ErrorNote message={errorMessage(addMember.error)} /> : null}
      {addMember.isSuccess ? (
        <SuccessNote
          message={`${memberSummary(addMember.data)} masuk tim dengan peran ${addMember.data.role.name}.`}
        />
      ) : null}
    </form>
  )
}

function EditMemberForm({ member, update, onDone }: { member: ProjectMember; update: UpdateMember; onDone: () => void }) {
  const [values, setValues] = useState<MemberEditValues>(() => memberEditValues(member))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.mutate({ memberId: member.id, values }, { onSuccess: onDone })
  }

  return (
    <FormSection title={`Ubah posisi dan peran ${member.name}`}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field
            id="member-edit-position"
            label="Posisi (opsional)"
            placeholder="Pilih atau ketik posisi"
            list={POSITION_OPTIONS_ID}
            autoComplete="off"
            autoFocus
            maxLength={MEMBER_POSITION_MAX_LENGTH}
            hint="Kosongkan untuk menghapus posisi"
            value={values.position}
            onChange={(event) => setValues((current) => ({ ...current, position: event.target.value }))}
          />
          <SelectField
            id="member-edit-role"
            label="Peran proyek"
            hint={ROLE_HINT}
            value={values.role}
            onChange={(event) => setValues((current) => ({ ...current, role: event.target.value as RoleCode }))}
          >
            <RoleOptions />
          </SelectField>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan anggota
          </Button>
          <Button variant="subtle" onClick={onDone}>
            Batal
          </Button>
        </div>

        {update.isError ? <ErrorNote message={errorMessage(update.error)} /> : null}
      </form>
    </FormSection>
  )
}

function RoleOptions() {
  return (
    <>
      {ROLE_OPTIONS.map((role) => (
        <option key={role.code} value={role.code}>
          {role.label}
        </option>
      ))}
    </>
  )
}
