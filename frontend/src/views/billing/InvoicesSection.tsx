import { useState } from 'react'

import { useBillingList } from '../../controllers/useBilling'
import {
  BILLING_STATUSES,
  BILLING_STATUS_LABEL,
  INVOICE_SEARCH_MAX_LENGTH,
  PARTY_TYPES,
  PARTY_TYPE_LABEL,
  type BillingStatus,
  type PartyType,
} from '../../models/billing'
import { rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { BalanceCell, DueDateCell, StatusChip } from './BillingCells'
import { BillingDetail } from './BillingDetail'
import { InvoiceForm } from './BillingForms'
import { partyName, projectName, type Directory } from './directory'

const PAGE_SIZE = 10

export function InvoicesSection({ directory, today }: { directory: Directory; today: string }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<BillingStatus | ''>('')
  const [partyType, setPartyType] = useState<PartyType | ''>('')
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const invoices = useBillingList('invoices', {
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    partyType: partyType === '' ? undefined : partyType,
    projectId: projectId === '' ? undefined : projectId,
    page,
    pageSize: PAGE_SIZE,
  })
  const isFiltered = search.trim() !== '' || status !== '' || partyType !== '' || projectId !== ''

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  return (
    <div className="space-y-6">
      <Card
        title="Daftar faktur"
        description="Tagihan bernomor ke pelanggan atau vendor, diurutkan dari jatuh tempo paling awal"
      >
        <Toolbar>
          <ToolbarInput
            id="invoice-search"
            label="Cari nomor atau catatan faktur"
            type="search"
            placeholder="Cari nomor atau catatan"
            className="w-56"
            maxLength={INVOICE_SEARCH_MAX_LENGTH}
            value={search}
            onChange={(event) => filterChanged(setSearch)(event.target.value)}
          />
          <FilterSelect
            id="invoice-filter-status"
            label="Saring menurut status faktur"
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
            id="invoice-filter-party-type"
            label="Saring menurut pihak yang ditagih"
            value={partyType}
            onChange={(event) => filterChanged(setPartyType)(event.target.value as PartyType | '')}
          >
            <option value="">Pelanggan dan vendor</option>
            {PARTY_TYPES.map((option) => (
              <option key={option} value={option}>
                {PARTY_TYPE_LABEL[option]}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect
            id="invoice-filter-project"
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
          <FormToggle isOpen={isFormOpen} openLabel="Buat faktur" onToggle={() => setIsFormOpen((open) => !open)} />
        </Toolbar>

        {isFormOpen ? (
          <FormPanel>
            {directory.isLoading ? (
              <Loading label="Mengambil pelanggan, vendor, dan proyek..." />
            ) : (
              <InvoiceForm directory={directory} />
            )}
          </FormPanel>
        ) : null}

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {invoices.isPending ? <Loading /> : null}
        {invoices.isError ? <LoadFailed onRetry={() => invoices.refetch()} /> : null}
        {invoices.data ? (
          <>
            <Table
              rows={invoices.data.items}
              emptyMessage={
                isFiltered
                  ? 'Tidak ada faktur yang cocok dengan penyaringan ini.'
                  : 'Belum ada faktur. Buat faktur pertama lewat tombol Buat faktur.'
              }
              columns={[
                {
                  header: 'Nomor',
                  cell: (row) => (
                    <span className="flex flex-col">
                      <span className="font-medium text-slate-900">{row.number}</span>
                      {row.note ? <span className="text-xs text-slate-500">{row.note}</span> : null}
                    </span>
                  ),
                },
                {
                  header: 'Ditagihkan kepada',
                  cell: (row) => (
                    <span className="flex flex-col">
                      <span>{partyName(directory, row.partyType, row.partyId)}</span>
                      <span className="text-xs text-slate-500">{PARTY_TYPE_LABEL[row.partyType]}</span>
                    </span>
                  ),
                },
                { header: 'Proyek', cell: (row) => projectName(directory, row.projectId) },
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
              page={invoices.data.page}
              totalPages={invoices.data.totalPages}
              totalItems={invoices.data.totalItems}
              unit="faktur"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <BillingDetail
          key={selectedId}
          kind="invoices"
          id={selectedId}
          today={today}
          heading={(invoice) => `faktur ${invoice.number}`}
          fields={(invoice) => [
            { label: 'Nomor faktur', value: invoice.number },
            {
              label: 'Ditagihkan kepada',
              value: `${partyName(directory, invoice.partyType, invoice.partyId)} (${PARTY_TYPE_LABEL[invoice.partyType].toLowerCase()})`,
            },
            { label: 'Proyek', value: projectName(directory, invoice.projectId) },
            { label: 'Catatan', value: invoice.note || '-' },
          ]}
          renderEdit={(invoice, controls) => (
            <InvoiceForm
              record={invoice}
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
