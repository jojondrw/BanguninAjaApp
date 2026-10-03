import { useState } from 'react'

import { useBillingList } from '../../controllers/useBilling'
import {
  BILLING_STATUSES,
  BILLING_STATUS_LABEL,
  INVOICE_SEARCH_MAX_LENGTH,
  PARTY_TYPES,
  PARTY_TYPE_LABEL,
  type BillingStatus,
  type Invoice,
  type PartyType,
} from '../../models/billing'
import { projectOptions } from '../../models/lookupApi'
import { rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { BalanceCell, DueDateCell, ReferenceCell, StatusChip } from './BillingCells'
import { BillingDetail } from './BillingDetail'
import { InvoiceForm } from './BillingForms'

const PAGE_SIZE = 10
const NO_PROJECT = 'Tanpa proyek'

// Satu faktur pelanggan hanya boleh dicatat sekali sebagai piutang. Faktur
// vendor dan faktur yang sudah lunas tidak ditawari pencatatan.
function ReceivableStatus({ invoice, onRecord }: { invoice: Invoice; onRecord: (invoice: Invoice) => void }) {
  if (invoice.receivableReference) {
    return <>Sudah dicatat sebagai piutang {invoice.receivableReference}</>
  }
  if (invoice.partyType !== 'customer') {
    return <>Faktur vendor tidak dicatat sebagai piutang</>
  }
  if (invoice.status === 'paid') {
    return <>Belum dicatat, dan faktur sudah lunas</>
  }
  return (
    <span className="flex flex-col items-start gap-1.5">
      <span>Belum dicatat sebagai piutang</span>
      <Button variant="subtle" onClick={() => onRecord(invoice)}>
        Catat sebagai piutang
      </Button>
    </span>
  )
}

export function InvoicesSection({ today, onRecordReceivable }: {
  today: string
  onRecordReceivable: (invoice: Invoice) => void
}) {
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
          <SearchSelect
            {...projectOptions}
            id="invoice-filter-project"
            label="Saring menurut proyek"
            compact
            allowEmpty
            emptyLabel="Semua proyek"
            className="w-52"
            value={projectId}
            onChange={(value) => filterChanged(setProjectId)(value)}
          />
          <FormToggle isOpen={isFormOpen} openLabel="Buat faktur" onToggle={() => setIsFormOpen((open) => !open)} />
        </Toolbar>

        {isFormOpen ? (
          <FormPanel>
            <InvoiceForm />
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
                      <ReferenceCell reference={row.number} sources={[['Piutang', row.receivableReference]]} />
                      {row.note ? <span className="text-xs text-slate-500">{row.note}</span> : null}
                    </span>
                  ),
                },
                {
                  header: 'Ditagihkan kepada',
                  cell: (row) => (
                    <span className="flex flex-col">
                      <span>{row.partyName ?? '-'}</span>
                      <span className="text-xs text-slate-500">{PARTY_TYPE_LABEL[row.partyType]}</span>
                    </span>
                  ),
                },
                { header: 'Proyek', cell: (row) => row.projectName ?? NO_PROJECT },
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
              value: `${invoice.partyName ?? '-'} (${PARTY_TYPE_LABEL[invoice.partyType].toLowerCase()})`,
            },
            { label: 'Proyek', value: invoice.projectName ?? NO_PROJECT },
            { label: 'Catatan', value: invoice.note || '-' },
            { label: 'Piutang', value: <ReceivableStatus invoice={invoice} onRecord={onRecordReceivable} /> },
          ]}
          linkedTo={(invoice) => (invoice.receivableReference ? `piutang ${invoice.receivableReference}` : null)}
          renderEdit={(invoice, controls) => (
            <InvoiceForm
              record={invoice}
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
