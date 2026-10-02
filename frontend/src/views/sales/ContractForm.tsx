import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../../controllers/useErp'
import { useCreateContract, useCustomers, useUnits, useUpdateContract } from '../../controllers/useSales'
import {
  CONTRACT_NUMBER_MAX_LENGTH,
  CONTRACT_TYPE_LABEL,
  contractToForm,
  emptyContractForm,
  type Contract,
  type ContractFormValues,
  type ContractType,
} from '../../models/sales'
import { errorMessage } from '../../shared/errorMessage'
import { Empty, Loading } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { FormPanel, SelectField } from '../components/RecordControls'
import { OPTION_LIMIT, amountHint, projectLabel } from './salesShared'

const CONTRACT_TYPES: ContractType[] = ['installment', 'mortgage', 'cash']

// Kontrak baru memilih unit yang tersedia. Saat mengubah draf, unitnya tetap
// karena backend menolak penggantian unit (contract_unit_change_invalid).
export function ContractForm({ contract, onSaved, onCancel }: {
  contract: Contract | null
  onSaved?: (saved: Contract) => void
  onCancel?: () => void
}) {
  const [values, setValues] = useState<ContractFormValues>(() =>
    contract ? contractToForm(contract) : emptyContractForm(),
  )
  const createContract = useCreateContract()
  const updateContract = useUpdateContract()
  const saving = contract ? updateContract : createContract
  const customers = useCustomers({ pageSize: OPTION_LIMIT })
  const units = useUnits({ status: 'available', pageSize: OPTION_LIMIT })
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const customerItems = customers.data?.items ?? []
  const unitItems = units.data?.items ?? []
  const projectItems = projects.data?.items ?? []
  const isCurrentCustomerListed = customerItems.some((customer) => customer.id === values.customerId)

  const update = (key: keyof ContractFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  // Nilai kontrak diisi otomatis dari harga unit, selama pengguna belum
  // mengetik nilainya sendiri.
  const chooseUnit = (event: ChangeEvent<HTMLSelectElement>) => {
    const unit = unitItems.find((item) => item.id === event.target.value)
    setValues((current) => ({
      ...current,
      unitId: event.target.value,
      value: current.value === '' && unit && unit.price > 0 ? String(unit.price) : current.value,
    }))
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (contract) {
      updateContract.mutate({ id: contract.id, values }, { onSuccess: (saved) => onSaved?.(saved) })
      return
    }
    createContract.mutate(values, { onSuccess: () => setValues(emptyContractForm()) })
  }

  const isLoadingOptions = customers.isPending || (contract === null && units.isPending)
  const missing =
    isLoadingOptions || contract
      ? null
      : customerItems.length === 0
        ? 'Belum ada pelanggan. Tambahkan dulu di bagian Pelanggan.'
        : unitItems.length === 0
          ? 'Tidak ada unit yang tersedia. Tambahkan unit atau lepas unit yang ditahan di bagian Unit.'
          : null

  return (
    <FormPanel>
      <div className="mb-4">
        <h3 className="text-sm font-semibold text-slate-900">
          {contract ? `Ubah draf kontrak ${contract.number}` : 'Kontrak baru'}
        </h3>
        <p className="mt-0.5 text-xs text-slate-500">
          {contract
            ? 'Draf masih bisa diubah. Setelah diaktifkan, isi kontrak terkunci.'
            : 'Kontrak baru berstatus Draf dan langsung mengunci unitnya menjadi Dipesan.'}
        </p>
      </div>
      {isLoadingOptions ? <Loading label="Mengambil pelanggan dan unit..." /> : null}
      {missing ? <Empty message={missing} /> : null}
      {!isLoadingOptions && missing === null ? (
        <form onSubmit={submit} className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            <Field
              id="contract-number"
              label="Nomor kontrak"
              placeholder="KTR-2026-001"
              autoComplete="off"
              autoFocus={contract !== null}
              maxLength={CONTRACT_NUMBER_MAX_LENGTH}
              required
              value={values.number}
              onChange={update('number')}
            />
            <SelectField
              id="contract-customer"
              label="Pelanggan"
              required
              value={values.customerId}
              onChange={update('customerId')}
            >
              <option value="">Pilih pelanggan</option>
              {contract && !isCurrentCustomerListed ? (
                <option value={contract.customerId}>{contract.customerName}</option>
              ) : null}
              {customerItems.map((customer) => (
                <option key={customer.id} value={customer.id}>
                  {customer.name} ({customer.identityNumber})
                </option>
              ))}
            </SelectField>
            {contract ? (
              <Field
                id="contract-unit"
                label="Unit"
                hint="Unit tidak bisa diganti. Hapus draf lalu buat kontrak baru."
                disabled
                value={contract.unitCode}
                readOnly
              />
            ) : (
              <SelectField
                id="contract-unit"
                label="Unit"
                hint="Hanya unit berstatus Tersedia"
                required
                value={values.unitId}
                onChange={chooseUnit}
              >
                <option value="">Pilih unit</option>
                {unitItems.map((unit) => (
                  <option key={unit.id} value={unit.id}>
                    {unit.code}, {unit.unitType} ({projectLabel(projectItems, unit.projectId)})
                  </option>
                ))}
              </SelectField>
            )}
            <SelectField id="contract-type" label="Cara bayar" value={values.type} onChange={update('type')}>
              {CONTRACT_TYPES.map((type) => (
                <option key={type} value={type}>
                  {CONTRACT_TYPE_LABEL[type]}
                </option>
              ))}
            </SelectField>
            <Field
              id="contract-value"
              label="Nilai kontrak"
              type="number"
              inputMode="numeric"
              min={1}
              step={1}
              required
              hint={amountHint(values.value, 'Terisi dari harga unit, boleh diubah')}
              value={values.value}
              onChange={update('value')}
            />
            <Field
              id="contract-date"
              label="Tanggal kontrak"
              type="date"
              required
              value={values.date}
              onChange={update('date')}
            />
          </div>

          <div className="flex flex-wrap gap-2">
            <Button type="submit" isPending={saving.isPending} pendingLabel="Menyimpan kontrak">
              {contract ? 'Simpan perubahan' : 'Simpan kontrak'}
            </Button>
            {onCancel ? (
              <Button variant="subtle" onClick={onCancel}>
                Batal ubah
              </Button>
            ) : null}
          </div>

          {saving.isError ? <ErrorNote message={errorMessage(saving.error)} /> : null}
          {createContract.isSuccess && !contract ? (
            <SuccessNote
              message={`Kontrak ${createContract.data.number} untuk ${createContract.data.customerName} berhasil dibuat.`}
            />
          ) : null}
        </form>
      ) : null}
    </FormPanel>
  )
}
