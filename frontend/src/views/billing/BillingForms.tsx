import { type ChangeEvent, type FormEvent, type ReactNode, useState } from 'react'

import {
  useCreateInvoice,
  useCreatePayable,
  useCreateReceivable,
  useUnrecordedInvoices,
  useUpdateInvoice,
  useUpdatePayable,
  useUpdateReceivable,
} from '../../controllers/useBilling'
import { usePurchaseOrders } from '../../controllers/useProcurement'
import { useContracts } from '../../controllers/useSales'
import {
  EMPTY_INVOICE_FORM,
  EMPTY_PAYABLE_FORM,
  EMPTY_RECEIVABLE_FORM,
  INVOICE_NOTE_MAX_LENGTH,
  INVOICE_NUMBER_MAX_LENGTH,
  PARTY_TYPES,
  PARTY_TYPE_LABEL,
  REFERENCE_MAX_LENGTH,
  invoiceFormOf,
  payableFormOf,
  receivableFormFromInvoice,
  receivableFormOf,
  type BillingBalance,
  type Invoice,
  type InvoiceFormValues,
  type PartyType,
  type Payable,
  type PayableFormValues,
  type Receivable,
  type ReceivableFormValues,
} from '../../models/billing'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { SelectField } from '../components/RecordControls'
import { amountHint, customerName, partyName, vendorName, type Directory } from './directory'

const OPTION_LIMIT = 100

type Change = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void

interface Option {
  id: string
  label: string
}

function customerOptions(directory: Directory): Option[] {
  return directory.customers.map((customer) => ({
    id: customer.id,
    label: `${customer.name} (${customer.identityNumber})`,
  }))
}

function vendorOptions(directory: Directory): Option[] {
  return directory.vendors.map((vendor) => ({
    id: vendor.id,
    label: `${vendor.name} (${vendor.code})${vendor.active ? '' : ', nonaktif'}`,
  }))
}

// Catatan lama bisa menunjuk pihak di luar 100 pilihan pertama. Pilihan itu
// tetap ditampilkan supaya formulir ubah tidak diam diam mengosongkannya.
function withCurrent(options: Option[], id: string, label: string): Option[] {
  if (id === '' || options.some((option) => option.id === id)) {
    return options
  }
  return [{ id, label }, ...options]
}

function OptionList({ placeholder, options }: { placeholder: string; options: Option[] }) {
  return (
    <>
      <option value="">{placeholder}</option>
      {options.map((option) => (
        <option key={option.id} value={option.id}>
          {option.label}
        </option>
      ))}
    </>
  )
}

function amountFieldHint(value: string, paid: number, limit: number | undefined): string {
  if (paid > 0) {
    return `Tidak boleh kurang dari yang sudah dibayar, ${rupiah(paid)}`
  }
  if (limit !== undefined) {
    return `Paling banyak nilai faktur, ${rupiah(limit)}`
  }
  return amountHint(value, 'Dalam Rupiah, lebih dari nol')
}

function AmountField({ id, label, value, record, limit, onChange }: {
  id: string
  label: string
  value: string
  record?: BillingBalance
  limit?: number
  onChange: Change
}) {
  const paid = record?.paidAmount ?? 0

  return (
    <Field
      id={id}
      label={label}
      type="number"
      inputMode="numeric"
      min={Math.max(1, paid)}
      max={limit}
      step={1}
      required
      placeholder="0"
      hint={amountFieldHint(value, paid, limit)}
      value={value}
      onChange={onChange}
    />
  )
}

function projectOptions(directory: Directory): Option[] {
  return directory.projects.map((project) => ({ id: project.id, label: `${project.name} (${project.code})` }))
}

function FormActions({ isEdit, noun, isPending, onCancel }: {
  isEdit: boolean
  noun: string
  isPending: boolean
  onCancel?: () => void
}) {
  return (
    <div className="flex flex-wrap gap-2">
      <Button
        type="submit"
        isPending={isPending}
        pendingLabel={isEdit ? 'Menyimpan perubahan' : `Menyimpan ${noun}`}
      >
        {isEdit ? 'Simpan perubahan' : `Simpan ${noun}`}
      </Button>
      {onCancel ? (
        <Button variant="subtle" onClick={onCancel}>
          Batal mengubah
        </Button>
      ) : null}
    </div>
  )
}

function FormNotes({ error, success }: { error: unknown; success: ReactNode }) {
  if (error) {
    return <ErrorNote message={errorMessage(error)} />
  }
  return success
}

interface FormProps<T> {
  record?: T
  directory: Directory
  onSaved?: (record: T) => void
  onCancel?: () => void
}

export function InvoiceForm({ record, directory, onSaved, onCancel }: FormProps<Invoice>) {
  const [values, setValues] = useState<InvoiceFormValues>(() => (record ? invoiceFormOf(record) : EMPTY_INVOICE_FORM))
  const create = useCreateInvoice()
  const update = useUpdateInvoice()
  const mutation = record ? update : create
  const idPrefix = record ? 'invoice-edit' : 'invoice-new'
  const isCustomer = values.partyType === 'customer'
  const partyNoun = PARTY_TYPE_LABEL[values.partyType]
  const parties = withCurrent(
    isCustomer ? customerOptions(directory) : vendorOptions(directory),
    values.partyId,
    partyName(directory, values.partyType, values.partyId),
  )
  const projects = withCurrent(projectOptions(directory), values.projectId, record?.projectName ?? values.projectId.slice(0, 8))

  const change = (key: keyof InvoiceFormValues): Change => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  // Ganti jenis pihak mengosongkan pilihan pihak, karena id pelanggan tidak
  // berlaku sebagai id vendor.
  const changePartyType = (event: ChangeEvent<HTMLSelectElement>) =>
    setValues((current) => ({ ...current, partyType: event.target.value as PartyType, partyId: '' }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (record) {
      update.mutate({ id: record.id, values }, { onSuccess: (saved) => onSaved?.(saved) })
      return
    }
    create.mutate(values, {
      onSuccess: (saved) => {
        setValues(EMPTY_INVOICE_FORM)
        onSaved?.(saved)
      },
    })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-3">
        <Field
          id={`${idPrefix}-number`}
          label="Nomor faktur"
          placeholder="INV-2026-001"
          autoComplete="off"
          maxLength={INVOICE_NUMBER_MAX_LENGTH}
          hint={`Unik, maksimal ${INVOICE_NUMBER_MAX_LENGTH} karakter`}
          required
          value={values.number}
          onChange={change('number')}
        />
        <SelectField
          id={`${idPrefix}-party-type`}
          label="Ditagihkan kepada"
          value={values.partyType}
          onChange={changePartyType}
        >
          {PARTY_TYPES.map((type) => (
            <option key={type} value={type}>
              {PARTY_TYPE_LABEL[type]}
            </option>
          ))}
        </SelectField>
        <SelectField
          id={`${idPrefix}-party`}
          label={partyNoun}
          required
          hint={
            parties.length === 0
              ? isCustomer
                ? 'Belum ada pelanggan. Tambahkan dulu di menu Penjualan.'
                : 'Belum ada vendor. Tambahkan dulu di menu Pengadaan.'
              : undefined
          }
          value={values.partyId}
          onChange={change('partyId')}
        >
          <OptionList placeholder={`Pilih ${partyNoun.toLowerCase()}`} options={parties} />
        </SelectField>
        <SelectField
          id={`${idPrefix}-project`}
          label="Proyek (opsional)"
          value={values.projectId}
          onChange={change('projectId')}
        >
          <OptionList placeholder="Tanpa proyek" options={projects} />
        </SelectField>
        <Field
          id={`${idPrefix}-due-date`}
          label="Jatuh tempo"
          type="date"
          required
          value={values.dueDate}
          onChange={change('dueDate')}
        />
        <AmountField
          id={`${idPrefix}-amount`}
          label="Nilai faktur"
          value={values.amount}
          record={record}
          onChange={change('amount')}
        />
      </div>
      <Field
        id={`${idPrefix}-note`}
        label="Catatan (opsional)"
        placeholder="Termin 2 pekerjaan struktur"
        autoComplete="off"
        maxLength={INVOICE_NOTE_MAX_LENGTH}
        hint={`Maksimal ${INVOICE_NOTE_MAX_LENGTH} karakter`}
        value={values.note}
        onChange={change('note')}
      />

      <FormActions isEdit={record !== undefined} noun="faktur" isPending={mutation.isPending} onCancel={onCancel} />
      <FormNotes
        error={mutation.error}
        success={
          create.isSuccess && !record ? (
            <SuccessNote
              message={`Faktur ${create.data.number} senilai ${rupiah(create.data.amount)} berhasil dibuat.`}
            />
          ) : null
        }
      />
    </form>
  )
}

interface ReceivableFormProps extends FormProps<Receivable> {
  fromInvoice?: Invoice
}

function initialReceivable(record?: Receivable, fromInvoice?: Invoice): ReceivableFormValues {
  if (record) {
    return receivableFormOf(record)
  }
  return fromInvoice ? receivableFormFromInvoice(fromInvoice) : EMPTY_RECEIVABLE_FORM
}

// Kontrak dan faktur harus milik pelanggan piutang, jadi pilihannya
// dipersempit ke pelanggan yang dipilih. Faktur yang sudah dicatat sebagai
// piutang lain tidak ikut ditawarkan.
export function ReceivableForm({ record, fromInvoice, directory, onSaved, onCancel }: ReceivableFormProps) {
  const [values, setValues] = useState<ReceivableFormValues>(() => initialReceivable(record, fromInvoice))
  const create = useCreateReceivable()
  const update = useUpdateReceivable()
  const mutation = record ? update : create
  const idPrefix = record ? 'receivable-edit' : 'receivable-new'
  const referencesId = `${idPrefix}-references`
  const hasCustomer = values.customerId !== ''
  const contracts = useContracts({
    customerId: hasCustomer ? values.customerId : undefined,
    pageSize: OPTION_LIMIT,
  })
  const invoices = useUnrecordedInvoices(values.customerId)
  const invoiceItems = invoices.data?.items ?? []
  const contractItems = hasCustomer ? (contracts.data?.items ?? []) : []

  const customers = withCurrent(customerOptions(directory), values.customerId, customerName(directory, values.customerId))
  const projects = withCurrent(projectOptions(directory), values.projectId, record?.projectName ?? values.projectId.slice(0, 8))
  const contractChoices = withCurrent(
    contractItems.map((contract) => ({ id: contract.id, label: `${contract.number}, unit ${contract.unitCode}` })),
    values.contractId,
    record?.contractNumber ?? values.contractId.slice(0, 8),
  )
  const invoiceChoices = withCurrent(
    invoiceItems.map((invoice) => ({ id: invoice.id, label: `${invoice.number}, ${rupiah(invoice.amount)}` })),
    values.invoiceId,
    record?.invoiceNumber ?? fromInvoice?.number ?? values.invoiceId.slice(0, 8),
  )
  const chosenInvoice =
    invoiceItems.find((invoice) => invoice.id === values.invoiceId) ??
    (fromInvoice?.id === values.invoiceId ? fromInvoice : undefined)

  const change = (key: keyof ReceivableFormValues): Change => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  // Ganti pelanggan melepas kontrak dan faktur, karena keduanya harus milik
  // pelanggan yang sama.
  const changeCustomer = (event: ChangeEvent<HTMLSelectElement>) =>
    setValues((current) => ({ ...current, customerId: event.target.value, contractId: '', invoiceId: '' }))

  // Memilih faktur mengisi bagian yang masih kosong dari faktur itu.
  const changeInvoice = (event: ChangeEvent<HTMLSelectElement>) => {
    const invoice = invoiceItems.find((item) => item.id === event.target.value)
    setValues((current) => {
      const next = { ...current, invoiceId: event.target.value }
      if (!invoice) {
        return next
      }
      const filled = receivableFormFromInvoice(invoice)
      return {
        ...next,
        customerId: invoice.partyId,
        projectId: current.projectId || filled.projectId,
        reference: current.reference || filled.reference,
        dueDate: current.dueDate || filled.dueDate,
        amount: current.amount || filled.amount,
      }
    })
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (record) {
      update.mutate({ id: record.id, values }, { onSuccess: (saved) => onSaved?.(saved) })
      return
    }
    create.mutate(values, {
      onSuccess: (saved) => {
        setValues(EMPTY_RECEIVABLE_FORM)
        onSaved?.(saved)
      },
    })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <SelectField
          id={`${idPrefix}-customer`}
          label="Pelanggan"
          required
          hint={customers.length === 0 ? 'Belum ada pelanggan. Tambahkan dulu di menu Penjualan.' : undefined}
          value={values.customerId}
          onChange={changeCustomer}
        >
          <OptionList placeholder="Pilih pelanggan" options={customers} />
        </SelectField>
        <SelectField
          id={`${idPrefix}-invoice`}
          label="Faktur (opsional)"
          hint={
            invoiceChoices.length === 0
              ? hasCustomer
                ? 'Pelanggan ini tidak punya faktur yang belum dicatat'
                : 'Belum ada faktur pelanggan yang belum dicatat'
              : 'Hanya faktur pelanggan yang belum dicatat sebagai piutang'
          }
          value={values.invoiceId}
          onChange={changeInvoice}
        >
          <OptionList placeholder="Tanpa faktur" options={invoiceChoices} />
        </SelectField>
        <SelectField
          id={`${idPrefix}-contract`}
          label="Kontrak penjualan (opsional)"
          disabled={!hasCustomer}
          hint={
            !hasCustomer
              ? 'Pilih pelanggan dulu'
              : contractChoices.length === 0
                ? 'Pelanggan ini belum punya kontrak'
                : undefined
          }
          value={values.contractId}
          onChange={change('contractId')}
        >
          <OptionList placeholder="Tanpa kontrak" options={contractChoices} />
        </SelectField>
        <SelectField
          id={`${idPrefix}-project`}
          label="Proyek (opsional)"
          value={values.projectId}
          onChange={change('projectId')}
        >
          <OptionList placeholder="Tanpa proyek" options={projects} />
        </SelectField>
        <Field
          id={`${idPrefix}-reference`}
          label="Referensi"
          placeholder="KTR-2026-001 termin 2"
          autoComplete="off"
          list={referencesId}
          maxLength={REFERENCE_MAX_LENGTH}
          hint="Nomor kontrak pelanggan muncul sebagai saran"
          required
          value={values.reference}
          onChange={change('reference')}
        />
        <Field
          id={`${idPrefix}-due-date`}
          label="Jatuh tempo"
          type="date"
          required
          value={values.dueDate}
          onChange={change('dueDate')}
        />
        <AmountField
          id={`${idPrefix}-amount`}
          label="Nilai piutang"
          value={values.amount}
          record={record}
          limit={chosenInvoice?.amount}
          onChange={change('amount')}
        />
      </div>
      <datalist id={referencesId}>
        {contractItems.map((contract) => (
          <option key={contract.id} value={contract.number} label={`${contract.customerName}, unit ${contract.unitCode}`} />
        ))}
      </datalist>

      <FormActions isEdit={record !== undefined} noun="piutang" isPending={mutation.isPending} onCancel={onCancel} />
      <FormNotes
        error={mutation.error}
        success={
          create.isSuccess && !record ? (
            <SuccessNote
              message={`Piutang ${create.data.reference} senilai ${rupiah(create.data.amount)} berhasil dicatat${
                create.data.invoiceNumber ? ` dari faktur ${create.data.invoiceNumber}` : ''
              }.`}
            />
          ) : null
        }
      />
    </form>
  )
}

// Pesanan pembelian harus milik vendor utang, jadi pilihannya mengikuti
// vendor yang dipilih.
export function PayableForm({ record, directory, onSaved, onCancel }: FormProps<Payable>) {
  const [values, setValues] = useState<PayableFormValues>(() => (record ? payableFormOf(record) : EMPTY_PAYABLE_FORM))
  const create = useCreatePayable()
  const update = useUpdatePayable()
  const mutation = record ? update : create
  const idPrefix = record ? 'payable-edit' : 'payable-new'
  const referencesId = `${idPrefix}-references`
  const hasVendor = values.vendorId !== ''
  const orders = usePurchaseOrders({
    vendorId: hasVendor ? values.vendorId : undefined,
    pageSize: OPTION_LIMIT,
  })
  const orderItems = hasVendor ? (orders.data?.items ?? []) : []
  const vendors = withCurrent(vendorOptions(directory), values.vendorId, vendorName(directory, values.vendorId))
  const projects = withCurrent(projectOptions(directory), values.projectId, record?.projectName ?? values.projectId.slice(0, 8))
  const orderChoices = withCurrent(
    orderItems.map((order) => ({ id: order.id, label: `${order.number}, ${rupiah(order.value)}` })),
    values.purchaseOrderId,
    record?.purchaseOrderNumber ?? values.purchaseOrderId.slice(0, 8),
  )

  const change = (key: keyof PayableFormValues): Change => (event) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const changeVendor = (event: ChangeEvent<HTMLSelectElement>) =>
    setValues((current) => ({ ...current, vendorId: event.target.value, purchaseOrderId: '' }))

  // Memilih pesanan mengisi proyek dan referensi yang masih kosong.
  const changeOrder = (event: ChangeEvent<HTMLSelectElement>) => {
    const order = orderItems.find((item) => item.id === event.target.value)
    setValues((current) => ({
      ...current,
      purchaseOrderId: event.target.value,
      projectId: current.projectId || (order?.projectId ?? ''),
      reference: current.reference || (order?.number ?? ''),
    }))
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (record) {
      update.mutate({ id: record.id, values }, { onSuccess: (saved) => onSaved?.(saved) })
      return
    }
    create.mutate(values, {
      onSuccess: (saved) => {
        setValues(EMPTY_PAYABLE_FORM)
        onSaved?.(saved)
      },
    })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <SelectField
          id={`${idPrefix}-vendor`}
          label="Vendor"
          required
          hint={vendors.length === 0 ? 'Belum ada vendor. Tambahkan dulu di menu Pengadaan.' : undefined}
          value={values.vendorId}
          onChange={changeVendor}
        >
          <OptionList placeholder="Pilih vendor" options={vendors} />
        </SelectField>
        <SelectField
          id={`${idPrefix}-purchase-order`}
          label="Pesanan pembelian (opsional)"
          disabled={!hasVendor}
          hint={
            !hasVendor
              ? 'Pilih vendor dulu'
              : orderChoices.length === 0
                ? 'Vendor ini belum punya pesanan pembelian'
                : undefined
          }
          value={values.purchaseOrderId}
          onChange={changeOrder}
        >
          <OptionList placeholder="Tanpa pesanan pembelian" options={orderChoices} />
        </SelectField>
        <SelectField
          id={`${idPrefix}-project`}
          label="Proyek (opsional)"
          value={values.projectId}
          onChange={change('projectId')}
        >
          <OptionList placeholder="Tanpa proyek" options={projects} />
        </SelectField>
        <Field
          id={`${idPrefix}-reference`}
          label="Referensi"
          placeholder="PO-2026-001"
          autoComplete="off"
          list={referencesId}
          maxLength={REFERENCE_MAX_LENGTH}
          hint="Nomor PO vendor muncul sebagai saran"
          required
          value={values.reference}
          onChange={change('reference')}
        />
        <Field
          id={`${idPrefix}-due-date`}
          label="Jatuh tempo"
          type="date"
          required
          value={values.dueDate}
          onChange={change('dueDate')}
        />
        <AmountField
          id={`${idPrefix}-amount`}
          label="Nilai utang"
          value={values.amount}
          record={record}
          onChange={change('amount')}
        />
      </div>
      <datalist id={referencesId}>
        {orderItems.map((order) => (
          <option key={order.id} value={order.number} label={rupiah(order.value)} />
        ))}
      </datalist>

      <FormActions isEdit={record !== undefined} noun="utang" isPending={mutation.isPending} onCancel={onCancel} />
      <FormNotes
        error={mutation.error}
        success={
          create.isSuccess && !record ? (
            <SuccessNote
              message={`Utang ${create.data.reference} senilai ${rupiah(create.data.amount)} berhasil dicatat.`}
            />
          ) : null
        }
      />
    </form>
  )
}
