import type { Employee } from '../../models/hr'

export function EmployeeOptions({ employees }: { employees: Employee[] }) {
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
