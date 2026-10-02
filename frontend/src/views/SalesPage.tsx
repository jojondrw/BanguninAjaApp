import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useProjects } from '../controllers/useErp'
import {
  useContracts,
  useCreateContract,
  useCreateCustomer,
  useCreateInstallment,
  useCreateUnit,
  useCustomers,
  useInstallmentCount,
  useInstallments,
  useUnitSummary,
  useUnits,
  useUpdateContractStatus,
} from '../controllers/useSales'
import type { Project } from '../models/project'
import {
  CONTRACT_NUMBER_MAX_LENGTH,
  CONTRACT_STATUS_LABEL,
  CONTRACT_STATUS_TONE,
  CONTRACT_TYPE_LABEL,
  CUSTOMER_ADDRESS_MAX_LENGTH,
  CUSTOMER_CONTACT_MAX_LENGTH,
  CUSTOMER_EMAIL_MAX_LENGTH,
  CUSTOMER_IDENTITY_MAX_LENGTH,
  CUSTOMER_NAME_MAX_LENGTH,
  EMPTY_CUSTOMER_FORM,
  EMPTY_UNIT_FORM,
  INSTALLMENT_STATUS_LABEL,
  INSTALLMENT_STATUS_TONE,
  UNIT_CODE_MAX_LENGTH,
  UNIT_STATUS_LABEL,
  UNIT_STATUS_TONE,
  UNIT_TYPE_MAX_LENGTH,
  emptyContractForm,
  type Contract,
  type ContractFormValues,
  type ContractType,
  type CustomerFormValues,
  type InstallmentFormValues,
  type NewUnitStatus,
  type UnitFormValues,
  type UnitStatus,
} from '../models/sales'
import { errorMessage } from '../shared/errorMessage'
import { number, rupiah, rupiahShort, shortDate } from '../shared/format'
import { kpiText } from '../shared/kpiText'
import { AppShell } from './components/AppShell'
import { Card, Empty, Kpi, KpiRow, LoadFailed, Loading, Table } from './components/Data'
import { Button, CONTROL_CLASS, ErrorNote, Field, SuccessNote } from './components/Form'
import { Chip, Pager, RowAction, SectionTabs, SelectField } from './components/ListTools'

const PAGE_SIZE = 10
const OPTION_LIMIT = 100

type SalesTab = 'units' | 'customers' | 'contracts'

const TABS: { value: SalesTab; label: string }[] = [
  { value: 'units', label: 'Unit' },
  { value: 'customers', label: 'Pelanggan' },
  { value: 'contracts', label: 'Kontrak' },
]

const CONTRACT_TYPES: ContractType[] = ['installment', 'mortgage', 'cash']
const NEW_UNIT_STATUSES: NewUnitStatus[] = ['available', 'on_hold']

function amountHint(value: string, fallback: string): string {
  const amount = Number(value)
  if (value === '' || !Number.isFinite(amount)) {
    return fallback
  }
  return rupiah(amount)
}

function projectLabel(projects: Project[], id: string): string {
  const project = projects.find((item) => item.id === id)
  return project ? project.name : '-'
}

function SalesKpis() {
  const summary = useUnitSummary()
  const contracts = useContracts({ pageSize: OPTION_LIMIT })
  const overdue = useInstallmentCount('overdue')
  const due = useInstallmentCount('due')

  const countOf = (status: UnitStatus) =>
    summary.data?.find((item) => item.status === status)?.total ?? 0
  const totalUnits = summary.data?.reduce((total, item) => total + item.total, 0) ?? 0

  const contractItems = contracts.data?.items ?? []
  const binding = contractItems.filter((contract) => contract.status !== 'cancelled')
  const contractValue = binding.reduce((total, contract) => total + contract.value, 0)
  const isPartial = (contracts.data?.totalItems ?? 0) > contractItems.length

  return (
    <KpiRow>
      <Kpi
        label="Unit tersedia"
        value={kpiText(summary, () => number(countOf('available')))}
        note={summary.data ? `dari ${number(totalUnits)} unit, ${number(countOf('on_hold'))} ditahan` : undefined}
      />
      <Kpi
        label="Unit terjual"
        value={kpiText(summary, () => number(countOf('sold')))}
        note={summary.data ? `${number(countOf('reserved'))} unit sedang dipesan` : undefined}
      />
      <Kpi
        label="Nilai kontrak"
        value={kpiText(contracts, () => rupiahShort(contractValue))}
        note={
          contracts.data
            ? `${number(binding.length)} kontrak di luar yang batal${isPartial ? `, dari ${OPTION_LIMIT} terbaru` : ''}`
            : undefined
        }
      />
      <Kpi
        label="Cicilan terlambat"
        value={kpiText(overdue, number)}
        note={due.data === undefined ? undefined : `${number(due.data)} jatuh tempo dalam 7 hari`}
      />
    </KpiRow>
  )
}

function NewUnitCard({ projects, isLoading }: { projects: Project[]; isLoading: boolean }) {
  const [values, setValues] = useState<UnitFormValues>(EMPTY_UNIT_FORM)
  const createUnit = useCreateUnit()

  const update = (key: keyof UnitFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createUnit.mutate(values, { onSuccess: () => setValues(EMPTY_UNIT_FORM) })
  }

  return (
    <Card title="Unit baru" description="Unit yang sudah terikat kontrak tidak bisa diubah manual.">
      {isLoading ? <Loading label="Mengambil proyek..." /> : null}
      {!isLoading && projects.length === 0 ? (
        <Empty message="Belum ada proyek. Buat proyek dulu di menu Proyek, lalu tambahkan unitnya di sini." />
      ) : null}
      {!isLoading && projects.length > 0 ? (
        <form onSubmit={submit} className="space-y-5">
          <div className="grid gap-4 md:grid-cols-3">
            <Field
              id="unit-code"
              label="Kode unit"
              placeholder="A-01"
              autoComplete="off"
              maxLength={UNIT_CODE_MAX_LENGTH}
              hint={`Unik, maksimal ${UNIT_CODE_MAX_LENGTH} karakter`}
              required
              value={values.code}
              onChange={update('code')}
            />
            <SelectField id="unit-project" label="Proyek" required value={values.projectId} onChange={update('projectId')}>
              <option value="">Pilih proyek</option>
              {projects.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name} ({project.code})
                </option>
              ))}
            </SelectField>
            <Field
              id="unit-type"
              label="Tipe"
              placeholder="Tipe 36/72"
              autoComplete="off"
              maxLength={UNIT_TYPE_MAX_LENGTH}
              required
              value={values.unitType}
              onChange={update('unitType')}
            />
            <Field
              id="unit-area"
              label="Luas (m²)"
              type="number"
              inputMode="decimal"
              min={0.01}
              step={0.01}
              placeholder="72"
              required
              value={values.areaSqm}
              onChange={update('areaSqm')}
            />
            <Field
              id="unit-price"
              label="Harga jual"
              type="number"
              inputMode="numeric"
              min={0}
              step={1}
              placeholder="0"
              hint={amountHint(values.price, 'Dalam Rupiah')}
              value={values.price}
              onChange={update('price')}
            />
            <SelectField id="unit-status" label="Status awal" value={values.status} onChange={update('status')}>
              {NEW_UNIT_STATUSES.map((status) => (
                <option key={status} value={status}>
                  {UNIT_STATUS_LABEL[status]}
                </option>
              ))}
            </SelectField>
          </div>

          <Button type="submit" isPending={createUnit.isPending} pendingLabel="Menyimpan unit">
            Simpan unit
          </Button>

          {createUnit.isError ? <ErrorNote message={errorMessage(createUnit.error)} /> : null}
          {createUnit.isSuccess ? (
            <SuccessNote message={`Unit ${createUnit.data.code} berhasil ditambahkan.`} />
          ) : null}
        </form>
      ) : null}
    </Card>
  )
}

function UnitsSection() {
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const units = useUnits({ projectId: projectId === '' ? undefined : projectId, page, pageSize: PAGE_SIZE })
  const projectItems = projects.data?.items ?? []

  return (
    <div className="space-y-6">
      {isFormOpen ? <NewUnitCard projects={projectItems} isLoading={projects.isPending} /> : null}

      <Card title="Daftar unit" description={units.data ? `${number(units.data.totalItems)} unit` : undefined}>
        <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
          <div className="w-64">
            <SelectField
              id="unit-project-filter"
              label="Saring per proyek"
              value={projectId}
              disabled={projects.isPending}
              onChange={(event) => {
                setProjectId(event.target.value)
                setPage(1)
              }}
            >
              <option value="">Semua proyek</option>
              {projectItems.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name}
                </option>
              ))}
            </SelectField>
          </div>
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={() => setIsFormOpen((open) => !open)}>
            {isFormOpen ? 'Tutup formulir' : 'Tambah unit'}
          </Button>
        </div>

        {units.isPending ? <Loading /> : null}
        {units.isError ? <LoadFailed onRetry={() => units.refetch()} /> : null}
        {units.data ? (
          <>
            <Table
              rows={units.data.items}
              emptyMessage={
                projectId === ''
                  ? 'Belum ada unit. Tambahkan unit pertama lewat tombol Tambah unit.'
                  : 'Proyek ini belum punya unit.'
              }
              columns={[
                { header: 'Kode', cell: (row) => row.code },
                { header: 'Proyek', cell: (row) => projectLabel(projectItems, row.projectId) },
                { header: 'Tipe', cell: (row) => row.unitType },
                { header: 'Luas', align: 'right', cell: (row) => `${number(row.areaSqm)} m²` },
                { header: 'Harga', align: 'right', cell: (row) => rupiahShort(row.price) },
                {
                  header: 'Status',
                  cell: (row) => <Chip label={UNIT_STATUS_LABEL[row.status]} tone={UNIT_STATUS_TONE[row.status]} />,
                },
              ]}
            />
            <Pager
              page={units.data.page}
              totalPages={units.data.totalPages}
              totalItems={units.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>
    </div>
  )
}

function NewCustomerCard() {
  const [values, setValues] = useState<CustomerFormValues>(EMPTY_CUSTOMER_FORM)
  const createCustomer = useCreateCustomer()

  const update = (key: keyof CustomerFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createCustomer.mutate(values, { onSuccess: () => setValues(EMPTY_CUSTOMER_FORM) })
  }

  return (
    <Card title="Pelanggan baru" description="Nomor identitas dipakai untuk mencegah pelanggan ganda.">
      <form onSubmit={submit} className="space-y-5">
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

        <Button type="submit" isPending={createCustomer.isPending} pendingLabel="Menyimpan pelanggan">
          Simpan pelanggan
        </Button>

        {createCustomer.isError ? <ErrorNote message={errorMessage(createCustomer.error)} /> : null}
        {createCustomer.isSuccess ? (
          <SuccessNote message={`Pelanggan ${createCustomer.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </Card>
  )
}

function CustomersSection() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const customers = useCustomers({
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <div className="space-y-6">
      {isFormOpen ? <NewCustomerCard /> : null}

      <Card
        title="Daftar pelanggan"
        description={customers.data ? `${number(customers.data.totalItems)} pelanggan` : undefined}
      >
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <label htmlFor="customer-search" className="sr-only">
              Cari pelanggan
            </label>
            <input
              id="customer-search"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
              placeholder="Cari nama pelanggan"
              className={`${CONTROL_CLASS} h-8 w-56 px-3 text-[13px]`}
            />
          </div>
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={() => setIsFormOpen((open) => !open)}>
            {isFormOpen ? 'Tutup formulir' : 'Tambah pelanggan'}
          </Button>
        </div>

        {customers.isPending ? <Loading /> : null}
        {customers.isError ? <LoadFailed onRetry={() => customers.refetch()} /> : null}
        {customers.data ? (
          <>
            <Table
              rows={customers.data.items}
              emptyMessage={
                search.trim() === ''
                  ? 'Belum ada pelanggan. Tambahkan lewat tombol Tambah pelanggan.'
                  : 'Tidak ada pelanggan dengan nama itu.'
              }
              columns={[
                { header: 'Nama', cell: (row) => row.name },
                { header: 'No. identitas', cell: (row) => row.identityNumber },
                { header: 'Telepon', cell: (row) => row.contact || '-' },
                { header: 'Email', cell: (row) => row.email || '-' },
                { header: 'Terdaftar', cell: (row) => shortDate(row.createdAt) },
              ]}
            />
            <Pager
              page={customers.data.page}
              totalPages={customers.data.totalPages}
              totalItems={customers.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>
    </div>
  )
}

function NewContractCard() {
  const [values, setValues] = useState<ContractFormValues>(emptyContractForm)
  const createContract = useCreateContract()
  const customers = useCustomers({ pageSize: OPTION_LIMIT })
  const units = useUnits({ status: 'available', pageSize: OPTION_LIMIT })
  const projects = useProjects({ pageSize: OPTION_LIMIT })
  const customerItems = customers.data?.items ?? []
  const unitItems = units.data?.items ?? []
  const projectItems = projects.data?.items ?? []

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
    createContract.mutate(values, { onSuccess: () => setValues(emptyContractForm()) })
  }

  const isLoadingOptions = customers.isPending || units.isPending
  const missing =
    !isLoadingOptions && customerItems.length === 0
      ? 'Belum ada pelanggan. Tambahkan dulu di tab Pelanggan.'
      : !isLoadingOptions && unitItems.length === 0
        ? 'Tidak ada unit yang tersedia. Tambahkan unit atau lepas unit yang ditahan di tab Unit.'
        : null

  return (
    <Card
      title="Kontrak baru"
      description="Kontrak baru berstatus Draf dan langsung mengunci unitnya menjadi Dipesan."
    >
      {isLoadingOptions ? <Loading label="Mengambil pelanggan dan unit..." /> : null}
      {missing ? <Empty message={missing} /> : null}
      {!isLoadingOptions && missing === null ? (
        <form onSubmit={submit} className="space-y-5">
          <div className="grid gap-4 md:grid-cols-3">
            <Field
              id="contract-number"
              label="Nomor kontrak"
              placeholder="KTR-2026-001"
              autoComplete="off"
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
              {customerItems.map((customer) => (
                <option key={customer.id} value={customer.id}>
                  {customer.name} ({customer.identityNumber})
                </option>
              ))}
            </SelectField>
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

          <Button type="submit" isPending={createContract.isPending} pendingLabel="Menyimpan kontrak">
            Simpan kontrak
          </Button>

          {createContract.isError ? <ErrorNote message={errorMessage(createContract.error)} /> : null}
          {createContract.isSuccess ? (
            <SuccessNote
              message={`Kontrak ${createContract.data.number} untuk ${createContract.data.customerName} berhasil dibuat.`}
            />
          ) : null}
        </form>
      ) : null}
    </Card>
  )
}

const EMPTY_INSTALLMENT_FORM: InstallmentFormValues = { installmentNumber: '', dueDate: '', amount: '' }

function ContractInstallmentsCard({ contract }: { contract: Contract }) {
  const [values, setValues] = useState<InstallmentFormValues>(EMPTY_INSTALLMENT_FORM)
  const installments = useInstallments({ contractId: contract.id, pageSize: OPTION_LIMIT })
  const createInstallment = useCreateInstallment()
  const updateStatus = useUpdateContractStatus()

  const items = installments.data?.items ?? []
  const scheduled = items.reduce((total, item) => total + item.amount, 0)
  const paid = items.filter((item) => item.status === 'paid').reduce((total, item) => total + item.amount, 0)
  const unscheduled = Math.max(0, contract.value - scheduled)
  const nextNumber = items.reduce((highest, item) => Math.max(highest, item.installmentNumber), 0) + 1
  const canSchedule = contract.status === 'draft' || contract.status === 'active'

  const update = (key: keyof InstallmentFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const filled = {
      ...values,
      installmentNumber: values.installmentNumber === '' ? String(nextNumber) : values.installmentNumber,
    }
    createInstallment.mutate(
      { contractId: contract.id, values: filled },
      { onSuccess: () => setValues(EMPTY_INSTALLMENT_FORM) },
    )
  }

  return (
    <Card
      title={`Cicilan kontrak ${contract.number}`}
      description={`${contract.customerName}, unit ${contract.unitCode}, nilai ${rupiah(contract.value)}`}
    >
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-slate-600">
          Dijadwalkan {rupiahShort(scheduled)}, terbayar {rupiahShort(paid)}, belum dijadwalkan{' '}
          {rupiahShort(unscheduled)}.
        </p>
        {contract.status === 'draft' ? (
          <Button
            isPending={updateStatus.isPending}
            pendingLabel="Mengaktifkan"
            onClick={() => updateStatus.mutate({ id: contract.id, status: 'active' })}
          >
            Aktifkan kontrak
          </Button>
        ) : null}
      </div>
      {updateStatus.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(updateStatus.error)} />
        </div>
      ) : null}

      {installments.isPending ? <Loading /> : null}
      {installments.isError ? <LoadFailed onRetry={() => installments.refetch()} /> : null}
      {installments.data ? (
        <Table
          rows={items}
          emptyMessage="Kontrak ini belum punya jadwal cicilan."
          columns={[
            { header: 'Ke', cell: (row) => row.installmentNumber },
            { header: 'Jatuh tempo', cell: (row) => shortDate(row.dueDate) },
            { header: 'Nominal', align: 'right', cell: (row) => rupiah(row.amount) },
            {
              header: 'Status',
              cell: (row) => (
                <span className="flex items-center gap-2">
                  <Chip label={INSTALLMENT_STATUS_LABEL[row.status]} tone={INSTALLMENT_STATUS_TONE[row.status]} />
                  {row.daysOverdue > 0 ? (
                    <span className="text-xs text-red-700">{row.daysOverdue} hari</span>
                  ) : null}
                </span>
              ),
            },
            { header: 'Dibayar', cell: (row) => shortDate(row.paidDate) },
          ]}
        />
      ) : null}

      {canSchedule ? (
        <form onSubmit={submit} className="mt-6 space-y-4 border-t border-slate-200 pt-5">
          <h3 className="text-sm font-semibold text-slate-900">Tambah jadwal cicilan</h3>
          <div className="grid gap-4 md:grid-cols-3">
            <Field
              id="installment-number"
              label="Angsuran ke"
              type="number"
              inputMode="numeric"
              min={1}
              step={1}
              placeholder={String(nextNumber)}
              hint={`Kosongkan untuk memakai nomor ${nextNumber}`}
              value={values.installmentNumber}
              onChange={update('installmentNumber')}
            />
            <Field
              id="installment-due-date"
              label="Jatuh tempo"
              type="date"
              required
              value={values.dueDate}
              onChange={update('dueDate')}
            />
            <Field
              id="installment-amount"
              label="Nominal"
              type="number"
              inputMode="numeric"
              min={1}
              step={1}
              required
              hint={amountHint(values.amount, `Sisa yang belum dijadwalkan ${rupiah(unscheduled)}`)}
              value={values.amount}
              onChange={update('amount')}
            />
          </div>
          <Button type="submit" isPending={createInstallment.isPending} pendingLabel="Menyimpan cicilan">
            Simpan cicilan
          </Button>
          {createInstallment.isError ? <ErrorNote message={errorMessage(createInstallment.error)} /> : null}
          {createInstallment.isSuccess ? (
            <SuccessNote message={`Angsuran ke-${createInstallment.data.installmentNumber} berhasil dijadwalkan.`} />
          ) : null}
        </form>
      ) : (
        <p className="mt-4 text-xs text-slate-500">
          Kontrak yang sudah lunas atau batal tidak bisa ditambah jadwal cicilannya.
        </p>
      )}
    </Card>
  )
}

function ContractsSection() {
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const contracts = useContracts({ page, pageSize: PAGE_SIZE })
  const selected = contracts.data?.items.find((contract) => contract.id === selectedId) ?? null

  return (
    <div className="space-y-6">
      {isFormOpen ? <NewContractCard /> : null}

      <Card
        title="Daftar kontrak"
        description={contracts.data ? `${number(contracts.data.totalItems)} kontrak` : undefined}
      >
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <p className="text-xs text-slate-500">Pilih Cicilan untuk melihat dan menambah jadwal pembayaran.</p>
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={() => setIsFormOpen((open) => !open)}>
            {isFormOpen ? 'Tutup formulir' : 'Tambah kontrak'}
          </Button>
        </div>

        {contracts.isPending ? <Loading /> : null}
        {contracts.isError ? <LoadFailed onRetry={() => contracts.refetch()} /> : null}
        {contracts.data ? (
          <>
            <Table
              rows={contracts.data.items}
              emptyMessage="Belum ada kontrak. Buat kontrak dari pelanggan dan unit yang tersedia."
              columns={[
                { header: 'Nomor', cell: (row) => row.number },
                { header: 'Pelanggan', cell: (row) => row.customerName },
                { header: 'Unit', cell: (row) => row.unitCode },
                { header: 'Cara bayar', cell: (row) => CONTRACT_TYPE_LABEL[row.type] },
                { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                { header: 'Nilai', align: 'right', cell: (row) => rupiahShort(row.value) },
                {
                  header: 'Status',
                  cell: (row) => (
                    <Chip label={CONTRACT_STATUS_LABEL[row.status]} tone={CONTRACT_STATUS_TONE[row.status]} />
                  ),
                },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Cicilan"
                      isActive={row.id === selectedId}
                      onClick={() => setSelectedId((current) => (current === row.id ? null : row.id))}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={contracts.data.page}
              totalPages={contracts.data.totalPages}
              totalItems={contracts.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selected ? <ContractInstallmentsCard key={selected.id} contract={selected} /> : null}
    </div>
  )
}

export function SalesPage() {
  const [tab, setTab] = useState<SalesTab>('units')

  return (
    <AppShell title="Penjualan" description="Unit properti, pelanggan, kontrak, dan jadwal cicilannya">
      <SalesKpis />

      <div className="mt-6 mb-4">
        <SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian penjualan" />
      </div>

      {tab === 'units' ? <UnitsSection /> : null}
      {tab === 'customers' ? <CustomersSection /> : null}
      {tab === 'contracts' ? <ContractsSection /> : null}
    </AppShell>
  )
}
