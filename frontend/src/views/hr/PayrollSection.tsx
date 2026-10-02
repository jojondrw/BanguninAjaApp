import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useActiveEmployeeOptions,
  useCreatePayroll,
  useDeletePayroll,
  usePayPayroll,
  usePayrolls,
  useUpdatePayroll,
} from '../../controllers/useHr'
import { EMPTY_PAYROLL_FORM, payrollFormValues, type Payroll, type PayrollFormValues } from '../../models/hr'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { monthLabel, rupiah } from '../../shared/format'
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
import { amountHint, PAGE_SIZE, pageAfterRemoval, refusalText } from './hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from './RowActions'

type PaidFilter = 'true' | 'false' | ''
type PayrollAsk = { id: string; kind: 'pay' | 'delete' } | null
type ChangePayroll = (key: keyof PayrollFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

function usePayrollValues(initial: PayrollFormValues) {
  const [values, setValues] = useState<PayrollFormValues>(initial)

  const change: ChangePayroll = (key) => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

function AdjustmentFields({ idPrefix, values, onChange }: {
  idPrefix: string
  values: PayrollFormValues
  onChange: ChangePayroll
}) {
  return (
    <>
      <Field
        id={`${idPrefix}-allowance`}
        label="Tunjangan (opsional)"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        hint={amountHint(values.allowance, 'Dalam Rupiah')}
        value={values.allowance}
        onChange={onChange('allowance')}
      />
      <Field
        id={`${idPrefix}-deduction`}
        label="Potongan (opsional)"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        hint={amountHint(values.deduction, 'Tidak boleh melebihi gaji pokok ditambah tunjangan')}
        value={values.deduction}
        onChange={onChange('deduction')}
      />
    </>
  )
}

function NewPayrollForm({ period }: { period: string }) {
  const { values, setValues, change } = usePayrollValues(EMPTY_PAYROLL_FORM)
  const employees = useActiveEmployeeOptions()
  const createPayroll = useCreatePayroll(period)

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
            onChange={change('employeeId')}
          >
            <EmployeeOptions employees={employees.data?.items ?? []} />
          </SelectField>
          <AdjustmentFields idPrefix="payroll-new" values={values} onChange={change} />
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

function EditPayrollForm({ payroll, onSaved, onCancel }: {
  payroll: Payroll
  onSaved: (saved: Payroll) => void
  onCancel: () => void
}) {
  const { values, change } = usePayrollValues(payrollFormValues(payroll))
  const updatePayroll = useUpdatePayroll()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updatePayroll.mutate({ payroll, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <div>
          <h3 className="text-[15px] font-semibold text-slate-900">
            Ubah slip {payroll.employeeName}, {monthLabel(payroll.period)}
          </h3>
          <p className="mt-1 text-sm text-slate-600">
            Gaji pokok dihitung ulang dari data karyawan dan absensi terbaru saat disimpan. Sekarang{' '}
            {rupiah(payroll.basicPay)}.
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <AdjustmentFields idPrefix="payroll-edit" values={values} onChange={change} />
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updatePayroll.isPending} pendingLabel="Menghitung gaji">
            Simpan perubahan slip
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updatePayroll.isError ? <ErrorNote message={errorMessage(updatePayroll.error)} /> : null}
      </form>
    </FormPanel>
  )
}

// Pembayaran tidak bisa dibatalkan, jadi diminta konfirmasi sekali lagi di
// baris yang sama, bukan lewat confirm() yang memblokir halaman.
function PayAction({ payroll, isAsking, isPaying, onAsk, onCancel, onConfirm }: {
  payroll: Payroll
  isAsking: boolean
  isPaying: boolean
  onAsk: () => void
  onCancel: () => void
  onConfirm: () => void
}) {
  if (payroll.paid) {
    return <Chip tone="bg-green-100 text-green-800">Lunas</Chip>
  }

  if (!isAsking && !isPaying) {
    return (
      <button
        type="button"
        onClick={onAsk}
        className="h-7 rounded-lg bg-white px-2.5 text-xs font-medium text-slate-700 shadow-hairline transition-[background-color,transform]
                   hover:bg-slate-50 motion-safe:active:scale-[0.97]"
      >
        Tandai dibayar<span className="sr-only">: {payroll.employeeName}</span>
      </button>
    )
  }

  return (
    <span className="inline-flex items-center gap-2">
      <button
        type="button"
        disabled={isPaying}
        aria-busy={isPaying}
        onClick={onConfirm}
        className="rounded-lg bg-navy-700 px-3 py-1 text-xs font-medium text-white hover:bg-navy-900
                   disabled:cursor-not-allowed disabled:bg-navy-700/60"
      >
        {isPaying ? 'Menyimpan' : 'Ya, sudah dibayar'}
      </button>
      {isPaying ? null : (
        <button
          type="button"
          autoFocus
          onClick={onCancel}
          className="rounded-lg px-2 py-1 text-xs text-slate-600 hover:bg-slate-100"
        >
          Batal
        </button>
      )}
    </span>
  )
}

export function PayrollSection({ period, onPeriodChange, projects }: {
  period: string
  onPeriodChange: (period: string) => void
  projects: Project[]
}) {
  const [periodDraft, setPeriodDraft] = useState(period)
  const [paid, setPaid] = useState<PaidFilter>('')
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Payroll | null>(null)
  const [asking, setAsking] = useState<PayrollAsk>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const pay = usePayPayroll()
  const deletePayroll = useDeletePayroll()
  const isFormOpen = isCreating || editing !== null

  const payrolls = usePayrolls({
    period,
    projectId: projectId === '' ? undefined : projectId,
    paid: paid === '' ? undefined : paid,
    page,
    pageSize: PAGE_SIZE,
  })

  const clearNotes = () => {
    setNotice(null)
    pay.reset()
    deletePayroll.reset()
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

  const startEdit = (payroll: Payroll) => {
    clearNotes()
    setAsking(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === payroll.id ? null : payroll))
  }

  const ask = (payroll: Payroll, kind: 'pay' | 'delete') => {
    clearNotes()
    setAsking({ id: payroll.id, kind })
  }

  const isAsking = (payroll: Payroll, kind: 'pay' | 'delete') =>
    asking !== null && asking.id === payroll.id && asking.kind === kind

  const markPaid = (payroll: Payroll) => {
    pay.mutate(payroll.id, {
      onSuccess: (saved) => {
        setNotice(`Gaji ${saved.employeeName} periode ${monthLabel(saved.period)} ditandai lunas.`)
        if (editing?.id === payroll.id) {
          closeForm()
        }
      },
      onSettled: () => setAsking(null),
    })
  }

  const remove = (payroll: Payroll) => {
    deletePayroll.mutate(payroll, {
      onSuccess: () => {
        setNotice(`Slip ${payroll.employeeName} periode ${monthLabel(payroll.period)} dihapus.`)
        setPage((current) => pageAfterRemoval(current, payrolls.data?.items.length ?? 0))
        if (editing?.id === payroll.id) {
          closeForm()
        }
      },
      onSettled: () => setAsking(null),
    })
  }

  const rowError = pay.isError
    ? refusalText('Slip belum ditandai lunas.', pay.error)
    : deletePayroll.isError
      ? refusalText(`Slip ${deletePayroll.variables.employeeName} tidak bisa dihapus.`, deletePayroll.error)
      : null

  return (
    <Card title="Penggajian" description="Slip gaji per periode. Slip yang sudah dibayar tidak bisa diubah atau dihapus.">
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
              setEditing(null)
            }
          }}
        />
        <FilterSelect
          id="payroll-filter-project"
          label="Saring menurut proyek karyawan"
          value={projectId}
          onChange={(event) => {
            setProjectId(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Semua proyek</option>
          {projects.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name}
            </option>
          ))}
        </FilterSelect>
        <FilterSelect
          id="payroll-filter-paid"
          label="Saring menurut status pembayaran"
          value={paid}
          onChange={(event) => {
            setPaid(event.target.value as PaidFilter)
            setPage(1)
          }}
        >
          <option value="">Semua slip</option>
          <option value="false">Belum dibayar</option>
          <option value="true">Sudah dibayar</option>
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Buat slip gaji" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewPayrollForm period={period} /> : null}
      {editing ? (
        <EditPayrollForm
          key={editing.id}
          payroll={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Slip ${saved.employeeName} diperbarui, gaji bersih ${rupiah(saved.netPay)}.`)
          }}
        />
      ) : null}

      <RowNotice success={notice} error={rowError} />

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
                    payroll={row}
                    isAsking={isAsking(row, 'pay')}
                    isPaying={pay.isPending && pay.variables === row.id}
                    onAsk={() => ask(row, 'pay')}
                    onCancel={() => setAsking(null)}
                    onConfirm={() => markPaid(row)}
                  />
                ),
              },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) =>
                  row.paid ? (
                    <span className="text-xs text-slate-500">Terkunci</span>
                  ) : (
                    <ActionGroup>
                      {asking?.id === row.id ? null : (
                        <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                      )}
                      {isAsking(row, 'pay') ? null : (
                        <ConfirmAction
                          label="Hapus"
                          srLabel={`slip ${row.employeeName}`}
                          question={`Hapus slip ${row.employeeName}?`}
                          confirmLabel="Ya, hapus"
                          pendingLabel="Menghapus"
                          isAsking={isAsking(row, 'delete')}
                          isPending={deletePayroll.isPending && deletePayroll.variables.id === row.id}
                          onAsk={() => ask(row, 'delete')}
                          onCancel={() => setAsking(null)}
                          onConfirm={() => remove(row)}
                        />
                      )}
                    </ActionGroup>
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
