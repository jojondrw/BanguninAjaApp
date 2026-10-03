import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useCreateEmployee, useUpdateEmployee } from '../../controllers/useHr'
import {
  EMPLOYEE_IDENTITY_MAX_LENGTH,
  EMPLOYEE_NAME_MAX_LENGTH,
  EMPLOYEE_POSITION_MAX_LENGTH,
  EMPLOYMENT_TYPES,
  EMPLOYMENT_TYPE_LABEL,
  EMPTY_EMPLOYEE_FORM,
  employeeFormValues,
  type Employee,
  type EmployeeFormValues,
} from '../../models/hr'
import { projectOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { FormPanel, SelectField } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { amountHint } from './hrShared'

function useEmployeeValues(initial: EmployeeFormValues) {
  const [values, setValues] = useState<EmployeeFormValues>(initial)

  const change =
    (key: keyof EmployeeFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  return { values, setValues, change }
}

function EmployeeFields({ idPrefix, values, onChange, onProject, withLeftDate = false }: {
  idPrefix: string
  values: EmployeeFormValues
  onChange: (key: keyof EmployeeFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onProject: (projectId: string) => void
  withLeftDate?: boolean
}) {
  const isDaily = values.employmentType === 'daily'

  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Field
        id={`${idPrefix}-identity`}
        label="Nomor induk"
        placeholder="KRY-001"
        autoComplete="off"
        maxLength={EMPLOYEE_IDENTITY_MAX_LENGTH}
        hint={`Unik, maksimal ${EMPLOYEE_IDENTITY_MAX_LENGTH} karakter`}
        required
        value={values.identityNumber}
        onChange={onChange('identityNumber')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama lengkap"
        placeholder="Budi Santoso"
        autoComplete="off"
        maxLength={EMPLOYEE_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <Field
        id={`${idPrefix}-position`}
        label="Jabatan"
        placeholder="Mandor"
        autoComplete="off"
        maxLength={EMPLOYEE_POSITION_MAX_LENGTH}
        required
        value={values.position}
        onChange={onChange('position')}
      />
      <SelectField
        id={`${idPrefix}-type`}
        label="Jenis kepegawaian"
        hint="Karyawan harian digaji per hari hadir"
        value={values.employmentType}
        onChange={onChange('employmentType')}
      >
        {EMPLOYMENT_TYPES.map((type) => (
          <option key={type} value={type}>
            {EMPLOYMENT_TYPE_LABEL[type]}
          </option>
        ))}
      </SelectField>
      <SearchSelect
        {...projectOptions}
        id={`${idPrefix}-project`}
        label="Ditempatkan di proyek (opsional)"
        placeholder="Cari proyek"
        allowEmpty
        emptyLabel="Kantor pusat, tanpa proyek"
        value={values.projectId}
        onChange={onProject}
      />
      <Field
        id={`${idPrefix}-joined`}
        label="Tanggal bergabung"
        type="date"
        required
        max={values.leftDate || undefined}
        value={values.joinedDate}
        onChange={onChange('joinedDate')}
      />
      {withLeftDate ? (
        <Field
          id={`${idPrefix}-left`}
          label="Tanggal keluar (opsional)"
          type="date"
          min={values.joinedDate || undefined}
          hint="Kosongkan selama masih bekerja. Hari terakhir masih dihitung aktif."
          value={values.leftDate}
          onChange={onChange('leftDate')}
        />
      ) : null}
      <Field
        id={`${idPrefix}-salary`}
        label={isDaily ? 'Upah harian' : 'Gaji pokok bulanan'}
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        placeholder="0"
        hint={amountHint(values.baseSalary, 'Dalam Rupiah')}
        value={values.baseSalary}
        onChange={onChange('baseSalary')}
      />
    </div>
  )
}

export function NewEmployeeForm() {
  const { values, setValues, change } = useEmployeeValues(EMPTY_EMPLOYEE_FORM)
  const createEmployee = useCreateEmployee()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createEmployee.mutate(values, { onSuccess: () => setValues(EMPTY_EMPLOYEE_FORM) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <EmployeeFields
          idPrefix="employee-new"
          values={values}
          onChange={change}
          onProject={(projectId) => setValues((current) => ({ ...current, projectId }))}
        />

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

export function EditEmployeeForm({ employee, onSaved, onCancel }: {
  employee: Employee
  onSaved: (saved: Employee) => void
  onCancel: () => void
}) {
  const { values, setValues, change } = useEmployeeValues(employeeFormValues(employee))
  const updateEmployee = useUpdateEmployee()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateEmployee.mutate({ employee, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">Ubah data {employee.name}</h3>
        <EmployeeFields
          idPrefix="employee-edit"
          values={values}
          onChange={change}
          onProject={(projectId) => setValues((current) => ({ ...current, projectId }))}
          withLeftDate
        />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateEmployee.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan karyawan
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateEmployee.isError ? <ErrorNote message={errorMessage(updateEmployee.error)} /> : null}
      </form>
    </FormPanel>
  )
}
