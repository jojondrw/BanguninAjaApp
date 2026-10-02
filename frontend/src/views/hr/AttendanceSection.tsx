import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useActiveEmployeeOptions,
  useAttendances,
  useCreateAttendance,
  useDeleteAttendance,
  useUpdateAttendance,
} from '../../controllers/useHr'
import {
  ATTENDANCE_STATUSES,
  ATTENDANCE_STATUS_LABEL,
  ATTENDANCE_STATUS_TONE,
  EMPTY_ATTENDANCE_FORM,
  attendanceFormValues,
  type Attendance,
  type AttendanceFormValues,
  type AttendanceStatus,
} from '../../models/hr'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import {
  Chip,
  FilterSelect,
  FormPanel,
  FormToggle,
  Pager,
  SelectField,
  Toolbar,
  ToolbarInput,
} from '../components/RecordControls'
import { EmployeeOptions } from './EmployeeOptions'
import { PAGE_SIZE, pageAfterRemoval, refusalText } from './hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from './RowActions'

type ChangeAttendance = (
  key: keyof AttendanceFormValues,
) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

function useAttendanceValues(initial: AttendanceFormValues) {
  const [values, setValues] = useState<AttendanceFormValues>(initial)

  const change: ChangeAttendance = (key) => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

// Jam masuk dan pulang hanya untuk yang hadir. Backend menolak jam untuk
// status lain, jadi isiannya dikunci dan tidak dikirim.
function AttendanceTimeFields({ idPrefix, values, onChange }: {
  idPrefix: string
  values: AttendanceFormValues
  onChange: ChangeAttendance
}) {
  const isPresent = values.status === 'present'

  return (
    <>
      <SelectField id={`${idPrefix}-status`} label="Status" value={values.status} onChange={onChange('status')}>
        {ATTENDANCE_STATUSES.map((status) => (
          <option key={status} value={status}>
            {ATTENDANCE_STATUS_LABEL[status]}
          </option>
        ))}
      </SelectField>
      <Field
        id={`${idPrefix}-check-in`}
        label="Jam masuk"
        type="time"
        required={isPresent}
        disabled={!isPresent}
        hint={isPresent ? 'Wajib untuk yang hadir' : 'Hanya untuk yang hadir'}
        value={isPresent ? values.checkInTime : ''}
        onChange={onChange('checkInTime')}
      />
      <Field
        id={`${idPrefix}-check-out`}
        label="Jam pulang (opsional)"
        type="time"
        disabled={!isPresent}
        min={values.checkInTime || undefined}
        hint={isPresent ? 'Tidak boleh lebih awal dari jam masuk' : undefined}
        value={isPresent ? values.checkOutTime : ''}
        onChange={onChange('checkOutTime')}
      />
    </>
  )
}

function NewAttendanceForm({ date }: { date: string }) {
  const { values, setValues, change } = useAttendanceValues(EMPTY_ATTENDANCE_FORM)
  const employees = useActiveEmployeeOptions()
  const createAttendance = useCreateAttendance(date)

  // Formulir tetap memegang pilihan status dan jam, supaya mencatat beberapa
  // karyawan berturut turut cukup dengan mengganti nama.
  const submit = (event: FormEvent) => {
    event.preventDefault()
    createAttendance.mutate(values, {
      onSuccess: () => setValues((current) => ({ ...current, employeeId: '' })),
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <p className="text-sm text-slate-600">
          Dicatat untuk tanggal <span className="font-medium text-slate-900">{shortDate(`${date}T00:00:00`)}</span>.
          Ganti tanggal di penyaring di atas.
        </p>
        <div className="grid gap-4 md:grid-cols-4">
          <SelectField
            id="attendance-employee"
            label="Karyawan"
            required
            disabled={employees.isPending}
            hint={employees.isError ? 'Daftar karyawan gagal dimuat' : undefined}
            value={values.employeeId}
            onChange={change('employeeId')}
          >
            <EmployeeOptions employees={employees.data?.items ?? []} />
          </SelectField>
          <AttendanceTimeFields idPrefix="attendance-new" values={values} onChange={change} />
        </div>

        <Button type="submit" isPending={createAttendance.isPending} pendingLabel="Menyimpan absensi">
          Simpan absensi
        </Button>

        {createAttendance.isError ? <ErrorNote message={errorMessage(createAttendance.error)} /> : null}
        {createAttendance.isSuccess ? (
          <SuccessNote
            message={`Absensi ${createAttendance.data.employeeName} tercatat: ${ATTENDANCE_STATUS_LABEL[createAttendance.data.status].toLowerCase()}.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditAttendanceForm({ attendance, onSaved, onCancel }: {
  attendance: Attendance
  onSaved: (saved: Attendance) => void
  onCancel: () => void
}) {
  const { values, change } = useAttendanceValues(attendanceFormValues(attendance))
  const updateAttendance = useUpdateAttendance()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateAttendance.mutate({ attendance, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">
          Ubah absensi {attendance.employeeName}, {shortDate(attendance.date)}
        </h3>
        <div className="grid gap-4 md:grid-cols-3">
          <AttendanceTimeFields idPrefix="attendance-edit" values={values} onChange={change} />
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateAttendance.isPending} pendingLabel="Menyimpan absensi">
            Simpan perubahan absensi
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateAttendance.isError ? <ErrorNote message={errorMessage(updateAttendance.error)} /> : null}
      </form>
    </FormPanel>
  )
}

export function AttendanceSection({ today }: { today: string }) {
  const [date, setDate] = useState(today)
  const [status, setStatus] = useState<AttendanceStatus | ''>('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Attendance | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteAttendance = useDeleteAttendance()
  const isFormOpen = isCreating || editing !== null

  const attendances = useAttendances({
    dateFrom: date,
    dateTo: date,
    status: status === '' ? undefined : status,
    page,
    pageSize: PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    deleteAttendance.reset()
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

  const startEdit = (attendance: Attendance) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === attendance.id ? null : attendance))
  }

  const remove = (attendance: Attendance) => {
    deleteAttendance.mutate(attendance, {
      onSuccess: () => {
        setNotice(`Absensi ${attendance.employeeName} tanggal ${shortDate(attendance.date)} dihapus.`)
        setPage((current) => pageAfterRemoval(current, attendances.data?.items.length ?? 0))
        if (editing?.id === attendance.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card title="Absensi harian" description="Satu catatan per karyawan per tanggal">
      <Toolbar>
        <ToolbarInput
          id="attendance-date"
          label="Tanggal"
          showLabel
          type="date"
          required
          value={date}
          onChange={(event) => {
            setDate(event.target.value === '' ? today : event.target.value)
            setPage(1)
            setEditing(null)
          }}
        />
        <FilterSelect
          id="attendance-filter-status"
          label="Saring menurut status absensi"
          value={status}
          onChange={(event) => {
            setStatus(event.target.value as AttendanceStatus | '')
            setPage(1)
          }}
        >
          <option value="">Semua status</option>
          {ATTENDANCE_STATUSES.map((option) => (
            <option key={option} value={option}>
              {ATTENDANCE_STATUS_LABEL[option]}
            </option>
          ))}
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Catat absensi" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewAttendanceForm date={date} /> : null}
      {editing ? (
        <EditAttendanceForm
          key={editing.id}
          attendance={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(
              `Absensi ${saved.employeeName} diperbarui: ${ATTENDANCE_STATUS_LABEL[saved.status].toLowerCase()}.`,
            )
          }}
        />
      ) : null}

      <RowNotice
        success={notice}
        error={
          deleteAttendance.isError
            ? refusalText(`Absensi ${deleteAttendance.variables.employeeName} tidak bisa dihapus.`, deleteAttendance.error)
            : null
        }
      />

      {attendances.isPending ? <Loading /> : null}
      {attendances.isError ? <LoadFailed onRetry={() => attendances.refetch()} /> : null}
      {attendances.data ? (
        <>
          <Table
            rows={attendances.data.items}
            emptyMessage="Belum ada absensi untuk tanggal dan status ini."
            columns={[
              { header: 'Karyawan', cell: (row) => row.employeeName },
              {
                header: 'Status',
                cell: (row) => <Chip tone={ATTENDANCE_STATUS_TONE[row.status]}>{ATTENDANCE_STATUS_LABEL[row.status]}</Chip>,
              },
              { header: 'Masuk', align: 'right', cell: (row) => row.checkInTime ?? '-' },
              { header: 'Pulang', align: 'right', cell: (row) => row.checkOutTime ?? '-' },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <ActionGroup>
                    {askingId === row.id ? null : (
                      <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                    )}
                    <ConfirmAction
                      label="Hapus"
                      srLabel={`absensi ${row.employeeName}`}
                      question={`Hapus absensi ${row.employeeName}?`}
                      confirmLabel="Ya, hapus"
                      pendingLabel="Menghapus"
                      isAsking={askingId === row.id}
                      isPending={deleteAttendance.isPending && deleteAttendance.variables.id === row.id}
                      onAsk={() => {
                        clearNotes()
                        setAskingId(row.id)
                      }}
                      onCancel={() => setAskingId(null)}
                      onConfirm={() => remove(row)}
                    />
                  </ActionGroup>
                ),
              },
            ]}
          />
          <Pager
            page={attendances.data.page}
            totalPages={attendances.data.totalPages}
            totalItems={attendances.data.totalItems}
            unit="catatan"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
