import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../controllers/useErp'
import {
  useActiveEmployeeOptions,
  useAttendanceCount,
  useAttendances,
  useCreateAttendance,
  useCreateEmployee,
  useCreatePayroll,
  useEmployeeHeadcount,
  useEmployees,
  usePayPayroll,
  usePayrollTotals,
  usePayrolls,
} from '../controllers/useHr'
import {
  ATTENDANCE_STATUSES,
  ATTENDANCE_STATUS_LABEL,
  ATTENDANCE_STATUS_TONE,
  EMPLOYEE_IDENTITY_MAX_LENGTH,
  EMPLOYEE_NAME_MAX_LENGTH,
  EMPLOYEE_POSITION_MAX_LENGTH,
  EMPLOYMENT_TYPES,
  EMPLOYMENT_TYPE_LABEL,
  EMPTY_ATTENDANCE_FORM,
  EMPTY_EMPLOYEE_FORM,
  EMPTY_PAYROLL_FORM,
  type AttendanceFormValues,
  type AttendanceStatus,
  type Employee,
  type EmployeeFormValues,
  type EmploymentType,
  type PayrollFormValues,
} from '../models/hr'
import type { Project } from '../models/project'
import { errorMessage } from '../shared/errorMessage'
import { monthLabel, number, rupiah, rupiahShort, shortDate } from '../shared/format'
import { currentPeriod, todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { Card, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'
import { Button, ErrorNote, Field, SuccessNote } from './components/Form'
import {
  Chip,
  FilterSelect,
  FormPanel,
  FormToggle,
  Pager,
  SelectField,
  Toolbar,
  ToolbarInput,
} from './components/RecordControls'
import { kpiText } from './components/kpiText'

const PAGE_SIZE = 20

type ActiveFilter = 'true' | 'false' | ''

function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return fallback
  }
  return rupiah(amount)
}

function headcountNote(byType: { employmentType: EmploymentType; total: number }[]): string {
  if (byType.length === 0) {
    return 'Belum ada karyawan aktif'
  }
  return byType.map((count) => `${count.total} ${EMPLOYMENT_TYPE_LABEL[count.employmentType].toLowerCase()}`).join(', ')
}

function salaryText(employee: Employee): string {
  const amount = rupiah(employee.baseSalary)
  return employee.employmentType === 'daily' ? `${amount} /hari` : amount
}

function EmployeeOptions({ employees }: { employees: Employee[] }) {
  return (
    <>
      <option value="">Pilih karyawan</option>
      {employees.map((employee) => (
        <option key={employee.id} value={employee.id}>
          {employee.name} ({employee.identityNumber})
        </option>
      ))}
    </>
  )
}

function NewEmployeeForm({ projects }: { projects: Project[] }) {
  const [values, setValues] = useState<EmployeeFormValues>(EMPTY_EMPLOYEE_FORM)
  const createEmployee = useCreateEmployee()
  const isDaily = values.employmentType === 'daily'

  const update =
    (key: keyof EmployeeFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createEmployee.mutate(values, { onSuccess: () => setValues(EMPTY_EMPLOYEE_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="employee-identity"
            label="Nomor induk"
            placeholder="KRY-001"
            autoComplete="off"
            maxLength={EMPLOYEE_IDENTITY_MAX_LENGTH}
            hint={`Unik, maksimal ${EMPLOYEE_IDENTITY_MAX_LENGTH} karakter`}
            required
            value={values.identityNumber}
            onChange={update('identityNumber')}
          />
          <Field
            id="employee-name"
            label="Nama lengkap"
            placeholder="Budi Santoso"
            autoComplete="off"
            maxLength={EMPLOYEE_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="employee-position"
            label="Jabatan"
            placeholder="Mandor"
            autoComplete="off"
            maxLength={EMPLOYEE_POSITION_MAX_LENGTH}
            required
            value={values.position}
            onChange={update('position')}
          />
          <SelectField
            id="employee-type"
            label="Jenis kepegawaian"
            hint="Karyawan harian digaji per hari hadir"
            value={values.employmentType}
            onChange={update('employmentType')}
          >
            {EMPLOYMENT_TYPES.map((type) => (
              <option key={type} value={type}>
                {EMPLOYMENT_TYPE_LABEL[type]}
              </option>
            ))}
          </SelectField>
          <SelectField
            id="employee-project"
            label="Ditempatkan di proyek (opsional)"
            value={values.projectId}
            onChange={update('projectId')}
          >
            <option value="">Kantor pusat, tanpa proyek</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name}
              </option>
            ))}
          </SelectField>
          <Field
            id="employee-joined"
            label="Tanggal bergabung"
            type="date"
            required
            value={values.joinedDate}
            onChange={update('joinedDate')}
          />
          <Field
            id="employee-salary"
            label={isDaily ? 'Upah harian' : 'Gaji pokok bulanan'}
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={amountHint(values.baseSalary, 'Dalam Rupiah')}
            value={values.baseSalary}
            onChange={update('baseSalary')}
          />
        </div>

        <Button type="submit" isPending={createEmployee.isPending} pendingLabel="Menyimpan karyawan">
          Simpan karyawan
        </Button>

        {createEmployee.isError ? <ErrorNote message={errorMessage(createEmployee.error)} /> : null}
        {createEmployee.isSuccess ? (
          <SuccessNote message={`${createEmployee.data.name} berhasil ditambahkan sebagai karyawan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EmployeesCard({ projects, projectName }: {
  projects: Project[]
  projectName: (id: string | null) => string
}) {
  const [search, setSearch] = useState('')
  const [projectId, setProjectId] = useState('')
  const [employmentType, setEmploymentType] = useState<EmploymentType | ''>('')
  const [active, setActive] = useState<ActiveFilter>('true')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)

  const employees = useEmployees({
    search: search.trim() === '' ? undefined : search.trim(),
    projectId: projectId === '' ? undefined : projectId,
    employmentType: employmentType === '' ? undefined : employmentType,
    active: active === '' ? undefined : active,
    page,
    pageSize: PAGE_SIZE,
  })

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  return (
    <Card title="Karyawan" description="Disaring dan dicari langsung di server">
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
        <FormToggle
          isOpen={isFormOpen}
          openLabel="Tambah karyawan"
          onToggle={() => setIsFormOpen((open) => !open)}
        />
      </Toolbar>

      {isFormOpen ? <NewEmployeeForm projects={projects} /> : null}

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
              {
                header: 'Status',
                cell: (row) => {
                  if (row.active) {
                    return <Chip tone="bg-green-100 text-green-800">Aktif</Chip>
                  }
                  if (row.leftDate === null) {
                    return <Chip tone="bg-blue-100 text-blue-800">Mulai {shortDate(row.joinedDate)}</Chip>
                  }
                  return <Chip tone="bg-slate-100 text-slate-700">Keluar {shortDate(row.leftDate)}</Chip>
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

function NewAttendanceForm({ date }: { date: string }) {
  const [values, setValues] = useState<AttendanceFormValues>(EMPTY_ATTENDANCE_FORM)
  const employees = useActiveEmployeeOptions()
  const createAttendance = useCreateAttendance(date)
  const isPresent = values.status === 'present'

  const update =
    (key: keyof AttendanceFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

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
            onChange={update('employeeId')}
          >
            <EmployeeOptions employees={employees.data?.items ?? []} />
          </SelectField>
          <SelectField id="attendance-status" label="Status" value={values.status} onChange={update('status')}>
            {ATTENDANCE_STATUSES.map((status) => (
              <option key={status} value={status}>
                {ATTENDANCE_STATUS_LABEL[status]}
              </option>
            ))}
          </SelectField>
          <Field
            id="attendance-check-in"
            label="Jam masuk"
            type="time"
            required={isPresent}
            disabled={!isPresent}
            hint={isPresent ? 'Wajib untuk yang hadir' : 'Hanya untuk yang hadir'}
            value={isPresent ? values.checkInTime : ''}
            onChange={update('checkInTime')}
          />
          <Field
            id="attendance-check-out"
            label="Jam pulang (opsional)"
            type="time"
            disabled={!isPresent}
            min={values.checkInTime || undefined}
            value={isPresent ? values.checkOutTime : ''}
            onChange={update('checkOutTime')}
          />
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

function AttendanceCard({ today }: { today: string }) {
  const [date, setDate] = useState(today)
  const [status, setStatus] = useState<AttendanceStatus | ''>('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)

  const attendances = useAttendances({
    dateFrom: date,
    dateTo: date,
    status: status === '' ? undefined : status,
    page,
    pageSize: PAGE_SIZE,
  })

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
        <FormToggle isOpen={isFormOpen} openLabel="Catat absensi" onToggle={() => setIsFormOpen((open) => !open)} />
      </Toolbar>

      {isFormOpen ? <NewAttendanceForm date={date} /> : null}

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

function NewPayrollForm({ period }: { period: string }) {
  const [values, setValues] = useState<PayrollFormValues>(EMPTY_PAYROLL_FORM)
  const employees = useActiveEmployeeOptions()
  const createPayroll = useCreatePayroll(period)

  const update =
    (key: keyof PayrollFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createPayroll.mutate(values, { onSuccess: () => setValues(EMPTY_PAYROLL_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <p className="text-sm text-slate-600">
          Slip untuk periode <span className="font-medium text-slate-900">{monthLabel(period)}</span>. Gaji pokok
          dihitung server: gaji bulanan untuk karyawan tetap dan kontrak, upah harian dikali hari hadir untuk
          karyawan harian.
        </p>
        <div className="grid gap-4 md:grid-cols-3">
          <SelectField
            id="payroll-employee"
            label="Karyawan"
            required
            disabled={employees.isPending}
            hint={employees.isError ? 'Daftar karyawan gagal dimuat' : undefined}
            value={values.employeeId}
            onChange={update('employeeId')}
          >
            <EmployeeOptions employees={employees.data?.items ?? []} />
          </SelectField>
          <Field
            id="payroll-allowance"
            label="Tunjangan (opsional)"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={amountHint(values.allowance, 'Dalam Rupiah')}
            value={values.allowance}
            onChange={update('allowance')}
          />
          <Field
            id="payroll-deduction"
            label="Potongan (opsional)"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={amountHint(values.deduction, 'Tidak boleh melebihi gaji pokok ditambah tunjangan')}
            value={values.deduction}
            onChange={update('deduction')}
          />
        </div>

        <Button type="submit" isPending={createPayroll.isPending} pendingLabel="Menghitung gaji">
          Buat slip gaji
        </Button>

        {createPayroll.isError ? <ErrorNote message={errorMessage(createPayroll.error)} /> : null}
        {createPayroll.isSuccess ? (
          <SuccessNote
            message={`Slip ${createPayroll.data.employeeName} dibuat, gaji bersih ${rupiah(createPayroll.data.netPay)}.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}

function PayAction({ payrollId, isPaid, confirmingId, onConfirm, pay }: {
  payrollId: string
  isPaid: boolean
  confirmingId: string | null
  onConfirm: (id: string | null) => void
  pay: ReturnType<typeof usePayPayroll>
}) {
  if (isPaid) {
    return <Chip tone="bg-green-100 text-green-800">Lunas</Chip>
  }

  const isPaying = pay.isPending && pay.variables === payrollId

  if (confirmingId !== payrollId && !isPaying) {
    return (
      <button
        type="button"
        onClick={() => onConfirm(payrollId)}
        className="rounded-lg border border-slate-300 bg-white px-3 py-1 text-xs font-medium text-slate-700 hover:bg-slate-50"
      >
        Tandai dibayar
      </button>
    )
  }

  // Pembayaran tidak bisa dibatalkan, jadi diminta konfirmasi sekali lagi di
  // baris yang sama, bukan lewat confirm() yang memblokir halaman.
  return (
    <span className="inline-flex items-center gap-2">
      <button
        type="button"
        disabled={isPaying}
        aria-busy={isPaying}
        onClick={() => pay.mutate(payrollId, { onSettled: () => onConfirm(null) })}
        className="rounded-lg bg-navy-700 px-3 py-1 text-xs font-medium text-white hover:bg-navy-900
                   disabled:cursor-not-allowed disabled:bg-navy-700/60"
      >
        {isPaying ? 'Menyimpan' : 'Ya, sudah dibayar'}
      </button>
      {isPaying ? null : (
        <button
          type="button"
          onClick={() => onConfirm(null)}
          className="rounded-lg px-2 py-1 text-xs text-slate-600 hover:bg-slate-100"
        >
          Batal
        </button>
      )}
    </span>
  )
}

function PayrollCard({ period, onPeriodChange }: { period: string; onPeriodChange: (period: string) => void }) {
  const [periodDraft, setPeriodDraft] = useState(period)
  const [paid, setPaid] = useState<'true' | 'false' | ''>('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [confirmingId, setConfirmingId] = useState<string | null>(null)
  const pay = usePayPayroll()

  const payrolls = usePayrolls({
    period,
    paid: paid === '' ? undefined : paid,
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <Card title="Penggajian" description="Slip gaji per periode. Slip yang sudah dibayar tidak bisa diubah.">
      <Toolbar>
        <ToolbarInput
          id="payroll-period"
          label="Periode"
          showLabel
          type="month"
          required
          placeholder="2026-09"
          pattern="[0-9]{4}-[0-9]{2}"
          value={periodDraft}
          onChange={(event) => {
            // Peramban tanpa pemilih bulan menampilkan isian teks biasa, jadi
            // ketikan setengah jadi disimpan dulu dan periode baru berganti
            // setelah formatnya lengkap.
            setPeriodDraft(event.target.value)
            if (/^\d{4}-(0[1-9]|1[0-2])$/.test(event.target.value)) {
              onPeriodChange(event.target.value)
              setPage(1)
            }
          }}
        />
        <FilterSelect
          id="payroll-filter-paid"
          label="Saring menurut status pembayaran"
          value={paid}
          onChange={(event) => {
            setPaid(event.target.value as 'true' | 'false' | '')
            setPage(1)
          }}
        >
          <option value="">Semua slip</option>
          <option value="false">Belum dibayar</option>
          <option value="true">Sudah dibayar</option>
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Buat slip gaji" onToggle={() => setIsFormOpen((open) => !open)} />
      </Toolbar>

      {isFormOpen ? <NewPayrollForm period={period} /> : null}

      {pay.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(pay.error)} />
        </div>
      ) : null}
      {pay.isSuccess ? (
        <div className="mb-4">
          <SuccessNote message={`Gaji ${pay.data.employeeName} periode ${monthLabel(pay.data.period)} ditandai lunas.`} />
        </div>
      ) : null}

      {payrolls.isPending ? <Loading /> : null}
      {payrolls.isError ? <LoadFailed onRetry={() => payrolls.refetch()} /> : null}
      {payrolls.data ? (
        <>
          <Table
            rows={payrolls.data.items}
            emptyMessage={`Belum ada slip gaji untuk ${monthLabel(period)}.`}
            columns={[
              { header: 'Karyawan', cell: (row) => row.employeeName },
              { header: 'Gaji pokok', align: 'right', cell: (row) => rupiah(row.basicPay) },
              { header: 'Tunjangan', align: 'right', cell: (row) => rupiah(row.allowance) },
              { header: 'Potongan', align: 'right', cell: (row) => rupiah(row.deduction) },
              {
                header: 'Gaji bersih',
                align: 'right',
                cell: (row) => <span className="font-medium text-slate-900">{rupiah(row.netPay)}</span>,
              },
              {
                header: 'Pembayaran',
                align: 'right',
                cell: (row) => (
                  <PayAction
                    payrollId={row.id}
                    isPaid={row.paid}
                    confirmingId={confirmingId}
                    onConfirm={setConfirmingId}
                    pay={pay}
                  />
                ),
              },
            ]}
          />
          <Pager
            page={payrolls.data.page}
            totalPages={payrolls.data.totalPages}
            totalItems={payrolls.data.totalItems}
            unit="slip"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}

export function HrPage() {
  const [today] = useState(todayDate)
  const [period, setPeriod] = useState(currentPeriod)
  const headcount = useEmployeeHeadcount()
  const presentToday = useAttendanceCount({ dateFrom: today, dateTo: today, status: 'present' })
  const recordedToday = useAttendanceCount({ dateFrom: today, dateTo: today })
  const payrollTotals = usePayrollTotals(period)
  const projects = useProjects({ pageSize: 100 })
  const projectList = projects.data?.items ?? []

  const projectName = (id: string | null) => {
    if (id === null) {
      return 'Kantor pusat'
    }
    return projectList.find((project) => project.id === id)?.name ?? id.slice(0, 8)
  }

  const totalsNote = payrollTotals.data?.isPartial ? 'dari 100 slip pertama' : undefined

  return (
    <AppShell title="SDM" description="Karyawan, absensi harian, dan penggajian per periode">
      <KpiRow>
        <Kpi
          label="Karyawan aktif"
          value={kpiText(headcount, (data) => number(data.total))}
          note={headcount.data ? headcountNote(headcount.data.byType) : undefined}
        />
        <Kpi
          label="Hadir hari ini"
          value={kpiText(presentToday, number)}
          note={recordedToday.data === undefined ? undefined : `dari ${number(recordedToday.data)} absensi tercatat`}
        />
        <Kpi
          label={`Gaji bersih ${monthLabel(period)}`}
          value={kpiText(payrollTotals, (data) => rupiahShort(data.netPay))}
          note={payrollTotals.data ? (totalsNote ?? `${number(payrollTotals.data.count)} slip gaji`) : undefined}
        />
        <Kpi
          label="Belum dibayar"
          value={kpiText(payrollTotals, (data) => rupiahShort(data.unpaidNetPay))}
          note={payrollTotals.data ? `${number(payrollTotals.data.unpaidCount)} slip menunggu pembayaran` : undefined}
        />
      </KpiRow>

      <div className="mt-6 grid gap-6">
        <EmployeesCard projects={projectList} projectName={projectName} />
        <AttendanceCard today={today} />
        <PayrollCard period={period} onPeriodChange={setPeriod} />
      </div>
    </AppShell>
  )
}
