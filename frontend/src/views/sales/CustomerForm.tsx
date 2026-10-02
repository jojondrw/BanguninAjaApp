import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useConvertLead, useCreateCustomer, useUpdateCustomer } from '../../controllers/useSales'
import {
  CUSTOMER_ADDRESS_MAX_LENGTH,
  CUSTOMER_CONTACT_MAX_LENGTH,
  CUSTOMER_EMAIL_MAX_LENGTH,
  CUSTOMER_IDENTITY_MAX_LENGTH,
  CUSTOMER_NAME_MAX_LENGTH,
  EMPTY_CUSTOMER_FORM,
  customerToForm,
  type Customer,
  type CustomerFormValues,
} from '../../models/sales'
import { errorMessage } from '../../shared/errorMessage'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { FormPanel } from '../components/RecordControls'

// Satu formulir untuk tiga keperluan: pelanggan baru, ubah pelanggan, dan
// pelanggan baru yang diisi dari data prospek.
export function CustomerForm({ customer, initial, title, convertLeadId, onSaved, onCancel }: {
  customer: Customer | null
  initial?: CustomerFormValues
  title: string
  convertLeadId?: string
  onSaved?: (saved: Customer) => void
  onCancel?: () => void
}) {
  const [values, setValues] = useState<CustomerFormValues>(
    () => (customer ? customerToForm(customer) : (initial ?? EMPTY_CUSTOMER_FORM)),
  )
  const createCustomer = useCreateCustomer()
  const updateCustomer = useUpdateCustomer()
  const convertLead = useConvertLead()
  const saving = customer ? updateCustomer : convertLeadId ? convertLead : createCustomer
  const isPrefilled = customer !== null || initial !== undefined

  const update = (key: keyof CustomerFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (customer) {
      updateCustomer.mutate({ id: customer.id, values }, { onSuccess: (saved) => onSaved?.(saved) })
      return
    }
    if (convertLeadId) {
      convertLead.mutate({ leadId: convertLeadId, values }, { onSuccess: (result) => onSaved?.(result.customer) })
      return
    }
    createCustomer.mutate(values, {
      onSuccess: (saved) => {
        setValues(EMPTY_CUSTOMER_FORM)
        onSaved?.(saved)
      },
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <div>
          <h3 className="text-sm font-semibold text-slate-900">{title}</h3>
          <p className="mt-0.5 text-xs text-slate-500">Nomor identitas dipakai untuk mencegah pelanggan ganda.</p>
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="customer-name"
            label="Nama lengkap"
            autoComplete="off"
            maxLength={CUSTOMER_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="customer-identity"
            label="Nomor identitas"
            placeholder="NIK 16 digit"
            autoComplete="off"
            autoFocus={isPrefilled}
            inputMode="numeric"
            maxLength={CUSTOMER_IDENTITY_MAX_LENGTH}
            hint={`Unik, maksimal ${CUSTOMER_IDENTITY_MAX_LENGTH} karakter`}
            required
            value={values.identityNumber}
            onChange={update('identityNumber')}
          />
          <Field
            id="customer-contact"
            label="Telepon (opsional)"
            type="tel"
            autoComplete="off"
            maxLength={CUSTOMER_CONTACT_MAX_LENGTH}
            value={values.contact}
            onChange={update('contact')}
          />
          <Field
            id="customer-email"
            label="Email (opsional)"
            type="email"
            autoComplete="off"
            maxLength={CUSTOMER_EMAIL_MAX_LENGTH}
            value={values.email}
            onChange={update('email')}
          />
          <div className="md:col-span-2">
            <Field
              id="customer-address"
              label="Alamat (opsional)"
              autoComplete="off"
              maxLength={CUSTOMER_ADDRESS_MAX_LENGTH}
              value={values.address}
              onChange={update('address')}
            />
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={saving.isPending} pendingLabel="Menyimpan pelanggan">
            {customer ? 'Simpan perubahan' : 'Simpan pelanggan'}
          </Button>
          {onCancel ? (
            <Button variant="subtle" onClick={onCancel}>
              {customer ? 'Batal ubah' : 'Batal'}
            </Button>
          ) : null}
        </div>

        {saving.isError ? <ErrorNote message={errorMessage(saving.error)} /> : null}
        {createCustomer.isSuccess && !customer && !onSaved ? (
          <SuccessNote message={`Pelanggan ${createCustomer.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}
