import { useState } from 'react'

import { useDeleteEmployee, useEmployees, useSetEmployeeLeftDate } from '../../controllers/useHr'
import {
  EMPLOYMENT_TYPES,
  EMPLOYMENT_TYPE_LABEL,
  inputDate,
  type Employee,
  type EmploymentType,
} from '../../models/hr'
import type { Project } from '../../models/project'
import { rupiah, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { CONTROL_CLASS } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FilterSelect, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { EditEmployeeForm, NewEmployeeForm } from './EmployeeForm'
import { PAGE_SIZE, pageAfterRemoval, refusalText } from './hrShared'
import { ActionGroup, ConfirmAction, RowNotice, TextAction } from './RowActions'

type ActiveFilter = 'true' | 'false' | ''
type EmployeeAsk = { id: string; kind: 'deactivate' | 'delete' } | null

function salaryText(employee: Employee): string {
  const amount = rupiah(employee.baseSalary)
  return employee.employmentType === 'daily' ? `${amount} /hari` : amount
}

// Tanggal keluar tetap dihitung hari kerja, jadi karyawan yang keluar hari ini
// atau nanti masih aktif sampai tanggal itu lewat.
function EmploymentChip({ employee }: { employee: Employee }) {
  if (employee.active && employee.leftDate !== null) {
    return <Chip tone="bg-amber-100 text-amber-800">Sampai {shortDate(employee.leftDate)}</Chip>
  }
  if (employee.active) {
    return <Chip tone="bg-green-100 text-green-800">Aktif</Chip>
  }
  if (employee.leftDate === null) {
    return <Chip tone="bg-blue-100 text-blue-800">Mulai {shortDate(employee.joinedDate)}</Chip>
  }
  return <Chip tone="bg-slate-100 text-slate-700">Keluar {shortDate(employee.leftDate)}</Chip>
}

function laterDate(first: string, second: string): string {
  return first > second ? first : second
}

export function EmployeesSection({ projects, projectName, today }: {
  projects: Project[]
  projectName: (id: string | null) => string
  today: string
}) {
  const [search, setSearch] = useState('')
  const [projectId, setProjectId] = useState('')
  const [employmentType, setEmploymentType] = useState<EmploymentType | ''>('')
  const [active, setActive] = useState<ActiveFilter>('true')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Employee | null>(null)
  const [asking, setAsking] = useState<EmployeeAsk>(null)
  const [leftDate, setLeftDate] = useState(today)
  const [notice, setNotice] = useState<string | null>(null)
  const setEmployeeLeftDate = useSetEmployeeLeftDate()
  const deleteEmployee = useDeleteEmployee()
  const isFormOpen = isCreating || editing !== null

  const employees = useEmployees({
    search: search.trim() === '' ? undefined : search.trim(),
    projectId: projectId === '' ? undefined : projectId,
    employmentType: employmentType === '' ? undefined : employmentType,
    active: active === '' ? undefined : active,
    page,
    pageSize: PAGE_SIZE,
  })
  const rowCount = employees.data?.items.length ?? 0

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  const clearNotes = () => {
    setNotice(null)
    setEmployeeLeftDate.reset()
    deleteEmployee.reset()
  }

  const closeForm = () => {
    setIsCreating(false)
    setEditing(null)
  }

  const toggleForm = () => {
    clearNotes()
    if (isFormOpen) {
      closeForm()
      return
    }
    setIsCreating(true)
  }

  const startEdit = (employee: Employee) => {
    clearNotes()
    setAsking(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === employee.id ? null : employee))
  }

  const ask = (employee: Employee, kind: 'deactivate' | 'delete') => {
    clearNotes()
    setLeftDate(laterDate(today, inputDate(employee.joinedDate)))
    setAsking({ id: employee.id, kind })
  }

  const isAsking = (employee: Employee, kind: 'deactivate' | 'delete') =>
    asking !== null && asking.id === employee.id && asking.kind === kind

  const deactivate = (employee: Employee) => {
    setEmployeeLeftDate.mutate(
      { employee, leftDate },
      {
        onSuccess: (saved) => {
          setNotice(`${saved.name} dinonaktifkan. Hari kerja terakhir ${shortDate(saved.leftDate)}.`)
          if (editing?.id === employee.id) {
            closeForm()
          }
        },
        onSettled: () => setAsking(null),
      },
    )
  }

  const reactivate = (employee: Employee) => {
    clearNotes()
    setAsking(null)
    setEmployeeLeftDate.mutate(
      { employee, leftDate: '' },
      { onSuccess: (saved) => setNotice(`${saved.name} kembali tercatat sebagai karyawan aktif.`) },
    )
  }

  const remove = (employee: Employee) => {
    deleteEmployee.mutate(employee, {
      onSuccess: () => {
        setNotice(`Karyawan ${employee.name} dihapus.`)
        setPage((current) => pageAfterRemoval(current, rowCount))
        if (editing?.id === employee.id) {
          closeForm()
        }
      },
      onSettled: () => setAsking(null),
    })
  }

  const rowError = setEmployeeLeftDate.isError
    ? refusalText(`Tanggal keluar ${setEmployeeLeftDate.variables.employee.name} tidak tersimpan.`, setEmployeeLeftDate.error)
    : deleteEmployee.isError
      ? refusalText(`${deleteEmployee.variables.name} tidak bisa dihapus.`, deleteEmployee.error)
      : null

  return (
    <Card title="Karyawan" description="Karyawan yang sudah punya absensi atau slip gaji dinonaktifkan, bukan dihapus.">
      <Toolbar>
        <ToolbarInput
          id="employee-search"
          label="Cari nama karyawan"
          type="search"
          placeholder="Cari nama karyawan"
          className="w-56"
          value={search}
          onChange={(event) => filterChanged(setSearch)(event.target.value)}
        />
        <FilterSelect
          id="employee-filter-project"
          label="Saring menurut proyek"
          value={projectId}
          onChange={(event) => filterChanged(setProjectId)(event.target.value)}
        >
          <option value="">Semua proyek</option>
          {projects.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name}
            </option>
          ))}
        </FilterSelect>
        <FilterSelect
          id="employee-filter-type"
          label="Saring menurut jenis kepegawaian"
          value={employmentType}
          onChange={(event) => filterChanged(setEmploymentType)(event.target.value as EmploymentType | '')}
        >
          <option value="">Semua jenis</option>
          {EMPLOYMENT_TYPES.map((type) => (
            <option key={type} value={type}>
              {EMPLOYMENT_TYPE_LABEL[type]}
            </option>
          ))}
        </FilterSelect>
        <FilterSelect
          id="employee-filter-active"
          label="Saring menurut status kerja"
          value={active}
          onChange={(event) => filterChanged(setActive)(event.target.value as ActiveFilter)}
        >
          <option value="true">Masih bekerja</option>
          <option value="false">Tidak aktif</option>
          <option value="">Semua status</option>
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Tambah karyawan" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewEmployeeForm projects={projects} /> : null}
      {editing ? (
        <EditEmployeeForm
          key={editing.id}
          employee={editing}
          projects={projects}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Data karyawan ${saved.name} berhasil diperbarui.`)
          }}
        />
      ) : null}

      <RowNotice success={notice} error={rowError} />

      {employees.isPending ? <Loading /> : null}
      {employees.isError ? <LoadFailed onRetry={() => employees.refetch()} /> : null}
      {employees.data ? (
        <>
          <Table
            rows={employees.data.items}
            emptyMessage="Tidak ada karyawan yang cocok dengan penyaringan ini."
            columns={[
              { header: 'Nomor induk', cell: (row) => row.identityNumber },
              {
                header: 'Nama',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="font-medium text-slate-900">{row.name}</span>
                    <span className="text-xs text-slate-500">{row.position}</span>
                  </span>
                ),
              },
              { header: 'Proyek', cell: (row) => projectName(row.projectId) },
              { header: 'Jenis', cell: (row) => EMPLOYMENT_TYPE_LABEL[row.employmentType] },
              { header: 'Bergabung', cell: (row) => shortDate(row.joinedDate) },
              { header: 'Gaji pokok', align: 'right', cell: salaryText },
              { header: 'Status', cell: (row) => <EmploymentChip employee={row} /> },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => {
                  const isDeactivating = isAsking(row, 'deactivate')
                  const isDeleting = isAsking(row, 'delete')
                  const isSavingLeftDate =
                    setEmployeeLeftDate.isPending && setEmployeeLeftDate.variables.employee.id === row.id
                  const canDeactivate = row.leftDate === null && !isDeleting
                  const canReactivate = row.leftDate !== null && !isDeleting

                  return (
                    <ActionGroup>
                      {isDeactivating || isDeleting ? null : (
                        <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                      )}
                      {canDeactivate ? (
                        <ConfirmAction
                          label="Nonaktifkan"
                          srLabel={row.name}
                          confirmLabel="Simpan tanggal keluar"
                          pendingLabel="Menyimpan"
                          tone="primary"
                          isAsking={isDeactivating}
                          isPending={isSavingLeftDate}
                          onAsk={() => ask(row, 'deactivate')}
                          onCancel={() => setAsking(null)}
                          onConfirm={() => deactivate(row)}
                        >
                          <label htmlFor={`employee-left-${row.id}`} className="text-xs text-slate-700">
                            Hari kerja terakhir
                          </label>
                          <input
                            id={`employee-left-${row.id}`}
                            type="date"
                            required
                            autoFocus
                            min={inputDate(row.joinedDate)}
                            value={leftDate}
                            onChange={(event) => setLeftDate(event.target.value)}
                            className={`${CONTROL_CLASS} h-7 px-2 text-xs`}
                          />
                        </ConfirmAction>
                      ) : null}
                      {canReactivate ? (
                        <TextAction
                          label="Aktifkan lagi"
                          srLabel={row.name}
                          isPending={isSavingLeftDate}
                          pendingLabel="Menyimpan"
                          onClick={() => reactivate(row)}
                        />
                      ) : null}
                      {isDeactivating ? null : (
                        <ConfirmAction
                          label="Hapus"
                          srLabel={row.name}
                          question={`Hapus ${row.name}?`}
                          confirmLabel="Ya, hapus"
                          pendingLabel="Menghapus"
                          isAsking={isAsking(row, 'delete')}
                          isPending={deleteEmployee.isPending && deleteEmployee.variables.id === row.id}
                          onAsk={() => ask(row, 'delete')}
                          onCancel={() => setAsking(null)}
                          onConfirm={() => remove(row)}
                        />
                      )}
                    </ActionGroup>
                  )
                },
              },
            ]}
          />
          <Pager
            page={employees.data.page}
            totalPages={employees.data.totalPages}
            totalItems={employees.data.totalItems}
            unit="karyawan"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
