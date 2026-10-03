import { useState } from 'react'

import { useBillingList } from '../../controllers/useBilling'
import {
  BILLING_STATUSES,
  BILLING_STATUS_LABEL,
  REFERENCE_SEARCH_MAX_LENGTH,
  type BillingStatus,
  type Invoice,
} from '../../models/billing'
import { rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { BalanceCell, DueDateCell, ReferenceCell, StatusChip } from './BillingCells'
import { BillingDetail } from './BillingDetail'
import { ReceivableForm } from './BillingForms'
import type { Directory } from './directory'

const PAGE_SIZE = 10

// draftInvoice datang dari tombol "Catat sebagai piutang" di rincian faktur.
// Formulir langsung terbuka dengan isian dari faktur itu.
export function ReceivablesSection({ directory, today, draftInvoice }: {
  directory: Directory
  today: string
  draftInvoice: Invoice | null
}) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<BillingStatus | ''>('')
  const [customerId, setCustomerId] = useState('')
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [draft, setDraft] = useState(draftInvoice)
  const [isFormOpen, setIsFormOpen] = useState(draftInvoice !== null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const receivables = useBillingList('receivables', {
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    customerId: customerId === '' ? undefined : customerId,
    projectId: projectId === '' ? undefined : projectId,
    page,
    pageSize: PAGE_SIZE,
  })
  const isFiltered = search.trim() !== '' || status !== '' || customerId !== '' || projectId !== ''

  const toggleForm = () => {
    if (isFormOpen) {
      setDraft(null)
    }
    setIsFormOpen((open) => !open)
  }

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  return (
    <div className="space-y-6">
      <Card
        title="Daftar piutang"
        description="Uang yang masih harus diterima dari pelanggan, diurutkan dari jatuh tempo paling awal"
      >
        <Toolbar>
          <ToolbarInput
            id="receivable-search"
            label="Cari referensi piutang"
            type="search"
            placeholder="Cari referensi"
            className="w-56"
            maxLength={REFERENCE_SEARCH_MAX_LENGTH}
            value={search}
            onChange={(event) => filterChanged(setSearch)(event.target.value)}
          />
          <FilterSelect
            id="receivable-filter-status"
            label="Saring menurut status piutang"
            value={status}
            onChange={(event) => filterChanged(setStatus)(event.target.value as BillingStatus | '')}
          >
            <option value="">Semua status</option>
            {BILLING_STATUSES.map((option) => (
              <option key={option} value={option}>
                {BILLING_STATUS_LABEL[option]}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect
            id="receivable-filter-customer"
            label="Saring menurut pelanggan"
            value={customerId}
            onChange={(event) => filterChanged(setCustomerId)(event.target.value)}
          >
            <option value="">Semua pelanggan</option>
            {directory.customers.map((customer) => (
              <option key={customer.id} value={customer.id}>
                {customer.name}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect
            id="receivable-filter-project"
            label="Saring menurut proyek"
            value={projectId}
            onChange={(event) => filterChanged(setProjectId)(event.target.value)}
          >
            <option value="">Semua proyek</option>
            {directory.projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name}
              </option>
            ))}
          </FilterSelect>
          <FormToggle isOpen={isFormOpen} openLabel="Catat piutang" onToggle={toggleForm} />
        </Toolbar>

        {isFormOpen ? (
          <FormPanel>
            {directory.isLoading ? (
              <Loading label="Mengambil daftar pelanggan..." />
            ) : (
              <>
                {draft ? (
                  <p className="mb-4 text-sm text-slate-600">
                    Isian diambil dari faktur {draft.number}. Periksa lalu simpan untuk mencatatnya sebagai piutang.
                  </p>
                ) : null}
                <ReceivableForm directory={directory} fromInvoice={draft ?? undefined} />
              </>
            )}
          </FormPanel>
        ) : null}

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {receivables.isPending ? <Loading /> : null}
        {receivables.isError ? <LoadFailed onRetry={() => receivables.refetch()} /> : null}
        {receivables.data ? (
          <>
            <Table
              rows={receivables.data.items}
              emptyMessage={
                isFiltered
                  ? 'Tidak ada piutang yang cocok dengan penyaringan ini.'
                  : 'Belum ada piutang. Catat piutang pertama lewat tombol Catat piutang.'
              }
              columns={[
                {
                  header: 'Referensi',
                  cell: (row) => (
                    <ReferenceCell
                      reference={row.reference}
                      sources={[
                        ['Faktur', row.invoiceNumber],
                        ['Kontrak', row.contractNumber],
                      ]}
                    />
                  ),
                },
                { header: 'Pelanggan', cell: (row) => row.customerName },
                { header: 'Proyek', cell: (row) => row.projectName ?? '-' },
                { header: 'Jatuh tempo', cell: (row) => <DueDateCell record={row} today={today} /> },
                { header: 'Status', cell: (row) => <StatusChip status={row.status} /> },
                { header: 'Nilai', align: 'right', cell: (row) => rupiah(row.amount) },
                { header: 'Sisa', align: 'right', cell: (row) => <BalanceCell record={row} /> },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Rincian"
                      isActive={row.id === selectedId}
                      onClick={() => {
                        setSelectedId((current) => (current === row.id ? null : row.id))
                        setNotice(null)
                      }}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={receivables.data.page}
              totalPages={receivables.data.totalPages}
              totalItems={receivables.data.totalItems}
              unit="piutang"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <BillingDetail
          key={selectedId}
          kind="receivables"
          id={selectedId}
          today={today}
          heading={(receivable) => `piutang ${receivable.reference}`}
          fields={(receivable) => [
            { label: 'Referensi', value: receivable.reference },
            { label: 'Pelanggan', value: receivable.customerName },
            { label: 'Proyek', value: receivable.projectName ?? 'Tanpa proyek' },
            { label: 'Kontrak penjualan', value: receivable.contractNumber ?? 'Tidak ditautkan' },
            { label: 'Faktur', value: receivable.invoiceNumber ?? 'Tidak ditautkan' },
          ]}
          renderEdit={(receivable, controls) => (
            <ReceivableForm
              record={receivable}
              directory={directory}
              onSaved={controls.onSaved}
              onCancel={controls.onCancel}
            />
          )}
          onDeleted={(message) => {
            setSelectedId(null)
            setNotice(message)
          }}
        />
      ) : null}
    </div>
  )
}
